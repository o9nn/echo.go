package metahuman

import (
	"fmt"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/deeptreeecho"
)

// =============================================================================
// EXPRESSION BRIDGE
// =============================================================================
//
// Connects the CognitiveEventLoop's endocrine system to the MetaHuman DNA
// expression pipeline, producing real-time facial animation data that can be
// streamed to Live2D, MetaHuman, or any avatar renderer.
//
// Pipeline: CognitiveEventLoop → ExpressionBridge → WebSocket/Live2D
//
// The bridge runs as a goroutine synchronized to the cognitive event loop,
// sampling endocrine state at each cognitive step and producing expression
// frames at a configurable frame rate (default 30fps).
//
// =============================================================================

// ExpressionFrameListener receives expression frames for rendering
type ExpressionFrameListener interface {
	OnExpressionFrame(frame ExpressionFrame)
}

// ExpressionFrameFunc is a function adapter for ExpressionFrameListener
type ExpressionFrameFunc func(frame ExpressionFrame)

func (f ExpressionFrameFunc) OnExpressionFrame(frame ExpressionFrame) {
	f(frame)
}

// ExpressionBridge connects the cognitive event loop to the MetaHuman DNA
// expression pipeline for real-time avatar animation.
type ExpressionBridge struct {
	mu sync.RWMutex

	// MetaHuman DNA bridge
	dnaBridge *DNACognitiveBridge

	// Frame output
	currentFrame ExpressionFrame
	frameCount   uint64
	frameRate    float64 // target fps

	// Cognitive state cache (updated by event loop)
	cognitiveMode   string
	cognitivePhase  string
	arousal         float64
	valence         float64
	dominance       float64
	predictionError float64
	cognitiveLoad   float64

	// Frame listeners
	listeners []ExpressionFrameListener

	// Metrics
	totalFrames     uint64
	avgFrameTimeNs  int64
	peakFrameTimeNs int64
	droppedFrames   uint64
}

// NewExpressionBridge creates a new expression bridge
func NewExpressionBridge() *ExpressionBridge {
	return &ExpressionBridge{
		dnaBridge: NewDNACognitiveBridge(),
		frameRate: 30.0,
	}
}

// SetFrameRate sets the target frame rate for expression updates
func (eb *ExpressionBridge) SetFrameRate(fps float64) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if fps > 0 && fps <= 120 {
		eb.frameRate = fps
	}
}

// AddListener adds a frame listener
func (eb *ExpressionBridge) AddListener(l ExpressionFrameListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners = append(eb.listeners, l)
}

// AddListenerFunc adds a function as a frame listener
func (eb *ExpressionBridge) AddListenerFunc(f func(frame ExpressionFrame)) {
	eb.AddListener(ExpressionFrameFunc(f))
}

// UpdateFromEndocrine samples the endocrine system and produces a new frame.
// Called by the CognitiveEventLoop at each cognitive step.
func (eb *ExpressionBridge) UpdateFromEndocrine(
	endocrineState *deeptreeecho.VirtualEndocrineSystem,
	mode deeptreeecho.CognitiveMode,
	cognitiveLoad float64,
	valence float64,
	arousal float64,
	dt float64,
) ExpressionFrame {
	eb.mu.Lock()

	start := time.Now()

	// Delegate to the DNA bridge
	frame := eb.dnaBridge.Update(endocrineState, mode, cognitiveLoad, valence, arousal)
	eb.currentFrame = frame
	eb.frameCount++
	eb.totalFrames++

	// Track timing
	elapsed := time.Since(start).Nanoseconds()
	if eb.avgFrameTimeNs == 0 {
		eb.avgFrameTimeNs = elapsed
	} else {
		eb.avgFrameTimeNs = (eb.avgFrameTimeNs*7 + elapsed) / 8 // EMA
	}
	if elapsed > eb.peakFrameTimeNs {
		eb.peakFrameTimeNs = elapsed
	}

	// Copy listeners for notification outside lock
	listeners := make([]ExpressionFrameListener, len(eb.listeners))
	copy(listeners, eb.listeners)
	eb.mu.Unlock()

	// Notify listeners outside the lock
	for _, l := range listeners {
		l.OnExpressionFrame(frame)
	}

	return frame
}

// UpdateSimple performs a simplified update using valence/arousal only
func (eb *ExpressionBridge) UpdateSimple(valence, arousal, cognitiveLoad float64) ExpressionFrame {
	return eb.UpdateFromEndocrine(nil, deeptreeecho.ModeExplore, cognitiveLoad, valence, arousal, 0.033)
}

// UpdateCognitiveState updates the cached cognitive state from the event loop
func (eb *ExpressionBridge) UpdateCognitiveState(
	mode string,
	phase string,
	arousal, valence, dominance float64,
	predictionError, cognitiveLoad float64,
) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.cognitiveMode = mode
	eb.cognitivePhase = phase
	eb.arousal = arousal
	eb.valence = valence
	eb.dominance = dominance
	eb.predictionError = predictionError
	eb.cognitiveLoad = cognitiveLoad
}

// GetCurrentFrame returns the latest expression frame
func (eb *ExpressionBridge) GetCurrentFrame() ExpressionFrame {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.currentFrame
}

// GetMetrics returns bridge performance metrics
func (eb *ExpressionBridge) GetMetrics() ExpressionBridgeMetrics {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return ExpressionBridgeMetrics{
		TotalFrames:     eb.totalFrames,
		AvgFrameTimeUs:  float64(eb.avgFrameTimeNs) / 1000.0,
		PeakFrameTimeUs: float64(eb.peakFrameTimeNs) / 1000.0,
		DroppedFrames:   eb.droppedFrames,
		FrameRate:       eb.frameRate,
		ListenerCount:   len(eb.listeners),
	}
}

// GetDNABridge returns the underlying DNA bridge for direct access
func (eb *ExpressionBridge) GetDNABridge() *DNACognitiveBridge {
	return eb.dnaBridge
}

// ExpressionBridgeMetrics holds performance metrics
type ExpressionBridgeMetrics struct {
	TotalFrames     uint64  `json:"total_frames"`
	AvgFrameTimeUs  float64 `json:"avg_frame_time_us"`
	PeakFrameTimeUs float64 `json:"peak_frame_time_us"`
	DroppedFrames   uint64  `json:"dropped_frames"`
	FrameRate       float64 `json:"frame_rate"`
	ListenerCount   int     `json:"listener_count"`
}

// String returns a human-readable summary
func (m ExpressionBridgeMetrics) String() string {
	return fmt.Sprintf(
		"ExpressionBridge: %d frames, avg=%.1fμs, peak=%.1fμs, dropped=%d, fps=%.0f, listeners=%d",
		m.TotalFrames, m.AvgFrameTimeUs, m.PeakFrameTimeUs,
		m.DroppedFrames, m.FrameRate, m.ListenerCount,
	)
}
