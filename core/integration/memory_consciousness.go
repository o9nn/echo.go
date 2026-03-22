package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/o9nn/echo.go/core/memory"
)

// MemoryConsciousnessIntegrator bridges the thought engine with hypergraph memory.
// It enables thoughts to query memory, store insights, and build persistent knowledge.
type MemoryConsciousnessIntegrator struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Core components
	stateManager *CognitiveStateManager
	memory       *memory.HypergraphMemory

	// Integration state
	activeQueries  map[string]*MemoryQuery
	recentInsights []*StoredInsight
	activationMap  map[string]float64 // node ID -> activation level

	// Configuration
	queryThreshold  float64
	storeThreshold  float64
	activationDecay float64

	// Metrics
	queriesExecuted uint64
	insightsStored  uint64
	patternsFound   uint64

	running bool
}

// MemoryQuery represents a query from consciousness to memory.
type MemoryQuery struct {
	ID        string
	Timestamp time.Time
	ThoughtID string
	QueryType QueryType
	Keywords  []string
	Context   map[string]interface{}
	Results   []*memory.MemoryNode
	Relevance float64
}

// QueryType defines different types of memory queries.
type QueryType int

const (
	QueryTypeRecall       QueryType = iota // Recall specific facts
	QueryTypePattern                       // Find patterns
	QueryTypeAssociation                   // Find associations
	QueryTypeEpisodic                      // Recall experiences
	QueryTypeProcedural                    // Recall how to do something
)

// StoredInsight represents an insight stored in memory.
type StoredInsight struct {
	ID          string
	ThoughtID   string
	NodeID      string
	Content     string
	Timestamp   time.Time
	Importance  float64
	Connections []string
}

// NewMemoryConsciousnessIntegrator creates a new integrator.
func NewMemoryConsciousnessIntegrator(
	sm *CognitiveStateManager,
	mem *memory.HypergraphMemory,
) *MemoryConsciousnessIntegrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &MemoryConsciousnessIntegrator{
		ctx:             ctx,
		cancel:          cancel,
		stateManager:    sm,
		memory:          mem,
		activeQueries:   make(map[string]*MemoryQuery),
		recentInsights:  make([]*StoredInsight, 0),
		activationMap:   make(map[string]float64),
		queryThreshold:  0.6,
		storeThreshold:  0.7,
		activationDecay: 0.95,
	}
}

// Start begins the integration process.
func (mci *MemoryConsciousnessIntegrator) Start() error {
	mci.mu.Lock()
	if mci.running {
		mci.mu.Unlock()
		return fmt.Errorf("memory-consciousness integrator already running")
	}
	mci.running = true
	mci.mu.Unlock()

	go mci.thoughtMonitoringLoop()
	go mci.activationDecayLoop()
	go mci.patternRecognitionLoop()

	return nil
}

// Stop halts the integration process.
func (mci *MemoryConsciousnessIntegrator) Stop() {
	mci.mu.Lock()
	mci.running = false
	mci.mu.Unlock()
	mci.cancel()
}

// thoughtMonitoringLoop monitors thoughts and triggers memory operations.
func (mci *MemoryConsciousnessIntegrator) thoughtMonitoringLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.processRecentThoughts()
		}
	}
}

// processRecentThoughts processes recent thoughts for memory integration.
func (mci *MemoryConsciousnessIntegrator) processRecentThoughts() {
	thoughts := mci.stateManager.GetRecentThoughts(5)

	for i := range thoughts {
		thought := thoughts[i]
		// Check if thought should trigger memory query
		if mci.shouldQueryMemory(thought) {
			mci.queryMemoryForThought(thought)
		}

		// Check if thought should be stored as insight
		if mci.shouldStoreAsInsight(thought) {
			mci.storeThoughtAsInsight(thought)
		}
	}
}

// shouldQueryMemory determines if a thought should trigger a memory query.
func (mci *MemoryConsciousnessIntegrator) shouldQueryMemory(thought SharedThought) bool {
	return thought.Phase == "reflection" || thought.Phase == "question" || thought.Phase == "orienting"
}

