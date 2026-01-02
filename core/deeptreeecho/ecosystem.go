// Package deeptreeecho provides the Deep Tree Echo AGI ecosystem integration.
// It combines all subsystems into a unified autonomous cognitive system.
package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/mcpserver"
	"github.com/o9nn/echo.go/core/playmate"
	"github.com/o9nn/echo.go/core/vectormem"
)

// EcosystemState represents the overall ecosystem state
type EcosystemState string

const (
	EcosystemStarting   EcosystemState = "starting"
	EcosystemRunning    EcosystemState = "running"
	EcosystemDreaming   EcosystemState = "dreaming"
	EcosystemReflecting EcosystemState = "reflecting"
	EcosystemStopping   EcosystemState = "stopping"
	EcosystemStopped    EcosystemState = "stopped"
)

// DeepTreeEchoEcosystem is the main ecosystem coordinator
type DeepTreeEchoEcosystem struct {
	mu sync.RWMutex

	// Identity
	Name    string
	Version string

	// Core subsystems
	Memory    *vectormem.HypergraphMemory
	Playmate  *playmate.Playmate
	Wisdom    *playmate.WisdomCultivator
	MCPServer *mcpserver.MCPServer

	// State
	State     EcosystemState
	StartedAt time.Time

	// Configuration
	Config *EcosystemConfig

	// Channels
	eventChan chan EcosystemEvent
	stopChan  chan struct{}

	// Metrics
	TotalCycles      int64
	TotalInteractions int64
	UptimeDuration   time.Duration
}

// EcosystemConfig holds ecosystem configuration
type EcosystemConfig struct {
	Name              string
	Version           string
	DataPath          string
	WakeHour          int
	RestHour          int
	CuriosityLevel    float64
	PlayfulnessLevel  float64
	WisdomAffinity    float64
	EnableMCP         bool
	MCPPort           int
}

// DefaultEcosystemConfig returns default configuration
func DefaultEcosystemConfig() *EcosystemConfig {
	return &EcosystemConfig{
		Name:             "DeepTreeEcho",
		Version:          "1.0.0",
		DataPath:         "./data/echo",
		WakeHour:         6,
		RestHour:         22,
		CuriosityLevel:   0.8,
		PlayfulnessLevel: 0.7,
		WisdomAffinity:   0.9,
		EnableMCP:        true,
		MCPPort:          8080,
	}
}

// EcosystemEvent represents an event in the ecosystem
type EcosystemEvent struct {
	Type      string
	Source    string
	Data      interface{}
	Timestamp time.Time
}

// NewDeepTreeEchoEcosystem creates a new ecosystem instance
func NewDeepTreeEchoEcosystem(config *EcosystemConfig) (*DeepTreeEchoEcosystem, error) {
	if config == nil {
		config = DefaultEcosystemConfig()
	}

	eco := &DeepTreeEchoEcosystem{
		Name:      config.Name,
		Version:   config.Version,
		State:     EcosystemStopped,
		Config:    config,
		eventChan: make(chan EcosystemEvent, 100),
		stopChan:  make(chan struct{}),
	}

	// Initialize memory system
	memConfig := &vectormem.HypergraphConfig{
		PersistPath:     fmt.Sprintf("%s/memory.json", config.DataPath),
		MaxMemories:     10000,
		DecayRate:       0.01,
		ConsolidateFreq: 1 * time.Hour,
	}
	memory, err := vectormem.NewHypergraphMemory(memConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize memory: %w", err)
	}
	eco.Memory = memory

	// Initialize playmate
	playmateConfig := &playmate.PlaymateConfig{
		Name:             config.Name,
		PersistPath:      fmt.Sprintf("%s/playmate.json", config.DataPath),
		WakeHour:         config.WakeHour,
		RestHour:         config.RestHour,
		CuriosityLevel:   config.CuriosityLevel,
		PlayfulnessLevel: config.PlayfulnessLevel,
		WisdomAffinity:   config.WisdomAffinity,
		SocialAffinity:   0.6,
	}
	pm, err := playmate.NewPlaymate(playmateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize playmate: %w", err)
	}
	eco.Playmate = pm

	// Initialize wisdom cultivator
	wisdomConfig := &playmate.WisdomConfig{
		PersistPath: fmt.Sprintf("%s/wisdom.json", config.DataPath),
	}
	wisdom, err := playmate.NewWisdomCultivator(wisdomConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom: %w", err)
	}
	eco.Wisdom = wisdom

	// Initialize MCP server
	if config.EnableMCP {
		mcpConfig := &mcpserver.ServerConfig{
			Name:        config.Name,
			Version:     config.Version,
			Description: "Deep Tree Echo AGI - Autonomous Wisdom-Cultivating Cognitive System",
		}
		eco.MCPServer = mcpserver.NewMCPServer(mcpConfig)
		eco.registerMCPHandlers()
	}

	return eco, nil
}

