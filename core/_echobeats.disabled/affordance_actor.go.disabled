package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tochemey/goakt/v2/actors"
)

// AffordanceEngineActor processes past experiences and actual interactions
// Handles Steps 0-5: Conditioning from past performance
type AffordanceEngineActor struct {
	mu           sync.RWMutex
	sharedState  *SharedCognitiveState
	config       *GoAktConfig
	
	// Affordance processing state
	pastExperiences []Affordance
	affordances     []Affordance
	selectedAction  *Affordance
	
	// Step handlers for steps 0-5
	stepHandlers    map[int]func(context.Context, *StepMsg) (*StepResultMsg, error)
	
	// Metrics
	stepsProcessed  int64
	totalTime       time.Duration
	errors          int64
}

// Affordance represents an action possibility from past experience
type Affordance struct {
	Action      string
	Context     interface{}
	PastSuccess float64
	Confidence  float64
	Timestamp   time.Time
}

// NewAffordanceEngineActor creates a new affordance engine actor
func NewAffordanceEngineActor(sharedState *SharedCognitiveState, config *GoAktConfig) *AffordanceEngineActor {
	actor := &AffordanceEngineActor{
		sharedState:     sharedState,
		config:          config,
		pastExperiences: make([]Affordance, 0),
		affordances:     make([]Affordance, 0),
		stepHandlers:    make(map[int]func(context.Context, *StepMsg) (*StepResultMsg, error)),
	}
	
	// Register step handlers
	actor.stepHandlers[0] = actor.handleStep0 // Pivotal relevance (shared with relevance engine)
	actor.stepHandlers[1] = actor.handleStep1 // Recall past experiences
	actor.stepHandlers[2] = actor.handleStep2 // Evaluate affordances
	actor.stepHandlers[3] = actor.handleStep3 // Select action candidates
	actor.stepHandlers[4] = actor.handleStep4 // Refine selection
	actor.stepHandlers[5] = actor.handleStep5 // Commit to action
	
	return actor
}

// PreStart is called before the actor starts
func (a *AffordanceEngineActor) PreStart(ctx context.Context) error {
	return nil
}

// Receive handles incoming messages
func (a *AffordanceEngineActor) Receive(ctx actors.ReceiveContext) {
	switch msg := ctx.Message().(type) {
	case *StepMsg:
		a.handleStep(ctx, msg)
	case *PivotalSyncMsg:
		a.handlePivotalSync(ctx, msg)
	case *StateUpdateMsg:
		a.handleStateUpdate(ctx, msg)
	default:
		ctx.Unhandled()
	}
}

// PostStop is called after the actor stops
func (a *AffordanceEngineActor) PostStop(ctx context.Context) error {
	return nil
}

// handleStep processes a cognitive step
func (a *AffordanceEngineActor) handleStep(ctx actors.ReceiveContext, msg *StepMsg) {
	startTime := time.Now()
	
	handler, ok := a.stepHandlers[msg.StepNumber]
	if !ok {
		// This step is not handled by the affordance engine
		return
	}
	
	result, err := handler(context.Background(), msg)
	if err != nil {
		a.mu.Lock()
		a.errors++
		a.mu.Unlock()
		
		result = &StepResultMsg{
			StepNumber: msg.StepNumber,
			EngineID:   "affordance",
			Success:    false,
			Error:      err,
		}
	}
	
	result.ProcessingTime = time.Since(startTime)
	
	a.mu.Lock()
	a.stepsProcessed++
	a.totalTime += result.ProcessingTime
	a.mu.Unlock()
	
	// Reply with result
	ctx.Response(result)
}

// Step 0: Pivotal relevance realization (orienting to present)
func (a *AffordanceEngineActor) handleStep0(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// At step 0, we orient to the present moment and prepare for affordance processing
	// This is a pivotal step shared with the relevance engine
	
	// Clear previous cycle's affordances
	a.affordances = make([]Affordance, 0)
	
	return &StepResultMsg{
		StepNumber: 0,
		EngineID:   "affordance",
		Success:    true,
		Output:     "Oriented to present, ready for affordance processing",
		Confidence: 1.0,
	}, nil
}

