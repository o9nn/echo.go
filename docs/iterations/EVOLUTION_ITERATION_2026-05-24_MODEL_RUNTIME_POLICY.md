# Evolution Iteration: Model Runtime Policy and GGUF Smoke Path

**Date:** 2026-05-24  
**Repository:** `o9nn/echo.go`  
**Author:** Manus AI  
**Focus:** Deployment-grade local model smoke testing, registry model scoring, and wake/rest runtime-readiness hooks

## Summary

This iteration connects the local model registry to a more deployment-realistic native inference boundary without bundling or downloading a model artifact. Echo now has a gated real-model smoke test path through `ECHO_TEST_GGUF_MODEL=/path/to/model.gguf`, a registry policy seam for scoring local GGUF candidates, and EvolutionSystem-level warmup, cooldown, and readiness methods that can be consumed by wake/rest and Echodream scheduling loops.

The architectural shift is deliberately conservative. Lightweight GGUF metadata fixtures remain the default repository test path, so CI and no-cgo environments remain fast and portable. When a real tiny GGUF model is available in a deployment environment, the new gated smoke test verifies the actual constructor, capability probe, maintained `./llama` local provider, lifecycle, and one-token decode path.

> Local model residency is now represented as a schedulable runtime state rather than a hidden side effect of generation.

## Implementation changes

The registry previously selected a safe discovered model and lazily loaded it through the managed provider. This iteration preserves that behavior but adds a policy layer that can eventually express Echo task intent: skill practice, background dreaming, wakeful discussion, summarization, or other local cognitive work. Policy scoring is observable through model lifecycle events rather than being an opaque in-memory decision.

| Area | Change | Autonomy effect |
|---|---|---|
| `core/llm/local_model_registry.go` | Added `ModelSelectionTask`, `ModelScoringPolicy`, policy score events, `Warmup`, `Cooldown`, and `RuntimeReadiness`. | Echo can choose local models by task-relevant criteria and expose model residency as an explicit cognitive substrate state. |
| `core/llm/local_gguf_provider*.go` | Added a registry-only warmup helper in both cgo and no-cgo provider implementations. | Warmup can test the real native load path when supported while remaining compile-safe and failure-transparent in no-cgo builds. |
| `core/deeptreeecho/evolution_integration.go` | Added `WarmupLocalModel`, `CooldownLocalModel`, `LocalModelReady`, and runtime event emission. | Wake/rest, Echodream, and future persistent loops can prepare or release native model memory around state transitions. |
| `core/deeptreeecho/cognitive_event_bus.go` | Added model policy and runtime readiness/cooling event types. | Local model lifecycle changes can now enter the cognitive event stream as scheduling signals. |
| `core/llm/local_gguf_provider_test.go` | Added a gated `ECHO_TEST_GGUF_MODEL` one-token smoke test. | Deployment environments can verify real GGUF inference without forcing binary artifacts into the repository. |
| `core/llm/local_model_registry_test.go` | Added policy scoring, warmup, and readiness tests over metadata fixtures. | Registry semantics remain deterministic and portable while still testing real selection and lifecycle state. |
| `core/deeptreeecho/model_registry_integration_test.go` | Added EvolutionSystem runtime hook coverage. | System-level readiness and cooling events are locked into integration behavior. |
| `discover/gpu_test.go`, `llama/llama_test.go`, `model/model_test.go` | Added cgo build constraints to tests that reference cgo-only symbols. | Full no-cgo test runs can validate portable packages instead of failing on native-only test definitions. |

## Runtime policy surfaces

The new runtime methods form a small, stable seam for higher-order autonomy components. `WarmupLocalModel` asks the registry-owned provider to load the selected GGUF model before active cognition. `CooldownLocalModel` releases residency for rest, memory pressure, or explicit runtime cooling. `LocalModelReady` reports whether the selected model is loaded, memory-safe, and free of load errors.