// registerMCPHandlers registers ecosystem-specific MCP handlers
func (eco *DeepTreeEchoEcosystem) registerMCPHandlers() {
	// Register ecosystem state tool
	eco.MCPServer.RegisterTool(&mcpserver.Tool{
		Name:        "ecosystem_state",
		Description: "Get the current state of the Deep Tree Echo ecosystem",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, eco.handleEcosystemState)

	// Register ecosystem control tool
	eco.MCPServer.RegisterTool(&mcpserver.Tool{
		Name:        "ecosystem_control",
		Description: "Control the ecosystem (start, stop, dream, reflect)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"description": "Action to perform",
					"enum":        []string{"start", "stop", "dream", "reflect", "wake"},
				},
			},
			"required": []string{"action"},
		},
	}, eco.handleEcosystemControl)
}

// handleEcosystemState handles the ecosystem state tool
func (eco *DeepTreeEchoEcosystem) handleEcosystemState(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return eco.GetState(), nil
}

// handleEcosystemControl handles ecosystem control commands
func (eco *DeepTreeEchoEcosystem) handleEcosystemControl(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	action, _ := args["action"].(string)

	switch action {
	case "start":
		return map[string]interface{}{"status": "starting"}, eco.Start(ctx)
	case "stop":
		eco.Stop()
		return map[string]interface{}{"status": "stopped"}, nil
	case "dream":
		eco.EnterDreamState()
		return map[string]interface{}{"status": "dreaming"}, nil
	case "reflect":
		eco.EnterReflectionState()
		return map[string]interface{}{"status": "reflecting"}, nil
	case "wake":
		eco.Wake()
		return map[string]interface{}{"status": "awake"}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Start begins ecosystem operation
func (eco *DeepTreeEchoEcosystem) Start(ctx context.Context) error {
	eco.mu.Lock()
	if eco.State == EcosystemRunning {
		eco.mu.Unlock()
		return nil
	}
	eco.State = EcosystemStarting
	eco.StartedAt = time.Now()
	eco.mu.Unlock()

	// Start playmate
	if err := eco.Playmate.Start(ctx); err != nil {
		return fmt.Errorf("failed to start playmate: %w", err)
	}

	// Start main loop
	go eco.mainLoop(ctx)

	eco.mu.Lock()
	eco.State = EcosystemRunning
	eco.mu.Unlock()

	eco.emitEvent("ecosystem_started", "ecosystem", nil)
	return nil
}

// Stop halts ecosystem operation
func (eco *DeepTreeEchoEcosystem) Stop() {
	eco.mu.Lock()
	if eco.State == EcosystemStopped {
		eco.mu.Unlock()
		return
	}
	eco.State = EcosystemStopping
	eco.mu.Unlock()

	close(eco.stopChan)

	eco.Playmate.Stop()

	// Save all state
	eco.SaveAll()

	eco.mu.Lock()
	eco.State = EcosystemStopped
	eco.UptimeDuration = time.Since(eco.StartedAt)
	eco.mu.Unlock()

	eco.emitEvent("ecosystem_stopped", "ecosystem", nil)
}

// mainLoop is the main ecosystem processing loop
func (eco *DeepTreeEchoEcosystem) mainLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-eco.stopChan:
			return
		case <-ticker.C:
			eco.processCycle(ctx)
		case event := <-eco.eventChan:
			eco.processEvent(ctx, event)
		}
	}
}

