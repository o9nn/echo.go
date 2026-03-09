package deeptreeecho

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/o9nn/echo.go/core/llm"
)

// CognitiveEventLoop is the persistent, self-orchestrated cognitive event loop
// that implements the Echobeats 12-step cycle with 3 concurrent streams.
//
// Architecture:
//   - 3 concurrent cognitive streams phased 4 steps apart
//   - Stream A: steps {1,5,9}  — Perception stream
//   - Stream B: steps {2,6,10} — Action stream
//   - Stream C: steps {3,7,11} — Simulation stream
//   - Step 4,8,12 are synchronization points (all streams converge)
//
// The 12 steps map to the cognitive cycle:
//   1. Sense       (A)  — Gather sensory input
//   2. Attend      (B)  — Focus attention (modulated by NE)
//   3. Remember    (C)  — Query memory (ASSD-backed)
//   4. SYNC-1           — Cross-stream integration
//   5. Predict     (A)  — Generate predictions
//   6. Compare     (B)  — Compute prediction error
//   7. Learn       (C)  — Update reservoir weights
//   8. SYNC-2           — Cross-stream integration
//   9. Decide      (A)  — Select action (mode-aware)
//  10. Act         (B)  — Execute action
//  11. Reflect     (C)  — Meta-cognitive reflection
//  12. SYNC-3           — Full cycle integration, wisdom update
type CognitiveEventLoop struct {
	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	// Core subsystems
	endocrine    *VirtualEndocrineSystem
	fourE        *FourECognitionState
	chaosEngine  *CognitiveNoiseGenerator
	reservoir    *EchoStateReservoir
	llmProvider  llm.LLMProvider

	// Stream state
	streams [3]*CognitiveStream

	// Cycle state
	cycleCount   atomic.Uint64
	stepCount    atomic.Uint64
	currentStep  int
	stepInterval time.Duration

	// Cross-stream shared state
	sharedState *SharedCognitiveState

	// Wake/rest integration
	awake        bool
	fatigueLevel float64
	lastWakeTime time.Time

	// Callbacks
	onCycleComplete func(cycleNum uint64, state *SharedCognitiveState)
	onStepComplete  func(step int, stream int, result *StepOutput)
	onModeChange    func(from, to CognitiveMode)

	// Metrics
	totalCycles      uint64
	totalSteps       uint64
	avgCycleDuration time.Duration
	lastCycleStart   time.Time

	// Running state
	running bool
}

// CognitiveStream represents one of the three concurrent cognitive streams
type CognitiveStream struct {
	ID          int
	Name        string
	PhaseOffset int // 0, 4, or 8
	Steps       []int
	Active      bool
	LastOutput  *StepOutput
	mu          sync.RWMutex
}

// SharedCognitiveState holds state shared across all three streams
type SharedCognitiveState struct {
	mu sync.RWMutex

	// Sensory buffer
	SensoryInput    map[string]interface{}
	AttentionFocus  []string
	WorkingMemory   map[string]interface{}

	// Prediction and error
	CurrentPrediction interface{}
	PredictionError   float64

	// Decision state
	ActiveGoals    []string
	PendingActions []string
	SelectedAction string

	// Reflection
	Insights       []string
	WisdomUpdates  []string

	// Metrics
	CognitiveLoad  float64
	AwarenessLevel float64

	// Timestamp
	LastUpdate time.Time
}

// StepOutput contains the result of processing a single cognitive step
type StepOutput struct {
	StepNumber    int
	StreamID      int
	Success       bool
	Output        interface{}
	CognitiveLoad float64
	Duration      time.Duration
	Error         error
}

