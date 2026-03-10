package metahuman

import (
	"github.com/o9nn/echo.go/core/deeptreeecho"
)

// EndocrineExpressionMapper maps virtual endocrine hormone concentrations
// to FACS Action Unit activations. Each hormone modulates specific AUs
// with biologically-grounded intensity formulas.
type EndocrineExpressionMapper struct {
	// Mapping table: hormone → list of (AU, scale factor)
	mappings map[deeptreeecho.Hormone][]auMapping
}

type auMapping struct {
	AU    ActionUnit
	Scale float64
}

// NewEndocrineExpressionMapper creates a mapper with the canonical
// hormone-to-AU mappings from the meta-echo-dna skill specification.
func NewEndocrineExpressionMapper() *EndocrineExpressionMapper {
	m := &EndocrineExpressionMapper{
		mappings: make(map[deeptreeecho.Hormone][]auMapping),
	}

	// Cortisol → worry/concern/stress
	m.mappings[deeptreeecho.Cortisol] = []auMapping{
		{AU4, 0.8},  // Brow Lowerer — worry
		{AU1, 0.5},  // Inner Brow Raise — distress
		{AU15, 0.4}, // Lip Corner Depress — sadness/stress
	}

	// Dopamine (phasic) → reward/surprise burst
	m.mappings[deeptreeecho.DopaminePhasic] = []auMapping{
		{AU12, 0.9}, // Lip Corner Pull — smile/reward
		{AU6, 0.7},  // Cheek Raise — genuine smile (Duchenne)
	}

	// Dopamine (tonic) → baseline contentment
	m.mappings[deeptreeecho.DopamineTonic] = []auMapping{
		{AU12, 0.3}, // Lip Corner Pull — baseline contentment
	}

	// Serotonin → warm contentment
	m.mappings[deeptreeecho.Serotonin] = []auMapping{
		{AU6, 0.4},  // Cheek Raise — warm contentment
		{AU12, 0.3}, // Lip Corner Pull — gentle smile
	}

	// Norepinephrine → alertness/attention
	m.mappings[deeptreeecho.Norepinephrine] = []auMapping{
		{AU5, 0.8},  // Upper Lid Raise — alertness/surprise
		{AU7, 0.5},  // Lid Tightener — focus/vigilance
		{AU20, 0.3}, // Lip Stretch — tension
	}

	// Oxytocin → social bonding/warmth
	m.mappings[deeptreeecho.Oxytocin] = []auMapping{
		{AU6, 0.6},  // Cheek Raise — warmth
		{AU12, 0.5}, // Lip Corner Pull — social smile
		{AU25, 0.3}, // Lips Part — openness
	}

	// Melatonin → drowsiness/rest
	m.mappings[deeptreeecho.Melatonin] = []auMapping{
		{AU43, 0.7}, // Eyes Closed — drowsiness
		{AU7, 0.4},  // Lid Tightener — heavy lids
	}

	// Cytokine IL-6 → inflammation/sickness behavior
	m.mappings[deeptreeecho.CytokineIL6] = []auMapping{
		{AU4, 0.5},  // Brow Lowerer — discomfort
		{AU10, 0.4}, // Upper Lip Raise — disgust/illness
	}

	// Anandamide → bliss/flow/relaxation
	m.mappings[deeptreeecho.Anandamide] = []auMapping{
		{AU6, 0.5},  // Cheek Raise — relaxed contentment
		{AU25, 0.3}, // Lips Part — relaxed jaw
	}

	return m
}

// MapToFACS applies all hormone concentrations to the FACS state.
// Hormone concentrations should be in [0, 1] range.
func (m *EndocrineExpressionMapper) MapToFACS(concentrations map[deeptreeecho.Hormone]float64, facs *FACSState) {
	for hormone, auMaps := range m.mappings {
		conc, ok := concentrations[hormone]
		if !ok {
			continue
		}
		for _, am := range auMaps {
			facs.Add(am.AU, conc*am.Scale)
		}
	}
}

