package deeptreeecho

import (
	"context"
	"testing"

	"github.com/o9nn/echo.go/core/backendcap"
	"github.com/o9nn/echo.go/core/llm"
)

func TestLocalModelProviderExposesBackendDecisionAndFallsBack(t *testing.T) {
	fallback := &llm.SimpleFallbackProvider{}
	provider := NewLocalModelProvider(LocalModelConfig{
		ModelPath:        "/tmp/nonexistent-echo-model.gguf",
		ModelName:        "nonexistent",
		FallbackProvider: fallback,
	})

	decision := provider.BackendDecision()
	if decision.Selected.Name == "" {
		t.Fatalf("expected backend decision to be recorded: %+v", decision)
	}
	if len(provider.BackendSnapshot()) == 0 {
		t.Fatalf("expected backend snapshot to be recorded")
	}
	if !provider.Available() {
		t.Fatalf("expected fallback provider to preserve availability")
	}

	response, err := provider.Generate(context.Background(), "hello", llm.GenerateOptions{MaxTokens: 8})
	if err != nil {
		t.Fatalf("expected fallback generation to succeed: %v", err)
	}
	if response == "" {
		t.Fatalf("expected non-empty fallback response")
	}
}

func TestEchobeatsSchedulerReportsBackendCapabilityState(t *testing.T) {
	scheduler := NewEchobeatsSchedulerV2(nil)
	decision := scheduler.BackendDecision()
	if decision.Selected.Name == "" {
		t.Fatalf("expected initial backend decision: %+v", decision)
	}

	scheduler.processBeat()
	updated := scheduler.BackendDecision()
	if updated.Selected.Name == "" {
		t.Fatalf("expected updated backend decision after beat: %+v", updated)
	}
	if len(scheduler.BackendSnapshot()) == 0 {
		t.Fatalf("expected backend snapshot after beat")
	}

	metrics := scheduler.GetMetrics()
	if _, ok := metrics["backend_decision"].(backendcap.Decision); !ok {
		t.Fatalf("expected backend decision in metrics, got %#v", metrics["backend_decision"])
	}
	if _, ok := metrics["backend_snapshot"].([]backendcap.Capability); !ok {
		t.Fatalf("expected backend snapshot in metrics, got %#v", metrics["backend_snapshot"])
	}
}

func TestEvolutionSystemStatusIncludesBackendCapabilityDecision(t *testing.T) {
	es := &EvolutionSystem{providerManager: llm.NewProviderManager(), initialized: true}
	decision := es.BackendDecision()
	if decision.Selected.Name == "" {
		t.Fatalf("expected evolution system backend decision: %+v", decision)
	}

	status := es.GetStatus()
	if _, ok := status["backend_decision"].(backendcap.Decision); !ok {
		t.Fatalf("expected backend decision in evolution status, got %#v", status["backend_decision"])
	}
	if _, ok := status["backend_capabilities"].([]backendcap.Capability); !ok {
		t.Fatalf("expected backend capabilities in evolution status, got %#v", status["backend_capabilities"])
	}
	if _, ok := status["backend_degraded"].(bool); !ok {
		t.Fatalf("expected backend degraded flag in evolution status, got %#v", status["backend_degraded"])
	}
}