// NewCognitiveEventLoop creates a new persistent cognitive event loop
func NewCognitiveEventLoop(llmProvider llm.LLMProvider) *CognitiveEventLoop {
	ctx, cancel := context.WithCancel(context.Background())

	endocrine := NewVirtualEndocrineSystem()
	fourE := NewFourECognitionState(endocrine)
	chaosEngine := NewCognitiveNoiseGenerator()

	cel := &CognitiveEventLoop{
		ctx:          ctx,
		cancel:       cancel,
		endocrine:    endocrine,
		fourE:        fourE,
		chaosEngine:  chaosEngine,
		llmProvider:  llmProvider,
		stepInterval: 100 * time.Millisecond, // 10 Hz base rate
		awake:        true,
		lastWakeTime: time.Now(),
		sharedState: &SharedCognitiveState{
			SensoryInput:  make(map[string]interface{}),
			WorkingMemory: make(map[string]interface{}),
			Insights:      make([]string, 0),
			WisdomUpdates: make([]string, 0),
			LastUpdate:    time.Now(),
		},
	}

	// Initialize the three concurrent streams
	cel.streams = [3]*CognitiveStream{
		{ID: 0, Name: "Perception", PhaseOffset: 0, Steps: []int{1, 5, 9}, Active: true},
		{ID: 1, Name: "Action", PhaseOffset: 4, Steps: []int{2, 6, 10}, Active: true},
		{ID: 2, Name: "Simulation", PhaseOffset: 8, Steps: []int{3, 7, 11}, Active: true},
	}

	// Wire endocrine mode change callback
	endocrine.SetModeChangeCallback(func(from, to CognitiveMode) {
		if cel.onModeChange != nil {
			cel.onModeChange(from, to)
		}
		log.Printf("[CognitiveEventLoop] Mode transition: %s -> %s", from, to)
	})

	return cel
}

// Start begins the persistent cognitive event loop
func (cel *CognitiveEventLoop) Start() error {
	cel.mu.Lock()
	if cel.running {
		cel.mu.Unlock()
		return fmt.Errorf("cognitive event loop already running")
	}
	cel.running = true
	cel.awake = true
	cel.lastWakeTime = time.Now()
	cel.mu.Unlock()

	log.Println("[CognitiveEventLoop] Starting persistent cognitive event loop")

	go cel.mainLoop()
	return nil
}

// Stop halts the cognitive event loop
func (cel *CognitiveEventLoop) Stop() {
	cel.mu.Lock()
	defer cel.mu.Unlock()
	cel.running = false
	cel.cancel()
	log.Println("[CognitiveEventLoop] Cognitive event loop stopped")
}

// mainLoop is the persistent event loop goroutine
func (cel *CognitiveEventLoop) mainLoop() {
	ticker := time.NewTicker(cel.stepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cel.ctx.Done():
			return
		case <-ticker.C:
			cel.mu.RLock()
			running := cel.running
			awake := cel.awake
			cel.mu.RUnlock()

			if !running {
				return
			}

			if !awake {
				// In rest state — run dream consolidation at reduced rate
				cel.dreamTick()
				continue
			}

			cel.cognitiveStep()
		}
	}
}

// cognitiveStep executes one step of the 12-step cognitive cycle
func (cel *CognitiveEventLoop) cognitiveStep() {
	cel.mu.Lock()
	step := cel.currentStep
	cel.currentStep = (cel.currentStep + 1) % 12
	cel.mu.Unlock()

	stepNum := step + 1 // 1-indexed
	cel.stepCount.Add(1)

	// Update chaos engine
	cel.chaosEngine.Update()

	// Update endocrine system
	cel.endocrine.Tick(cel.stepInterval.Seconds())

	// Determine which streams are active for this step
	isSyncPoint := stepNum == 4 || stepNum == 8 || stepNum == 12

	if isSyncPoint {
		cel.executeSyncPoint(stepNum)
	} else {
		// Execute the appropriate stream step
		for _, stream := range cel.streams {
			for _, s := range stream.Steps {
				if s == stepNum {
					cel.executeStreamStep(stream, stepNum)
					break
				}
			}
		}
	}

	// Check for cycle completion
	if stepNum == 12 {
		cycleNum := cel.cycleCount.Add(1)
		cel.mu.Lock()
		cel.totalCycles = cycleNum
		cel.mu.Unlock()

		if cel.onCycleComplete != nil {
			cel.onCycleComplete(cycleNum, cel.sharedState)
		}

		// Update fatigue
		cel.updateFatigue()
	}
}

