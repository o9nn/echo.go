package llm

import (
	"context"
	"strings"
	"testing"
	"time"

	"path/filepath"
)

func TestLocalModelRegistryTracksSelectionAndState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeLLMTinyGGUF(t, path)

	registry := NewLocalModelRegistry(LocalModelRegistryOptions{ModelPaths: []string{dir}})
	state := registry.State()
	if state.SelectedModel.ModelPath != path {
		t.Fatalf("expected selected model %s, got %s", path, state.SelectedModel.ModelPath)
	}
	if len(state.DiscoveredModels) != 1 {
		t.Fatalf("expected one discovered model, got %d", len(state.DiscoveredModels))
	}
	if registry.Provider() == nil {
		t.Fatalf("expected registry-managed provider")
	}
	if state.Loaded {
		t.Fatalf("expected lazy registry to start unloaded")
	}
}

func TestLocalModelRegistryEmitsLoadFailedForInvalidTinyFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeLLMTinyGGUF(t, path)

	var events []LocalModelEvent
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths: []string{dir},
		OnEvent: func(event LocalModelEvent) {
			events = append(events, event)
		},
	})
	provider := registry.Provider()
	if provider == nil {
		t.Fatalf("expected provider")
	}
	_, _ = provider.Generate(context.Background(), "hello", GenerateOptions{MaxTokens: 1})
	if len(events) == 0 {
		t.Fatalf("expected lifecycle event after failed lazy load")
	}
	last := events[len(events)-1]
	if last.Type != ModelLifecycleLoadFailed {
		t.Fatalf("expected %s event, got %s", ModelLifecycleLoadFailed, last.Type)
	}
	if last.ModelPath != path {
		t.Fatalf("expected event model path %s, got %s", path, last.ModelPath)
	}
	if !strings.Contains(registry.State().LoadError, "failed") && registry.State().LoadError == "" {
		t.Fatalf("expected registry load error to be recorded")
	}
}

func TestLocalModelRegistryIdleUnloadPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeLLMTinyGGUF(t, path)

	now := time.Unix(100, 0)
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths:      []string{dir},
		IdleUnloadAfter: time.Second,
		Now:             func() time.Time { return now },
	})
	registry.mu.Lock()
	registry.loaded = true
	registry.lastUsed = now.Add(-2 * time.Second)
	registry.mu.Unlock()
	if !registry.UnloadIdle("test idle") {
		t.Fatalf("expected idle unload")
	}
	state := registry.State()
	if state.Loaded {
		t.Fatalf("expected unloaded state")
	}
	if state.UnloadReason != "test idle" {
		t.Fatalf("expected unload reason, got %q", state.UnloadReason)
	}
}
