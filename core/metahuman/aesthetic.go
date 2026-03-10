package metahuman

// AestheticParameters holds the SuperHotGirl aesthetic modulation parameters.
// These bias expressions toward confident, charismatic presentations and
// drive dynamic material updates for the MetaHuman avatar.
type AestheticParameters struct {
	// ConfidencePosture biases toward upright posture, raised chin, squared shoulders.
	// Range: 0.0 - 1.0
	ConfidencePosture float64

	// Charisma amplifies smile warmth and eye contact intensity.
	// Range: 0.0 - 1.0
	Charisma float64

	// EyeSparkle controls specular highlight intensity in iris material.
	// Range: 0.0 - 1.0
	EyeSparkle float64

	// GracefulMovement controls motion smoothing and acceleration curves.
	// Range: 0.0 - 1.0
	GracefulMovement float64

	// EmissiveGlow controls skin subsurface scattering boost.
	// Range: 0.0 - 0.5
	EmissiveGlow float64
}

// DefaultAestheticParameters returns the default SuperHotGirl aesthetic preset.
func DefaultAestheticParameters() AestheticParameters {
	return AestheticParameters{
		ConfidencePosture: 0.7,
		Charisma:          0.6,
		EyeSparkle:        0.5,
		GracefulMovement:  0.6,
		EmissiveGlow:      0.15,
	}
}

// ApplyToFACS biases the FACS state toward confident, charismatic expressions.
// This is applied after endocrine and cognitive mappings but before final output.
func (ap *AestheticParameters) ApplyToFACS(facs *FACSState) {
	// Confidence → slight chin raise, reduced brow lowering
	if ap.ConfidencePosture > 0.5 {
		boost := (ap.ConfidencePosture - 0.5) * 2.0 // 0-1 range
		facs.Add(AU17, boost*0.3)                    // Chin Raise
		facs.Add(AU2, boost*0.2)                     // Outer Brow Raise (confident)
		// Reduce stress indicators
		current := facs.Get(AU4)
		facs.Set(AU4, current*(1.0-boost*0.3)) // Reduce brow lowering
		currentDepress := facs.Get(AU15)
		facs.Set(AU15, currentDepress*(1.0-boost*0.4)) // Reduce lip corner depress
	}

	// Charisma → warmer smile, more eye engagement
	if ap.Charisma > 0.3 {
		boost := (ap.Charisma - 0.3) / 0.7 // Normalize to 0-1
		facs.Add(AU12, boost*0.2)           // Smile boost
		facs.Add(AU6, boost*0.15)           // Cheek raise boost
	}

	// EyeSparkle → wider eyes, more alert appearance
	if ap.EyeSparkle > 0.3 {
		boost := (ap.EyeSparkle - 0.3) / 0.7
		facs.Add(AU5, boost*0.1) // Slight upper lid raise
	}
}

// MaterialParameters returns the dynamic material instance values
// for MetaHuman rendering based on the aesthetic parameters.
func (ap *AestheticParameters) MaterialParameters() map[string]float64 {
	return map[string]float64{
		"EyeSparkleIntensity": ap.EyeSparkle,
		"SkinGlowIntensity":   ap.EmissiveGlow,
		"IrisSpecular":        ap.EyeSparkle * 2.0,
		"SkinSSSBoost":        ap.EmissiveGlow * 1.5,
		"MotionSmoothing":     ap.GracefulMovement,
	}
}

// CompositeExpressions defines named expression presets that combine
// multiple AUs with aesthetic modifiers.
type CompositeExpression struct {
	Name     string
	AUs      map[ActionUnit]float64
	Modifier string // Which aesthetic parameter to boost
}

// CompositeExpressionPresets defines the canonical composite expressions
// from the meta-echo-dna FACS mapping specification.
var CompositeExpressionPresets = []CompositeExpression{
	{
		Name:     "GenuineSmile",
		AUs:      map[ActionUnit]float64{AU6: 0.8, AU12: 0.9},
		Modifier: "ConfidencePosture",
	},
	{
		Name:     "Flirtatious",
		AUs:      map[ActionUnit]float64{AU12: 0.6, AU6: 0.5, AU46: 0.7},
		Modifier: "Charisma",
	},
	{
		Name:     "Curious",
		AUs:      map[ActionUnit]float64{AU1: 0.6, AU2: 0.5, AU5: 0.7},
		Modifier: "EyeSparkle",
	},
	{
		Name:     "Confident",
		AUs:      map[ActionUnit]float64{AU2: 0.5, AU12: 0.6, AU17: 0.7},
		Modifier: "ConfidencePosture",
	},
	{
		Name:     "Playful",
		AUs:      map[ActionUnit]float64{AU12: 0.7, AU25: 0.5, AU6: 0.6},
		Modifier: "Charisma",
	},
}

// ApplyCompositeExpression applies a named composite expression to the FACS state.
func ApplyCompositeExpression(name string, facs *FACSState, aesthetic *AestheticParameters) bool {
	for _, preset := range CompositeExpressionPresets {
		if preset.Name == name {
			for au, value := range preset.AUs {
				facs.Add(au, value)
			}
			// Boost the associated aesthetic parameter
			switch preset.Modifier {
			case "ConfidencePosture":
				aesthetic.ConfidencePosture = clamp(aesthetic.ConfidencePosture+0.1, 0, 1)
			case "Charisma":
				aesthetic.Charisma = clamp(aesthetic.Charisma+0.1, 0, 1)
			case "EyeSparkle":
				aesthetic.EyeSparkle = clamp(aesthetic.EyeSparkle+0.1, 0, 1)
			}
			return true
		}
	}
	return false
}