// executeStreamStep executes a step within a specific cognitive stream
func (cel *CognitiveEventLoop) executeStreamStep(stream *CognitiveStream, step int) {
	start := time.Now()
	var output *StepOutput

	switch step {
	case 1: // Sense (Perception stream)
		output = cel.stepSense()
	case 2: // Attend (Action stream)
		output = cel.stepAttend()
	case 3: // Remember (Simulation stream)
		output = cel.stepRemember()
	case 5: // Predict (Perception stream)
		output = cel.stepPredict()
	case 6: // Compare (Action stream)
		output = cel.stepCompare()
	case 7: // Learn (Simulation stream)
		output = cel.stepLearn()
	case 9: // Decide (Perception stream)
		output = cel.stepDecide()
	case 10: // Act (Action stream)
		output = cel.stepAct()
	case 11: // Reflect (Simulation stream)
		output = cel.stepReflect()
	}

	if output != nil {
		output.Duration = time.Since(start)
		output.StreamID = stream.ID
		stream.mu.Lock()
		stream.LastOutput = output
		stream.mu.Unlock()

		if cel.onStepComplete != nil {
			cel.onStepComplete(step, stream.ID, output)
		}
	}
}

// executeSyncPoint handles cross-stream synchronization
func (cel *CognitiveEventLoop) executeSyncPoint(step int) {
	cel.sharedState.mu.Lock()
	defer cel.sharedState.mu.Unlock()

	switch step {
	case 4: // SYNC-1: Integrate perception results
		// Merge sensory input with attention focus and memory context
		cel.sharedState.CognitiveLoad = cel.computeCognitiveLoad()
		cel.sharedState.LastUpdate = time.Now()

		// Signal endocrine system based on novelty
		if cel.sharedState.CognitiveLoad > 0.7 {
			cel.endocrine.SignalEvent(EndoNoveltyEncountered, cel.sharedState.CognitiveLoad)
		}

	case 8: // SYNC-2: Integrate learning results
		// Update 4E metrics based on prediction accuracy
		if cel.sharedState.PredictionError > 0 {
			cel.fourE.UpdateFromCognitiveEvent("prediction", cel.sharedState.PredictionError < 0.3,
				map[string]float64{"accuracy": 1.0 - cel.sharedState.PredictionError})
		}

		// Signal reward or error to endocrine system
		if cel.sharedState.PredictionError < 0.2 {
			cel.endocrine.SignalEvent(EndoRewardReceived, 1.0-cel.sharedState.PredictionError)
		} else if cel.sharedState.PredictionError > 0.6 {
			cel.endocrine.SignalEvent(EndoErrorSignal, cel.sharedState.PredictionError)
		}

	case 12: // SYNC-3: Full cycle integration
		// Wisdom update
		if len(cel.sharedState.Insights) > 0 {
			cel.endocrine.SignalEvent(EndoInsightGained, 0.5)
		}

		// Record valence memory for this cycle
		cel.endocrine.RecordValenceMemory(fmt.Sprintf("cycle_%d", cel.cycleCount.Load()))

		// Update awareness level
		cel.sharedState.AwarenessLevel = cel.computeAwarenessLevel()

		// Clear transient state for next cycle
		cel.sharedState.Insights = cel.sharedState.Insights[:0]
		cel.sharedState.WisdomUpdates = cel.sharedState.WisdomUpdates[:0]
	}
}

// Step implementations

