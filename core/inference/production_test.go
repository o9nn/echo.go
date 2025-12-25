// Package inference provides tests for the production inference engine
package inference

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// MEMORY POOL TESTS
// =============================================================================

func TestMemoryPool(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewMemoryPool(config)
	
	t.Run("Alloc", func(t *testing.T) {
		ptr, err := pool.Alloc(1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ptr == nil {
			t.Fatal("expected non-nil pointer")
		}
	})
	
	t.Run("AllocFloat32", func(t *testing.T) {
		slice, err := pool.AllocFloat32(256)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slice) != 256 {
			t.Errorf("expected length 256, got %d", len(slice))
		}
	})
	
	t.Run("AllocInt32", func(t *testing.T) {
		slice, err := pool.AllocInt32(128)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slice) != 128 {
			t.Errorf("expected length 128, got %d", len(slice))
		}
	})
	
	t.Run("Stats", func(t *testing.T) {
		stats := pool.Stats()
		if stats.TotalAllocated == 0 {
			t.Error("expected non-zero total allocated")
		}
	})
	
	t.Run("Reset", func(t *testing.T) {
		pool.Reset()
		stats := pool.Stats()
		if stats.CurrentUsage != 0 {
			t.Error("expected zero current usage after reset")
		}
	})
	
	pool.Close()
}

