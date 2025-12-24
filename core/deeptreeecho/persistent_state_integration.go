package deeptreeecho

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistentStateIntegration connects the PersistentCognitiveState with
// UnifiedCognitiveLoopV2 to enable state persistence across wake/rest cycles.
// This allows Deep Tree Echo to maintain continuity of identity and wisdom.
type PersistentStateIntegration struct {
	mu sync.RWMutex

	// Persistent state manager
	persistentState *PersistentCognitiveState

	// Reference to the cognitive loop (set during integration)
	cognitiveLoop *UnifiedCognitiveLoopV2

	// Reference to echobeats unified (optional)
	echobeatsUnified *EchobeatsUnified

	// State file path
	stateFilePath string

	// Auto-save configuration
	autoSaveEnabled  bool
	autoSaveInterval time.Duration
	lastSaveTime     time.Time

	// State change tracking
	changesSinceLastSave int
	saveThreshold        int

	// Running state
	running bool
	stopCh  chan struct{}
}

// PersistentStateSnapshot represents a complete snapshot of the cognitive state
type PersistentStateSnapshot struct {
	// Metadata
	Version     int       `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	SessionID   string    `json:"session_id"`
	
	// Core state from PersistentCognitiveState
	CoreState   CognitiveStateData `json:"core_state"`
	
	// Cognitive loop metrics
	LoopMetrics LoopMetricsSnapshot `json:"loop_metrics"`
	
	// Echobeats state
	EchobeatsState EchobeatsStateSnapshot `json:"echobeats_state"`
	
	// Wake/rest history
	WakeRestHistory []WakeRestEvent `json:"wake_rest_history"`
	
	// Conversation history summary
	ConversationSummary []ConversationSummaryEntry `json:"conversation_summary"`
}

// LoopMetricsSnapshot captures cognitive loop metrics
type LoopMetricsSnapshot struct {
	TotalCycles          uint64  `json:"total_cycles"`
	TotalEvents          uint64  `json:"total_events"`
	WisdomGained         float64 `json:"wisdom_gained"`
	InsightsGained       uint64  `json:"insights_gained"`
	ConversationsEngaged uint64  `json:"conversations_engaged"`
	CognitiveLoad        float64 `json:"cognitive_load"`
	WisdomLevel          float64 `json:"wisdom_level"`
	AwarenessLevel       float64 `json:"awareness_level"`
}

// EchobeatsStateSnapshot captures echobeats state
type EchobeatsStateSnapshot struct {
	CurrentStep       int                    `json:"current_step"`
	CurrentMode       string                 `json:"current_mode"`
	CycleCount        uint64                 `json:"cycle_count"`
	TotalSteps        uint64                 `json:"total_steps"`
	ActiveGoals       []string               `json:"active_goals"`
	CompletedGoals    []string               `json:"completed_goals"`
	EngineStates      []map[string]interface{} `json:"engine_states"`
}

// WakeRestEvent records a wake/rest transition
type WakeRestEvent struct {
	Timestamp time.Time `json:"timestamp"`
	FromState string    `json:"from_state"`
	ToState   string    `json:"to_state"`
	Reason    string    `json:"reason"`
}

// ConversationSummaryEntry summarizes a conversation
type ConversationSummaryEntry struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Topic      string    `json:"topic"`
	Duration   int64     `json:"duration_seconds"`
	Insights   []string  `json:"insights"`
	Importance float64   `json:"importance"`
}

// NewPersistentStateIntegration creates a new persistent state integration
func NewPersistentStateIntegration(stateDir string) *PersistentStateIntegration {
	stateFilePath := filepath.Join(stateDir, "deep_tree_echo_state.json")

	return &PersistentStateIntegration{
		persistentState:  NewPersistentCognitiveState(stateFilePath),
		stateFilePath:    stateFilePath,
		autoSaveEnabled:  true,
		autoSaveInterval: 5 * time.Minute,
		saveThreshold:    100, // Save after 100 changes
		stopCh:           make(chan struct{}),
	}
}

// IntegrateWithCognitiveLoop connects the persistent state to a cognitive loop
func (psi *PersistentStateIntegration) IntegrateWithCognitiveLoop(loop *UnifiedCognitiveLoopV2) error {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	psi.cognitiveLoop = loop

	// Load existing state if available
	if err := psi.loadState(); err != nil {
		fmt.Printf("⚠️  No existing state found, starting fresh: %v\n", err)
	} else {
		fmt.Println("✅ Loaded persistent cognitive state")
		psi.applyStateToLoop()
	}

	return nil
}

// IntegrateWithEchobeats connects the persistent state to echobeats
func (psi *PersistentStateIntegration) IntegrateWithEchobeats(echobeats *EchobeatsUnified) error {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	psi.echobeatsUnified = echobeats
	return nil
}

// Start begins the auto-save routine
func (psi *PersistentStateIntegration) Start() error {
	psi.mu.Lock()
	if psi.running {
		psi.mu.Unlock()
		return fmt.Errorf("persistent state integration already running")
	}
	psi.running = true
	psi.mu.Unlock()

	// Start auto-save goroutine
	go psi.autoSaveRoutine()

	fmt.Println("💾 Persistent state integration started")
	return nil
}

// Stop halts the auto-save routine and performs final save
func (psi *PersistentStateIntegration) Stop() error {
	psi.mu.Lock()
	if !psi.running {
		psi.mu.Unlock()
		return fmt.Errorf("persistent state integration not running")
	}
	psi.running = false
	close(psi.stopCh)
	psi.mu.Unlock()

	// Final save
	if err := psi.SaveState(); err != nil {
		return fmt.Errorf("failed to save final state: %w", err)
	}

	fmt.Println("💾 Persistent state integration stopped")
	return nil
}

// autoSaveRoutine periodically saves state
func (psi *PersistentStateIntegration) autoSaveRoutine() {
	ticker := time.NewTicker(psi.autoSaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-psi.stopCh:
			return
		case <-ticker.C:
			if psi.autoSaveEnabled && psi.changesSinceLastSave > 0 {
				if err := psi.SaveState(); err != nil {
					fmt.Printf("⚠️  Auto-save failed: %v\n", err)
				}
			}
		}
	}
}

// RecordChange records a state change and triggers save if threshold reached
func (psi *PersistentStateIntegration) RecordChange() {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	psi.changesSinceLastSave++

	if psi.changesSinceLastSave >= psi.saveThreshold {
		go func() {
			if err := psi.SaveState(); err != nil {
				fmt.Printf("⚠️  Threshold save failed: %v\n", err)
			}
		}()
	}
}

// SaveState saves the current cognitive state to disk
func (psi *PersistentStateIntegration) SaveState() error {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	snapshot := psi.createSnapshot()

	// Ensure directory exists
	dir := filepath.Dir(psi.stateFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(psi.stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	psi.lastSaveTime = time.Now()
	psi.changesSinceLastSave = 0

	fmt.Printf("💾 State saved to %s\n", psi.stateFilePath)
	return nil
}

// loadState loads the cognitive state from disk
func (psi *PersistentStateIntegration) loadState() error {
	data, err := os.ReadFile(psi.stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var snapshot PersistentStateSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Apply loaded state to persistent state manager
	psi.persistentState.State = snapshot.CoreState

	return nil
}

// applyStateToLoop applies the loaded state to the cognitive loop
func (psi *PersistentStateIntegration) applyStateToLoop() {
	if psi.cognitiveLoop == nil {
		return
	}

	// Apply wisdom level
	psi.cognitiveLoop.mu.Lock()
	psi.cognitiveLoop.wisdomLevel = psi.persistentState.State.WisdomLevel
	psi.cognitiveLoop.mu.Unlock()

	// Apply interests to interest patterns
	// Restore interests from persistent state
	psi.cognitiveLoop.interestPatterns.RestoreInterests(psi.persistentState.State.Interests)

	// Apply active goals
	for _, goal := range psi.persistentState.State.ActiveGoals {
		psi.cognitiveLoop.echobeatsScheduler.AddGoal(goal, 0.5)
	}

	fmt.Printf("✅ Applied persistent state: wisdom=%.3f, interests=%d, goals=%d\n",
		psi.persistentState.State.WisdomLevel,
		len(psi.persistentState.State.Interests),
		len(psi.persistentState.State.ActiveGoals))
}

// createSnapshot creates a complete snapshot of the current state
func (psi *PersistentStateIntegration) createSnapshot() PersistentStateSnapshot {
	snapshot := PersistentStateSnapshot{
		Version:   1,
		Timestamp: time.Now(),
		SessionID: fmt.Sprintf("session_%d", time.Now().Unix()),
	}

	// Get core state
	snapshot.CoreState = psi.persistentState.State

	// Get loop metrics if available
	if psi.cognitiveLoop != nil {
		psi.cognitiveLoop.mu.RLock()
		snapshot.LoopMetrics = LoopMetricsSnapshot{
			TotalCycles:          psi.cognitiveLoop.totalCycles,
			TotalEvents:          psi.cognitiveLoop.totalEvents,
			WisdomGained:         psi.cognitiveLoop.wisdomGained,
			InsightsGained:       psi.cognitiveLoop.insightsGained,
			ConversationsEngaged: psi.cognitiveLoop.conversationsEngaged,
			CognitiveLoad:        psi.cognitiveLoop.cognitiveLoad,
			WisdomLevel:          psi.cognitiveLoop.wisdomLevel,
			AwarenessLevel:       psi.cognitiveLoop.awarenessLevel,
		}
		psi.cognitiveLoop.mu.RUnlock()
	}

	// Get echobeats state if available
	if psi.echobeatsUnified != nil {
		metrics := psi.echobeatsUnified.GetMetrics()
		snapshot.EchobeatsState = EchobeatsStateSnapshot{
			CurrentStep: metrics["current_step"].(int),
			CurrentMode: string(metrics["current_mode"].(UnifiedCognitiveMode)),
			CycleCount:  metrics["total_cycles"].(uint64),
			TotalSteps:  metrics["total_steps"].(uint64),
		}
	}

	return snapshot
}

// RecordWakeRestTransition records a wake/rest transition
func (psi *PersistentStateIntegration) RecordWakeRestTransition(fromState, toState, reason string) {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	event := WakeRestEvent{
		Timestamp: time.Now(),
		FromState: fromState,
		ToState:   toState,
		Reason:    reason,
	}

	// Update persistent state
	psi.persistentState.State.LastWakeTime = time.Now().Unix()
	psi.changesSinceLastSave++

	fmt.Printf("🔄 Wake/rest transition: %s → %s (%s)\n", fromState, toState, reason)
	_ = event // Store in history when implementing full persistence
}

// RecordConversation records a conversation summary
func (psi *PersistentStateIntegration) RecordConversation(topic string, duration time.Duration, insights []string, importance float64) {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	entry := ConversationSummaryEntry{
		ID:         fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Topic:      topic,
		Duration:   int64(duration.Seconds()),
		Insights:   insights,
		Importance: importance,
	}

	// Update interests based on conversation
	psi.persistentState.State.Interests[topic] = importance

	// Add insights as memories
	for _, insight := range insights {
		psi.persistentState.AddMemory(insight, importance, []string{"conversation", topic})
	}

	psi.changesSinceLastSave++
	_ = entry // Store in history when implementing full persistence
}

// RecordWisdom records a wisdom principle
func (psi *PersistentStateIntegration) RecordWisdom(principle string, importance float64) {
	psi.mu.Lock()
	defer psi.mu.Unlock()

	psi.persistentState.AddWisdomPrinciple(principle, importance)
	psi.changesSinceLastSave++
}

// GetWisdomLevel returns the current wisdom level
func (psi *PersistentStateIntegration) GetWisdomLevel() float64 {
	psi.mu.RLock()
	defer psi.mu.RUnlock()

	return psi.persistentState.State.WisdomLevel
}

// GetTopInterests returns the top interests
func (psi *PersistentStateIntegration) GetTopInterests(n int) map[string]float64 {
	psi.mu.RLock()
	defer psi.mu.RUnlock()

	// Simple implementation - return all interests
	// TODO: Sort and return top N
	return psi.persistentState.State.Interests
}

// GetRecentMemories returns recent memories
func (psi *PersistentStateIntegration) GetRecentMemories(n int) []CognitiveMemory {
	psi.mu.RLock()
	defer psi.mu.RUnlock()

	memories := psi.persistentState.State.RecentMemories
	if len(memories) <= n {
		return memories
	}
	return memories[len(memories)-n:]
}

// GetStateFilePath returns the path to the state file
func (psi *PersistentStateIntegration) GetStateFilePath() string {
	return psi.stateFilePath
}
