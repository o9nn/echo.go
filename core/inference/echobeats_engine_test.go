package inference

import (
	"context"
	"testing"
	"time"
)

func TestStreamID(t *testing.T) {
	tests := []struct {
		stream StreamID
		name   string
		phase  int
		triad  [3]int
	}{
		{StreamAlpha, "Alpha", 0, [3]int{1, 5, 9}},
		{StreamBeta, "Beta", 4, [3]int{2, 6, 10}},
		{StreamGamma, "Gamma", 8, [3]int{3, 7, 11}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.stream.String(); got != tt.name {
				t.Errorf("StreamID.String() = %v, want %v", got, tt.name)
			}
			if got := tt.stream.Phase(); got != tt.phase {
				t.Errorf("StreamID.Phase() = %v, want %v", got, tt.phase)
			}
			if got := tt.stream.StepTriad(); got != tt.triad {
				t.Errorf("StreamID.StepTriad() = %v, want %v", got, tt.triad)
			}
		})
	}
}

func TestStepType(t *testing.T) {
	// Test step type mapping for all 12 steps
	expectedTypes := map[int]StepType{
		1:  StepRelevanceRealization,
		2:  StepAffordanceInteraction,
		3:  StepAffordanceInteraction,
		4:  StepAffordanceInteraction,
		5:  StepAffordanceInteraction,
		6:  StepAffordanceInteraction,
		7:  StepRelevanceRealization,
		8:  StepSalienceSimulation,
		9:  StepSalienceSimulation,
		10: StepSalienceSimulation,
		11: StepSalienceSimulation,
		12: StepSalienceSimulation,
	}

	for step, expected := range expectedTypes {
		got := GetStepType(step)
		if got != expected {
			t.Errorf("GetStepType(%d) = %v, want %v", step, got, expected)
		}
	}
}

func TestEchobeatsEngineInitialization(t *testing.T) {
	engine := NewEchobeatsEngine()
	if engine == nil {
		t.Fatal("NewEchobeatsEngine() returned nil")
	}

	config := DefaultEngineConfig()
	err := engine.Initialize("/tmp/test_model.gguf", config)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	state := engine.GetState()
	if !state.Initialized {
		t.Error("Engine should be initialized")
	}

	// Test double initialization
	err = engine.Initialize("/tmp/test_model.gguf", config)
	if err == nil {
		t.Error("Double initialization should return error")
	}

	engine.Close()
}

func TestEchobeatsEngineSingleStreamInference(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()
	req := &InferenceRequest{
		Step:        1,
		StepType:    StepRelevanceRealization,
		Prompt:      "Test prompt for inference",
		MaxTokens:   100,
		Temperature: 0.7,
	}

	resp, err := engine.InferStream(ctx, StreamAlpha, req)
	if err != nil {
		t.Fatalf("InferStream() error = %v", err)
	}

	if resp.StreamID != StreamAlpha {
		t.Errorf("Response StreamID = %v, want %v", resp.StreamID, StreamAlpha)
	}
	if resp.Step != 1 {
		t.Errorf("Response Step = %v, want 1", resp.Step)
	}
	if resp.Output == "" {
		t.Error("Response Output should not be empty")
	}
}

func TestEchobeatsEngineConcurrentInference(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()
	requests := [3]*InferenceRequest{
		{Step: 1, Prompt: "Alpha prompt", MaxTokens: 50},
		{Step: 2, Prompt: "Beta prompt", MaxTokens: 50},
		{Step: 3, Prompt: "Gamma prompt", MaxTokens: 50},
	}

	responses, err := engine.InferConcurrent(ctx, requests)
	if err != nil {
		t.Fatalf("InferConcurrent() error = %v", err)
	}

	for i, resp := range responses {
		if resp == nil {
			t.Errorf("Response %d is nil", i)
			continue
		}
		if resp.StreamID != StreamID(i) {
			t.Errorf("Response %d StreamID = %v, want %v", i, resp.StreamID, StreamID(i))
		}
	}
}

func TestEchobeatsEngineCognitiveStep(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()

	// Test step 1 (should activate Alpha stream)
	responses, err := engine.ExecuteCognitiveStep(ctx, 1, "Initial prompt")
	if err != nil {
		t.Fatalf("ExecuteCognitiveStep(1) error = %v", err)
	}

	// Step 1 is in Alpha's triad {1,5,9}
	if responses[0] == nil {
		t.Error("Alpha stream should have response for step 1")
	}

	// Test invalid step
	_, err = engine.ExecuteCognitiveStep(ctx, 0, "Invalid step")
	if err == nil {
		t.Error("ExecuteCognitiveStep(0) should return error")
	}

	_, err = engine.ExecuteCognitiveStep(ctx, 13, "Invalid step")
	if err == nil {
		t.Error("ExecuteCognitiveStep(13) should return error")
	}
}

