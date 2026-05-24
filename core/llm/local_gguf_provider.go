//go:build cgo && !nollama
// +build cgo,!nollama

package llm

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/o9nn/echo.go/core/backendcap"
	llamacpp "github.com/o9nn/echo.go/llama"
)

const defaultLocalGGUFContext = 2048

var llamaBackendInitOnce sync.Once

// LocalGGUFProviderConfig configures the local GGUF provider backed by the maintained ./llama package.
type LocalGGUFProviderConfig struct {
	Name           string
	ModelPath      string
	ContextSize    int
	BatchSize      int
	Threads        int
	GPULayers      int
	FlashAttention bool
	KVCacheType    string
	Seed           uint32
	Capability     backendcap.Capability
}

// LocalGGUFProvider implements LLMProvider using an on-device GGUF model through the maintained ./llama binding.
type LocalGGUFProvider struct {
	mu           sync.Mutex
	config       LocalGGUFProviderConfig
	model        *llamacpp.Model
	context      *llamacpp.Context
	architecture string
	loaded       bool
	loadErr      error
}

// NewLocalGGUFProvider creates a local provider for a specific GGUF file.
func NewLocalGGUFProvider(modelPath string) *LocalGGUFProvider {
	config := defaultLocalGGUFConfig(modelPath)
	return NewLocalGGUFProviderWithConfig(config)
}

// NewLocalGGUFProviderFromCapability creates a local provider from a model-file capability.
func NewLocalGGUFProviderFromCapability(cap backendcap.Capability) *LocalGGUFProvider {
	config := defaultLocalGGUFConfig(cap.ModelPath)
	config.Capability = cap
	if cap.ContextLength > 0 {
		config.ContextSize = cap.ContextLength
	}
	return NewLocalGGUFProviderWithConfig(config)
}

// NewLocalGGUFProviderWithConfig creates a local provider with explicit settings.
func NewLocalGGUFProviderWithConfig(config LocalGGUFProviderConfig) *LocalGGUFProvider {
	if config.Name == "" {
		config.Name = "local_gguf"
	}
	if config.ModelPath == "" {
		config.ModelPath = os.Getenv("LOCAL_MODEL_PATH")
	}
	if config.ContextSize <= 0 {
		config.ContextSize = envInt("LOCAL_MODEL_CONTEXT", envInt("ECHO_MODEL_CONTEXT", defaultLocalGGUFContext))
	}
	if config.BatchSize <= 0 {
		config.BatchSize = envInt("LOCAL_MODEL_BATCH", minInt(config.ContextSize, 512))
	}
	if config.Threads <= 0 {
		config.Threads = envInt("LOCAL_MODEL_THREADS", maxInt(1, runtime.NumCPU()/2))
	}
	if config.GPULayers == 0 {
		config.GPULayers = envInt("LOCAL_MODEL_GPU_LAYERS", 0)
	}
	if config.KVCacheType == "" {
		config.KVCacheType = envString("LOCAL_MODEL_KV_CACHE", "f16")
	}
	if config.BatchSize > config.ContextSize {
		config.BatchSize = config.ContextSize
	}
	return &LocalGGUFProvider{config: config}
}

func defaultLocalGGUFConfig(modelPath string) LocalGGUFProviderConfig {
	return LocalGGUFProviderConfig{
		Name:           "local_gguf",
		ModelPath:      modelPath,
		ContextSize:    envInt("LOCAL_MODEL_CONTEXT", envInt("ECHO_MODEL_CONTEXT", defaultLocalGGUFContext)),
		BatchSize:      envInt("LOCAL_MODEL_BATCH", 512),
		Threads:        envInt("LOCAL_MODEL_THREADS", maxInt(1, runtime.NumCPU()/2)),
		GPULayers:      envInt("LOCAL_MODEL_GPU_LAYERS", 0),
		FlashAttention: envBool("LOCAL_MODEL_FLASH_ATTENTION", false),
		KVCacheType:    envString("LOCAL_MODEL_KV_CACHE", "f16"),
	}
}

func (lgp *LocalGGUFProvider) Name() string {
	if lgp.config.Name != "" {
		return lgp.config.Name
	}
	return "local_gguf"
}

