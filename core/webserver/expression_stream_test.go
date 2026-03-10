package webserver

import (
	"testing"
	"time"

	"github.com/o9nn/echo.go/core/metahuman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpressionStream(t *testing.T) {
	hub := NewWebSocketHub()
	es := NewExpressionStream(hub)
	require.NotNil(t, es)
	assert.Equal(t, time.Millisecond*33, es.minFrameInterval)
}

func TestExpressionStreamSetMaxBroadcastRate(t *testing.T) {
	es := NewExpressionStream(nil)

	es.SetMaxBroadcastRate(60.0)
	assert.InDelta(t, time.Millisecond*16, es.minFrameInterval, float64(time.Millisecond))

	es.SetMaxBroadcastRate(0) // invalid
	assert.InDelta(t, time.Millisecond*16, es.minFrameInterval, float64(time.Millisecond))
}

func TestExpressionStreamOnFrame(t *testing.T) {
	es := NewExpressionStream(nil) // no hub, just test frame processing

	frame := metahuman.ExpressionFrame{
		Timestamp: time.Now(),
		ActionUnits: map[metahuman.ActionUnit]float64{
			metahuman.AU6:  0.7,
			metahuman.AU12: 0.8,
		},
		MorphTargets: map[string]float64{
			"CTRL_expressions_cheekRaiseL": 0.7,
			"CTRL_expressions_mouthSmileL": 0.8,
		},
		MaterialParams: map[string]float64{
			"SkinFlush":     0.3,
			"EyeWetness":    0.1,
			"PupilDilation": 0.5,
		},
		CognitiveMode:    "exploration",
		LyapunovExponent: 0.66,
	}

	es.OnExpressionFrame(frame)
	assert.Equal(t, uint64(1), es.frameCount)
}

func TestExpressionStreamThrottling(t *testing.T) {
	es := NewExpressionStream(nil)
	es.SetMaxBroadcastRate(10) // 10fps = 100ms interval

	frame := metahuman.ExpressionFrame{
		Timestamp: time.Now(),
	}

	// Send 5 frames rapidly
	for i := 0; i < 5; i++ {
		es.OnExpressionFrame(frame)
	}

	// First frame should broadcast, rest should be dropped
	assert.Equal(t, uint64(5), es.frameCount)
	assert.Equal(t, uint64(4), es.droppedCount) // 4 dropped due to throttle
}

func TestExpressionStreamWireFrame(t *testing.T) {
	es := NewExpressionStream(nil)

	frame := metahuman.ExpressionFrame{
		Timestamp: time.Now(),
		ActionUnits: map[metahuman.ActionUnit]float64{
			metahuman.AU6:  0.7,
			metahuman.AU12: 0.8,
			metahuman.AU1:  0.005, // below threshold, should be excluded
		},
		MorphTargets: map[string]float64{
			"CTRL_expressions_cheekRaiseL": 0.7,
			"CTRL_expressions_mouthSmileL": 0.8,
			"CTRL_tiny_value":              0.005, // below threshold
		},
		MaterialParams: map[string]float64{
			"SkinFlush":     0.3,
			"PupilDilation": 0.5,
		},
		CognitiveMode:    "creative",
		LyapunovExponent: 0.66,
	}

	wire := es.buildWireFrame(frame)

	assert.Equal(t, "creative", wire.Mode)
	assert.InDelta(t, 0.66, wire.Lyapunov, 0.01)

	// Check that below-threshold values are excluded
	if wire.AUs != nil {
		_, hasAU1 := wire.AUs["AU1"]
		assert.False(t, hasAU1, "AU1 below threshold should be excluded")
		_, hasAU6 := wire.AUs["AU6"]
		assert.True(t, hasAU6, "AU6 above threshold should be included")
	}

	// Check material params
	if wire.Materials != nil {
		_, hasSkinFlush := wire.Materials["SkinFlush"]
		assert.True(t, hasSkinFlush, "SkinFlush should be included")
	}
}

func TestExpressionStreamMetrics(t *testing.T) {
	es := NewExpressionStream(nil)

	frame := metahuman.ExpressionFrame{Timestamp: time.Now()}
	es.OnExpressionFrame(frame)

	metrics := es.GetMetrics()
	assert.Equal(t, uint64(1), metrics.TotalFrames)
	assert.Contains(t, metrics.String(), "ExpressionStream")
}

func TestRoundTo3(t *testing.T) {
	assert.Equal(t, 0.123, roundTo3(0.1234))
	assert.Equal(t, 0.0, roundTo3(0.0))
	assert.Equal(t, 1.0, roundTo3(1.0))
	assert.Equal(t, 0.667, roundTo3(0.6666))
}

func BenchmarkExpressionStreamOnFrame(b *testing.B) {
	es := NewExpressionStream(nil)
	es.SetMaxBroadcastRate(120) // high rate to avoid throttling

	frame := metahuman.ExpressionFrame{
		Timestamp: time.Now(),
		ActionUnits: map[metahuman.ActionUnit]float64{
			metahuman.AU6:  0.7,
			metahuman.AU12: 0.8,
			metahuman.AU1:  0.3,
			metahuman.AU2:  0.4,
		},
		MorphTargets: map[string]float64{
			"CTRL_expressions_cheekRaiseL": 0.7,
			"CTRL_expressions_mouthSmileL": 0.8,
		},
		CognitiveMode:    "exploration",
		LyapunovExponent: 0.66,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		es.OnExpressionFrame(frame)
	}
}
