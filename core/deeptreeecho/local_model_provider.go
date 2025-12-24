package deeptreeecho

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/cogpy/echo9llama/core/llm"
)

// LocalModelProvider provides an abstraction layer for local model execution.
// This serves as a bridge to llama.cpp when CGO is available, and falls back
// to API providers when local execution is not possible.
//
// The local model path enables:
// - Fully autonomous operation without external API dependencies
// - Lower latency for rapid cognitive cycles
// - Privacy-preserving inference
// - Massively parallel inference for echo subsystems
type LocalModelProvider struct {
	mu sync.RWMutex

	// Model configuration
	modelPath    string
	modelName    string
	contextSize  int
	batchSize    int
	threads      int

	// State
	loaded       bool
	available    bool

	// Fallback provider for when local execution is unavailable
	fallbackProvider llm.LLMProvider

	// Statistics
	totalInferences uint64
	localInferences uint64
	apiInferences   uint64
}

// LocalModelConfig configures the local model provider
type LocalModelConfig struct {
	ModelPath       string
	ModelName       string
	ContextSize     int
	BatchSize       int
	Threads         int
	FallbackProvider llm.LLMProvider
}

// NewLocalModelProvider creates a new local model provider
func NewLocalModelProvider(config LocalModelConfig) *LocalModelProvider {
	lmp := &LocalModelProvider{
		modelPath:        config.ModelPath,
		modelName:        config.ModelName,
		contextSize:      config.ContextSize,
		batchSize:        config.BatchSize,
		threads:          config.Threads,
		fallbackProvider: config.FallbackProvider,
	}

	// Set defaults
	if lmp.contextSize == 0 {
		lmp.contextSize = 4096
	}
	if lmp.batchSize == 0 {
		lmp.batchSize = 512
	}
	if lmp.threads == 0 {
		lmp.threads = 4
	}

	// Check if local model is available
	lmp.checkLocalAvailability()

	return lmp
}

// checkLocalAvailability checks if local model execution is possible
func (lmp *LocalModelProvider) checkLocalAvailability() {
	lmp.mu.Lock()
	defer lmp.mu.Unlock()

	// Check if model file exists
	if lmp.modelPath != "" {
		if _, err := os.Stat(lmp.modelPath); err == nil {
			// Model file exists, check if llama.cpp is available
			// This would require CGO and the llama package to be properly compiled
			// For now, we mark as unavailable and use fallback
			lmp.available = false
			fmt.Println("⚠️  Local model file found but llama.cpp bindings not available")
			fmt.Println("    Using API fallback provider")
		}
	}

	// If no local model, check fallback
	if !lmp.available && lmp.fallbackProvider != nil {
		lmp.available = lmp.fallbackProvider.Available()
	}
}

// Name returns the provider name
func (lmp *LocalModelProvider) Name() string {
	if lmp.loaded {
		return fmt.Sprintf("local:%s", lmp.modelName)
	}
	if lmp.fallbackProvider != nil {
		return fmt.Sprintf("fallback:%s", lmp.fallbackProvider.Name())
	}
	return "local:unavailable"
}

// Available checks if the provider is available
func (lmp *LocalModelProvider) Available() bool {
	lmp.mu.RLock()
	defer lmp.mu.RUnlock()
	return lmp.available
}

// MaxTokens returns the maximum tokens supported
func (lmp *LocalModelProvider) MaxTokens() int {
	if lmp.loaded {
		return lmp.contextSize
	}
	if lmp.fallbackProvider != nil {
		return lmp.fallbackProvider.MaxTokens()
	}
	return 4096
}

// Generate generates a response using the local model or fallback
func (lmp *LocalModelProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	lmp.mu.Lock()
	lmp.totalInferences++
	lmp.mu.Unlock()

	// Try local model first if loaded
	if lmp.loaded {
		response, err := lmp.generateLocal(ctx, prompt, opts)
		if err == nil {
			lmp.mu.Lock()
			lmp.localInferences++
			lmp.mu.Unlock()
			return response, nil
		}
		// Fall through to API on error
		fmt.Printf("⚠️  Local inference failed, using fallback: %v\n", err)
	}

	// Use fallback provider
	if lmp.fallbackProvider != nil && lmp.fallbackProvider.Available() {
		lmp.mu.Lock()
		lmp.apiInferences++
		lmp.mu.Unlock()
		return lmp.fallbackProvider.Generate(ctx, prompt, opts)
	}

	return "", fmt.Errorf("no available provider for generation")
}

// Chat performs a chat completion using the local model or fallback
// Note: This converts chat messages to a single prompt for providers that don't support chat
func (lmp *LocalModelProvider) Chat(ctx context.Context, messages []ChatMessageLocal, opts GenerateOptionsLocal) (string, error) {
	lmp.mu.Lock()
	lmp.totalInferences++
	lmp.mu.Unlock()

	// Convert messages to prompt
	prompt := convertMessagesToPrompt(messages)

	// Try local model first if loaded
	if lmp.loaded {
		llmOpts := llm.GenerateOptions{
			SystemPrompt: opts.SystemPrompt,
			MaxTokens:    opts.MaxTokens,
			Temperature:  opts.Temperature,
		}
		response, err := lmp.generateLocal(ctx, prompt, llmOpts)
		if err == nil {
			lmp.mu.Lock()
			lmp.localInferences++
			lmp.mu.Unlock()
			return response, nil
		}
		fmt.Printf("⚠️  Local chat failed, using fallback: %v\n", err)
	}

	// Use fallback provider with Generate
	if lmp.fallbackProvider != nil && lmp.fallbackProvider.Available() {
		lmp.mu.Lock()
		lmp.apiInferences++
		lmp.mu.Unlock()
		llmOpts := llm.GenerateOptions{
			SystemPrompt: opts.SystemPrompt,
			MaxTokens:    opts.MaxTokens,
			Temperature:  opts.Temperature,
		}
		return lmp.fallbackProvider.Generate(ctx, prompt, llmOpts)
	}

	return "", fmt.Errorf("no available provider for chat")
}

