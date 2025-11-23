package echobeats

import "time"

// Term represents a cognitive term in System 4 architecture
type Term int

const (
	T1_Perception       Term = 1 // Perception (Need vs Capacity)
	T2_IdeaFormation    Term = 2 // Idea Formation
	T4_SensoryInput     Term = 4 // Sensory Input
	T5_ActionSequence   Term = 5 // Action Sequence
	T7_MemoryEncoding   Term = 7 // Memory Encoding
	T8_BalancedResponse Term = 8 // Balanced Response
)

func (t Term) String() string {
	switch t {
	case T1_Perception:
		return "T1_Perception"
	case T2_IdeaFormation:
		return "T2_IdeaFormation"
	case T4_SensoryInput:
		return "T4_SensoryInput"
	case T5_ActionSequence:
		return "T5_ActionSequence"
	case T7_MemoryEncoding:
		return "T7_MemoryEncoding"
	case T8_BalancedResponse:
		return "T8_BalancedResponse"
	default:
		return "Unknown"
	}
}

// Mode represents processing mode
type Mode int

const (
	Expressive Mode = iota // E - Reactive, Action-oriented
	Reflective             // R - Anticipatory, Simulation-oriented
)

func (m Mode) String() string {
	if m == Expressive {
		return "E"
	}
	return "R"
}

// StepConfig defines the configuration for a single step in the 12-step cycle
type StepConfig struct {
	Step  int
	Phase int
	Term  Term
	Mode  Mode
}

// buildStepConfigs creates the 12-step configuration matrix
func buildStepConfigs() []StepConfig {
	return []StepConfig{
		// Step 0: Phase 0 - T4E (Sensory Input, Expressive)
		{Step: 0, Phase: 0, Term: T4_SensoryInput, Mode: Expressive},

		// Step 1: Phase 1 - T1R (Perception, Reflective) - Pivotal Relevance Realization
		{Step: 1, Phase: 1, Term: T1_Perception, Mode: Reflective},

		// Step 2: Phase 2 - T2E (Idea Formation, Expressive)
		{Step: 2, Phase: 2, Term: T2_IdeaFormation, Mode: Expressive},

		// Step 3: Phase 0 - T7R (Memory Encoding, Reflective)
		{Step: 3, Phase: 0, Term: T7_MemoryEncoding, Mode: Reflective},

		// Step 4: Phase 1 - T4E (Sensory Input, Expressive) - Affordance Interaction
		{Step: 4, Phase: 1, Term: T4_SensoryInput, Mode: Expressive},

		// Step 5: Phase 2 - T1R (Perception, Reflective) - Affordance Interaction
		{Step: 5, Phase: 2, Term: T1_Perception, Mode: Reflective},

		// Step 6: Phase 0 - T2E (Idea Formation, Expressive) - Affordance Interaction
		{Step: 6, Phase: 0, Term: T2_IdeaFormation, Mode: Expressive},

		// Step 7: Phase 1 - T5E (Action Sequence, Expressive) - Pivotal Relevance Realization
		{Step: 7, Phase: 1, Term: T5_ActionSequence, Mode: Expressive},

		// Step 8: Phase 2 - T8E (Balanced Response, Expressive) - Affordance Interaction
		{Step: 8, Phase: 2, Term: T8_BalancedResponse, Mode: Expressive},

		// Step 9: Phase 0 - T8E (Balanced Response, Expressive) - Affordance Interaction
		{Step: 9, Phase: 0, Term: T8_BalancedResponse, Mode: Expressive},

		// Step 10: Phase 1 - T7R (Memory Encoding, Reflective) - Salience Simulation
		{Step: 10, Phase: 1, Term: T7_MemoryEncoding, Mode: Reflective},

		// Step 11: Phase 2 - T5E (Action Sequence, Expressive) - Salience Simulation
		{Step: 11, Phase: 2, Term: T5_ActionSequence, Mode: Expressive},
	}
}

// StepHandler is a function that handles a cognitive step
type StepHandler func(step int, mode Mode) error

// CognitiveMetrics tracks metrics for the cognitive system
type CognitiveMetrics struct {
	CurrentStep      int
	TotalSteps       uint64
	ExpressiveSteps  uint64
	ReflectiveSteps  uint64
	LastStepTime     time.Time
	AverageStepTime  time.Duration
	CognitiveLoad    float64
}
