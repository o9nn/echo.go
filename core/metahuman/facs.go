// Package metahuman implements the MetaHuman DNA Cognitive Bridge.
//
// This package provides the complete expression pipeline from the meta-echo-dna
// skill: mapping Deep Tree Echo's endocrine and cognitive states through FACS
// Action Units, chaotic micro-expression dynamics (Lorenz attractor), and
// SuperHotGirl aesthetic parameters to MetaHuman CTRL_ morph targets.
//
// Architecture:
//
//	Endocrine State ──┐
//	                  ├→ FACS Action Units → MetaHuman CTRL_ Morph Targets
//	Cognitive State ──┘        ↕                    ↕
//	                   Chaotic Dynamics      Aesthetic Parameters
//	                   (Lorenz Attractor)    (SuperHotGirl)
package metahuman

import (
	"fmt"
	"math"
	"sync"
)

// ActionUnit represents a FACS Action Unit with its current activation level.
type ActionUnit int

const (
	AU1  ActionUnit = iota // Inner Brow Raise
	AU2                    // Outer Brow Raise
	AU4                    // Brow Lowerer
	AU5                    // Upper Lid Raise
	AU6                    // Cheek Raise
	AU7                    // Lid Tightener
	AU9                    // Nose Wrinkle
	AU10                   // Upper Lip Raise
	AU12                   // Lip Corner Pull (Smile)
	AU14                   // Dimpler
	AU15                   // Lip Corner Depress
	AU17                   // Chin Raise
	AU20                   // Lip Stretch
	AU23                   // Lip Tightener
	AU25                   // Lips Part
	AU26                   // Jaw Drop
	AU28                   // Lip Suck
	AU43                   // Eyes Closed
	AU45                   // Blink
	AU46                   // Wink
	AUCount                // Sentinel for iteration
)

// ActionUnitNames maps action units to their FACS descriptions.
var ActionUnitNames = map[ActionUnit]string{
	AU1:  "Inner Brow Raise",
	AU2:  "Outer Brow Raise",
	AU4:  "Brow Lowerer",
	AU5:  "Upper Lid Raise",
	AU6:  "Cheek Raise",
	AU7:  "Lid Tightener",
	AU9:  "Nose Wrinkle",
	AU10: "Upper Lip Raise",
	AU12: "Lip Corner Pull",
	AU14: "Dimpler",
	AU15: "Lip Corner Depress",
	AU17: "Chin Raise",
	AU20: "Lip Stretch",
	AU23: "Lip Tightener",
	AU25: "Lips Part",
	AU26: "Jaw Drop",
	AU28: "Lip Suck",
	AU43: "Eyes Closed",
	AU45: "Blink",
	AU46: "Wink",
}

// MetaHumanMorphTarget maps FACS AUs to MetaHuman CTRL_ morph target names.
var MetaHumanMorphTarget = map[ActionUnit]string{
	AU1:  "CTRL_brow_inner_UP",
	AU2:  "CTRL_brow_outer_UP",
	AU4:  "CTRL_brow_down",
	AU5:  "CTRL_eye_upperLid_UP",
	AU6:  "CTRL_cheek_raise",
	AU7:  "CTRL_eye_squint",
	AU9:  "CTRL_nose_wrinkle",
	AU10: "CTRL_mouth_upperLip_UP",
	AU12: "CTRL_mouth_cornerPull",
	AU14: "CTRL_mouth_dimple",
	AU15: "CTRL_mouth_cornerDepress",
	AU17: "CTRL_chin_raise",
	AU20: "CTRL_mouth_stretch",
	AU23: "CTRL_mouth_tighten",
	AU25: "CTRL_mouth_lipsPart",
	AU26: "CTRL_jaw_open",
	AU28: "CTRL_mouth_lipSuck",
	AU43: "CTRL_eye_blink",
	AU45: "CTRL_eye_blink",
	AU46: "CTRL_eye_blink_L",
}

// FACSState holds the current activation levels for all tracked action units.
type FACSState struct {
	mu          sync.RWMutex
	activations [AUCount]float64
}

// NewFACSState creates a new FACS state with all AUs at zero.
func NewFACSState() *FACSState {
	return &FACSState{}
}

// Set sets the activation level for a specific action unit, clamped to [0, 1].
func (f *FACSState) Set(au ActionUnit, value float64) {
	if au < 0 || au >= AUCount {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activations[au] = clamp(value, 0.0, 1.0)
}

// Get returns the current activation level for a specific action unit.
func (f *FACSState) Get(au ActionUnit) float64 {
	if au < 0 || au >= AUCount {
		return 0.0
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.activations[au]
}

// Add adds a value to the current activation level, clamped to [0, 1].
func (f *FACSState) Add(au ActionUnit, value float64) {
	if au < 0 || au >= AUCount {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activations[au] = clamp(f.activations[au]+value, 0.0, 1.0)
}

// Reset sets all action units to zero.
func (f *FACSState) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.activations {
		f.activations[i] = 0.0
	}
}

// Snapshot returns a copy of all current AU activations.
func (f *FACSState) Snapshot() map[ActionUnit]float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make(map[ActionUnit]float64, int(AUCount))
	for i := ActionUnit(0); i < AUCount; i++ {
		if f.activations[i] > 0.001 {
			result[i] = f.activations[i]
		}
	}
	return result
}

// ToMorphTargets converts current FACS state to MetaHuman CTRL_ morph targets.
func (f *FACSState) ToMorphTargets() map[string]float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	targets := make(map[string]float64)
	for au := ActionUnit(0); au < AUCount; au++ {
		if f.activations[au] > 0.001 {
			if target, ok := MetaHumanMorphTarget[au]; ok {
				// Sum contributions if multiple AUs map to same target
				targets[target] += f.activations[au]
				if targets[target] > 1.0 {
					targets[target] = 1.0
				}
			}
		}
	}
	return targets
}

// clamp restricts a value to the range [min, max].
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ActionUnitCodes maps ActionUnit constants to their FACS code strings (e.g., AU6 → "AU6").
var ActionUnitCodes = map[ActionUnit]string{
	AU1:  "AU1",
	AU2:  "AU2",
	AU4:  "AU4",
	AU5:  "AU5",
	AU6:  "AU6",
	AU7:  "AU7",
	AU9:  "AU9",
	AU10: "AU10",
	AU12: "AU12",
	AU14: "AU14",
	AU15: "AU15",
	AU17: "AU17",
	AU20: "AU20",
	AU23: "AU23",
	AU25: "AU25",
	AU26: "AU26",
	AU28: "AU28",
	AU43: "AU43",
	AU45: "AU45",
	AU46: "AU46",
}

// ActionUnitCode returns the FACS code string for an ActionUnit (e.g., "AU6").
func ActionUnitCode(au ActionUnit) string {
	if code, ok := ActionUnitCodes[au]; ok {
		return code
	}
	return fmt.Sprintf("AU%d", au)
}

// Ensure math is used
var _ = math.Abs