// Step 1: Recall past experiences
func (a *AffordanceEngineActor) handleStep1(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Retrieve relevant past experiences based on current context
	// In production, this would query the hypergraph memory
	
	relevantExperiences := a.recallRelevantExperiences(msg.Payload)
	a.pastExperiences = relevantExperiences
	
	// Share context with other engines
	a.sharedState.AddPastContext(relevantExperiences)
	
	return &StepResultMsg{
		StepNumber: 1,
		EngineID:   "affordance",
		Success:    true,
		Output:     fmt.Sprintf("Recalled %d relevant experiences", len(relevantExperiences)),
		Confidence: 0.9,
	}, nil
}

// Step 2: Evaluate affordances
func (a *AffordanceEngineActor) handleStep2(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Generate affordances from past experiences
	for _, exp := range a.pastExperiences {
		affordance := Affordance{
			Action:      exp.Action,
			Context:     exp.Context,
			PastSuccess: exp.PastSuccess,
			Confidence:  a.evaluateAffordanceConfidence(exp),
			Timestamp:   time.Now(),
		}
		a.affordances = append(a.affordances, affordance)
	}
	
	return &StepResultMsg{
		StepNumber: 2,
		EngineID:   "affordance",
		Success:    true,
		Output:     fmt.Sprintf("Evaluated %d affordances", len(a.affordances)),
		Confidence: 0.85,
	}, nil
}

// Step 3: Select action candidates
func (a *AffordanceEngineActor) handleStep3(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Filter and rank affordances by confidence and past success
	candidates := a.selectTopCandidates(a.affordances, 5)
	
	return &StepResultMsg{
		StepNumber: 3,
		EngineID:   "affordance",
		Success:    true,
		Output:     fmt.Sprintf("Selected %d action candidates", len(candidates)),
		Confidence: 0.8,
	}, nil
}

// Step 4: Refine selection
func (a *AffordanceEngineActor) handleStep4(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Refine selection based on current context and relevance
	if len(a.affordances) > 0 {
		// Select the best affordance
		best := a.affordances[0]
		for _, aff := range a.affordances {
			if aff.Confidence*aff.PastSuccess > best.Confidence*best.PastSuccess {
				best = aff
			}
		}
		a.selectedAction = &best
	}
	
	return &StepResultMsg{
		StepNumber: 4,
		EngineID:   "affordance",
		Success:    true,
		Output:     "Refined action selection",
		Confidence: 0.85,
	}, nil
}

// Step 5: Commit to action
func (a *AffordanceEngineActor) handleStep5(ctx context.Context, msg *StepMsg) (*StepResultMsg, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Commit to the selected action
	if a.selectedAction != nil {
		return &StepResultMsg{
			StepNumber: 5,
			EngineID:   "affordance",
			Success:    true,
			Output:     fmt.Sprintf("Committed to action: %s", a.selectedAction.Action),
			Confidence: a.selectedAction.Confidence,
		}, nil
	}
	
	return &StepResultMsg{
		StepNumber: 5,
		EngineID:   "affordance",
		Success:    true,
		Output:     "No action selected",
		Confidence: 0.5,
	}, nil
}

// handlePivotalSync handles synchronization at pivotal steps
func (a *AffordanceEngineActor) handlePivotalSync(ctx actors.ReceiveContext, msg *PivotalSyncMsg) {
	// Acknowledge the sync
	ctx.Response(&SyncAckMsg{
		PivotalStep: msg.PivotalStep,
		EngineID:    "affordance",
		Ready:       true,
	})
}

// handleStateUpdate handles state updates from other engines
func (a *AffordanceEngineActor) handleStateUpdate(ctx actors.ReceiveContext, msg *StateUpdateMsg) {
	// Process state updates from relevance and salience engines
	// This enables inter-engine coordination
}

// Helper functions

func (a *AffordanceEngineActor) recallRelevantExperiences(context interface{}) []Affordance {
	// In production, this would query the Dgraph hypergraph
	// For now, return empty slice
	return []Affordance{}
}

func (a *AffordanceEngineActor) evaluateAffordanceConfidence(exp Affordance) float64 {
	// Calculate confidence based on past success and recency
	baseConfidence := exp.PastSuccess
	// Decay based on time since last use
	timeSince := time.Since(exp.Timestamp)
	decayFactor := 1.0 / (1.0 + timeSince.Hours()/24.0)
	return baseConfidence * decayFactor
}

func (a *AffordanceEngineActor) selectTopCandidates(affordances []Affordance, n int) []Affordance {
	if len(affordances) <= n {
		return affordances
	}
	// Simple selection of top n by confidence
	// In production, use more sophisticated ranking
	return affordances[:n]
}