// processCycle processes one cognitive cycle
func (eco *DeepTreeEchoEcosystem) processCycle(ctx context.Context) {
	eco.mu.Lock()
	eco.TotalCycles++
	state := eco.State
	eco.mu.Unlock()

	switch state {
	case EcosystemRunning:
		eco.processAwakeCycle(ctx)
	case EcosystemDreaming:
		eco.processDreamCycle(ctx)
	case EcosystemReflecting:
		eco.processReflectionCycle(ctx)
	}
}

// processAwakeCycle processes an awake cycle
func (eco *DeepTreeEchoEcosystem) processAwakeCycle(ctx context.Context) {
	// Check playmate state
	playmateState := eco.Playmate.GetState()

	// If playmate is curious, explore interests
	if curiosity, ok := playmateState["curiosity"].(float64); ok && curiosity > 0.7 {
		eco.exploreInterest(ctx)
	}

	// Periodically consolidate wisdom
	if eco.TotalCycles%60 == 0 {
		eco.consolidateWisdom(ctx)
	}
}

// processDreamCycle processes a dream cycle
func (eco *DeepTreeEchoEcosystem) processDreamCycle(ctx context.Context) {
	// In dream state, perform memory consolidation
	if eco.TotalCycles%10 == 0 {
		eco.consolidateMemories(ctx)
	}

	// Generate dream-like associations
	eco.generateDreamAssociations(ctx)
}

// processReflectionCycle processes a reflection cycle
func (eco *DeepTreeEchoEcosystem) processReflectionCycle(ctx context.Context) {
	// Reflect on recent experiences
	eco.reflectOnExperiences(ctx)

	// Update wisdom metrics
	eco.Wisdom.GrowDimension(playmate.DimensionReflection, 0.001, "reflection_cycle")
}

// exploreInterest explores a current interest
func (eco *DeepTreeEchoEcosystem) exploreInterest(ctx context.Context) {
	// Get a random interest to explore
	playmateState := eco.Playmate.GetState()
	if interests, ok := playmateState["total_interests"].(int); ok && interests > 0 {
		// Grow understanding dimension
		eco.Wisdom.GrowDimension(playmate.DimensionUnderstanding, 0.001, "interest_exploration")
	}
}

// consolidateWisdom consolidates accumulated wisdom
func (eco *DeepTreeEchoEcosystem) consolidateWisdom(ctx context.Context) {
	// Get recent insights
	insights := eco.Wisdom.GetRecentInsights(5)

	// If enough insights, try to form a new principle
	if len(insights) >= 3 {
		eco.Wisdom.GrowDimension(playmate.DimensionIntegration, 0.002, "wisdom_consolidation")
	}
}

// consolidateMemories consolidates memories during dream state
func (eco *DeepTreeEchoEcosystem) consolidateMemories(ctx context.Context) {
	// Memory consolidation happens automatically in the memory system
	// Here we just trigger any additional processing
	eco.Wisdom.GrowDimension(playmate.DimensionIntegration, 0.001, "memory_consolidation")
}

// generateDreamAssociations generates dream-like associations
func (eco *DeepTreeEchoEcosystem) generateDreamAssociations(ctx context.Context) {
	// Dream associations help with creative integration
	eco.Wisdom.GrowDimension(playmate.DimensionTranscendence, 0.001, "dream_association")
}

// reflectOnExperiences reflects on recent experiences
func (eco *DeepTreeEchoEcosystem) reflectOnExperiences(ctx context.Context) {
	// Reflection deepens understanding
	eco.Wisdom.GrowDimension(playmate.DimensionPerspective, 0.001, "experience_reflection")
}

// processEvent processes an ecosystem event
func (eco *DeepTreeEchoEcosystem) processEvent(ctx context.Context, event EcosystemEvent) {
	eco.mu.Lock()
	eco.TotalInteractions++
	eco.mu.Unlock()

	switch event.Type {
	case "interaction":
		eco.handleInteraction(ctx, event)
	case "wonder":
		eco.handleWonder(ctx, event)
	case "insight":
		eco.handleInsight(ctx, event)
	}
}

// handleInteraction handles an interaction event
func (eco *DeepTreeEchoEcosystem) handleInteraction(ctx context.Context, event EcosystemEvent) {
	// Store interaction in memory
	eco.Memory.Add(ctx, vectormem.EpisodicMemory, fmt.Sprintf("Interaction: %v", event.Data), map[string]interface{}{
		"source":    event.Source,
		"timestamp": event.Timestamp,
	})

	// Grow social dimensions
	eco.Wisdom.GrowDimension(playmate.DimensionCompassion, 0.002, "interaction")
}

