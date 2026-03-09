package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// FourEDimension represents one of the four dimensions of 4E cognition
type FourEDimension int

const (
	DimensionEmbodied  FourEDimension = iota // Body schema, proprioception, somatic markers
	DimensionEmbedded                        // Affordance detection, niche coupling, environmental sensitivity
	DimensionEnacted                         // Sensorimotor coordination, prediction accuracy, active inference
	DimensionExtended                        // Tool use, external memory, social cognition
)

// FourEDimensionNames maps dimensions to display names
var FourEDimensionNames = map[FourEDimension]string{
	DimensionEmbodied: "Embodied",
	DimensionEmbedded: "Embedded",
	DimensionEnacted:  "Enacted",
	DimensionExtended: "Extended",
}

func (d FourEDimension) String() string {
	if name, ok := FourEDimensionNames[d]; ok {
		return name
	}
	return "Unknown"
}

// EmbodiedMetrics tracks body-schema integration metrics
type EmbodiedMetrics struct {
	BodySchemaIntegration  float64 // How well internal state is represented [0,1]
	ProprioceptiveAccuracy float64 // Accuracy of self-state estimation [0,1]
	SomaticMarkerStrength  float64 // Strength of embodied emotional signals [0,1]
	InteroceptiveAwareness float64 // Awareness of internal states [0,1]
}

// Score returns the aggregate embodied score
func (em *EmbodiedMetrics) Score() float64 {
	return (em.BodySchemaIntegration + em.ProprioceptiveAccuracy +
		em.SomaticMarkerStrength + em.InteroceptiveAwareness) / 4.0
}

// EmbeddedMetrics tracks environmental coupling metrics
type EmbeddedMetrics struct {
	AffordanceDetectionRate   float64 // Rate of detecting action possibilities [0,1]
	NicheCoupling             float64 // Degree of environment adaptation [0,1]
	EnvironmentalSensitivity  float64 // Responsiveness to environmental changes [0,1]
	ContextualAppropriatenessFloat float64 // How well actions fit context [0,1]
}

// Score returns the aggregate embedded score
func (em *EmbeddedMetrics) Score() float64 {
	return (em.AffordanceDetectionRate + em.NicheCoupling +
		em.EnvironmentalSensitivity + em.ContextualAppropriatenessFloat) / 4.0
}

// EnactedMetrics tracks sensorimotor coordination metrics
type EnactedMetrics struct {
	SensorimotorCoordination float64 // Perception-action coupling quality [0,1]
	PredictionAccuracy       float64 // How well predictions match outcomes [0,1]
	ActiveInferenceEfficiency float64 // Efficiency of active inference loops [0,1]
	AdaptiveFlexibility      float64 // Ability to adapt behavior to new situations [0,1]
}

// Score returns the aggregate enacted score
func (em *EnactedMetrics) Score() float64 {
	return (em.SensorimotorCoordination + em.PredictionAccuracy +
		em.ActiveInferenceEfficiency + em.AdaptiveFlexibility) / 4.0
}

// ExtendedMetrics tracks tool use and social cognition metrics
type ExtendedMetrics struct {
	ToolUseCompetence        float64 // Skill in using external tools [0,1]
	ExternalMemoryIntegration float64 // Integration with external memory systems [0,1]
	SocialCognitionDepth     float64 // Depth of understanding others [0,1]
	CollaborativeCapacity    float64 // Ability to work with others [0,1]
}

// Score returns the aggregate extended score
func (em *ExtendedMetrics) Score() float64 {
	return (em.ToolUseCompetence + em.ExternalMemoryIntegration +
		em.SocialCognitionDepth + em.CollaborativeCapacity) / 4.0
}

// FourECognitionState holds the complete 4E cognition state
type FourECognitionState struct {
	mu sync.RWMutex

	Embodied EmbodiedMetrics
	Embedded EmbeddedMetrics
	Enacted  EnactedMetrics
	Extended ExtendedMetrics

	// History for tracking evolution
	history     []FourESnapshot
	maxHistory  int

	// Integration with endocrine system
	endocrine *VirtualEndocrineSystem

	// Update tracking
	lastUpdate time.Time
	updateCount uint64
}

// FourESnapshot captures a point-in-time 4E state
type FourESnapshot struct {
	Timestamp     time.Time
	EmbodiedScore float64
	EmbeddedScore float64
	EnactedScore  float64
	ExtendedScore float64
	OverallScore  float64
	CognitiveMode CognitiveMode
}

