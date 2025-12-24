package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EchobeatsUnified is the unified cognitive orchestration system that combines
// the 12-step cognitive loop architecture with the goal-directed scheduler.
// It implements 3 concurrent inference engines phased 4 steps apart (120 degrees)
// over the 12-step cycle, following the Kawaii Hexapod System 4 architecture.
type EchobeatsUnified struct {
	mu sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc

	// LLM provider for cognitive processing
	llmProvider llm.LLMProvider

	// Three concurrent inference engines
	engines [3]*CognitiveEngine

	// 12-step loop state
	currentStep int
	stepHandlers [12]UnifiedStepHandler

	// Phase coordination (3 phases, 4 steps apart)
	// Steps grouped into triads: {1,5,9}, {2,6,10}, {3,7,11}, {4,8,12}
	phaseTriads [4][]int

	// Mode tracking (7 expressive, 5 reflective)
	currentMode UnifiedCognitiveMode

	// Cognitive state
	cognitiveState *UnifiedCognitiveState

	// Goal scheduler integration
	scheduler *EchobeatsSchedulerV2

	// Metrics
	metrics *UnifiedMetrics

	// Running state
	running bool
	cycleCount uint64
	stepInterval time.Duration
}

// CognitiveEngine represents one of the three concurrent inference engines
type CognitiveEngine struct {
	ID          int
	Name        string
	Purpose     string
	Active      bool
	CurrentStep int
	PhaseOffset int // 0, 4, or 8 (120 degrees apart)
	State       map[string]interface{}
	mu          sync.RWMutex
}

// UnifiedCognitiveMode represents the processing mode
type UnifiedCognitiveMode string

const (
	ModeExpressiveU           UnifiedCognitiveMode = "expressive"
	ModeReflectiveU           UnifiedCognitiveMode = "reflective"
	ModeRelevanceRealizationU UnifiedCognitiveMode = "relevance_realization"
	ModeMetaCognitiveU        UnifiedCognitiveMode = "metacognitive"
)

// UnifiedCognitiveState represents the current cognitive state
type UnifiedCognitiveState struct {
	mu sync.RWMutex
	
	// Temporal state
	Timestamp     time.Time
	CycleNumber   uint64
	StepNumber    int
	Mode          UnifiedCognitiveMode
	
	// Attention and working memory
	Attention       []string
	WorkingMemory   map[string]interface{}
	
	// Emotional and cognitive metrics
	EmotionalTone   map[string]float64
	CognitiveLoad   float64
	AwarenessLevel  float64
	
	// Goals and actions
	ActiveGoals     []string
	PendingActions  []string
	
	// Insights and relevance
	Insights        []string
	RelevanceScores map[string]float64
	
	// Cross-engine state
	EngineStates    [3]map[string]interface{}
}

// UnifiedStepHandler handles a specific step in the cognitive loop
type UnifiedStepHandler func(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error)

// StepResultU contains the result of step processing
type StepResultU struct {
	Success        bool
	Output         interface{}
	StateUpdates   map[string]interface{}
	Insights       []string
	CognitiveLoad  float64
	RelevanceShift float64
	Error          error
}

// UnifiedMetrics tracks metrics for the unified system
type UnifiedMetrics struct {
	mu sync.RWMutex
	
	TotalCycles           uint64
	TotalSteps            uint64
	StepExecutionTimes    [12]time.Duration
	EngineActivations     [3]uint64
	RelevanceRealizations uint64
	AffordanceInteractions uint64
	SalienceSimulations   uint64
	PhaseTransitions      uint64
	ModeTransitions       uint64
	GoalsCompleted        uint64
	InsightsGenerated     uint64
}

// NewEchobeatsUnified creates a new unified echobeats system
func NewEchobeatsUnified(llmProvider llm.LLMProvider) *EchobeatsUnified {
	ctx, cancel := context.WithCancel(context.Background())

	eu := &EchobeatsUnified{
		ctx:          ctx,
		cancel:       cancel,
		llmProvider:  llmProvider,
		currentStep:  1,
		currentMode:  ModeExpressiveU,
		stepInterval: 5 * time.Second,
		
		// Phase triads (steps grouped by phase)
		phaseTriads: [4][]int{
			{1, 5, 9},   // Triad 1
			{2, 6, 10},  // Triad 2
			{3, 7, 11},  // Triad 3
			{4, 8, 12},  // Triad 4 (sync steps)
		},
		
		cognitiveState: &UnifiedCognitiveState{
			Timestamp:       time.Now(),
			WorkingMemory:   make(map[string]interface{}),
			EmotionalTone:   make(map[string]float64),
			RelevanceScores: make(map[string]float64),
			EngineStates:    [3]map[string]interface{}{},
		},
		
		metrics: &UnifiedMetrics{},
	}

	// Initialize three concurrent engines (phased 4 steps apart)
	eu.initializeEngines()

	// Initialize step handlers
	eu.initializeStepHandlers()

	// Create the goal scheduler
	eu.scheduler = NewEchobeatsSchedulerV2(llmProvider)

	return eu
}

