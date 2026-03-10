package metahuman

import (
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/deeptreeecho"
)

// DNACognitiveBridge is the main orchestrator that implements the complete
// meta-echo-dna expression pipeline. It bridges the Deep Tree Echo cognitive
// architecture to MetaHuman avatar expression through:
//
//  1. Endocrine hormone → FACS Action Unit mapping
//  2. Cognitive state → FACS Action Unit mapping
//  3. Lorenz attractor chaotic micro-expression injection
//  4. SuperHotGirl aesthetic parameter biasing
//  5. Final FACS → MetaHuman CTRL_ morph target conversion
//
// This implements the echo-angel composition:
//
//	echo-angel' = echo-introspect ⊗ (meta-echo-dna ⊗ (aiangel-platform ⊕ (circled-operators ⊗ unreal-echo)))
type DNACognitiveBridge struct {
	mu sync.RWMutex

	// Components
	FACS       *FACSState
	Attractor  *LorenzAttractor
	Aesthetic  AestheticParameters
	Mapper     *EndocrineExpressionMapper

	// State
	lastUpdate time.Time
	frameCount uint64

	// Output cache
	lastMorphTargets  map[string]float64
	lastMaterialParams map[string]float64

	// Configuration
	EnableChaos     bool    // Enable Lorenz micro-expressions
	EnableAesthetic bool    // Enable SuperHotGirl aesthetic bias
	SmoothingFactor float64 // Parameter smoothing (0=none, 1=max)
}

// NewDNACognitiveBridge creates a fully initialized bridge with all components.
func NewDNACognitiveBridge() *DNACognitiveBridge {
	return &DNACognitiveBridge{
		FACS:            NewFACSState(),
		Attractor:       NewLorenzAttractor(),
		Aesthetic:       DefaultAestheticParameters(),
		Mapper:          NewEndocrineExpressionMapper(),
		lastUpdate:      time.Now(),
		EnableChaos:     true,
		EnableAesthetic: true,
		SmoothingFactor: 0.3,
	}
}

// ExpressionFrame represents one frame of the expression pipeline output.
type ExpressionFrame struct {
	// Timestamp of this frame
	Timestamp time.Time `json:"timestamp"`

	// Frame number
	Frame uint64 `json:"frame"`

	// FACS Action Unit activations (AU → intensity 0-1)
	ActionUnits map[ActionUnit]float64 `json:"action_units"`

	// MetaHuman CTRL_ morph targets (target name → value 0-1)
	MorphTargets map[string]float64 `json:"morph_targets"`

	// Dynamic material parameters
	MaterialParams map[string]float64 `json:"material_params"`

	// Diagnostic info
	LyapunovExponent float64       `json:"lyapunov_exponent"`
	CognitiveMode    string        `json:"cognitive_mode"`
	DeltaTime        time.Duration `json:"delta_time"`
}

