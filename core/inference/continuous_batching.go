// Package inference provides continuous batching for high-throughput inference
// This implements dynamic batch management for the echobeats inference engine
package inference

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// BATCH CONFIGURATION
// =============================================================================

// BatchConfig configures the continuous batching system
type BatchConfig struct {
	MaxBatchSize      int           // Maximum sequences in a batch
	MaxTokensPerBatch int           // Maximum total tokens per batch
	MaxWaitTime       time.Duration // Maximum time to wait for batch fill
	MinBatchSize      int           // Minimum sequences before processing
	PrefillBatchSize  int           // Batch size for prefill phase
	DecodeBatchSize   int           // Batch size for decode phase
	
	// Priority configuration
	EnablePriority    bool          // Enable priority-based scheduling
	StreamPriorities  [3]int        // Priority levels for each stream (higher = more priority)
	
	// Memory configuration
	MaxKVCacheTokens  int           // Maximum tokens in KV cache
	KVCacheEviction   string        // Eviction policy: "lru", "fifo", "priority"
}

// DefaultBatchConfig returns default batch configuration
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxBatchSize:      64,
		MaxTokensPerBatch: 8192,
		MaxWaitTime:       50 * time.Millisecond,
		MinBatchSize:      1,
		PrefillBatchSize:  32,
		DecodeBatchSize:   64,
		EnablePriority:    true,
		StreamPriorities:  [3]int{3, 2, 1}, // Alpha > Beta > Gamma
		MaxKVCacheTokens:  32768,
		KVCacheEviction:   "lru",
	}
}

// =============================================================================
// SEQUENCE
// =============================================================================

// SequenceState represents the state of a sequence
type SequenceState int

const (
	SequenceStatePending SequenceState = iota
	SequenceStatePrefill
	SequenceStateDecode
	SequenceStateComplete
	SequenceStateCancelled
)

// Sequence represents a single inference sequence
type Sequence struct {
	ID           string
	StreamID     StreamID
	Step         int
	Priority     int
	State        SequenceState
	
	// Input
	PromptTokens []int32
	MaxNewTokens int
	
	// Generation state
	GeneratedTokens []int32
	CurrentPosition int
	
	// KV cache
	KVCacheSlot int
	KVCacheLen  int
	
	// Timing
	SubmitTime   time.Time
	StartTime    time.Time
	EndTime      time.Time
	
	// Output
	Response     *StreamingResponse
	Error        error
	
	// Callbacks
	OnToken      func(*Token)
	OnComplete   func(*Sequence)
}

// TokenCount returns total tokens (prompt + generated)
func (s *Sequence) TokenCount() int {
	return len(s.PromptTokens) + len(s.GeneratedTokens)
}

// IsComplete returns whether generation is complete
func (s *Sequence) IsComplete() bool {
	return s.State == SequenceStateComplete || s.State == SequenceStateCancelled
}

// =============================================================================
// PRIORITY QUEUE
// =============================================================================

// SequenceQueue is a priority queue for sequences
type SequenceQueue struct {
	sequences []*Sequence
	mu        sync.RWMutex
}

func (pq *SequenceQueue) Len() int {
	return len(pq.sequences)
}

func (pq *SequenceQueue) Less(i, j int) bool {
	// Higher priority first, then earlier submit time
	if pq.sequences[i].Priority != pq.sequences[j].Priority {
		return pq.sequences[i].Priority > pq.sequences[j].Priority
	}
	return pq.sequences[i].SubmitTime.Before(pq.sequences[j].SubmitTime)
}

func (pq *SequenceQueue) Swap(i, j int) {
	pq.sequences[i], pq.sequences[j] = pq.sequences[j], pq.sequences[i]
}

func (pq *SequenceQueue) Push(x interface{}) {
	pq.sequences = append(pq.sequences, x.(*Sequence))
}

func (pq *SequenceQueue) Pop() interface{} {
	old := pq.sequences
	n := len(old)
	seq := old[n-1]
	pq.sequences = old[0 : n-1]
	return seq
}

// =============================================================================
// KV CACHE MANAGER
// =============================================================================

// KVCacheSlot represents a slot in the KV cache
type KVCacheSlot struct {
	ID          int
	SequenceID  string
	TokenCount  int
	LastAccess  time.Time
	InUse       bool
}

// KVCacheManager manages KV cache allocation
type KVCacheManager struct {
	slots       []*KVCacheSlot
	maxTokens   int
	usedTokens  int
	policy      string
	mu          sync.RWMutex
}

