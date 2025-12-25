// Package inference provides cognitive state persistence
// This implements save/restore for the echobeats inference engine
package inference

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// STATE CONFIGURATION
// =============================================================================

// StateConfig configures state persistence
type StateConfig struct {
	// Storage configuration
	StorageDir       string        // Directory for state files
	MaxSnapshots     int           // Maximum snapshots to keep
	AutoSaveInterval time.Duration // Auto-save interval (0 = disabled)
	
	// Compression
	EnableCompression bool   // Enable gzip compression
	CompressionLevel  int    // Compression level (1-9)
	
	// Checkpointing
	EnableCheckpoints bool   // Enable periodic checkpoints
	CheckpointInterval time.Duration // Checkpoint interval
	
	// Recovery
	RecoveryMode      string // "latest", "checkpoint", "specific"
	RecoveryTimestamp int64  // Specific timestamp for recovery
}

// DefaultStateConfig returns default state configuration
func DefaultStateConfig() StateConfig {
	return StateConfig{
		StorageDir:         "/tmp/echo_state",
		MaxSnapshots:       10,
		AutoSaveInterval:   0, // Disabled by default
		EnableCompression:  true,
		CompressionLevel:   6,
		EnableCheckpoints:  true,
		CheckpointInterval: 5 * time.Minute,
		RecoveryMode:       "latest",
	}
}

// =============================================================================
// COGNITIVE STATE
// =============================================================================

// CognitiveState represents the complete state of the cognitive system
type CognitiveState struct {
	// Metadata
	Version     int       `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	Checksum    string    `json:"checksum"`
	Description string    `json:"description"`
	
	// Engine state
	EngineState EchobeatsState `json:"engine_state"`
	
	// Stream states
	StreamStates [3]StreamState `json:"stream_states"`
	
	// KV cache state
	KVCacheState KVCacheSnapshot `json:"kv_cache_state"`
	
	// Cognitive loop state
	CognitiveLoopState CognitiveLoopSnapshot `json:"cognitive_loop_state"`
	
	// Memory pool state
	MemoryPoolState MemoryPoolSnapshot `json:"memory_pool_state"`
	
	// Opponent process state (from N+19)
	OpponentState OpponentSnapshot `json:"opponent_state"`
	
	// Metrics
	Metrics StateMetrics `json:"metrics"`
}

// StreamState represents the state of a single stream
type StreamState struct {
	StreamID      StreamID  `json:"stream_id"`
	CurrentStep   int       `json:"current_step"`
	TokensGenerated int64   `json:"tokens_generated"`
	LastInference time.Time `json:"last_inference"`
	
	// Context state
	ContextTokens []int32   `json:"context_tokens"`
	ContextLength int       `json:"context_length"`
	
	// Sampling state
	SamplerState  []byte    `json:"sampler_state"`
	
	// Stream-specific metadata
	Metadata      map[string]string `json:"metadata"`
}

// KVCacheSnapshot represents KV cache state
type KVCacheSnapshot struct {
	NumSlots    int              `json:"num_slots"`
	UsedSlots   int              `json:"used_slots"`
	TotalTokens int              `json:"total_tokens"`
	SlotStates  []KVSlotSnapshot `json:"slot_states"`
}

// KVSlotSnapshot represents a single KV cache slot
type KVSlotSnapshot struct {
	SlotID     int    `json:"slot_id"`
	SequenceID string `json:"sequence_id"`
	TokenCount int    `json:"token_count"`
	InUse      bool   `json:"in_use"`
	// KV data would be stored separately for large caches
	KVDataPath string `json:"kv_data_path,omitempty"`
}

// CognitiveLoopSnapshot represents cognitive loop state
type CognitiveLoopSnapshot struct {
	CurrentStep   int32  `json:"current_step"`
	CycleCount    uint64 `json:"cycle_count"`
	Phase         string `json:"phase"`
	
	// Step history
	StepHistory   []StepRecord `json:"step_history"`
	
	// Inter-stream state
	StreamSync    [3]int64 `json:"stream_sync"`
}

// StepRecord records a cognitive step
type StepRecord struct {
	Step      int       `json:"step"`
	StepType  string    `json:"step_type"`
	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration_ms"`
	StreamID  StreamID  `json:"stream_id"`
}