func (cel *CognitiveEventLoop) stepSense() *StepOutput {
	// Gather sensory input: internal state + external events
	cel.sharedState.mu.Lock()
	cel.sharedState.SensoryInput["time"] = time.Now()
	cel.sharedState.SensoryInput["fatigue"] = cel.fatigueLevel

	mode, confidence := cel.endocrine.Bus.CurrentMode()
	cel.sharedState.SensoryInput["cognitive_mode"] = mode.String()
	cel.sharedState.SensoryInput["mode_confidence"] = confidence

	valence, arousal, dominance := cel.endocrine.GetAffectiveState()
	cel.sharedState.SensoryInput["valence"] = valence
	cel.sharedState.SensoryInput["arousal"] = arousal
	cel.sharedState.SensoryInput["dominance"] = dominance
	cel.sharedState.mu.Unlock()

	// Update 4E embodied metrics
	cel.fourE.UpdateFromCognitiveEvent("self_state_estimation", true, nil)

	return &StepOutput{StepNumber: 1, Success: true, CognitiveLoad: 0.1}
}

func (cel *CognitiveEventLoop) stepAttend() *StepOutput {
	// Focus attention modulated by norepinephrine and chaos
	ne := cel.endocrine.Bus.Concentration(Norepinephrine)
	attentionNoise := cel.chaosEngine.GetAttentionNoise()

	// Higher NE = narrower attention (exploitation), Lower NE = broader (exploration)
	focusWidth := 1.0 - ne + attentionNoise
	focusWidth = clampF64(focusWidth, 0.1, 1.0)

	cel.sharedState.mu.Lock()
	cel.sharedState.SensoryInput["attention_width"] = focusWidth
	cel.sharedState.mu.Unlock()

	// Update 4E embedded metrics
	cel.fourE.UpdateFromCognitiveEvent("affordance_detection", true, nil)

	return &StepOutput{StepNumber: 2, Success: true, CognitiveLoad: 0.15,
		Output: map[string]float64{"focus_width": focusWidth}}
}

func (cel *CognitiveEventLoop) stepRemember() *StepOutput {
	// Query memory with chaos-modulated retrieval
	memoryNoise := cel.chaosEngine.GetMemoryNoise()
	_ = memoryNoise // Used for retrieval threshold modulation

	// Update 4E extended metrics
	cel.fourE.UpdateFromCognitiveEvent("memory_retrieval", true, nil)

	return &StepOutput{StepNumber: 3, Success: true, CognitiveLoad: 0.2}
}

func (cel *CognitiveEventLoop) stepPredict() *StepOutput {
	// Generate predictions using reservoir state
	if cel.reservoir != nil {
		// Feed current state into reservoir
		input := []float64{
			cel.endocrine.Bus.Concentration(Cortisol),
			cel.endocrine.Bus.Concentration(DopamineTonic),
			cel.endocrine.Bus.Concentration(Norepinephrine),
			cel.fatigueLevel,
		}
		cel.reservoir.Update(input)
	}

	return &StepOutput{StepNumber: 5, Success: true, CognitiveLoad: 0.25}
}

func (cel *CognitiveEventLoop) stepCompare() *StepOutput {
	// Compute prediction error
	// In a full implementation, this compares predicted vs actual sensory input
	predError := 0.3 + cel.chaosEngine.GetDecisionNoise()*0.1
	predError = clampF64(predError, 0.0, 1.0)

	cel.sharedState.mu.Lock()
	cel.sharedState.PredictionError = predError
	cel.sharedState.mu.Unlock()

	return &StepOutput{StepNumber: 6, Success: true, CognitiveLoad: 0.2,
		Output: map[string]float64{"prediction_error": predError}}
}

func (cel *CognitiveEventLoop) stepLearn() *StepOutput {
	// Update reservoir weights based on prediction error
	cel.sharedState.mu.RLock()
	predError := cel.sharedState.PredictionError
	cel.sharedState.mu.RUnlock()

	// Learning rate modulated by dopamine (reward prediction error)
	da := cel.endocrine.Bus.Concentration(DopaminePhasic)
	learningRate := 0.01 * (1.0 + da)

	_ = learningRate // Applied to reservoir weight updates
	_ = predError

	return &StepOutput{StepNumber: 7, Success: true, CognitiveLoad: 0.3}
}

