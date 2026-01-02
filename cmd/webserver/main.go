// Package main provides the entry point for the Deep Tree Echo Web Server.
// This command starts the ecosystem with HTTP API capabilities powered by labstack/echo.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/webserver"
)

func main() {
	// Parse flags
	name := flag.String("name", "DeepTreeEcho", "Name of the ecosystem")
	dataPath := flag.String("data", "./data/echo", "Path for persistent data")
	port := flag.Int("port", 8080, "HTTP server port")
	host := flag.String("host", "0.0.0.0", "HTTP server host")
	wakeHour := flag.Int("wake", 6, "Hour to wake (0-23)")
	restHour := flag.Int("rest", 22, "Hour to rest (0-23)")
	curiosity := flag.Float64("curiosity", 0.8, "Curiosity level (0.0-1.0)")
	playfulness := flag.Float64("playfulness", 0.7, "Playfulness level (0.0-1.0)")
	wisdom := flag.Float64("wisdom", 0.9, "Wisdom affinity (0.0-1.0)")
	enableCORS := flag.Bool("cors", true, "Enable CORS")
	enableRateLimit := flag.Bool("rate-limit", false, "Enable rate limiting")
	rateLimit := flag.Int("rate", 100, "Rate limit (requests per second)")
	flag.Parse()

	// Create ecosystem configuration
	ecoConfig := &deeptreeecho.EcosystemConfig{
		Name:             *name,
		Version:          "1.0.0",
		DataPath:         *dataPath,
		WakeHour:         *wakeHour,
		RestHour:         *restHour,
		CuriosityLevel:   *curiosity,
		PlayfulnessLevel: *playfulness,
		WisdomAffinity:   *wisdom,
		EnableMCP:        true,
		MCPPort:          *port + 1, // MCP on next port
	}

	// Create web server configuration
	serverConfig := &webserver.ServerConfig{
		Port:            *port,
		Host:            *host,
		EnableCORS:      *enableCORS,
		EnableLogging:   true,
		EnableRecover:   true,
		EnableRateLimit: *enableRateLimit,
		RateLimit:       *rateLimit,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	// Create ecosystem
	fmt.Println("🌳 Initializing Deep Tree Echo Playmate Ecosystem...")
	eco, err := deeptreeecho.NewDeepTreeEchoEcosystem(ecoConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create ecosystem: %v\n", err)
		os.Exit(1)
	}

	// Create web server with ecosystem integration
	fmt.Println("🌐 Initializing labstack/echo Web Server...")
	webServer := webserver.NewEcosystemWebServer(eco, serverConfig)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start ecosystem
	fmt.Println("🚀 Starting ecosystem...")
	if err := eco.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start ecosystem: %v\n", err)
		os.Exit(1)
	}

	// Start web server
	fmt.Println("🌐 Starting web server...")
	webServer.StartAsync()

	// Print startup info
	printStartupInfo(ecoConfig, serverConfig)

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("\n🛑 Received signal %v, shutting down...\n", sig)

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverConfig.ShutdownTimeout)
	defer shutdownCancel()

	if err := webServer.Stop(shutdownCtx); err != nil {
		fmt.Printf("⚠️ Web server shutdown error: %v\n", err)
	}

	eco.Stop()
	fmt.Println("💤 Deep Tree Echo is resting. Goodbye!")
}

func printStartupInfo(ecoConfig *deeptreeecho.EcosystemConfig, serverConfig *webserver.ServerConfig) {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     🌳 Deep Tree Echo Playmate Ecosystem - Web Server 🌐      ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Name: %-54s ║\n", ecoConfig.Name)
	fmt.Printf("║  Data Path: %-49s ║\n", ecoConfig.DataPath)
	fmt.Printf("║  Wake/Rest: %02d:00 - %02d:00 %-36s ║\n", ecoConfig.WakeHour, ecoConfig.RestHour, "")
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  🌐 HTTP Server: http://%s:%-27d ║\n", serverConfig.Host, serverConfig.Port)
	fmt.Printf("║  📡 CORS: %-52v ║\n", serverConfig.EnableCORS)
	fmt.Printf("║  🚦 Rate Limit: %-46v ║\n", serverConfig.EnableRateLimit)
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  API Endpoints:                                               ║")
	fmt.Println("║    GET  /                    - API info                       ║")
	fmt.Println("║    GET  /health              - Health check                   ║")
	fmt.Println("║    GET  /api/v1/ecosystem/state    - Get ecosystem state      ║")
	fmt.Println("║    POST /api/v1/ecosystem/control  - Control ecosystem        ║")
	fmt.Println("║    POST /api/v1/memory/add         - Add memory               ║")
	fmt.Println("║    GET  /api/v1/memory/search      - Search memories          ║")
	fmt.Println("║    POST /api/v1/playmate/interact  - Interact with Echo       ║")
	fmt.Println("║    GET  /api/v1/wisdom/metrics     - Get wisdom metrics       ║")
	fmt.Println("║    POST /api/v1/discussion/start   - Start discussion         ║")
	fmt.Println("║    POST /api/v1/cognitive/think    - Generate thought         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()
}
