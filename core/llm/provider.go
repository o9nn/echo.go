package llm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/backendcap"
)

// LLMProvider defines the interface for all LLM providers
type LLMProvider interface {
	// Generate produces a completion for the given prompt
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error)

	// StreamGenerate produces a streaming completion
	StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error)

	// Name returns the provider name
	Name() string

	// Available checks if the provider is configured and available
	Available() bool

	// MaxTokens returns the maximum tokens supported
	MaxTokens() int
}

// GenerateOptions contains options for generation
type GenerateOptions struct {
	MaxTokens    int
	Temperature  float64
	TopP         float64
	Stop         []string
	SystemPrompt string
}

// DefaultGenerateOptions returns sensible defaults
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		MaxTokens:   500,
		Temperature: 0.7,
		TopP:        0.9,
		Stop:        []string{},
	}
}

// StreamChunk represents a chunk of streaming output
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// ProviderManager manages multiple LLM providers with fallback
type ProviderManager struct {
	mu              sync.RWMutex
	providers       map[string]LLMProvider
	fallbackChain   []string
	defaultProvider string

	// Metrics
	requestCount map[string]uint64
	errorCount   map[string]uint64
	totalLatency map[string]time.Duration
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers:     make(map[string]LLMProvider),
		fallbackChain: []string{},
		requestCount:  make(map[string]uint64),
		errorCount:    make(map[string]uint64),
		totalLatency:  make(map[string]time.Duration),
	}
}

// RegisterProvider adds a provider to the manager
func (pm *ProviderManager) RegisterProvider(provider LLMProvider) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := provider.Name()
	if _, exists := pm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	pm.providers[name] = provider

	// If this is the first available provider, make it default
	if pm.defaultProvider == "" && provider.Available() {
		pm.defaultProvider = name
	}

	return nil
}

// SetFallbackChain sets the order of providers to try
func (pm *ProviderManager) SetFallbackChain(chain []string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate all providers exist
	for _, name := range chain {
		if _, exists := pm.providers[name]; !exists {
			return fmt.Errorf("provider %s not registered", name)
		}
	}

	pm.fallbackChain = dedupeProviderNames(chain)
	pm.defaultProvider = firstAvailableProvider(pm.providers, pm.fallbackChain)

	return nil
}

// ApplyBackendDecision reorders the fallback chain so generation follows the
// substrate selected by backendcap.Select while preserving registered provider
// availability and existing preference order within each substrate class.
func (pm *ProviderManager) ApplyBackendDecision(decision backendcap.Decision) []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	base := append([]string{}, pm.fallbackChain...)
	if len(base) == 0 {
		base = sortedProviderNames(pm.providers)
	} else {
		present := make(map[string]bool, len(base))
		for _, name := range base {
			present[name] = true
		}
		for _, name := range sortedProviderNames(pm.providers) {
			if !present[name] {
				base = append(base, name)
			}
		}
	}

	ordered := orderProvidersForDecision(base, decision)
	pm.fallbackChain = ordered
	pm.defaultProvider = firstAvailableProvider(pm.providers, ordered)
	return append([]string{}, ordered...)
}

// GetFallbackChain returns the current provider routing order.
func (pm *ProviderManager) GetFallbackChain() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return append([]string{}, pm.fallbackChain...)
}

// Generate generates text using the default provider with fallback
func (pm *ProviderManager) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return pm.GenerateWithProvider(ctx, "", prompt, opts)
}

// GenerateWithProvider generates text using a specific provider with fallback
func (pm *ProviderManager) GenerateWithProvider(ctx context.Context, providerName string, prompt string, opts GenerateOptions) (string, error) {
	pm.mu.RLock()

	// Determine which providers to try
	providersToTry := []string{}
	if providerName != "" {
		providersToTry = append(providersToTry, providerName)
	} else if pm.defaultProvider != "" {
		providersToTry = append(providersToTry, pm.defaultProvider)
	}

	// Add fallback chain
	for _, name := range pm.fallbackChain {
		if name != providerName && name != pm.defaultProvider {
			providersToTry = append(providersToTry, name)
		}
	}

	pm.mu.RUnlock()

	if len(providersToTry) == 0 {
		return "", errors.New("no LLM providers available")
	}

	var lastErr error
	for _, name := range providersToTry {
		pm.mu.RLock()
		provider, exists := pm.providers[name]
		pm.mu.RUnlock()

		if !exists || !provider.Available() {
			continue
		}

		// Try this provider
		start := time.Now()
		result, err := provider.Generate(ctx, prompt, opts)
		latency := time.Since(start)

		// Update metrics
		pm.mu.Lock()
		pm.requestCount[name]++
		pm.totalLatency[name] += latency
		if err != nil {
			pm.errorCount[name]++
		}
		pm.mu.Unlock()

		if err == nil {
			return result, nil
		}

		lastErr = err
		// Continue to next provider in fallback chain
	}

	if lastErr != nil {
		return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
	}

	return "", errors.New("no available providers")
}

