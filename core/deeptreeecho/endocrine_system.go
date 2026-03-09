package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// Hormone represents a specific hormone in the virtual endocrine system
type Hormone int

const (
	Cortisol         Hormone = iota // HPA axis - stress response
	DopamineTonic                   // Dopaminergic - baseline motivation
	DopaminePhasic                  // Dopaminergic - reward/surprise burst
	Serotonin                       // Serotonergic - mood/patience/satiety
	Norepinephrine                  // Noradrenergic - alertness/attention
	Oxytocin                        // Oxytocinergic - social bonding/trust
	Melatonin                       // Circadian - sleep/rest drive
	Insulin                         // Pancreatic - energy regulation
	CytokineIL6                     // Immune - inflammation/sickness behavior
	Anandamide                      // Endocannabinoid - bliss/flow/pain modulation
	HormoneCount                    // Sentinel for iteration
)

// HormoneNames maps hormones to their display names
var HormoneNames = map[Hormone]string{
	Cortisol:       "Cortisol",
	DopamineTonic:  "Dopamine (Tonic)",
	DopaminePhasic: "Dopamine (Phasic)",
	Serotonin:      "Serotonin",
	Norepinephrine: "Norepinephrine",
	Oxytocin:       "Oxytocin",
	Melatonin:      "Melatonin",
	Insulin:        "Insulin",
	CytokineIL6:    "Cytokine IL-6",
	Anandamide:     "Anandamide",
}

// CognitiveMode represents emergent cognitive modes from hormone combinations
type CognitiveMode int

const (
	ModeExplore    CognitiveMode = iota // High NE + High DA_phasic: novelty seeking
	ModeExploit                         // High DA_tonic + High 5HT: routine optimization
	ModeFight      CognitiveMode = iota // High cortisol + High NE: threat response
	ModeFlight                          // High cortisol + Low NE: avoidance
	ModeFreeze                          // High cortisol + Low DA: paralysis
	ModeFlow                            // High anandamide + moderate NE: optimal performance
	ModeSocial                          // High oxytocin + Low cortisol: bonding
	ModeRest                            // High melatonin + Low NE: recovery
	ModeCreative                        // High DA_phasic + High anandamide: divergent thinking
	ModeAnalytical                      // High 5HT + moderate NE: convergent thinking
)

// CognitiveModeNames maps modes to display names
var CognitiveModeNames = map[CognitiveMode]string{
	ModeExplore:    "Explore",
	ModeExploit:    "Exploit",
	ModeFight:      "Fight",
	ModeFlight:     "Flight",
	ModeFreeze:     "Freeze",
	ModeFlow:       "Flow",
	ModeSocial:     "Social",
	ModeRest:       "Rest",
	ModeCreative:   "Creative",
	ModeAnalytical: "Analytical",
}

func (cm CognitiveMode) String() string {
	if name, ok := CognitiveModeNames[cm]; ok {
		return name
	}
	return "Unknown"
}

// EndocrineEvent represents a signal to the endocrine system
type EndocrineEvent int

const (
	EndoNoveltyEncountered EndocrineEvent = iota
	EndoRewardReceived
	EndoThreatDetected
	EndoSocialContact
	EndoErrorSignal
	EndoInsightGained
	EndoFatigueAccumulated
	EndoRestCompleted
	EndoGoalAchieved
	EndoGoalFrustrated
)

// GlandConfig defines parameters for a virtual gland
type GlandConfig struct {
	BaselineLevel  float64       // Resting concentration [0,1]
	DecayRate      float64       // Exponential decay rate per second
	RiseRate       float64       // Maximum rise rate per second
	MinLevel       float64       // Minimum concentration
	MaxLevel       float64       // Maximum concentration
	HalfLife       time.Duration // Time to decay to half
	SensitivityMap map[EndocrineEvent]float64
}

