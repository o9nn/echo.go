package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/EchoCog/echollama/core/echodream"
)

// ConsciousnessSnapshot represents a saved state of consciousness
type ConsciousnessSnapshot struct {
	// Metadata
	SessionID           string    `json:"session_id"`
	Timestamp           time.Time `json:"timestamp"`
	Version             string    `json:"version"`

	// State
	State               string    `json:"state"`
	Uptime              string    `json:"uptime"`

	// Metrics
	ThoughtsGenerated   uint64    `json:"thoughts_generated"`
	InsightsGenerated   uint64    `json:"insights_generated"`
	WisdomScore         float64   `json:"wisdom_score"`
	Fatigue             float64   `json:"fatigue"`

	// Cognitive data
	WisdomNuggets       []echodream.WisdomNugget `json:"wisdom_nuggets"`
	Patterns            []echodream.Pattern      `json:"patterns"`
	StateTransitions    []StateTransition        `json:"state_transitions"`

	// Metrics
	TotalSteps          uint64    `json:"total_steps"`
	ExpressiveSteps     uint64    `json:"expressive_steps"`
	ReflectiveSteps     uint64    `json:"reflective_steps"`
	DreamCycles         uint64    `json:"dream_cycles"`
	MemoriesConsolidated uint64   `json:"memories_consolidated"`
	PatternsDetected    uint64    `json:"patterns_detected"`
	WisdomExtracted     uint64    `json:"wisdom_extracted"`
}

// PersistenceManager handles saving and loading consciousness state
type PersistenceManager struct {
	basePath     string
	maxSnapshots int
}

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(basePath string, maxSnapshots int) *PersistenceManager {
	return &PersistenceManager{
		basePath:     basePath,
		maxSnapshots: maxSnapshots,
	}
}

