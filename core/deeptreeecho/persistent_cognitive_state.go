package deeptreeecho

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistentCognitiveState manages the persistent state of the cognitive system
// This enables the agent to maintain continuity across wake/rest cycles
type PersistentCognitiveState struct {
	mu sync.RWMutex

	// File path for persistence
	statePath string

	// Core state
	State CognitiveStateData

	// Auto-save interval
	autoSaveInterval time.Duration
	lastSave         time.Time
}

// CognitiveStateData represents the serializable cognitive state
type CognitiveStateData struct {
	// Identity
	Identity           string   `json:"identity"`
	CoreValues         []string `json:"core_values"`
	
	// Temporal state
	CreatedAt          int64    `json:"created_at"`
	LastWakeTime       int64    `json:"last_wake_time"`
	LastRestTime       int64    `json:"last_rest_time"`
	TotalWakeTime      int64    `json:"total_wake_time"`
	TotalRestTime      int64    `json:"total_rest_time"`
	WakeRestCycles     int      `json:"wake_rest_cycles"`
	
	// Cognitive metrics
	WisdomLevel        float64  `json:"wisdom_level"`
	AwarenessLevel     float64  `json:"awareness_level"`
	CognitiveLoad      float64  `json:"cognitive_load"`
	
	// Experience counters
	TotalThoughts      uint64   `json:"total_thoughts"`
	TotalInteractions  uint64   `json:"total_interactions"`
	TotalDreams        uint64   `json:"total_dreams"`
	TotalInsights      uint64   `json:"total_insights"`
	
	// Knowledge state
	KnowledgeGaps      map[string]float64 `json:"knowledge_gaps"`
	Interests          map[string]float64 `json:"interests"`
	ActiveGoals        []string           `json:"active_goals"`
	
	// Wisdom principles learned
	WisdomPrinciples   []PersistentWisdomPrinciple `json:"wisdom_principles"`
	
	// Recent memories for continuity
	RecentMemories     []CognitiveMemory `json:"recent_memories"`
	
	// Semantic network summary
	SemanticConcepts   []string `json:"semantic_concepts"`
	
	// Current mood and focus
	CurrentMood        string   `json:"current_mood"`
	CurrentFocus       string   `json:"current_focus"`
	
	// Version for migration
	Version            int      `json:"version"`
}

// PersistentWisdomPrinciple represents a wisdom principle that persists
type PersistentWisdomPrinciple struct {
	ID          string  `json:"id"`
	Principle   string  `json:"principle"`
	Confidence  float64 `json:"confidence"`
	Applications int    `json:"applications"`
	CreatedAt   int64   `json:"created_at"`
}

// CognitiveMemory represents a memory that persists across sessions
type CognitiveMemory struct {
	ID          string  `json:"id"`
	Content     string  `json:"content"`
	Importance  float64 `json:"importance"`
	Timestamp   int64   `json:"timestamp"`
	Tags        []string `json:"tags"`
}

// NewPersistentCognitiveState creates a new persistent state manager
func NewPersistentCognitiveState(statePath string) *PersistentCognitiveState {
	pcs := &PersistentCognitiveState{
		statePath:        statePath,
		autoSaveInterval: 5 * time.Minute,
		State: CognitiveStateData{
			CreatedAt:        time.Now().Unix(),
			KnowledgeGaps:    make(map[string]float64),
			Interests:        make(map[string]float64),
			ActiveGoals:      make([]string, 0),
			WisdomPrinciples: make([]PersistentWisdomPrinciple, 0),
			RecentMemories:   make([]CognitiveMemory, 0),
			SemanticConcepts: make([]string, 0),
			Version:          1,
		},
	}

	// Try to load existing state
	if err := pcs.Load(); err != nil {
		fmt.Printf("No existing state found, starting fresh: %v\n", err)
	}

	return pcs
}

