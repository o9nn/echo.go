# Iteration Report: First-Class Local GGUF Provider

**Date:** 2026-05-24  
**Repository:** `o9nn/echo.go`  
**Focus:** Concrete local GGUF provider, maintained `./llama` package routing, model-path configuration, and model-aware backend decisions.  
**Status:** Complete and validated locally.

## Executive Summary

This iteration closes the loop between **model-file capability discovery**, **substrate-aware routing**, and **native inference execution**. The previous iteration made concrete GGUF files visible as backend capabilities and allowed provider routing to respond to `backendcap.Decision`; this iteration adds an actual `local_gguf` provider backed by the maintained in-repository `./llama` package. Echo can now discover a model file, select it as a concrete backend capability when the host appears safe, register a local provider from that model capability, and route generation through the same provider-manager path used by remote API providers and continuity fallback.

The patch also adds `EvolutionSystemConfig.ModelPaths` and environment-driven model path discovery through `ECHO_MODEL_PATHS` and `LOCAL_MODEL_PATH`. This makes `backendcap.SnapshotWithModelPaths(paths)` part of the normal evolution-system decision loop, status surface, and backend capability event payload. As a result, Echo is now substrate-aware at the level of **specific local model files**, not merely at the broader level of native, API, and fallback backend classes.

## Key Outcomes

| Area | Change | Practical Gain |
|---|---|---|
| Local inference provider | Added a cgo-enabled `core/llm.LocalGGUFProvider` backed by the maintained `./llama` package. | Provider routing now has a real native GGUF execution target rather than only diagnostics or a compile stub. |
| No-cgo continuity | Replaced the previous local-GGUF stub with an interface-correct no-cgo / `nollama` stub. | CI and lightweight builds continue to compile and degrade safely when native inference is unavailable. |
| Model-aware selection | Added `backendcap.SelectFromCapabilities` and `backendcap.SelectWithModelPaths`. | Routing can select concrete model-file capabilities enriched with context length, quantization, memory estimate, and path. |
| Model path configuration | Added `EvolutionSystemConfig.ModelPaths`, `ECHO_MODEL_PATHS`, and `LOCAL_MODEL_PATH` discovery. | Echo can scan configured files or directories for local models during normal startup and status evaluation. |
| Provider registration | EvolutionSystem now registers `local_gguf` from the selected concrete model capability when safe. | The provider manager can prioritize native local inference before remote APIs when capability policy selects it. |
| Resource cleanup | Added `Context.Free()` to the maintained `./llama` wrapper and `Close()` on the provider. | Native llama contexts can be released explicitly when the provider lifecycle ends. |
| Event/status visibility | Status and backend capability events now include model paths and model-enriched snapshots. | Echodream, wake/rest policies, and diagnostics can observe concrete substrate changes. |

## Implementation Notes

The real `LocalGGUFProvider` is build-tagged for `cgo && !nollama`, because the maintained `./llama` package only exposes the native binding API when cgo is available. The no-cgo provider remains compile-safe and implements the same constructor and `LLMProvider` surface, but it returns unavailable status and explicit generation errors instead of pretending to provide local inference. This preserves the repository’s existing no-cgo test contract while allowing the native build to grow real functionality.

The provider intentionally keeps `Available()` lightweight. It checks that a model path exists and that the `backendcap` memory and GGUF metadata probe considers the model safe, but it does not eagerly allocate native model or context memory during provider registration or diagnostic polling. The model is loaded lazily on first generation, which prevents status checks and backend capability events from consuming large amounts of memory.

Generation uses the maintained `./llama` package directly. The provider tokenizes the prompt, clears the KV cache for a fresh generation turn, decodes the prompt into a llama batch, samples tokens through `SamplingContext`, streams token pieces through the existing `llm.StreamChunk` contract, and applies stop-sequence trimming. The implementation is deliberately narrow: it gives Echo a working local substrate without expanding the older `core/inference/llama` surface.

## Files Changed

