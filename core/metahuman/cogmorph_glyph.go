package metahuman

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GlyphProjection provides a visual debug representation of the current
// expression state using the CogMorph glyph system from codex-grassmania.
// Each AU activation is projected to a Unicode glyph with intensity encoding.
type GlyphProjection struct {
	// AU → glyph mapping
	glyphs map[ActionUnit]rune
}

// NewGlyphProjection creates a new glyph projection system.
func NewGlyphProjection() *GlyphProjection {
	return &GlyphProjection{
		glyphs: map[ActionUnit]rune{
			AU1:  '⌢', // Inner Brow Raise
			AU2:  '⌣', // Outer Brow Raise
			AU4:  '⌐', // Brow Lowerer
			AU5:  '◉', // Upper Lid Raise
			AU6:  '☺', // Cheek Raise
			AU7:  '◎', // Lid Tightener
			AU9:  '≋', // Nose Wrinkle
			AU10: '⌒', // Upper Lip Raise
			AU12: '☻', // Smile
			AU14: '◦', // Dimpler
			AU15: '☹', // Lip Corner Depress
			AU17: '▲', // Chin Raise
			AU20: '━', // Lip Stretch
			AU23: '▬', // Lip Tightener
			AU25: '○', // Lips Part
			AU26: '▽', // Jaw Drop
			AU28: '●', // Lip Suck
			AU43: '◡', // Eyes Closed
			AU45: '◠', // Blink
			AU46: '◐', // Wink
		},
	}
}

// intensityBlock returns a block character representing intensity level.
func intensityBlock(intensity float64) string {
	if intensity < 0.125 {
		return "░"
	}
	if intensity < 0.375 {
		return "▒"
	}
	if intensity < 0.625 {
		return "▓"
	}
	return "█"
}

// Project generates a text-based glyph projection of the FACS state.
func (gp *GlyphProjection) Project(facs *FACSState) string {
	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════╗\n")
	sb.WriteString("║   CogMorph Expression Glyph      ║\n")
	sb.WriteString("╠══════════════════════════════════╣\n")

	snap := facs.Snapshot()
	if len(snap) == 0 {
		sb.WriteString("║   (neutral — no active AUs)      ║\n")
	} else {
		for au := ActionUnit(0); au < AUCount; au++ {
			if intensity, ok := snap[au]; ok {
				glyph := gp.glyphs[au]
				name := ActionUnitNames[au]
				bar := intensityBlock(intensity)
				sb.WriteString(fmt.Sprintf("║ %c AU%-2d %-20s %s %.2f ║\n",
					glyph, au, name, bar, intensity))
			}
		}
	}

	sb.WriteString("╚══════════════════════════════════╝")
	return sb.String()
}

// ProjectJSON generates a JSON representation of the expression state
// suitable for transmission to external visualization tools.
func (gp *GlyphProjection) ProjectJSON(frame ExpressionFrame) (string, error) {
	projection := map[string]interface{}{
		"frame":     frame.Frame,
		"timestamp": frame.Timestamp,
		"action_units": func() map[string]interface{} {
			aus := make(map[string]interface{})
			for au, intensity := range frame.ActionUnits {
				aus[fmt.Sprintf("AU%d", au)] = map[string]interface{}{
					"name":      ActionUnitNames[au],
					"intensity": intensity,
					"glyph":     string(gp.glyphs[au]),
				}
			}
			return aus
		}(),
		"morph_targets":    frame.MorphTargets,
		"material_params":  frame.MaterialParams,
		"lyapunov":         frame.LyapunovExponent,
		"cognitive_mode":   frame.CognitiveMode,
	}

	data, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