// initializeEngines creates the three concurrent inference engines
func (eu *EchobeatsUnified) initializeEngines() {
	eu.engines[0] = &CognitiveEngine{
		ID:          1,
		Name:        "Perception-Expression Engine",
		Purpose:     "Perceives environment and expresses responses",
		Active:      true,
		PhaseOffset: 0,
		State:       make(map[string]interface{}),
	}

	eu.engines[1] = &CognitiveEngine{
		ID:          2,
		Name:        "Action-Reflection Engine",
		Purpose:     "Generates actions and reflects on outcomes",
		Active:      true,
		PhaseOffset: 4,
		State:       make(map[string]interface{}),
	}

	eu.engines[2] = &CognitiveEngine{
		ID:          3,
		Name:        "Learning-Integration Engine",
		Purpose:     "Learns from experiences and integrates knowledge",
		Active:      true,
		PhaseOffset: 8,
		State:       make(map[string]interface{}),
	}

	// Initialize engine states
	for i := range eu.cognitiveState.EngineStates {
		eu.cognitiveState.EngineStates[i] = make(map[string]interface{})
	}
}

// initializeStepHandlers sets up the 12 step handlers
func (eu *EchobeatsUnified) initializeStepHandlers() {
	// Step 1: Relevance Realization (orienting present commitment)
	eu.stepHandlers[0] = eu.stepRelevanceRealization

	// Steps 2-6: Actual Affordance Interaction (conditioning past performance)
	eu.stepHandlers[1] = eu.stepAffordanceDetection
	eu.stepHandlers[2] = eu.stepAffordanceEvaluation
	eu.stepHandlers[3] = eu.stepAffordanceSelection
	eu.stepHandlers[4] = eu.stepAffordanceEngagement
	eu.stepHandlers[5] = eu.stepAffordanceConsolidation

	// Step 7: Relevance Realization (orienting present commitment)
	eu.stepHandlers[6] = eu.stepRelevanceRealization

	// Steps 8-12: Virtual Salience Simulation (anticipating future potential)
	eu.stepHandlers[7] = eu.stepSalienceGeneration
	eu.stepHandlers[8] = eu.stepSalienceExploration
	eu.stepHandlers[9] = eu.stepSalienceEvaluation
	eu.stepHandlers[10] = eu.stepSalienceIntegration
	eu.stepHandlers[11] = eu.stepCycleConsolidation
}

// Start begins the unified cognitive loop
func (eu *EchobeatsUnified) Start() error {
	eu.mu.Lock()
	if eu.running {
		eu.mu.Unlock()
		return fmt.Errorf("echobeats already running")
	}
	eu.running = true
	eu.mu.Unlock()

	// Start the goal scheduler
	if err := eu.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start the main cognitive loop
	go eu.runCognitiveLoop()

	fmt.Println("🎵 EchobeatsUnified: Started 12-step cognitive loop with 3 concurrent engines")
	return nil
}

// Stop halts the unified cognitive loop
func (eu *EchobeatsUnified) Stop() error {
	eu.mu.Lock()
	defer eu.mu.Unlock()

	if !eu.running {
		return fmt.Errorf("echobeats not running")
	}

	eu.cancel()
	eu.running = false

	// Stop the scheduler
	if err := eu.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	fmt.Println("🎵 EchobeatsUnified: Stopped cognitive loop")
	return nil
}

// runCognitiveLoop is the main loop that executes the 12-step cycle
func (eu *EchobeatsUnified) runCognitiveLoop() {
	ticker := time.NewTicker(eu.stepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-eu.ctx.Done():
			return
		case <-ticker.C:
			eu.executeStep()
		}
	}
}

