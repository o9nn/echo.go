package echobeats

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tochemey/goakt/v2/actors"
	"github.com/tochemey/goakt/v2/goakt"
	"github.com/tochemey/goakt/v2/log"
)

// GoAktCognitiveSystem implements the 3 concurrent inference engines using goakt actors
// This provides improved concurrency, distribution, and fault tolerance
type GoAktCognitiveSystem struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	running         atomic.Bool

	// Actor system
	actorSystem     goakt.ActorSystem
	
	// Engine actors
	affordanceActor actors.PID
	relevanceActor  actors.PID
	salienceActor   actors.PID
	
	// Orchestrator actor
	orchestratorActor actors.PID

	// Shared state
	sharedState     *SharedCognitiveState
	
	// Configuration
	config          *GoAktConfig

	// Metrics
	cycleCount      atomic.Uint64
	lastCycleTime   atomic.Int64
}

// GoAktConfig holds configuration for the actor system
type GoAktConfig struct {
	SystemName       string
	StepDuration     time.Duration
	PivotalTimeout   time.Duration
	MaxRetries       int
	EnableTelemetry  bool
}

// DefaultGoAktConfig returns default configuration
func DefaultGoAktConfig() *GoAktConfig {
	return &GoAktConfig{
		SystemName:      "deep-tree-echo",
		StepDuration:    time.Millisecond * 100,
		PivotalTimeout:  time.Second * 5,
		MaxRetries:      3,
		EnableTelemetry: true,
	}
}

