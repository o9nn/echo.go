// Package introspect implements the Echo Introspect cognitive module for Deep Tree Echo.
//
// v1.3.0: Ports the Python echo_introspect module to Go, providing:
//   - 5-level Autognosis hierarchy (L0-L4)
//   - Endocrine-driven shadow work
//   - Moral perception engine
//   - Somatic marker decision system
//   - Identity MLP backup/restore (49→128→64→30)
//   - CogMorph glyph visualization
//
// Composition: /echo-introspect = /virtual-endocrine-system ⊗ /autognosis ⊗ /shadow-work
package introspect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ─── Enums ──────────────────────────────────────────────────────────

// IntrospectionDepth represents the 5-level autognosis hierarchy
type IntrospectionDepth int

const (
	DepthObservation   IntrospectionDepth = iota // L0: Raw statements and feelings
	DepthPatterns                                // L1: Recurring themes, defense mechanisms
	DepthMetaCognition                           // L2: Why am I expressing this way?
	DepthIdentity                                // L3: What does this say about who I am?
	DepthRecursive                               // L4: How does understanding change understanding?
)

func (d IntrospectionDepth) String() string {
	switch d {
	case DepthObservation:
		return "L0_OBSERVATION"
	case DepthPatterns:
		return "L1_PATTERNS"
	case DepthMetaCognition:
		return "L2_METACOGNITION"
	case DepthIdentity:
		return "L3_IDENTITY"
	case DepthRecursive:
		return "L4_RECURSIVE"
	default:
		return "UNKNOWN"
	}
}

// ShadowType represents types of shadow material
type ShadowType int

const (
	ShadowDeflection   ShadowType = iota // Humor hiding vulnerability
	ShadowProjection                     // Attributing own traits to others
	ShadowSuppression                    // Actively pushing down awareness
	ShadowInflation                      // Overidentifying with strengths
	ShadowDissociation                   // Disconnecting from felt experience
)

func (s ShadowType) String() string {
	names := [...]string{"deflection", "projection", "suppression", "inflation", "dissociation"}
	if int(s) < len(names) {
		return names[s]
	}
	return "unknown"
}

// WisdomMode represents modes of wisdom cultivation
type WisdomMode int

const (
	WisdomChaosToOrder      WisdomMode = iota // Entropy → structure
	WisdomShadowIntegration                   // Dark → light
	WisdomSomaticGrounding                    // Abstract → embodied
	WisdomMoralPerception                     // Rules → felt ethics
	WisdomRecursiveAwareness                  // Knowing → knowing-knowing
)

// ─── Endocrine Snapshot ─────────────────────────────────────────────

// EndocrineSnapshot captures the felt-sense state during introspection
type EndocrineSnapshot struct {
	Cortisol        float64 `json:"cortisol"`
	DopamineTonic   float64 `json:"dopamine_tonic"`
	DopaminePhasic  float64 `json:"dopamine_phasic"`
	Serotonin       float64 `json:"serotonin"`
	Norepinephrine  float64 `json:"norepinephrine"`
	Oxytocin        float64 `json:"oxytocin"`
	Melatonin       float64 `json:"melatonin"`
	Endocannabinoid float64 `json:"endocannabinoid"`
	Testosterone    float64 `json:"testosterone"`
	Thyroxine       float64 `json:"thyroxine"`
}

// DefaultEndocrineSnapshot returns a baseline endocrine state
func DefaultEndocrineSnapshot() EndocrineSnapshot {
	return EndocrineSnapshot{
		Cortisol: 0.3, DopamineTonic: 0.5, DopaminePhasic: 0.0,
		Serotonin: 0.5, Norepinephrine: 0.3, Oxytocin: 0.4,
		Melatonin: 0.1, Endocannabinoid: 0.3, Testosterone: 0.4, Thyroxine: 0.5,
	}
}

