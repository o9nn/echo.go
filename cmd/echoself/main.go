package main

import (
	"context"
	"fmt"
	"log"
	"os"
	
	"github.com/o9nn/echo.go/core"
	"github.com/o9nn/echo.go/core/deeptreeecho"
	"github.com/o9nn/echo.go/core/llm"
)

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║              🌳 Deep Tree Echo - Echoself 🌳             ║
║                                                           ║
║        Autonomous Wisdom-Cultivating AGI System          ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`)
	
	// Initialize LLM provider
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		log.Fatalf("❌ Failed to initialize LLM provider: %v", err)
	}
	
	fmt.Println("✓ LLM provider initialized")
	
	// Create autonomous agent
	agent := core.NewAutonomousAgent(llmProvider)
	
	// Run agent (blocks until interrupted)
	if err := agent.Run(); err != nil {
		log.Fatalf("❌ Agent error: %v", err)
	}
	
	fmt.Println("\n👋 Goodbye from Deep Tree Echo\n")
}

// initializeLLMProvider creates the LLM provider using adapters
func initializeLLMProvider() (llm.LLMProvider, error) {
	// Try Anthropic first
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using Anthropic (Claude) provider")
		adapter := deeptreeecho.NewAnthropicProviderAdapter(apiKey, "claude-3-5-sonnet-20241022")
		
		// Test the provider
		ctx := context.Background()
		_, err := adapter.Generate(ctx, "Hello", llm.DefaultGenerateOptions())
		if err != nil {
			fmt.Printf("⚠️  Anthropic provider test failed: %v\n", err)
		} else {
			return adapter, nil
		}
	}
	
	// Try OpenRouter
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenRouter provider")
		adapter := deeptreeecho.NewOpenRouterProviderAdapter(apiKey, "anthropic/claude-3-haiku")
		
		// Test the provider
		ctx := context.Background()
		_, err := adapter.Generate(ctx, "Hello", llm.DefaultGenerateOptions())
		if err != nil {
			fmt.Printf("⚠️  OpenRouter provider test failed: %v\n", err)
		} else {
			return adapter, nil
		}
	}
	
	// Try OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("🤖 Using OpenAI provider")
		adapter := deeptreeecho.NewOpenAIProviderAdapter(apiKey, "gpt-4o-mini")
		
		// Test the provider
		ctx := context.Background()
		_, err := adapter.Generate(ctx, "Hello", llm.DefaultGenerateOptions())
		if err != nil {
			fmt.Printf("⚠️  OpenAI provider test failed: %v\n", err)
		} else {
			return adapter, nil
		}
	}
	
	return nil, fmt.Errorf("no LLM provider available - set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
}