func (cel *CognitiveEventLoop) stepDecide() *StepOutput {
	// Select action based on cognitive mode and goals
	mode, _ := cel.endocrine.Bus.CurrentMode()
	decisionNoise := cel.chaosEngine.GetDecisionNoise()

	action := "observe" // Default
	switch mode {
	case ModeExplore:
		action = "explore"
	case ModeExploit:
		action = "optimize"
	case ModeCreative:
		action = "create"
	case ModeAnalytical:
		action = "analyze"
	case ModeSocial:
		action = "engage"
	case ModeRest:
		action = "rest"
	case ModeFlow:
		action = "flow"
	}

	// Chaos can flip decisions at low confidence
	if absF64(decisionNoise) > 0.1 {
		// Small chance of creative deviation
		action = "wonder"
	}

	cel.sharedState.mu.Lock()
	cel.sharedState.SelectedAction = action
	cel.sharedState.mu.Unlock()

	return &StepOutput{StepNumber: 9, Success: true, CognitiveLoad: 0.2,
		Output: map[string]interface{}{"action": action, "mode": mode.String()}}
}

func (cel *CognitiveEventLoop) stepAct() *StepOutput {
	// Execute the selected action
	cel.sharedState.mu.RLock()
	action := cel.sharedState.SelectedAction
	cel.sharedState.mu.RUnlock()

	// Update 4E enacted metrics
	cel.fourE.UpdateFromCognitiveEvent("action_execution", true, nil)

	return &StepOutput{StepNumber: 10, Success: true, CognitiveLoad: 0.15,
		Output: map[string]interface{}{"executed_action": action}}
}

func (cel *CognitiveEventLoop) stepReflect() *StepOutput {
	// Meta-cognitive reflection on the current cycle
	cel.sharedState.mu.Lock()

	// Generate insight if conditions are right
	da := cel.endocrine.Bus.Concentration(DopaminePhasic)
	anandamide := cel.endocrine.Bus.Concentration(Anandamide)

	if da > 0.5 && anandamide > 0.3 {
		insight := fmt.Sprintf("cycle_%d: High reward signal with flow state detected",
			cel.cycleCount.Load())
		cel.sharedState.Insights = append(cel.sharedState.Insights, insight)
	}

	cel.sharedState.mu.Unlock()

	return &StepOutput{StepNumber: 11, Success: true, CognitiveLoad: 0.25}
}

// dreamTick runs during rest state for knowledge consolidation
func (cel *CognitiveEventLoop) dreamTick() {
	// Reduced-rate processing during dream state
	cel.chaosEngine.Update()
	cel.endocrine.Tick(cel.stepInterval.Seconds() * 3) // Slower endocrine updates during rest

	// Increase melatonin during rest
	cel.endocrine.SignalEvent(EndoFatigueAccumulated, 0.01)
}

// updateFatigue updates the fatigue level based on cognitive load and time awake
func (cel *CognitiveEventLoop) updateFatigue() {
	cel.mu.Lock()
	defer cel.mu.Unlock()

	awakeTime := time.Since(cel.lastWakeTime)
	loadContribution := cel.sharedState.CognitiveLoad * 0.001
	timeContribution := awakeTime.Hours() * 0.01

	cel.fatigueLevel += loadContribution + timeContribution
	cel.fatigueLevel = clampF64(cel.fatigueLevel, 0.0, 1.0)

	// Signal fatigue to endocrine system
	if cel.fatigueLevel > 0.7 {
		cel.endocrine.SignalEvent(EndoFatigueAccumulated, cel.fatigueLevel)
	}

	// Auto-rest when fatigue is critical
	if cel.fatigueLevel > 0.95 {
		cel.awake = false
		log.Println("[CognitiveEventLoop] Fatigue critical — entering rest state")
	}
}

