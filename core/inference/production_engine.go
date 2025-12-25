// Package inference provides the production-ready llama engine with full CGO integration
// This is the main entry point for the echobeats inference system
package inference

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// PRODUCTION ENGINE CONFIGURATION
// =============================================================================

// ProductionConfig is the complete configuration for the production engine
type ProductionConfig struct {
	// Model configuration
	ModelPath     string // Path to main model
	DraftModelPath string // Path to draft model (for speculative decoding)
	
	// Engine configuration
	Engine        EngineConfig
	Batch         BatchConfig
	Speculative   SpeculativeConfig
	State         StateConfig
	Pool          PoolConfig
	
	// Runtime configuration
	NumStreams    int           // Number of concurrent streams (default: 3)
	EnableGPU     bool          // Enable GPU acceleration
	GPUDevices    []int         // GPU device indices to use
	NumThreads    int           // Number of CPU threads per stream
	
	// Feature flags
	EnableStreaming        bool // Enable token streaming
	EnableBatching         bool // Enable continuous batching
	EnableSpeculative      bool // Enable speculative decoding
	EnableStatePersistence bool // Enable state persistence
	EnableMemoryPool       bool // Enable memory pooling
	
	// Logging
	LogLevel      string // "debug", "info", "warn", "error"
	MetricsPort   int    // Port for metrics endpoint (0 = disabled)
}

// DefaultProductionConfig returns default production configuration
func DefaultProductionConfig() ProductionConfig {
	return ProductionConfig{
		Engine:        DefaultEngineConfig(),
		Batch:         DefaultBatchConfig(),
		Speculative:   DefaultSpeculativeConfig(),
		State:         DefaultStateConfig(),
		Pool:          DefaultPoolConfig(),
		NumStreams:    3,
		EnableGPU:     true,
		GPUDevices:    []int{0},
		NumThreads:    runtime.NumCPU() / 3,
		EnableStreaming:        true,
		EnableBatching:         true,
		EnableSpeculative:      false, // Disabled by default
		EnableStatePersistence: true,
		EnableMemoryPool:       true,
		LogLevel:      "info",
		MetricsPort:   0,
	}
}

// =============================================================================
// PRODUCTION ENGINE
// =============================================================================

// ProductionEngine is the complete production-ready inference engine
type ProductionEngine struct {
	config ProductionConfig
	
	// Core components
	echobeatsEngine *EchobeatsEngine
	streamingEngine *StreamingEchobeatsEngine
	batchedEngine   *BatchedEchobeatsEngine
	specEngine      *SpeculativeEchobeatsEngine
	
	// Support systems
	stateManager    *StateManager
	memoryPool      *MemoryPool
	streamAllocator *StreamAllocator
	
	// State
	initialized atomic.Bool
	running     atomic.Bool
	startTime   time.Time
	
	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup
	
	// Shutdown
	shutdownChan chan struct{}
	
	// Metrics
	metrics *ProductionMetrics
}

// ProductionMetrics tracks production engine metrics
type ProductionMetrics struct {
	mu sync.RWMutex
	
	// Request metrics
	TotalRequests     uint64
	SuccessfulRequests uint64
	FailedRequests    uint64
	
	// Token metrics
	TotalTokensGenerated uint64
	TotalTokensPrompt    uint64
	
	// Latency metrics
	TotalLatencyMs    uint64
	MinLatencyMs      uint64
	MaxLatencyMs      uint64
	
	// Throughput
	TokensPerSecond   float64
	RequestsPerSecond float64
	
	// Resource metrics
	MemoryUsedBytes   uint64
	GPUMemoryUsedBytes uint64
	
	// Stream metrics
	StreamMetrics [3]StreamMetrics
}

// StreamMetrics tracks per-stream metrics
type StreamMetrics struct {
	Requests       uint64
	Tokens         uint64
	LatencyMs      uint64
	AcceptanceRate float64 // For speculative decoding
}

// NewProductionEngine creates a new production engine
func NewProductionEngine(config ProductionConfig) *ProductionEngine {
	return &ProductionEngine{
		config:       config,
		shutdownChan: make(chan struct{}),
		metrics:      &ProductionMetrics{MinLatencyMs: ^uint64(0)},
	}
}

