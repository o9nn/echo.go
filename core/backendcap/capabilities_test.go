package backendcap

import "testing"

func TestSnapshotIncludesContinuityFallbacks(t *testing.T) {
	caps := Snapshot()
	if len(caps) == 0 {
		t.Fatal("expected at least one backend capability")
	}

	seenRemote := false
	seenFallback := false
	for _, cap := range caps {
		switch cap.Name {
		case "remote_llm_provider":
			seenRemote = true
			if !cap.Available || cap.Offline {
				t.Fatalf("remote provider capability should be available but online-only: %+v", cap)
			}
		case "simple_fallback":
			seenFallback = true
			if !cap.Available || !cap.Offline {
				t.Fatalf("simple fallback should preserve offline cognitive continuity: %+v", cap)
			}
		}
	}
	if !seenRemote || !seenFallback {
		t.Fatalf("expected remote and simple fallback continuity capabilities, got %+v", caps)
	}
}

func TestSelectPreservesOfflineContinuity(t *testing.T) {
	decision := Select(Workload{NeedOffline: true, PreferNative: true, MinMemoryTier: MemoryConstrained})
	if decision.Selected.Name == "" {
		t.Fatalf("expected a selected capability: %+v", decision)
	}
	if !decision.Selected.Offline {
		t.Fatalf("offline workload selected non-offline backend: %+v", decision)
	}
	if !decision.Selected.Available && decision.Selected.Name != "no_backend_available" {
		t.Fatalf("selected unavailable backend unexpectedly: %+v", decision)
	}
}

func TestSelectReportsUnsatisfiedStressWorkload(t *testing.T) {
	t.Setenv("ECHO_GGML_STRESS", "")
	decision := Select(Workload{NeedOffline: true, PreferNative: true, RequireStress: true, MinMemoryTier: MemoryStress})
	if !decision.Degraded {
		t.Fatalf("unsatisfied stress workload should be degraded: %+v", decision)
	}
	if decision.Selected.Name != "no_backend_available" {
		t.Fatalf("expected no_backend_available for unstressed default environment, got %+v", decision)
	}
	if len(decision.Alternatives) == 0 {
		t.Fatalf("expected alternatives to explain why selection failed")
	}
}
