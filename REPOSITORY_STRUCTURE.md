# Echo9llama Repository Structure

**Version:** 2.0  
**Date:** December 19, 2025  
**Purpose:** Optimal cognitive grip, performance, and deep integration readiness

---

## Directory Architecture

```
echo9llama/
│
├── README.md                    # Project overview and quickstart
├── LICENSE                      # License file
├── go.mod / go.sum              # Go module definition
├── Makefile                     # Build automation
├── Dockerfile                   # Container build
│
├── cmd/                         # Entry points (executables)
│   ├── echoself/                # Main autonomous echoself entry
│   ├── echobridge/              # Bridge server entry
│   └── runner/                  # Model runner entry
│
├── core/                        # Core cognitive architecture (Go packages)
│   │
│   ├── cognitive/               # 12-Step Cognitive Loop (Echobeats)
│   │   ├── loop.go              # Main cognitive loop orchestration
│   │   ├── engines.go           # 3 concurrent inference engines
│   │   ├── steps.go             # 12-step processors
│   │   └── triads.go            # Triad synchronization {1,5,9} etc.
│   │
│   ├── consciousness/           # Stream of Consciousness
│   │   ├── stream.go            # Thought generation
│   │   ├── awareness.go         # Self-awareness mechanisms
│   │   └── attention.go         # Attention management
│   │
│   ├── dream/                   # Echodream Knowledge Integration
│   │   ├── cycle.go             # Dream cycle management
│   │   ├── consolidation.go     # Memory consolidation
│   │   └── wisdom.go            # Wisdom extraction
│   │
│   ├── goals/                   # Goal Orchestration
│   │   ├── orchestrator.go      # Goal management
│   │   ├── generator.go         # Goal generation from identity
│   │   └── pursuit.go           # Goal pursuit logic
│   │
│   ├── memory/                  # Memory Systems
│   │   ├── hypergraph.go        # Hypergraph knowledge store
│   │   ├── episodic.go          # Episodic memory
│   │   └── semantic.go          # Semantic memory
│   │
│   ├── identity/                # Identity Kernel
│   │   ├── kernel.go            # Core identity definition
│   │   └── interests.go         # Interest pattern system
│   │
│   ├── llm/                     # LLM Provider Abstraction
│   │   ├── provider.go          # Provider interface
│   │   ├── anthropic.go         # Anthropic Claude
│   │   ├── openrouter.go        # OpenRouter
│   │   └── openai.go            # OpenAI
│   │
│   ├── skills/                  # Skill Learning System
│   │   ├── learning.go          # Skill acquisition
│   │   └── practice.go          # Skill practice
│   │
│   ├── discussion/              # Discussion Autonomy
│   │   ├── manager.go           # Discussion management
│   │   └── engagement.go        # Engagement decisions
│   │
│   ├── persistence/             # State Persistence
│   │   ├── state.go             # State management
│   │   └── storage.go           # Storage backends
│   │
│   └── integration/             # System Integration
│       ├── opencog.go           # OpenCog integration
│       ├── scheme.go            # Scheme interpreter
│       └── hgql.go              # HyperGraphQL
│
├── server/                      # HTTP/gRPC Servers
│   ├── routes.go                # API routes
│   ├── handlers.go              # Request handlers
│   └── autonomous/              # Autonomous server mode
│
├── model/                       # Model Loading & Processing
│   ├── loader.go                # Model loading
│   └── models/                  # Model-specific implementations
│
├── ml/                          # Machine Learning Backend
│   └── backend/                 # GGML/GGUF backend
│
├── api/                         # API Definitions
│   └── examples/                # API usage examples
│
├── assets/                      # Static Assets
│   ├── live2d/                  # Live2D models
│   └── ui/                      # UI textures/resources
│
├── docs/                        # Documentation
│   ├── README.md                # Documentation index
│   ├── architecture/            # Architecture documentation
│   │   ├── cognitive-loop.md    # 12-step cognitive loop spec
│   │   ├── consciousness.md     # Consciousness architecture
│   │   └── integration.md       # Integration specifications
│   ├── guides/                  # User guides
│   │   ├── quickstart.md        # Getting started
│   │   ├── deployment.md        # Deployment guide
│   │   └── development.md       # Development guide
│   ├── api/                     # API documentation
│   │   └── reference.md         # API reference
│   └── iterations/              # Evolution history
│       └── ITERATION_*.md       # Iteration reports
│
├── archive/                     # Deprecated/Historical
│   ├── deprecated/              # Deprecated code
│   ├── iterations/              # Old iteration docs
│   └── experiments/             # Experimental code
│
├── scripts/                     # Build/Deploy Scripts
│   ├── build.sh                 # Build script
│   └── deploy.sh                # Deployment script
│
└── test/                        # Test Files
    ├── integration/             # Integration tests
    └── fixtures/                # Test fixtures
```