// Initialize sets up all engine components
func (pe *ProductionEngine) Initialize(ctx context.Context) error {
	if pe.initialized.Load() {
		return errors.New("engine already initialized")
	}
	
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	pe.startTime = time.Now()
	
	// Validate model path
	if pe.config.ModelPath == "" {
		return errors.New("model path is required")
	}
	if _, err := os.Stat(pe.config.ModelPath); os.IsNotExist(err) {
		return fmt.Errorf("model file not found: %s", pe.config.ModelPath)
	}
	
	// Initialize memory pool
	if pe.config.EnableMemoryPool {
		pe.memoryPool = NewMemoryPool(pe.config.Pool)
		pe.streamAllocator = NewStreamAllocator(pe.config.Pool)
	}
	
	// Initialize state manager
	if pe.config.EnableStatePersistence {
		pe.stateManager = NewStateManager(pe.config.State)
		if err := pe.stateManager.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize state manager: %w", err)
		}
	}
	
	// Initialize the appropriate engine based on configuration
	if pe.config.EnableSpeculative && pe.config.DraftModelPath != "" {
		pe.specEngine = NewSpeculativeEchobeatsEngine(pe.config.Batch, pe.config.Speculative)
		if err := pe.specEngine.Initialize(pe.config.ModelPath, pe.config.Engine); err != nil {
			return fmt.Errorf("failed to initialize speculative engine: %w", err)
		}
		if err := pe.specEngine.InitializeSpeculative(pe.config.ModelPath, pe.config.DraftModelPath, pe.config.Engine); err != nil {
			return fmt.Errorf("failed to initialize speculative decoding: %w", err)
		}
		pe.batchedEngine = pe.specEngine.BatchedEchobeatsEngine
		pe.streamingEngine = pe.batchedEngine.StreamingEchobeatsEngine
		pe.echobeatsEngine = pe.streamingEngine.EchobeatsEngine
	} else if pe.config.EnableBatching {
		pe.batchedEngine = NewBatchedEchobeatsEngine(pe.config.Batch)
		if err := pe.batchedEngine.Initialize(pe.config.ModelPath, pe.config.Engine); err != nil {
			return fmt.Errorf("failed to initialize batched engine: %w", err)
		}
		pe.streamingEngine = pe.batchedEngine.StreamingEchobeatsEngine
		pe.echobeatsEngine = pe.streamingEngine.EchobeatsEngine
	} else if pe.config.EnableStreaming {
		pe.streamingEngine = NewStreamingEchobeatsEngine()
		if err := pe.streamingEngine.Initialize(pe.config.ModelPath, pe.config.Engine); err != nil {
			return fmt.Errorf("failed to initialize streaming engine: %w", err)
		}
		pe.echobeatsEngine = pe.streamingEngine.EchobeatsEngine
	} else {
		pe.echobeatsEngine = NewEchobeatsEngine()
		if err := pe.echobeatsEngine.Initialize(pe.config.ModelPath, pe.config.Engine); err != nil {
			return fmt.Errorf("failed to initialize echobeats engine: %w", err)
		}
	}
	
	// Try to restore previous state
	if pe.config.EnableStatePersistence {
		if snapshot, err := pe.stateManager.GetLatestSnapshot(); err == nil {
			if state, err := pe.stateManager.LoadState(snapshot.Path); err == nil {
				pe.stateManager.RestoreState(pe.echobeatsEngine, state)
			}
		}
	}
	
	pe.initialized.Store(true)
	return nil
}

// Start starts the production engine
func (pe *ProductionEngine) Start(ctx context.Context) error {
	if !pe.initialized.Load() {
		return errors.New("engine not initialized")
	}
	if pe.running.Swap(true) {
		return errors.New("engine already running")
	}
	
	// Start batched engine if enabled
	if pe.config.EnableBatching && pe.batchedEngine != nil {
		if err := pe.batchedEngine.Start(ctx); err != nil {
			pe.running.Store(false)
			return fmt.Errorf("failed to start batched engine: %w", err)
		}
	}
	
	// Start background tasks
	pe.wg.Add(1)
	go pe.metricsLoop(ctx)
	
	if pe.config.EnableStatePersistence && pe.config.State.AutoSaveInterval > 0 {
		pe.wg.Add(1)
		go pe.autoSaveLoop(ctx)
	}
	
	return nil
}

