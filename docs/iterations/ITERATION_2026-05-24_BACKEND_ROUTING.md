# Deep Tree Echo Iteration — Substrate-Aware Backend Routing

**Date:** 2026-05-24  
**Repository:** `o9nn/echo.go`  
**Focus:** Provider-manager routing policy, host/model capability probing, backend capability events, and migration planning away from the legacy `core/inference/llama` surface.  
**Status:** Complete and validated locally.

## Executive Summary

This iteration converts the backend capability work from a **diagnostic-only surface** into an operational scheduling influence. The provider manager now consumes `backendcap.Decision` values and dynamically reorders registered local, remote, and fallback providers before generation. This means Echo’s inference routing can now follow the substrate selected by backend capability evaluation rather than merely reporting it in diagnostics.

The patch also introduces a lightweight host-memory probe and GGUF model metadata reader. Local model files can now be represented as capabilities that include model path, context length, quantization, and estimated memory footprint. This gives Echo a more concrete vocabulary for future local-model selection, memory-safe wake planning, and Echodream substrate adaptation.

## Key Outcomes

| Area | Change | Practical Gain |
|---|---|---|
| Provider routing | Added `ProviderManager.ApplyBackendDecision(decision backendcap.Decision)` and route ordering by backend class. | Local/native, remote/API, and fallback providers are now prioritized according to backend capability decisions. |
| Continuity fallback | Registered `SimpleFallback` in evolution-system provider wiring and adapter factory routing. | Echo can remain wakeful in a degraded continuity mode even when no API provider is configured. |
| Host memory probing | Added `ProbeHostMemory()` and `HostMemoryTier()` based on `/proc/meminfo`. | Scheduler decisions can distinguish constrained, standard, and stress-grade memory envelopes. |
| Model-file capabilities | Added GGUF metadata probing and `DiscoverModelCapabilities(paths)` / `SnapshotWithModelPaths(paths)`. | Specific local GGUF models can be surfaced with context length, quantization, estimated memory, and availability. |
| Cognitive event visibility | Added `EventBackendCapabilityChanged` and emitted routing/capability payloads through the existing cognitive event bus. | Wake/rest and dream systems can react to substrate changes through the same event stream as other cognitive state changes. |
| Test coverage | Added focused routing tests, host/model capability tests, and updated continuity fallback behavior tests. | The patch is covered under both cgo and no-cgo test modes. |

## Implementation Notes

The provider manager now treats `backendcap.Decision` as an active routing input. When the selected backend is native, registered local providers such as `local_gguf` are promoted ahead of remote API providers. When the selected backend is a remote/API class, API providers remain ahead of fallback. When the decision selects degraded fallback, `SimpleFallback` becomes the first provider. The existing fallback-chain semantics remain intact: explicit provider requests still start with the requested provider, and the fallback chain is preserved as a secondary route.

The evolution-system initialization path now refreshes backend routing after registering providers. The generation path refreshes routing again before calling the provider manager, which allows future capability changes, such as model discovery or host-memory changes, to influence generation without requiring a process restart. Backend capability changes emit a cognitive event containing the decision, capability snapshot, host-memory probe, selected backend, degradation flag, and provider route.

## Files Changed

| File | Purpose |
|---|---|
| `core/backendcap/capabilities.go` | Extended capability metadata fields and made GGML memory tier host-aware. |
| `core/backendcap/host_model.go` | Added host-memory probing, GGUF metadata parsing, model capability discovery, and snapshot enrichment. |
| `core/backendcap/host_model_test.go` | Added synthetic GGUF tests for model metadata and capability discovery. |
| `core/llm/provider.go` | Added backend-decision routing policy, fallback-chain inspection, and provider class scoring. |
| `core/llm/provider_routing_test.go` | Added tests for native, remote, and fallback route ordering. |
| `core/deeptreeecho/evolution_integration.go` | Wired routing refresh, fallback continuity, event bus access, and backend capability event emission. |
| `core/deeptreeecho/cognitive_event_bus.go` | Added `backend_capability_changed` event type. |
| `core/deeptreeecho/llm_provider_adapter.go` | Applied backend routing in the provider manager factory and included fallback continuity. |
| `core/deeptreeecho/evolution_integration_test.go` | Updated provider-absence behavior to assert degraded fallback continuity. |

## Validation Results

| Validation Command | Result |
|---|---|
| `go test ./core/backendcap ./core/llm` | Passed. |
| `go test ./core/deeptreeecho` | Passed. |
| `go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed. |
| `CGO_ENABLED=0 go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed. |
| `go test ./...` | Completed with all reported packages passing in the terminal output. |

## Llama Migration Plan

This iteration deliberately avoids expanding the legacy `core/inference/llama` dependency surface. The capability registry still reports `llama_legacy`, but its guidance now remains to prefer the maintained source-based `./llama` binding. The next migration should proceed in small, reversible steps so Echo does not lose operational continuity while local inference is modernized.

| Step | Action | Expected Gain |
|---|---|---|
| 1 | Introduce a narrow adapter interface over the maintained `./llama` package for model loading, context configuration, token generation, and streaming. | Creates a stable seam without forcing all callers to import low-level bindings. |
| 2 | Implement a concrete `local_gguf` provider backed by the maintained `./llama` adapter and the new model capability metadata. | Gives provider routing a real native target instead of only a stub or diagnostic capability. |
| 3 | Add a model registry that scans configured model directories and exposes model capabilities to backend selection. | Allows Echo to choose among specific local models by memory footprint, context length, and quantization. |
| 4 | Deprecate direct `core/inference/llama` usage in favor of the adapter while retaining tests for compatibility. | Reduces ABI drift and maintenance risk while preserving fallback behavior. |
| 5 | Emit model-selection events into the cognitive stream and let Echodream observe substrate changes. | Connects inference substrate changes to wake/rest and knowledge consolidation cycles. |

## Recommended Next Iteration

The next iteration should convert the newly discovered model-file capabilities into a first-class local GGUF provider backed by the maintained `./llama` package. That will close the loop between **capability discovery**, **provider routing**, and **actual native inference execution**. The most direct implementation path is to add a small `LocalGGUFProvider` that consumes selected model metadata, estimates whether the host can safely load the model, and exposes `Available()`, `MaxTokens()`, `Generate()`, and `StreamGenerate()` through the existing `LLMProvider` interface.

A secondary improvement should add an environment or config field for model directories, such as `ECHO_MODEL_PATHS` or `EvolutionSystemConfig.ModelPaths`, so `backendcap.SnapshotWithModelPaths(paths)` becomes part of the normal backend decision loop. This will allow Echo to become substrate-aware at the level of concrete local models rather than only backend classes.
