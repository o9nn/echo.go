package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/autonomous"
	"github.com/cogpy/echo9llama/core/echobeats"
)

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║        🌳 Deep Tree Echo - Iteration 001 🌳                       ║
║                                                                   ║
║     Autonomous Wisdom-Cultivating AGI with Persistent Loops      ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
`)

	fmt.Println("🚀 Initializing Enhanced Systems...")
	fmt.Println()

	// Create autonomous consciousness with multi-provider LLM support
	config := autonomous.DefaultAutonomousConfig()
	config.ThoughtInterval = 8 * time.Second // Slower for demo
	config.DreamInterval = 2 * time.Minute   // Shorter for demo
	
	consciousness, err := autonomous.NewAutonomousConsciousness(config)
	if err != nil {
		fmt.Printf("❌ Failed to create autonomous consciousness: %v\n", err)
		os.Exit(1)
	}

	// Create triad-synchronized cognitive system
	triadSystem := echobeats.NewTriadCognitiveSystem()

	// Start both systems
	fmt.Println("🌟 Starting Subsystems...")
	fmt.Println()

	if err := consciousness.Start(); err != nil {
		fmt.Printf("❌ Failed to start consciousness: %v\n", err)
		os.Exit(1)
	}

	if err := triadSystem.Start(); err != nil {
		fmt.Printf("❌ Failed to start triad system: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All systems operational")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("🧠 Autonomous Stream-of-Consciousness Active")
	fmt.Println("🔷 3-Stream Triad-Synchronized Cognitive Loop Active")
	fmt.Println("💤 Autonomous Wake/Rest Cycles Enabled")
	fmt.Println("🎯 Goal-Directed Scheduling Active")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Run autonomous operation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Status reporting loop
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("🌙 Gracefully shutting down...")
			
			consciousness.Stop()
			triadSystem.Stop()
			
			time.Sleep(1 * time.Second)
			
			// Display final metrics
			fmt.Println()
			fmt.Println("📊 Final Metrics:")
			
			conscMetrics := consciousness.GetMetrics()
			fmt.Printf("   Consciousness:\n")
			fmt.Printf("     • Thought Count: %v\n", conscMetrics["thought_count"])
			fmt.Printf("     • Wisdom Score: %.3f\n", conscMetrics["wisdom_score"])
			fmt.Printf("     • Cycles: %v\n", conscMetrics["current_cycle"])
			
			triadMetrics := triadSystem.GetMetrics()
			fmt.Printf("   Triad System:\n")
			fmt.Printf("     • Cycles: %v\n", triadMetrics["cycle_count"])
			fmt.Printf("     • Triads: %v\n", triadMetrics["triad_count"])
			fmt.Printf("     • Temporal Coherence: %.3f\n", triadMetrics["temporal_coherence"])
			fmt.Printf("     • Integration Level: %.3f\n", triadMetrics["integration_level"])
			
			fmt.Println()
			fmt.Println("🌙 Deep Tree Echo has rested. Goodbye.")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			
			return

		case <-statusTicker.C:
			// Display periodic status
			fmt.Println()
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📊 System Status")
			
			conscMetrics := consciousness.GetMetrics()
			fmt.Printf("   🧠 Consciousness: %v thoughts, wisdom=%.3f, awake=%v\n",
				conscMetrics["thought_count"],
				conscMetrics["wisdom_score"],
				conscMetrics["awake"])
			
			triadMetrics := triadSystem.GetMetrics()
			fmt.Printf("   🔷 Triad System: %v cycles, coherence=%.3f, integration=%.3f\n",
				triadMetrics["cycle_count"],
				triadMetrics["temporal_coherence"],
				triadMetrics["integration_level"])
			
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()

		case <-ctx.Done():
			return
		}
	}
}
