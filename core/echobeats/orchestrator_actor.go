package echobeats

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tochemey/goakt/v2/actors"
)

// OrchestratorActor coordinates the 12-step cognitive loop
// It manages the three engine actors and ensures proper synchronization
type OrchestratorActor struct {
	mu          sync.RWMutex
	sharedState *SharedCognitiveState
	config      *GoAktConfig

	// Engine actor references
	affordanceActor actors.PID
	relevanceActor  actors.PID
	salienceActor   actors.PID

	// Loop state
	currentStep    atomic.Int32
	cycleCount     atomic.Uint64
	running        atomic.Bool
	lastCycleStart time.Time

	// Synchronization
	pivotalBarrier sync.WaitGroup
	syncResponses  map[string]bool

	// Step-to-engine mapping
	stepEngineMap map[int][]actors.PID

	// Metrics
	cycleMetrics *CycleMetrics
}

// CycleMetrics tracks performance of cognitive cycles
type CycleMetrics struct {
	mu                sync.RWMutex
	TotalCycles       uint64
	AverageCycleTime  time.Duration
	LastCycleTime     time.Duration
	StepTimes         map[int]time.Duration
	EngineMetrics     map[string]*EngineMetric
}

// EngineMetric tracks individual engine performance
type EngineMetric struct {
	StepsProcessed int64
	TotalTime      time.Duration
	Errors         int64
	AverageTime    time.Duration
}

// NewOrchestratorActor creates a new orchestrator actor
func NewOrchestratorActor(
	affordanceActor, relevanceActor, salienceActor actors.PID,
	sharedState *SharedCognitiveState,
	config *GoAktConfig,
) *OrchestratorActor {
	actor := &OrchestratorActor{
		sharedState:     sharedState,
		config:          config,
		affordanceActor: affordanceActor,
		relevanceActor:  relevanceActor,
		salienceActor:   salienceActor,
		syncResponses:   make(map[string]bool),
		stepEngineMap:   make(map[int][]actors.PID),
		cycleMetrics:    newCycleMetrics(),
	}

	// Map steps to engines based on the 12-step cognitive architecture
	// Steps 0-5: Affordance engine (with step 0 also handled by relevance)
	// Steps 6-11: Salience engine (with step 6 also handled by relevance)
	// Steps 0 and 6: Pivotal steps handled by relevance engine

	// Pivotal steps (all engines synchronize)
	actor.stepEngineMap[0] = []actors.PID{relevanceActor, affordanceActor, salienceActor}
	actor.stepEngineMap[6] = []actors.PID{relevanceActor, affordanceActor, salienceActor}

	// Affordance steps (1-5)
	for i := 1; i <= 5; i++ {
		actor.stepEngineMap[i] = []actors.PID{affordanceActor}
	}

	// Salience steps (7-11)
	for i := 7; i <= 11; i++ {
		actor.stepEngineMap[i] = []actors.PID{salienceActor}
	}

	return actor
}

func newCycleMetrics() *CycleMetrics {
	return &CycleMetrics{
		StepTimes:     make(map[int]time.Duration),
		EngineMetrics: make(map[string]*EngineMetric),
	}
}

// PreStart is called before the actor starts
func (o *OrchestratorActor) PreStart(ctx context.Context) error {
	return nil
}

// Receive handles incoming messages
func (o *OrchestratorActor) Receive(ctx actors.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *StartCycleMsg:
		o.handleStartCycle(ctx, msg)
	case *StepResultMsg:
		o.handleStepResult(ctx, msg)
	case *SyncAckMsg:
		o.handleSyncAck(ctx, msg)
	case *GetStatusMsg:
		o.handleGetStatus(ctx)
	default:
		ctx.Unhandled()
	}
}

// PostStop is called after the actor stops
func (o *OrchestratorActor) PostStop(ctx context.Context) error {
	return nil
}

// handleStartCycle initiates a new cognitive cycle
func (o *OrchestratorActor) handleStartCycle(ctx actors.ReceiveContext, msg *StartCycleMsg) {
	if o.running.Load() {
		// Already running a cycle
		return
	}

	o.running.Store(true)
	o.lastCycleStart = time.Now()
	o.currentStep.Store(0)

	// Start the cycle by processing step 0
	go o.runCycle(ctx)
}

// runCycle executes the full 12-step cognitive loop
func (o *OrchestratorActor) runCycle(ctx actors.ReceiveContext) {
	defer func() {
		o.running.Store(false)
		o.cycleCount.Add(1)
		
		// Record cycle metrics
		o.cycleMetrics.mu.Lock()
		o.cycleMetrics.TotalCycles++
		o.cycleMetrics.LastCycleTime = time.Since(o.lastCycleStart)
		o.cycleMetrics.mu.Unlock()
	}()

	// Execute all 12 steps
	for step := 0; step < 12; step++ {
		o.currentStep.Store(int32(step))
		o.sharedState.SetCurrentStep(step)

		stepStart := time.Now()

		// Check if this is a pivotal step
		isPivotal := step == 0 || step == 6
		if isPivotal {
			o.executePivotalStep(ctx, step)
		} else {
			o.executeStep(ctx, step)
		}

		// Record step time
		o.cycleMetrics.mu.Lock()
		o.cycleMetrics.StepTimes[step] = time.Since(stepStart)
		o.cycleMetrics.mu.Unlock()

		// Wait for step duration
		time.Sleep(o.config.StepDuration)
	}
}