// NewKVCacheManager creates a new KV cache manager
func NewKVCacheManager(maxTokens int, numSlots int, policy string) *KVCacheManager {
	slots := make([]*KVCacheSlot, numSlots)
	for i := range slots {
		slots[i] = &KVCacheSlot{ID: i}
	}
	return &KVCacheManager{
		slots:     slots,
		maxTokens: maxTokens,
		policy:    policy,
	}
}

// Allocate allocates a KV cache slot for a sequence
func (kv *KVCacheManager) Allocate(seqID string, tokenCount int) (int, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	
	// Check if we have space
	if kv.usedTokens+tokenCount > kv.maxTokens {
		// Try to evict
		if err := kv.evictLocked(tokenCount); err != nil {
			return -1, err
		}
	}
	
	// Find free slot
	for _, slot := range kv.slots {
		if !slot.InUse {
			slot.InUse = true
			slot.SequenceID = seqID
			slot.TokenCount = tokenCount
			slot.LastAccess = time.Now()
			kv.usedTokens += tokenCount
			return slot.ID, nil
		}
	}
	
	return -1, errors.New("no free KV cache slots")
}

// Release releases a KV cache slot
func (kv *KVCacheManager) Release(slotID int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	
	if slotID >= 0 && slotID < len(kv.slots) {
		slot := kv.slots[slotID]
		kv.usedTokens -= slot.TokenCount
		slot.InUse = false
		slot.SequenceID = ""
		slot.TokenCount = 0
	}
}

// Update updates the token count for a slot
func (kv *KVCacheManager) Update(slotID int, tokenCount int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	
	if slotID >= 0 && slotID < len(kv.slots) {
		slot := kv.slots[slotID]
		kv.usedTokens += tokenCount - slot.TokenCount
		slot.TokenCount = tokenCount
		slot.LastAccess = time.Now()
	}
}

func (kv *KVCacheManager) evictLocked(needed int) error {
	// Sort slots by eviction policy
	var candidates []*KVCacheSlot
	for _, slot := range kv.slots {
		if slot.InUse {
			candidates = append(candidates, slot)
		}
	}
	
	if len(candidates) == 0 {
		return errors.New("no slots to evict")
	}
	
	// Sort by last access (LRU)
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].LastAccess.After(candidates[j].LastAccess) {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	
	// Evict until we have enough space
	freed := 0
	for _, slot := range candidates {
		if freed >= needed {
			break
		}
		freed += slot.TokenCount
		kv.usedTokens -= slot.TokenCount
		slot.InUse = false
		slot.SequenceID = ""
		slot.TokenCount = 0
	}
	
	if freed < needed {
		return fmt.Errorf("could not free enough space: needed %d, freed %d", needed, freed)
	}
	
	return nil
}

// Stats returns KV cache statistics
func (kv *KVCacheManager) Stats() (used, total, slots int) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	inUse := 0
	for _, slot := range kv.slots {
		if slot.InUse {
			inUse++
		}
	}
	
	return kv.usedTokens, kv.maxTokens, inUse
}

// =============================================================================
// BATCH
// =============================================================================

// Batch represents a batch of sequences for processing
type Batch struct {
	ID         int
	Sequences  []*Sequence
	Phase      string // "prefill" or "decode"
	TokenCount int
	CreateTime time.Time
}

// NewBatch creates a new batch
func NewBatch(id int, phase string) *Batch {
	return &Batch{
		ID:         id,
		Sequences:  make([]*Sequence, 0),
		Phase:      phase,
		CreateTime: time.Now(),
	}
}

// Add adds a sequence to the batch
func (b *Batch) Add(seq *Sequence) {
	b.Sequences = append(b.Sequences, seq)
	b.TokenCount += seq.TokenCount()
}

// Size returns the number of sequences in the batch
func (b *Batch) Size() int {
	return len(b.Sequences)
}

// =============================================================================
// CONTINUOUS BATCHER
// =============================================================================

// ContinuousBatcher manages continuous batching
type ContinuousBatcher struct {
	config BatchConfig
	
	// Queues
	prefillQueue *SequenceQueue
	decodeQueue  *SequenceQueue
	
	// KV cache
	kvCache *KVCacheManager
	
	// State
	running     atomic.Bool
	batchCount  int64
	seqCount    int64
	
	// Channels
	submitChan  chan *Sequence
	batchChan   chan *Batch
	doneChan    chan struct{}
	
	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup
	
	// Active sequences
	activeSeqs   map[string]*Sequence
	activeSeqsMu sync.RWMutex
}

