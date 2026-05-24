package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/backendcap"
)

const (
	// ModelLifecycleLoaded is emitted when a selected local model becomes resident.
	ModelLifecycleLoaded = "model_loaded"
	// ModelLifecycleUnloaded is emitted when a resident local model is released.
	ModelLifecycleUnloaded = "model_unloaded"
	// ModelLifecycleLoadFailed is emitted when a selected local model cannot be loaded.
	ModelLifecycleLoadFailed = "model_load_failed"
	// ModelLifecyclePolicyScored is emitted when the registry evaluates a concrete model candidate.
	ModelLifecyclePolicyScored = "model_policy_scored"
)

// ModelSelectionTask describes Echo's intent when selecting among concrete GGUF files.
type ModelSelectionTask struct {
	Intent                string
	RequiredContextTokens int
}

// ModelScoringPolicy scores a concrete GGUF capability for a specific Echo task.
// Higher scores are preferred. It operates below backend-class selection, where
// backendcap.SelectFromCapabilities has already established that concrete model
// files are schedulable native substrate candidates.
type ModelScoringPolicy func(cap backendcap.Capability, task ModelSelectionTask) int

// LocalModelEvent describes a state transition in the local model runtime registry.
type LocalModelEvent struct {
	Type                  string                     `json:"type"`
	ModelName             string                     `json:"model_name,omitempty"`
	ModelPath             string                     `json:"model_path,omitempty"`
	Capability            backendcap.Capability      `json:"capability"`
	EstimatedMemoryBytes  uint64                     `json:"estimated_memory_bytes,omitempty"`
	Loaded                bool                       `json:"loaded"`
	Reason                string                     `json:"reason,omitempty"`
	Error                 string                     `json:"error,omitempty"`
	PolicyScore           int                        `json:"policy_score,omitempty"`
	PolicyIntent          string                     `json:"policy_intent,omitempty"`
	RequiredContextTokens int                        `json:"required_context_tokens,omitempty"`
	Timestamp             time.Time                  `json:"timestamp"`
	HostMemory            backendcap.HostMemoryProbe `json:"host_memory"`
}

// LocalModelRegistryOptions configures persistent local GGUF model lifecycle management.
type LocalModelRegistryOptions struct {
	ModelPaths        []string
	ProviderName      string
	MemorySafetyRatio float64
	IdleUnloadAfter   time.Duration
	ScoringPolicy     ModelScoringPolicy
	SelectionTask     ModelSelectionTask
	OnEvent           func(LocalModelEvent)
	Now               func() time.Time
}

// LocalModelRegistry owns local GGUF model discovery, selection, lazy load state,
// and safe unload/reload policy. It keeps Echo's native substrate state distinct
// from one-shot provider construction so autonomous loops can observe and govern
// local model residency over time.
type LocalModelRegistry struct {
	mu sync.Mutex

	options  LocalModelRegistryOptions
	models   []backendcap.Capability
	selected backendcap.Capability
	provider *localModelRegistryProvider

	loaded       bool
	loadErr      error
	lastUsed     time.Time
	lastLoaded   time.Time
	lastUnloaded time.Time
	unloadReason string
}

// LocalModelRegistryState is a copyable status surface for diagnostics, Echodream,
// and wake/rest policy.
type LocalModelRegistryState struct {
	ModelPaths           []string                   `json:"model_paths"`
	DiscoveredModels     []backendcap.Capability    `json:"discovered_models"`
	SelectedModel        backendcap.Capability      `json:"selected_model"`
	ProviderName         string                     `json:"provider_name"`
	Loaded               bool                       `json:"loaded"`
	LoadError            string                     `json:"load_error,omitempty"`
	EstimatedMemoryBytes uint64                     `json:"estimated_memory_bytes,omitempty"`
	LastUsed             time.Time                  `json:"last_used,omitempty"`
	LastLoaded           time.Time                  `json:"last_loaded,omitempty"`
	LastUnloaded         time.Time                  `json:"last_unloaded,omitempty"`
	UnloadReason         string                     `json:"unload_reason,omitempty"`
	HostMemory           backendcap.HostMemoryProbe `json:"host_memory"`
	MemorySafe           bool                       `json:"memory_safe"`
	RuntimeReady         bool                       `json:"runtime_ready"`
}

