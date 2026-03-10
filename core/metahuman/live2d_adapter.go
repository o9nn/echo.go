package metahuman

import (
	"github.com/o9nn/echo.go/core/live2d"
)

// Live2DAdapter bridges the MetaHuman DNA expression pipeline to the
// existing Live2D avatar system. It converts FACS-based morph targets
// to Live2D Cubism parameter updates.
type Live2DAdapter struct {
	bridge *DNACognitiveBridge

	// Mapping from MetaHuman CTRL_ targets to Live2D parameters
	targetToParam map[string]string
}

// NewLive2DAdapter creates an adapter that connects the MetaHuman DNA
// bridge to the Live2D avatar manager.
func NewLive2DAdapter(bridge *DNACognitiveBridge) *Live2DAdapter {
	adapter := &Live2DAdapter{
		bridge: bridge,
		targetToParam: map[string]string{
			// MetaHuman CTRL_ → Live2D Cubism Parameter mapping
			"CTRL_mouth_cornerPull":   "ParamMouthSmile",
			"CTRL_cheek_raise":        "ParamEyeLSmile",
			"CTRL_eye_blink":          "ParamEyeLOpen",
			"CTRL_eye_upperLid_UP":    "ParamEyeLOpen",
			"CTRL_eye_squint":         "ParamEyeLSmile",
			"CTRL_brow_inner_UP":      "ParamBrowLY",
			"CTRL_brow_outer_UP":      "ParamBrowLY",
			"CTRL_brow_down":          "ParamBrowLY",
			"CTRL_jaw_open":           "ParamMouthOpenY",
			"CTRL_mouth_lipsPart":     "ParamMouthOpenY",
			"CTRL_mouth_cornerDepress": "ParamMouthForm",
			"CTRL_nose_wrinkle":       "ParamMouthForm",
		},
	}
	return adapter
}

// ToLive2DParameters converts the current bridge state to Live2D parameters.
func (a *Live2DAdapter) ToLive2DParameters() []live2d.ModelParameter {
	morphTargets := a.bridge.GetLastMorphTargets()
	params := make([]live2d.ModelParameter, 0, len(morphTargets))

	for target, value := range morphTargets {
		if paramID, ok := a.targetToParam[target]; ok {
			// Handle inverted mappings
			paramValue := value
			switch target {
			case "CTRL_eye_blink":
				// Blink inverts eye openness
				paramValue = 1.0 - value
			case "CTRL_brow_down":
				// Brow down is negative direction
				paramValue = -value * 30.0 // Scale to Live2D range
			case "CTRL_mouth_cornerDepress":
				// Depress is negative mouth form
				paramValue = -value
			default:
				// Scale to Live2D parameter ranges
				if paramID == "ParamBrowLY" {
					paramValue = value * 30.0
				} else if paramID == "ParamAngleX" || paramID == "ParamAngleY" || paramID == "ParamAngleZ" {
					paramValue = value * 30.0
				}
			}

			params = append(params, live2d.ModelParameter{
				ID:    paramID,
				Value: paramValue,
				Min:   -30.0,
				Max:   30.0,
			})
		}
	}

	return params
}

// ToEmotionalState converts the current bridge FACS state to the
// existing Live2D EmotionalState format for backward compatibility.
func (a *Live2DAdapter) ToEmotionalState() live2d.EmotionalState {
	facs := a.bridge.FACS

	// Derive valence from smile vs frown AUs
	smile := facs.Get(AU12) + facs.Get(AU6)*0.5
	frown := facs.Get(AU15) + facs.Get(AU4)*0.5
	valence := clamp(smile-frown, -1.0, 1.0)

	// Derive arousal from eye/mouth openness
	arousal := clamp(facs.Get(AU5)*0.5+facs.Get(AU25)*0.3+facs.Get(AU26)*0.2, 0, 1)

	// Derive dominance from confidence indicators
	dominance := clamp(0.5+facs.Get(AU17)*0.3-facs.Get(AU1)*0.2, 0, 1)

	// Curiosity from brow raise
	curiosity := clamp(facs.Get(AU1)*0.5+facs.Get(AU2)*0.5, 0, 1)

	// Confidence from aesthetic parameters
	confidence := a.bridge.Aesthetic.ConfidencePosture

	return live2d.EmotionalState{
		Valence:    valence,
		Arousal:    arousal,
		Dominance:  dominance,
		Curiosity:  curiosity,
		Confidence: confidence,
	}
}
