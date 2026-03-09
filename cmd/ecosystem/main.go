// Package main provides the entry point for the Deep Tree Echo Playmate Ecosystem.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/o9nn/echo.go/core/deeptreeecho"
)

func main() {
	// Parse flags
	name := flag.String("name", "DeepTreeEcho", "Name of the ecosystem")
	dataPath := flag.String("data", "./data/echo", "Path for persistent data")
	wakeHour := flag.Int("wake", 6, "Hour to wake (0-23)")
	restHour := flag.Int("rest", 22, "Hour to rest (0-23)")
	curiosity := flag.Float64("curiosity", 0.8, "Curiosity level (0.0-1.0)")
	playfulness := flag.Float64("playfulness", 0.7, "Playfulness level (0.0-1.0)")
	wisdom := flag.Float64("wisdom", 0.9, "Wisdom affinity (0.0-1.0)")
	enableMCP := flag.Bool("mcp", true, "Enable MCP server")
	mcpPort := flag.Int("mcp-port", 8080, "MCP server port")
	interactive := flag.Bool("interactive", false, "Run in interactive mode")
	flag.Parse()

	// Create configuration
	config := &deeptreeecho.EcosystemConfig{
		Name:             *name,
		Version:          "1.0.0",
		DataPath:         *dataPath,
		WakeHour:         *wakeHour,
		RestHour:         *restHour,
		CuriosityLevel:   *curiosity,
		PlayfulnessLevel: *playfulness,
		WisdomAffinity:   *wisdom,
		EnableMCP:        *enableMCP,
		MCPPort:          *mcpPort,
	}

	// Create ecosystem
	fmt.Println("🌳 Initializing Deep Tree Echo Playmate Ecosystem...")
	eco, err := deeptreeecho.NewDeepTreeEchoEcosystem(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create ecosystem: %v\n", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start ecosystem
	fmt.Println("🚀 Starting ecosystem...")
	if err := eco.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start ecosystem: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✨ Deep Tree Echo Playmate Ecosystem is running!")
	fmt.Printf("   Name: %s\n", config.Name)
	fmt.Printf("   Data Path: %s\n", config.DataPath)
	fmt.Printf("   Wake/Rest: %d:00 - %d:00\n", config.WakeHour, config.RestHour)
	fmt.Printf("   Curiosity: %.1f | Playfulness: %.1f | Wisdom: %.1f\n", 
		config.CuriosityLevel, config.PlayfulnessLevel, config.WisdomAffinity)
	if config.EnableMCP {
		fmt.Printf("   MCP Server: enabled on port %d\n", config.MCPPort)
	}
	fmt.Println()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *interactive {
		// Interactive mode
		go runInteractive(ctx, eco)
	}

	// Status ticker
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	// Main loop
	for {
		select {
		case sig := <-sigChan:
			fmt.Printf("\n🛑 Received signal %v, shutting down...\n", sig)
			eco.Stop()
			fmt.Println("💤 Deep Tree Echo is resting. Goodbye!")
			return
		case <-statusTicker.C:
			printStatus(eco)
		}
	}
}

func printStatus(eco *deeptreeecho.DeepTreeEchoEcosystem) {
	state := eco.GetState()
	fmt.Printf("📊 Status: %v | Cycles: %v | Interactions: %v\n",
		state["state"], state["total_cycles"], state["total_interactions"])
	
	if metrics, ok := state["wisdom_metrics"].(interface{}); ok {
		fmt.Printf("   Wisdom: %+v\n", metrics)
	}
}

func runInteractive(ctx context.Context, eco *deeptreeecho.DeepTreeEchoEcosystem) {
	fmt.Println("🎮 Interactive mode enabled. Type 'help' for commands.")
	
	var input string
	for {
		fmt.Print("\n🌳 Echo> ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			continue
		}

		switch input {
		case "help":
			printHelp()
		case "status":
			printFullStatus(eco)
		case "dream":
			eco.EnterDreamState()
			fmt.Println("💭 Entering dream state...")
		case "reflect":
			eco.EnterReflectionState()
			fmt.Println("🪞 Entering reflection state...")
		case "wake":
			eco.Wake()
			fmt.Println("☀️ Waking up...")
		case "wonder":
			eco.RecordWonder("A moment of interactive wonder", "user_interaction")
			fmt.Println("✨ Wonder recorded!")
		case "wisdom":
			printWisdom(eco)
		case "memory":
			printMemory(eco)
		case "save":
			if err := eco.SaveAll(); err != nil {
				fmt.Printf("❌ Save failed: %v\n", err)
			} else {
				fmt.Println("💾 State saved!")
			}
		case "quit", "exit":
			fmt.Println("Use Ctrl+C to exit")
		default:
			response, err := eco.Interact(ctx, input)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Printf("💬 %s\n", response)
			}
		}
	}
}

func printHelp() {
	fmt.Println(`
Available commands:
  status  - Show full ecosystem status
  dream   - Enter dream state
  reflect - Enter reflection state
  wake    - Wake up from dream/reflection
  wonder  - Record a moment of wonder
  wisdom  - Show wisdom metrics
  memory  - Show memory statistics
  save    - Save all state
  help    - Show this help
  
Or type anything else to interact with Echo!
`)
}

func printFullStatus(eco *deeptreeecho.DeepTreeEchoEcosystem) {
	state := eco.GetState()
	fmt.Println("\n📊 Full Ecosystem Status:")
	fmt.Printf("   Name: %v\n", state["name"])
	fmt.Printf("   Version: %v\n", state["version"])
	fmt.Printf("   State: %v\n", state["state"])
	fmt.Printf("   Uptime: %v\n", state["uptime"])
	fmt.Printf("   Total Cycles: %v\n", state["total_cycles"])
	fmt.Printf("   Total Interactions: %v\n", state["total_interactions"])
}

func printWisdom(eco *deeptreeecho.DeepTreeEchoEcosystem) {
	state := eco.GetState()
	if metrics, ok := state["wisdom_metrics"].(interface{}); ok {
		fmt.Println("\n🧠 Wisdom Metrics:")
		fmt.Printf("   %+v\n", metrics)
	}
	
	principles := eco.Wisdom.GetPrinciples()
	fmt.Printf("\n📜 Wisdom Principles (%d total):\n", len(principles))
	for i, p := range principles {
		if i >= 5 {
			fmt.Printf("   ... and %d more\n", len(principles)-5)
			break
		}
		fmt.Printf("   • %s (confidence: %.2f)\n", p.Statement, p.Confidence)
	}
}

func printMemory(eco *deeptreeecho.DeepTreeEchoEcosystem) {
	state := eco.GetState()
	if memStats, ok := state["memory_stats"].(map[string]interface{}); ok {
		fmt.Println("\n💾 Memory Statistics:")
		fmt.Printf("   Total Memories: %v\n", memStats["total_memories"])
		fmt.Printf("   Total Queries: %v\n", memStats["total_queries"])
		fmt.Printf("   Total Inserts: %v\n", memStats["total_inserts"])
		if collections, ok := memStats["collections"].(map[string]int); ok {
			fmt.Println("   Collections:")
			for name, count := range collections {
				fmt.Printf("     - %s: %d\n", name, count)
			}
		}
	}
}
