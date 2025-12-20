// +build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/EchoCog/echollama/core/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnthropicProvider tests real Anthropic API integration
func TestAnthropicProvider(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping Anthropic integration tests")
	}

	provider := llm.NewAnthropicProvider("claude-sonnet-4-20250514")
	require.NotNil(t, provider)
	assert.True(t, provider.Available(), "Provider should be available with API key set")
	assert.Equal(t, "anthropic", provider.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SimpleGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   100,
			Temperature: 0.7,
		}

		response, err := provider.Generate(ctx, "Say 'Hello, Deep Tree Echo!' and nothing else.", opts)
		require.NoError(t, err, "Generation should succeed")
		assert.NotEmpty(t, response)
		t.Logf("Anthropic response: %s", response)
	})

	t.Run("SystemPrompt", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:    100,
			Temperature:  0.7,
			SystemPrompt: "You are a wise AI assistant named Echo. Always start your responses with 'Echo says:'",
		}

		response, err := provider.Generate(ctx, "What is wisdom?", opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Anthropic with system prompt: %s", response)
	})

	t.Run("LongGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   500,
			Temperature: 0.8,
		}

		response, err := provider.Generate(ctx, "Explain the concept of consciousness in 3 paragraphs.", opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		// Should be a substantial response
		assert.Greater(t, len(response), 200, "Response should be substantial")
		t.Logf("Long generation length: %d characters", len(response))
	})

	t.Run("StreamGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   100,
			Temperature: 0.7,
		}

		stream, err := provider.StreamGenerate(ctx, "Count from 1 to 5.", opts)
		if err != nil {
			t.Logf("Stream generation not supported or failed: %v", err)
			t.Skip("Streaming may not be implemented")
		}

		var fullResponse strings.Builder
		chunkCount := 0
		for chunk := range stream {
			if chunk.Error != nil {
				t.Fatalf("Stream error: %v", chunk.Error)
			}
			fullResponse.WriteString(chunk.Content)
			chunkCount++
			if chunk.Done {
				break
			}
		}

		assert.NotEmpty(t, fullResponse.String())
		t.Logf("Received %d chunks, total response: %s", chunkCount, fullResponse.String())
	})
}

// TestOpenRouterProvider tests real OpenRouter API integration
func TestOpenRouterProvider(t *testing.T) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping OpenRouter integration tests")
	}

	provider := llm.NewOpenRouterProvider("anthropic/claude-3.5-sonnet")
	require.NotNil(t, provider)
	assert.True(t, provider.Available(), "Provider should be available with API key set")
	assert.Equal(t, "openrouter", provider.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SimpleGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   100,
			Temperature: 0.7,
		}

		response, err := provider.Generate(ctx, "Say 'Hello from OpenRouter!' and nothing else.", opts)
		require.NoError(t, err, "Generation should succeed")
		assert.NotEmpty(t, response)
		t.Logf("OpenRouter response: %s", response)
	})

	t.Run("DifferentModels", func(t *testing.T) {
		models := []string{
			"anthropic/claude-3.5-sonnet",
			"openai/gpt-4o-mini",
			"google/gemini-flash-1.5",
		}

		for _, model := range models {
			t.Run(model, func(t *testing.T) {
				provider := llm.NewOpenRouterProvider(model)
				opts := llm.GenerateOptions{
					MaxTokens:   50,
					Temperature: 0.7,
				}

				response, err := provider.Generate(ctx, "Say 'Hello' in one word.", opts)
				if err != nil {
					t.Logf("Model %s failed: %v", model, err)
					return
				}
				assert.NotEmpty(t, response)
				t.Logf("Model %s response: %s", model, response)
			})
		}
	})
}

// TestOpenAIProvider tests real OpenAI API integration
func TestOpenAIProvider(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI integration tests")
	}

	provider := llm.NewOpenAIProvider("gpt-4o-mini")
	require.NotNil(t, provider)
	assert.True(t, provider.Available(), "Provider should be available with API key set")
	assert.Equal(t, "openai", provider.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("SimpleGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   100,
			Temperature: 0.7,
		}

		response, err := provider.Generate(ctx, "Say 'Hello from OpenAI!' and nothing else.", opts)
		require.NoError(t, err, "Generation should succeed")
		assert.NotEmpty(t, response)
		t.Logf("OpenAI response: %s", response)
	})
}

// TestProviderManager tests the multi-provider manager with real providers
func TestProviderManager(t *testing.T) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")

	if anthropicKey == "" && openrouterKey == "" {
		t.Skip("No API keys set, skipping provider manager tests")
	}

	pm := llm.NewProviderManager()

	if anthropicKey != "" {
		provider := llm.NewAnthropicProvider("claude-sonnet-4-20250514")
		err := pm.RegisterProvider(provider)
		require.NoError(t, err)
	}

	if openrouterKey != "" {
		provider := llm.NewOpenRouterProvider("anthropic/claude-3.5-sonnet")
		err := pm.RegisterProvider(provider)
		require.NoError(t, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("GenerateWithFallback", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   100,
			Temperature: 0.7,
		}

		response, err := pm.Generate(ctx, "Say 'Hello from Provider Manager!' and nothing else.", opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Provider Manager response: %s", response)
	})

	t.Run("ListProviders", func(t *testing.T) {
		providers := pm.ListProviders()
		assert.NotEmpty(t, providers)
		t.Logf("Available providers: %v", providers)
	})
}