// NewContinuousBatcher creates a new continuous batcher
func NewContinuousBatcher(config BatchConfig) *ContinuousBatcher {
	return &ContinuousBatcher{
		config:       config,
		prefillQueue: &SequenceQueue{sequences: make([]*Sequence, 0)},
		decodeQueue:  &SequenceQueue{sequences: make([]*Sequence, 0)},
		kvCache:      NewKVCacheManager(config.MaxKVCacheTokens, config.MaxBatchSize*2, config.KVCacheEviction),
		submitChan:   make(chan *Sequence, 1000),
		batchChan:    make(chan *Batch, 100),
		doneChan:     make(chan struct{}),
		activeSeqs:   make(map[string]*Sequence),
	}
}

// Start starts the continuous batcher
func (cb *ContinuousBatcher) Start(ctx context.Context) error {
	if cb.running.Swap(true) {
		return errors.New("batcher already running")
	}
	
	// Start batch formation goroutine
	cb.wg.Add(1)
	go cb.batchFormationLoop(ctx)
	
	return nil
}

// Stop stops the continuous batcher
func (cb *ContinuousBatcher) Stop() {
	if !cb.running.Swap(false) {
		return
	}
	
	close(cb.doneChan)
	cb.wg.Wait()
}

// Submit submits a sequence for processing
func (cb *ContinuousBatcher) Submit(seq *Sequence) error {
	if !cb.running.Load() {
		return errors.New("batcher not running")
	}
	
	seq.ID = fmt.Sprintf("seq-%d", atomic.AddInt64(&cb.seqCount, 1))
	seq.SubmitTime = time.Now()
	seq.State = SequenceStatePending
	
	// Set priority based on stream
	if cb.config.EnablePriority {
		seq.Priority = cb.config.StreamPriorities[seq.StreamID]
	}
	
	// Track active sequence
	cb.activeSeqsMu.Lock()
	cb.activeSeqs[seq.ID] = seq
	cb.activeSeqsMu.Unlock()
	
	select {
	case cb.submitChan <- seq:
		return nil
	case <-cb.doneChan:
		return errors.New("batcher stopped")
	}
}

// GetBatch returns the next batch to process
func (cb *ContinuousBatcher) GetBatch(ctx context.Context) (*Batch, error) {
	select {
	case batch := <-cb.batchChan:
		return batch, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-cb.doneChan:
		return nil, errors.New("batcher stopped")
	}
}

// Complete marks a sequence as complete
func (cb *ContinuousBatcher) Complete(seqID string, err error) {
	cb.activeSeqsMu.Lock()
	seq, ok := cb.activeSeqs[seqID]
	if ok {
		delete(cb.activeSeqs, seqID)
	}
	cb.activeSeqsMu.Unlock()
	
	if !ok {
		return
	}
	
	seq.EndTime = time.Now()
	if err != nil {
		seq.State = SequenceStateCancelled
		seq.Error = err
	} else {
		seq.State = SequenceStateComplete
	}
	
	// Release KV cache
	cb.kvCache.Release(seq.KVCacheSlot)
	
	// Call completion callback
	if seq.OnComplete != nil {
		seq.OnComplete(seq)
	}
}

// batchFormationLoop forms batches from queued sequences
func (cb *ContinuousBatcher) batchFormationLoop(ctx context.Context) {
	defer cb.wg.Done()
	
	ticker := time.NewTicker(cb.config.MaxWaitTime / 2)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-cb.doneChan:
			return
		case seq := <-cb.submitChan:
			// Add to prefill queue
			cb.prefillQueue.mu.Lock()
			heap.Push(cb.prefillQueue, seq)
			cb.prefillQueue.mu.Unlock()
			
			// Try to form a batch
			cb.tryFormBatch()
			
		case <-ticker.C:
			// Periodic batch formation
			cb.tryFormBatch()
		}
	}
}

// tryFormBatch attempts to form and dispatch a batch
func (cb *ContinuousBatcher) tryFormBatch() {
	// Try prefill batch first (higher priority)
	if batch := cb.formPrefillBatch(); batch != nil {
		select {
		case cb.batchChan <- batch:
			atomic.AddInt64(&cb.batchCount, 1)
		default:
			// Put sequences back in queue
			cb.prefillQueue.mu.Lock()
			for _, seq := range batch.Sequences {
				heap.Push(cb.prefillQueue, seq)
			}
			cb.prefillQueue.mu.Unlock()
		}
		return
	}
	
	// Try decode batch
	if batch := cb.formDecodeBatch(); batch != nil {
		select {
		case cb.batchChan <- batch:
			atomic.AddInt64(&cb.batchCount, 1)
		default:
			// Put sequences back in queue
			cb.decodeQueue.mu.Lock()
			for _, seq := range batch.Sequences {
				heap.Push(cb.decodeQueue, seq)
			}
			cb.decodeQueue.mu.Unlock()
		}
	}
}

