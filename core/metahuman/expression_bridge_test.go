package metahuman

import (
	"sync"
	"testing"

	"github.com/o9nn/echo.go/core/deeptreeecho"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExpressionBridge(t *testing.T) {
	eb := NewExpressionBridge()
	require.NotNil(t, eb)
	assert.NotNil(t, eb.dnaBridge)
	assert.Equal(t, 30.0, eb.frameRate)
	assert.Equal(t, uint64(0), eb.totalFrames)
}

func TestExpressionBridgeSetFrameRate(t *testing.T) {
	eb := NewExpressionBridge()

	eb.SetFrameRate(60.0)
	assert.Equal(t, 60.0, eb.frameRate)

	eb.SetFrameRate(0) // invalid, should not change
	assert.Equal(t, 60.0, eb.frameRate)

	eb.SetFrameRate(121) // too high, should not change
	assert.Equal(t, 60.0, eb.frameRate)

	eb.SetFrameRate(120) // max valid
	assert.Equal(t, 120.0, eb.frameRate)
}

func TestExpressionBridgeUpdateFromEndocrine(t *testing.T) {
	eb := NewExpressionBridge()

	// Create a VES and signal some events
	ves := deeptreeecho.NewVirtualEndocrineSystem()
	ves.SignalEvent(deeptreeecho.EndoRewardReceived, 0.8)
	ves.Tick(0.1)

	// Update the bridge
	frame := eb.UpdateFromEndocrine(ves, deeptreeecho.ModeExplore, 0.3, 0.7, 0.5, 0.033)

	// Should have produced a frame
	assert.NotZero(t, frame.Timestamp)
	assert.Equal(t, uint64(1), eb.totalFrames)

	// Morph targets should have some non-zero values
	hasNonZero := false
	for _, v := range frame.MorphTargets {
		if v != 0 {
			hasNonZero = true
			break
		}
	}
	assert.True(t, hasNonZero, "Expected some non-zero morph targets after endocrine update")
}

func TestExpressionBridgeListeners(t *testing.T) {
	eb := NewExpressionBridge()

	var received []ExpressionFrame
	var mu sync.Mutex

	eb.AddListenerFunc(func(frame ExpressionFrame) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, frame)
	})

	ves := deeptreeecho.NewVirtualEndocrineSystem()
	ves.SignalEvent(deeptreeecho.EndoNoveltyEncountered, 0.9)
	ves.Tick(0.1)

	// Update 3 times
	for i := 0; i < 3; i++ {
		eb.UpdateFromEndocrine(ves, deeptreeecho.ModeCreative, 0.5, 0.6, 0.7, 0.033)
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 3, "Listener should receive 3 frames")
}

func TestExpressionBridgeMetrics(t *testing.T) {
	eb := NewExpressionBridge()

	ves := deeptreeecho.NewVirtualEndocrineSystem()
	ves.Tick(0.1)

	for i := 0; i < 10; i++ {
		eb.UpdateFromEndocrine(ves, deeptreeecho.ModeAnalytical, 0.3, 0.5, 0.4, 0.033)
	}

	metrics := eb.GetMetrics()
	assert.Equal(t, uint64(10), metrics.TotalFrames)
	assert.Greater(t, metrics.AvgFrameTimeUs, 0.0)
	assert.Greater(t, metrics.PeakFrameTimeUs, 0.0)
	assert.Equal(t, 30.0, metrics.FrameRate)
	assert.Contains(t, metrics.String(), "ExpressionBridge")
}

func TestExpressionBridgeCognitiveStateUpdate(t *testing.T) {
	eb := NewExpressionBridge()

	eb.UpdateCognitiveState("exploration", "perceive", 0.7, 0.5, 0.6, 0.1, 0.4)

	eb.mu.RLock()
	defer eb.mu.RUnlock()
	assert.Equal(t, "exploration", eb.cognitiveMode)
	assert.Equal(t, "perceive", eb.cognitivePhase)
	assert.Equal(t, 0.7, eb.arousal)
	assert.Equal(t, 0.5, eb.valence)
	assert.Equal(t, 0.6, eb.dominance)
}

func TestExpressionBridgeSimpleUpdate(t *testing.T) {
	eb := NewExpressionBridge()

	frame := eb.UpdateSimple(0.7, 0.5, 0.3)
	assert.NotZero(t, frame.Timestamp)
	assert.Equal(t, uint64(1), eb.totalFrames)
}

func TestExpressionBridgeMultipleModes(t *testing.T) {
	eb := NewExpressionBridge()

	ves := deeptreeecho.NewVirtualEndocrineSystem()

	modes := []deeptreeecho.CognitiveMode{
		deeptreeecho.ModeExplore,
		deeptreeecho.ModeCreative,
		deeptreeecho.ModeSocial,
		deeptreeecho.ModeAnalytical,
		deeptreeecho.ModeFlow,
		deeptreeecho.ModeRest,
	}

	for _, mode := range modes {
		ves.SignalEvent(deeptreeecho.EndoRewardReceived, 0.5)
		ves.Tick(0.1)
		eb.UpdateFromEndocrine(ves, mode, 0.3, 0.5, 0.6, 0.033)
	}

	assert.Equal(t, uint64(len(modes)), eb.totalFrames)

	// Each mode should produce different expression patterns
	frame := eb.GetCurrentFrame()
	assert.NotZero(t, frame.Timestamp)
}

func TestExpressionBridgeGetDNABridge(t *testing.T) {
	eb := NewExpressionBridge()
	dnaBridge := eb.GetDNABridge()
	require.NotNil(t, dnaBridge)
}

func BenchmarkExpressionBridgeUpdate(b *testing.B) {
	eb := NewExpressionBridge()
	ves := deeptreeecho.NewVirtualEndocrineSystem()
	ves.SignalEvent(deeptreeecho.EndoRewardReceived, 0.7)
	ves.Tick(0.1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.UpdateFromEndocrine(ves, deeptreeecho.ModeExplore, 0.3, 0.7, 0.5, 0.033)
	}
}

func BenchmarkExpressionBridgeWithListeners(b *testing.B) {
	eb := NewExpressionBridge()
	ves := deeptreeecho.NewVirtualEndocrineSystem()
	ves.SignalEvent(deeptreeecho.EndoRewardReceived, 0.7)
	ves.Tick(0.1)

	// Add 5 listeners
	for i := 0; i < 5; i++ {
		eb.AddListenerFunc(func(frame ExpressionFrame) {
			_ = frame.Timestamp
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eb.UpdateFromEndocrine(ves, deeptreeecho.ModeCreative, 0.5, 0.6, 0.7, 0.033)
	}
}