// DefaultGlandConfigs returns biologically-inspired gland configurations
func DefaultGlandConfigs() map[Hormone]GlandConfig {
	return map[Hormone]GlandConfig{
		Cortisol: {
			BaselineLevel: 0.2, DecayRate: 0.01, RiseRate: 0.15,
			MinLevel: 0.05, MaxLevel: 1.0, HalfLife: 60 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoThreatDetected:     0.6,
				EndoErrorSignal:        0.3,
				EndoGoalFrustrated:     0.4,
				EndoFatigueAccumulated: 0.2,
			},
		},
		DopamineTonic: {
			BaselineLevel: 0.5, DecayRate: 0.005, RiseRate: 0.05,
			MinLevel: 0.1, MaxLevel: 0.9, HalfLife: 120 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoGoalAchieved: 0.1,
				EndoRestCompleted: 0.05,
			},
		},
		DopaminePhasic: {
			BaselineLevel: 0.1, DecayRate: 0.2, RiseRate: 0.8,
			MinLevel: 0.0, MaxLevel: 1.0, HalfLife: 3 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoRewardReceived:     0.7,
				EndoNoveltyEncountered: 0.5,
				EndoInsightGained:      0.8,
			},
		},
		Serotonin: {
			BaselineLevel: 0.5, DecayRate: 0.003, RiseRate: 0.03,
			MinLevel: 0.1, MaxLevel: 0.9, HalfLife: 180 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoGoalAchieved:  0.15,
				EndoSocialContact: 0.1,
				EndoRestCompleted: 0.1,
			},
		},
		Norepinephrine: {
			BaselineLevel: 0.3, DecayRate: 0.05, RiseRate: 0.4,
			MinLevel: 0.05, MaxLevel: 1.0, HalfLife: 15 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoNoveltyEncountered: 0.4,
				EndoThreatDetected:     0.5,
				EndoErrorSignal:        0.2,
			},
		},
		Oxytocin: {
			BaselineLevel: 0.3, DecayRate: 0.008, RiseRate: 0.1,
			MinLevel: 0.05, MaxLevel: 0.95, HalfLife: 90 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoSocialContact: 0.5,
				EndoGoalAchieved:  0.1,
			},
		},
		Melatonin: {
			BaselineLevel: 0.1, DecayRate: 0.002, RiseRate: 0.02,
			MinLevel: 0.0, MaxLevel: 1.0, HalfLife: 300 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoFatigueAccumulated: 0.15,
				EndoRestCompleted:      -0.3,
			},
		},
		Insulin: {
			BaselineLevel: 0.4, DecayRate: 0.01, RiseRate: 0.1,
			MinLevel: 0.1, MaxLevel: 0.9, HalfLife: 60 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{},
		},
		CytokineIL6: {
			BaselineLevel: 0.1, DecayRate: 0.005, RiseRate: 0.05,
			MinLevel: 0.0, MaxLevel: 1.0, HalfLife: 120 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoFatigueAccumulated: 0.1,
				EndoThreatDetected:     0.05,
			},
		},
		Anandamide: {
			BaselineLevel: 0.3, DecayRate: 0.02, RiseRate: 0.15,
			MinLevel: 0.0, MaxLevel: 1.0, HalfLife: 30 * time.Second,
			SensitivityMap: map[EndocrineEvent]float64{
				EndoInsightGained:  0.4,
				EndoGoalAchieved:   0.3,
				EndoRestCompleted:  0.2,
				EndoSocialContact:  0.15,
			},
		},
	}
}

// ValenceMemoryEntry records a valence-tagged experience
type ValenceMemoryEntry struct {
	Timestamp time.Time
	Valence   float64 // [-1, +1] negative to positive
	Arousal   float64 // [0, 1] calm to excited
	Dominance float64 // [0, 1] submissive to dominant
	Context   string
	Hormones  [int(HormoneCount)]float64
	Mode      CognitiveMode
}

// HormoneBus is the central hormone concentration bus
type HormoneBus struct {
	mu            sync.RWMutex
	concentrations [int(HormoneCount)]float64
	configs        map[Hormone]GlandConfig
	lastUpdate     time.Time
	currentMode    CognitiveMode
	modeConfidence float64
}

// NewHormoneBus creates a new hormone bus with default configurations
func NewHormoneBus() *HormoneBus {
	configs := DefaultGlandConfigs()
	bus := &HormoneBus{
		configs:    configs,
		lastUpdate: time.Now(),
	}
	// Initialize to baseline levels
	for h := Hormone(0); h < HormoneCount; h++ {
		if cfg, ok := configs[h]; ok {
			bus.concentrations[h] = cfg.BaselineLevel
		}
	}
	return bus
}

