# Iteration Report: Persistent Local Model Registry

**Date:** 2026-05-24  
**Repository:** `o9nn/echo.go`  
**Focus:** Persistent local GGUF model lifecycle management, registry-managed provider ownership, memory-aware unload policy, lifecycle event publication, and no-cgo-safe degradation.  
**Status:** Complete and validated locally.

## Executive Summary

This iteration turns the previous first-class `local_gguf` provider into a **persistent native model substrate**. The earlier patch made concrete GGUF model files discoverable and routable; this patch adds a `LocalModelRegistry` that owns model discovery, selection, lazy-load state, load errors, last-use timestamps, estimated memory footprint, and safe unload policy. Echo no longer treats a local model provider as a one-shot construction detail inside `EvolutionSystem`; instead, the system now has a long-lived local model runtime manager that can be observed, queried, and governed by future wake/rest and Echodream policies.

The registry publishes model lifecycle transitions through the existing cognitive event bus as `model_loaded`, `model_unloaded`, and `model_load_failed`. These events are intentionally routed through the same event-stream infrastructure used by backend capability decisions so that later autonomous loops can react to native model availability, model load failures, and substrate pressure without inventing a parallel event mechanism.

## Key Outcomes

| Area | Change | Practical Gain |
|---|---|---|
| Persistent native substrate | Added `core/llm.LocalModelRegistry` as the long-lived owner of local model discovery, selection, provider wrapping, load state, and unload policy. | Echo gains a native model memory surface rather than reconstructing a provider directly from each backend decision. |
| Lifecycle event stream | Added lifecycle event constants and cognitive event-bus support for model load, unload, and load-failure transitions. | Echodream, wake/rest scheduling, and diagnostics can observe local model residency changes as cognitive events. |
| Registry-managed provider | EvolutionSystem now registers the provider returned by the registry instead of directly constructing a standalone `LocalGGUFProvider`. | Generation attempts flow through the registry wrapper, allowing lazy-load success or failure to update persistent state. |
| Memory pressure policy | Added `MaybeUnloadForMemoryPressure` and idle-unload hooks around backend routing refreshes. | Autonomous loops can eventually shed native resources without process restart when memory pressure or rest policy requires it. |
| Status visibility | `GetStatus()` now exposes `local_model_registry` state with discovered models, selected model, loaded state, load error, memory estimate, host memory, and timestamps. | Operators and future cognitive policies can inspect concrete substrate readiness at runtime. |
| No-cgo continuity | The registry keeps discovered model metadata visible even when native llama is unavailable and records load-failure events when generation attempts the registry-managed candidate. | Lightweight CI and no-cgo builds remain safe while still exercising lifecycle state and degraded routing behavior. |

## Implementation Notes

The registry is deliberately placed in `core/llm` rather than inside `core/deeptreeecho` because it is a provider-runtime concern. It accepts model paths, performs a backend capability discovery pass, selects the best concrete model candidate, and wraps the existing `LocalGGUFProvider` behind the normal `LLMProvider` interface. This keeps provider-manager routing unchanged while adding lifecycle awareness around the provider.

The registry tracks whether the selected model is currently resident, the last load error, the last time the provider was used, the last successful load timestamp, the last unload timestamp, and the reason for unload. It also exposes a copyable `LocalModelRegistryState` that can be serialized into system status and future cognitive memory surfaces. The current implementation keeps loading lazy: status refreshes and routing decisions still avoid eager native allocation.

Memory handling is intentionally conservative. The registry checks the selected model’s estimated footprint against a fresh host-memory probe using a configurable safety ratio, and it can unload a resident model when the footprint exceeds the safe threshold. An idle-unload hook is also present so a future sleep/rest loop can release local model memory after a configured quiet interval.

EvolutionSystem now refreshes the registry during backend routing, invokes memory and idle unload policy hooks, and routes generation through the registry-managed provider when a concrete model candidate is selected. In no-cgo builds, this permits a controlled attempt against the registry-managed candidate, records `model_load_failed`, and then falls back through the existing provider manager continuity chain.

## Files Changed