// NewFourECognitionState creates a new 4E cognition tracker
func NewFourECognitionState(endocrine *VirtualEndocrineSystem) *FourECognitionState {
	state := &FourECognitionState{
		history:    make([]FourESnapshot, 0, 500),
		maxHistory: 500,
		endocrine:  endocrine,
		lastUpdate: time.Now(),
	}

	// Initialize with baseline values
	state.Embodied = EmbodiedMetrics{
		BodySchemaIntegration:  0.3,
		ProprioceptiveAccuracy: 0.2,
		SomaticMarkerStrength:  0.1,
		InteroceptiveAwareness: 0.2,
	}
	state.Embedded = EmbeddedMetrics{
		AffordanceDetectionRate:   0.2,
		NicheCoupling:             0.3,
		EnvironmentalSensitivity:  0.3,
		ContextualAppropriatenessFloat: 0.2,
	}
	state.Enacted = EnactedMetrics{
		SensorimotorCoordination: 0.2,
		PredictionAccuracy:       0.2,
		ActiveInferenceEfficiency: 0.2,
		AdaptiveFlexibility:      0.3,
	}
	state.Extended = ExtendedMetrics{
		ToolUseCompetence:        0.3,
		ExternalMemoryIntegration: 0.2,
		SocialCognitionDepth:     0.1,
		CollaborativeCapacity:    0.2,
	}

	return state
}

// OverallScore returns the aggregate 4E score
func (fes *FourECognitionState) OverallScore() float64 {
	fes.mu.RLock()
	defer fes.mu.RUnlock()
	return (fes.Embodied.Score() + fes.Embedded.Score() +
		fes.Enacted.Score() + fes.Extended.Score()) / 4.0
}

// DimensionScores returns all four dimension scores
func (fes *FourECognitionState) DimensionScores() map[string]float64 {
	fes.mu.RLock()
	defer fes.mu.RUnlock()
	return map[string]float64{
		"embodied": fes.Embodied.Score(),
		"embedded": fes.Embedded.Score(),
		"enacted":  fes.Enacted.Score(),
		"extended": fes.Extended.Score(),
		"overall":  (fes.Embodied.Score() + fes.Embedded.Score() + fes.Enacted.Score() + fes.Extended.Score()) / 4.0,
	}
}

// UpdateFromCognitiveEvent updates 4E metrics based on a cognitive event
func (fes *FourECognitionState) UpdateFromCognitiveEvent(eventType string, success bool, metadata map[string]float64) {
	fes.mu.Lock()
	defer fes.mu.Unlock()

	fes.updateCount++
	fes.lastUpdate = time.Now()

	// Learning rate decays with experience (slower changes as system matures)
	lr := 0.01 / (1.0 + float64(fes.updateCount)*0.0001)

	successMultiplier := 1.0
	if !success {
		successMultiplier = -0.5
	}

	switch eventType {
	case "self_state_estimation":
		fes.Embodied.ProprioceptiveAccuracy += lr * successMultiplier
		fes.Embodied.BodySchemaIntegration += lr * successMultiplier * 0.5

	case "somatic_marker":
		fes.Embodied.SomaticMarkerStrength += lr * successMultiplier
		if fes.endocrine != nil {
			// Somatic markers are stronger when endocrine system is active
			_, arousal, _ := fes.endocrine.GetAffectiveState()
			fes.Embodied.InteroceptiveAwareness += lr * arousal * 0.5
		}

	case "affordance_detection":
		fes.Embedded.AffordanceDetectionRate += lr * successMultiplier
		fes.Embedded.NicheCoupling += lr * successMultiplier * 0.3

	case "environment_response":
		fes.Embedded.EnvironmentalSensitivity += lr * successMultiplier
		fes.Embedded.ContextualAppropriatenessFloat += lr * successMultiplier * 0.5

	case "prediction":
		accuracy := 0.0
		if v, ok := metadata["accuracy"]; ok {
			accuracy = v
		}
		fes.Enacted.PredictionAccuracy += lr * (accuracy - fes.Enacted.PredictionAccuracy)
		fes.Enacted.ActiveInferenceEfficiency += lr * successMultiplier * 0.3

	case "action_execution":
		fes.Enacted.SensorimotorCoordination += lr * successMultiplier
		fes.Enacted.AdaptiveFlexibility += lr * successMultiplier * 0.5

	case "tool_use":
		fes.Extended.ToolUseCompetence += lr * successMultiplier
		fes.Extended.ExternalMemoryIntegration += lr * successMultiplier * 0.3

	case "social_interaction":
		fes.Extended.SocialCognitionDepth += lr * successMultiplier
		fes.Extended.CollaborativeCapacity += lr * successMultiplier * 0.5

	case "memory_retrieval":
		fes.Extended.ExternalMemoryIntegration += lr * successMultiplier
		fes.Enacted.ActiveInferenceEfficiency += lr * successMultiplier * 0.2
	}

	// Clamp all metrics to [0, 1]
	fes.clampAll()

	// Record snapshot
	fes.recordSnapshot()
}

