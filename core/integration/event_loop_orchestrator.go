package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/consciousness"
	"github.com/o9nn/echo.go/core/goals"
)

// CognitiveEventLoopOrchestrator unifies consciousness, scheduling, and goal pursuit
// into a single persistent cognitive loop with bidirectional influence.
type CognitiveEventLoopOrchestrator struct {
	mu sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	thoughtEngine *consciousness.AutonomousThoughtEngine
	goalOrch      *goals.GoalOrchestrator
	stateManager  *CognitiveStateManager

	// Cognitive state tracking
	currentFocus   string
	cognitiveLoad  float64
	fatigueLevel   float64
	awarenessLevel float64


	// Metrics
	eventsTriggered   uint64
	thoughtsTriggered uint64
	goalsGenerated    uint64
	cyclesCompleted   uint64

	running bool
}

// NewCognitiveEventLoopOrchestrator creates a new orchestrator.
func NewCognitiveEventLoopOrchestrator(
	engine *consciousness.AutonomousThoughtEngine,
	goalOrch *goals.GoalOrchestrator,
	sm *CognitiveStateManager,
) *CognitiveEventLoopOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	elo := &CognitiveEventLoopOrchestrator{
		ctx:            ctx,
		cancel:         cancel,
		thoughtEngine:  engine,
		goalOrch:       goalOrch,
		stateManager:   sm,
		cognitiveLoad:  0.3,
		fatigueLevel:   0.0,
		awarenessLevel: 0.7,
	}

	return elo
}

// Start begins the orchestrated cognitive loop.
func (elo *CognitiveEventLoopOrchestrator) Start() error {
	elo.mu.Lock()
	if elo.running {
		elo.mu.Unlock()
		return fmt.Errorf("cognitive event loop orchestrator already running")
	}
	elo.running = true
	elo.mu.Unlock()

	go elo.thoughtRoutingLoop()
	go elo.goalDrivenLoop()
	go elo.cognitiveStateLoop()
	go elo.autonomousCycleLoop()

	return nil
}

// Stop halts the orchestrated cognitive loop.
func (elo *CognitiveEventLoopOrchestrator) Stop() {
	elo.mu.Lock()
	elo.running = false
	elo.mu.Unlock()
	elo.cancel()
}

// thoughtRoutingLoop routes thoughts to the state manager.
func (elo *CognitiveEventLoopOrchestrator) thoughtRoutingLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-elo.ctx.Done():
			return
		case <-ticker.C:
			thought, err := elo.thoughtEngine.GenerateThought(elo.ctx)
			if err == nil && thought != nil {
				elo.stateManager.AddThought(
					thought.Content,
					string(thought.Type),
					"event_loop_orchestrator",
					0.7,
					[]string{"autonomous"},
				)

				elo.mu.Lock()
				elo.thoughtsTriggered++
				elo.mu.Unlock()
			}
		}
	}
}

// goalDrivenLoop generates thoughts based on active goals.
func (elo *CognitiveEventLoopOrchestrator) goalDrivenLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-elo.ctx.Done():
			return
		case <-ticker.C:
			activeGoals := elo.goalOrch.GetActiveGoals()
			for _, goal := range activeGoals {
				elo.stateManager.AddThought(
					fmt.Sprintf("Pursuing goal: %s", goal.Title),
					"goal_pursuit",
					"goal_orchestrator",
					0.8,
					[]string{"goal", goal.ID},
				)

				elo.mu.Lock()
				elo.goalsGenerated++
				elo.mu.Unlock()
			}
		}
	}
}

// cognitiveStateLoop monitors and updates cognitive state.
func (elo *CognitiveEventLoopOrchestrator) cognitiveStateLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-elo.ctx.Done():
			return
		case <-ticker.C:
			elo.updateCognitiveState()
		}
	}
}

// updateCognitiveState updates cognitive load, fatigue, and awareness.
func (elo *CognitiveEventLoopOrchestrator) updateCognitiveState() {
	elo.mu.Lock()
	defer elo.mu.Unlock()

	// Increase fatigue based on cognitive load
	elo.fatigueLevel += elo.cognitiveLoad * 0.01
	// Decrease awareness if fatigued
	if elo.fatigueLevel > 0.7 {
		elo.awarenessLevel -= 0.01
	}
	// Clamp values
	elo.fatigueLevel = clamp(elo.fatigueLevel, 0.0, 1.0)
	elo.awarenessLevel = clamp(elo.awarenessLevel, 0.0, 1.0)

	// If fatigue is high, trigger rest
	if elo.fatigueLevel > 0.8 && elo.awarenessLevel < 0.5 {
		elo.fatigueLevel = 0.0
		elo.awarenessLevel = 0.7
	}
}

// autonomousCycleLoop implements the 12-step cognitive cycle.
func (elo *CognitiveEventLoopOrchestrator) autonomousCycleLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-elo.ctx.Done():
			return
		case <-ticker.C:
			elo.executeAutonomousCycle()
		}
	}
}

// executeAutonomousCycle runs one complete cognitive cycle.
func (elo *CognitiveEventLoopOrchestrator) executeAutonomousCycle() {
	// Phase 1: Orienting (Steps 1-4) — Relevance realization
	elo.stateManager.AddThought(
		"What is most relevant to my current goals?",
		"orienting",
		"autonomous_cycle",
		0.9,
		[]string{"cycle"},
	)
	time.Sleep(1 * time.Second)

	// Phase 2: Conditioning (Steps 5-8) — Actual affordance interaction
	elo.stateManager.AddThought(
		"What have I learned that applies here?",
		"conditioning",
		"autonomous_cycle",
		0.9,
		[]string{"cycle"},
	)
	time.Sleep(1 * time.Second)

	// Phase 3: Anticipating (Steps 9-12) — Virtual salience simulation
	elo.stateManager.AddThought(
		"What possibilities should I explore?",
		"anticipating",
		"autonomous_cycle",
		0.9,
		[]string{"cycle"},
	)

	elo.mu.Lock()
	elo.cyclesCompleted++
	elo.mu.Unlock()
}

// GetMetrics returns orchestrator metrics.
func (elo *CognitiveEventLoopOrchestrator) GetMetrics() map[string]interface{} {
	elo.mu.RLock()
	defer elo.mu.RUnlock()

	return map[string]interface{}{
		"events_triggered":   elo.eventsTriggered,
		"thoughts_triggered": elo.thoughtsTriggered,
		"goals_generated":    elo.goalsGenerated,
		"cycles_completed":   elo.cyclesCompleted,
		"cognitive_load":     elo.cognitiveLoad,
		"fatigue_level":      elo.fatigueLevel,
		"awareness_level":    elo.awarenessLevel,
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