// Stop stops the production engine
func (pe *ProductionEngine) Stop(ctx context.Context) error {
	if !pe.running.Swap(false) {
		return nil
	}
	
	close(pe.shutdownChan)
	
	// Wait for background tasks
	done := make(chan struct{})
	go func() {
		pe.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	
	// Save final state
	if pe.config.EnableStatePersistence && pe.stateManager != nil {
		pe.stateManager.CreateCheckpoint(pe.echobeatsEngine, "shutdown checkpoint")
	}
	
	// Stop batched engine
	if pe.batchedEngine != nil {
		pe.batchedEngine.Stop()
	}
	
	return nil
}

// Close releases all resources
func (pe *ProductionEngine) Close() error {
	if !pe.initialized.Swap(false) {
		return nil
	}
	
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	var errs []error
	
	// Close engines
	if pe.specEngine != nil {
		if err := pe.specEngine.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if pe.echobeatsEngine != nil {
		if err := pe.echobeatsEngine.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	// Close state manager
	if pe.stateManager != nil {
		if err := pe.stateManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	// Close memory pool
	if pe.memoryPool != nil {
		if err := pe.memoryPool.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if pe.streamAllocator != nil {
		if err := pe.streamAllocator.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}

// =============================================================================
// INFERENCE METHODS
// =============================================================================

// Infer performs inference using the best available method
func (pe *ProductionEngine) Infer(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error) {
	if !pe.running.Load() {
		return nil, errors.New("engine not running")
	}
	
	start := time.Now()
	
	var resp *InferenceResponse
	var err error
	
	// Use speculative decoding if available
	if pe.specEngine != nil {
		resp, err = pe.specEngine.SpeculativeInferOnStream(ctx, req.StreamID, req)
	} else {
		resp, err = pe.echobeatsEngine.InferStream(ctx, req.StreamID, req)
	}
	
	// Update metrics
	pe.updateMetrics(req.StreamID, resp, time.Since(start), err)
	
	return resp, err
}

// InferStreaming performs streaming inference
func (pe *ProductionEngine) InferStreaming(ctx context.Context, req *StreamingRequest) (*StreamingResponse, error) {
	if !pe.running.Load() {
		return nil, errors.New("engine not running")
	}
	
	if pe.streamingEngine == nil {
		return nil, errors.New("streaming not enabled")
	}
	
	return pe.streamingEngine.InferStreamOnStream(ctx, req.StreamID, req)
}

// InferConcurrent performs concurrent inference on all 3 streams
func (pe *ProductionEngine) InferConcurrent(ctx context.Context, requests [3]*InferenceRequest) ([3]*InferenceResponse, error) {
	if !pe.running.Load() {
		return [3]*InferenceResponse{}, errors.New("engine not running")
	}
	
	if pe.specEngine != nil {
		return pe.specEngine.SpeculativeInferConcurrent(ctx, requests)
	}
	
	return pe.echobeatsEngine.InferConcurrent(ctx, requests)
}

// ExecuteCognitiveStep executes a single cognitive step
func (pe *ProductionEngine) ExecuteCognitiveStep(ctx context.Context, step int, prompt string) ([3]*InferenceResponse, error) {
	if !pe.running.Load() {
		return [3]*InferenceResponse{}, errors.New("engine not running")
	}
	
	return pe.echobeatsEngine.ExecuteCognitiveStep(ctx, step, prompt)
}

// ExecuteFullCycle executes a complete 12-step cognitive cycle
func (pe *ProductionEngine) ExecuteFullCycle(ctx context.Context, initialPrompt string) ([][3]*InferenceResponse, error) {
	if !pe.running.Load() {
		return nil, errors.New("engine not running")
	}
	
	return pe.echobeatsEngine.ExecuteFullCycle(ctx, initialPrompt)
}

// =============================================================================
// STATE MANAGEMENT
// =============================================================================

// SaveState saves the current state
func (pe *ProductionEngine) SaveState(description string) (*SnapshotInfo, error) {
	if pe.stateManager == nil {
		return nil, errors.New("state persistence not enabled")
	}
	
	return pe.stateManager.SaveState(pe.echobeatsEngine, description)
}

// LoadState loads a specific state
func (pe *ProductionEngine) LoadState(path string) error {
	if pe.stateManager == nil {
		return errors.New("state persistence not enabled")
	}
	
	state, err := pe.stateManager.LoadState(path)
	if err != nil {
		return err
	}
	
	return pe.stateManager.RestoreState(pe.echobeatsEngine, state)
}

// CreateCheckpoint creates a checkpoint
func (pe *ProductionEngine) CreateCheckpoint(description string) (*SnapshotInfo, error) {
	if pe.stateManager == nil {
		return nil, errors.New("state persistence not enabled")
	}
	
	return pe.stateManager.CreateCheckpoint(pe.echobeatsEngine, description)
}

// ListSnapshots returns all available snapshots
func (pe *ProductionEngine) ListSnapshots() []SnapshotInfo {
	if pe.stateManager == nil {
		return nil
	}
	return pe.stateManager.ListSnapshots()
}

// =============================================================================
// METRICS
// =============================================================================

// GetMetrics returns current metrics
func (pe *ProductionEngine) GetMetrics() ProductionMetrics {
	pe.metrics.mu.RLock()
	defer pe.metrics.mu.RUnlock()
	return *pe.metrics
}

// GetEngineMetrics returns echobeats engine metrics
func (pe *ProductionEngine) GetEngineMetrics() EngineMetrics {
	if pe.echobeatsEngine == nil {
		return EngineMetrics{}
	}
	return pe.echobeatsEngine.GetMetrics()
}

// GetBatcherStats returns batcher statistics
func (pe *ProductionEngine) GetBatcherStats() *BatcherStats {
	if pe.batchedEngine == nil {
		return nil
	}
	stats := pe.batchedEngine.Stats()
	return &stats
}

// GetSpeculativeStats returns speculative decoding statistics
func (pe *ProductionEngine) GetSpeculativeStats() *[3]SpeculativeStats {
	if pe.specEngine == nil {
		return nil
	}
	stats := pe.specEngine.SpeculativeStats()
	return &stats
}

// GetMemoryStats returns memory pool statistics
func (pe *ProductionEngine) GetMemoryStats() *PoolStats {
	if pe.memoryPool == nil {
		return nil
	}
	stats := pe.memoryPool.Stats()
	return &stats
}

func (pe *ProductionEngine) updateMetrics(streamID StreamID, resp *InferenceResponse, latency time.Duration, err error) {
	pe.metrics.mu.Lock()
	defer pe.metrics.mu.Unlock()
	
	pe.metrics.TotalRequests++
	
	if err != nil {
		pe.metrics.FailedRequests++
		return
	}
	
	pe.metrics.SuccessfulRequests++
	
	if resp != nil {
		pe.metrics.TotalTokensGenerated += uint64(len(resp.Tokens))
	}
	
	latencyMs := uint64(latency.Milliseconds())
	pe.metrics.TotalLatencyMs += latencyMs
	
	if latencyMs < pe.metrics.MinLatencyMs {
		pe.metrics.MinLatencyMs = latencyMs
	}
	if latencyMs > pe.metrics.MaxLatencyMs {
		pe.metrics.MaxLatencyMs = latencyMs
	}
	
	// Update stream metrics
	pe.metrics.StreamMetrics[streamID].Requests++
	if resp != nil {
		pe.metrics.StreamMetrics[streamID].Tokens += uint64(len(resp.Tokens))
	}
	pe.metrics.StreamMetrics[streamID].LatencyMs += latencyMs
}

func (pe *ProductionEngine) metricsLoop(ctx context.Context) {
	defer pe.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	var lastRequests, lastTokens uint64
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pe.shutdownChan:
			return
		case <-ticker.C:
			pe.metrics.mu.Lock()
			
			// Calculate throughput
			requestDelta := pe.metrics.TotalRequests - lastRequests
			tokenDelta := pe.metrics.TotalTokensGenerated - lastTokens
			
			pe.metrics.RequestsPerSecond = float64(requestDelta)
			pe.metrics.TokensPerSecond = float64(tokenDelta)
			
			lastRequests = pe.metrics.TotalRequests
			lastTokens = pe.metrics.TotalTokensGenerated
			
			pe.metrics.mu.Unlock()
		}
	}
}

func (pe *ProductionEngine) autoSaveLoop(ctx context.Context) {
	defer pe.wg.Done()
	
	ticker := time.NewTicker(pe.config.State.AutoSaveInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pe.shutdownChan:
			return
		case <-ticker.C:
			pe.stateManager.SaveState(pe.echobeatsEngine, "auto-save")
		}
	}
}

// =============================================================================
// HEALTH CHECK
// =============================================================================

// HealthStatus represents engine health
type HealthStatus struct {
	Healthy     bool              `json:"healthy"`
	Initialized bool              `json:"initialized"`
	Running     bool              `json:"running"`
	Uptime      time.Duration     `json:"uptime"`
	Components  map[string]bool   `json:"components"`
	Errors      []string          `json:"errors,omitempty"`
}

// Health returns the engine health status
func (pe *ProductionEngine) Health() HealthStatus {
	status := HealthStatus{
		Initialized: pe.initialized.Load(),
		Running:     pe.running.Load(),
		Uptime:      time.Since(pe.startTime),
		Components:  make(map[string]bool),
	}
	
	status.Components["echobeats"] = pe.echobeatsEngine != nil
	status.Components["streaming"] = pe.streamingEngine != nil
	status.Components["batching"] = pe.batchedEngine != nil
	status.Components["speculative"] = pe.specEngine != nil
	status.Components["state_manager"] = pe.stateManager != nil
	status.Components["memory_pool"] = pe.memoryPool != nil
	
	status.Healthy = status.Initialized && status.Running
	
	return status
}

// =============================================================================
// FACTORY FUNCTIONS
// =============================================================================

// CreateProductionEngine creates a fully configured production engine
func CreateProductionEngine(modelPath string, opts ...ProductionOption) (*ProductionEngine, error) {
	config := DefaultProductionConfig()
	config.ModelPath = modelPath
	
	for _, opt := range opts {
		opt(&config)
	}
	
	return NewProductionEngine(config), nil
}

// ProductionOption is a functional option for configuring the production engine
type ProductionOption func(*ProductionConfig)

// WithDraftModel enables speculative decoding with a draft model
func WithDraftModel(path string) ProductionOption {
	return func(c *ProductionConfig) {
		c.DraftModelPath = path
		c.EnableSpeculative = true
	}
}

// WithGPU enables GPU acceleration
func WithGPU(devices ...int) ProductionOption {
	return func(c *ProductionConfig) {
		c.EnableGPU = true
		if len(devices) > 0 {
			c.GPUDevices = devices
		}
	}
}

// WithContextSize sets the context size
func WithContextSize(size uint32) ProductionOption {
	return func(c *ProductionConfig) {
		c.Engine.ContextSize = size
	}
}

// WithBatchSize sets the batch size
func WithBatchSize(size uint32) ProductionOption {
	return func(c *ProductionConfig) {
		c.Engine.BatchSize = size
	}
}

// WithStateDir sets the state storage directory
func WithStateDir(dir string) ProductionOption {
	return func(c *ProductionConfig) {
		c.State.StorageDir = dir
		c.EnableStatePersistence = true
	}
}

// WithAutoSave enables auto-save with the specified interval
func WithAutoSave(interval time.Duration) ProductionOption {
	return func(c *ProductionConfig) {
		c.State.AutoSaveInterval = interval
		c.EnableStatePersistence = true
	}
}

// WithoutBatching disables continuous batching
func WithoutBatching() ProductionOption {
	return func(c *ProductionConfig) {
		c.EnableBatching = false
	}
}

// WithoutStreaming disables token streaming
func WithoutStreaming() ProductionOption {
	return func(c *ProductionConfig) {
		c.EnableStreaming = false
	}
}

// =============================================================================
// QUICK START
// =============================================================================

// QuickStart creates and starts a production engine with minimal configuration
func QuickStart(ctx context.Context, modelPath string) (*ProductionEngine, error) {
	engine, err := CreateProductionEngine(modelPath)
	if err != nil {
		return nil, err
	}
	
	if err := engine.Initialize(ctx); err != nil {
		return nil, err
	}
	
	if err := engine.Start(ctx); err != nil {
		engine.Close()
		return nil, err
	}
	
	return engine, nil
}

// QuickStartWithState creates and starts a production engine with state persistence
func QuickStartWithState(ctx context.Context, modelPath, stateDir string) (*ProductionEngine, error) {
	// Ensure state directory exists
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}
	
	engine, err := CreateProductionEngine(modelPath, WithStateDir(stateDir))
	if err != nil {
		return nil, err
	}
	
	if err := engine.Initialize(ctx); err != nil {
		return nil, err
	}
	
	if err := engine.Start(ctx); err != nil {
		engine.Close()
		return nil, err
	}
	
	return engine, nil
}

// GetModelDir returns the default model directory
func GetModelDir() string {
	// Check environment variable
	if dir := os.Getenv("ECHO_MODEL_DIR"); dir != "" {
		return dir
	}
	
	// Check common locations
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".echo", "models"),
		"/opt/echo/models",
		"./models",
	}
	
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}
	
	return "./models"
}
