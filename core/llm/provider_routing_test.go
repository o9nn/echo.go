package llm

import (
	"context"
	"testing"

	"github.com/o9nn/echo.go/core/backendcap"
)

type routingTestProvider struct {
	name      string
	available bool
}

func (p routingTestProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return p.name, nil
}

func (p routingTestProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	ch <- StreamChunk{Content: p.name, Done: true}
	close(ch)
	return ch, nil
}

func (p routingTestProvider) Name() string    { return p.name }
func (p routingTestProvider) Available() bool { return p.available }
func (p routingTestProvider) MaxTokens() int  { return 4096 }

func TestApplyBackendDecisionRoutesNativeBeforeRemote(t *testing.T) {
	pm := NewProviderManager()
	mustRegister(t, pm, routingTestProvider{name: "anthropic", available: true})
	mustRegister(t, pm, routingTestProvider{name: "local_gguf", available: true})
	mustRegister(t, pm, &SimpleFallbackProvider{})
	if err := pm.SetFallbackChain([]string{"anthropic", "local_gguf", "SimpleFallback"}); err != nil {
		t.Fatal(err)
	}

	route := pm.ApplyBackendDecision(backendcap.Decision{Selected: backendcap.Capability{Name: "ggml", Kind: backendcap.BackendNativeCPU, Native: true}})
	if got := route[0]; got != "local_gguf" {
		t.Fatalf("expected native provider first, got route %v", route)
	}

	result, err := pm.Generate(context.Background(), "hello", DefaultGenerateOptions())
	if err != nil {
		t.Fatal(err)
	}
	if result != "local_gguf" {
		t.Fatalf("expected generation to use local_gguf, got %q", result)
	}
}

func TestApplyBackendDecisionRoutesRemoteBeforeFallback(t *testing.T) {
	pm := NewProviderManager()
	mustRegister(t, pm, routingTestProvider{name: "openai", available: true})
	mustRegister(t, pm, &SimpleFallbackProvider{})
	if err := pm.SetFallbackChain([]string{"SimpleFallback", "openai"}); err != nil {
		t.Fatal(err)
	}

	route := pm.ApplyBackendDecision(backendcap.Decision{Selected: backendcap.Capability{Name: "remote_llm_provider", Kind: backendcap.BackendRemoteAPI}})
	if got := route[0]; got != "openai" {
		t.Fatalf("expected remote provider first, got route %v", route)
	}
}

func TestApplyBackendDecisionRoutesFallbackWhenSelected(t *testing.T) {
	pm := NewProviderManager()
	mustRegister(t, pm, routingTestProvider{name: "openai", available: true})
	mustRegister(t, pm, &SimpleFallbackProvider{})
	if err := pm.SetFallbackChain([]string{"openai", "SimpleFallback"}); err != nil {
		t.Fatal(err)
	}

	route := pm.ApplyBackendDecision(backendcap.Decision{Selected: backendcap.Capability{Name: "simple_fallback", Kind: backendcap.BackendFallback}, Degraded: true})
	if got := route[0]; got != "SimpleFallback" {
		t.Fatalf("expected fallback provider first, got route %v", route)
	}
}

func mustRegister(t *testing.T, pm *ProviderManager, provider LLMProvider) {
	t.Helper()
	if err := pm.RegisterProvider(provider); err != nil {
		t.Fatal(err)
	}
}
