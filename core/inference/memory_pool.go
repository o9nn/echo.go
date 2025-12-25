// Package inference provides memory pooling for high-throughput tensor allocations
// This implements arena-based allocation for the echobeats inference engine
package inference

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

// =============================================================================
// MEMORY POOL CONFIGURATION
// =============================================================================

// PoolConfig configures the memory pool
type PoolConfig struct {
	// Arena configuration
	ArenaSize      int64 // Size of each arena in bytes (default: 256MB)
	MaxArenas      int   // Maximum number of arenas (default: 16)
	PreallocArenas int   // Number of arenas to preallocate (default: 2)
	
	// Block configuration
	MinBlockSize int64 // Minimum block size (default: 64 bytes)
	MaxBlockSize int64 // Maximum block size (default: 64MB)
	Alignment    int64 // Memory alignment (default: 64 bytes for cache line)
	
	// Pool behavior
	GrowOnDemand    bool // Grow pool when exhausted (default: true)
	ZeroOnAlloc     bool // Zero memory on allocation (default: false)
	ZeroOnFree      bool // Zero memory on free (default: false)
	TrackAllocations bool // Track allocation statistics (default: true)
}

// DefaultPoolConfig returns default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		ArenaSize:        256 * 1024 * 1024, // 256MB
		MaxArenas:        16,
		PreallocArenas:   2,
		MinBlockSize:     64,
		MaxBlockSize:     64 * 1024 * 1024, // 64MB
		Alignment:        64,               // Cache line alignment
		GrowOnDemand:     true,
		ZeroOnAlloc:      false,
		ZeroOnFree:       false,
		TrackAllocations: true,
	}
}

// =============================================================================
// MEMORY ARENA
// =============================================================================

// Arena represents a contiguous memory region
type Arena struct {
	data      []byte
	offset    int64
	size      int64
	id        int
	allocations int64
	mu        sync.Mutex
}

// newArena creates a new arena
func newArena(size int64, id int) *Arena {
	return &Arena{
		data: make([]byte, size),
		size: size,
		id:   id,
	}
}

// alloc allocates memory from the arena
func (a *Arena) alloc(size, alignment int64) (unsafe.Pointer, int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Align the offset
	alignedOffset := (a.offset + alignment - 1) &^ (alignment - 1)
	
	// Check if we have enough space
	if alignedOffset+size > a.size {
		return nil, 0, errors.New("arena exhausted")
	}
	
	ptr := unsafe.Pointer(&a.data[alignedOffset])
	offset := alignedOffset
	a.offset = alignedOffset + size
	a.allocations++
	
	return ptr, offset, nil
}

// reset resets the arena for reuse
func (a *Arena) reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.offset = 0
	a.allocations = 0
}

// available returns the available space in the arena
func (a *Arena) available() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.size - a.offset
}

// =============================================================================
// BLOCK POOL (for small allocations)
// =============================================================================

// BlockPool manages fixed-size memory blocks
type BlockPool struct {
	blockSize int64
	pool      sync.Pool
	allocated int64
	freed     int64
}

// newBlockPool creates a new block pool
func newBlockPool(blockSize int64) *BlockPool {
	bp := &BlockPool{
		blockSize: blockSize,
	}
	bp.pool.New = func() interface{} {
		return make([]byte, blockSize)
	}
	return bp
}

// get gets a block from the pool
func (bp *BlockPool) get() []byte {
	atomic.AddInt64(&bp.allocated, 1)
	return bp.pool.Get().([]byte)
}

// put returns a block to the pool
func (bp *BlockPool) put(block []byte) {
	atomic.AddInt64(&bp.freed, 1)
	bp.pool.Put(block)
}

// =============================================================================
// MEMORY POOL
// =============================================================================

// MemoryPool manages memory for tensor allocations
type MemoryPool struct {
	config PoolConfig
	
	// Arenas for large allocations
	arenas    []*Arena
	arenasMu  sync.RWMutex
	
	// Block pools for small allocations (power of 2 sizes)
	blockPools map[int64]*BlockPool
	blocksMu   sync.RWMutex
	
	// Statistics
	stats PoolStats
	
	// State
	closed atomic.Bool
}

// PoolStats tracks memory pool statistics
type PoolStats struct {
	TotalAllocated   int64
	TotalFreed       int64
	CurrentUsage     int64
	PeakUsage        int64
	ArenaAllocations int64
	BlockAllocations int64
	AllocationCount  int64
	FreeCount        int64
	OOMCount         int64
	mu               sync.RWMutex
}