// NewLocalModelRegistry creates a registry and performs an initial discovery pass.
func NewLocalModelRegistry(options LocalModelRegistryOptions) *LocalModelRegistry {
	if options.ProviderName == "" {
		options.ProviderName = "local_gguf"
	}
	if options.MemorySafetyRatio <= 0 || options.MemorySafetyRatio > 1 {
		options.MemorySafetyRatio = 0.85
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	registry := &LocalModelRegistry{options: options}
	registry.Refresh()
	return registry
}

// Refresh reprobes configured model paths and rebuilds the selected model if the
// backend-capability decision changed.
func (r *LocalModelRegistry) Refresh() LocalModelRegistryState {
	r.mu.Lock()
	defer r.mu.Unlock()

	models := backendcap.DiscoverModelCapabilities(r.options.ModelPaths)
	selected := r.selectModelLocked(models)
	changed := selected.ModelPath != r.selected.ModelPath
	if changed && r.provider != nil {
		r.unloadLocked("selected model changed")
	}
	r.models = models
	r.selected = selected
	if selected.ModelPath != "" && (r.provider == nil || changed) {
		r.provider = &localModelRegistryProvider{registry: r, provider: NewLocalGGUFProviderFromCapability(selected)}
	}
	if selected.ModelPath == "" {
		r.provider = nil
		r.loaded = false
		r.loadErr = nil
	}
	return r.stateLocked()
}

func (r *LocalModelRegistry) selectModelLocked(models []backendcap.Capability) backendcap.Capability {
	if r.options.ScoringPolicy != nil {
		var selected backendcap.Capability
		bestScore := 0
		hasSelection := false
		for _, model := range models {
			if model.ModelPath == "" || !model.Available {
				continue
			}
			score := r.options.ScoringPolicy(model, r.options.SelectionTask)
			r.emitPolicyScoreLocked(model, score)
			if !hasSelection || score > bestScore {
				selected = model
				bestScore = score
				hasSelection = true
			}
		}
		if hasSelection {
			return selected
		}
	}

	decision := backendcap.SelectFromCapabilities(backendcap.Workload{
		NeedOffline:   false,
		PreferNative:  true,
		MinMemoryTier: backendcap.MemoryConstrained,
	}, models)

	if decision.Selected.ModelPath != "" {
		return decision.Selected
	}
	return backendcap.Capability{}
}

// Provider returns the registry-managed provider candidate, if a model is selected.
func (r *LocalModelRegistry) Provider() LLMProvider {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.selected.ModelPath == "" {
		return nil
	}
	if r.provider == nil {
		r.provider = &localModelRegistryProvider{registry: r, provider: NewLocalGGUFProviderFromCapability(r.selected)}
	}
	return r.provider
}

// State returns a snapshot of registry lifecycle state.
func (r *LocalModelRegistry) State() LocalModelRegistryState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stateLocked()
}

// Warmup eagerly loads the selected local model through the registry-owned provider.
// It is intended for wake transitions and explicit diagnostics; normal generation
// still preserves lazy loading through the provider path.
func (r *LocalModelRegistry) Warmup(ctx context.Context) error {
	r.mu.Lock()
	if r.selected.ModelPath == "" {
		r.mu.Unlock()
		return fmt.Errorf("no local GGUF model selected")
	}
	if r.provider == nil {
		r.provider = &localModelRegistryProvider{registry: r, provider: NewLocalGGUFProviderFromCapability(r.selected)}
	}
	provider := r.provider.provider
	r.mu.Unlock()

	if ctx != nil {
		select {
		case <-ctx.Done():
			r.recordUse(provider, ctx.Err())
			return ctx.Err()
		default:
		}
	}
	err := provider.loadModelForRegistryWarmup()
	r.recordUse(provider, err)
	return err
}