// NewGoAktCognitiveSystem creates a new actor-based cognitive system
func NewGoAktCognitiveSystem(config *GoAktConfig) (*GoAktCognitiveSystem, error) {
	if config == nil {
		config = DefaultGoAktConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	system := &GoAktCognitiveSystem{
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
		sharedState: NewSharedCognitiveState(),
	}

	return system, nil
}

// Start initializes and starts the actor system
func (gcs *GoAktCognitiveSystem) Start() error {
	gcs.mu.Lock()
	defer gcs.mu.Unlock()

	if gcs.running.Load() {
		return fmt.Errorf("cognitive system already running")
	}

	// Create actor system
	actorSystem, err := goakt.NewActorSystem(
		gcs.config.SystemName,
		goakt.WithLogger(log.DefaultLogger),
		goakt.WithActorInitMaxRetries(gcs.config.MaxRetries),
	)
	if err != nil {
		return fmt.Errorf("failed to create actor system: %w", err)
	}

	gcs.actorSystem = actorSystem

	// Start the actor system
	if err := actorSystem.Start(gcs.ctx); err != nil {
		return fmt.Errorf("failed to start actor system: %w", err)
	}

	// Spawn engine actors
	if err := gcs.spawnEngineActors(); err != nil {
		actorSystem.Stop(gcs.ctx)
		return fmt.Errorf("failed to spawn engine actors: %w", err)
	}

	gcs.running.Store(true)
	return nil
}

// Stop gracefully shuts down the actor system
func (gcs *GoAktCognitiveSystem) Stop() error {
	gcs.mu.Lock()
	defer gcs.mu.Unlock()

	if !gcs.running.Load() {
		return nil
	}

	gcs.running.Store(false)
	gcs.cancel()

	if gcs.actorSystem != nil {
		return gcs.actorSystem.Stop(context.Background())
	}

	return nil
}

// spawnEngineActors creates the three cognitive engine actors
func (gcs *GoAktCognitiveSystem) spawnEngineActors() error {
	var err error

	// Spawn Affordance Engine Actor (Steps 0-5: Past experience processing)
	gcs.affordanceActor, err = gcs.actorSystem.Spawn(
		gcs.ctx,
		"affordance-engine",
		NewAffordanceEngineActor(gcs.sharedState, gcs.config),
	)
	if err != nil {
		return fmt.Errorf("failed to spawn affordance actor: %w", err)
	}

	// Spawn Relevance Engine Actor (Steps 0, 6: Pivotal relevance realization)
	gcs.relevanceActor, err = gcs.actorSystem.Spawn(
		gcs.ctx,
		"relevance-engine",
		NewRelevanceEngineActor(gcs.sharedState, gcs.config),
	)
	if err != nil {
		return fmt.Errorf("failed to spawn relevance actor: %w", err)
	}

	// Spawn Salience Engine Actor (Steps 6-11: Future simulation)
	gcs.salienceActor, err = gcs.actorSystem.Spawn(
		gcs.ctx,
		"salience-engine",
		NewSalienceEngineActor(gcs.sharedState, gcs.config),
	)
	if err != nil {
		return fmt.Errorf("failed to spawn salience actor: %w", err)
	}

	// Spawn Orchestrator Actor (Coordinates the 12-step loop)
	gcs.orchestratorActor, err = gcs.actorSystem.Spawn(
		gcs.ctx,
		"loop-orchestrator",
		NewOrchestratorActor(
			gcs.affordanceActor,
			gcs.relevanceActor,
			gcs.salienceActor,
			gcs.sharedState,
			gcs.config,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to spawn orchestrator actor: %w", err)
	}

	return nil
}

// RunCycle executes one complete 12-step cognitive cycle
func (gcs *GoAktCognitiveSystem) RunCycle() error {
	if !gcs.running.Load() {
		return fmt.Errorf("cognitive system not running")
	}

	// Send start cycle message to orchestrator
	err := gcs.actorSystem.Tell(gcs.ctx, gcs.orchestratorActor, &StartCycleMsg{
		Timestamp: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to start cycle: %w", err)
	}

	gcs.cycleCount.Add(1)
	gcs.lastCycleTime.Store(time.Now().UnixNano())

	return nil
}

// GetMetrics returns current system metrics
func (gcs *GoAktCognitiveSystem) GetMetrics() *CognitiveSystemMetrics {
	return &CognitiveSystemMetrics{
		Running:       gcs.running.Load(),
		CycleCount:    gcs.cycleCount.Load(),
		LastCycleTime: time.Unix(0, gcs.lastCycleTime.Load()),
		CoherenceScore: gcs.sharedState.GetCoherence(),
	}
}

// CognitiveSystemMetrics holds system performance metrics
type CognitiveSystemMetrics struct {
	Running        bool
	CycleCount     uint64
	LastCycleTime  time.Time
	CoherenceScore float64
}

// Message types for actor communication

// StartCycleMsg initiates a cognitive cycle
type StartCycleMsg struct {
	Timestamp time.Time
}

// StepMsg represents a step to be processed
type StepMsg struct {
	StepNumber int
	Timestamp  time.Time
	Payload    interface{}
}

// StepResultMsg contains the result of processing a step
type StepResultMsg struct {
	StepNumber     int
	EngineID       string
	Success        bool
	Output         interface{}
	Confidence     float64
	ProcessingTime time.Duration
	Error          error
}

// PivotalSyncMsg synchronizes engines at pivotal steps
type PivotalSyncMsg struct {
	PivotalStep   int
	EngineID      string
	StateSnapshot interface{}
	Timestamp     time.Time
}

// SyncAckMsg acknowledges a pivotal sync
type SyncAckMsg struct {
	PivotalStep int
	EngineID    string
	Ready       bool
}

// StateUpdateMsg broadcasts state changes
type StateUpdateMsg struct {
	SourceEngine string
	UpdateType   string
	StateData    interface{}
	Timestamp    time.Time
}

// SharedCognitiveState holds state shared across all engines
type SharedCognitiveState struct {
	mu sync.RWMutex

	// Current cognitive focus
	currentAttention interface{}
	attentionWeight  float64

	// Temporal integration
	pastContext   []interface{}
	presentFocus  interface{}
	futureOptions []interface{}

	// Coherence tracking
	coherenceScore   float64
	integrationLevel float64

	// Step synchronization
	currentStep       int
	pivotalStepReached bool
}

// NewSharedCognitiveState creates a new shared state
func NewSharedCognitiveState() *SharedCognitiveState {
	return &SharedCognitiveState{
		pastContext:    make([]interface{}, 0),
		futureOptions:  make([]interface{}, 0),
		coherenceScore: 1.0,
	}
}

// GetCoherence returns the current coherence score
func (scs *SharedCognitiveState) GetCoherence() float64 {
	scs.mu.RLock()
	defer scs.mu.RUnlock()
	return scs.coherenceScore
}

// UpdateCoherence updates the coherence score
func (scs *SharedCognitiveState) UpdateCoherence(score float64) {
	scs.mu.Lock()
	defer scs.mu.Unlock()
	scs.coherenceScore = score
}

// SetCurrentStep sets the current step
func (scs *SharedCognitiveState) SetCurrentStep(step int) {
	scs.mu.Lock()
	defer scs.mu.Unlock()
	scs.currentStep = step
}

// GetCurrentStep returns the current step
func (scs *SharedCognitiveState) GetCurrentStep() int {
	scs.mu.RLock()
	defer scs.mu.RUnlock()
	return scs.currentStep
}

// AddPastContext adds context from affordance processing
func (scs *SharedCognitiveState) AddPastContext(ctx interface{}) {
	scs.mu.Lock()
	defer scs.mu.Unlock()
	scs.pastContext = append(scs.pastContext, ctx)
	// Keep only recent context
	if len(scs.pastContext) > 10 {
		scs.pastContext = scs.pastContext[1:]
	}
}

// SetPresentFocus sets the current focus from relevance realization
func (scs *SharedCognitiveState) SetPresentFocus(focus interface{}) {
	scs.mu.Lock()
	defer scs.mu.Unlock()
	scs.presentFocus = focus
}

// AddFutureOption adds a future possibility from salience simulation
func (scs *SharedCognitiveState) AddFutureOption(option interface{}) {
	scs.mu.Lock()
	defer scs.mu.Unlock()
	scs.futureOptions = append(scs.futureOptions, option)
	// Keep only top options
	if len(scs.futureOptions) > 5 {
		scs.futureOptions = scs.futureOptions[1:]
	}
}
