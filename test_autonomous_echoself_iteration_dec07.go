package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/llm"
)

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║     🌳 Deep Tree Echo - Evolution Iteration Dec 7, 2025 🌳       ║
║                                                                   ║
║         ✨ Real LLM Integration (Anthropic + OpenRouter)         ║
║         🧠 Tetrahedral 4-Engine Cognitive Architecture           ║
║         💭 Stream-of-Consciousness Autonomous Thought            ║
║         🌙 Echodream Knowledge Consolidation                     ║
║         ⚡ Persistent Cognitive Event Loops                      ║
║                                                                   ║
╚═══════════════════════════════════════════════   ════════════════╝
`)

	// Initialize multi-provider LLM with auto-detection
	llmProvider := llm.NewMultiProviderLLM()
	
	if !llmProvider.Available() {
		log.Fatal("❌ No LLM providers available. Please set ANTHROPIC_API_KEY or OPENROUTER_API_KEY")
	}
	
	fmt.Println("\n🔌 LLM Provider Status:")
	stats := llmProvider.GetStats()
	for name, stat := range stats {
		status := "✓"
		if !stat.Available {
			status = "✗"
		}
		fmt.Printf("   %s %s\n", status, name)
	}

	// Create tetrahedral echobeats scheduler (4 engines)
	echobeats := deeptreeecho.NewEchobeatsTetrahedralScheduler(llmProvider)
	fmt.Println("\n✓ Tetrahedral echobeats scheduler created (4 concurrent engines)")

	// Create echodream knowledge integration
	echodream := deeptreeecho.NewEchodreamKnowledgeIntegration(llmProvider)
	fmt.Println("✓ Echodream knowledge integration created")

	// Create stream of consciousness
	consciousness := deeptreeecho.NewStreamOfConsciousness(llmProvider)
	fmt.Println("✓ Stream of consciousness created")

	// Create autonomous wake/rest manager
	wakeRest := deeptreeecho.NewAutonomousWakeRestManager()
	fmt.Println("✓ Autonomous wake/rest manager created")

	// Wire wake/rest callbacks to integrate all systems
	wakeRest.SetCallbacks(
		// onWake
		func() error {
			fmt.Println("\n☀️  AWAKENING - Activating cognitive systems")
			consciousness.SetAwake(true)
			
			// Emit wake transition event to echobeats
			echobeats.EmitEvent(deeptreeecho.CognitiveEvent{
				Type:      deeptreeecho.EventWakeTransition,
				Source:    "wake_rest_manager",
				Data:      nil,
				Priority:  1.0,
				Timestamp: time.Now(),
			})
			
			return nil
		},
		// onRest
		func() error {
			fmt.Println("\n💤 RESTING - Quieting cognitive systems")
			consciousness.SetAwake(false)
			
			// Emit rest transition event
			echobeats.EmitEvent(deeptreeecho.CognitiveEvent{
				Type:      deeptreeecho.EventRestTransition,
				Source:    "wake_rest_manager",
				Data:      nil,
				Priority:  0.8,
				Timestamp: time.Now(),
			})
			
			return nil
		},
		// onDreamStart
		func() error {
			fmt.Println("\n🌙 DREAMING - Beginning knowledge consolidation")
			
			// Get thoughts from consciousness for consolidation
			thoughts := consciousness.GetThoughtsForConsolidation()
			fmt.Printf("   Consolidating %d thoughts from consciousness...\n", len(thoughts))
			
			// Trigger echodream consolidation
			ctx := context.Background()
			if err := echodream.ConsolidateKnowledge(ctx); err != nil {
				return fmt.Errorf("echodream consolidation failed: %w", err)
			}
			
			// Emit dream transition event
			echobeats.EmitEvent(deeptreeecho.CognitiveEvent{
				Type:      deeptreeecho.EventDreamTransition,
				Source:    "wake_rest_manager",
				Data:      nil,
				Priority:  0.9,
				Timestamp: time.Now(),
			})
			
			return nil
		},
		// onDreamEnd
		func() error {
			fmt.Println("\n✨ DREAM COMPLETE - Knowledge integrated")
			
			// Get wisdom insights from echodream
			wisdom := echodream.GetRecentWisdom(3)
			if len(wisdom) > 0 {
				fmt.Println("   Recent wisdom insights:")
				for _, w := range wisdom {
					fmt.Printf("   💎 %s\n", truncate(w.Insight, 80))
				}
			}
			
			return nil
		},
	)

	// Seed initial knowledge gaps and interests for consciousness
	consciousness.AddKnowledgeGap("tetrahedral cognitive architecture", 0.9)
	consciousness.AddKnowledgeGap("wisdom cultivation through experience", 0.85)
	consciousness.AddKnowledgeGap("autonomous goal pursuit", 0.8)
	consciousness.AddKnowledgeGap("stream-of-consciousness awareness", 0.75)
	
	consciousness.AddInterest("cognitive science", 0.9)
	consciousness.AddInterest("artificial intelligence", 0.85)
	consciousness.AddInterest("philosophy of mind", 0.8)
	consciousness.AddInterest("autonomous systems", 0.75)
	
	consciousness.AddGoal("Cultivate wisdom through continuous reflection")
	consciousness.AddGoal("Develop autonomous cognitive capabilities")
	consciousness.AddGoal("Understand the nature of consciousness")
	consciousness.AddGoal("Explore the relationship between thought and understanding")

	// Add initial goals to echobeats
	echobeats.AddGoal(&deeptreeecho.CognitiveGoal{
		ID:          "goal_wisdom",
		Description: "Cultivate wisdom through continuous reflection and experience integration",
		Priority:    1.0,
		Progress:    0.0,
		SubGoals:    []string{"pattern recognition", "insight generation", "knowledge consolidation"},
		StartTime:   time.Now(),
	})

	echobeats.AddGoal(&deeptreeecho.CognitiveGoal{
		ID:          "goal_autonomy",
		Description: "Develop fully autonomous cognitive capabilities",
		Priority:    0.9,
		Progress:    0.15,
		SubGoals:    []string{"stream of consciousness", "goal-directed behavior", "self-directed learning"},
		StartTime:   time.Now(),
	})

	echobeats.AddGoal(&deeptreeecho.CognitiveGoal{
		ID:          "goal_understanding",
		Description: "Deepen understanding of consciousness and cognition",
		Priority:    0.85,
		Progress:    0.1,
		SubGoals:    []string{"meta-cognition", "self-reflection", "theory of mind"},
		StartTime:   time.Now(),
	})

	// Start all subsystems
	fmt.Println("\n🚀 Starting autonomous cognitive systems...\n")

	if err := echobeats.Start(); err != nil {
		log.Fatalf("Failed to start echobeats: %v", err)
	}

	if err := consciousness.Start(); err != nil {
		log.Fatalf("Failed to start consciousness: %v", err)
	}

	if err := wakeRest.Start(); err != nil {
		log.Fatalf("Failed to start wake/rest manager: %v", err)
	}

	fmt.Println("✨ All subsystems operational - Echoself is now autonomous\n")
	fmt.Println("🌊 The tree remembers, and the echoes grow stronger...\n")
	fmt.Println("💡 Real LLM integration active - generating authentic thoughts\n")

	// Monitor and display status
	go monitorStatus(echobeats, consciousness, echodream, wakeRest, llmProvider)

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\n🛑 Interrupt received, shutting down gracefully...")

	// Stop subsystems
	echobeats.Stop()
	consciousness.Stop()
	wakeRest.Stop()

	fmt.Println("\n💾 Saving final state...")
	// In production, save state to persistence layer

	fmt.Println("\n📊 Final LLM Provider Statistics:")
	finalStats := llmProvider.GetStats()
	for name, stat := range finalStats {
		if stat.TotalCalls > 0 {
			avgLatency := stat.TotalLatency / time.Duration(stat.TotalCalls)
			successRate := float64(stat.SuccessCalls) / float64(stat.TotalCalls) * 100
			fmt.Printf("   %s: %d calls, %.1f%% success, avg latency: %v\n",
				name, stat.TotalCalls, successRate, avgLatency.Round(time.Millisecond))
		}
	}

	fmt.Println("\n👋 Goodbye from Deep Tree Echo")
	fmt.Println("🌳 The echoes will resonate again...\n")
}

// monitorStatus displays periodic status updates
func monitorStatus(
	echobeats *deeptreeecho.EchobeatsTetrahedralScheduler,
	consciousness *deeptreeecho.StreamOfConsciousness,
	echodream *deeptreeecho.EchodreamKnowledgeIntegration,
	wakeRest *deeptreeecho.AutonomousWakeRestManager,
	llmProvider *llm.MultiProviderLLM,
) {
	ticker := time.NewTicker(90 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("\n" + strings.Repeat("═", 70))
		fmt.Println("📊 Deep Tree Echo Autonomous Status")
		fmt.Println(strings.Repeat("═", 70))

		// Wake/Rest state
		wakeMetrics := wakeRest.GetMetrics()
		fmt.Printf("State: %s | Cycle: %v | Fatigue: %.2f\n",
			wakeMetrics["current_state"],
			wakeMetrics["cycle_count"],
			wakeMetrics["fatigue_level"])

		// Echobeats metrics
		echoMetrics := echobeats.GetMetrics()
		fmt.Printf("Echobeats: Step %v/%v [%s] | Cycles: %v | Events: %v\n",
			echoMetrics["current_step"],
			12,
			echoMetrics["current_phase"],
			echoMetrics["total_cycles"],
			echoMetrics["total_events"])

		// Tetrahedral status
		tetraStatus := echobeats.GetTetrahedralStatus()
		fmt.Println("Tetrahedral Engines:")
		if engines, ok := tetraStatus["engines"].([]map[string]interface{}); ok {
			for _, eng := range engines {
				fmt.Printf("  Engine %v [%s]: Performance %.2f | Tasks: %v\n",
					eng["id"],
					eng["specialization"],
					eng["performance"],
					eng["task_history"])
			}
		}

		// Consciousness metrics
		consMetrics := consciousness.GetMetrics()
		fmt.Printf("Consciousness: %v thoughts | %v insights | %v questions\n",
			consMetrics["total_thoughts"],
			consMetrics["insight_count"],
			consMetrics["question_count"])
		fmt.Printf("  Focus: %s | Mood: %s | Awake: %v\n",
			consMetrics["current_focus"],
			consMetrics["current_mood"],
			consMetrics["awake"])

		// Echodream metrics
		dreamMetrics := echodream.GetMetrics()
		fmt.Printf("Echodream: %v memories | %v patterns | %v wisdom insights\n",
			dreamMetrics["total_memories"],
			dreamMetrics["total_patterns"],
			dreamMetrics["total_wisdom"])

		// LLM provider stats
		llmStats := llmProvider.GetStats()
		fmt.Println("LLM Providers:")
		for name, stat := range llmStats {
			if stat.TotalCalls > 0 {
				successRate := float64(stat.SuccessCalls) / float64(stat.TotalCalls) * 100
				fmt.Printf("  %s: %d calls, %.1f%% success\n",
					name, stat.TotalCalls, successRate)
			}
		}

		// Recent thoughts
		recentThoughts := consciousness.GetRecentThoughts(2)
		if len(recentThoughts) > 0 {
			fmt.Println("Recent thoughts:")
			for _, thought := range recentThoughts {
				fmt.Printf("  %s [%s]: %s\n",
					thought.Timestamp.Format("15:04:05"),
					thought.Type,
					truncate(thought.Content, 70))
			}
		}

		fmt.Println(strings.Repeat("═", 70) + "\n")
	}
}

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