// MemoryPoolSnapshot represents memory pool state
type MemoryPoolSnapshot struct {
	TotalAllocated int64 `json:"total_allocated"`
	CurrentUsage   int64 `json:"current_usage"`
	PeakUsage      int64 `json:"peak_usage"`
	ArenaCount     int   `json:"arena_count"`
}

// OpponentSnapshot represents opponent process state
type OpponentSnapshot struct {
	OrdoInfluence float64            `json:"ordo_influence"`
	ChaoInfluence float64            `json:"chao_influence"`
	Balance       float64            `json:"balance"`
	WisdomScore   float64            `json:"wisdom_score"`
	PatternStates map[string]float64 `json:"pattern_states"`
}

// StateMetrics tracks state metrics
type StateMetrics struct {
	TotalInferences   uint64  `json:"total_inferences"`
	TotalTokens       uint64  `json:"total_tokens"`
	TotalCycles       uint64  `json:"total_cycles"`
	AverageLatencyMs  float64 `json:"average_latency_ms"`
	UptimeSeconds     int64   `json:"uptime_seconds"`
}

// =============================================================================
// STATE MANAGER
// =============================================================================

// StateManager manages cognitive state persistence
type StateManager struct {
	config StateConfig
	
	// Current state
	currentState *CognitiveState
	stateMu      sync.RWMutex
	
	// Snapshot management
	snapshots    []SnapshotInfo
	snapshotsMu  sync.RWMutex
	
	// Auto-save
	autoSaveStop chan struct{}
	autoSaveWg   sync.WaitGroup
	
	// State
	initialized atomic.Bool
	startTime   time.Time
}