// NewMemoryPool creates a new memory pool
func NewMemoryPool(config PoolConfig) *MemoryPool {
	mp := &MemoryPool{
		config:     config,
		arenas:     make([]*Arena, 0, config.MaxArenas),
		blockPools: make(map[int64]*BlockPool),
	}
	
	// Preallocate arenas
	for i := 0; i < config.PreallocArenas; i++ {
		arena := newArena(config.ArenaSize, i)
		mp.arenas = append(mp.arenas, arena)
	}
	
	// Create block pools for power-of-2 sizes
	for size := config.MinBlockSize; size <= config.MaxBlockSize; size *= 2 {
		mp.blockPools[size] = newBlockPool(size)
	}
	
	return mp
}

// Alloc allocates memory from the pool
func (mp *MemoryPool) Alloc(size int64) (unsafe.Pointer, error) {
	if mp.closed.Load() {
		return nil, errors.New("pool is closed")
	}
	
	if size <= 0 {
		return nil, errors.New("invalid allocation size")
	}
	
	// Round up to alignment
	alignedSize := (size + mp.config.Alignment - 1) &^ (mp.config.Alignment - 1)
	
	var ptr unsafe.Pointer
	var err error
	
	// Use block pool for small allocations
	if alignedSize <= mp.config.MaxBlockSize {
		ptr, err = mp.allocFromBlockPool(alignedSize)
	} else {
		ptr, err = mp.allocFromArena(alignedSize)
	}
	
	if err != nil {
		mp.stats.mu.Lock()
		mp.stats.OOMCount++
		mp.stats.mu.Unlock()
		return nil, err
	}
	
	// Update statistics
	if mp.config.TrackAllocations {
		mp.stats.mu.Lock()
		mp.stats.TotalAllocated += alignedSize
		mp.stats.CurrentUsage += alignedSize
		if mp.stats.CurrentUsage > mp.stats.PeakUsage {
			mp.stats.PeakUsage = mp.stats.CurrentUsage
		}
		mp.stats.AllocationCount++
		mp.stats.mu.Unlock()
	}
	
	// Zero memory if configured
	if mp.config.ZeroOnAlloc {
		mp.zero(ptr, alignedSize)
	}
	
	return ptr, nil
}

// allocFromBlockPool allocates from the appropriate block pool
func (mp *MemoryPool) allocFromBlockPool(size int64) (unsafe.Pointer, error) {
	// Find the smallest block size that fits
	blockSize := mp.config.MinBlockSize
	for blockSize < size {
		blockSize *= 2
	}
	
	mp.blocksMu.RLock()
	pool, ok := mp.blockPools[blockSize]
	mp.blocksMu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("no block pool for size %d", blockSize)
	}
	
	block := pool.get()
	
	mp.stats.mu.Lock()
	mp.stats.BlockAllocations++
	mp.stats.mu.Unlock()
	
	return unsafe.Pointer(&block[0]), nil
}

// allocFromArena allocates from an arena
func (mp *MemoryPool) allocFromArena(size int64) (unsafe.Pointer, error) {
	mp.arenasMu.RLock()
	
	// Try existing arenas
	for _, arena := range mp.arenas {
		if arena.available() >= size {
			ptr, _, err := arena.alloc(size, mp.config.Alignment)
			if err == nil {
				mp.arenasMu.RUnlock()
				mp.stats.mu.Lock()
				mp.stats.ArenaAllocations++
				mp.stats.mu.Unlock()
				return ptr, nil
			}
		}
	}
	mp.arenasMu.RUnlock()
	
	// Need to grow
	if !mp.config.GrowOnDemand {
		return nil, errors.New("pool exhausted and growth disabled")
	}
	
	mp.arenasMu.Lock()
	defer mp.arenasMu.Unlock()
	
	if len(mp.arenas) >= mp.config.MaxArenas {
		return nil, errors.New("maximum arenas reached")
	}
	
	// Create new arena
	arenaSize := mp.config.ArenaSize
	if size > arenaSize {
		arenaSize = size * 2 // Create arena at least 2x the requested size
	}
	
	arena := newArena(arenaSize, len(mp.arenas))
	mp.arenas = append(mp.arenas, arena)
	
	ptr, _, err := arena.alloc(size, mp.config.Alignment)
	if err != nil {
		return nil, err
	}
	
	mp.stats.mu.Lock()
	mp.stats.ArenaAllocations++
	mp.stats.mu.Unlock()
	
	return ptr, nil
}