// Cooldown explicitly releases a resident model, using memory-pressure language
// for wake/rest policy consumers while still allowing deliberate rest transitions.
func (r *LocalModelRegistry) Cooldown(reason string) bool {
	if reason == "" {
		reason = "runtime cooldown requested"
	}
	if r.MaybeUnloadForMemoryPressure(reason) {
		return true
	}
	return r.Unload(reason)
}

// RuntimeReadiness reports whether the selected model is resident and memory-safe.
func (r *LocalModelRegistry) RuntimeReadiness() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.stateLocked()
	return state.RuntimeReady
}

// MaybeUnloadForMemoryPressure releases the selected model if the current host
// memory probe suggests its estimated footprint no longer fits safely.
func (r *LocalModelRegistry) MaybeUnloadForMemoryPressure(reason string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.loaded || r.provider == nil || r.selected.EstimatedMemoryBytes == 0 {
		return false
	}
	host := backendcap.ProbeHostMemory()
	limit := uint64(float64(host.AvailableBytes) * r.options.MemorySafetyRatio)
	if host.AvailableBytes > 0 && r.selected.EstimatedMemoryBytes > limit {
		if reason == "" {
			reason = fmt.Sprintf("estimated footprint %.2f GiB exceeds %.0f%% of available memory", bytesToGiBLocal(r.selected.EstimatedMemoryBytes), r.options.MemorySafetyRatio*100)
		}
		r.unloadLocked(reason)
		return true
	}
	return false
}

// UnloadIdle releases the model when it has not been used for the configured idle interval.
func (r *LocalModelRegistry) UnloadIdle(reason string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.loaded || r.provider == nil || r.options.IdleUnloadAfter <= 0 || r.lastUsed.IsZero() {
		return false
	}
	if r.options.Now().Sub(r.lastUsed) >= r.options.IdleUnloadAfter {
		if reason == "" {
			reason = "idle unload policy"
		}
		r.unloadLocked(reason)
		return true
	}
	return false
}

// Unload releases the model explicitly.
func (r *LocalModelRegistry) Unload(reason string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.unloadLocked(reason)
}

func (r *LocalModelRegistry) recordUse(provider *LocalGGUFProvider, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.options.Now()
	r.lastUsed = now
	if err != nil {
		r.loadErr = err
		r.loaded = false
		r.emitLocked(ModelLifecycleLoadFailed, err, "generation attempted lazy load")
		return
	}
	loadedNow := provider != nil && provider.Loaded()
	if loadedNow && !r.loaded {
		r.loaded = true
		r.loadErr = nil
		r.lastLoaded = now
		r.emitLocked(ModelLifecycleLoaded, nil, "lazy load completed")
	} else if loadedNow {
		r.loaded = true
		r.loadErr = nil
	}
}

func (r *LocalModelRegistry) unloadLocked(reason string) bool {
	if r.provider == nil || !r.loaded {
		return false
	}
	if reason == "" {
		reason = "explicit unload"
	}
	_ = r.provider.provider.Close()
	r.loaded = false
	r.lastUnloaded = r.options.Now()
	r.unloadReason = reason
	r.emitLocked(ModelLifecycleUnloaded, nil, reason)
	return true
}

func (r *LocalModelRegistry) stateLocked() LocalModelRegistryState {
	models := append([]backendcap.Capability{}, r.models...)
	state := LocalModelRegistryState{
		ModelPaths:           append([]string{}, r.options.ModelPaths...),
		DiscoveredModels:     models,
		SelectedModel:        r.selected,
		ProviderName:         r.options.ProviderName,
		Loaded:               r.loaded,
		EstimatedMemoryBytes: r.selected.EstimatedMemoryBytes,
		LastUsed:             r.lastUsed,
		LastLoaded:           r.lastLoaded,
		LastUnloaded:         r.lastUnloaded,
		UnloadReason:         r.unloadReason,
		HostMemory:           backendcap.ProbeHostMemory(),
	}
	if r.provider != nil && r.provider.provider != nil {
		state.Loaded = r.provider.provider.Loaded()
		if err := r.provider.provider.LoadError(); err != nil {
			state.LoadError = err.Error()
		}
	}
	if r.loadErr != nil {
		state.LoadError = r.loadErr.Error()
	}
	state.MemorySafe = r.memorySafeLocked(state.HostMemory)
	state.RuntimeReady = state.SelectedModel.ModelPath != "" && state.Loaded && state.MemorySafe && state.LoadError == ""
	return state
}