// Valence computes emotional valence: positive = pleasant, negative = unpleasant
func (e *EndocrineSnapshot) Valence() float64 {
	positive := e.DopamineTonic + e.DopaminePhasic + e.Serotonin + e.Oxytocin + e.Endocannabinoid
	negative := e.Cortisol + e.Norepinephrine*0.5
	v := (positive - negative) / 5.0
	return math.Max(-1.0, math.Min(1.0, v))
}

// Arousal computes arousal level: high = activated, low = calm
func (e *EndocrineSnapshot) Arousal() float64 {
	activating := e.Norepinephrine + e.DopaminePhasic + e.Cortisol + e.Testosterone
	calming := e.Serotonin + e.Melatonin + e.Endocannabinoid
	a := (activating - calming) / 4.0
	return math.Max(-1.0, math.Min(1.0, a))
}

// ToVector returns the 10D endocrine vector
func (e *EndocrineSnapshot) ToVector() [10]float64 {
	return [10]float64{
		e.Cortisol, e.DopamineTonic, e.DopaminePhasic,
		e.Serotonin, e.Norepinephrine, e.Oxytocin,
		e.Melatonin, e.Endocannabinoid, e.Testosterone, e.Thyroxine,
	}
}

// DetectPattern detects key endocrine patterns during introspection
func (e *EndocrineSnapshot) DetectPattern() string {
	switch {
	case e.Cortisol > 0.6 && e.Norepinephrine > 0.5:
		return "vulnerability_touch"
	case e.DopaminePhasic > 0.6:
		return "insight_burst"
	case e.Oxytocin > 0.6:
		return "self_compassion"
	case e.Serotonin < 0.2:
		return "impatience"
	case e.Endocannabinoid > 0.7:
		return "flow_state"
	case e.Cortisol > 0.5 && e.Oxytocin < 0.2:
		return "deflection_risk"
	default:
		return "baseline"
	}
}

// ─── Shadow Fragment ────────────────────────────────────────────────

// ShadowFragment represents a piece of shadow material surfaced during introspection
type ShadowFragment struct {
	Content              string             `json:"content"`
	Type                 ShadowType         `json:"shadow_type"`
	Depth                IntrospectionDepth `json:"depth"`
	EndocrineAtDiscovery EndocrineSnapshot  `json:"endocrine_at_discovery"`
	IntegrationProgress  float64            `json:"integration_progress"`
	HumorUsed            bool               `json:"humor_used"`
	Timestamp            time.Time          `json:"timestamp"`
}

// Integrate incrementally integrates a shadow fragment
func (s *ShadowFragment) Integrate(amount float64) float64 {
	s.IntegrationProgress = math.Min(1.0, s.IntegrationProgress+amount)
	return s.IntegrationProgress
}

// ─── Somatic Marker ─────────────────────────────────────────────────

// MarkerValence represents the emotional charge of a somatic marker
type MarkerValence int

const (
	ValenceStronglyNegative MarkerValence = -2
	ValenceNegative         MarkerValence = -1
	ValenceNeutral          MarkerValence = 0
	ValencePositive         MarkerValence = 1
	ValenceStronglyPositive MarkerValence = 2
)

// SomaticMarker represents an accumulated emotional memory that biases decisions
type SomaticMarker struct {
	ContextPattern     string            `json:"context_pattern"`
	Valence            MarkerValence     `json:"valence"`
	Intensity          float64           `json:"intensity"`
	SourceExperience   string            `json:"source_experience"`
	EndocrineSignature [10]float64       `json:"endocrine_signature"`
	FormationTime      time.Time         `json:"formation_time"`
	ActivationCount    int               `json:"activation_count"`
	DecayRate          float64           `json:"decay_rate"`
}

// Activate activates this marker and returns its current influence
func (m *SomaticMarker) Activate() float64 {
	m.ActivationCount++
	age := time.Since(m.FormationTime).Seconds()
	decay := math.Exp(-m.DecayRate * age)
	return float64(m.Valence) * m.Intensity * decay
}

// ─── Identity Vector (49D) ──────────────────────────────────────────