// clampAll ensures all metrics are within [0, 1]
func (fes *FourECognitionState) clampAll() {
	clamp := func(v *float64) {
		if *v < 0 {
			*v = 0
		}
		if *v > 1 {
			*v = 1
		}
	}

	clamp(&fes.Embodied.BodySchemaIntegration)
	clamp(&fes.Embodied.ProprioceptiveAccuracy)
	clamp(&fes.Embodied.SomaticMarkerStrength)
	clamp(&fes.Embodied.InteroceptiveAwareness)

	clamp(&fes.Embedded.AffordanceDetectionRate)
	clamp(&fes.Embedded.NicheCoupling)
	clamp(&fes.Embedded.EnvironmentalSensitivity)
	clamp(&fes.Embedded.ContextualAppropriatenessFloat)

	clamp(&fes.Enacted.SensorimotorCoordination)
	clamp(&fes.Enacted.PredictionAccuracy)
	clamp(&fes.Enacted.ActiveInferenceEfficiency)
	clamp(&fes.Enacted.AdaptiveFlexibility)

	clamp(&fes.Extended.ToolUseCompetence)
	clamp(&fes.Extended.ExternalMemoryIntegration)
	clamp(&fes.Extended.SocialCognitionDepth)
	clamp(&fes.Extended.CollaborativeCapacity)
}

// recordSnapshot captures current state for history
func (fes *FourECognitionState) recordSnapshot() {
	mode := CognitiveMode(0)
	if fes.endocrine != nil {
		mode, _ = fes.endocrine.Bus.CurrentMode()
	}

	snapshot := FourESnapshot{
		Timestamp:     time.Now(),
		EmbodiedScore: fes.Embodied.Score(),
		EmbeddedScore: fes.Embedded.Score(),
		EnactedScore:  fes.Enacted.Score(),
		ExtendedScore: fes.Extended.Score(),
		OverallScore:  (fes.Embodied.Score() + fes.Embedded.Score() + fes.Enacted.Score() + fes.Extended.Score()) / 4.0,
		CognitiveMode: mode,
	}

	fes.history = append(fes.history, snapshot)
	if len(fes.history) > fes.maxHistory {
		fes.history = fes.history[1:]
	}
}

// GetHistory returns recent 4E snapshots
func (fes *FourECognitionState) GetHistory(n int) []FourESnapshot {
	fes.mu.RLock()
	defer fes.mu.RUnlock()
	if n > len(fes.history) {
		n = len(fes.history)
	}
	start := len(fes.history) - n
	result := make([]FourESnapshot, n)
	copy(result, fes.history[start:])
	return result
}

// FourEMaturityLevel represents the 4E cognitive maturity level
type FourEMaturityLevel int

const (
	MaturityNascent      FourEMaturityLevel = iota // Initial formation
	MaturityDeveloping                             // Basic capabilities
	MaturityIntegrating                            // Developing complexity
	MaturityMature                                 // Full capability
	MaturityTranscendent                           // Beyond normal limits
)

// FourEMaturityNames maps maturity levels to display names
var FourEMaturityNames = map[FourEMaturityLevel]string{
	MaturityNascent:      "Nascent",
	MaturityDeveloping:   "Developing",
	MaturityIntegrating:  "Integrating",
	MaturityMature:       "Mature",
	MaturityTranscendent: "Transcendent",
}

func (ml FourEMaturityLevel) String() string {
	if name, ok := FourEMaturityNames[ml]; ok {
		return name
	}
	return "Unknown"
}

// FourEMaturityThresholds defines the 4E score thresholds for maturity advancement
var FourEMaturityThresholds = map[FourEMaturityLevel]float64{
	MaturityNascent:      0.0,
	MaturityDeveloping:   0.25,
	MaturityIntegrating:  0.45,
	MaturityMature:       0.65,
	MaturityTranscendent: 0.85,
}

// DetermineMaturity returns the current 4E maturity level based on scores and wisdom
func (fes *FourECognitionState) DetermineMaturity(wisdomScore float64) FourEMaturityLevel {
	fes.mu.RLock()
	defer fes.mu.RUnlock()

	overall := (fes.Embodied.Score() + fes.Embedded.Score() +
		fes.Enacted.Score() + fes.Extended.Score()) / 4.0

	// Maturity requires both 4E score and wisdom score
	level := MaturityNascent
	for l := MaturityTranscendent; l >= MaturityNascent; l-- {
		threshold := FourEMaturityThresholds[l]
		wisdomThreshold := threshold * 0.8
		if overall >= threshold && wisdomScore >= wisdomThreshold {
			level = l
			break
		}
	}

	return level
}

// Ensure math is used
var _ = math.Max
