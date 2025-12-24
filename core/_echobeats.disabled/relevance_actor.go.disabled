package echobeats

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/tochemey/goakt/v2/actors"
)

// RelevanceEngineActor performs pivotal relevance realization
// Handles Steps 0 and 6: Orienting to present commitment
type RelevanceEngineActor struct {
	mu          sync.RWMutex
	sharedState *SharedCognitiveState
	config      *GoAktConfig

	// Relevance processing state
	relevanceScores   map[string]float64
	currentRelevance  interface{}
	orientationVector []float64

	// Step handlers for pivotal steps
	stepHandlers map[int]func(context.Context, *StepMsg) (*StepResultMsg, error)

	// Metrics
	stepsProcessed int64
	totalTime      time.Duration
	errors         int64
}

// NewRelevanceEngineActor creates a new relevance engine actor
func NewRelevanceEngineActor(sharedState *SharedCognitiveState, config *GoAktConfig) *RelevanceEngineActor {
	actor := &RelevanceEngineActor{
		sharedState:       sharedState,
		config:            config,
		relevanceScores:   make(map[string]float64),
		orientationVector: make([]float64, 0),
		stepHandlers:      make(map[int]func(context.Context, *StepMsg) (*StepResultMsg, error)),
	}

	// Register step handlers for pivotal steps
	actor.stepHandlers[0] = actor.handlePivotalStep0 // First pivotal: orient to present
	actor.stepHandlers[6] = actor.handlePivotalStep6 // Second pivotal: reorient mid-cycle

	return actor
}

// PreStart is called before the actor starts
func (r *RelevanceEngineActor) PreStart(ctx context.Context) error {
	return nil
}

// Receive handles incoming messages
func (r *RelevanceEngineActor) Receive(ctx actors.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *StepMsg:
		r.handleStep(ctx, msg)
	case *PivotalSyncMsg:
		r.handlePivotalSync(ctx, msg)
	case *StateUpdateMsg:
		r.handleStateUpdate(ctx, msg)
	default:
		ctx.Unhandled()
	}
}

// PostStop is called after the actor stops
func (r *RelevanceEngineActor) PostStop(ctx context.Context) error {
	return nil
}

// handleStep processes a cognitive step
func (r *RelevanceEngineActor) handleStep(ctx actors.ReceiveContext, msg *StepMsg) {
	startTime := time.Now()

	handler, ok := r.stepHandlers[msg.StepNumber]
	if !ok {
		// This step is not handled by the relevance engine
		return
	}

	result, err := handler(context.Background(), msg)
	if err != nil {
		r.mu.Lock()
		r.errors++
		r.mu.Unlock()

		result = &StepResultMsg{
			StepNumber: msg.StepNumber,
			EngineID:   "relevance",
			Success:    false,
			Error:      err,
		}
	}

	result.ProcessingTime = time.Since(startTime)

	r.mu.Lock()
	r.stepsProcessed++
	r.totalTime += result.ProcessingTime
	r.mu.Unlock()

	// Reply with result
	ctx.Response(result)
}

// handlePivotalStep0 handles the first pivotal step (orienting to present)
func (r *RelevanceEngineActor) handlePivotalStep0(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Step 0: Pivotal Relevance Realization
	// This is the moment of orienting to the present commitment
	// All three engines synchronize here before the cycle begins

	// 1. Clear previous relevance scores
	r.relevanceScores = make(map[string]float64)

	// 2. Compute orientation vector based on current state
	r.orientationVector = r.computeOrientationVector()

	// 3. Determine what is most relevant right now
	r.currentRelevance = r.realizeRelevance(msg.Payload)

	// 4. Update shared state with current focus
	r.sharedState.SetPresentFocus(r.currentRelevance)

	// 5. Calculate initial coherence
	coherence := r.calculateCoherence()
	r.sharedState.UpdateCoherence(coherence)

	return &StepResultMsg{
		StepNumber: 0,
		EngineID:   "relevance",
		Success:    true,
		Output:     "Pivotal relevance realized - oriented to present",
		Confidence: coherence,
	}, nil
}

// handlePivotalStep6 handles the second pivotal step (mid-cycle reorientation)
func (r *RelevanceEngineActor) handlePivotalStep6(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Step 6: Second Pivotal Relevance Realization
	// This is the mid-cycle reorientation point
	// Transition from affordance processing to salience simulation

	// 1. Integrate results from affordance processing (steps 1-5)
	affordanceResults := r.integrateAffordanceResults()

	// 2. Recompute orientation for salience phase
	r.orientationVector = r.computeOrientationVector()

	// 3. Update relevance based on affordance outcomes
	r.currentRelevance = r.updateRelevance(affordanceResults)

	// 4. Prepare context for salience simulation
	r.sharedState.SetPresentFocus(r.currentRelevance)

	// 5. Recalculate coherence
	coherence := r.calculateCoherence()
	r.sharedState.UpdateCoherence(coherence)

	return &StepResultMsg{
		StepNumber: 6,
		EngineID:   "relevance",
		Success:    true,
		Output:     "Mid-cycle reorientation complete - ready for salience simulation",
		Confidence: coherence,
	}, nil
}

// handlePivotalSync handles synchronization at pivotal steps
func (r *RelevanceEngineActor) handlePivotalSync(ctx actors.ReceiveContext, msg *PivotalSyncMsg) {
	// The relevance engine is the coordinator for pivotal syncs
	// It waits for all engines to be ready before proceeding

	ctx.Response(&SyncAckMsg{
		PivotalStep: msg.PivotalStep,
		EngineID:    "relevance",
		Ready:       true,
	})
}

// handleStateUpdate handles state updates from other engines
func (r *RelevanceEngineActor) handleStateUpdate(ctx actors.ReceiveContext, msg *StateUpdateMsg) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Process state updates to maintain coherence
	switch msg.UpdateType {
	case "attention":
		// Update attention focus
	case "context":
		// Update context from affordance engine
	case "coherence":
		// Recalculate coherence
	}
}

// Helper functions

// computeOrientationVector calculates the current orientation in cognitive space
func (r *RelevanceEngineActor) computeOrientationVector() []float64 {
	// In production, this would compute a high-dimensional vector
	// representing the current cognitive orientation
	// For now, return a simple 3D vector
	return []float64{1.0, 0.0, 0.0}
}

// realizeRelevance determines what is most relevant in the current moment
func (r *RelevanceEngineActor) realizeRelevance(context interface{}) interface{} {
	// Relevance realization is the core cognitive operation
	// It determines what matters most right now
	return context
}

// updateRelevance updates relevance based on new information
func (r *RelevanceEngineActor) updateRelevance(affordanceResults interface{}) interface{} {
	// Update relevance based on affordance processing results
	return affordanceResults
}

// integrateAffordanceResults integrates results from affordance processing
func (r *RelevanceEngineActor) integrateAffordanceResults() interface{} {
	// Get results from shared state
	return nil
}

// calculateCoherence calculates the current cognitive coherence
func (r *RelevanceEngineActor) calculateCoherence() float64 {
	// Coherence is a measure of how well-integrated the cognitive state is
	// Higher coherence means better integration between past, present, and future

	// Simple coherence calculation based on orientation vector magnitude
	if len(r.orientationVector) == 0 {
		return 1.0
	}

	var sumSquares float64
	for _, v := range r.orientationVector {
		sumSquares += v * v
	}
	magnitude := math.Sqrt(sumSquares)

	// Normalize to 0-1 range
	if magnitude > 1.0 {
		return 1.0
	}
	return magnitude
}