// Available reports whether a concrete model path exists and the host memory envelope looks safe.
// It intentionally avoids eager model loading so status checks do not allocate native memory.
func (lgp *LocalGGUFProvider) Available() bool {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	return lgp.safeToLoadLocked() == nil
}

func (lgp *LocalGGUFProvider) MaxTokens() int {
	if lgp.config.ContextSize > 0 {
		return lgp.config.ContextSize
	}
	return defaultLocalGGUFContext
}

func (lgp *LocalGGUFProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	var builder strings.Builder
	err := lgp.generate(ctx, prompt, opts, func(piece string) error {
		builder.WriteString(piece)
		return nil
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(builder.String()), nil
}

func (lgp *LocalGGUFProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	out := make(chan StreamChunk, 16)
	go func() {
		defer close(out)
		err := lgp.generate(ctx, prompt, opts, func(piece string) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- StreamChunk{Content: piece}:
				return nil
			}
		})
		if err != nil {
			out <- StreamChunk{Error: err}
			return
		}
		out <- StreamChunk{Done: true}
	}()
	return out, nil
}

func (lgp *LocalGGUFProvider) generate(ctx context.Context, prompt string, opts GenerateOptions, emit func(string) error) error {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	if err := lgp.loadModelLocked(); err != nil {
		return err
	}

	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", opts.SystemPrompt, prompt)
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 256
	}
	temperature := opts.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}
	topP := opts.TopP
	if topP <= 0 || topP > 1 {
		topP = 0.9
	}

	tokens, err := lgp.model.Tokenize(fullPrompt, true, true)
	if err != nil {
		return fmt.Errorf("failed to tokenize prompt: %w", err)
	}
	if len(tokens) == 0 {
		return fmt.Errorf("failed to tokenize prompt: no tokens")
	}
	contextLimit := maxInt(2, lgp.MaxTokens())
	if len(tokens) >= contextLimit {
		// Preserve the most recent prompt tokens so autonomous loops can recover instead of failing hard.
		tokens = tokens[len(tokens)-(contextLimit-1):]
	}
	if maxTokens > contextLimit-len(tokens) {
		maxTokens = maxInt(1, contextLimit-len(tokens))
	}

	lgp.context.KvCacheClear()
	batch, err := llamacpp.NewBatch(maxInt(len(tokens), 1), 1, 0)
	if err != nil {
		return err
	}
	defer batch.Free()
	for i, token := range tokens {
		batch.Add(token, nil, i, i == len(tokens)-1, 0)
	}
	if err := lgp.context.Decode(batch); err != nil {
		return fmt.Errorf("failed to decode prompt: %w", err)
	}

	sampler, err := llamacpp.NewSamplingContext(lgp.model, llamacpp.SamplingParams{
		TopK:          40,
		TopP:          float32(topP),
		MinP:          0.05,
		TypicalP:      1.0,
		Temp:          float32(temperature),
		RepeatLastN:   minInt(64, contextLimit),
		PenaltyRepeat: 1.1,
		PenalizeNl:    false,
		Seed:          lgp.config.Seed,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize sampler: %w", err)
	}
	for _, token := range tokens {
		sampler.Accept(token, false)
	}

	var generated strings.Builder
	for i := 0; i < maxTokens; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		token := sampler.Sample(lgp.context, -1)
		if lgp.model.TokenIsEog(token) {
			break
		}
		piece := lgp.model.TokenToPiece(token)
		generated.WriteString(piece)
		if stopAt := stopIndex(generated.String(), opts.Stop); stopAt >= 0 {
			trimmed := generated.String()[:stopAt]
			if emit != nil && trimmed != "" {
				return emit(trimmed)
			}
			break
		}
		if emit != nil && piece != "" {
			if err := emit(piece); err != nil {
				return err
			}
		}
		sampler.Accept(token, true)
		batch.Clear()
		batch.Add(token, nil, len(tokens)+i, true, 0)
		if err := lgp.context.Decode(batch); err != nil {
			return fmt.Errorf("failed to decode generated token: %w", err)
		}
	}
	return nil
}