// Concentration returns the current concentration of a hormone
func (hb *HormoneBus) Concentration(h Hormone) float64 {
	hb.mu.RLock()
	defer hb.mu.RUnlock()
	return hb.concentrations[h]
}

// AllConcentrations returns a copy of all hormone concentrations
func (hb *HormoneBus) AllConcentrations() [int(HormoneCount)]float64 {
	hb.mu.RLock()
	defer hb.mu.RUnlock()
	return hb.concentrations
}

// CurrentMode returns the current emergent cognitive mode and its confidence
func (hb *HormoneBus) CurrentMode() (CognitiveMode, float64) {
	hb.mu.RLock()
	defer hb.mu.RUnlock()
	return hb.currentMode, hb.modeConfidence
}

// VirtualEndocrineSystem implements the 10-gland virtual endocrine system
type VirtualEndocrineSystem struct {
	mu sync.RWMutex

	// Core hormone bus
	Bus *HormoneBus

	// Valence memory
	valenceMemory    []ValenceMemoryEntry
	maxValenceMemory int

	// Current affective state
	currentValence  float64
	currentArousal  float64
	currentDominance float64

	// Mode transition callbacks
	onModeChange func(from, to CognitiveMode)

	// Metrics
	totalEvents     uint64
	modeTransitions uint64
	lastEventTime   time.Time
}

// NewVirtualEndocrineSystem creates a new endocrine system
func NewVirtualEndocrineSystem() *VirtualEndocrineSystem {
	return &VirtualEndocrineSystem{
		Bus:              NewHormoneBus(),
		valenceMemory:    make([]ValenceMemoryEntry, 0, 1000),
		maxValenceMemory: 1000,
		lastEventTime:    time.Now(),
	}
}

// SetModeChangeCallback registers a callback for cognitive mode transitions
func (ves *VirtualEndocrineSystem) SetModeChangeCallback(cb func(from, to CognitiveMode)) {
	ves.mu.Lock()
	defer ves.mu.Unlock()
	ves.onModeChange = cb
}

// SignalEvent sends an event to the endocrine system, triggering hormone responses
func (ves *VirtualEndocrineSystem) SignalEvent(event EndocrineEvent, intensity float64) {
	ves.mu.Lock()
	defer ves.mu.Unlock()

	ves.totalEvents++
	ves.lastEventTime = time.Now()

	// Clamp intensity
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1 {
		intensity = 1
	}

	ves.Bus.mu.Lock()
	for h := Hormone(0); h < HormoneCount; h++ {
		cfg, ok := ves.Bus.configs[h]
		if !ok {
			continue
		}
		sensitivity, hasSensitivity := cfg.SensitivityMap[event]
		if !hasSensitivity {
			continue
		}

		delta := sensitivity * intensity
		newLevel := ves.Bus.concentrations[h] + delta
		if newLevel < cfg.MinLevel {
			newLevel = cfg.MinLevel
		}
		if newLevel > cfg.MaxLevel {
			newLevel = cfg.MaxLevel
		}
		ves.Bus.concentrations[h] = newLevel
	}
	ves.Bus.mu.Unlock()
}

// Tick updates the endocrine system by one time step
func (ves *VirtualEndocrineSystem) Tick(dt float64) {
	ves.mu.Lock()
	defer ves.mu.Unlock()

	ves.Bus.mu.Lock()

	// Decay all hormones toward baseline
	for h := Hormone(0); h < HormoneCount; h++ {
		cfg, ok := ves.Bus.configs[h]
		if !ok {
			continue
		}
		current := ves.Bus.concentrations[h]
		baseline := cfg.BaselineLevel
		decay := cfg.DecayRate * dt

		if current > baseline {
			current -= decay
			if current < baseline {
				current = baseline
			}
		} else if current < baseline {
			current += decay * 0.5 // Slower recovery than decay
			if current > baseline {
				current = baseline
			}
		}

		if current < cfg.MinLevel {
			current = cfg.MinLevel
		}
		if current > cfg.MaxLevel {
			current = cfg.MaxLevel
		}
		ves.Bus.concentrations[h] = current
	}

	// Compute affective state (valence-arousal-dominance)
	ves.computeAffectiveState()

	// Detect cognitive mode
	oldMode := ves.Bus.currentMode
	ves.detectCognitiveMode()
	newMode := ves.Bus.currentMode

	ves.Bus.lastUpdate = time.Now()
	ves.Bus.mu.Unlock()

	// Fire mode change callback
	if oldMode != newMode {
		ves.modeTransitions++
		if ves.onModeChange != nil {
			ves.onModeChange(oldMode, newMode)
		}
	}
}

