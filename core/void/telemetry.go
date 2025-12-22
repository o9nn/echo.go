// Package void implements telemetry collection and local cores
package void

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TelemetryCollector collects telemetry from all cores.
type TelemetryCollector struct {
	// Observation
	observers    map[string]*Observer
	observations chan *TelemetryObservation

	// Configuration
	samplingRate time.Duration
	bufferSize   int

	// State
	running bool

	// Synchronization
	mu       sync.RWMutex
	stopChan chan struct{}
}

// Observer observes a single core.
type Observer struct {
	ID      string
	CoreID  string
	Metrics []string
	Sampler func() *TelemetryObservation
}

// TelemetryObservation represents a telemetry observation.
type TelemetryObservation struct {
	CoreID    string
	Timestamp time.Time
	Metrics   map[string]interface{}
	State     interface{}
}

// NewTelemetryCollector creates a new telemetry collector.
func NewTelemetryCollector(samplingRate time.Duration, bufferSize int) *TelemetryCollector {
	return &TelemetryCollector{
		observers:    make(map[string]*Observer),
		observations: make(chan *TelemetryObservation, bufferSize),
		samplingRate: samplingRate,
		bufferSize:   bufferSize,
		running:      false,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the telemetry collector.
func (tc *TelemetryCollector) Start(ctx context.Context) error {
	tc.mu.Lock()
	if tc.running {
		tc.mu.Unlock()
		return fmt.Errorf("telemetry collector already running")
	}
	tc.running = true
	tc.mu.Unlock()

	// Start sampling loop
	go tc.run(ctx)

	return nil
}

// Stop stops the telemetry collector.
func (tc *TelemetryCollector) Stop() error {
	tc.mu.Lock()
	if !tc.running {
		tc.mu.Unlock()
		return fmt.Errorf("telemetry collector not running")
	}
	tc.running = false
	tc.mu.Unlock()

	close(tc.stopChan)
	return nil
}

// run is the main sampling loop.
func (tc *TelemetryCollector) run(ctx context.Context) {
	ticker := time.NewTicker(tc.samplingRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tc.stopChan:
			return
		case <-ticker.C:
			tc.sample()
		}
	}
}

// sample samples all observers.
func (tc *TelemetryCollector) sample() {
	tc.mu.RLock()
	observers := make([]*Observer, 0, len(tc.observers))
	for _, obs := range tc.observers {
		observers = append(observers, obs)
	}
	tc.mu.RUnlock()

	for _, obs := range observers {
		observation := obs.Sampler()
		select {
		case tc.observations <- observation:
		default:
			// Buffer full, skip observation
		}
	}
}

// RegisterObserver registers an observer.
func (tc *TelemetryCollector) RegisterObserver(observer *Observer) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if _, exists := tc.observers[observer.ID]; exists {
		return fmt.Errorf("observer %s already registered", observer.ID)
	}

	tc.observers[observer.ID] = observer
	return nil
}

// CollectAll collects all pending observations.
func (tc *TelemetryCollector) CollectAll() []*TelemetryObservation {
	observations := make([]*TelemetryObservation, 0)

	// Drain the observations channel
	for {
		select {
		case obs := <-tc.observations:
			observations = append(observations, obs)
		default:
			return observations
		}
	}
}

// LocalCore represents a local process with inherited execution context.
type LocalCore struct {
	// Identity
	ID   string
	Name string
	Type CoreType

	// Context (inherited from GTS)
	ExecutionContext *ExecutionContext

	// State
	State   interface{}
	History []interface{}

	// Communication
	InputChannels  []*Channel
	OutputChannels []*Channel

	// Processing
	Processor func(context.Context, interface{}) (interface{}, error)

	// Synchronization
	mu sync.RWMutex
}

// CoreType represents the type of a core.
type CoreType int

const (
	CoreAffordance CoreType = iota
	CoreRelevance
	CoreSalience
)

// String returns the string representation of the core type.
func (ct CoreType) String() string {
	switch ct {
	case CoreAffordance:
		return "Affordance"
	case CoreRelevance:
		return "Relevance"
	case CoreSalience:
		return "Salience"
	default:
		return "Unknown"
	}
}

// NewLocalCore creates a new local core.
func NewLocalCore(id string, name string, coreType CoreType) *LocalCore {
	return &LocalCore{
		ID:             id,
		Name:           name,
		Type:           coreType,
		State:          nil,
		History:        make([]interface{}, 0),
		InputChannels:  make([]*Channel, 0),
		OutputChannels: make([]*Channel, 0),
	}
}

// Start starts the local core.
func (lc *LocalCore) Start(ctx context.Context) error {
	// Core-specific initialization
	return nil
}

// Stop stops the local core.
func (lc *LocalCore) Stop() error {
	// Core-specific cleanup
	return nil
}

// Process processes input and produces output.
func (lc *LocalCore) Process(ctx context.Context, input interface{}) (interface{}, error) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.Processor == nil {
		return nil, fmt.Errorf("no processor defined for core %s", lc.ID)
	}

	output, err := lc.Processor(ctx, input)
	if err != nil {
		return nil, err
	}

	// Update state
	lc.State = output

	// Add to history
	lc.History = append(lc.History, output)
	if len(lc.History) > 100 {
		lc.History = lc.History[1:]
	}

	return output, nil
}

