package webserver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/metahuman"
)

// =============================================================================
// EXPRESSION STREAM
// =============================================================================
//
// Provides WebSocket streaming of MetaHuman DNA expression frames to connected
// avatar renderers (Live2D, Three.js, Unreal Engine, etc.).
//
// Protocol:
//   - Channel: "expression"
//   - Message type: "expression_frame"
//   - Data: ExpressionStreamFrame (compact JSON)
//
// Clients subscribe to the "expression" channel to receive real-time
// expression frames at the configured frame rate.
//
// =============================================================================

// ExpressionStreamFrame is the compact wire format for expression frames
type ExpressionStreamFrame struct {
	// Frame metadata
	FrameID   uint64  `json:"fid"`
	Timestamp int64   `json:"ts"`  // Unix millis
	DeltaMs   float64 `json:"dt"`  // Time since last frame

	// FACS Action Units (compact: only non-zero values)
	AUs map[string]float64 `json:"aus,omitempty"`

	// MetaHuman morph targets (compact: only non-zero values)
	Morphs map[string]float64 `json:"morphs,omitempty"`

	// Material parameters (compact: only non-zero values)
	Materials map[string]float64 `json:"mats,omitempty"`

	// Cognitive context
	Mode       string  `json:"mode,omitempty"`
	Lyapunov   float64 `json:"lyapunov,omitempty"`
}

// ExpressionStream manages streaming of expression frames to WebSocket clients
type ExpressionStream struct {
	mu sync.RWMutex

	// WebSocket hub for broadcasting
	hub *WebSocketHub

	// Frame state
	lastFrameTime time.Time
	frameCount    uint64

	// Throttling
	minFrameInterval time.Duration // minimum time between broadcasts
	lastBroadcast    time.Time

	// Metrics
	broadcastCount uint64
	droppedCount   uint64
}

// NewExpressionStream creates a new expression stream
func NewExpressionStream(hub *WebSocketHub) *ExpressionStream {
	return &ExpressionStream{
		hub:              hub,
		minFrameInterval: time.Millisecond * 33, // ~30fps max broadcast rate
	}
}

// SetMaxBroadcastRate sets the maximum broadcast rate in fps
func (es *ExpressionStream) SetMaxBroadcastRate(fps float64) {
	es.mu.Lock()
	defer es.mu.Unlock()
	if fps > 0 && fps <= 120 {
		es.minFrameInterval = time.Duration(float64(time.Second) / fps)
	}
}

// OnExpressionFrame implements ExpressionFrameListener for the expression bridge
func (es *ExpressionStream) OnExpressionFrame(frame metahuman.ExpressionFrame) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.frameCount++

	// Throttle broadcasts
	now := time.Now()
	if now.Sub(es.lastBroadcast) < es.minFrameInterval {
		es.droppedCount++
		return
	}
	es.lastBroadcast = now

	// Build compact wire frame
	wireFrame := es.buildWireFrame(frame)

	// Broadcast to WebSocket clients
	if es.hub != nil {
		data, err := json.Marshal(wireFrame)
		if err != nil {
			return
		}

		msg := &WebSocketMessage{
			Type:      "expression_frame",
			Channel:   "expression",
			Data:      json.RawMessage(data),
			Timestamp: now,
		}

		// Non-blocking send
		select {
		case es.hub.broadcast <- msg:
			es.broadcastCount++
		default:
			es.droppedCount++
		}
	}
}

// buildWireFrame converts a full ExpressionFrame to compact wire format
func (es *ExpressionStream) buildWireFrame(frame metahuman.ExpressionFrame) ExpressionStreamFrame {
	dt := float64(0)
	if !es.lastFrameTime.IsZero() {
		dt = time.Since(es.lastFrameTime).Seconds() * 1000
	}
	es.lastFrameTime = time.Now()

	wire := ExpressionStreamFrame{
		FrameID:   es.frameCount,
		Timestamp: time.Now().UnixMilli(),
		DeltaMs:   dt,
		Mode:      frame.CognitiveMode,
		Lyapunov:  roundTo3(frame.LyapunovExponent),
	}

	// Only include non-zero AU values (compact)
	aus := make(map[string]float64)
	for au, val := range frame.ActionUnits {
		if val > 0.01 { // threshold for wire transmission
			code := metahuman.ActionUnitCode(au)
			aus[code] = roundTo3(val)
		}
	}
	if len(aus) > 0 {
		wire.AUs = aus
	}

	// Only include non-zero morph targets
	morphs := make(map[string]float64)
	for name, val := range frame.MorphTargets {
		if val > 0.01 || val < -0.01 {
			morphs[name] = roundTo3(val)
		}
	}
	if len(morphs) > 0 {
		wire.Morphs = morphs
	}

	// Only include non-zero material params
	mats := make(map[string]float64)
	for name, val := range frame.MaterialParams {
		if val > 0.01 || val < -0.01 {
			mats[name] = roundTo3(val)
		}
	}
	if len(mats) > 0 {
		wire.Materials = mats
	}

	return wire
}

// GetMetrics returns streaming metrics
func (es *ExpressionStream) GetMetrics() ExpressionStreamMetrics {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return ExpressionStreamMetrics{
		TotalFrames:    es.frameCount,
		BroadcastCount: es.broadcastCount,
		DroppedCount:   es.droppedCount,
		MaxFPS:         1.0 / es.minFrameInterval.Seconds(),
	}
}

// ExpressionStreamMetrics holds streaming performance metrics
type ExpressionStreamMetrics struct {
	TotalFrames    uint64  `json:"total_frames"`
	BroadcastCount uint64  `json:"broadcast_count"`
	DroppedCount   uint64  `json:"dropped_count"`
	MaxFPS         float64 `json:"max_fps"`
}

// String returns a human-readable summary
func (m ExpressionStreamMetrics) String() string {
	return fmt.Sprintf(
		"ExpressionStream: %d frames, %d broadcasts, %d dropped, max=%.0ffps",
		m.TotalFrames, m.BroadcastCount, m.DroppedCount, m.MaxFPS,
	)
}

// roundTo3 rounds a float to 3 decimal places for compact wire format
func roundTo3(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}