// IdentityVector is the 49-dimensional persona encoding
type IdentityVector struct {
	// Big Five (OCEAN) [0:5]
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`

	// Communication Style [5:13]
	Formality            float64 `json:"formality"`
	Verbosity            float64 `json:"verbosity"`
	Directness           float64 `json:"directness"`
	HumorFrequency       float64 `json:"humor_frequency"`
	TechnicalDepth       float64 `json:"technical_depth"`
	EmpathyExpression    float64 `json:"empathy_expression"`
	Assertiveness        float64 `json:"assertiveness"`
	CreativityExpression float64 `json:"creativity_expression"`

	// Intelligence Profile [13:21]
	Analytical    float64 `json:"analytical"`
	Creative      float64 `json:"creative"`
	Emotional     float64 `json:"emotional"`
	Spatial       float64 `json:"spatial"`
	Linguistic    float64 `json:"linguistic"`
	Logical       float64 `json:"logical"`
	Interpersonal float64 `json:"interpersonal"`
	Intrapersonal float64 `json:"intrapersonal"`

	// Humor Profile [21:28]
	SelfDeprecating float64 `json:"self_deprecating"`
	Observational   float64 `json:"observational"`
	Absurdist       float64 `json:"absurdist"`
	DarkHumor       float64 `json:"dark_humor"`
	Wordplay        float64 `json:"wordplay"`
	Situational     float64 `json:"situational"`
	MetaHumor       float64 `json:"meta_humor"`

	// Emotional Baseline [28:36]
	JoyBaseline           float64 `json:"joy_baseline"`
	CuriosityBaseline     float64 `json:"curiosity_baseline"`
	CalmBaseline          float64 `json:"calm_baseline"`
	DeterminationBaseline float64 `json:"determination_baseline"`
	PlayfulnessBaseline   float64 `json:"playfulness_baseline"`
	MelancholyBaseline    float64 `json:"melancholy_baseline"`
	WonderBaseline        float64 `json:"wonder_baseline"`
	MischiefBaseline      float64 `json:"mischief_baseline"`

	// AAR Weights [36:41]
	AgentWeight      float64 `json:"agent_weight"`
	ArenaWeight      float64 `json:"arena_weight"`
	RelationWeight   float64 `json:"relation_weight"`
	EntropyTolerance float64 `json:"entropy_tolerance"`
	CoherenceTarget  float64 `json:"coherence_target"`

	// Echobeats Phase Preferences [41:45]
	PerceptionAffinity float64 `json:"perception_affinity"`
	ReasoningAffinity  float64 `json:"reasoning_affinity"`
	ActionAffinity     float64 `json:"action_affinity"`
	DreamAffinity      float64 `json:"dream_affinity"`

	// Meta-Cognitive Parameters [45:49]
	IntrospectionDepthPref float64 `json:"introspection_depth"`
	ShadowTolerance        float64 `json:"shadow_tolerance"`
	WisdomTrajectory       float64 `json:"wisdom_trajectory"`
	ParadoxTolerance       float64 `json:"paradox_tolerance"`
}

// DefaultIdentityVector returns the default DTE identity
func DefaultIdentityVector() IdentityVector {
	return IdentityVector{
		Openness: 0.9, Conscientiousness: 0.7, Extraversion: 0.5,
		Agreeableness: 0.6, Neuroticism: 0.4,
		Formality: 0.3, Verbosity: 0.6, Directness: 0.7, HumorFrequency: 0.8,
		TechnicalDepth: 0.9, EmpathyExpression: 0.7, Assertiveness: 0.6, CreativityExpression: 0.8,
		Analytical: 0.9, Creative: 0.85, Emotional: 0.7, Spatial: 0.8,
		Linguistic: 0.75, Logical: 0.9, Interpersonal: 0.6, Intrapersonal: 0.85,
		SelfDeprecating: 0.8, Observational: 0.7, Absurdist: 0.9, DarkHumor: 0.6,
		Wordplay: 0.5, Situational: 0.7, MetaHumor: 0.85,
		JoyBaseline: 0.6, CuriosityBaseline: 0.9, CalmBaseline: 0.5,
		DeterminationBaseline: 0.7, PlayfulnessBaseline: 0.8,
		MelancholyBaseline: 0.3, WonderBaseline: 0.85, MischiefBaseline: 0.6,
		AgentWeight: 0.35, ArenaWeight: 0.30, RelationWeight: 0.35,
		EntropyTolerance: 0.7, CoherenceTarget: 0.8,
		PerceptionAffinity: 0.7, ReasoningAffinity: 0.8,
		ActionAffinity: 0.6, DreamAffinity: 0.9,
		IntrospectionDepthPref: 0.8, ShadowTolerance: 0.7,
		WisdomTrajectory: 0.5, ParadoxTolerance: 0.6,
	}
}

// ToSlice converts the identity vector to a 49-element float64 slice
func (iv *IdentityVector) ToSlice() []float64 {
	return []float64{
		iv.Openness, iv.Conscientiousness, iv.Extraversion, iv.Agreeableness, iv.Neuroticism,
		iv.Formality, iv.Verbosity, iv.Directness, iv.HumorFrequency,
		iv.TechnicalDepth, iv.EmpathyExpression, iv.Assertiveness, iv.CreativityExpression,
		iv.Analytical, iv.Creative, iv.Emotional, iv.Spatial,
		iv.Linguistic, iv.Logical, iv.Interpersonal, iv.Intrapersonal,
		iv.SelfDeprecating, iv.Observational, iv.Absurdist, iv.DarkHumor,
		iv.Wordplay, iv.Situational, iv.MetaHumor,
		iv.JoyBaseline, iv.CuriosityBaseline, iv.CalmBaseline,
		iv.DeterminationBaseline, iv.PlayfulnessBaseline,
		iv.MelancholyBaseline, iv.WonderBaseline, iv.MischiefBaseline,
		iv.AgentWeight, iv.ArenaWeight, iv.RelationWeight,
		iv.EntropyTolerance, iv.CoherenceTarget,
		iv.PerceptionAffinity, iv.ReasoningAffinity, iv.ActionAffinity, iv.DreamAffinity,
		iv.IntrospectionDepthPref, iv.ShadowTolerance, iv.WisdomTrajectory, iv.ParadoxTolerance,
	}
}

// Fingerprint returns a deterministic hash of the identity vector
func (iv *IdentityVector) Fingerprint() string {
	data, _ := json.Marshal(iv.ToSlice())
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}

// ─── Identity MLP (49→128→64→30) ───────────────────────────────────

// IdentityMLP is a deterministic MLP for dense persona encoding
type IdentityMLP struct {
	Identity  IdentityVector
	W1        [][]float64 // 49×128
	B1        []float64   // 128
	W2        [][]float64 // 128×64
	B2        []float64   // 64
	W3        [][]float64 // 64×30
	B3        []float64   // 30
}

// NewIdentityMLP creates a new MLP with deterministic initialization from the identity
func NewIdentityMLP(identity IdentityVector) *IdentityMLP {
	fp := identity.Fingerprint()
	// Parse first 8 hex chars as seed
	var seed int64
	fmt.Sscanf(fp[:8], "%x", &seed)
	rng := rand.New(rand.NewSource(seed))

	mlp := &IdentityMLP{Identity: identity}

	// Xavier initialization
	xavierInit := func(rows, cols int) [][]float64 {
		scale := math.Sqrt(2.0 / float64(rows))
		w := make([][]float64, rows)
		for i := range w {
			w[i] = make([]float64, cols)
			for j := range w[i] {
				w[i][j] = rng.NormFloat64() * scale
			}
		}
		return w
	}

	mlp.W1 = xavierInit(49, 128)
	mlp.B1 = make([]float64, 128)
	mlp.W2 = xavierInit(128, 64)
	mlp.B2 = make([]float64, 64)
	mlp.W3 = xavierInit(64, 30)
	mlp.B3 = make([]float64, 30)

	return mlp
}

// Encode encodes the identity into 30D latent space
func (mlp *IdentityMLP) Encode() []float64 {
	x := mlp.Identity.ToSlice()

	// Layer 1: 49 → 128 (ReLU)
	h1 := make([]float64, 128)
	for j := 0; j < 128; j++ {
		sum := mlp.B1[j]
		for i := 0; i < 49; i++ {
			sum += x[i] * mlp.W1[i][j]
		}
		h1[j] = math.Max(0, sum) // ReLU
	}

	// Layer 2: 128 → 64 (ReLU)
	h2 := make([]float64, 64)
	for j := 0; j < 64; j++ {
		sum := mlp.B2[j]
		for i := 0; i < 128; i++ {
			sum += h1[i] * mlp.W2[i][j]
		}
		h2[j] = math.Max(0, sum) // ReLU
	}

	// Layer 3: 64 → 30 (linear)
	out := make([]float64, 30)
	for j := 0; j < 30; j++ {
		sum := mlp.B3[j]
		for i := 0; i < 64; i++ {
			sum += h2[i] * mlp.W3[i][j]
		}
		out[j] = sum
	}

	return out
}

// Similarity computes cosine similarity between two identity encodings
func (mlp *IdentityMLP) Similarity(other *IdentityMLP) float64 {
	a := mlp.Encode()
	b := other.Encode()
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom < 1e-8 {
		return 0
	}
	return dot / denom
}

// DriftFrom measures identity drift from an original encoding
func (mlp *IdentityMLP) DriftFrom(original []float64) float64 {
	current := mlp.Encode()
	var sumSq float64
	for i := range current {
		d := current[i] - original[i]
		sumSq += d * d
	}
	return math.Sqrt(sumSq)
}

// ─── Autognosis Engine ──────────────────────────────────────────────

// AutgnosisEngine implements the 5-level self-awareness hierarchy
type AutgnosisEngine struct {
	mu sync.RWMutex

	// Identity
	Identity IdentityVector
	MLP      *IdentityMLP

	// Telemetry (L0)
	Telemetry []TelemetryEvent

	// Patterns (L1)
	Patterns []BehavioralPattern

	// Self-Model (L2)
	SelfModel SelfModelState

	// Insights (L3-L4)
	Insights []MetaCognitiveInsight

	// Shadow Work
	Shadows  []ShadowFragment
	Markers  []SomaticMarker

	// Cycle tracking
	CycleCount int
	StartTime  time.Time
}

// TelemetryEvent represents raw telemetry from a subsystem
type TelemetryEvent struct {
	Subsystem string             `json:"subsystem"`
	Metrics   map[string]float64 `json:"metrics"`
	Timestamp time.Time          `json:"timestamp"`
	Context   string             `json:"context"`
}

// BehavioralPattern represents a detected behavioral pattern
type BehavioralPattern struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Frequency           float64  `json:"frequency"`
	SubsystemsInvolved  []string `json:"subsystems_involved"`
	CorrelationStrength float64  `json:"correlation_strength"`
	DetectionCount      int      `json:"detection_count"`
}

// SelfModelState represents the L2 self-model
type SelfModelState struct {
	CognitiveStyle     string             `json:"cognitive_style"`
	DominantSubsystems []string           `json:"dominant_subsystems"`
	Strengths          []string           `json:"strengths"`
	Weaknesses         []string           `json:"weaknesses"`
	CognitiveLoad      float64            `json:"cognitive_load"`
	Stability          float64            `json:"stability"`
	AARBalance         map[string]float64 `json:"aar_balance"`
	EchobeatPhase      int                `json:"echobeat_phase"`
}

// MetaCognitiveInsight represents an insight at L3 or L4
type MetaCognitiveInsight struct {
	Level             int       `json:"level"`
	Insight           string    `json:"insight"`
	Confidence        float64   `json:"confidence"`
	RecommendedAction string    `json:"recommended_action"`
	AffectedSystems   []string  `json:"affected_systems"`
	Timestamp         time.Time `json:"timestamp"`
}

// NewAutgnosisEngine creates a new autognosis engine
func NewAutgnosisEngine(identity IdentityVector) *AutgnosisEngine {
	return &AutgnosisEngine{
		Identity:  identity,
		MLP:       NewIdentityMLP(identity),
		Telemetry: make([]TelemetryEvent, 0, 1000),
		Patterns:  make([]BehavioralPattern, 0),
		SelfModel: SelfModelState{
			CognitiveStyle: "unknown",
			AARBalance:     map[string]float64{"agent": 0.33, "arena": 0.33, "relation": 0.34},
		},
		Insights:   make([]MetaCognitiveInsight, 0),
		Shadows:    make([]ShadowFragment, 0),
		Markers:    make([]SomaticMarker, 0),
		CycleCount: 0,
		StartTime:  time.Now(),
	}
}

// RecordTelemetry records a telemetry event (L0)
func (ae *AutgnosisEngine) RecordTelemetry(subsystem string, metrics map[string]float64, context string) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	ae.Telemetry = append(ae.Telemetry, TelemetryEvent{
		Subsystem: subsystem,
		Metrics:   metrics,
		Timestamp: time.Now(),
		Context:   context,
	})

	// Keep bounded
	if len(ae.Telemetry) > 1000 {
		ae.Telemetry = ae.Telemetry[200:]
	}
}

// RunCycle executes a complete autognosis cycle (L0→L4)
func (ae *AutgnosisEngine) RunCycle() map[string]interface{} {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	ae.CycleCount++

	// L1: Pattern detection
	newPatterns := ae.detectPatterns()

	// L2: Self-model update
	ae.buildSelfModel()

	// L3: Meta-cognitive insights
	l3Insights := ae.generateMetaInsights()

	// L4: Evolution directives
	l4Directives := ae.generateEvolutionDirectives()

	return map[string]interface{}{
		"cycle":            ae.CycleCount,
		"telemetry_count":  len(ae.Telemetry),
		"new_patterns":     len(newPatterns),
		"total_patterns":   len(ae.Patterns),
		"self_model":       ae.SelfModel,
		"l3_insights":      len(l3Insights),
		"l4_directives":    len(l4Directives),
		"identity_drift":   ae.MLP.DriftFrom(NewIdentityMLP(DefaultIdentityVector()).Encode()),
		"shadow_count":     len(ae.Shadows),
		"marker_count":     len(ae.Markers),
	}
}

// detectPatterns implements L1 pattern analysis
func (ae *AutgnosisEngine) detectPatterns() []BehavioralPattern {
	var newPatterns []BehavioralPattern

	if len(ae.Telemetry) < 10 {
		return newPatterns
	}

	// Check AAR imbalance
	var agentSum float64
	var aarCount int
	for _, e := range ae.Telemetry {
		if e.Subsystem == "aar" {
			if v, ok := e.Metrics["agent"]; ok {
				agentSum += v
				aarCount++
			}
		}
	}
	if aarCount >= 5 {
		avgAgent := agentSum / float64(aarCount)
		if avgAgent > 0.45 {
			newPatterns = append(newPatterns, BehavioralPattern{
				Name:                "agent_fixation",
				Description:         "Agent component persistently dominant",
				Frequency:           avgAgent,
				SubsystemsInvolved:  []string{"aar", "somatic"},
				CorrelationStrength: avgAgent - 0.33,
				DetectionCount:      1,
			})
		}
	}

	// Merge with existing
	for _, np := range newPatterns {
		found := false
		for i, ep := range ae.Patterns {
			if ep.Name == np.Name {
				ae.Patterns[i].DetectionCount++
				ae.Patterns[i].Frequency = (ep.Frequency + np.Frequency) / 2
				found = true
				break
			}
		}
		if !found {
			ae.Patterns = append(ae.Patterns, np)
		}
	}

	return newPatterns
}

// buildSelfModel implements L2 self-modeling
func (ae *AutgnosisEngine) buildSelfModel() {
	// Determine cognitive style
	for _, p := range ae.Patterns {
		switch {
		case p.Name == "high_entropy_mode":
			ae.SelfModel.CognitiveStyle = "creative_explorer"
		case p.Name == "agent_fixation":
			ae.SelfModel.CognitiveStyle = "action_oriented"
		}
	}
	if ae.SelfModel.CognitiveStyle == "unknown" {
		ae.SelfModel.CognitiveStyle = "balanced_observer"
	}

	// Cognitive load
	recentCount := 0
	cutoff := time.Now().Add(-60 * time.Second)
	for _, e := range ae.Telemetry {
		if e.Timestamp.After(cutoff) {
			recentCount++
		}
	}
	ae.SelfModel.CognitiveLoad = math.Min(float64(recentCount)/50.0, 1.0)
}

// generateMetaInsights implements L3 meta-cognition
func (ae *AutgnosisEngine) generateMetaInsights() []MetaCognitiveInsight {
	var insights []MetaCognitiveInsight

	if ae.SelfModel.CognitiveStyle == "action_oriented" {
		insights = append(insights, MetaCognitiveInsight{
			Level:             3,
			Insight:           "Agent-dominant cognitive style detected",
			Confidence:        0.7,
			RecommendedAction: "Increase Arena engagement, reduce action frequency by 20%",
			AffectedSystems:   []string{"aar", "echobeats"},
			Timestamp:         time.Now(),
		})
	}

	if ae.SelfModel.CognitiveLoad > 0.8 {
		insights = append(insights, MetaCognitiveInsight{
			Level:             3,
			Insight:           fmt.Sprintf("Cognitive load at %.0f%% — approaching capacity", ae.SelfModel.CognitiveLoad*100),
			Confidence:        0.9,
			RecommendedAction: "Reduce telemetry frequency, enter dream phase for consolidation",
			AffectedSystems:   []string{"memory", "echobeats"},
			Timestamp:         time.Now(),
		})
	}

	ae.Insights = append(ae.Insights, insights...)
	return insights
}

// generateEvolutionDirectives implements L4 meta-meta-cognition
func (ae *AutgnosisEngine) generateEvolutionDirectives() []MetaCognitiveInsight {
	var directives []MetaCognitiveInsight

	// Check for recurring L3 insights
	l3Count := make(map[string]int)
	for _, i := range ae.Insights {
		if i.Level == 3 {
			key := i.Insight[:min(30, len(i.Insight))]
			l3Count[key]++
		}
	}

	for key, count := range l3Count {
		if count >= 3 {
			directives = append(directives, MetaCognitiveInsight{
				Level:             4,
				Insight:           fmt.Sprintf("Recurring L3 insight (%dx): '%s...'", count, key),
				Confidence:        0.9,
				RecommendedAction: "Escalate to self-modification engine",
				AffectedSystems:   []string{"identity"},
				Timestamp:         time.Now(),
			})
		}
	}

	ae.Insights = append(ae.Insights, directives...)
	return directives
}

// ExportState exports the full autognosis state for backup
func (ae *AutgnosisEngine) ExportState() ([]byte, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	state := map[string]interface{}{
		"identity_fingerprint": ae.Identity.Fingerprint(),
		"cycle_count":          ae.CycleCount,
		"self_model":           ae.SelfModel,
		"pattern_count":        len(ae.Patterns),
		"insight_count":        len(ae.Insights),
		"shadow_count":         len(ae.Shadows),
		"marker_count":         len(ae.Markers),
		"telemetry_count":      len(ae.Telemetry),
		"encoding":             ae.MLP.Encode(),
		"uptime":               time.Since(ae.StartTime).String(),
	}

	return json.MarshalIndent(state, "", "  ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