// StreamGenerate generates streaming text using the default provider
func (pm *ProviderManager) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	return pm.StreamGenerateWithProvider(ctx, "", prompt, opts)
}

// StreamGenerateWithProvider generates streaming text using a specific provider
func (pm *ProviderManager) StreamGenerateWithProvider(ctx context.Context, providerName string, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Determine which provider to use
	targetProvider := providerName
	if targetProvider == "" {
		targetProvider = pm.defaultProvider
	}

	if targetProvider == "" {
		outChan := make(chan StreamChunk, 1)
		outChan <- StreamChunk{Error: errors.New("no LLM providers available")}
		close(outChan)
		return outChan, errors.New("no LLM providers available")
	}

	provider, exists := pm.providers[targetProvider]
	if !exists || !provider.Available() {
		outChan := make(chan StreamChunk, 1)
		outChan <- StreamChunk{Error: fmt.Errorf("provider %s not available", targetProvider)}
		close(outChan)
		return outChan, fmt.Errorf("provider %s not available", targetProvider)
	}

	return provider.StreamGenerate(ctx, prompt, opts)
}

// GetProvider returns a specific provider
func (pm *ProviderManager) GetProvider(name string) (LLMProvider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// ListProviders returns all registered provider names
func (pm *ProviderManager) ListProviders() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.providers))
	for name := range pm.providers {
		names = append(names, name)
	}
	return names
}

// GetMetrics returns usage metrics for all providers
func (pm *ProviderManager) GetMetrics() map[string]ProviderMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make(map[string]ProviderMetrics)
	for name := range pm.providers {
		requests := pm.requestCount[name]
		errors := pm.errorCount[name]
		totalLatency := pm.totalLatency[name]

		var avgLatency time.Duration
		if requests > 0 {
			avgLatency = totalLatency / time.Duration(requests)
		}

		var errorRate float64
		if requests > 0 {
			errorRate = float64(errors) / float64(requests)
		}

		metrics[name] = ProviderMetrics{
			RequestCount:   requests,
			ErrorCount:     errors,
			ErrorRate:      errorRate,
			AverageLatency: avgLatency,
		}
	}

	return metrics
}

// ProviderMetrics contains usage statistics for a provider
type ProviderMetrics struct {
	RequestCount   uint64
	ErrorCount     uint64
	ErrorRate      float64
	AverageLatency time.Duration
}

// Name returns the provider manager name
func (pm *ProviderManager) Name() string {
	return "ProviderManager"
}

// Available checks if any provider is available
func (pm *ProviderManager) Available() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, provider := range pm.providers {
		if provider.Available() {
			return true
		}
	}
	return false
}

// MaxTokens returns the maximum tokens of the default provider
func (pm *ProviderManager) MaxTokens() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.defaultProvider != "" {
		if provider, exists := pm.providers[pm.defaultProvider]; exists {
			return provider.MaxTokens()
		}
	}

	// Return a reasonable default
	return 4096
}

func dedupeProviderNames(chain []string) []string {
	seen := make(map[string]bool, len(chain))
	out := make([]string, 0, len(chain))
	for _, name := range chain {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func sortedProviderNames(providers map[string]LLMProvider) []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func firstAvailableProvider(providers map[string]LLMProvider, chain []string) string {
	for _, name := range chain {
		if provider, ok := providers[name]; ok && provider.Available() {
			return name
		}
	}
	return ""
}

func orderProvidersForDecision(base []string, decision backendcap.Decision) []string {
	ordered := append([]string{}, dedupeProviderNames(base)...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return providerScoreForDecision(ordered[i], decision) > providerScoreForDecision(ordered[j], decision)
	})
	return ordered
}

func providerScoreForDecision(name string, decision backendcap.Decision) int {
	kind := providerBackendKind(name)
	score := 0
	if kind == decision.Selected.Kind {
		score += 100
	}
	if decision.Selected.Name == "simple_fallback" && kind == backendcap.BackendFallback {
		score += 100
	}
	if decision.Selected.Name == "remote_llm_provider" && kind == backendcap.BackendRemoteAPI {
		score += 80
	}
	if decision.Selected.Native && kind == backendcap.BackendNativeCPU {
		score += 80
	}
	if kind == backendcap.BackendFallback {
		score -= 50
	}
	return score
}

func providerBackendKind(name string) backendcap.BackendKind {
	switch name {
	case "local_gguf", "local_gguf_stub", "llama", "llama_legacy", "ggml":
		return backendcap.BackendNativeCPU
	case "SimpleFallback", "simple_fallback", "fallback":
		return backendcap.BackendFallback
	default:
		return backendcap.BackendRemoteAPI
	}
}