// generateLocal performs local model inference
// This is a placeholder for when llama.cpp bindings are available
func (lmp *LocalModelProvider) generateLocal(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	// TODO: Implement using llama.cpp bindings when CGO is available
	// This would involve:
	// 1. Loading the model if not already loaded
	// 2. Tokenizing the prompt
	// 3. Running inference
	// 4. Decoding the output tokens
	return "", fmt.Errorf("local inference not implemented - llama.cpp bindings required")
}

// ChatMessageLocal represents a chat message for local processing
type ChatMessageLocal struct {
	Role    string
	Content string
}

// GenerateOptionsLocal represents generation options for local processing
type GenerateOptionsLocal struct {
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
}

// convertMessagesToPrompt converts chat messages to a single prompt string
func convertMessagesToPrompt(messages []ChatMessageLocal) string {
	var prompt strings.Builder
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt.WriteString(fmt.Sprintf("System: %s\n\n", msg.Content))
		case "user":
			prompt.WriteString(fmt.Sprintf("User: %s\n\n", msg.Content))
		case "assistant":
			prompt.WriteString(fmt.Sprintf("Assistant: %s\n\n", msg.Content))
		}
	}
	prompt.WriteString("Assistant: ")
	return prompt.String()
}

// LoadModel attempts to load a local model
func (lmp *LocalModelProvider) LoadModel(modelPath string) error {
	lmp.mu.Lock()
	defer lmp.mu.Unlock()

	// Check if model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", modelPath)
	}

	// TODO: Implement model loading using llama.cpp bindings
	// This would involve:
	// 1. Initializing the llama backend
	// 2. Loading the model weights
	// 3. Creating the context

	lmp.modelPath = modelPath
	// lmp.loaded = true // Would be set after successful loading

	return fmt.Errorf("model loading not implemented - llama.cpp bindings required")
}

// UnloadModel unloads the current model
func (lmp *LocalModelProvider) UnloadModel() error {
	lmp.mu.Lock()
	defer lmp.mu.Unlock()

	if !lmp.loaded {
		return nil
	}

	// TODO: Implement model unloading
	lmp.loaded = false
	return nil
}

// GetStats returns inference statistics
func (lmp *LocalModelProvider) GetStats() map[string]interface{} {
	lmp.mu.RLock()
	defer lmp.mu.RUnlock()

	return map[string]interface{}{
		"total_inferences": lmp.totalInferences,
		"local_inferences": lmp.localInferences,
		"api_inferences":   lmp.apiInferences,
		"model_loaded":     lmp.loaded,
		"model_path":       lmp.modelPath,
		"model_name":       lmp.modelName,
		"context_size":     lmp.contextSize,
		"available":        lmp.available,
	}
}

// SetFallbackProvider sets the fallback API provider
func (lmp *LocalModelProvider) SetFallbackProvider(provider llm.LLMProvider) {
	lmp.mu.Lock()
	defer lmp.mu.Unlock()
	lmp.fallbackProvider = provider
	lmp.checkLocalAvailability()
}

// IsLocalAvailable returns whether local inference is available
func (lmp *LocalModelProvider) IsLocalAvailable() bool {
	lmp.mu.RLock()
	defer lmp.mu.RUnlock()
	return lmp.loaded
}

// GetFallbackProvider returns the fallback provider
func (lmp *LocalModelProvider) GetFallbackProvider() llm.LLMProvider {
	lmp.mu.RLock()
	defer lmp.mu.RUnlock()
	return lmp.fallbackProvider
}

// LocalModelProviderBuilder helps construct a LocalModelProvider
type LocalModelProviderBuilder struct {
	config LocalModelConfig
}

// NewLocalModelProviderBuilder creates a new builder
func NewLocalModelProviderBuilder() *LocalModelProviderBuilder {
	return &LocalModelProviderBuilder{
		config: LocalModelConfig{
			ContextSize: 4096,
			BatchSize:   512,
			Threads:     4,
		},
	}
}

// WithModelPath sets the model path
func (b *LocalModelProviderBuilder) WithModelPath(path string) *LocalModelProviderBuilder {
	b.config.ModelPath = path
	return b
}

// WithModelName sets the model name
func (b *LocalModelProviderBuilder) WithModelName(name string) *LocalModelProviderBuilder {
	b.config.ModelName = name
	return b
}

// WithContextSize sets the context size
func (b *LocalModelProviderBuilder) WithContextSize(size int) *LocalModelProviderBuilder {
	b.config.ContextSize = size
	return b
}

// WithBatchSize sets the batch size
func (b *LocalModelProviderBuilder) WithBatchSize(size int) *LocalModelProviderBuilder {
	b.config.BatchSize = size
	return b
}

// WithThreads sets the number of threads
func (b *LocalModelProviderBuilder) WithThreads(threads int) *LocalModelProviderBuilder {
	b.config.Threads = threads
	return b
}

// WithFallbackProvider sets the fallback provider
func (b *LocalModelProviderBuilder) WithFallbackProvider(provider llm.LLMProvider) *LocalModelProviderBuilder {
	b.config.FallbackProvider = provider
	return b
}

// Build creates the LocalModelProvider
func (b *LocalModelProviderBuilder) Build() *LocalModelProvider {
	return NewLocalModelProvider(b.config)
}