| File | Purpose |
|---|---|
| `core/llm/local_gguf_provider.go` | Added the first-class cgo local GGUF provider backed by the maintained `./llama` binding. |
| `core/llm/local_gguf_provider_stub.go` | Added an interface-correct no-cgo / `nollama` stub for safe degraded builds. |
| `core/llm/local_gguf_provider_test.go` | Added provider-constructor contract coverage shared by cgo and no-cgo modes. |
| `core/backendcap/capabilities.go` | Added model-path-aware selection helpers and scoring for concrete model capabilities. |
| `core/backendcap/host_model.go` | Marks discovered model capabilities unavailable in no-cgo builds while preserving their metadata. |
| `core/backendcap/model_selection_test.go` | Added tests for concrete model selection and context-length filtering. |
| `core/deeptreeecho/evolution_integration.go` | Added model-path configuration, local provider registration, model-aware decisions, and model path status/event payloads. |
| `core/deeptreeecho/evolution_integration_test.go` | Added model capability status coverage and updated default provider expectations. |
| `llama/llama.go` | Added `Context.Free()` to support explicit provider cleanup. |

## Validation Results

| Validation Command | Result |
|---|---|
| `go test ./core/backendcap ./core/llm` | Passed. |
| `go test ./core/deeptreeecho` | Passed. |
| `CGO_ENABLED=0 go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed. |
| `go test ./...` | Completed with all reported packages passing in the terminal output. |
| `git diff --check` | Passed with no whitespace errors. |

## Capability Flow After This Iteration

| Stage | Before | After |
|---|---|---|
| Discovery | GGUF files could be surfaced as diagnostic capabilities. | GGUF files are discovered from `EvolutionSystemConfig.ModelPaths`, `ECHO_MODEL_PATHS`, and `LOCAL_MODEL_PATH`. |
| Selection | Backend policy could choose native classes but had no concrete local provider target. | `backendcap.SelectWithModelPaths` can choose an actual model-file capability when safe and suitable. |
| Registration | `local_gguf` existed only as a stub or archived implementation. | EvolutionSystem registers a real `local_gguf` provider from the selected model capability. |
| Routing | Provider manager could reorder abstract provider classes. | Provider manager can prioritize a concrete native GGUF provider ahead of API providers. |
| Execution | Native local generation was not part of the normal provider path. | `Generate()` and `StreamGenerate()` can execute through the maintained `./llama` package in cgo builds. |
| Degradation | No-cgo behavior needed to remain safe. | No-cgo builds still compile and report native model capability metadata as unavailable rather than failing. |

## Recommended Next Iteration

The next evolution should connect local model selection to **persistent model lifecycle management**. The current provider lazily loads one selected GGUF model and can release it through `Close()`, but Echo still lacks a long-lived local model registry that can cache loaded models, evict them under memory pressure, and publish lifecycle events such as `model_loaded`, `model_unloaded`, and `model_load_failed` into the cognitive event stream.

A strong follow-up would add a small `LocalModelRegistry` or `ModelRuntimeManager` that owns model discovery, selection, lazy load state, and memory pressure decisions. This manager should expose the currently selected model, loaded state, load error, estimated memory footprint, and last-use timestamp. EvolutionSystem can then ask the registry for the best provider candidate instead of constructing the provider directly from each backend decision.

| Priority | Next Step | Expected Gain |
|---|---|---|
| 1 | Add a `LocalModelRegistry` that tracks discovered, selected, loaded, and failed local models. | Echo gains a persistent native substrate memory instead of one-shot provider construction. |
| 2 | Emit model lifecycle events into the cognitive event bus. | Wake/rest and Echodream can respond to native model availability, load failures, and substrate changes. |
| 3 | Add safe model unload and reload policy based on host memory probe and recent usage. | Autonomous loops can rest or shed native resources without process restart. |
| 4 | Add an integration test with a tiny valid GGUF fixture or gated real-model smoke test. | The generation path can be verified beyond constructor and routing contracts when a model fixture is available. |
| 5 | Continue reducing direct dependence on the legacy `core/inference/llama` surface. | The maintained `./llama` path becomes the durable native inference substrate for Echo’s growth. |

## Notes for Future Evolution

This patch deliberately avoids downloading or bundling a model file. Model paths are configuration-driven so repository tests remain lightweight and no-cgo-safe. For an actual autonomous deployment, set `ECHO_MODEL_PATHS` to one or more directories containing GGUF files, or set `LOCAL_MODEL_PATH` to a single preferred model. A practical next deployment test would use a small local GGUF model first, then graduate to a larger Echo-specific or NanEcho-specific model once the model registry and memory-pressure loop are in place.