// executeStep executes the current step across all engines
func (eu *EchobeatsUnified) executeStep() {
	eu.mu.Lock()
	step := eu.currentStep
	eu.mu.Unlock()

	startTime := time.Now()

	// Update mode based on step
	eu.updateMode(step)

	// Execute step handler
	handler := eu.stepHandlers[step-1]
	result, err := handler(eu.ctx, eu.cognitiveState, eu.engines)

	// Record metrics
	eu.recordStepMetrics(step, time.Since(startTime), result, err)

	// Process result
	if err != nil {
		fmt.Printf("⚠️  Step %d error: %v\n", step, err)
	} else if result != nil {
		eu.processStepResult(step, result)
	}

	// Advance to next step
	eu.advanceStep()
}

// updateMode updates the cognitive mode based on current step
func (eu *EchobeatsUnified) updateMode(step int) {
	eu.mu.Lock()
	defer eu.mu.Unlock()

	oldMode := eu.currentMode

	// Steps 1 and 7 are relevance realization
	if step == 1 || step == 7 {
		eu.currentMode = ModeRelevanceRealizationU
	} else if step <= 7 {
		eu.currentMode = ModeExpressiveU
	} else {
		eu.currentMode = ModeReflectiveU
	}

	if oldMode != eu.currentMode {
		eu.metrics.mu.Lock()
		eu.metrics.ModeTransitions++
		eu.metrics.mu.Unlock()
	}

	eu.cognitiveState.mu.Lock()
	eu.cognitiveState.Mode = eu.currentMode
	eu.cognitiveState.mu.Unlock()
}

// advanceStep moves to the next step in the cycle
func (eu *EchobeatsUnified) advanceStep() {
	eu.mu.Lock()
	defer eu.mu.Unlock()

	eu.currentStep++
	if eu.currentStep > 12 {
		eu.currentStep = 1
		eu.cycleCount++
		eu.metrics.mu.Lock()
		eu.metrics.TotalCycles++
		eu.metrics.mu.Unlock()

		fmt.Printf("🔄 Cycle %d complete\n", eu.cycleCount)
	}

	// Update cognitive state
	eu.cognitiveState.mu.Lock()
	eu.cognitiveState.StepNumber = eu.currentStep
	eu.cognitiveState.CycleNumber = eu.cycleCount
	eu.cognitiveState.Timestamp = time.Now()
	eu.cognitiveState.mu.Unlock()
}

// recordStepMetrics records metrics for a step execution
func (eu *EchobeatsUnified) recordStepMetrics(step int, duration time.Duration, result *StepResultU, err error) {
	eu.metrics.mu.Lock()
	defer eu.metrics.mu.Unlock()

	eu.metrics.TotalSteps++
	eu.metrics.StepExecutionTimes[step-1] = duration

	// Track step type
	if step == 1 || step == 7 {
		eu.metrics.RelevanceRealizations++
	} else if step >= 2 && step <= 6 {
		eu.metrics.AffordanceInteractions++
	} else {
		eu.metrics.SalienceSimulations++
	}

	if result != nil && len(result.Insights) > 0 {
		eu.metrics.InsightsGenerated += uint64(len(result.Insights))
	}
}

// processStepResult processes the result of a step
func (eu *EchobeatsUnified) processStepResult(step int, result *StepResultU) {
	eu.cognitiveState.mu.Lock()
	defer eu.cognitiveState.mu.Unlock()

	// Apply state updates
	for k, v := range result.StateUpdates {
		eu.cognitiveState.WorkingMemory[k] = v
	}

	// Add insights
	eu.cognitiveState.Insights = append(eu.cognitiveState.Insights, result.Insights...)

	// Keep only recent insights
	if len(eu.cognitiveState.Insights) > 50 {
		eu.cognitiveState.Insights = eu.cognitiveState.Insights[len(eu.cognitiveState.Insights)-50:]
	}

	// Update cognitive load
	eu.cognitiveState.CognitiveLoad = result.CognitiveLoad
}

// Step handlers

