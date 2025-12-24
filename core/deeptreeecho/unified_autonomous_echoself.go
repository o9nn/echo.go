package deeptreeecho

import (
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// UnifiedAutonomousEchoself is the top-level autonomous agent that wraps
// the UnifiedCognitiveLoopV2 and provides a high-level interface for
// autonomous operation with identity and values
type UnifiedAutonomousEchoself struct {
	mu sync.RWMutex

	// Core cognitive loop
	cognitiveLoop *UnifiedCognitiveLoopV2

	// Identity
	identity   string
	coreValues []string

	// State
	running   bool
	startTime time.Time

	// Metrics
	totalInteractions uint64
	totalThoughts     uint64
	totalDreams       uint64
}

// NewUnifiedAutonomousEchoself creates a new autonomous echoself agent
func NewUnifiedAutonomousEchoself(llmProvider llm.LLMProvider, identity string, coreValues []string) *UnifiedAutonomousEchoself {
	return &UnifiedAutonomousEchoself{
		cognitiveLoop: NewUnifiedCognitiveLoopV2(llmProvider),
		identity:      identity,
		coreValues:    coreValues,
	}
}

// Start begins autonomous operation
func (uae *UnifiedAutonomousEchoself) Start() error {
	uae.mu.Lock()
	defer uae.mu.Unlock()

	if uae.running {
		return fmt.Errorf("agent is already running")
	}

	// Start the cognitive loop
	if err := uae.cognitiveLoop.Start(); err != nil {
		return fmt.Errorf("failed to start cognitive loop: %w", err)
	}

	uae.running = true
	uae.startTime = time.Now()

	return nil
}

// Stop gracefully shuts down the agent
func (uae *UnifiedAutonomousEchoself) Stop() error {
	uae.mu.Lock()
	defer uae.mu.Unlock()

	if !uae.running {
		return fmt.Errorf("agent is not running")
	}

	// Stop the cognitive loop
	if err := uae.cognitiveLoop.Stop(); err != nil {
		return fmt.Errorf("failed to stop cognitive loop: %w", err)
	}

	uae.running = false

	return nil
}

// ProcessExternalMessage handles an external message/interaction
func (uae *UnifiedAutonomousEchoself) ProcessExternalMessage(message string) (string, error) {
	uae.mu.Lock()
	uae.totalInteractions++
	uae.mu.Unlock()

	// Process through the cognitive loop
	response, err := uae.cognitiveLoop.ProcessExternalInput(message)
	if err != nil {
		return "", fmt.Errorf("failed to process message: %w", err)
	}

	return response, nil
}

// GetCognitiveState returns the current cognitive state
func (uae *UnifiedAutonomousEchoself) GetCognitiveState() map[string]interface{} {
	uae.mu.RLock()
	defer uae.mu.RUnlock()

	state := make(map[string]interface{})

	// Get state from cognitive loop
	loopState := uae.cognitiveLoop.GetState()
	for k, v := range loopState {
		state[k] = v
	}

	// Add agent-level state
	state["identity"] = uae.identity
	state["core_values"] = uae.coreValues
	state["total_interactions"] = uae.totalInteractions
	state["total_thoughts"] = uae.totalThoughts
	state["total_dreams"] = uae.totalDreams

	if uae.running {
		state["uptime"] = time.Since(uae.startTime).String()
	} else {
		state["uptime"] = "not running"
	}

	return state
}

// GetIdentity returns the agent's identity
func (uae *UnifiedAutonomousEchoself) GetIdentity() string {
	return uae.identity
}

// GetCoreValues returns the agent's core values
func (uae *UnifiedAutonomousEchoself) GetCoreValues() []string {
	return uae.coreValues
}

// IsRunning returns whether the agent is currently running
func (uae *UnifiedAutonomousEchoself) IsRunning() bool {
	uae.mu.RLock()
	defer uae.mu.RUnlock()
	return uae.running
}

// GetUptime returns the duration since the agent started
func (uae *UnifiedAutonomousEchoself) GetUptime() time.Duration {
	uae.mu.RLock()
	defer uae.mu.RUnlock()
	if !uae.running {
		return 0
	}
	return time.Since(uae.startTime)
}