// Free returns memory to the pool
// Note: For arena allocations, memory is only truly freed on Reset()
func (mp *MemoryPool) Free(ptr unsafe.Pointer, size int64) {
	if ptr == nil {
		return
	}
	
	alignedSize := (size + mp.config.Alignment - 1) &^ (mp.config.Alignment - 1)
	
	// Zero memory if configured
	if mp.config.ZeroOnFree {
		mp.zero(ptr, alignedSize)
	}
	
	// Return to block pool if applicable
	if alignedSize <= mp.config.MaxBlockSize {
		blockSize := mp.config.MinBlockSize
		for blockSize < alignedSize {
			blockSize *= 2
		}
		
		mp.blocksMu.RLock()
		pool, ok := mp.blockPools[blockSize]
		mp.blocksMu.RUnlock()
		
		if ok {
			// Convert pointer back to slice
			block := unsafe.Slice((*byte)(ptr), blockSize)
			pool.put(block)
		}
	}
	
	// Update statistics
	if mp.config.TrackAllocations {
		mp.stats.mu.Lock()
		mp.stats.TotalFreed += alignedSize
		mp.stats.CurrentUsage -= alignedSize
		mp.stats.FreeCount++
		mp.stats.mu.Unlock()
	}
}

// AllocSlice allocates a slice of the given type and length
func AllocSlice[T any](mp *MemoryPool, length int) ([]T, error) {
	var zero T
	elemSize := int64(unsafe.Sizeof(zero))
	totalSize := elemSize * int64(length)
	
	ptr, err := mp.Alloc(totalSize)
	if err != nil {
		return nil, err
	}
	
	return unsafe.Slice((*T)(ptr), length), nil
}

// AllocFloat32 allocates a float32 slice
func (mp *MemoryPool) AllocFloat32(length int) ([]float32, error) {
	return AllocSlice[float32](mp, length)
}

// AllocInt32 allocates an int32 slice
func (mp *MemoryPool) AllocInt32(length int) ([]int32, error) {
	return AllocSlice[int32](mp, length)
}

// Reset resets all arenas for reuse
func (mp *MemoryPool) Reset() {
	mp.arenasMu.Lock()
	defer mp.arenasMu.Unlock()
	
	for _, arena := range mp.arenas {
		arena.reset()
	}
	
	// Reset statistics
	mp.stats.mu.Lock()
	mp.stats.CurrentUsage = 0
	mp.stats.mu.Unlock()
}

// Stats returns the current pool statistics
func (mp *MemoryPool) Stats() PoolStats {
	mp.stats.mu.RLock()
	defer mp.stats.mu.RUnlock()
	return mp.stats
}

// Close closes the memory pool and releases all resources
func (mp *MemoryPool) Close() error {
	if mp.closed.Swap(true) {
		return errors.New("pool already closed")
	}
	
	mp.arenasMu.Lock()
	defer mp.arenasMu.Unlock()
	
	// Clear arenas
	for i := range mp.arenas {
		mp.arenas[i] = nil
	}
	mp.arenas = nil
	
	// Clear block pools
	mp.blocksMu.Lock()
	mp.blockPools = nil
	mp.blocksMu.Unlock()
	
	// Force GC
	runtime.GC()
	
	return nil
}

// zero zeros memory at the given pointer
func (mp *MemoryPool) zero(ptr unsafe.Pointer, size int64) {
	slice := unsafe.Slice((*byte)(ptr), size)
	for i := range slice {
		slice[i] = 0
	}
}

// =============================================================================
// TENSOR ALLOCATOR
// =============================================================================

// TensorAllocator provides tensor-specific memory allocation
type TensorAllocator struct {
	pool *MemoryPool
	
	// Tensor metadata tracking
	tensors   map[uintptr]*TensorMeta
	tensorsMu sync.RWMutex
}

// TensorMeta tracks tensor metadata
type TensorMeta struct {
	Ptr       unsafe.Pointer
	Size      int64
	Shape     []int64
	Dtype     string
	StreamID  StreamID
	Timestamp int64
}

// NewTensorAllocator creates a new tensor allocator
func NewTensorAllocator(pool *MemoryPool) *TensorAllocator {
	return &TensorAllocator{
		pool:    pool,
		tensors: make(map[uintptr]*TensorMeta),
	}
}