func (r *LocalModelRegistry) memorySafeLocked(host backendcap.HostMemoryProbe) bool {
	if r.selected.ModelPath == "" || r.selected.EstimatedMemoryBytes == 0 || host.AvailableBytes == 0 {
		return true
	}
	return r.selected.EstimatedMemoryBytes <= uint64(float64(host.AvailableBytes)*r.options.MemorySafetyRatio)
}

func (r *LocalModelRegistry) emitPolicyScoreLocked(model backendcap.Capability, score int) {
	if r.options.OnEvent == nil {
		return
	}
	event := LocalModelEvent{
		Type:                  ModelLifecyclePolicyScored,
		ModelName:             strings.TrimPrefix(model.Name, "model:"),
		ModelPath:             model.ModelPath,
		Capability:            model,
		EstimatedMemoryBytes:  model.EstimatedMemoryBytes,
		Loaded:                r.loaded && model.ModelPath == r.selected.ModelPath,
		Reason:                "registry scoring policy evaluated concrete GGUF candidate",
		PolicyScore:           score,
		PolicyIntent:          r.options.SelectionTask.Intent,
		RequiredContextTokens: r.options.SelectionTask.RequiredContextTokens,
		Timestamp:             r.options.Now(),
		HostMemory:            backendcap.ProbeHostMemory(),
	}
	r.options.OnEvent(event)
}

func (r *LocalModelRegistry) emitLocked(eventType string, err error, reason string) {
	if r.options.OnEvent == nil {
		return
	}
	event := LocalModelEvent{
		Type:                 eventType,
		ModelName:            strings.TrimPrefix(r.selected.Name, "model:"),
		ModelPath:            r.selected.ModelPath,
		Capability:           r.selected,
		EstimatedMemoryBytes: r.selected.EstimatedMemoryBytes,
		Loaded:               r.loaded,
		Reason:               reason,
		Timestamp:            r.options.Now(),
		HostMemory:           backendcap.ProbeHostMemory(),
	}
	if err != nil {
		event.Error = err.Error()
	}
	r.options.OnEvent(event)
}

type localModelRegistryProvider struct {
	registry *LocalModelRegistry
	provider *LocalGGUFProvider
}

func (p *localModelRegistryProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	result, err := p.provider.Generate(ctx, prompt, opts)
	p.registry.recordUse(p.provider, err)
	return result, err
}

func (p *localModelRegistryProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	stream, err := p.provider.StreamGenerate(ctx, prompt, opts)
	if err != nil {
		p.registry.recordUse(p.provider, err)
		return stream, err
	}
	out := make(chan StreamChunk, 16)
	go func() {
		defer close(out)
		var streamErr error
		for chunk := range stream {
			if chunk.Error != nil {
				streamErr = chunk.Error
			}
			out <- chunk
		}
		p.registry.recordUse(p.provider, streamErr)
	}()
	return out, nil
}

func (p *localModelRegistryProvider) Name() string {
	return p.provider.Name()
}

func (p *localModelRegistryProvider) Available() bool {
	p.registry.MaybeUnloadForMemoryPressure("memory pressure before availability check")
	state := p.registry.State()
	return state.SelectedModel.ModelPath != "" && state.MemorySafe
}

func (p *localModelRegistryProvider) MaxTokens() int {
	return p.provider.MaxTokens()
}

func bytesToGiBLocal(bytes uint64) float64 {
	return float64(bytes) / float64(1024*1024*1024)
}