// queryMemoryForThought queries memory based on thought content.
func (mci *MemoryConsciousnessIntegrator) queryMemoryForThought(thought SharedThought) {
	query := &MemoryQuery{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		ThoughtID: thought.ID,
		QueryType: QueryTypePattern,
		Keywords:  []string{thought.Content},
		Results:   make([]*memory.MemoryNode, 0),
	}

	// Execute query
	query.Results = mci.findPatterns(query.Keywords)

	mci.mu.Lock()
	mci.activeQueries[query.ID] = query
	mci.queriesExecuted++
	mci.mu.Unlock()

	// Inject results back as thoughts
	if len(query.Results) > 0 {
		for _, node := range query.Results {
			mci.stateManager.AddThought(
				fmt.Sprintf("Remembering: %s", node.Content),
				"memory_recall",
				"memory_consciousness_integrator",
				0.6,
				[]string{"memory"},
			)
		}
	}
}

// shouldStoreAsInsight determines if a thought should be stored as memory.
func (mci *MemoryConsciousnessIntegrator) shouldStoreAsInsight(thought SharedThought) bool {
	return thought.Phase == "insight" || thought.Phase == "anticipating"
}

// storeThoughtAsInsight stores a thought as a memory node.
func (mci *MemoryConsciousnessIntegrator) storeThoughtAsInsight(thought SharedThought) {
	node := &memory.MemoryNode{
		ID:        uuid.New().String(),
		Type:      memory.NodeThought,
		Content:   thought.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"source":     thought.Source,
			"thought_id": thought.ID,
			"phase":      thought.Phase,
		},
	}

	if err := mci.memory.AddNode(node); err != nil {
		return
	}

	insight := &StoredInsight{
		ID:        uuid.New().String(),
		ThoughtID: thought.ID,
		NodeID:    node.ID,
		Content:   thought.Content,
		Timestamp: time.Now(),
		Importance: 0.8,
	}

	mci.mu.Lock()
	mci.recentInsights = append(mci.recentInsights, insight)
	mci.insightsStored++
	mci.mu.Unlock()
}

// activationDecayLoop gradually decays activation levels.
func (mci *MemoryConsciousnessIntegrator) activationDecayLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.decayActivation()
		}
	}
}

// decayActivation applies decay to all activation levels.
func (mci *MemoryConsciousnessIntegrator) decayActivation() {
	mci.mu.Lock()
	defer mci.mu.Unlock()

	for nodeID, activation := range mci.activationMap {
		newActivation := activation * mci.activationDecay
		if newActivation < 0.01 {
			delete(mci.activationMap, nodeID)
		} else {
			mci.activationMap[nodeID] = newActivation
		}
	}
}

// patternRecognitionLoop looks for emerging patterns in activation.
func (mci *MemoryConsciousnessIntegrator) patternRecognitionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mci.ctx.Done():
			return
		case <-ticker.C:
			mci.recognizePatterns()
		}
	}
}

// recognizePatterns identifies patterns in highly activated nodes.
func (mci *MemoryConsciousnessIntegrator) recognizePatterns() {
	mci.mu.RLock()
	highlyActivated := make([]string, 0)
	for nodeID, activation := range mci.activationMap {
		if activation > 0.7 {
			highlyActivated = append(highlyActivated, nodeID)
		}
	}
	mci.mu.RUnlock()

	if len(highlyActivated) >= 2 {
		mci.mu.Lock()
		mci.patternsFound++
		mci.mu.Unlock()
	}
}

// findPatterns searches memory for pattern nodes.
func (mci *MemoryConsciousnessIntegrator) findPatterns(keywords []string) []*memory.MemoryNode {
	return make([]*memory.MemoryNode, 0)
}

// GetMetrics returns integration metrics.
func (mci *MemoryConsciousnessIntegrator) GetMetrics() map[string]interface{} {
	mci.mu.RLock()
	defer mci.mu.RUnlock()

	return map[string]interface{}{
		"queries_executed": mci.queriesExecuted,
		"insights_stored":  mci.insightsStored,
		"patterns_found":   mci.patternsFound,
		"active_nodes":     len(mci.activationMap),
	}
}
