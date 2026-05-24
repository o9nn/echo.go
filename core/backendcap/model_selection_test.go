package backendcap

import (
	"path/filepath"
	"testing"
)

func TestSelectWithModelPathsPrefersConcreteGGUFModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeTinyGGUF(t, path)

	decision := SelectWithModelPaths(Workload{NeedOffline: true, PreferNative: true}, []string{dir})
	if !cgoEnabled {
		if decision.Selected.ModelPath != "" {
			t.Fatalf("did not expect no-cgo build to select a native model, got %+v", decision.Selected)
		}
		return
	}
	if decision.Selected.ModelPath != path {
		t.Fatalf("expected concrete model capability %s, got %+v", path, decision.Selected)
	}
	if decision.Selected.ContextLength != 2048 || decision.Selected.Quantization != "Q4_K_M" {
		t.Fatalf("expected selected model metadata to be preserved, got %+v", decision.Selected)
	}
}

func TestSelectWithModelPathsRespectsRequiredTokens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeTinyGGUF(t, path)

	decision := SelectWithModelPaths(Workload{NeedOffline: true, PreferNative: true, RequiredTokens: 4096}, []string{dir})
	if decision.Selected.ModelPath == path {
		t.Fatalf("model context is too small for workload but was selected: %+v", decision.Selected)
	}
}