// Update performs one complete expression pipeline cycle.
// This is the per-frame update function that should be called from the
// cognitive loop or animation tick.
//
// Pipeline steps:
//  1. Read endocrine state → compute hormone-driven AU activations
//  2. Read cognitive state → compute cognitive-driven AU activations
//  3. Blend hormone + cognitive AUs (sum, clamp [0,1])
//  4. Step Lorenz attractor, add micro-expression noise
//  5. Apply aesthetic parameter biases (SuperHotGirl)
//  6. Map final AU values to MetaHuman CTRL_ morph targets
//  7. Compute dynamic material instance parameters
func (b *DNACognitiveBridge) Update(
	endocrineState *deeptreeecho.VirtualEndocrineSystem,
	cognitiveMode deeptreeecho.CognitiveMode,
	cognitiveLoad float64,
	valence float64,
	arousal float64,
) ExpressionFrame {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	dt := now.Sub(b.lastUpdate)
	b.lastUpdate = now
	b.frameCount++

	// Step 1: Reset FACS for this frame
	b.FACS.Reset()

	// Step 2: Apply endocrine-driven AU activations
	if endocrineState != nil && endocrineState.Bus != nil {
		b.Mapper.MapFromBus(endocrineState.Bus, b.FACS)
	}

	// Step 3: Apply cognitive-driven AU activations
	CognitiveModePreset(cognitiveMode, b.FACS)
	CognitiveLoadToFACS(cognitiveLoad, b.FACS)
	ValenceArousalToFACS(valence, arousal, b.FACS)

	// Step 4: Apply chaotic micro-expression noise
	if b.EnableChaos {
		b.Attractor.ApplyMicroExpressionsDelta(b.FACS, dt.Seconds())
	}

	// Step 5: Apply aesthetic parameter biases
	if b.EnableAesthetic {
		b.Aesthetic.ApplyToFACS(b.FACS)
	}

	// Step 6: Convert to morph targets
	morphTargets := b.FACS.ToMorphTargets()

	// Step 7: Apply temporal smoothing
	if b.lastMorphTargets != nil && b.SmoothingFactor > 0 {
		for key, newVal := range morphTargets {
			if oldVal, ok := b.lastMorphTargets[key]; ok {
				morphTargets[key] = oldVal*b.SmoothingFactor + newVal*(1.0-b.SmoothingFactor)
			}
		}
	}
	b.lastMorphTargets = morphTargets

	// Step 8: Compute material parameters
	materialParams := b.Aesthetic.MaterialParameters()

	// Modulate material by endocrine state
	if endocrineState != nil && endocrineState.Bus != nil {
		allConc := endocrineState.Bus.AllConcentrations()
		// Dopamine boosts sparkle
		materialParams["EyeSparkleIntensity"] += allConc[deeptreeecho.DopaminePhasic] * 0.3
		// Serotonin boosts glow
		materialParams["SkinGlowIntensity"] += allConc[deeptreeecho.Serotonin] * 0.1
	}
	b.lastMaterialParams = materialParams

	// Build frame
	frame := ExpressionFrame{
		Timestamp:        now,
		Frame:            b.frameCount,
		ActionUnits:      b.FACS.Snapshot(),
		MorphTargets:     morphTargets,
		MaterialParams:   materialParams,
		LyapunovExponent: b.Attractor.LyapunovExponent(),
		CognitiveMode:    deeptreeecho.CognitiveModeNames[cognitiveMode],
		DeltaTime:        dt,
	}

	return frame
}

// UpdateSimple performs a simplified update using valence/arousal only
// (without a full endocrine system). Useful for Live2D integration.
func (b *DNACognitiveBridge) UpdateSimple(valence, arousal, cognitiveLoad float64) ExpressionFrame {
	return b.Update(nil, deeptreeecho.ModeExplore, cognitiveLoad, valence, arousal)
}

// GetLastMorphTargets returns the most recent morph target values.
func (b *DNACognitiveBridge) GetLastMorphTargets() map[string]float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.lastMorphTargets == nil {
		return make(map[string]float64)
	}
	result := make(map[string]float64, len(b.lastMorphTargets))
	for k, v := range b.lastMorphTargets {
		result[k] = v
	}
	return result
}

// GetLastMaterialParams returns the most recent material parameter values.
func (b *DNACognitiveBridge) GetLastMaterialParams() map[string]float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.lastMaterialParams == nil {
		return make(map[string]float64)
	}
	result := make(map[string]float64, len(b.lastMaterialParams))
	for k, v := range b.lastMaterialParams {
		result[k] = v
	}
	return result
}

// SetChaosIntensity adjusts the Lorenz attractor chaos intensity.
func (b *DNACognitiveBridge) SetChaosIntensity(intensity float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Attractor.ChaosIntensity = clamp(intensity, 0, 1)
}

// SetAesthetic updates the aesthetic parameters.
func (b *DNACognitiveBridge) SetAesthetic(params AestheticParameters) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Aesthetic = params
}

// IsHealthy returns true if all components are functioning correctly.
func (b *DNACognitiveBridge) IsHealthy() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Attractor.IsHealthy()
}

// Metrics returns diagnostic metrics for the bridge.
func (b *DNACognitiveBridge) Metrics() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return map[string]interface{}{
		"frame_count":       b.frameCount,
		"lyapunov_exponent": b.Attractor.LyapunovExponent(),
		"chaos_intensity":   b.Attractor.ChaosIntensity,
		"chaos_healthy":     b.Attractor.IsHealthy(),
		"enable_chaos":      b.EnableChaos,
		"enable_aesthetic":  b.EnableAesthetic,
		"smoothing_factor":  b.SmoothingFactor,
		"confidence":        b.Aesthetic.ConfidencePosture,
		"charisma":          b.Aesthetic.Charisma,
		"eye_sparkle":       b.Aesthetic.EyeSparkle,
	}
}
