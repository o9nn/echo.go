package llm

import (
	"context"
	"fmt"
)

// OpenAIProvider stub - can be implemented if needed
type OpenAIProvider struct {
	apiKey string
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4"
	}
	
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
	}
}

// Generate generates text (stub implementation)
func (oai *OpenAIProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return "", fmt.Errorf("OpenAI provider not yet implemented")
}

// StreamGenerate generates streaming text (stub implementation)
func (oai *OpenAIProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	close(ch)
	return ch, fmt.Errorf("OpenAI provider not yet implemented")
}

// Name returns the provider name
func (oai *OpenAIProvider) Name() string {
	return "OpenAI"
}

// Available returns false as this is a stub
func (oai *OpenAIProvider) Available() bool {
	return false
}

// MaxTokens returns the maximum token limit
func (oai *OpenAIProvider) MaxTokens() int {
	return 8192
}