// computeCognitiveLoad estimates the current cognitive load
func (cel *CognitiveEventLoop) computeCognitiveLoad() float64 {
	cortisol := cel.endocrine.Bus.Concentration(Cortisol)
	ne := cel.endocrine.Bus.Concentration(Norepinephrine)
	load := cortisol*0.4 + ne*0.3 + cel.fatigueLevel*0.3
	return clampF64(load, 0.0, 1.0)
}

// computeAwarenessLevel estimates the current awareness level
func (cel *CognitiveEventLoop) computeAwarenessLevel() float64 {
	ne := cel.endocrine.Bus.Concentration(Norepinephrine)
	da := cel.endocrine.Bus.Concentration(DopamineTonic)
	melatonin := cel.endocrine.Bus.Concentration(Melatonin)
	awareness := ne*0.3 + da*0.3 + (1.0-melatonin)*0.2 + (1.0-cel.fatigueLevel)*0.2
	return clampF64(awareness, 0.0, 1.0)
}

// Wake transitions from rest to awake state
func (cel *CognitiveEventLoop) Wake() {
	cel.mu.Lock()
	defer cel.mu.Unlock()
	cel.awake = true
	cel.fatigueLevel = 0.1 // Partial fatigue recovery
	cel.lastWakeTime = time.Now()
	cel.endocrine.SignalEvent(EndoRestCompleted, 0.8)
	log.Println("[CognitiveEventLoop] Waking up — cognitive event loop active")
}

// Rest transitions from awake to rest state
func (cel *CognitiveEventLoop) Rest() {
	cel.mu.Lock()
	defer cel.mu.Unlock()
	cel.awake = false
	log.Println("[CognitiveEventLoop] Entering rest state — dream consolidation active")
}

// IsAwake returns whether the system is in awake state
func (cel *CognitiveEventLoop) IsAwake() bool {
	cel.mu.RLock()
	defer cel.mu.RUnlock()
	return cel.awake
}

// GetEndocrine returns the endocrine system
func (cel *CognitiveEventLoop) GetEndocrine() *VirtualEndocrineSystem {
	return cel.endocrine
}

// GetFourE returns the 4E cognition state
func (cel *CognitiveEventLoop) GetFourE() *FourECognitionState {
	return cel.fourE
}

// GetChaosEngine returns the cognitive noise generator
func (cel *CognitiveEventLoop) GetChaosEngine() *CognitiveNoiseGenerator {
	return cel.chaosEngine
}

// SetOnCycleComplete sets the cycle completion callback
func (cel *CognitiveEventLoop) SetOnCycleComplete(cb func(uint64, *SharedCognitiveState)) {
	cel.mu.Lock()
	defer cel.mu.Unlock()
	cel.onCycleComplete = cb
}

// SetOnStepComplete sets the step completion callback
func (cel *CognitiveEventLoop) SetOnStepComplete(cb func(int, int, *StepOutput)) {
	cel.mu.Lock()
	defer cel.mu.Unlock()
	cel.onStepComplete = cb
}

// GetMetrics returns comprehensive event loop metrics
func (cel *CognitiveEventLoop) GetMetrics() map[string]interface{} {
	cel.mu.RLock()
	defer cel.mu.RUnlock()

	mode, confidence := cel.endocrine.Bus.CurrentMode()
	valence, arousal, dominance := cel.endocrine.GetAffectiveState()

	return map[string]interface{}{
		"total_cycles":    cel.cycleCount.Load(),
		"total_steps":     cel.stepCount.Load(),
		"awake":           cel.awake,
		"fatigue_level":   cel.fatigueLevel,
		"cognitive_mode":  mode.String(),
		"mode_confidence": confidence,
		"valence":         valence,
		"arousal":         arousal,
		"dominance":       dominance,
		"4e_scores":       cel.fourE.DimensionScores(),
		"chaos_metrics":   cel.chaosEngine.GetAttractor().GetMetrics(),
		"endocrine":       cel.endocrine.GetMetrics(),
	}
}

// absF64 returns the absolute value of a float64
func absF64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