// executePivotalStep handles pivotal steps with synchronization
func (o *OrchestratorActor) executePivotalStep(ctx actors.ReceiveContext, step int) {
	o.mu.Lock()
	o.syncResponses = make(map[string]bool)
	o.mu.Unlock()

	// Send pivotal sync to all engines
	syncMsg := &PivotalSyncMsg{
		PivotalStep: step,
		EngineID:    "orchestrator",
		Timestamp:   time.Now(),
	}

	engines := o.stepEngineMap[step]
	for _, engine := range engines {
		ctx.Tell(engine, syncMsg)
	}

	// Wait for all engines to acknowledge (with timeout)
	timeout := time.After(o.config.PivotalTimeout)
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Timeout - proceed anyway
			return
		case <-ticker.C:
			o.mu.RLock()
			allReady := len(o.syncResponses) >= len(engines)
			o.mu.RUnlock()
			if allReady {
				// All engines synchronized
				o.executeStep(ctx, step)
				return
			}
		}
	}
}

// executeStep sends a step message to the appropriate engines
func (o *OrchestratorActor) executeStep(ctx actors.ReceiveContext, step int) {
	engines, ok := o.stepEngineMap[step]
	if !ok {
		return
	}

	stepMsg := &StepMsg{
		StepNumber: step,
		Timestamp:  time.Now(),
		Payload:    nil, // In production, include relevant context
	}

	for _, engine := range engines {
		ctx.Tell(engine, stepMsg)
	}
}

// handleStepResult processes results from engine actors
func (o *OrchestratorActor) handleStepResult(ctx actors.ReceiveContext, msg *StepResultMsg) {
	// Update metrics
	o.cycleMetrics.mu.Lock()
	defer o.cycleMetrics.mu.Unlock()

	metric, ok := o.cycleMetrics.EngineMetrics[msg.EngineID]
	if !ok {
		metric = &EngineMetric{}
		o.cycleMetrics.EngineMetrics[msg.EngineID] = metric
	}

	metric.StepsProcessed++
	metric.TotalTime += msg.ProcessingTime
	if !msg.Success {
		metric.Errors++
	}
	metric.AverageTime = metric.TotalTime / time.Duration(metric.StepsProcessed)
}

// handleSyncAck processes synchronization acknowledgments
func (o *OrchestratorActor) handleSyncAck(ctx actors.ReceiveContext, msg *SyncAckMsg) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.syncResponses[msg.EngineID] = msg.Ready
}

// handleGetStatus returns current orchestrator status
func (o *OrchestratorActor) handleGetStatus(ctx actors.ReceiveContext) {
	status := &LoopStatusMsg{
		Running:     o.running.Load(),
		CycleCount:  o.cycleCount.Load(),
		CurrentStep: int(o.currentStep.Load()),
		Metrics:     o.cycleMetrics,
	}
	ctx.Response(status)
}

// GetStatusMsg requests current status
type GetStatusMsg struct{}

// LoopStatusMsg contains current loop status
type LoopStatusMsg struct {
	Running     bool
	CycleCount  uint64
	CurrentStep int
	Metrics     *CycleMetrics
}

// Triad step groupings for the 12-step cognitive architecture
// These represent the concurrent processing patterns across the three engines

// TriadSteps defines the four triads that occur every 4 steps
var TriadSteps = map[string][]int{
	"pivotal_relevance_1": {1, 5, 9},  // First pivotal relevance triad
	"affordance_action":   {2, 6, 10}, // Affordance-action triad
	"salience_simulation": {3, 7, 11}, // Salience-simulation triad
	"meta_reflection":     {4, 8, 12}, // Meta-cognitive reflection triad (12 wraps to 0)
}

// GetTriadForStep returns which triad a step belongs to
func GetTriadForStep(step int) string {
	normalizedStep := step % 12
	for triad, steps := range TriadSteps {
		for _, s := range steps {
			if s == normalizedStep || (normalizedStep == 0 && s == 12) {
				return triad
			}
		}
	}
	return "unknown"
}

// PhaseForStep returns the phase (affordance/relevance/salience) for a step
func PhaseForStep(step int) string {
	normalizedStep := step % 12
	switch {
	case normalizedStep == 0 || normalizedStep == 6:
		return "relevance" // Pivotal steps
	case normalizedStep >= 1 && normalizedStep <= 5:
		return "affordance"
	case normalizedStep >= 7 && normalizedStep <= 11:
		return "salience"
	default:
		return "unknown"
	}
}
