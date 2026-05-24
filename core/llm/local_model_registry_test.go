package llm

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/o9nn/echo.go/core/backendcap"
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

func TestLocalModelRegistryPolicyScoringSelectsBestCandidate(t *testing.T) {
	dir := t.TempDir()
	small := filepath.Join(dir, "small.gguf")
	large := filepath.Join(dir, "large.gguf")
	writeLLMTinyGGUFWithContext(t, small, 1024)
	writeLLMTinyGGUFWithContext(t, large, 4096)

	var scores []LocalModelEvent
	registry := NewLocalModelRegistry(LocalModelRegistryOptions{
		ModelPaths: []string{dir},
		SelectionTask: ModelSelectionTask{
			Intent:                "wakeful_skill_practice",
			RequiredContextTokens: 3000,
		},
		ScoringPolicy: func(cap backendcap.Capability, task ModelSelectionTask) int {
			score := cap.ContextLength
			if task.RequiredContextTokens > 0 && cap.ContextLength >= task.RequiredContextTokens {
				score += 10000
			}
			return score
		},
		OnEvent: func(event LocalModelEvent) {
			if event.Type == ModelLifecyclePolicyScored {
				scores = append(scores, event)
			}
		},
	})
	state := registry.State()
	if state.SelectedModel.ModelPath != large {
		t.Fatalf("expected higher-context model %s, got %s", large, state.SelectedModel.ModelPath)
	}
	if len(scores) != 2 {
		t.Fatalf("expected two policy score events, got %d", len(scores))
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

func TestLocalModelRegistryWarmupEmitsLoadedOrFailed(t *testing.T) {
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
	_ = registry.Warmup(context.Background())
	found := false
	for _, event := range events {
		if event.Type == ModelLifecycleLoaded || event.Type == ModelLifecycleLoadFailed {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected warmup to emit loaded or load-failed lifecycle event, got %#v", events)
	}
}

func TestLocalModelRegistryRuntimeReadiness(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeLLMTinyGGUF(t, path)

	registry := NewLocalModelRegistry(LocalModelRegistryOptions{ModelPaths: []string{dir}})
	if registry.RuntimeReadiness() {
		t.Fatalf("expected lazy unloaded registry to start not runtime-ready")
	}
	state := registry.State()
	if state.RuntimeReady {
		t.Fatalf("expected state runtime_ready to be false before warmup")
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

func writeLLMTinyGGUFWithContext(t *testing.T, path string, contextLength uint32) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.Write([]byte("GGUF"))
	binary.Write(file, binary.LittleEndian, uint32(3))
	binary.Write(file, binary.LittleEndian, uint64(0))
	binary.Write(file, binary.LittleEndian, uint64(4))
	writeLLMKVString(t, file, "general.name", filepath.Base(path))
	writeLLMKVString(t, file, "general.architecture", "llama")
	writeLLMKVUint32(t, file, "llama.context_length", contextLength)
	writeLLMKVUint32(t, file, "general.file_type", 15)
}