// GetTelemetry returns telemetry observation for this core.
func (lc *LocalCore) GetTelemetry() *TelemetryObservation {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	metrics := make(map[string]interface{})
	metrics["type"] = lc.Type.String()
	metrics["history_size"] = len(lc.History)
	metrics["input_channels"] = len(lc.InputChannels)
	metrics["output_channels"] = len(lc.OutputChannels)

	return &TelemetryObservation{
		CoreID:    lc.ID,
		Timestamp: time.Now(),
		Metrics:   metrics,
		State:     lc.State,
	}
}

// ReceiveGestalt receives a gestalt broadcast and updates execution context.
func (lc *LocalCore) ReceiveGestalt(broadcast *GestaltBroadcast) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.ExecutionContext != nil {
		lc.ExecutionContext.Gestalt = &GestaltState{
			globalState:      broadcast.Snapshot.State,
			processGraph:     broadcast.Snapshot.Graph,
			currentTimestamp: broadcast.Timestamp,
		}
	}

	return nil
}

// GetState returns the current state.
func (lc *LocalCore) GetState() interface{} {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.State
}

// Channel represents a communication channel between cores.
type Channel struct {
	// Identity
	ID   string
	Name string

	// Endpoints
	From string // Core ID
	To   string // Core ID

	// Protocol
	Protocol *Protocol

	// Buffer
	Buffer     chan *Message
	BufferSize int

	// State
	IsOpen       bool
	MessageCount int64

	// Synchronization
	mu sync.RWMutex
}

// Message represents a message sent through a channel.
type Message struct {
	ID        string
	From      string
	To        string
	Timestamp time.Time
	Payload   interface{}
	Metadata  map[string]interface{}
}

// NewChannel creates a new channel.
func NewChannel(id string, name string, from string, to string, bufferSize int) *Channel {
	return &Channel{
		ID:           id,
		Name:         name,
		From:         from,
		To:           to,
		Buffer:       make(chan *Message, bufferSize),
		BufferSize:   bufferSize,
		IsOpen:       false,
		MessageCount: 0,
	}
}

// Open opens the channel.
func (ch *Channel) Open() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.IsOpen {
		return fmt.Errorf("channel %s already open", ch.ID)
	}

	ch.IsOpen = true
	return nil
}

// Close closes the channel.
func (ch *Channel) Close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.IsOpen {
		return fmt.Errorf("channel %s already closed", ch.ID)
	}

	ch.IsOpen = false
	close(ch.Buffer)
	return nil
}

// Send sends a message through the channel.
func (ch *Channel) Send(msg *Message) error {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if !ch.IsOpen {
		return fmt.Errorf("channel %s is closed", ch.ID)
	}

	select {
	case ch.Buffer <- msg:
		ch.MessageCount++
		return nil
	default:
		return fmt.Errorf("channel %s buffer full", ch.ID)
	}
}

// Receive receives a message from the channel.
func (ch *Channel) Receive() (*Message, error) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if !ch.IsOpen {
		return nil, fmt.Errorf("channel %s is closed", ch.ID)
	}

	msg, ok := <-ch.Buffer
	if !ok {
		return nil, fmt.Errorf("channel %s closed during receive", ch.ID)
	}

	return msg, nil
}

// Pipe represents a data flow pathway.
type Pipe struct {
	// Identity
	ID        string
	Name      string
	Direction PipeDirection

	// Endpoints
	Source string
	Sink   string

	// Stream
	Stream     chan interface{}
	StreamSize int

	// State
	IsOpen    bool
	DataCount int64

	// Synchronization
	mu sync.RWMutex
}

// PipeDirection represents the direction of a pipe.
type PipeDirection int

const (
	PipeToGTS PipeDirection = iota
	PipeFromGTS
	PipeBidirectional
)

// NewPipe creates a new pipe.
func NewPipe(id string, name string, source string, sink string, direction PipeDirection, streamSize int) *Pipe {
	return &Pipe{
		ID:         id,
		Name:       name,
		Direction:  direction,
		Source:     source,
		Sink:       sink,
		Stream:     make(chan interface{}, streamSize),
		StreamSize: streamSize,
		IsOpen:     false,
		DataCount:  0,
	}
}

// Open opens the pipe.
func (p *Pipe) Open() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.IsOpen {
		return fmt.Errorf("pipe %s already open", p.ID)
	}

	p.IsOpen = true
	return nil
}

// Close closes the pipe.
func (p *Pipe) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.IsOpen {
		return fmt.Errorf("pipe %s already closed", p.ID)
	}

	p.IsOpen = false
	close(p.Stream)
	return nil
}

// Write writes data to the pipe.
func (p *Pipe) Write(data interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.IsOpen {
		return fmt.Errorf("pipe %s is closed", p.ID)
	}

	select {
	case p.Stream <- data:
		p.DataCount++
		return nil
	default:
		return fmt.Errorf("pipe %s stream full", p.ID)
	}
}

// Read reads data from the pipe.
func (p *Pipe) Read() (interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.IsOpen {
		return nil, fmt.Errorf("pipe %s is closed", p.ID)
	}

	data, ok := <-p.Stream
	if !ok {
		return nil, fmt.Errorf("pipe %s closed during read", p.ID)
	}

	return data, nil
}
