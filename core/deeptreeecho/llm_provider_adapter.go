package deeptreeecho

import (
	"context"
	"fmt"

	"github.com/cogpy/echo9llama/core/llm"
)

// LLMProviderAdapter adapts deeptreeecho providers to the llm.LLMProvider interface
// This enables seamless integration with the core LLM provider manager

// AnthropicProviderAdapter wraps AnthropicProvider to implement llm.LLMProvider
type AnthropicProviderAdapter struct {
	provider *AnthropicProvider
}

// NewAnthropicProviderAdapter creates a new adapter for Anthropic provider
func NewAnthropicProviderAdapter(apiKey, model string) *AnthropicProviderAdapter {
	return &AnthropicProviderAdapter{
		provider: NewAnthropicProvider(apiKey, model),
	}
}

// Generate implements llm.LLMProvider
func (a *AnthropicProviderAdapter) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("anthropic provider not initialized")
	}

	req := LLMRequest{
		SystemPrompt: opts.SystemPrompt,
		UserPrompt:   prompt,
		Temperature:  opts.Temperature,
		MaxTokens:    opts.MaxTokens,
		Context:      []Message{},
	}

	resp, err := a.provider.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

// StreamGenerate implements llm.LLMProvider
func (a *AnthropicProviderAdapter) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	// For now, implement as a single chunk response
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		result, err := a.Generate(ctx, prompt, opts)
		if err != nil {
			ch <- llm.StreamChunk{Error: err}
			return
		}
		ch <- llm.StreamChunk{Content: result, Done: true}
	}()
	return ch, nil
}

// Name implements llm.LLMProvider
func (a *AnthropicProviderAdapter) Name() string {
	return "anthropic"
}

// Available implements llm.LLMProvider
func (a *AnthropicProviderAdapter) Available() bool {
	return a.provider != nil && a.provider.available
}

// MaxTokens implements llm.LLMProvider
func (a *AnthropicProviderAdapter) MaxTokens() int {
	return 200000 // Claude 3.5 Sonnet supports 200K context
}

// OpenRouterProviderAdapter wraps OpenRouterProvider to implement llm.LLMProvider
type OpenRouterProviderAdapter struct {
	provider *OpenRouterProvider
}

// NewOpenRouterProviderAdapter creates a new adapter for OpenRouter provider
func NewOpenRouterProviderAdapter(apiKey, model string) *OpenRouterProviderAdapter {
	return &OpenRouterProviderAdapter{
		provider: NewOpenRouterProvider(apiKey, model),
	}
}

// Generate implements llm.LLMProvider
func (a *OpenRouterProviderAdapter) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("openrouter provider not initialized")
	}

	// Use the existing GenerateThought method with system prompt context
	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s", opts.SystemPrompt, prompt)
	}

	return a.provider.GenerateThought(ctx, fullPrompt)
}

// StreamGenerate implements llm.LLMProvider
func (a *OpenRouterProviderAdapter) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		result, err := a.Generate(ctx, prompt, opts)
		if err != nil {
			ch <- llm.StreamChunk{Error: err}
			return
		}
		ch <- llm.StreamChunk{Content: result, Done: true}
	}()
	return ch, nil
}

// Name implements llm.LLMProvider
func (a *OpenRouterProviderAdapter) Name() string {
	return "openrouter"
}

// Available implements llm.LLMProvider
func (a *OpenRouterProviderAdapter) Available() bool {
	return a.provider != nil && a.provider.available
}

// MaxTokens implements llm.LLMProvider
func (a *OpenRouterProviderAdapter) MaxTokens() int {
	return 128000 // Most OpenRouter models support at least 128K
}

// OpenAIProviderAdapter wraps OpenAIProvider to implement llm.LLMProvider
type OpenAIProviderAdapter struct {
	provider *OpenAIProvider
}

// NewOpenAIProviderAdapter creates a new adapter for OpenAI provider
func NewOpenAIProviderAdapter(apiKey, model string) *OpenAIProviderAdapter {
	return &OpenAIProviderAdapter{
		provider: NewOpenAIProvider(apiKey, model),
	}
}

// Generate implements llm.LLMProvider
func (a *OpenAIProviderAdapter) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("openai provider not initialized")
	}

	fullPrompt := prompt
	if opts.SystemPrompt != "" {
		fullPrompt = fmt.Sprintf("System: %s\n\nUser: %s", opts.SystemPrompt, prompt)
	}

	return a.provider.GenerateThought(ctx, fullPrompt)
}

// StreamGenerate implements llm.LLMProvider
func (a *OpenAIProviderAdapter) StreamGenerate(ctx context.Context, prompt string, opts llm.GenerateOptions) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk, 1)
	go func() {
		defer close(ch)
		result, err := a.Generate(ctx, prompt, opts)
		if err != nil {
			ch <- llm.StreamChunk{Error: err}
			return
		}
		ch <- llm.StreamChunk{Content: result, Done: true}
	}()
	return ch, nil
}

// Name implements llm.LLMProvider
func (a *OpenAIProviderAdapter) Name() string {
	return "openai"
}

// Available implements llm.LLMProvider
func (a *OpenAIProviderAdapter) Available() bool {
	return a.provider != nil && a.provider.available
}

// MaxTokens implements llm.LLMProvider
func (a *OpenAIProviderAdapter) MaxTokens() int {
	return 128000 // GPT-4o supports 128K context
}

// CreateLLMProviderManager creates a provider manager with all available providers
func CreateLLMProviderManager(anthropicKey, openRouterKey, openAIKey string) *llm.ProviderManager {
	pm := llm.NewProviderManager()

	// Register Anthropic (highest priority)
	if anthropicKey != "" {
		adapter := NewAnthropicProviderAdapter(anthropicKey, "claude-3-5-sonnet-20241022")
		pm.RegisterProvider(adapter)
	}

	// Register OpenRouter (medium priority)
	if openRouterKey != "" {
		adapter := NewOpenRouterProviderAdapter(openRouterKey, "anthropic/claude-3-haiku")
		pm.RegisterProvider(adapter)
	}

	// Register OpenAI (lower priority)
	if openAIKey != "" {
		adapter := NewOpenAIProviderAdapter(openAIKey, "gpt-4o-mini")
		pm.RegisterProvider(adapter)
	}

	// Set fallback chain
	chain := []string{}
	if anthropicKey != "" {
		chain = append(chain, "anthropic")
	}
	if openRouterKey != "" {
		chain = append(chain, "openrouter")
	}
	if openAIKey != "" {
		chain = append(chain, "openai")
	}
	if len(chain) > 0 {
		pm.SetFallbackChain(chain)
	}

	return pm
}