// MapFromBus reads hormone concentrations from a HormoneBus and applies them.
func (m *EndocrineExpressionMapper) MapFromBus(bus *deeptreeecho.HormoneBus, facs *FACSState) {
	allConc := bus.AllConcentrations()
	concentrations := make(map[deeptreeecho.Hormone]float64, int(deeptreeecho.HormoneCount))
	for h := deeptreeecho.Hormone(0); h < deeptreeecho.HormoneCount; h++ {
		concentrations[h] = allConc[h]
	}
	m.MapToFACS(concentrations, facs)
}

// CognitiveModePreset maps a cognitive mode to a FACS expression preset.
// This provides the cognitive-driven AU activations complementary to the
// endocrine-driven activations.
func CognitiveModePreset(mode deeptreeecho.CognitiveMode, facs *FACSState) {
	switch mode {
	case deeptreeecho.ModeExplore:
		// Curious, open — AU5+AU25 moderate
		facs.Add(AU5, 0.5)
		facs.Add(AU25, 0.4)
		facs.Add(AU1, 0.3)
	case deeptreeecho.ModeExploit:
		// Focused, routine — AU4+AU7 moderate
		facs.Add(AU4, 0.4)
		facs.Add(AU7, 0.3)
	case deeptreeecho.ModeFight:
		// Threat response — AU1+AU4+AU5+AU20 high
		facs.Add(AU1, 0.6)
		facs.Add(AU4, 0.7)
		facs.Add(AU5, 0.5)
		facs.Add(AU20, 0.6)
	case deeptreeecho.ModeFlight:
		// Avoidance — AU1+AU5+AU20
		facs.Add(AU1, 0.7)
		facs.Add(AU5, 0.6)
		facs.Add(AU20, 0.4)
	case deeptreeecho.ModeFreeze:
		// Paralysis — wide eyes, tense
		facs.Add(AU5, 0.8)
		facs.Add(AU7, 0.6)
		facs.Add(AU20, 0.3)
	case deeptreeecho.ModeFlow:
		// Optimal performance — relaxed focus
		facs.Add(AU6, 0.4)
		facs.Add(AU12, 0.3)
		facs.Add(AU7, 0.2)
	case deeptreeecho.ModeSocial:
		// Warm, engaged — AU6+AU12 high
		facs.Add(AU6, 0.7)
		facs.Add(AU12, 0.6)
		facs.Add(AU25, 0.3)
	case deeptreeecho.ModeRest:
		// Drowsy, peaceful — AU43+AU7 moderate
		facs.Add(AU43, 0.5)
		facs.Add(AU7, 0.3)
	case deeptreeecho.ModeCreative:
		// Divergent thinking — curious + joyful
		facs.Add(AU1, 0.4)
		facs.Add(AU2, 0.3)
		facs.Add(AU6, 0.5)
		facs.Add(AU12, 0.4)
		facs.Add(AU25, 0.3)
	case deeptreeecho.ModeAnalytical:
		// Convergent thinking — focused concentration
		facs.Add(AU4, 0.5)
		facs.Add(AU7, 0.4)
	}
}

// CognitiveLoadToFACS maps cognitive load (0-1) to expression.
// Higher load → more brow lowering and lid tightening.
func CognitiveLoadToFACS(load float64, facs *FACSState) {
	facs.Add(AU4, load*0.6) // Brow Lowerer
	facs.Add(AU7, load*0.4) // Lid Tightener
}

// ValenceArousalToFACS maps valence-arousal emotional dimensions to FACS.
// Valence: -1 (negative) to +1 (positive)
// Arousal: 0 (calm) to 1 (excited)
func ValenceArousalToFACS(valence, arousal float64, facs *FACSState) {
	if valence > 0 {
		facs.Add(AU6, valence*0.6)  // Cheek Raise
		facs.Add(AU12, valence*0.7) // Smile
	} else {
		facs.Add(AU15, -valence*0.5) // Lip Corner Depress
		facs.Add(AU4, -valence*0.3)  // Brow Lowerer
	}
	facs.Add(AU5, arousal*0.5)  // Upper Lid Raise
	facs.Add(AU25, arousal*0.3) // Lips Part
	facs.Add(AU26, arousal*0.2) // Jaw Drop
}