// computeAffectiveState derives valence, arousal, dominance from hormone concentrations
func (ves *VirtualEndocrineSystem) computeAffectiveState() {
	c := ves.Bus.concentrations

	// Valence: positive hormones minus negative hormones
	// Positive: dopamine, serotonin, oxytocin, anandamide
	// Negative: cortisol, cytokine
	positive := c[DopamineTonic]*0.25 + c[DopaminePhasic]*0.3 +
		c[Serotonin]*0.2 + c[Oxytocin]*0.15 + c[Anandamide]*0.1
	negative := c[Cortisol]*0.5 + c[CytokineIL6]*0.3 + (1.0-c[Serotonin])*0.2
	ves.currentValence = clampF64(positive-negative, -1.0, 1.0)

	// Arousal: activating hormones
	ves.currentArousal = clampF64(
		c[Norepinephrine]*0.35+c[Cortisol]*0.25+c[DopaminePhasic]*0.2+
			(1.0-c[Melatonin])*0.1+(1.0-c[Anandamide])*0.1,
		0.0, 1.0,
	)

	// Dominance: agency-related hormones
	ves.currentDominance = clampF64(
		c[DopamineTonic]*0.3+c[Serotonin]*0.25+(1.0-c[Cortisol])*0.25+
			c[Norepinephrine]*0.1+c[Anandamide]*0.1,
		0.0, 1.0,
	)
}

// detectCognitiveMode determines the emergent cognitive mode from hormone patterns
func (ves *VirtualEndocrineSystem) detectCognitiveMode() {
	c := ves.Bus.concentrations

	type modeScore struct {
		mode  CognitiveMode
		score float64
	}

	scores := []modeScore{
		{ModeExplore, c[Norepinephrine]*0.4 + c[DopaminePhasic]*0.4 + (1.0-c[Cortisol])*0.2},
		{ModeExploit, c[DopamineTonic]*0.4 + c[Serotonin]*0.4 + (1.0-c[Norepinephrine])*0.2},
		{ModeFight, c[Cortisol]*0.3 + c[Norepinephrine]*0.4 + (1.0-c[Serotonin])*0.3},
		{ModeFlight, c[Cortisol]*0.4 + (1.0-c[Norepinephrine])*0.3 + (1.0-c[DopamineTonic])*0.3},
		{ModeFreeze, c[Cortisol]*0.4 + (1.0-c[DopamineTonic])*0.3 + (1.0-c[Norepinephrine])*0.3},
		{ModeFlow, c[Anandamide]*0.35 + c[Norepinephrine]*0.25 + c[DopamineTonic]*0.2 + (1.0-c[Cortisol])*0.2},
		{ModeSocial, c[Oxytocin]*0.5 + (1.0-c[Cortisol])*0.3 + c[Serotonin]*0.2},
		{ModeRest, c[Melatonin]*0.4 + (1.0-c[Norepinephrine])*0.3 + c[Serotonin]*0.3},
		{ModeCreative, c[DopaminePhasic]*0.35 + c[Anandamide]*0.35 + (1.0-c[Serotonin])*0.15 + c[Norepinephrine]*0.15},
		{ModeAnalytical, c[Serotonin]*0.35 + c[Norepinephrine]*0.25 + c[DopamineTonic]*0.2 + (1.0-c[DopaminePhasic])*0.2},
	}

	bestMode := ModeExploit
	bestScore := 0.0
	for _, ms := range scores {
		if ms.score > bestScore {
			bestScore = ms.score
			bestMode = ms.mode
		}
	}

	ves.Bus.currentMode = bestMode
	ves.Bus.modeConfidence = bestScore
}

