package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("🌳 Deep Tree Echo V13 - Concurrent Inference Edition")
	fmt.Println("   Iteration 13: True 3 Concurrent Inference Engines")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	// Create V13 autonomous consciousness
	ac, err := deeptreeecho.NewAutonomousConsciousnessV13("EchoSelf")
	if err != nil {
		log.Fatalf("❌ Failed to create V13 consciousness: %v", err)
	}
	
	fmt.Println()
	
	// Start the system
	if err := ac.Start(); err != nil {
		log.Fatalf("❌ Failed to start V13 consciousness: %v", err)
	}
	
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("✅ V13 System fully operational")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println()
	
	// Monitor status
	go monitorStatusV13(ac)
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("🛑 Shutdown signal received")
	fmt.Println("=" + strings.Repeat("=", 70))
	
	// Stop the system
	if err := ac.Stop(); err != nil {
		log.Printf("⚠️  Error during shutdown: %v", err)
	}
	
	fmt.Println()
	fmt.Println("✅ V13 Test complete")
}

func monitorStatusV13(ac *deeptreeecho.AutonomousConsciousnessV13) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		status := ac.GetStatus()
		
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 70))
		fmt.Println("📊 V13 STATUS REPORT")
		fmt.Println("=" + strings.Repeat("=", 70))
		fmt.Printf("Version:            %v\n", status["version"])
		fmt.Printf("Running:            %v\n", status["running"])
		fmt.Printf("Awake:              %v\n", status["awake"])
		fmt.Printf("Uptime:             %v\n", status["uptime"])
		fmt.Println()
		fmt.Println("🔥 CONCURRENT ENGINE METRICS:")
		fmt.Printf("  Temporal Coherence:  %.3f\n", status["temporal_coherence"])
		fmt.Printf("  Integration Level:   %.3f\n", status["integration_level"])
		fmt.Printf("  Engine Coherence:    %v\n", status["engine_coherence"])
		fmt.Printf("  Engine Integration:  %v\n", status["engine_integration"])
		fmt.Printf("  Current Step:        %v\n", status["current_step"])
		fmt.Println()
		fmt.Println("🧠 WISDOM METRICS:")
		fmt.Printf("  Overall Score:       %.3f\n", status["wisdom_score"])
		fmt.Println("=" + strings.Repeat("=", 70))
		fmt.Println()
	}
}