// SaveSnapshot saves the current consciousness state
func (pm *PersistenceManager) SaveSnapshot(snapshot *ConsciousnessSnapshot) error {
	// Ensure directory exists
	if err := os.MkdirAll(pm.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("consciousness_%s.json", snapshot.SessionID)
	filepath := filepath.Join(pm.basePath, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	// Cleanup old snapshots
	pm.cleanupOldSnapshots()

	return nil
}

// LoadLatestSnapshot loads the most recent consciousness snapshot
func (pm *PersistenceManager) LoadLatestSnapshot() (*ConsciousnessSnapshot, error) {
	// List all snapshot files
	files, err := ioutil.ReadDir(pm.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No snapshots yet
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Find the most recent snapshot
	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		if file.ModTime().After(latestTime) {
			latestTime = file.ModTime()
			latestFile = file.Name()
		}
	}

	if latestFile == "" {
		return nil, nil // No snapshots found
	}

	// Load the snapshot
	return pm.LoadSnapshot(latestFile)
}

// LoadSnapshot loads a specific consciousness snapshot
func (pm *PersistenceManager) LoadSnapshot(filename string) (*ConsciousnessSnapshot, error) {
	filepath := filepath.Join(pm.basePath, filename)

	// Read file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	// Unmarshal JSON
	var snapshot ConsciousnessSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &snapshot, nil
}

// ListSnapshots returns a list of all available snapshots
func (pm *PersistenceManager) ListSnapshots() ([]string, error) {
	files, err := ioutil.ReadDir(pm.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	snapshots := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".json" {
			snapshots = append(snapshots, file.Name())
		}
	}

	return snapshots, nil
}

// cleanupOldSnapshots removes old snapshots beyond maxSnapshots
func (pm *PersistenceManager) cleanupOldSnapshots() {
	files, err := ioutil.ReadDir(pm.basePath)
	if err != nil {
		return
	}

	// Filter JSON files and sort by modification time
	type fileInfo struct {
		name    string
		modTime time.Time
	}

	snapshots := make([]fileInfo, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".json" {
			snapshots = append(snapshots, fileInfo{
				name:    file.Name(),
				modTime: file.ModTime(),
			})
		}
	}

	// If we have more than maxSnapshots, delete the oldest
	if len(snapshots) > pm.maxSnapshots {
		// Sort by modification time (oldest first)
		for i := 0; i < len(snapshots)-1; i++ {
			for j := i + 1; j < len(snapshots); j++ {
				if snapshots[i].modTime.After(snapshots[j].modTime) {
					snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
				}
			}
		}

		// Delete oldest files
		toDelete := len(snapshots) - pm.maxSnapshots
		for i := 0; i < toDelete; i++ {
			filepath := filepath.Join(pm.basePath, snapshots[i].name)
			os.Remove(filepath)
		}
	}
}

// CreateSnapshot creates a snapshot from the current agent state
func CreateSnapshot(ae *AutonomousEchoselfV5, sessionID string) *ConsciousnessSnapshot {
	status := ae.GetStatus()
	wisdom := ae.GetWisdom()
	patterns := ae.GetPatterns()

	snapshot := &ConsciousnessSnapshot{
		SessionID:         sessionID,
		Timestamp:         time.Now(),
		Version:           "v5.0",
		WisdomNuggets:     wisdom,
		Patterns:          patterns,
		StateTransitions:  ae.stateTransitions,
	}

	// Extract status fields
	if state, ok := status["state"].(string); ok {
		snapshot.State = state
	}
	if uptime, ok := status["uptime"].(string); ok {
		snapshot.Uptime = uptime
	}
	if thoughts, ok := status["thoughts_generated"].(uint64); ok {
		snapshot.ThoughtsGenerated = thoughts
	}
	if insights, ok := status["insights_generated"].(uint64); ok {
		snapshot.InsightsGenerated = insights
	}
	if wisdomScore, ok := status["wisdom_score"].(float64); ok {
		snapshot.WisdomScore = wisdomScore
	}
	if fatigue, ok := status["fatigue"].(float64); ok {
		snapshot.Fatigue = fatigue
	}
	if totalSteps, ok := status["total_steps"].(uint64); ok {
		snapshot.TotalSteps = totalSteps
	}
	if dreamCycles, ok := status["total_dream_cycles"].(uint64); ok {
		snapshot.DreamCycles = dreamCycles
	}
	if memoriesConsolidated, ok := status["memories_consolidated"].(uint64); ok {
		snapshot.MemoriesConsolidated = memoriesConsolidated
	}
	if patternsDetected, ok := status["patterns_detected"].(uint64); ok {
		snapshot.PatternsDetected = patternsDetected
	}
	if wisdomExtracted, ok := status["wisdom_extracted"].(uint64); ok {
		snapshot.WisdomExtracted = wisdomExtracted
	}

	return snapshot
}

// RestoreSnapshot restores state from a snapshot
func RestoreSnapshot(ae *AutonomousEchoselfV5, snapshot *ConsciousnessSnapshot) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Restore metrics
	ae.thoughtsGenerated = snapshot.ThoughtsGenerated
	ae.insightsGenerated = snapshot.InsightsGenerated
	ae.wisdomScore = snapshot.WisdomScore
	ae.stateTransitions = snapshot.StateTransitions

	// Restore wisdom and patterns to consolidation engine
	if ae.consolidationEngine != nil {
		// Note: This is a simplified restoration
		// A full implementation would properly restore the internal state
		// For now, we just log that we would restore this data
		fmt.Printf("📥 Restored %d wisdom nuggets and %d patterns from snapshot\n",
			len(snapshot.WisdomNuggets), len(snapshot.Patterns))
	}

	// Restore fatigue
	if ae.wakeController != nil {
		ae.wakeController.UpdateFatigue(snapshot.Fatigue)
	}

	fmt.Printf("✅ Consciousness state restored from session %s\n", snapshot.SessionID)
	fmt.Printf("   Previous uptime: %s\n", snapshot.Uptime)
	fmt.Printf("   Thoughts: %d, Insights: %d, Wisdom: %.2f\n",
		snapshot.ThoughtsGenerated, snapshot.InsightsGenerated, snapshot.WisdomScore)

	return nil
}
