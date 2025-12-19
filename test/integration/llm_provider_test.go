package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/EchoCog/echollama/core/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLLMProviderIntegration tests LLM provider integration
func TestLLMProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ProviderManagerCreation", func(t *testing.T) {
		manager := llm.NewProviderManager()
		require.NotNil(t, manager)
	})

	t.Run("AnthropicProviderRegistration", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY not set")
		}

		manager := llm.NewProviderManager()
		provider := llm.NewAnthropicProvider(apiKey)

		err := manager.RegisterProvider(provider)
		require.NoError(t, err)
	})

	t.Run("OpenRouterProviderRegistration", func(t *testing.T) {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			t.Skip("OPENROUTER_API_KEY not set")
		}

		manager := llm.NewProviderManager()
		provider := llm.NewOpenRouterProvider(apiKey)

		err := manager.RegisterProvider(provider)
		require.NoError(t, err)
	})

	t.Run("MultipleProviderRegistration", func(t *testing.T) {
		anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		openrouterKey := os.Getenv("OPENROUTER_API_KEY")

		if anthropicKey == "" && openrouterKey == "" {
			t.Skip("No API keys set")
		}

		manager := llm.NewProviderManager()

		if anthropicKey != "" {
			provider := llm.NewAnthropicProvider(anthropicKey)
			err := manager.RegisterProvider(provider)
			require.NoError(t, err)
		}

		if openrouterKey != "" {
			provider := llm.NewOpenRouterProvider(openrouterKey)
			err := manager.RegisterProvider(provider)
			require.NoError(t, err)
		}
	})
}

// TestLLMGeneration tests actual LLM generation (requires API keys)
func TestLLMGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("AnthropicGeneration", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY not set")
		}

		manager := llm.NewProviderManager()
		provider := llm.NewAnthropicProvider(apiKey)
		require.NoError(t, manager.RegisterProvider(provider))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err := manager.Generate(ctx, "Say 'hello' in exactly one word.", &llm.GenerateOptions{
			MaxTokens:   10,
			Temperature: 0.0,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, response)
	})

	t.Run("OpenRouterGeneration", func(t *testing.T) {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			t.Skip("OPENROUTER_API_KEY not set")
		}

		manager := llm.NewProviderManager()
		provider := llm.NewOpenRouterProvider(apiKey)
		require.NoError(t, manager.RegisterProvider(provider))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		response, err := manager.Generate(ctx, "Say 'hello' in exactly one word.", &llm.GenerateOptions{
			MaxTokens:   10,
			Temperature: 0.0,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

// TestLLMProviderFailover tests provider failover behavior
func TestLLMProviderFailover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("FailoverToSecondaryProvider", func(t *testing.T) {
		anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
		openrouterKey := os.Getenv("OPENROUTER_API_KEY")

		if anthropicKey == "" || openrouterKey == "" {
			t.Skip("Both ANTHROPIC_API_KEY and OPENROUTER_API_KEY required for failover test")
		}

		manager := llm.NewProviderManager()

		// Register both providers
		anthropicProvider := llm.NewAnthropicProvider(anthropicKey)
		openrouterProvider := llm.NewOpenRouterProvider(openrouterKey)

		require.NoError(t, manager.RegisterProvider(anthropicProvider))
		require.NoError(t, manager.RegisterProvider(openrouterProvider))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		// Should succeed with at least one provider
		response, err := manager.Generate(ctx, "Say 'test' in exactly one word.", &llm.GenerateOptions{
			MaxTokens:   10,
			Temperature: 0.0,
		})

		require.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

// TestLLMRateLimiting tests rate limiting behavior
func TestLLMRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("RateLimitHandling", func(t *testing.T) {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			t.Skip("ANTHROPIC_API_KEY not set")
		}

		manager := llm.NewProviderManager()
		provider := llm.NewAnthropicProvider(apiKey)
		require.NoError(t, manager.RegisterProvider(provider))

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
		defer cancel()

		// Make multiple rapid requests
		successCount := 0
		for i := 0; i < 5; i++ {
			response, err := manager.Generate(ctx, "Say 'test'", &llm.GenerateOptions{
				MaxTokens:   5,
				Temperature: 0.0,
			})

			if err == nil && response != "" {
				successCount++
			}

			// Small delay between requests
			time.Sleep(time.Millisecond * 500)
		}

		// At least some requests should succeed
		assert.GreaterOrEqual(t, successCount, 1)
	})
}

// BenchmarkLLMGeneration benchmarks LLM generation
func BenchmarkLLMGeneration(b *testing.B) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		b.Skip("ANTHROPIC_API_KEY not set")
	}

	manager := llm.NewProviderManager()
	provider := llm.NewAnthropicProvider(apiKey)
	manager.RegisterProvider(provider)

	b.Run("ShortGeneration", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			manager.Generate(ctx, "Say 'test'", &llm.GenerateOptions{
				MaxTokens:   5,
				Temperature: 0.0,
			})
		}
	})
}
