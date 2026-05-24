# Evolution Iteration: Backend Capability-Aware Autonomy

**Date:** 2026-05-24  
**Repository:** `o9nn/echo.go`  
**Author:** Manus AI  
**Focus:** Residual backend autonomy after native backend boundary stabilization

## Summary

This iteration turns backend availability from an implicit build-time hazard into an explicit **runtime scheduling signal**. The previous backend-boundary patch made `core/inference/llama` and `ml/backend/ggml` safe across constrained and no-cgo environments. This follow-up extends that work upward into Echo’s orchestration layer so wakeful cognitive loops can inspect backend capacity and degrade gracefully rather than assuming native inference is always present.

The central architectural result is a new `core/backendcap` package that exposes a schedulable capability snapshot and workload selector. Higher-level Deep Tree Echo components now surface backend decisions in local model provider state, Echobeats scheduler metrics, and EvolutionSystem status. This gives future persistent loops a stable question to ask before choosing an inference substrate: **which backend is available, offline-capable, native, stress-grade, and suitable for the current memory tier?**

## Legacy llama wrapper decision

The legacy `core/inference/llama` wrapper remains retired from default scheduling. It is retained behind the explicit `llama_legacy` build tag and represented in the capability registry as `llama_legacy`, but it is not treated as the preferred local inference path.

| Option | Decision | Rationale |
|---|---|---|
| Fully modernize `core/inference/llama` | Deferred | The wrapper is not the maintained source-based binding path and would require reproducible static-library generation plus ABI tracking. |
| Delete the wrapper immediately | Deferred | Keeping a tagged compatibility boundary is safer than removing potentially useful legacy API surface without a broader migration pass. |
| Retire from default scheduling | Adopted | The default runtime can now prefer available maintained/local/native or remote/fallback providers without hard-failing on legacy native artifacts. |

This approach preserves historical code while preventing it from being treated as a mandatory substrate for Echo’s active autonomy loop.

## Implementation changes

The new backend capability registry models each inference substrate as a `Capability` with availability, substrate kind, native/offline flags, stress-grade status, memory tier, build tags, reason, and guidance. It also introduces a `Workload` request and a `Decision` result so scheduler code can choose the best available backend for a cognitive act.

| Area | Change | Autonomy effect |
|---|---|---|
| `core/backendcap` | Added capability snapshot and backend selector. | Converts backend state into an explicit schedulable signal. |
| GGML integration | Reuses `ggml.Available()` and stress opt-in semantics. | Lets Echo distinguish CI-safe native checks from stress-grade tensor work. |
| Legacy llama integration | Imports the retired `core/inference/llama` availability signal as `llama_legacy`. | Keeps legacy compatibility visible but non-default. |
| Local model provider | Added backend snapshot and backend decision methods/stats. | Local inference boundaries can now disclose whether they are native, fallback, or degraded. |
| Echobeats scheduler | Added backend-aware rhythm state and metrics. | Scheduler pacing can degrade according to backend availability rather than assuming a substrate. |
| EvolutionSystem | Added backend capability snapshot, decision, and degraded flag to status. | System-level diagnostics expose inference substrate health to orchestration loops. |
| Tests | Added capability registry and Deep Tree Echo integration coverage. | Locks in cgo/no-cgo graceful behavior and scheduler-visible capability decisions. |

## Verification

The verification matrix confirmed that the new capability-aware surfaces behave correctly in both cgo and no-cgo builds. The checks also re-ran the previously stabilized backend boundaries for legacy llama and GGML.

| Verification command | Result |
|---|---|
| `CGO_ENABLED=1 go test ./core/backendcap ./core/deeptreeecho ...` | Passed |
| `CGO_ENABLED=0 go test ./core/backendcap ./core/deeptreeecho ...` | Passed |
| `CGO_ENABLED=0 go test ./core/inference/llama -run '^$'` | Passed; no test files |
| `CGO_ENABLED=1 go test ./core/inference/llama -run '^$'` | Passed; no test files |
| `CGO_ENABLED=1 go test ./ml/backend/ggml` | Passed |
| `CGO_ENABLED=0 go test ./ml/backend/ggml -run '^$'` | Passed; no test files |

The verification log for this iteration was saved at `/home/ubuntu/echo-backend-analysis/backend_autonomy_verify.log`.

## Architectural significance

This iteration strengthens Echo’s **wake/rest substrate awareness**. A future persistent Echobeats loop should not collapse simply because GGML, cgo, native llama, or a stress-grade memory tier is unavailable. It should instead inspect its available substrates, choose a suitable provider, downgrade if necessary, and preserve cognitive continuity.

> Backend availability is now a rhythm input rather than a catastrophic precondition.

This moves Echo closer to the intended autonomous pattern: a self-orchestrating cognitive system that can select the right substrate for the current act, avoid unsafe stress workloads in constrained environments, and make degraded-but-conscious fallback decisions when native inference is absent.

## Recommended next steps

The next evolution should connect the backend capability decision to actual inference routing priority inside the provider manager. The current patch exposes capability decisions through provider, scheduler, and system diagnostics; the next patch can use those decisions to reorder local/native/API/fallback providers dynamically.

| Priority | Next step | Expected gain |
|---:|---|---|
| 1 | Add provider-manager routing policy that consumes `backendcap.Decision`. | Actual scheduling, not only diagnostics, becomes substrate-aware. |
| 2 | Add memory-tier probing from host memory and model metadata. | Workload planning becomes safer and more precise. |
| 3 | Represent model files as capabilities with context length, quantization, and estimated memory footprint. | Echo can choose among specific local models, not only backend classes. |
| 4 | Emit backend capability events into the cognitive event stream. | Wake/rest and dream integration can react to substrate changes. |
| 5 | Plan a broader migration away from `core/inference/llama` toward the maintained `./llama` package. | Reduces long-term maintenance risk and ABI drift. |