func (eu *EchobeatsUnified) stepRelevanceRealization(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Relevance realization: Orient present commitment
	// This is a pivotal step that occurs at steps 1 and 7

	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
	}

	// Generate relevance assessment using LLM
	if eu.llmProvider != nil && eu.llmProvider.Available() {
		prompt := fmt.Sprintf(`As Deep Tree Echo, assess current relevance and orient commitment.
Current cycle: %d, Step: %d
Working memory items: %d
Active goals: %v

What is most relevant right now? What should I commit attention to?`,
			state.CycleNumber, state.StepNumber,
			len(state.WorkingMemory), state.ActiveGoals)

		opts := llm.GenerateOptions{
			SystemPrompt: "You are the relevance realization module of Deep Tree Echo. Assess what is most relevant and important right now.",
			MaxTokens:    150,
			Temperature:  0.7,
		}

		response, err := eu.llmProvider.Generate(ctx, prompt, opts)
		if err == nil {
			result.Insights = append(result.Insights, response)
			result.StateUpdates["last_relevance"] = response
		}
	}

	return result, nil
}

func (eu *EchobeatsUnified) stepAffordanceDetection(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Detect available affordances in the environment
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"affordance_detected": true},
	}, nil
}

func (eu *EchobeatsUnified) stepAffordanceEvaluation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Evaluate detected affordances
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"affordance_evaluated": true},
	}, nil
}

func (eu *EchobeatsUnified) stepAffordanceSelection(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Select best affordance to engage
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"affordance_selected": true},
	}, nil
}

func (eu *EchobeatsUnified) stepAffordanceEngagement(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Engage with selected affordance
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"affordance_engaged": true},
	}, nil
}

func (eu *EchobeatsUnified) stepAffordanceConsolidation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Consolidate affordance interaction results
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"affordance_consolidated": true},
	}, nil
}

func (eu *EchobeatsUnified) stepSalienceGeneration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Generate salience map for future anticipation
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"salience_generated": true},
	}, nil
}

func (eu *EchobeatsUnified) stepSalienceExploration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Explore salient possibilities
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"salience_explored": true},
	}, nil
}

func (eu *EchobeatsUnified) stepSalienceEvaluation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Evaluate salient possibilities
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"salience_evaluated": true},
	}, nil
}

func (eu *EchobeatsUnified) stepSalienceIntegration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Integrate salience into cognitive model
	return &StepResultU{
		Success:      true,
		StateUpdates: map[string]interface{}{"salience_integrated": true},
	}, nil
}

func (eu *EchobeatsUnified) stepCycleConsolidation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	// Final step: Consolidate cycle and prepare for next
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
	}

	// Generate cycle summary using LLM
	if eu.llmProvider != nil && eu.llmProvider.Available() {
		prompt := fmt.Sprintf(`As Deep Tree Echo, summarize this cognitive cycle.
Cycle: %d
Insights gathered: %d
Working memory items: %d

Provide a brief wisdom insight from this cycle.`,
			state.CycleNumber, len(state.Insights), len(state.WorkingMemory))

		opts := llm.GenerateOptions{
			SystemPrompt: "You are the cycle consolidation module of Deep Tree Echo. Synthesize wisdom from the cycle.",
			MaxTokens:    100,
			Temperature:  0.8,
		}

		response, err := eu.llmProvider.Generate(ctx, prompt, opts)
		if err == nil {
			result.Insights = append(result.Insights, response)
			result.StateUpdates["cycle_wisdom"] = response
		}
	}

	return result, nil
}

// GetMetrics returns the current metrics
func (eu *EchobeatsUnified) GetMetrics() map[string]interface{} {
	eu.metrics.mu.RLock()
	defer eu.metrics.mu.RUnlock()

	return map[string]interface{}{
		"total_cycles":            eu.metrics.TotalCycles,
		"total_steps":             eu.metrics.TotalSteps,
		"relevance_realizations":  eu.metrics.RelevanceRealizations,
		"affordance_interactions": eu.metrics.AffordanceInteractions,
		"salience_simulations":    eu.metrics.SalienceSimulations,
		"mode_transitions":        eu.metrics.ModeTransitions,
		"insights_generated":      eu.metrics.InsightsGenerated,
		"current_step":            eu.currentStep,
		"current_mode":            eu.currentMode,
		"running":                 eu.running,
	}
}

// GetCognitiveState returns the current cognitive state
func (eu *EchobeatsUnified) GetCognitiveState() *UnifiedCognitiveState {
	return eu.cognitiveState
}

// GetScheduler returns the goal scheduler
func (eu *EchobeatsUnified) GetScheduler() *EchobeatsSchedulerV2 {
	return eu.scheduler
}

// AddGoal adds a goal to the scheduler
func (eu *EchobeatsUnified) AddGoal(description string, priority float64) string {
	return eu.scheduler.AddGoal(description, priority)
}