func TestStreamAllocator(t *testing.T) {
	config := DefaultPoolConfig()
	allocator := NewStreamAllocator(config)
	
	t.Run("GetPool", func(t *testing.T) {
		pool := allocator.GetPool(StreamAlpha)
		if pool == nil {
			t.Fatal("expected non-nil pool")
		}
		
		ptr, err := pool.Alloc(1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ptr == nil {
			t.Fatal("expected non-nil pointer")
		}
	})
	
	t.Run("GetAllocator", func(t *testing.T) {
		ta := allocator.GetAllocator(StreamBeta)
		if ta == nil {
			t.Fatal("expected non-nil tensor allocator")
		}
	})
	
	t.Run("Stats", func(t *testing.T) {
		stats := allocator.Stats()
		if len(stats) != 3 {
			t.Errorf("expected 3 stats, got %d", len(stats))
		}
	})
	
	t.Run("ResetStream", func(t *testing.T) {
		allocator.ResetStream(StreamAlpha)
		// Just verify no panic
	})
	
	allocator.Close()
}

func TestTensorAllocator(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewMemoryPool(config)
	ta := NewTensorAllocator(pool)
	
	t.Run("AllocTensor", func(t *testing.T) {
		ptr, err := ta.AllocTensor([]int64{2, 3, 4}, "float32", StreamAlpha)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ptr == nil {
			t.Fatal("expected non-nil pointer")
		}
	})
	
	t.Run("TensorCount", func(t *testing.T) {
		count := ta.TensorCount()
		if count == 0 {
			t.Error("expected non-zero tensor count")
		}
	})
	
	t.Run("TensorsByStream", func(t *testing.T) {
		tensors := ta.TensorsByStream(StreamAlpha)
		if len(tensors) == 0 {
			t.Error("expected at least one tensor for StreamAlpha")
		}
	})
	
	pool.Close()
}

// =============================================================================
// STREAMING TESTS
// =============================================================================

func TestTokenStream(t *testing.T) {
	stream := NewTokenStream(StreamAlpha, 1, 100)
	
	t.Run("SendRecv", func(t *testing.T) {
		token := &Token{
			ID:       1,
			Text:     "hello",
			Position: 0,
		}
		
		go func() {
			stream.Send(token)
		}()
		
		received, err := stream.RecvWithTimeout(time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if received == nil {
			t.Fatal("expected non-nil token")
		}
		if received.Text != "hello" {
			t.Errorf("expected 'hello', got '%s'", received.Text)
		}
	})
	
	t.Run("TokenCount", func(t *testing.T) {
		count := stream.TokenCount()
		if count == 0 {
			t.Error("expected non-zero token count")
		}
	})
	
	t.Run("Close", func(t *testing.T) {
		stream.Close()
		
		if !stream.IsClosed() {
			t.Error("expected stream to be closed")
		}
	})
}

func TestStreamingResponse(t *testing.T) {
	resp := NewStreamingResponse(StreamAlpha, 1)
	
	t.Run("AccumulateToken", func(t *testing.T) {
		token := &Token{
			ID:       1,
			Text:     "test",
			Position: 0,
		}
		resp.AccumulateToken(token)
		
		text := resp.GetText()
		if text != "test" {
			t.Errorf("expected 'test', got '%s'", text)
		}
		
		tokens := resp.GetTokens()
		if len(tokens) != 1 {
			t.Errorf("expected 1 token, got %d", len(tokens))
		}
	})
	
	t.Run("ToInferenceResponse", func(t *testing.T) {
		ir := resp.ToInferenceResponse()
		if ir == nil {
			t.Fatal("expected non-nil inference response")
		}
		if ir.Output != "test" {
			t.Errorf("expected output 'test', got '%s'", ir.Output)
		}
	})
}

func TestStreamMultiplexer(t *testing.T) {
	stream1 := NewTokenStream(StreamAlpha, 1, 100)
	stream2 := NewTokenStream(StreamBeta, 1, 100)
	
	mux := NewStreamMultiplexer(stream1, stream2)
	
	t.Run("Recv", func(t *testing.T) {
		// Send a token
		go func() {
			stream1.Send(&Token{ID: 1, Text: "from_alpha"})
		}()
		
		// Receive should work
		token, err := mux.Recv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == nil {
			t.Fatal("expected non-nil token")
		}
	})
	
	t.Run("Close", func(t *testing.T) {
		mux.Close()
		stream1.Close()
		stream2.Close()
	})
}

// =============================================================================
// CONTINUOUS BATCHING TESTS
// =============================================================================

func TestKVCacheManager(t *testing.T) {
	kv := NewKVCacheManager(1000, 10, "lru")
	
	t.Run("Allocate", func(t *testing.T) {
		slot, err := kv.Allocate("seq1", 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slot < 0 {
			t.Error("expected non-negative slot")
		}
	})
	
	t.Run("Release", func(t *testing.T) {
		slot, _ := kv.Allocate("seq2", 100)
		kv.Release(slot)
		
		used, _, _ := kv.Stats()
		if used >= 200 {
			t.Error("expected used tokens to decrease after release")
		}
	})
	
	t.Run("Update", func(t *testing.T) {
		slot, _ := kv.Allocate("seq3", 100)
		kv.Update(slot, 150)
		
		used, _, _ := kv.Stats()
		if used < 150 {
			t.Error("expected used tokens to increase after update")
		}
	})
}

func TestContinuousBatcher(t *testing.T) {
	config := DefaultBatchConfig()
	config.MaxWaitTime = 10 * time.Millisecond
	
	batcher := NewContinuousBatcher(config)
	ctx := context.Background()
	
	t.Run("StartStop", func(t *testing.T) {
		err := batcher.Start(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		batcher.Stop()
	})
	
	t.Run("Submit", func(t *testing.T) {
		batcher = NewContinuousBatcher(config)
		batcher.Start(ctx)
		defer batcher.Stop()
		
		seq := &Sequence{
			StreamID:     StreamAlpha,
			PromptTokens: []int32{1, 2, 3},
			MaxNewTokens: 10,
		}
		
		err := batcher.Submit(seq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	
	t.Run("Stats", func(t *testing.T) {
		stats := batcher.Stats()
		if stats.SequenceCount == 0 {
			t.Error("expected non-zero sequence count")
		}
	})
}

func TestBatch(t *testing.T) {
	batch := NewBatch(1, "prefill")
	
	t.Run("Add", func(t *testing.T) {
		seq := &Sequence{
			StreamID:     StreamAlpha,
			PromptTokens: []int32{1, 2, 3},
		}
		batch.Add(seq)
		
		if batch.Size() != 1 {
			t.Errorf("expected size 1, got %d", batch.Size())
		}
	})
}

func TestSequence(t *testing.T) {
	seq := &Sequence{
		StreamID:     StreamAlpha,
		PromptTokens: []int32{1, 2, 3, 4, 5},
		MaxNewTokens: 10,
		GeneratedTokens: []int32{100, 101},
	}
	
	t.Run("TokenCount", func(t *testing.T) {
		count := seq.TokenCount()
		if count != 7 { // 5 prompt + 2 generated
			t.Errorf("expected 7 tokens, got %d", count)
		}
	})
	
	t.Run("IsComplete", func(t *testing.T) {
		if seq.IsComplete() {
			t.Error("expected sequence to not be complete")
		}
		
		seq.State = SequenceStateComplete
		if !seq.IsComplete() {
			t.Error("expected sequence to be complete")
		}
	})
}

// =============================================================================
// SPECULATIVE DECODING TESTS
// =============================================================================

func TestSpeculativeEngine(t *testing.T) {
	config := DefaultSpeculativeConfig()
	engine := NewSpeculativeEngine(config)
	
	t.Run("DraftLength", func(t *testing.T) {
		if engine.currentDraftLen[0] != config.DraftTokens {
			t.Errorf("expected draft length %d, got %d", config.DraftTokens, engine.currentDraftLen[0])
		}
	})
	
	t.Run("Stats", func(t *testing.T) {
		stats := engine.Stats()
		if stats.TotalIterations != 0 {
			t.Error("expected zero iterations initially")
		}
	})
}

func TestTreeSpeculator(t *testing.T) {
	config := DefaultSpeculativeConfig()
	ts := NewTreeSpeculator(config, 3, 2)
	
	t.Run("GenerateTree", func(t *testing.T) {
		ctx := context.Background()
		tree, err := ts.GenerateTree(ctx, []int32{1, 2, 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tree == nil {
			t.Fatal("expected non-nil tree")
		}
		if len(tree.Children) == 0 {
			t.Error("expected tree to have children")
		}
	})
}

func TestSpeculativeConfig(t *testing.T) {
	config := DefaultSpeculativeConfig()
	
	if config.DraftTokens != 4 {
		t.Errorf("expected draft tokens 4, got %d", config.DraftTokens)
	}
	if config.AcceptanceMethod != "typical" {
		t.Errorf("expected acceptance method 'typical', got '%s'", config.AcceptanceMethod)
	}
}

func TestDraftSequence(t *testing.T) {
	draft := &DraftSequence{
		Tokens:   []int32{1, 2, 3, 4},
		Logprobs: []float32{-0.1, -0.2, -0.3, -0.4},
		Accepted: 3,
		Rejected: 3,
	}
	
	if len(draft.Tokens) != 4 {
		t.Errorf("expected 4 tokens, got %d", len(draft.Tokens))
	}
}

// =============================================================================
// STATE PERSISTENCE TESTS
// =============================================================================

func TestStateManager(t *testing.T) {
	tmpDir := t.TempDir()
	
	config := DefaultStateConfig()
	config.StorageDir = tmpDir
	
	sm := NewStateManager(config)
	
	t.Run("Initialize", func(t *testing.T) {
		err := sm.Initialize()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	
	t.Run("SaveState", func(t *testing.T) {
		engine := NewEchobeatsEngine()
		
		info, err := sm.SaveState(engine, "test snapshot")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info == nil {
			t.Fatal("expected non-nil snapshot info")
		}
		if info.Path == "" {
			t.Error("expected non-empty path")
		}
	})
	
	t.Run("ListSnapshots", func(t *testing.T) {
		snapshots := sm.ListSnapshots()
		if len(snapshots) == 0 {
			t.Error("expected at least one snapshot")
		}
	})
	
	t.Run("LoadState", func(t *testing.T) {
		snapshots := sm.ListSnapshots()
		if len(snapshots) == 0 {
			t.Skip("no snapshots to load")
		}
		
		state, err := sm.LoadState(snapshots[0].Path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state == nil {
			t.Fatal("expected non-nil state")
		}
	})
	
	t.Run("CreateCheckpoint", func(t *testing.T) {
		engine := NewEchobeatsEngine()
		
		info, err := sm.CreateCheckpoint(engine, "test checkpoint")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !info.IsCheckpoint {
			t.Error("expected checkpoint flag to be true")
		}
	})
	
	sm.Close()
}

func TestBinaryStateFormat(t *testing.T) {
	state := &CognitiveState{
		Version:     1,
		Timestamp:   time.Now(),
		Description: "test state",
	}
	
	tmpFile := filepath.Join(t.TempDir(), "test_state.bin")
	
	t.Run("Write", func(t *testing.T) {
		f, err := os.Create(tmpFile)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		defer f.Close()
		
		err = WriteBinaryState(f, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	
	t.Run("Read", func(t *testing.T) {
		f, err := os.Open(tmpFile)
		if err != nil {
			t.Fatalf("failed to open file: %v", err)
		}
		defer f.Close()
		
		loaded, err := ReadBinaryState(f)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loaded.Description != state.Description {
			t.Errorf("expected description '%s', got '%s'", state.Description, loaded.Description)
		}
	})
}

func TestIncrementalStateManager(t *testing.T) {
	baseState := &CognitiveState{
		Version:   1,
		Timestamp: time.Now(),
		Checksum:  "test-checksum",
	}
	
	ism := NewIncrementalStateManager(baseState)
	
	t.Run("RecordDelta", func(t *testing.T) {
		ism.RecordDelta("current_step", float64(5))
		ism.RecordDelta("cycle_count", float64(100))
	})
	
	t.Run("ApplyDeltas", func(t *testing.T) {
		state := ism.ApplyDeltas()
		if state == nil {
			t.Fatal("expected non-nil state")
		}
	})
	
	t.Run("Compact", func(t *testing.T) {
		ism.Compact()
		state := ism.ApplyDeltas()
		if state == nil {
			t.Fatal("expected non-nil state after compact")
		}
	})
}

// =============================================================================
// PRODUCTION ENGINE TESTS
// =============================================================================

func TestProductionConfig(t *testing.T) {
	config := DefaultProductionConfig()
	
	if config.NumStreams != 3 {
		t.Errorf("expected 3 streams, got %d", config.NumStreams)
	}
	if !config.EnableStreaming {
		t.Error("expected streaming to be enabled by default")
	}
	if !config.EnableBatching {
		t.Error("expected batching to be enabled by default")
	}
	if config.EnableSpeculative {
		t.Error("expected speculative to be disabled by default")
	}
}

func TestProductionEngineCreation(t *testing.T) {
	t.Run("NewProductionEngine", func(t *testing.T) {
		config := DefaultProductionConfig()
		engine := NewProductionEngine(config)
		
		if engine == nil {
			t.Fatal("expected non-nil engine")
		}
	})
	
	t.Run("CreateWithOptions", func(t *testing.T) {
		engine, err := CreateProductionEngine(
			"/tmp/test_model.gguf",
			WithContextSize(4096),
			WithBatchSize(32),
		)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if engine == nil {
			t.Fatal("expected non-nil engine")
		}
		if engine.config.Engine.ContextSize != 4096 {
			t.Errorf("expected context size 4096, got %d", engine.config.Engine.ContextSize)
		}
	})
	
	t.Run("WithDraftModel", func(t *testing.T) {
		engine, _ := CreateProductionEngine(
			"/tmp/test_model.gguf",
			WithDraftModel("/tmp/draft_model.gguf"),
		)
		
		if !engine.config.EnableSpeculative {
			t.Error("expected speculative to be enabled with draft model")
		}
	})
	
	t.Run("WithGPU", func(t *testing.T) {
		engine, _ := CreateProductionEngine(
			"/tmp/test_model.gguf",
			WithGPU(0, 1),
		)
		
		if !engine.config.EnableGPU {
			t.Error("expected GPU to be enabled")
		}
		if len(engine.config.GPUDevices) != 2 {
			t.Errorf("expected 2 GPU devices, got %d", len(engine.config.GPUDevices))
		}
	})
}

func TestProductionEngineHealth(t *testing.T) {
	config := DefaultProductionConfig()
	engine := NewProductionEngine(config)
	
	health := engine.Health()
	
	if health.Initialized {
		t.Error("expected not initialized")
	}
	if health.Running {
		t.Error("expected not running")
	}
	if health.Healthy {
		t.Error("expected not healthy before initialization")
	}
	if len(health.Components) == 0 {
		t.Error("expected components map to be populated")
	}
}

func TestProductionMetrics(t *testing.T) {
	config := DefaultProductionConfig()
	engine := NewProductionEngine(config)
	
	metrics := engine.GetMetrics()
	
	if metrics.TotalRequests != 0 {
		t.Error("expected zero requests initially")
	}
	if metrics.MinLatencyMs != ^uint64(0) {
		t.Error("expected min latency to be max uint64 initially")
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestStreamTriads(t *testing.T) {
	// Test the 12-step cognitive loop triads
	triads := [][]int{
		{1, 5, 9},
		{2, 6, 10},
		{3, 7, 11},
		{4, 8, 12},
	}
	
	for i, triad := range triads {
		// Verify 4-step spacing
		if triad[1]-triad[0] != 4 {
			t.Errorf("triad %d: expected 4-step spacing between first and second, got %d", i, triad[1]-triad[0])
		}
		if triad[2]-triad[1] != 4 {
			t.Errorf("triad %d: expected 4-step spacing between second and third, got %d", i, triad[2]-triad[1])
		}
	}
}

func TestConcurrentStreams(t *testing.T) {
	config := DefaultBatchConfig()
	batcher := NewContinuousBatcher(config)
	ctx := context.Background()
	
	batcher.Start(ctx)
	defer batcher.Stop()
	
	// Submit sequences for all 3 streams
	for i := 0; i < 3; i++ {
		seq := &Sequence{
			StreamID:     StreamID(i),
			PromptTokens: []int32{int32(i * 100), int32(i*100 + 1)},
			MaxNewTokens: 10,
		}
		
		err := batcher.Submit(seq)
		if err != nil {
			t.Fatalf("failed to submit sequence for stream %d: %v", i, err)
		}
	}
	
	stats := batcher.Stats()
	if stats.SequenceCount < 3 {
		t.Errorf("expected at least 3 sequences, got %d", stats.SequenceCount)
	}
}

func TestStreamIDConstants(t *testing.T) {
	if StreamAlpha != 0 {
		t.Errorf("expected StreamAlpha to be 0, got %d", StreamAlpha)
	}
	if StreamBeta != 1 {
		t.Errorf("expected StreamBeta to be 1, got %d", StreamBeta)
	}
	if StreamGamma != 2 {
		t.Errorf("expected StreamGamma to be 2, got %d", StreamGamma)
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkMemoryPoolAlloc(b *testing.B) {
	config := DefaultPoolConfig()
	pool := NewMemoryPool(config)
	defer pool.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ptr, _ := pool.Alloc(1024)
		pool.Free(ptr, 1024)
	}
}

func BenchmarkTokenStreamSendRecv(b *testing.B) {
	stream := NewTokenStream(StreamAlpha, 1, 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := &Token{ID: int32(i), Text: "test"}
		go stream.Send(token)
		stream.RecvWithTimeout(time.Second)
	}
	
	stream.Close()
}

func BenchmarkKVCacheAllocate(b *testing.B) {
	kv := NewKVCacheManager(100000, 1000, "lru")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slot, _ := kv.Allocate("seq", 100)
		kv.Release(slot)
	}
}

func BenchmarkBatchFormation(b *testing.B) {
	config := DefaultBatchConfig()
	config.MaxWaitTime = time.Millisecond
	
	batcher := NewContinuousBatcher(config)
	ctx := context.Background()
	batcher.Start(ctx)
	defer batcher.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seq := &Sequence{
			StreamID:     StreamID(i % 3),
			PromptTokens: []int32{1, 2, 3},
			MaxNewTokens: 10,
		}
		batcher.Submit(seq)
	}
}