// formPrefillBatch forms a batch for prefill phase
func (cb *ContinuousBatcher) formPrefillBatch() *Batch {
	cb.prefillQueue.mu.Lock()
	defer cb.prefillQueue.mu.Unlock()
	
	if cb.prefillQueue.Len() < cb.config.MinBatchSize {
		return nil
	}
	
	batch := NewBatch(int(atomic.LoadInt64(&cb.batchCount)), "prefill")
	totalTokens := 0
	
	for cb.prefillQueue.Len() > 0 && batch.Size() < cb.config.PrefillBatchSize {
		seq := heap.Pop(cb.prefillQueue).(*Sequence)
		seqTokens := len(seq.PromptTokens)
		
		// Check token limit
		if totalTokens+seqTokens > cb.config.MaxTokensPerBatch {
			heap.Push(cb.prefillQueue, seq)
			break
		}
		
		// Allocate KV cache
		slot, err := cb.kvCache.Allocate(seq.ID, seqTokens)
		if err != nil {
			heap.Push(cb.prefillQueue, seq)
			break
		}
		
		seq.KVCacheSlot = slot
		seq.KVCacheLen = seqTokens
		seq.State = SequenceStatePrefill
		seq.StartTime = time.Now()
		
		batch.Add(seq)
		totalTokens += seqTokens
	}
	
	if batch.Size() == 0 {
		return nil
	}
	
	return batch
}

// formDecodeBatch forms a batch for decode phase
func (cb *ContinuousBatcher) formDecodeBatch() *Batch {
	cb.decodeQueue.mu.Lock()
	defer cb.decodeQueue.mu.Unlock()
	
	if cb.decodeQueue.Len() == 0 {
		return nil
	}
	
	batch := NewBatch(int(atomic.LoadInt64(&cb.batchCount)), "decode")
	
	for cb.decodeQueue.Len() > 0 && batch.Size() < cb.config.DecodeBatchSize {
		seq := heap.Pop(cb.decodeQueue).(*Sequence)
		seq.State = SequenceStateDecode
		batch.Add(seq)
	}
	
	if batch.Size() == 0 {
		return nil
	}
	
	return batch
}

// MoveToDecodeQueue moves a sequence from prefill to decode queue
func (cb *ContinuousBatcher) MoveToDecodeQueue(seq *Sequence) {
	cb.decodeQueue.mu.Lock()
	heap.Push(cb.decodeQueue, seq)
	cb.decodeQueue.mu.Unlock()
}

// Stats returns batcher statistics
type BatcherStats struct {
	BatchCount      int64
	SequenceCount   int64
	PrefillQueueLen int
	DecodeQueueLen  int
	ActiveSeqs      int
	KVCacheUsed     int
	KVCacheTotal    int
	KVCacheSlots    int
}

func (cb *ContinuousBatcher) Stats() BatcherStats {
	cb.prefillQueue.mu.RLock()
	prefillLen := cb.prefillQueue.Len()
	cb.prefillQueue.mu.RUnlock()
	
	cb.decodeQueue.mu.RLock()
	decodeLen := cb.decodeQueue.Len()
	cb.decodeQueue.mu.RUnlock()
	
	cb.activeSeqsMu.RLock()
	activeLen := len(cb.activeSeqs)
	cb.activeSeqsMu.RUnlock()
	
	kvUsed, kvTotal, kvSlots := cb.kvCache.Stats()
	
	return BatcherStats{
		BatchCount:      atomic.LoadInt64(&cb.batchCount),
		SequenceCount:   atomic.LoadInt64(&cb.seqCount),
		PrefillQueueLen: prefillLen,
		DecodeQueueLen:  decodeLen,
		ActiveSeqs:      activeLen,
		KVCacheUsed:     kvUsed,
		KVCacheTotal:    kvTotal,
		KVCacheSlots:    kvSlots,
	}
}

// =============================================================================
// BATCHED ECHOBEATS ENGINE
// =============================================================================

// BatchedEchobeatsEngine adds continuous batching to EchobeatsEngine
type BatchedEchobeatsEngine struct {
	*StreamingEchobeatsEngine
	
	batcher *ContinuousBatcher
	config  BatchConfig
	
	// Processing
	processingWg sync.WaitGroup
}