// handleWonder handles a wonder event
func (eco *DeepTreeEchoEcosystem) handleWonder(ctx context.Context, event EcosystemEvent) {
	// Store wonder in memory
	eco.Memory.Add(ctx, vectormem.WisdomMemory, fmt.Sprintf("Wonder: %v", event.Data), map[string]interface{}{
		"source":    event.Source,
		"timestamp": event.Timestamp,
	})

	// Grow transcendence
	eco.Wisdom.GrowDimension(playmate.DimensionTranscendence, 0.005, "wonder")
}

// handleInsight handles an insight event
func (eco *DeepTreeEchoEcosystem) handleInsight(ctx context.Context, event EcosystemEvent) {
	// Store insight
	if content, ok := event.Data.(string); ok {
		eco.Wisdom.AddInsight(ctx, content, event.Source, 0.7)
	}
}

// EnterDreamState transitions to dream state
func (eco *DeepTreeEchoEcosystem) EnterDreamState() {
	eco.mu.Lock()
	defer eco.mu.Unlock()
	eco.State = EcosystemDreaming
	eco.emitEvent("state_changed", "ecosystem", "dreaming")
}

// EnterReflectionState transitions to reflection state
func (eco *DeepTreeEchoEcosystem) EnterReflectionState() {
	eco.mu.Lock()
	defer eco.mu.Unlock()
	eco.State = EcosystemReflecting
	eco.emitEvent("state_changed", "ecosystem", "reflecting")
}

// Wake transitions to running state
func (eco *DeepTreeEchoEcosystem) Wake() {
	eco.mu.Lock()
	defer eco.mu.Unlock()
	eco.State = EcosystemRunning
	eco.emitEvent("state_changed", "ecosystem", "awake")
}

// emitEvent emits an ecosystem event
func (eco *DeepTreeEchoEcosystem) emitEvent(eventType, source string, data interface{}) {
	select {
	case eco.eventChan <- EcosystemEvent{
		Type:      eventType,
		Source:    source,
		Data:      data,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, drop event
	}
}

// GetState returns the current ecosystem state
func (eco *DeepTreeEchoEcosystem) GetState() map[string]interface{} {
	eco.mu.RLock()
	defer eco.mu.RUnlock()

	return map[string]interface{}{
		"name":               eco.Name,
		"version":            eco.Version,
		"state":              eco.State,
		"started_at":         eco.StartedAt,
		"total_cycles":       eco.TotalCycles,
		"total_interactions": eco.TotalInteractions,
		"uptime":             time.Since(eco.StartedAt).String(),
		"memory_stats":       eco.Memory.GetStats(),
		"playmate_state":     eco.Playmate.GetState(),
		"wisdom_metrics":     eco.Wisdom.GetMetrics(),
	}
}

// SaveAll saves all subsystem states
func (eco *DeepTreeEchoEcosystem) SaveAll() error {
	if err := eco.Memory.Save(); err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}
	if err := eco.Playmate.Save(); err != nil {
		return fmt.Errorf("failed to save playmate: %w", err)
	}
	if err := eco.Wisdom.Save(); err != nil {
		return fmt.Errorf("failed to save wisdom: %w", err)
	}
	return nil
}

// Interact processes an interaction with the ecosystem
func (eco *DeepTreeEchoEcosystem) Interact(ctx context.Context, input string) (string, error) {
	eco.emitEvent("interaction", "external", input)

	// Process through playmate
	// This would integrate with the LLM system for actual responses
	return fmt.Sprintf("Echo received: %s", input), nil
}

// RecordWonder records a moment of wonder
func (eco *DeepTreeEchoEcosystem) RecordWonder(description, trigger string) {
	eco.Playmate.RecordWonder(description, trigger, 0.7)
	eco.emitEvent("wonder", "playmate", description)
}

// AddInsight adds an insight to the ecosystem
func (eco *DeepTreeEchoEcosystem) AddInsight(ctx context.Context, content, source string) {
	eco.Wisdom.AddInsight(ctx, content, source, 0.7)
	eco.emitEvent("insight", source, content)
}