// AllocTensor allocates memory for a tensor
func (ta *TensorAllocator) AllocTensor(shape []int64, dtype string, streamID StreamID) (unsafe.Pointer, error) {
	// Calculate size based on dtype
	var elemSize int64
	switch dtype {
	case "float32", "f32":
		elemSize = 4
	case "float16", "f16":
		elemSize = 2
	case "int32", "i32":
		elemSize = 4
	case "int8", "i8":
		elemSize = 1
	case "bfloat16", "bf16":
		elemSize = 2
	default:
		return nil, fmt.Errorf("unsupported dtype: %s", dtype)
	}
	
	// Calculate total elements
	totalElements := int64(1)
	for _, dim := range shape {
		totalElements *= dim
	}
	
	size := totalElements * elemSize
	
	ptr, err := ta.pool.Alloc(size)
	if err != nil {
		return nil, err
	}
	
	// Track tensor metadata
	ta.tensorsMu.Lock()
	ta.tensors[uintptr(ptr)] = &TensorMeta{
		Ptr:       ptr,
		Size:      size,
		Shape:     shape,
		Dtype:     dtype,
		StreamID:  streamID,
		Timestamp: currentTimestamp(),
	}
	ta.tensorsMu.Unlock()
	
	return ptr, nil
}

// FreeTensor frees tensor memory
func (ta *TensorAllocator) FreeTensor(ptr unsafe.Pointer) {
	ta.tensorsMu.Lock()
	meta, ok := ta.tensors[uintptr(ptr)]
	if ok {
		delete(ta.tensors, uintptr(ptr))
	}
	ta.tensorsMu.Unlock()
	
	if ok {
		ta.pool.Free(ptr, meta.Size)
	}
}

// GetTensorMeta returns metadata for a tensor
func (ta *TensorAllocator) GetTensorMeta(ptr unsafe.Pointer) (*TensorMeta, bool) {
	ta.tensorsMu.RLock()
	defer ta.tensorsMu.RUnlock()
	meta, ok := ta.tensors[uintptr(ptr)]
	return meta, ok
}

// TensorCount returns the number of tracked tensors
func (ta *TensorAllocator) TensorCount() int {
	ta.tensorsMu.RLock()
	defer ta.tensorsMu.RUnlock()
	return len(ta.tensors)
}

// TensorsByStream returns tensors allocated for a specific stream
func (ta *TensorAllocator) TensorsByStream(streamID StreamID) []*TensorMeta {
	ta.tensorsMu.RLock()
	defer ta.tensorsMu.RUnlock()
	
	var result []*TensorMeta
	for _, meta := range ta.tensors {
		if meta.StreamID == streamID {
			result = append(result, meta)
		}
	}
	return result
}

// Reset frees all tensors and resets the pool
func (ta *TensorAllocator) Reset() {
	ta.tensorsMu.Lock()
	ta.tensors = make(map[uintptr]*TensorMeta)
	ta.tensorsMu.Unlock()
	
	ta.pool.Reset()
}

// =============================================================================
// STREAM-LOCAL ALLOCATOR
// =============================================================================

// StreamAllocator provides per-stream memory allocation
type StreamAllocator struct {
	allocators [3]*TensorAllocator
	pools      [3]*MemoryPool
}

// NewStreamAllocator creates allocators for all 3 streams
func NewStreamAllocator(config PoolConfig) *StreamAllocator {
	sa := &StreamAllocator{}
	
	for i := 0; i < 3; i++ {
		sa.pools[i] = NewMemoryPool(config)
		sa.allocators[i] = NewTensorAllocator(sa.pools[i])
	}
	
	return sa
}

// GetAllocator returns the allocator for a specific stream
func (sa *StreamAllocator) GetAllocator(streamID StreamID) *TensorAllocator {
	return sa.allocators[streamID]
}

// GetPool returns the pool for a specific stream
func (sa *StreamAllocator) GetPool(streamID StreamID) *MemoryPool {
	return sa.pools[streamID]
}

// ResetStream resets a specific stream's allocator
func (sa *StreamAllocator) ResetStream(streamID StreamID) {
	sa.allocators[streamID].Reset()
}

// ResetAll resets all stream allocators
func (sa *StreamAllocator) ResetAll() {
	for i := 0; i < 3; i++ {
		sa.allocators[i].Reset()
	}
}

// Stats returns combined statistics for all streams
func (sa *StreamAllocator) Stats() [3]PoolStats {
	var stats [3]PoolStats
	for i := 0; i < 3; i++ {
		stats[i] = sa.pools[i].Stats()
	}
	return stats
}

// Close closes all stream allocators
func (sa *StreamAllocator) Close() error {
	var errs []error
	for i := 0; i < 3; i++ {
		if err := sa.pools[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing pools: %v", errs)
	}
	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func currentTimestamp() int64 {
	return int64(runtime.NumCPU()) // Placeholder - would use time.Now().UnixNano()
}
