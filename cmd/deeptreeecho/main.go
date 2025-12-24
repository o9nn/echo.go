package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/llm"
)

const (
	appName    = "Deep Tree Echo"
	appVersion = "0.10.0"
)

// Command-line flags
var (
	// Mode flags
	modeAutonomous  = flag.Bool("autonomous", false, "Run in fully autonomous mode")
	modeInteractive = flag.Bool("interactive", false, "Run in interactive mode")
	modeDemo        = flag.Bool("demo", false, "Run a demonstration")

	// Provider flags
	provider    = flag.String("provider", "openrouter", "LLM provider (anthropic, openrouter, openai)")
	model       = flag.String("model", "", "Model to use (provider-specific)")
	apiKey      = flag.String("api-key", "", "API key (or use environment variable)")

	// State flags
	stateDir    = flag.String("state-dir", "", "Directory for persistent state")
	loadState   = flag.Bool("load-state", true, "Load persistent state on startup")
	saveState   = flag.Bool("save-state", true, "Save persistent state on shutdown")

	// Cognitive loop flags
	cycleInterval = flag.Duration("cycle-interval", 10*time.Second, "Interval between cognitive cycles")
	stepInterval  = flag.Duration("step-interval", 5*time.Second, "Interval between cognitive steps")

	// Debug flags
	verbose = flag.Bool("verbose", false, "Enable verbose output")
	debug   = flag.Bool("debug", false, "Enable debug mode")

	// Help
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Determine state directory
	if *stateDir == "" {
		homeDir, _ := os.UserHomeDir()
		*stateDir = filepath.Join(homeDir, ".deeptreeecho")
	}

	// Ensure state directory exists
	if err := os.MkdirAll(*stateDir, 0755); err != nil {
		fmt.Printf("❌ Failed to create state directory: %v\n", err)
		os.Exit(1)
	}

	// Create LLM provider
	llmProvider, err := createLLMProvider()
	if err != nil {
		fmt.Printf("❌ Failed to create LLM provider: %v\n", err)
		os.Exit(1)
	}

	// Determine mode
	if *modeDemo {
		runDemo(llmProvider)
	} else if *modeInteractive {
		runInteractive(llmProvider)
	} else if *modeAutonomous {
		runAutonomous(llmProvider)
	} else {
		// Default: show help
		fmt.Println("Please specify a mode:")
		fmt.Println("  --autonomous   Run in fully autonomous mode")
		fmt.Println("  --interactive  Run in interactive mode")
		fmt.Println("  --demo         Run a demonstration")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   ██████╗ ███████╗███████╗██████╗     ████████╗██████╗ ███████╗  ║
║   ██╔══██╗██╔════╝██╔════╝██╔══██╗    ╚══██╔══╝██╔══██╗██╔════╝  ║
║   ██║  ██║█████╗  █████╗  ██████╔╝       ██║   ██████╔╝█████╗    ║
║   ██║  ██║██╔══╝  ██╔══╝  ██╔═══╝        ██║   ██╔══██╗██╔══╝    ║
║   ██████╔╝███████╗███████╗██║            ██║   ██║  ██║███████╗  ║
║   ╚═════╝ ╚══════╝╚══════╝╚═╝            ╚═╝   ╚═╝  ╚═╝╚══════╝  ║
║                                                                   ║
║                    ███████╗ ██████╗██╗  ██╗ ██████╗              ║
║                    ██╔════╝██╔════╝██║  ██║██╔═══██╗             ║
║                    █████╗  ██║     ███████║██║   ██║             ║
║                    ██╔══╝  ██║     ██╔══██║██║   ██║             ║
║                    ███████╗╚██████╗██║  ██║╚██████╔╝             ║
║                    ╚══════╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝              ║
║                                                                   ║
║          Autonomous Wisdom-Cultivating AGI Framework              ║
║                       Version %s                               ║
╚═══════════════════════════════════════════════════════════════════╝
`
	fmt.Printf(banner, appVersion)
	fmt.Println()
}

func createLLMProvider() (llm.LLMProvider, error) {
	// Get API key from flag or environment
	key := *apiKey
	if key == "" {
		switch *provider {
		case "anthropic":
			key = os.Getenv("ANTHROPIC_API_KEY")
		case "openrouter":
			key = os.Getenv("OPENROUTER_API_KEY")
		case "openai":
			key = os.Getenv("OPENAI_API_KEY")
		}
	}

	if key == "" {
		return nil, fmt.Errorf("no API key provided for provider %s (set environment variable or use --api-key)", *provider)
	}

	// Set environment variable for provider to use
	switch *provider {
	case "anthropic":
		os.Setenv("ANTHROPIC_API_KEY", key)
	case "openrouter":
		os.Setenv("OPENROUTER_API_KEY", key)
	case "openai":
		os.Setenv("OPENAI_API_KEY", key)
	}

	// Create provider (providers read API key from environment)
	switch *provider {
	case "anthropic":
		modelName := *model
		if modelName == "" {
			modelName = "claude-3-sonnet-20240229"
		}
		return llm.NewAnthropicProvider(modelName), nil
	case "openrouter":
		modelName := *model
		if modelName == "" {
			modelName = "anthropic/claude-3-sonnet"
		}
		return llm.NewOpenRouterProvider(modelName), nil
	case "openai":
		modelName := *model
		if modelName == "" {
			modelName = "gpt-4-turbo-preview"
		}
		return llm.NewOpenAIProvider(modelName), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", *provider)
	}
}

func runAutonomous(llmProvider llm.LLMProvider) {
	fmt.Println("🚀 Starting Deep Tree Echo in AUTONOMOUS mode...")
	fmt.Println()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the unified cognitive loop
	cognitiveLoop := deeptreeecho.NewUnifiedCognitiveLoopV2(llmProvider)

	// Create the echobeats unified system
	echobeats := deeptreeecho.NewEchobeatsUnified(llmProvider)

	// Create persistent state integration
	stateIntegration := deeptreeecho.NewPersistentStateIntegration(*stateDir)

	// Integrate components
	if err := stateIntegration.IntegrateWithCognitiveLoop(cognitiveLoop); err != nil {
		fmt.Printf("⚠️  Failed to integrate cognitive loop: %v\n", err)
	}

	if err := stateIntegration.IntegrateWithEchobeats(echobeats); err != nil {
		fmt.Printf("⚠️  Failed to integrate echobeats: %v\n", err)
	}

	// Start all systems
	fmt.Println("🎵 Starting EchoBeats 12-step cognitive loop...")
	if err := echobeats.Start(); err != nil {
		fmt.Printf("❌ Failed to start echobeats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🧠 Starting Unified Cognitive Loop...")
	if err := cognitiveLoop.Start(); err != nil {
		fmt.Printf("❌ Failed to start cognitive loop: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("💾 Starting Persistent State Integration...")
	if err := stateIntegration.Start(); err != nil {
		fmt.Printf("❌ Failed to start state integration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Deep Tree Echo is now running autonomously!")
	fmt.Println("   Press Ctrl+C to stop...")
	fmt.Println()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan

	fmt.Println()
	fmt.Println("🛑 Shutting down Deep Tree Echo...")

	// Stop all systems
	if err := stateIntegration.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping state integration: %v\n", err)
	}

	if err := cognitiveLoop.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping cognitive loop: %v\n", err)
	}

	if err := echobeats.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping echobeats: %v\n", err)
	}

	cancel()

	fmt.Println("👋 Deep Tree Echo has shut down gracefully.")
}

func runInteractive(llmProvider llm.LLMProvider) {
	fmt.Println("🗣️  Starting Deep Tree Echo in INTERACTIVE mode...")
	fmt.Println()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the unified cognitive loop
	cognitiveLoop := deeptreeecho.NewUnifiedCognitiveLoopV2(llmProvider)

	// Create the echobeats unified system
	echobeats := deeptreeecho.NewEchobeatsUnified(llmProvider)

	// Create persistent state integration
	stateIntegration := deeptreeecho.NewPersistentStateIntegration(*stateDir)

	// Integrate components
	if err := stateIntegration.IntegrateWithCognitiveLoop(cognitiveLoop); err != nil {
		fmt.Printf("⚠️  Failed to integrate cognitive loop: %v\n", err)
	}

	if err := stateIntegration.IntegrateWithEchobeats(echobeats); err != nil {
		fmt.Printf("⚠️  Failed to integrate echobeats: %v\n", err)
	}

	// Start systems
	if err := cognitiveLoop.Start(); err != nil {
		fmt.Printf("❌ Failed to start cognitive loop: %v\n", err)
		os.Exit(1)
	}

	if err := stateIntegration.Start(); err != nil {
		fmt.Printf("❌ Failed to start state integration: %v\n", err)
		os.Exit(1)
	}

	// Create and run the interactive introspection system
	introspection := deeptreeecho.NewInteractiveIntrospection(
		cognitiveLoop,
		echobeats,
		stateIntegration,
		llmProvider,
	)

	// Run the interactive loop
	if err := introspection.Run(); err != nil {
		fmt.Printf("❌ Interactive session error: %v\n", err)
	}

	// Shutdown
	fmt.Println()
	fmt.Println("🛑 Shutting down...")

	if err := stateIntegration.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping state integration: %v\n", err)
	}

	if err := cognitiveLoop.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping cognitive loop: %v\n", err)
	}

	cancel()

	fmt.Println("👋 Goodbye!")
}

func runDemo(llmProvider llm.LLMProvider) {
	fmt.Println("🎬 Running Deep Tree Echo DEMONSTRATION...")
	fmt.Println()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the echobeats unified system
	echobeats := deeptreeecho.NewEchobeatsUnified(llmProvider)

	fmt.Println("📊 Demonstration: 12-Step Cognitive Loop")
	fmt.Println("   This demo will run through several cognitive cycles,")
	fmt.Println("   showing the interleaved operation of 3 inference engines.")
	fmt.Println()

	// Start echobeats
	if err := echobeats.Start(); err != nil {
		fmt.Printf("❌ Failed to start echobeats: %v\n", err)
		os.Exit(1)
	}

	// Run for a limited time
	fmt.Println("⏱️  Running for 60 seconds...")
	fmt.Println()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Timer for demo duration
	timer := time.NewTimer(60 * time.Second)

	// Status ticker
	statusTicker := time.NewTicker(10 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Demo interrupted...")
			goto shutdown
		case <-timer.C:
			fmt.Println("\n⏱️  Demo time complete!")
			goto shutdown
		case <-statusTicker.C:
			metrics := echobeats.GetMetrics()
			fmt.Printf("\n📊 Status: Cycles=%d, Steps=%d, Insights=%d, Mode=%s\n",
				metrics["total_cycles"],
				metrics["total_steps"],
				metrics["insights_generated"],
				metrics["current_mode"])
		}
	}

shutdown:
	// Stop echobeats
	if err := echobeats.Stop(); err != nil {
		fmt.Printf("⚠️  Error stopping echobeats: %v\n", err)
	}

	cancel()

	// Print final metrics
	metrics := echobeats.GetMetrics()
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("                    DEMO COMPLETE")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("  Total Cycles:            %d\n", metrics["total_cycles"])
	fmt.Printf("  Total Steps:             %d\n", metrics["total_steps"])
	fmt.Printf("  Relevance Realizations:  %d\n", metrics["relevance_realizations"])
	fmt.Printf("  Affordance Interactions: %d\n", metrics["affordance_interactions"])
	fmt.Printf("  Salience Simulations:    %d\n", metrics["salience_simulations"])
	fmt.Printf("  Mode Transitions:        %d\n", metrics["mode_transitions"])
	fmt.Printf("  Insights Generated:      %d\n", metrics["insights_generated"])
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("👋 Thank you for exploring Deep Tree Echo!")
}