// SnapshotInfo contains metadata about a snapshot
type SnapshotInfo struct {
	Path        string    `json:"path"`
	Timestamp   time.Time `json:"timestamp"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	Description string    `json:"description"`
	IsCheckpoint bool     `json:"is_checkpoint"`
}

// NewStateManager creates a new state manager
func NewStateManager(config StateConfig) *StateManager {
	return &StateManager{
		config:       config,
		snapshots:    make([]SnapshotInfo, 0),
		autoSaveStop: make(chan struct{}),
		startTime:    time.Now(),
	}
}

// Initialize sets up the state manager
func (sm *StateManager) Initialize() error {
	if sm.initialized.Load() {
		return errors.New("state manager already initialized")
	}
	
	// Create storage directory
	if err := os.MkdirAll(sm.config.StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	// Load existing snapshots
	if err := sm.loadSnapshotIndex(); err != nil {
		// Not fatal - might be first run
		sm.snapshots = make([]SnapshotInfo, 0)
	}
	
	// Initialize current state
	sm.currentState = &CognitiveState{
		Version:   1,
		Timestamp: time.Now(),
	}
	
	// Start auto-save if configured
	if sm.config.AutoSaveInterval > 0 {
		sm.autoSaveWg.Add(1)
		go sm.autoSaveLoop()
	}
	
	sm.initialized.Store(true)
	return nil
}

// SaveState saves the current cognitive state
func (sm *StateManager) SaveState(engine *EchobeatsEngine, description string) (*SnapshotInfo, error) {
	if !sm.initialized.Load() {
		return nil, errors.New("state manager not initialized")
	}
	
	sm.stateMu.Lock()
	defer sm.stateMu.Unlock()
	
	// Build state snapshot
	state := sm.buildState(engine, description)
	
	// Generate filename
	timestamp := time.Now()
	filename := fmt.Sprintf("state_%d.bin", timestamp.UnixNano())
	if sm.config.EnableCompression {
		filename += ".gz"
	}
	path := filepath.Join(sm.config.StorageDir, filename)
	
	// Serialize state
	data, err := sm.serializeState(state)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize state: %w", err)
	}
	
	// Compress if enabled
	if sm.config.EnableCompression {
		data, err = sm.compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress state: %w", err)
		}
	}
	
	// Calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	
	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write state file: %w", err)
	}
	
	// Create snapshot info
	info := SnapshotInfo{
		Path:        path,
		Timestamp:   timestamp,
		Size:        int64(len(data)),
		Checksum:    checksum,
		Description: description,
	}
	
	// Add to snapshots
	sm.snapshotsMu.Lock()
	sm.snapshots = append(sm.snapshots, info)
	sm.pruneSnapshots()
	sm.saveSnapshotIndex()
	sm.snapshotsMu.Unlock()
	
	return &info, nil
}

// LoadState loads a cognitive state
func (sm *StateManager) LoadState(path string) (*CognitiveState, error) {
	if !sm.initialized.Load() {
		return nil, errors.New("state manager not initialized")
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	
	// Decompress if needed
	if filepath.Ext(path) == ".gz" || (len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b) {
		data, err = sm.decompress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress state: %w", err)
		}
	}
	
	// Deserialize
	state, err := sm.deserializeState(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize state: %w", err)
	}
	
	return state, nil
}

// RestoreState restores state to an engine
func (sm *StateManager) RestoreState(engine *EchobeatsEngine, state *CognitiveState) error {
	if !sm.initialized.Load() {
		return errors.New("state manager not initialized")
	}
	
	// Restore engine state
	atomic.StoreInt32(&engine.currentStep, state.EngineState.CurrentStep)
	atomic.StoreUint64(&engine.cycleCount, state.EngineState.CycleCount)
	
	// Restore metrics
	engine.metrics.mu.Lock()
	engine.metrics.TotalInferences = state.Metrics.TotalInferences
	engine.metrics.TotalTokens = state.Metrics.TotalTokens
	engine.metrics.mu.Unlock()
	
	// Note: Full restoration would also restore:
	// - KV cache contents
	// - Stream contexts
	// - Sampler states
	// This requires the actual llama.cpp state restoration APIs
	
	return nil
}

// GetLatestSnapshot returns the most recent snapshot
func (sm *StateManager) GetLatestSnapshot() (*SnapshotInfo, error) {
	sm.snapshotsMu.RLock()
	defer sm.snapshotsMu.RUnlock()
	
	if len(sm.snapshots) == 0 {
		return nil, errors.New("no snapshots available")
	}
	
	latest := sm.snapshots[len(sm.snapshots)-1]
	return &latest, nil
}

// GetLatestCheckpoint returns the most recent checkpoint
func (sm *StateManager) GetLatestCheckpoint() (*SnapshotInfo, error) {
	sm.snapshotsMu.RLock()
	defer sm.snapshotsMu.RUnlock()
	
	for i := len(sm.snapshots) - 1; i >= 0; i-- {
		if sm.snapshots[i].IsCheckpoint {
			return &sm.snapshots[i], nil
		}
	}
	
	return nil, errors.New("no checkpoints available")
}

// ListSnapshots returns all available snapshots
func (sm *StateManager) ListSnapshots() []SnapshotInfo {
	sm.snapshotsMu.RLock()
	defer sm.snapshotsMu.RUnlock()
	
	result := make([]SnapshotInfo, len(sm.snapshots))
	copy(result, sm.snapshots)
	return result
}

// CreateCheckpoint creates a checkpoint snapshot
func (sm *StateManager) CreateCheckpoint(engine *EchobeatsEngine, description string) (*SnapshotInfo, error) {
	info, err := sm.SaveState(engine, "checkpoint: "+description)
	if err != nil {
		return nil, err
	}
	
	// Mark as checkpoint
	sm.snapshotsMu.Lock()
	for i := range sm.snapshots {
		if sm.snapshots[i].Path == info.Path {
			sm.snapshots[i].IsCheckpoint = true
			break
		}
	}
	sm.saveSnapshotIndex()
	sm.snapshotsMu.Unlock()
	
	info.IsCheckpoint = true
	return info, nil
}

// DeleteSnapshot deletes a snapshot
func (sm *StateManager) DeleteSnapshot(path string) error {
	sm.snapshotsMu.Lock()
	defer sm.snapshotsMu.Unlock()
	
	// Find and remove from list
	for i, snap := range sm.snapshots {
		if snap.Path == path {
			sm.snapshots = append(sm.snapshots[:i], sm.snapshots[i+1:]...)
			break
		}
	}
	
	// Delete file
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	sm.saveSnapshotIndex()
	return nil
}

// Close stops the state manager
func (sm *StateManager) Close() error {
	if !sm.initialized.Swap(false) {
		return nil
	}
	
	// Stop auto-save
	close(sm.autoSaveStop)
	sm.autoSaveWg.Wait()
	
	return nil
}

// =============================================================================
// INTERNAL METHODS
// =============================================================================

func (sm *StateManager) buildState(engine *EchobeatsEngine, description string) *CognitiveState {
	state := &CognitiveState{
		Version:     1,
		Timestamp:   time.Now(),
		Description: description,
	}
	
	// Engine state
	state.EngineState = engine.GetState()
	
	// Stream states
	for i := 0; i < 3; i++ {
		state.StreamStates[i] = StreamState{
			StreamID:    StreamID(i),
			CurrentStep: int(atomic.LoadInt32(&engine.currentStep)),
			Metadata:    make(map[string]string),
		}
	}
	
	// Cognitive loop state
	state.CognitiveLoopState = CognitiveLoopSnapshot{
		CurrentStep: atomic.LoadInt32(&engine.currentStep),
		CycleCount:  atomic.LoadUint64(&engine.cycleCount),
	}
	
	// Metrics
	metrics := engine.GetMetrics()
	state.Metrics = StateMetrics{
		TotalInferences: metrics.TotalInferences,
		TotalTokens:     metrics.TotalTokens,
		UptimeSeconds:   int64(time.Since(sm.startTime).Seconds()),
	}
	
	// Calculate checksum
	data, _ := json.Marshal(state)
	state.Checksum = fmt.Sprintf("%x", sha256.Sum256(data))
	
	return state
}

func (sm *StateManager) serializeState(state *CognitiveState) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(state); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sm *StateManager) deserializeState(data []byte) (*CognitiveState, error) {
	var state CognitiveState
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (sm *StateManager) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, sm.config.CompressionLevel)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (sm *StateManager) decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (sm *StateManager) autoSaveLoop() {
	defer sm.autoSaveWg.Done()
	
	ticker := time.NewTicker(sm.config.AutoSaveInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.autoSaveStop:
			return
		case <-ticker.C:
			// Auto-save would need access to the engine
			// This is a placeholder for the actual implementation
		}
	}
}

func (sm *StateManager) pruneSnapshots() {
	if len(sm.snapshots) <= sm.config.MaxSnapshots {
		return
	}
	
	// Keep checkpoints and most recent snapshots
	var keep []SnapshotInfo
	var regular []SnapshotInfo
	
	for _, snap := range sm.snapshots {
		if snap.IsCheckpoint {
			keep = append(keep, snap)
		} else {
			regular = append(regular, snap)
		}
	}
	
	// Keep most recent regular snapshots
	maxRegular := sm.config.MaxSnapshots - len(keep)
	if maxRegular < 0 {
		maxRegular = 0
	}
	
	if len(regular) > maxRegular {
		// Delete old snapshots
		for i := 0; i < len(regular)-maxRegular; i++ {
			os.Remove(regular[i].Path)
		}
		regular = regular[len(regular)-maxRegular:]
	}
	
	sm.snapshots = append(keep, regular...)
}

func (sm *StateManager) loadSnapshotIndex() error {
	path := filepath.Join(sm.config.StorageDir, "snapshots.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &sm.snapshots)
}

func (sm *StateManager) saveSnapshotIndex() error {
	path := filepath.Join(sm.config.StorageDir, "snapshots.json")
	data, err := json.MarshalIndent(sm.snapshots, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// =============================================================================
// KV CACHE SERIALIZATION
// =============================================================================

// KVCacheSerializer handles KV cache serialization
type KVCacheSerializer struct {
	storageDir string
}

// NewKVCacheSerializer creates a KV cache serializer
func NewKVCacheSerializer(storageDir string) *KVCacheSerializer {
	return &KVCacheSerializer{storageDir: storageDir}
}

// SaveKVCache saves KV cache to disk
func (s *KVCacheSerializer) SaveKVCache(slotID int, data []byte) (string, error) {
	filename := fmt.Sprintf("kv_slot_%d_%d.bin", slotID, time.Now().UnixNano())
	path := filepath.Join(s.storageDir, filename)
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	
	return path, nil
}

// LoadKVCache loads KV cache from disk
func (s *KVCacheSerializer) LoadKVCache(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// =============================================================================
// INCREMENTAL STATE
// =============================================================================

// IncrementalState represents incremental state updates
type IncrementalState struct {
	BaseChecksum string                 `json:"base_checksum"`
	Timestamp    time.Time              `json:"timestamp"`
	Delta        map[string]interface{} `json:"delta"`
}

// IncrementalStateManager manages incremental state updates
type IncrementalStateManager struct {
	baseState *CognitiveState
	deltas    []IncrementalState
	mu        sync.RWMutex
}

// NewIncrementalStateManager creates an incremental state manager
func NewIncrementalStateManager(baseState *CognitiveState) *IncrementalStateManager {
	return &IncrementalStateManager{
		baseState: baseState,
		deltas:    make([]IncrementalState, 0),
	}
}

// RecordDelta records a state delta
func (ism *IncrementalStateManager) RecordDelta(key string, value interface{}) {
	ism.mu.Lock()
	defer ism.mu.Unlock()
	
	delta := IncrementalState{
		BaseChecksum: ism.baseState.Checksum,
		Timestamp:    time.Now(),
		Delta:        map[string]interface{}{key: value},
	}
	ism.deltas = append(ism.deltas, delta)
}

// ApplyDeltas applies all deltas to get current state
func (ism *IncrementalStateManager) ApplyDeltas() *CognitiveState {
	ism.mu.RLock()
	defer ism.mu.RUnlock()
	
	// Deep copy base state
	data, _ := json.Marshal(ism.baseState)
	var state CognitiveState
	json.Unmarshal(data, &state)
	
	// Apply deltas (simplified - real implementation would be more sophisticated)
	for _, delta := range ism.deltas {
		for key, value := range delta.Delta {
			// Apply delta based on key
			switch key {
			case "current_step":
				if v, ok := value.(float64); ok {
					state.CognitiveLoopState.CurrentStep = int32(v)
				}
			case "cycle_count":
				if v, ok := value.(float64); ok {
					state.CognitiveLoopState.CycleCount = uint64(v)
				}
			}
		}
	}
	
	state.Timestamp = time.Now()
	return &state
}

// Compact compacts deltas into a new base state
func (ism *IncrementalStateManager) Compact() {
	ism.mu.Lock()
	defer ism.mu.Unlock()
	
	// Apply all deltas to create new base (inline to avoid deadlock)
	data, _ := json.Marshal(ism.baseState)
	var newBase CognitiveState
	json.Unmarshal(data, &newBase)
	
	for _, delta := range ism.deltas {
		for key, value := range delta.Delta {
			switch key {
			case "current_step":
				if v, ok := value.(float64); ok {
					newBase.CognitiveLoopState.CurrentStep = int32(v)
				}
			case "cycle_count":
				if v, ok := value.(float64); ok {
					newBase.CognitiveLoopState.CycleCount = uint64(v)
				}
			}
		}
	}
	newBase.Timestamp = time.Now()
	
	// Update checksum
	data2, _ := json.Marshal(newBase)
	newBase.Checksum = fmt.Sprintf("%x", sha256.Sum256(data2))
	
	ism.baseState = &newBase
	ism.deltas = make([]IncrementalState, 0)
}

// =============================================================================
// BINARY STATE FORMAT
// =============================================================================

// BinaryStateHeader is the header for binary state files
type BinaryStateHeader struct {
	Magic       [4]byte // "ECHO"
	Version     uint32
	Flags       uint32
	Timestamp   int64
	ChecksumLen uint32
	DataLen     uint64
}

// WriteBinaryState writes state in binary format
func WriteBinaryState(w io.Writer, state *CognitiveState) error {
	// Serialize state to JSON
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	
	// Calculate checksum
	checksum := sha256.Sum256(data)
	
	// Write header
	header := BinaryStateHeader{
		Magic:       [4]byte{'E', 'C', 'H', 'O'},
		Version:     1,
		Flags:       0,
		Timestamp:   state.Timestamp.UnixNano(),
		ChecksumLen: 32,
		DataLen:     uint64(len(data)),
	}
	
	if err := binary.Write(w, binary.LittleEndian, header); err != nil {
		return err
	}
	
	// Write checksum
	if _, err := w.Write(checksum[:]); err != nil {
		return err
	}
	
	// Write data
	if _, err := w.Write(data); err != nil {
		return err
	}
	
	return nil
}

// ReadBinaryState reads state from binary format
func ReadBinaryState(r io.Reader) (*CognitiveState, error) {
	// Read header
	var header BinaryStateHeader
	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	
	// Verify magic
	if header.Magic != [4]byte{'E', 'C', 'H', 'O'} {
		return nil, errors.New("invalid state file magic")
	}
	
	// Read checksum
	checksum := make([]byte, header.ChecksumLen)
	if _, err := io.ReadFull(r, checksum); err != nil {
		return nil, err
	}
	
	// Read data
	data := make([]byte, header.DataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	
	// Verify checksum
	computed := sha256.Sum256(data)
	if !bytes.Equal(checksum, computed[:]) {
		return nil, errors.New("checksum mismatch")
	}
	
	// Deserialize
	var state CognitiveState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	
	return &state, nil
}