func TestEchobeatsEngineFullCycle(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := engine.ExecuteFullCycle(ctx, "Start of cognitive cycle")
	if err != nil {
		t.Fatalf("ExecuteFullCycle() error = %v", err)
	}

	if len(results) != 12 {
		t.Errorf("ExecuteFullCycle() returned %d results, want 12", len(results))
	}

	state := engine.GetState()
	if state.CycleCount != 1 {
		t.Errorf("CycleCount = %d, want 1", state.CycleCount)
	}
}

func TestEchobeatsEngineMetrics(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()

	// Perform some inferences
	for i := 0; i < 3; i++ {
		req := &InferenceRequest{
			Step:      i + 1,
			Prompt:    "Test prompt",
			MaxTokens: 10,
		}
		engine.InferStream(ctx, StreamID(i), req)
	}

	metrics := engine.GetMetrics()
	if metrics.TotalInferences != 3 {
		t.Errorf("TotalInferences = %d, want 3", metrics.TotalInferences)
	}

	// Check per-stream metrics
	for i := 0; i < 3; i++ {
		if metrics.StreamInfers[i] != 1 {
			t.Errorf("StreamInfers[%d] = %d, want 1", i, metrics.StreamInfers[i])
		}
	}
}

func TestEchobeatsEngineContextCancellation(t *testing.T) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := engine.ExecuteFullCycle(ctx, "Should be cancelled")
	if err == nil {
		t.Error("ExecuteFullCycle() should return error on cancelled context")
	}
}

func TestMockInferenceEngine(t *testing.T) {
	engine := &MockInferenceEngine{streamID: StreamAlpha}
	config := DefaultEngineConfig()

	err := engine.Initialize("/tmp/test_model.gguf", config)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	state := engine.GetState()
	if !state.Initialized {
		t.Error("Engine should be initialized")
	}

	ctx := context.Background()
	req := &InferenceRequest{
		StreamID:  StreamAlpha,
		Step:      1,
		Prompt:    "Test prompt",
		MaxTokens: 50,
	}

	resp, err := engine.Infer(ctx, req)
	if err != nil {
		t.Fatalf("Infer() error = %v", err)
	}
	if resp.Output == "" {
		t.Error("Response Output should not be empty")
	}

	// Test embeddings
	embeddings, err := engine.Embed(ctx, "Test input")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if len(embeddings) != 768 {
		t.Errorf("Embeddings length = %d, want 768", len(embeddings))
	}

	// Test tokenization
	tokens, err := engine.Tokenize("Hello world")
	if err != nil {
		t.Fatalf("Tokenize() error = %v", err)
	}
	if len(tokens) == 0 {
		t.Error("Tokens should not be empty")
	}

	engine.Close()
}

func TestEngineFactory(t *testing.T) {
	tests := []struct {
		engineType EngineType
		name       string
	}{
		{EngineTypeMock, "Mock"},
		{EngineTypeLlama, "Llama"},
		{EngineTypeVulkan, "Vulkan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := CreateEngine(tt.engineType, StreamAlpha)
			if engine == nil {
				t.Errorf("CreateEngine(%v) returned nil", tt.engineType)
			}
		})
	}
}

func TestCreateEchobeatsEngineWithType(t *testing.T) {
	engine := CreateEchobeatsEngineWithType(EngineTypeMock)
	if engine == nil {
		t.Fatal("CreateEchobeatsEngineWithType() returned nil")
	}

	config := DefaultEngineConfig()
	err := engine.Initialize("/tmp/test_model.gguf", config)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	engine.Close()
}

func BenchmarkEchobeatsEngineSingleInference(b *testing.B) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()
	req := &InferenceRequest{
		Step:      1,
		Prompt:    "Benchmark prompt",
		MaxTokens: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.InferStream(ctx, StreamAlpha, req)
	}
}

func BenchmarkEchobeatsEngineConcurrentInference(b *testing.B) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()
	requests := [3]*InferenceRequest{
		{Step: 1, Prompt: "Alpha prompt", MaxTokens: 50},
		{Step: 2, Prompt: "Beta prompt", MaxTokens: 50},
		{Step: 3, Prompt: "Gamma prompt", MaxTokens: 50},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.InferConcurrent(ctx, requests)
	}
}

func BenchmarkEchobeatsEngineFullCycle(b *testing.B) {
	engine := NewEchobeatsEngine()
	config := DefaultEngineConfig()
	engine.Initialize("/tmp/test_model.gguf", config)
	defer engine.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ExecuteFullCycle(ctx, "Benchmark cycle")
	}
}