// NewBatchedEchobeatsEngine creates a batched echobeats engine
func NewBatchedEchobeatsEngine(config BatchConfig) *BatchedEchobeatsEngine {
	return &BatchedEchobeatsEngine{
		StreamingEchobeatsEngine: NewStreamingEchobeatsEngine(),
		batcher:                  NewContinuousBatcher(config),
		config:                   config,
	}
}

// Start starts the batched engine
func (be *BatchedEchobeatsEngine) Start(ctx context.Context) error {
	if err := be.batcher.Start(ctx); err != nil {
		return err
	}
	
	// Start batch processing goroutine
	be.processingWg.Add(1)
	go be.processBatches(ctx)
	
	return nil
}

// Stop stops the batched engine
func (be *BatchedEchobeatsEngine) Stop() {
	be.batcher.Stop()
	be.processingWg.Wait()
}

// SubmitRequest submits an inference request
func (be *BatchedEchobeatsEngine) SubmitRequest(req *StreamingRequest) (*Sequence, error) {
	// Create sequence
	seq := &Sequence{
		StreamID:     req.StreamID,
		Step:         req.Step,
		PromptTokens: make([]int32, 0), // Would be tokenized in production
		MaxNewTokens: req.MaxTokens,
		Response:     NewStreamingResponse(req.StreamID, req.Step),
		OnToken:      req.OnToken,
	}
	
	// Submit to batcher
	if err := be.batcher.Submit(seq); err != nil {
		return nil, err
	}
	
	return seq, nil
}

// processBatches processes batches from the batcher
func (be *BatchedEchobeatsEngine) processBatches(ctx context.Context) {
	defer be.processingWg.Done()
	
	for {
		batch, err := be.batcher.GetBatch(ctx)
		if err != nil {
			return
		}
		
		// Process batch
		be.processBatch(ctx, batch)
	}
}

// processBatch processes a single batch
func (be *BatchedEchobeatsEngine) processBatch(ctx context.Context, batch *Batch) {
	switch batch.Phase {
	case "prefill":
		be.processPrefillBatch(ctx, batch)
	case "decode":
		be.processDecodeBatch(ctx, batch)
	}
}

// processPrefillBatch processes a prefill batch
func (be *BatchedEchobeatsEngine) processPrefillBatch(ctx context.Context, batch *Batch) {
	// In production, this would:
	// 1. Build a batch tensor from all prompt tokens
	// 2. Run forward pass through the model
	// 3. Store KV cache for each sequence
	// 4. Move sequences to decode queue
	
	for _, seq := range batch.Sequences {
		// Simulate prefill
		seq.CurrentPosition = len(seq.PromptTokens)
		
		// Move to decode queue
		be.batcher.MoveToDecodeQueue(seq)
	}
}

// processDecodeBatch processes a decode batch
func (be *BatchedEchobeatsEngine) processDecodeBatch(ctx context.Context, batch *Batch) {
	// In production, this would:
	// 1. Build a batch tensor with one token per sequence
	// 2. Run forward pass through the model
	// 3. Sample next token for each sequence
	// 4. Update KV cache
	// 5. Check for completion
	
	for _, seq := range batch.Sequences {
		// Simulate token generation
		newToken := int32(len(seq.GeneratedTokens) + 1000)
		seq.GeneratedTokens = append(seq.GeneratedTokens, newToken)
		seq.CurrentPosition++
		
		// Update KV cache
		be.batcher.kvCache.Update(seq.KVCacheSlot, seq.TokenCount())
		
		// Create token for streaming
		token := &Token{
			ID:       newToken,
			Text:     fmt.Sprintf("tok%d", newToken),
			Position: seq.CurrentPosition,
			StreamID: seq.StreamID,
			Step:     seq.Step,
		}
		
		// Send to response stream
		if seq.Response != nil {
			seq.Response.AccumulateToken(token)
			seq.Response.Stream.Send(token)
		}
		
		// Call token callback
		if seq.OnToken != nil {
			seq.OnToken(token)
		}
		
		// Check completion
		if len(seq.GeneratedTokens) >= seq.MaxNewTokens {
			token.IsFinal = true
			be.batcher.Complete(seq.ID, nil)
		} else {
			// Re-queue for more decoding
			be.batcher.MoveToDecodeQueue(seq)
		}
	}
}

// Stats returns combined statistics
func (be *BatchedEchobeatsEngine) Stats() BatcherStats {
	return be.batcher.Stats()
}