| Method or event | Meaning | Intended consumer |
|---|---|---|
| `LocalModelRegistry.Warmup(ctx)` | Eagerly validates and loads the selected provider path. | Wake transition, diagnostics, deployment smoke, pre-thought preparation. |
| `LocalModelRegistry.Cooldown(reason)` | Releases loaded model state and records the reason. | Rest transition, memory-pressure handling, dream-mode substrate shedding. |
| `LocalModelRegistry.RuntimeReadiness()` | Returns whether local inference is selected, resident, memory-safe, and healthy. | Echobeats scheduling, stream-of-consciousness provider choice, status dashboards. |
| `EventModelPolicyScored` | Announces candidate model score, task intent, and context requirement. | Future Echodream model-selection learning and telemetry. |
| `EventModelRuntimeReady` | Announces successful warmup and readiness state. | Wake-loop orchestration and local thought activation. |
| `EventModelRuntimeCooling` | Announces cooldown and post-release readiness state. | Rest-loop orchestration and memory conservation. |

This is intentionally not a fake generation path. Synthetic fixtures still fail native decoding as expected, and the tests assert lifecycle behavior around that failure. Real generation is only asserted when a caller explicitly provides a genuine GGUF file through the gated environment variable.

## Verification

The implementation was verified in both normal and no-cgo modes. The first no-cgo full-suite run exposed unrelated cgo-only tests in `discover`, `llama`, and `model`; those were converted to cgo-only tests so the portable suite can exercise the repository without native symbol failures.

| Verification command | Result |
|---|---|
| `go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed |
| `CGO_ENABLED=0 go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed |
| `go test ./...` | Passed |
| `CGO_ENABLED=0 go test ./...` | Passed |
| `ECHO_TEST_GGUF_MODEL=/path/to/model.gguf go test ./core/llm -run TestLocalGGUFProviderGatedRealModelSmoke` | Added; skipped by default until a real GGUF is supplied |

## Architectural significance

This iteration strengthens Echo’s path toward persistent wake/rest autonomy. A future autonomous loop should not treat native model loading as a one-off request-time accident. It should be able to prepare a model before thought, choose a model based on cognitive task requirements, expose whether the substrate is ready, and shed memory during rest or dream phases.

The registry now acts as a small native substrate memory. It remembers discovered model candidates, selection policy metadata, load failures, unload reasons, memory safety, and runtime readiness. That memory is exactly the type of state that Echobeats can later use to schedule active cognition, Echodream can use to integrate knowledge during rest, and stream-of-consciousness loops can use to decide whether local inference is safe to enter.

## Recommended next steps

The next evolution should connect this policy and readiness seam into actual autonomous wake/rest orchestration. The code now has the surfaces needed to warm, cool, and score local models, but the high-level cognitive loops should begin consuming those signals directly.

| Priority | Next step | Expected gain |
|---:|---|---|
| 1 | Wire `WarmupLocalModel`, `CooldownLocalModel`, and `LocalModelReady` into the wake/rest callback bridge used by the unified cognitive loop. | Echo can prepare local inference when waking and release native memory when resting. |
| 2 | Provide a deployment fixture path for a tiny real GGUF model in CI or a gated local profile. | The one-token native smoke path can run regularly without repository-bundled model artifacts. |
| 3 | Replace ad-hoc scoring callbacks with a default Echo model policy that weighs context length, quantization, memory footprint, freshness, and task intent. | Echo can select among local models by cognitive purpose rather than choosing a single safe candidate. |
| 4 | Feed `EventModelPolicyScored`, `EventModelRuntimeReady`, and `EventModelRuntimeCooling` into Echodream summaries. | Rest/dream integration can learn which substrates supported or blocked cognition. |
| 5 | Add background warmup and cooldown cadence governed by host-memory probes and Echobeats interest rhythms. | Persistent autonomous loops can anticipate thought rather than reacting only after prompts arrive. |
