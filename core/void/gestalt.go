// Package void implements the gestalt state - the integrated awareness
package void

import (
	"sync"
	"time"
)

// GestaltState represents the integrated awareness - the unified whole.
type GestaltState struct {
	// Global integrated state
	globalState map[string]interface{}

	// Cross-process awareness
	processGraph *ProcessGraph

	// Temporal coherence
	history          *StateHistory
	currentTimestamp time.Time

	// Semantic unity
	sharedContext *SharedContext

	// Synchronization
	mu sync.RWMutex
}

// ProcessGraph represents the graph of processes and their relationships.
type ProcessGraph struct {
	Nodes map[string]*ProcessNode
	Edges map[string]*ProcessEdge
	mu    sync.RWMutex
}

// ProcessNode represents a node in the process graph.
type ProcessNode struct {
	CoreID    string
	State     interface{}
	Timestamp time.Time
}

// ProcessEdge represents an edge in the process graph.
type ProcessEdge struct {
	ID      string
	From    string
	To      string
	Channel string
	Weight  float64
}

// StateHistory maintains a history of gestalt snapshots.
type StateHistory struct {
	Snapshots []GestaltSnapshot
	MaxSize   int
	mu        sync.RWMutex
}

// GestaltSnapshot represents a snapshot of the gestalt state.
type GestaltSnapshot struct {
	Timestamp time.Time
	State     map[string]interface{}
	Graph     *ProcessGraph
}

// NewGestaltState creates a new gestalt state.
func NewGestaltState(maxHistorySize int) *GestaltState {
	return &GestaltState{
		globalState:      make(map[string]interface{}),
		processGraph:     NewProcessGraph(),
		history:          NewStateHistory(maxHistorySize),
		currentTimestamp: time.Now(),
		sharedContext:    &SharedContext{Semantics: make(map[string]interface{}), Ontology: NewOntology()},
	}
}

// Integrate integrates telemetry observations into the gestalt.
func (gs *GestaltState) Integrate(observations []*TelemetryObservation) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.currentTimestamp = time.Now()

	for _, obs := range observations {
		// Update global state
		gs.globalState[obs.CoreID] = obs.State

		// Update process graph node
		gs.processGraph.UpdateNode(&ProcessNode{
			CoreID:    obs.CoreID,
			State:     obs.State,
			Timestamp: obs.Timestamp,
		})

		// Update metrics in global state
		for key, value := range obs.Metrics {
			metricKey := obs.CoreID + "." + key
			gs.globalState[metricKey] = value
		}
	}

	return nil
}

// Snapshot creates a snapshot of the current gestalt state.
func (gs *GestaltState) Snapshot() *GestaltSnapshot {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Deep copy global state
	stateCopy := make(map[string]interface{})
	for k, v := range gs.globalState {
		stateCopy[k] = v
	}

	// Copy process graph
	graphCopy := gs.processGraph.Copy()

	snapshot := &GestaltSnapshot{
		Timestamp: gs.currentTimestamp,
		State:     stateCopy,
		Graph:     graphCopy,
	}

	// Add to history
	gs.history.Add(snapshot)

	return snapshot
}

// GetGlobalState returns the current global state.
func (gs *GestaltState) GetGlobalState() map[string]interface{} {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Return a copy
	stateCopy := make(map[string]interface{})
	for k, v := range gs.globalState {
		stateCopy[k] = v
	}
	return stateCopy
}

// GetProcessGraph returns the process graph.
func (gs *GestaltState) GetProcessGraph() *ProcessGraph {
	return gs.processGraph
}

// GetHistory returns the state history.
func (gs *GestaltState) GetHistory() *StateHistory {
	return gs.history
}

// GetSharedContext returns the shared context.
func (gs *GestaltState) GetSharedContext() *SharedContext {
	return gs.sharedContext
}

// NewProcessGraph creates a new process graph.
func NewProcessGraph() *ProcessGraph {
	return &ProcessGraph{
		Nodes: make(map[string]*ProcessNode),
		Edges: make(map[string]*ProcessEdge),
	}
}

// UpdateNode updates or adds a node in the process graph.
func (pg *ProcessGraph) UpdateNode(node *ProcessNode) {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	pg.Nodes[node.CoreID] = node
}

// AddEdge adds an edge to the process graph.
func (pg *ProcessGraph) AddEdge(edge *ProcessEdge) {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	pg.Edges[edge.ID] = edge
}

// GetNode retrieves a node by core ID.
func (pg *ProcessGraph) GetNode(coreID string) (*ProcessNode, bool) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	node, exists := pg.Nodes[coreID]
	return node, exists
}

// GetEdge retrieves an edge by ID.
func (pg *ProcessGraph) GetEdge(id string) (*ProcessEdge, bool) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()
	edge, exists := pg.Edges[id]
	return edge, exists
}

// Copy creates a deep copy of the process graph.
func (pg *ProcessGraph) Copy() *ProcessGraph {
	pg.mu.RLock()
	defer pg.mu.RUnlock()

	copy := NewProcessGraph()

	for id, node := range pg.Nodes {
		copy.Nodes[id] = &ProcessNode{
			CoreID:    node.CoreID,
			State:     node.State,
			Timestamp: node.Timestamp,
		}
	}

	for id, edge := range pg.Edges {
		copy.Edges[id] = &ProcessEdge{
			ID:      edge.ID,
			From:    edge.From,
			To:      edge.To,
			Channel: edge.Channel,
			Weight:  edge.Weight,
		}
	}

	return copy
}

// NewStateHistory creates a new state history.
func NewStateHistory(maxSize int) *StateHistory {
	return &StateHistory{
		Snapshots: make([]GestaltSnapshot, 0, maxSize),
		MaxSize:   maxSize,
	}
}

// Add adds a snapshot to the history.
func (sh *StateHistory) Add(snapshot *GestaltSnapshot) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.Snapshots = append(sh.Snapshots, *snapshot)

	// Trim if exceeds max size
	if len(sh.Snapshots) > sh.MaxSize {
		sh.Snapshots = sh.Snapshots[1:]
	}
}

// GetLatest returns the latest snapshot.
func (sh *StateHistory) GetLatest() (*GestaltSnapshot, bool) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	if len(sh.Snapshots) == 0 {
		return nil, false
	}

	return &sh.Snapshots[len(sh.Snapshots)-1], true
}

// GetAll returns all snapshots.
func (sh *StateHistory) GetAll() []GestaltSnapshot {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	// Return a copy
	snapshots := make([]GestaltSnapshot, len(sh.Snapshots))
	copy(snapshots, sh.Snapshots)
	return snapshots
}

// GetRange returns snapshots within a time range.
func (sh *StateHistory) GetRange(start, end time.Time) []GestaltSnapshot {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	result := make([]GestaltSnapshot, 0)
	for _, snapshot := range sh.Snapshots {
		if snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end) {
			result = append(result, snapshot)
		}
	}
	return result
}