---

## Core Package Consolidation

### Current State → Target State

| Current Package | Target Package | Rationale |
|:----------------|:---------------|:----------|
| `core/echobeats/` | `core/cognitive/` | Clearer naming for cognitive loop |
| `core/deeptreeecho/` | Distributed | Split into focused packages |
| `core/consciousness/` | `core/consciousness/` | Keep as-is |
| `core/echodream/` | `core/dream/` | Shorter, clearer naming |
| `core/goals/` | `core/goals/` | Keep as-is |
| `core/memory/` | `core/memory/` | Keep as-is |
| `core/identity/` | `core/identity/` | Keep as-is |
| `core/llm/` | `core/llm/` | Keep as-is |
| `core/skills/` | `core/skills/` | Keep as-is |
| `core/hgql/` | `core/integration/` | Merge into integration |
| `core/opencog/` | `core/integration/` | Merge into integration |
| `core/scheme/` | `core/integration/` | Merge into integration |

### Files to Archive

| File/Pattern | Reason | Destination |
|:-------------|:-------|:------------|
| `*.disabled` | Disabled code | `archive/deprecated/` |
| `*.bak` | Backup files | `archive/deprecated/` |
| `*.old` | Old versions | `archive/deprecated/` |
| `test_*_bin` | Test binaries | Delete (regenerate) |
| `*.log` | Log files | Delete (regenerate) |
| Root `ITERATION_*.md` | Iteration docs | `docs/iterations/` |
| Root `EVOLUTION_*.md` | Evolution docs | `docs/iterations/` |
| Root `*_GUIDE.md` | Guide docs | `docs/guides/` |
| Root `*_STUDY.md` | Analysis docs | `docs/architecture/` |

---

## Root File Organization

### Keep in Root
- `README.md` - Project overview
- `LICENSE` - License
- `go.mod`, `go.sum` - Go modules
- `Makefile` - Build automation
- `Dockerfile`, `Dockerfile.autonomous` - Container builds
- `CONTRIBUTING.md` - Contribution guide
- `SECURITY.md` - Security policy
- `REPOSITORY_STRUCTURE.md` - This document

### Move to docs/
- All `ITERATION_*.md` → `docs/iterations/`
- All `EVOLUTION_*.md` → `docs/iterations/`
- All `*_GUIDE.md` → `docs/guides/`
- All `*_STUDY.md`, `*_ANALYSIS.md` → `docs/architecture/`
- `TODO.md` → `docs/TODO.md`

### Move to archive/
- Empty files (0 bytes)
- Duplicate/versioned files
- Deprecated documentation

### Delete
- `*.log` files
- `test_*_bin` binaries
- `*.json` analysis outputs
- `__pycache__` directories

---

## Package Import Paths

After restructuring, import paths will be:

```go
import (
    "github.com/cogpy/echo9llama/core/cognitive"
    "github.com/cogpy/echo9llama/core/consciousness"
    "github.com/cogpy/echo9llama/core/dream"
    "github.com/cogpy/echo9llama/core/goals"
    "github.com/cogpy/echo9llama/core/memory"
    "github.com/cogpy/echo9llama/core/identity"
    "github.com/cogpy/echo9llama/core/llm"
    "github.com/cogpy/echo9llama/core/skills"
    "github.com/cogpy/echo9llama/core/discussion"
    "github.com/cogpy/echo9llama/core/persistence"
    "github.com/cogpy/echo9llama/core/integration"
)
```

---

## Integration Points

### Cognitive Loop ↔ Consciousness
- Stream of consciousness feeds into cognitive loop
- Cognitive loop generates thoughts for consciousness

### Dream ↔ Memory
- Dream consolidates episodic memories
- Memory provides content for dream processing

### Goals ↔ Identity
- Identity kernel drives goal generation
- Goals reflect identity interests

### LLM ↔ All
- LLM provider used by all cognitive components
- Centralized provider management

---

## Migration Strategy

1. **Phase 1: Documentation** - Move all loose docs to proper locations
2. **Phase 2: Archive** - Move deprecated/disabled files to archive
3. **Phase 3: Cleanup** - Delete generated artifacts (logs, binaries)
4. **Phase 4: Assets** - Reorganize Content/ to assets/
5. **Phase 5: Core** - Consolidate core packages (future iteration)

---

*This structure optimizes for cognitive grip, clear function boundaries, and deep integration readiness.*