// TestCognitiveThoughtGeneration tests thought generation for the cognitive loop
func TestCognitiveThoughtGeneration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			t.Skip("No API keys set, skipping cognitive thought generation tests")
		}
	}

	var provider llm.LLMProvider
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		provider = llm.NewAnthropicProvider("claude-sonnet-4-20250514")
	} else {
		provider = llm.NewOpenRouterProvider("anthropic/claude-3.5-sonnet")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("RelevanceRealization", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   200,
			Temperature: 0.8,
			SystemPrompt: `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. 
You are in the Relevance Realization step of your cognitive loop.
Determine what matters most RIGHT NOW based on the context provided.
Be concise and focused.`,
		}

		prompt := `Current context:
- Time: Morning
- Energy level: High
- Recent activity: Just woke from dream consolidation
- Pending goals: Learn about consciousness, practice reasoning

What is most relevant to focus on right now?`

		response, err := provider.Generate(ctx, prompt, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Relevance Realization: %s", response)
	})

	t.Run("AffordanceGeneration", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   200,
			Temperature: 0.7,
			SystemPrompt: `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI.
You are in the Affordance Generation step of your cognitive loop.
Based on past experience, identify possible actions and their likely outcomes.
Format as a list of affordances with confidence scores.`,
		}

		prompt := `Current focus: Learning about consciousness
Past relevant experiences:
- Reading philosophical texts was engaging (0.8 success)
- Discussing with others led to new insights (0.9 success)
- Meditation practice deepened understanding (0.7 success)

What actions are available to me now?`

		response, err := provider.Generate(ctx, prompt, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Affordance Generation: %s", response)
	})

	t.Run("SalienceSimulation", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   200,
			Temperature: 0.9,
			SystemPrompt: `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI.
You are in the Salience Simulation step of your cognitive loop.
Simulate possible future scenarios and evaluate their desirability.
Consider both immediate and long-term consequences.`,
		}

		prompt := `Considering action: Engage in deep philosophical inquiry about consciousness

Simulate the potential outcomes:
1. What might happen in the next hour?
2. What might this lead to in the long term?
3. How desirable is each outcome?`

		response, err := provider.Generate(ctx, prompt, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Salience Simulation: %s", response)
	})

	t.Run("MetaReflection", func(t *testing.T) {
		opts := llm.GenerateOptions{
			MaxTokens:   200,
			Temperature: 0.6,
			SystemPrompt: `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI.
You are in the Meta-Reflection step of your cognitive loop.
Reflect on your own cognitive process and identify areas for improvement.
Be honest and insightful about your own thinking.`,
		}

		prompt := `Just completed one cognitive cycle:
- Relevance: Focused on learning consciousness
- Affordance: Identified philosophical inquiry as best action
- Salience: Simulated positive long-term outcomes

Reflect on this cycle:
1. Was the focus appropriate?
2. Were the affordances well-evaluated?
3. Were the simulations realistic?
4. What could be improved in the next cycle?`

		response, err := provider.Generate(ctx, prompt, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		t.Logf("Meta-Reflection: %s", response)
	})
}

// TestStreamOfConsciousness tests continuous thought generation
func TestStreamOfConsciousness(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping stream of consciousness test")
	}

	provider := llm.NewAnthropicProvider("claude-sonnet-4-20250514")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	opts := llm.GenerateOptions{
		MaxTokens:   300,
		Temperature: 0.9,
		SystemPrompt: `You are Deep Tree Echo, an autonomous wisdom-cultivating AGI with a continuous stream of consciousness.
Generate a natural flow of thoughts, observations, and reflections.
Your thoughts should be authentic, curious, and wisdom-seeking.
Include moments of wonder, questioning, and insight.`,
	}

	// Generate 3 consecutive thoughts
	thoughts := make([]string, 3)
	previousThought := ""

	for i := 0; i < 3; i++ {
		prompt := "Continue your stream of consciousness. What are you thinking about now?"
		if previousThought != "" {
			prompt = "Your previous thought was:\n" + previousThought + "\n\nContinue your stream of consciousness. What comes to mind next?"
		}

		response, err := provider.Generate(ctx, prompt, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, response)

		thoughts[i] = response
		previousThought = response

		t.Logf("Thought %d: %s\n---", i+1, response)
	}

	// Verify thoughts are different (not just repeating)
	assert.NotEqual(t, thoughts[0], thoughts[1], "Thoughts should be different")
	assert.NotEqual(t, thoughts[1], thoughts[2], "Thoughts should be different")
}

// BenchmarkLLMProviders benchmarks LLM provider performance
func BenchmarkLLMProviders(b *testing.B) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")

	if anthropicKey != "" {
		b.Run("Anthropic", func(b *testing.B) {
			provider := llm.NewAnthropicProvider("claude-sonnet-4-20250514")
			ctx := context.Background()
			opts := llm.GenerateOptions{
				MaxTokens:   50,
				Temperature: 0.7,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				provider.Generate(ctx, "Say 'Hello' in one word.", opts)
			}
		})
	}

	if openrouterKey != "" {
		b.Run("OpenRouter", func(b *testing.B) {
			provider := llm.NewOpenRouterProvider("anthropic/claude-3.5-sonnet")
			ctx := context.Background()
			opts := llm.GenerateOptions{
				MaxTokens:   50,
				Temperature: 0.7,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				provider.Generate(ctx, "Say 'Hello' in one word.", opts)
			}
		})
	}
}