| File | Purpose |
|---|---|
| `core/llm/local_model_registry.go` | Added the persistent local model registry, lifecycle event model, provider wrapper, state surface, memory-pressure unload policy, and idle-unload hooks. |
| `core/llm/local_model_registry_test.go` | Added registry discovery, selection, state, provider wrapper, load-failure, and unload policy coverage. |
| `core/llm/local_gguf_provider.go` | Added provider lifecycle helpers such as loaded-state and load-error inspection for registry ownership. |
| `core/llm/local_gguf_provider_stub.go` | Added matching no-cgo lifecycle helper methods so registry code remains build-safe. |
| `core/deeptreeecho/cognitive_event_bus.go` | Added `EventModelLifecycle` to the existing cognitive event taxonomy. |
| `core/deeptreeecho/evolution_integration.go` | Wired registry creation, registry-managed provider registration, model lifecycle event publication, backend routing refresh policy, and registry status exposure. |
| `core/deeptreeecho/model_registry_integration_test.go` | Added EvolutionSystem integration coverage for registry status, routing, and lifecycle event publication. |
| `docs/iterations/ITERATION_2026-05-24_LOCAL_MODEL_REGISTRY.md` | Documents the iteration, validation results, and next evolution targets. |

## Validation Results

| Validation Command | Result |
|---|---|
| `go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed. |
| `CGO_ENABLED=0 go test ./core/backendcap ./core/llm ./core/deeptreeecho` | Passed. |
| `go test ./...` | Completed with all reported packages passing in the terminal output. |

## Capability Flow After This Iteration

| Stage | Before | After |
|---|---|---|
| Discovery | Model paths produced concrete GGUF capabilities during backend decisions. | The registry owns discovery and keeps discovered model metadata visible as persistent runtime state. |
| Selection | EvolutionSystem selected a concrete model capability and constructed a provider directly. | The registry selects the model candidate and exposes a registry-managed provider to EvolutionSystem. |
| Loading | `LocalGGUFProvider` lazily loaded a single model and only retained state inside itself. | Lazy-load success or failure is reflected in registry state and lifecycle events. |
| Routing | Provider routing could prioritize `local_gguf`, but the model lifecycle was not externally visible. | Routing through the registry-managed provider updates last-use, loaded-state, and load-error surfaces. |
| Resource release | The provider could be closed manually. | The registry can unload for explicit reasons, host-memory pressure, or future idle/rest policy. |
| Cognitive visibility | Backend capability events described substrate decisions. | Cognitive events now also describe model residency transitions and failures. |

## Recommended Next Iteration

The next evolution should connect the registry to an actual **deployment-grade native model smoke test** and then deepen the runtime policy loop. This iteration deliberately avoids downloading or bundling any model file, so tests still use lightweight GGUF metadata fixtures and no-cgo-safe failure paths. A practical follow-up should run against a tiny real GGUF model in a gated environment variable test, then graduate to an Echo-specific or NanEcho-specific local model once the smoke path is reliable.

| Priority | Next Step | Expected Gain |
|---|---|---|
| 1 | Add a gated real-model smoke test, for example `ECHO_TEST_GGUF_MODEL=/path/to/model.gguf`, that verifies one-token generation through the maintained `./llama` path. | The constructor, routing, lifecycle, and actual decode path become verifiable when a tiny model is available. |
| 2 | Add a model registry policy interface for scoring model candidates by context length, quantization, memory footprint, and Echo task type. | Echo can choose among multiple local models rather than simply selecting the first safe candidate. |
| 3 | Feed model lifecycle events into Echodream and wake/rest scheduling. | Echo can rest, shed native memory, or wake local inference based on substrate state. |
| 4 | Add a background model warmup and cool-down loop governed by registry state and host memory probe. | Persistent autonomous loops can prepare a model before active thought and release it during rest. |
| 5 | Continue shrinking direct dependence on legacy `core/inference/llama` surfaces. | The maintained `./llama` package becomes the durable native inference substrate for Echo’s growth. |

## Notes for Future Evolution

The registry is intentionally minimal but real. It does not fake model residency or deterministic generation, and it does not force repository tests to carry binary model artifacts. Its role is to give Echo a stable native substrate memory: a place where discovered local models can become selected, loaded, failed, unloaded, and eventually dream-governed resources. The next deployment-oriented pass should provide a tiny external GGUF fixture and allow the registry to prove not only that lifecycle state is coherent, but that native inference can complete through the same route used by autonomous thought.