// Load loads the state from disk
func (pcs *PersistentCognitiveState) Load() error {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	data, err := os.ReadFile(pcs.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state CognitiveStateData
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	pcs.State = state
	return nil
}

// Save saves the state to disk
func (pcs *PersistentCognitiveState) Save() error {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(pcs.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(pcs.State, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(pcs.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	pcs.lastSave = time.Now()
	return nil
}

// AutoSave saves if enough time has passed since last save
func (pcs *PersistentCognitiveState) AutoSave() error {
	if time.Since(pcs.lastSave) > pcs.autoSaveInterval {
		return pcs.Save()
	}
	return nil
}

// RecordWake records a wake event
func (pcs *PersistentCognitiveState) RecordWake() {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	now := time.Now().Unix()
	
	// Calculate rest time if we have a last rest time
	if pcs.State.LastRestTime > 0 {
		restDuration := now - pcs.State.LastRestTime
		pcs.State.TotalRestTime += restDuration
	}
	
	pcs.State.LastWakeTime = now
	pcs.State.WakeRestCycles++
}

// RecordRest records a rest event
func (pcs *PersistentCognitiveState) RecordRest() {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	now := time.Now().Unix()
	
	// Calculate wake time
	if pcs.State.LastWakeTime > 0 {
		wakeDuration := now - pcs.State.LastWakeTime
		pcs.State.TotalWakeTime += wakeDuration
	}
	
	pcs.State.LastRestTime = now
}

// UpdateWisdomLevel updates the wisdom level
func (pcs *PersistentCognitiveState) UpdateWisdomLevel(delta float64) {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	
	pcs.State.WisdomLevel = minFloat(1.0, maxFloat(0.0, pcs.State.WisdomLevel+delta))
}

// AddWisdomPrinciple adds a new wisdom principle
func (pcs *PersistentCognitiveState) AddWisdomPrinciple(principle string, confidence float64) {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	wp := PersistentWisdomPrinciple{
		ID:          fmt.Sprintf("wp_%d", time.Now().UnixNano()),
		Principle:   principle,
		Confidence:  confidence,
		Applications: 0,
		CreatedAt:   time.Now().Unix(),
	}
	
	pcs.State.WisdomPrinciples = append(pcs.State.WisdomPrinciples, wp)
	
	// Keep only the most recent 100 principles
	if len(pcs.State.WisdomPrinciples) > 100 {
		pcs.State.WisdomPrinciples = pcs.State.WisdomPrinciples[len(pcs.State.WisdomPrinciples)-100:]
	}
}

// AddMemory adds a memory to the persistent state
func (pcs *PersistentCognitiveState) AddMemory(content string, importance float64, tags []string) {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()

	memory := CognitiveMemory{
		ID:         fmt.Sprintf("mem_%d", time.Now().UnixNano()),
		Content:    content,
		Importance: importance,
		Timestamp:  time.Now().Unix(),
		Tags:       tags,
	}
	
	pcs.State.RecentMemories = append(pcs.State.RecentMemories, memory)
	
	// Keep only the most recent 50 memories
	if len(pcs.State.RecentMemories) > 50 {
		pcs.State.RecentMemories = pcs.State.RecentMemories[len(pcs.State.RecentMemories)-50:]
	}
}

// IncrementThoughts increments the thought counter
func (pcs *PersistentCognitiveState) IncrementThoughts() {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	pcs.State.TotalThoughts++
}

// IncrementInteractions increments the interaction counter
func (pcs *PersistentCognitiveState) IncrementInteractions() {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	pcs.State.TotalInteractions++
}

// IncrementDreams increments the dream counter
func (pcs *PersistentCognitiveState) IncrementDreams() {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	pcs.State.TotalDreams++
}

// SetMood sets the current mood
func (pcs *PersistentCognitiveState) SetMood(mood string) {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	pcs.State.CurrentMood = mood
}

// SetFocus sets the current focus
func (pcs *PersistentCognitiveState) SetFocus(focus string) {
	pcs.mu.Lock()
	defer pcs.mu.Unlock()
	pcs.State.CurrentFocus = focus
}

// GetSummary returns a summary of the cognitive state
func (pcs *PersistentCognitiveState) GetSummary() map[string]interface{} {
	pcs.mu.RLock()
	defer pcs.mu.RUnlock()

	return map[string]interface{}{
		"identity":           pcs.State.Identity,
		"wisdom_level":       pcs.State.WisdomLevel,
		"awareness_level":    pcs.State.AwarenessLevel,
		"total_thoughts":     pcs.State.TotalThoughts,
		"total_interactions": pcs.State.TotalInteractions,
		"total_dreams":       pcs.State.TotalDreams,
		"wake_rest_cycles":   pcs.State.WakeRestCycles,
		"wisdom_principles":  len(pcs.State.WisdomPrinciples),
		"recent_memories":    len(pcs.State.RecentMemories),
		"current_mood":       pcs.State.CurrentMood,
		"current_focus":      pcs.State.CurrentFocus,
	}
}

// Helper functions
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
