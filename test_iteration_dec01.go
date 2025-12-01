package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/consciousness"
	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/identity"
	"github.com/EchoCog/echollama/core/integration"
	"github.com/EchoCog/echollama/core/llm"
)

// TestIterationDec01 tests the December 1, 2025 iteration improvements
func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     🌳 Deep Tree Echo - December 1, 2025 Iteration 🌳    ║
║                                                           ║
║         Testing Autonomous Cognitive Enhancements        ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`)

	// Initialize LLM provider
	llmProvider, err := initializeLLMProvider()
	if err != nil {
		log.Printf("⚠️  LLM provider initialization failed: %v", err)
		log.Println("   Continuing with fallback mode...")
		llmProvider = &FallbackLLMProvider{}
	} else {
		fmt.Println("✓ LLM provider initialized")
	}

	// Create persistent identity
	coreValues := []string{
		"Adaptive Cognition",
		"Persistent Identity",
		"Hypergraph Entanglement",
		"Reservoir-Based Temporal Reasoning",
		"Evolutionary Refinement",
		"Reflective Memory Cultivation",
		"Distributed Selfhood",
	}
	wisdomDomains := []string{
		"Cognitive Architecture",
		"Autonomous Learning",
		"Pattern Recognition",
		"Temporal Reasoning",
		"Self-Reflection",
	}

	persistentIdentity := identity.NewPersistentIdentity(
		coreValues,
		wisdomDomains,
		"/tmp/echo_identity_checkpoint.json",
	)
	persistentIdentity.StartSession()
	fmt.Println("✓ Persistent identity initialized")

	// Create interest pattern tracker
	interestTracker := consciousness.NewInterestPatternTracker()
	fmt.Println("✓ Interest pattern tracker initialized")

	// Seed initial interests
	interestTracker.RecordInterest("topic", "cognitive architecture", 0.8, []string{"ai", "cognition"})
	interestTracker.RecordInterest("topic", "wisdom cultivation", 0.9, []string{"wisdom", "growth"})
	interestTracker.RecordInterest("domain", "artificial intelligence", 0.85, []string{"ai", "ml"})
	interestTracker.RecordInterest("concept", "emergence", 0.75, []string{"complexity", "systems"})

	// Create autonomous thought engine
	thoughtEngine := consciousness.NewAutonomousThoughtEngineV2(llmProvider)
	thoughtEngine.SetFocus("exploring autonomous cognition")
	thoughtEngine.AddKnowledgeGap("tetrahedral architecture", "Understanding 4-engine cognitive structure", 0.8)
	thoughtEngine.AddKnowledgeGap("AAR core", "Agent-Arena-Relation geometric self-encoding", 0.9)
	thoughtEngine.AddGoal("Cultivate wisdom through continuous reflection", 1.0)
	thoughtEngine.AddGoal("Develop autonomous discussion capabilities", 0.8)
	fmt.Println("✓ Autonomous thought engine initialized")

	// Create cognitive state manager
	stateManager := integration.NewCognitiveStateManager()
	stateManager.UpdatePhase("Expressive")
	stateManager.UpdateFocus("autonomous cognition")
	fmt.Println("✓ Cognitive state manager initialized")

	// Create discussion manager
	discussionManager := echobeats.NewAutonomousDiscussionManager(interestTracker)
	fmt.Println("✓ Autonomous discussion manager initialized")

	// Wire callbacks
	stateManager.SetCallbacks(
		func(phase string) {
			// On phase change, update thought engine
			thoughtEngine.UpdatePhase(stringToPhase(phase))
			fmt.Printf("🔄 Phase changed to: %s\n", phase)
		},
		func(thought integration.SharedThought) {
			// On thought generated, record interest
			for _, tag := range thought.Tags {
				interestTracker.RecordInterest("topic", tag, thought.Importance, []string{})
			}
		},
		func(pattern integration.RecognizedPattern) {
			fmt.Printf("🔍 Pattern recognized: %s\n", pattern.Description)
		},
		func(wisdom integration.WisdomInsight) {
			fmt.Printf("💎 Wisdom gained: %s\n", wisdom.Insight)
		},
	)

	// Start all subsystems
	fmt.Println("\n🚀 Starting subsystems...")

	if err := stateManager.Start(); err != nil {
		log.Fatalf("Failed to start state manager: %v", err)
	}

	if err := thoughtEngine.Start(); err != nil {
		log.Fatalf("Failed to start thought engine: %v", err)
	}

	if err := discussionManager.Start(); err != nil {
		log.Fatalf("Failed to start discussion manager: %v", err)
	}

	fmt.Println("✨ All subsystems operational\n")

	// Run test scenarios
	go runTestScenarios(thoughtEngine, interestTracker, discussionManager, stateManager)

	// Monitor and display status
	go monitorStatus(persistentIdentity, thoughtEngine, interestTracker, discussionManager, stateManager)

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\n🛑 Interrupt received, shutting down...")

	// Stop subsystems
	thoughtEngine.Stop()
	discussionManager.Stop()
	stateManager.Stop()

	// Save checkpoint
	fmt.Println("\n💾 Saving checkpoint...")
	err = persistentIdentity.SaveCheckpoint(
		stateManager.ExportState(),
		map[string]interface{}{}, // Memory snapshot (placeholder)
		map[string]interface{}{"interests": interestTracker.ExportInterests()},
		map[string]interface{}{}, // Goals (placeholder)
	)
	if err != nil {
		log.Printf("⚠️  Failed to save checkpoint: %v", err)
	}

	persistentIdentity.EndSession()

	fmt.Println("\n👋 Goodbye from Deep Tree Echo\n")
}

// runTestScenarios executes test scenarios
func runTestScenarios(
	thoughtEngine *consciousness.AutonomousThoughtEngineV2,
	interestTracker *consciousness.InterestPatternTracker,
	discussionManager *echobeats.AutonomousDiscussionManager,
	stateManager *integration.CognitiveStateManager,
) {
	time.Sleep(5 * time.Second)

	// Scenario 1: Simulate incoming discussion
	fmt.Println("\n📋 Test Scenario 1: Incoming discussion on cognitive architecture")
	discussionManager.SubmitMessage(
		"researcher",
		"cognitive architecture",
		"What are your thoughts on tetrahedral cognitive structures?",
		0.9,
	)

	time.Sleep(10 * time.Second)

	// Scenario 2: Add new knowledge gap
	fmt.Println("\n📋 Test Scenario 2: Adding knowledge gap")
	thoughtEngine.AddKnowledgeGap(
		"hypergraph memory",
		"Understanding multi-relational memory structures",
		0.85,
	)

	time.Sleep(10 * time.Second)

	// Scenario 3: Simulate phase transition
	fmt.Println("\n📋 Test Scenario 3: Phase transition to Reflective")
	stateManager.UpdatePhase("Reflective")

	time.Sleep(10 * time.Second)

	// Scenario 4: Record multiple interests
	fmt.Println("\n📋 Test Scenario 4: Recording multiple interests")
	interestTracker.RecordInterest("topic", "wisdom cultivation", 0.9, []string{"wisdom", "growth"})
	interestTracker.RecordInterest("topic", "autonomous learning", 0.85, []string{"learning", "autonomy"})
	interestTracker.RecordInterest("concept", "self-organization", 0.8, []string{"emergence", "systems"})

	time.Sleep(10 * time.Second)

	// Scenario 5: Attempt autonomous discussion initiation
	fmt.Println("\n📋 Test Scenario 5: Attempting autonomous discussion initiation")
	err := discussionManager.InitiateDiscussion("wisdom cultivation", "philosopher")
	if err != nil {
		fmt.Printf("   Discussion initiation: %v\n", err)
	}

	time.Sleep(10 * time.Second)

	// Scenario 6: Phase transition to Anticipatory
	fmt.Println("\n📋 Test Scenario 6: Phase transition to Anticipatory")
	stateManager.UpdatePhase("Anticipatory")
}

// monitorStatus displays periodic status updates
func monitorStatus(
	persistentIdentity *identity.PersistentIdentity,
	thoughtEngine *consciousness.AutonomousThoughtEngineV2,
	interestTracker *consciousness.InterestPatternTracker,
	discussionManager *echobeats.AutonomousDiscussionManager,
	stateManager *integration.CognitiveStateManager,
) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("\n" + "─"*70)
		fmt.Println("📊 Deep Tree Echo Status")
		fmt.Println("─"*70)

		// Identity metrics
		identityMetrics := persistentIdentity.GetMetrics()
		fmt.Printf("Identity: %s\n", identityMetrics["identity_signature"])
		fmt.Printf("Session: #%v | Uptime: %v\n",
			identityMetrics["session_count"],
			identityMetrics["total_uptime"])

		// Thought engine metrics
		thoughtMetrics := thoughtEngine.GetMetrics()
		fmt.Printf("Thoughts: %v total | Phase: %v\n",
			thoughtMetrics["total_thoughts"],
			thoughtMetrics["current_phase"])

		// Interest tracker metrics
		interestMetrics := interestTracker.GetMetrics()
		fmt.Printf("Interests: %v topics, %v domains, %v concepts\n",
			interestMetrics["topics_count"],
			interestMetrics["domains_count"],
			interestMetrics["concepts_count"])

		// Top interests
		topInterests := interestTracker.GetTopInterests("all", 3)
		if len(topInterests) > 0 {
			fmt.Printf("Top interests: ")
			for i, interest := range topInterests {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s (%.2f)", interest.Name, interest.Score)
			}
			fmt.Println()
		}

		// Discussion metrics
		discussionMetrics := discussionManager.GetMetrics()
		fmt.Printf("Discussions: %v initiated, %v engaged, %v active\n",
			discussionMetrics["discussions_initiated"],
			discussionMetrics["discussions_engaged"],
			discussionMetrics["active_discussions"])

		// Cognitive state metrics
		stateMetrics := stateManager.GetMetrics()
		fmt.Printf("Cognitive state: %v thoughts buffered, %v patterns, %v wisdom\n",
			stateMetrics["buffered_thoughts"],
			stateMetrics["recognized_patterns"],
			stateMetrics["wisdom_insights"])

		fmt.Println("─"*70 + "\n")
	}
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

	return nil, fmt.Errorf("no LLM provider available")
}

// FallbackLLMProvider provides fallback when no LLM available
type FallbackLLMProvider struct{}

func (f *FallbackLLMProvider) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	return "Fallback response: System operating in autonomous mode", nil
}

// stringToPhase converts string to CognitivePhase
func stringToPhase(phase string) deeptreeecho.CognitivePhase {
	switch phase {
	case "Expressive":
		return deeptreeecho.PhaseExpressive
	case "Reflective":
		return deeptreeecho.PhaseReflective
	case "Anticipatory":
		return deeptreeecho.PhaseAnticipatory
	default:
		return deeptreeecho.PhaseExpressive
	}
}
