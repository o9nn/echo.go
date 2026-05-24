package deeptreeecho

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/o9nn/echo.go/core/llm"
)

func TestEvolutionSystemExposesLocalModelRegistryState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeEvolutionTinyGGUF(t, path)

	cfg := DefaultEvolutionSystemConfig()
	cfg.ModelPaths = []string{dir}
	cfg.PreferredProviders = []string{"local_gguf"}
	es, err := NewEvolutionSystem(cfg)
	if err != nil {
		t.Fatalf("new evolution system: %v", err)
	}
	status := es.GetStatus()
	state, ok := status["local_model_registry"].(llm.LocalModelRegistryState)
	if !ok {
		t.Fatalf("expected local model registry state in status, got %T", status["local_model_registry"])
	}
	if state.SelectedModel.ModelPath != path {
		t.Fatalf("expected selected model %s, got %s", path, state.SelectedModel.ModelPath)
	}
}

func TestEvolutionSystemPublishesModelLoadFailedEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeEvolutionTinyGGUF(t, path)
	bus := NewCognitiveEventBus(context.Background())
	var received []CognitiveEvent
	bus.Subscribe(EventModelLoadFailed, func(event CognitiveEvent) {
		received = append(received, event)
	})

	cfg := DefaultEvolutionSystemConfig()
	cfg.ModelPaths = []string{dir}
	cfg.EventBus = bus
	es, err := NewEvolutionSystem(cfg)
	if err != nil {
		t.Fatalf("new evolution system: %v", err)
	}
	_, _ = es.Generate(context.Background(), "hello", llm.GenerateOptions{MaxTokens: 1})
	if len(received) == 0 {
		t.Fatalf("expected model_load_failed event")
	}
	data, ok := received[len(received)-1].Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map data")
	}
	if got := data["model_path"]; got != path {
		t.Fatalf("expected model path %s, got %v", path, got)
	}
}

func TestEvolutionSystemLocalModelRuntimeHooks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeEvolutionTinyGGUF(t, path)

	bus := NewCognitiveEventBus(context.Background())
	var coolingEvents []CognitiveEvent
	bus.Subscribe(EventModelRuntimeCooling, func(event CognitiveEvent) {
		coolingEvents = append(coolingEvents, event)
	})

	cfg := DefaultEvolutionSystemConfig()
	cfg.ModelPaths = []string{dir}
	cfg.EventBus = bus
	es, err := NewEvolutionSystem(cfg)
	if err != nil {
		t.Fatalf("new evolution system: %v", err)
	}
	if es.LocalModelReady() {
		t.Fatalf("expected synthetic fixture to start not ready")
	}
	_ = es.WarmupLocalModel(context.Background())
	if es.LocalModelReady() {
		t.Fatalf("expected invalid tiny fixture to remain not ready after failed warmup")
	}
	_ = es.CooldownLocalModel("test rest transition")
	if len(coolingEvents) == 0 {
		t.Fatalf("expected runtime cooling event")
	}
}