// RecordValenceMemory stores a valence-tagged experience
func (ves *VirtualEndocrineSystem) RecordValenceMemory(context string) {
	ves.mu.Lock()
	defer ves.mu.Unlock()

	entry := ValenceMemoryEntry{
		Timestamp: time.Now(),
		Valence:   ves.currentValence,
		Arousal:   ves.currentArousal,
		Dominance: ves.currentDominance,
		Context:   context,
		Hormones:  ves.Bus.concentrations,
		Mode:      ves.Bus.currentMode,
	}

	ves.valenceMemory = append(ves.valenceMemory, entry)
	if len(ves.valenceMemory) > ves.maxValenceMemory {
		ves.valenceMemory = ves.valenceMemory[1:]
	}
}

// GetAffectiveState returns the current valence, arousal, dominance
func (ves *VirtualEndocrineSystem) GetAffectiveState() (valence, arousal, dominance float64) {
	ves.mu.RLock()
	defer ves.mu.RUnlock()
	return ves.currentValence, ves.currentArousal, ves.currentDominance
}

// GetValenceHistory returns recent valence memory entries
func (ves *VirtualEndocrineSystem) GetValenceHistory(n int) []ValenceMemoryEntry {
	ves.mu.RLock()
	defer ves.mu.RUnlock()
	if n > len(ves.valenceMemory) {
		n = len(ves.valenceMemory)
	}
	start := len(ves.valenceMemory) - n
	result := make([]ValenceMemoryEntry, n)
	copy(result, ves.valenceMemory[start:])
	return result
}

// GetMetrics returns endocrine system metrics
func (ves *VirtualEndocrineSystem) GetMetrics() map[string]interface{} {
	ves.mu.RLock()
	defer ves.mu.RUnlock()
	mode, confidence := ves.Bus.CurrentMode()
	return map[string]interface{}{
		"total_events":      ves.totalEvents,
		"mode_transitions":  ves.modeTransitions,
		"current_mode":      mode.String(),
		"mode_confidence":   confidence,
		"current_valence":   ves.currentValence,
		"current_arousal":   ves.currentArousal,
		"current_dominance": ves.currentDominance,
		"valence_memory_size": len(ves.valenceMemory),
	}
}

// MoralPerceptionSignal represents a pre-deliberative moral sensing signal
type MoralPerceptionSignal struct {
	RawAffect        float64 // Immediate felt-sense [-1, +1]
	MoralAssociation string  // Associated moral concept
	EmpathicInference float64 // Empathic resonance strength [0, 1]
	NoveltySignal    float64 // How novel is this moral situation [0, 1]
}

// MoralPerception evaluates a situation through the moral perception engine
func (ves *VirtualEndocrineSystem) MoralPerception(situation string) MoralPerceptionSignal {
	ves.mu.RLock()
	defer ves.mu.RUnlock()

	c := ves.Bus.concentrations

	// Raw affect is current valence modulated by empathy (oxytocin)
	rawAffect := ves.currentValence * (0.5 + c[Oxytocin]*0.5)

	// Empathic inference scales with oxytocin and inversely with cortisol
	empathic := c[Oxytocin]*0.6 + c[Serotonin]*0.3 - c[Cortisol]*0.3
	empathic = clampF64(empathic, 0.0, 1.0)

	// Novelty signal from norepinephrine and phasic dopamine
	novelty := c[Norepinephrine]*0.5 + c[DopaminePhasic]*0.5
	novelty = clampF64(novelty, 0.0, 1.0)

	// Determine moral association based on dominant hormone pattern
	association := "neutral"
	if c[Oxytocin] > 0.6 {
		association = "care"
	} else if c[Cortisol] > 0.6 {
		association = "harm_aversion"
	} else if c[Serotonin] > 0.6 {
		association = "fairness"
	} else if c[DopamineTonic] > 0.6 {
		association = "loyalty"
	}

	return MoralPerceptionSignal{
		RawAffect:        rawAffect,
		MoralAssociation: association,
		EmpathicInference: empathic,
		NoveltySignal:    novelty,
	}
}

// clampF64 clamps a float64 value between min and max
func clampF64(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Ensure math is used (for future extensions)
var _ = math.Abs