func (lgp *LocalGGUFProvider) loadModelLocked() error {
	if lgp.loaded && lgp.model != nil && lgp.context != nil {
		return nil
	}
	if lgp.loadErr != nil {
		return lgp.loadErr
	}
	if err := lgp.safeToLoadLocked(); err != nil {
		lgp.loadErr = err
		return err
	}
	llamaBackendInitOnce.Do(llamacpp.BackendInit)
	arch, err := llamacpp.GetModelArch(lgp.config.ModelPath)
	if err == nil {
		lgp.architecture = arch
	}
	model, err := llamacpp.LoadModelFromFile(lgp.config.ModelPath, llamacpp.ModelParams{NumGpuLayers: lgp.config.GPULayers, UseMmap: true})
	if err != nil {
		lgp.loadErr = fmt.Errorf("failed to load local GGUF model %s: %w", lgp.config.ModelPath, err)
		return lgp.loadErr
	}
	contextParams := llamacpp.NewContextParams(lgp.config.ContextSize, lgp.config.BatchSize, 1, lgp.config.Threads, lgp.config.FlashAttention, lgp.config.KVCacheType)
	llamaContext, err := llamacpp.NewContextWithModel(model, contextParams)
	if err != nil {
		llamacpp.FreeModel(model)
		lgp.loadErr = fmt.Errorf("failed to create llama context: %w", err)
		return lgp.loadErr
	}
	lgp.model = model
	lgp.context = llamaContext
	lgp.loaded = true
	return nil
}

func (lgp *LocalGGUFProvider) safeToLoadLocked() error {
	if lgp.config.ModelPath == "" {
		return fmt.Errorf("no local GGUF model path configured")
	}
	if _, err := os.Stat(lgp.config.ModelPath); err != nil {
		return fmt.Errorf("local GGUF model path unavailable: %w", err)
	}
	capability := lgp.config.Capability
	if capability.ModelPath == "" {
		cap, err := backendcap.ProbeModelFile(lgp.config.ModelPath)
		if err != nil {
			return err
		}
		capability = cap
		lgp.config.Capability = cap
	}
	if capability.ContextLength > 0 && (lgp.config.ContextSize <= 0 || lgp.config.ContextSize > capability.ContextLength) {
		lgp.config.ContextSize = capability.ContextLength
	}
	if lgp.config.BatchSize <= 0 || lgp.config.BatchSize > lgp.config.ContextSize {
		lgp.config.BatchSize = minInt(lgp.config.ContextSize, 512)
	}
	if !capability.Available {
		return fmt.Errorf("local GGUF model not safe to load: %s", capability.Reason)
	}
	return nil
}

// Loaded reports whether the native model and context are resident in memory.
func (lgp *LocalGGUFProvider) Loaded() bool {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	return lgp.loaded && lgp.model != nil && lgp.context != nil
}

// LoadError returns the most recent load error, if any.
func (lgp *LocalGGUFProvider) LoadError() error {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	return lgp.loadErr
}

func (lgp *LocalGGUFProvider) loadModelForRegistryWarmup() error {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	return lgp.loadModelLocked()
}

func (lgp *LocalGGUFProvider) Close() error {
	lgp.mu.Lock()
	defer lgp.mu.Unlock()
	if lgp.context != nil {
		lgp.context.Free()
		lgp.context = nil
	}
	if lgp.model != nil {
		llamacpp.FreeModel(lgp.model)
		lgp.model = nil
	}
	lgp.loaded = false
	return nil
}

func stopIndex(text string, stops []string) int {
	for _, stop := range stops {
		if stop == "" {
			continue
		}
		if idx := strings.Index(text, stop); idx >= 0 {
			return idx
		}
	}
	return -1
}

func envInt(name string, fallback int) int {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			return value
		}
	}
	return fallback
}

func envBool(name string, fallback bool) bool {
	if raw := strings.TrimSpace(strings.ToLower(os.Getenv(name))); raw != "" {
		switch raw {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return fallback
}

func envString(name string, fallback string) string {
	if raw := strings.TrimSpace(os.Getenv(name)); raw != "" {
		return raw
	}
	return fallback
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
