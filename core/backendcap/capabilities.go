package backendcap

import (
	"os"
	"sort"
	"strings"

	legacyllama "github.com/o9nn/echo.go/core/inference/llama"
	"github.com/o9nn/echo.go/ml/backend/ggml"
)

// MemoryTier describes the amount of memory a backend path can safely use.
type MemoryTier string

const (
	MemoryConstrained MemoryTier = "constrained"
	MemoryStandard    MemoryTier = "standard"
	MemoryStress      MemoryTier = "stress"
)

// BackendKind describes the substrate class used by an inference backend.
type BackendKind string

const (
	BackendNativeCPU BackendKind = "native_cpu"
	BackendNativeGPU BackendKind = "native_gpu"
	BackendRemoteAPI BackendKind = "remote_api"
	BackendFallback  BackendKind = "fallback"
)

// Capability describes a schedulable backend substrate.
type Capability struct {
	Name        string      `json:"name"`
	Kind        BackendKind `json:"kind"`
	Available   bool        `json:"available"`
	Native      bool        `json:"native"`
	Offline     bool        `json:"offline"`
	StressGrade bool        `json:"stress_grade"`
	MemoryTier  MemoryTier  `json:"memory_tier"`
	BuildTags   []string    `json:"build_tags,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	Guidance    string      `json:"guidance,omitempty"`
}

// Workload describes an inference task from the perspective of the scheduler.
type Workload struct {
	NeedOffline    bool
	PreferNative   bool
	RequireStress  bool
	MinMemoryTier  MemoryTier
	RequiredTokens int
}

// Decision captures the backend selected for a workload and why.
type Decision struct {
	Selected     Capability   `json:"selected"`
	Degraded     bool         `json:"degraded"`
	Reason       string       `json:"reason"`
	Alternatives []Capability `json:"alternatives,omitempty"`
}

// StressEnabled reports whether stress-grade native backend work is explicitly enabled.
func StressEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("ECHO_GGML_STRESS")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// Snapshot returns the current backend capability surface. It intentionally avoids
// importing heavy native llama bindings so capability checks stay cheap and usable
// in both cgo and no-cgo builds.
func Snapshot() []Capability {
	stress := StressEnabled()
	caps := []Capability{
		{
			Name:        "ggml",
			Kind:        BackendNativeCPU,
			Available:   ggml.Available(),
			Native:      true,
			Offline:     true,
			MemoryTier:  memoryTierForGGML(stress),
			StressGrade: stress && ggml.Available(),
			Reason:      availabilityReason("ggml", ggml.Available()),
			Guidance:    "Use for local tensor execution and CI-safe native checks; enable ECHO_GGML_STRESS=1 only on machines with sufficient memory.",
		},
		{
			Name:       "llama_legacy",
			Kind:       BackendNativeCPU,
			Available:  legacyllama.Available(),
			Native:     true,
			Offline:    true,
			MemoryTier: MemoryStandard,
			BuildTags:  []string{"llama_legacy"},
			Reason:     availabilityReason("legacy core/inference/llama", legacyllama.Available()),
			Guidance:   "Retired from default scheduling; prefer the maintained source-based ./llama binding unless a legacy consumer requires this API.",
		},
		{
			Name:       "remote_llm_provider",
			Kind:       BackendRemoteAPI,
			Available:  true,
			Native:     false,
			Offline:    false,
			MemoryTier: MemoryConstrained,
			Reason:     "remote/API providers are evaluated by their own provider Available() methods",
			Guidance:   "Use when local native backends are absent, memory is constrained, or external model quality is preferred.",
		},
		{
			Name:       "simple_fallback",
			Kind:       BackendFallback,
			Available:  true,
			Native:     false,
			Offline:    true,
			MemoryTier: MemoryConstrained,
			Reason:     "always available as a degraded cognitive continuity surface",
			Guidance:   "Use only to preserve wakeful loop continuity when no real inference substrate is available.",
		},
	}

	if !cgoEnabled {
		for i := range caps {
			if caps[i].Native {
				caps[i].Available = false
				caps[i].StressGrade = false
				caps[i].Reason = "native backend unavailable: cgo is disabled"
			}
		}
	}

	sort.SliceStable(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })
	return caps
}

// Select chooses the best backend capability for a workload while preserving a
// degraded option instead of forcing orchestration failure.
func Select(workload Workload) Decision {
	caps := Snapshot()
	candidates := make([]Capability, 0, len(caps))
	for _, cap := range caps {
		if !cap.Available {
			continue
		}
		if workload.NeedOffline && !cap.Offline {
			continue
		}
		if workload.RequireStress && !cap.StressGrade {
			continue
		}
		if !satisfiesMemoryTier(cap.MemoryTier, workload.MinMemoryTier) {
			continue
		}
		candidates = append(candidates, cap)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return score(candidates[i], workload) > score(candidates[j], workload)
	})

	if len(candidates) == 0 {
		fallback := Capability{
			Name:       "no_backend_available",
			Kind:       BackendFallback,
			Available:  false,
			MemoryTier: MemoryConstrained,
			Reason:     "no backend capability satisfies the requested workload constraints",
			Guidance:   "Relax offline/stress requirements, enable cgo/native backends, or configure an API provider.",
		}
		return Decision{Selected: fallback, Degraded: true, Reason: fallback.Reason, Alternatives: caps}
	}

	selected := candidates[0]
	degraded := selected.Kind == BackendFallback || (!selected.Native && workload.PreferNative) || (workload.NeedOffline && !selected.Offline)
	reason := "selected highest-scoring available backend for workload"
	if degraded {
		reason = "selected degraded backend to preserve cognitive loop continuity"
	}
	return Decision{Selected: selected, Degraded: degraded, Reason: reason, Alternatives: candidates[1:]}
}

func memoryTierForGGML(stress bool) MemoryTier {
	if stress {
		return MemoryStress
	}
	return MemoryStandard
}

func availabilityReason(name string, available bool) string {
	if available {
		return name + " compiled and available in this build"
	}
	return name + " not available in this build"
}

func satisfiesMemoryTier(actual, required MemoryTier) bool {
	return tierRank(actual) >= tierRank(required)
}

func tierRank(t MemoryTier) int {
	switch t {
	case MemoryStress:
		return 3
	case MemoryStandard:
		return 2
	case MemoryConstrained:
		return 1
	default:
		return 0
	}
}

func score(cap Capability, workload Workload) int {
	score := tierRank(cap.MemoryTier)
	if cap.Native && workload.PreferNative {
		score += 6
	}
	if cap.Offline && workload.NeedOffline {
		score += 5
	}
	if cap.StressGrade && workload.RequireStress {
		score += 4
	}
	switch cap.Kind {
	case BackendNativeCPU, BackendNativeGPU:
		score += 3
	case BackendRemoteAPI:
		score += 2
	case BackendFallback:
		score -= 10
	}
	return score
}
