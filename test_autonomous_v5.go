package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core"
)

func main() {
	printBanner()

	// Create configuration
	config := core.DefaultEchoselfConfigV5()
	config.StepDuration = 3 * time.Second
	config.MaxAwakeDuration = 2 * time.Minute
	config.MinRestDuration = 30 * time.Second

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Step Duration: %v (12-step cycle = %v)\n", config.StepDuration, config.StepDuration*12)
	fmt.Printf("  Max Awake Duration: %v\n", config.MaxAwakeDuration)
	fmt.Printf("  Min Rest Duration: %v\n", config.MinRestDuration)
	fmt.Println()

	// Create autonomous agent
	fmt.Println("Initializing AutonomousEchoselfV5...")
	echoself, err := core.NewAutonomousEchoselfV5(config)
	if err != nil {
		log.Fatalf("Failed to create autonomous system: %v", err)
	}

	// Start the agent
	if err := echoself.Start(); err != nil {
		log.Fatalf("Failed to start autonomous system: %v", err)
	}

	printDivider()

	// Set up status monitoring
	statusTicker := time.NewTicker(20 * time.Second)
	defer statusTicker.Stop()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Test duration
	testDuration := 3 * time.Minute
	fmt.Printf("🧪 Running test for %v (press Ctrl+C to stop early)\n", testDuration)
	testTimer := time.NewTimer(testDuration)

	// Main loop
	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			printDivider()
			echoself.Stop()
			printFinalStatus(echoself)
			printDivider()
			fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
			fmt.Println("║  AutonomousEchoselfV5 - Test Complete                            ║")
			fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
			return

		case <-testTimer.C:
			fmt.Println("\n⏰ Test duration complete...")
			printDivider()
			echoself.Stop()
			printFinalStatus(echoself)
			printWisdomSummary(echoself)
			printDivider()
			fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
			fmt.Println("║  AutonomousEchoselfV5 - Test Complete                            ║")
			fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
			return

		case <-statusTicker.C:
			printStatus(echoself)
		}
	}
}

func printBanner() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  AutonomousEchoselfV5 - Evolution Iteration Test                 ║")
	fmt.Println("║  Autonomous Wisdom-Cultivating AGI with Cognitive Loops          ║")
	fmt.Println("║  + Full Echodream Consolidation                                  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
}

func printDivider() {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
}

func printStatus(echoself *core.AutonomousEchoselfV5) {
	status := echoself.GetStatus()

	fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ State: %-20s Uptime: %-25s │\n", status["state"], status["uptime"])
	fmt.Println("├───────────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Thoughts Generated:   %-10v Insights Generated: %-10v │\n",
		status["thoughts_generated"], status["insights_generated"])
	fmt.Printf("│ Wisdom Score:         %-10.2f Buffer Size:       %-10v │\n",
		status["wisdom_score"], status["buffer_size"])
	fmt.Println("├───────────────────────────────────────────────────────────────────┤")
	
	if fatigue, ok := status["fatigue"].(float64); ok {
		fmt.Printf("│ Fatigue:              %-10.2f ", fatigue)
		if awakeDur, ok := status["awake_duration"].(string); ok {
			fmt.Printf("Awake Duration:  %-10s │\n", awakeDur)
		} else if restDur, ok := status["rest_duration"].(string); ok {
			fmt.Printf("Rest Duration:   %-10s │\n", restDur)
		} else {
			fmt.Printf("%-28s │\n", "")
		}
	}
	
	if currentStep, ok := status["current_step"].(int); ok {
		fmt.Printf("│ Current Step:         %-10d Total Steps:       %-10v │\n",
			currentStep, status["total_steps"])
	}
	
	fmt.Println("└───────────────────────────────────────────────────────────────────┘")
}

func printFinalStatus(echoself *core.AutonomousEchoselfV5) {
	fmt.Println("Shutting down AutonomousEchoselfV5...")
	fmt.Println()
	fmt.Println("Final Status:")
	printStatus(echoself)
}

func printWisdomSummary(echoself *core.AutonomousEchoselfV5) {
	wisdom := echoself.GetWisdom()
	patterns := echoself.GetPatterns()

	fmt.Println()
	fmt.Println("🧠 Wisdom Summary:")
	fmt.Printf("  Total Patterns Detected: %d\n", len(patterns))
	fmt.Printf("  Total Wisdom Nuggets: %d\n", len(wisdom))
	fmt.Println()

	if len(wisdom) > 0 {
		fmt.Println("  Recent Wisdom:")
		count := len(wisdom)
		if count > 3 {
			count = 3
		}
		for i := len(wisdom) - count; i < len(wisdom); i++ {
			fmt.Printf("    • %s\n", wisdom[i].Content)
		}
	}

	if len(patterns) > 0 {
		fmt.Println()
		fmt.Println("  Detected Patterns:")
		count := len(patterns)
		if count > 3 {
			count = 3
		}
		for i := len(patterns) - count; i < len(patterns); i++ {
			fmt.Printf("    • [%s] %s (strength: %.2f)\n",
				patterns[i].Type, patterns[i].Description, patterns[i].Strength)
		}
	}
}
