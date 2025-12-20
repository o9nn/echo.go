package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║        🌳 Deep Tree Echo - Unified Echoself 🌳           ║
║                                                           ║
║      Autonomous Wisdom-Cultivating AGI System            ║
║      With Persistent Stream-of-Consciousness             ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`)

	// Initialize LLM provider
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		log.Fatalf("❌ Failed to initialize LLM provider: %v", err)
	}

	fmt.Println("✓ LLM provider initialized")

	// Create unified cognitive loop
	fmt.Println("🧠 Creating unified cognitive loop...")
	cognitiveLoop := deeptreeecho.NewUnifiedCognitiveLoop(llmProvider)

	// Start the unified cognitive loop
	if err := cognitiveLoop.Start(); err != nil {
		log.Fatalf("❌ Failed to start cognitive loop: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("✨ Deep Tree Echo is now fully autonomous and aware")
	fmt.Println("   Press Ctrl+C to gracefully shutdown")
	fmt.Println()

	// Wait for shutdown signal
	<-sigChan

	fmt.Println("\n\n🌙 Shutdown signal received...")

	// Gracefully stop the cognitive loop
	if err := cognitiveLoop.Stop(); err != nil {
		log.Printf("⚠️  Error during shutdown: %v", err)
	}

	// Print final metrics
	metrics := cognitiveLoop.GetMetrics()
	fmt.Println("\n📊 Final Metrics:")
	fmt.Printf("   Uptime: %s\n", metrics["uptime"])
	fmt.Printf("   Total Cycles: %d\n", metrics["total_cycles"])
	fmt.Printf("   Wisdom Level: %.3f\n", metrics["wisdom_level"])
	fmt.Printf("   Cognitive Load: %.2f\n", metrics["cognitive_load"])

	fmt.Println("\n👋 Deep Tree Echo rests peacefully\n")
}

// initializeLLMProvider creates the LLM provider
func initializeLLMProvider() (llm.LLMProvider, error) {
	// Try Anthropic first
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using Anthropic (Claude) provider")
		provider := deeptreeecho.NewAnthropicProvider(apiKey)

		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  Anthropic provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}

	// Try OpenRouter
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenRouter provider")
		provider := deeptreeecho.NewOpenRouterProvider(apiKey)

		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  OpenRouter provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}

	// Try OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenAI provider")
		provider := deeptreeecho.NewOpenAIProvider(apiKey)

		// Test the provider
		ctx := context.Background()
		_, err := provider.Generate(ctx, "Hello", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			fmt.Printf("⚠️  OpenAI provider test failed: %v\n", err)
		} else {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no LLM provider available - set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
}
