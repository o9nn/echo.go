# Echo9llama: Hybrid LLM Architecture with go-llama.cpp

**Date:** December 19, 2025  
**Author:** Manus AI  
**Version:** 1.0

---

## 1. Introduction

This document details the integration of `go-llama.cpp` into the `echo9llama` architecture, creating a powerful **hybrid LLM system**. This enhancement allows `echo9llama` to leverage both remote, high-capability models via `janecho-server` and local, efficient models via `llama.cpp` for direct inference.

This hybrid approach provides significant advantages:

- **Flexibility**: Choose the right tool for the job. Use large, remote models for complex reasoning and smaller, local models for speed, privacy, and offline capability.
- **Resilience**: If remote APIs are unavailable, `echo9llama` can fall back to its local model, ensuring continuous operation.
- **Cost-Effectiveness**: Offload routine tasks to a local model to reduce API costs.
- **Privacy**: Sensitive data can be processed entirely on-device using the local model.

## 2. Hybrid Architecture Overview

The `llm.ProviderManager` is the core of the hybrid architecture. It manages multiple registered LLM providers and orchestrates a fallback chain.

```
┌──────────────────────────────────────────────────┐
│           Autonomous Consciousness               │
└──────────────────────┬───────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────────┐
│              LLM Thought Engine                  │
└──────────────────────┬───────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────────┐
│               llm.ProviderManager                │
│                                                  │
│  Fallback Chain: ["janecho", "llamacpp"]         │
│                                                  │
│  ┌───────────────┐       ┌───────────────────┐   │
│  │  JanClient    │◀─────▶│  janecho-server   │   │
│  │ (Remote API)  │       │ (Anthropic, etc.) │   │
│  └───────────────┘       └───────────────────┘   │
│                                                  │
│  ┌───────────────┐       ┌───────────────────┐   │
│  │ LlamaCppProv. │◀─────▶│   llama.cpp       │   │
│  │ (Local CGo)   │       │ (GGUF Model File) │   │
│  └───────────────┘       └───────────────────┘   │
└──────────────────────────────────────────────────┘
```

- **`JanClient`**: Acts as the primary provider, connecting to the `janecho-server` for access to powerful, remote models.
- **`LlamaCppProvider`**: Acts as the secondary/fallback provider, enabling direct, local inference using a GGUF model file via `llama.cpp` CGo bindings.

## 3. `LlamaCppProvider` Implementation

- **File**: `core/llm/providers/llamacpp/provider.go`
- **Goal**: To provide a self-contained, local LLM provider that implements the `llm.LLMProvider` interface.

### 3.1. Key Features

- **Direct `llama.cpp` Integration**: Uses CGo to bind directly to the `llama.cpp` library within the `echo9llama` repository.
- **GGUF Model Support**: Loads any GGUF-formatted model for local inference.
- **Hardware Acceleration**: Automatically leverages GPU acceleration (Metal, CUDA, etc.) if available, with configurable GPU layer offloading.
- **Synchronous & Streaming Generation**: Implements both `Generate` and `StreamGenerate` methods of the `LLMProvider` interface.
- **Embedding Generation**: Includes a `GetEmbedding` method, allowing it to be used as an `EmbeddingProvider` for the memory system.
- **Configurable Parameters**: Supports runtime configuration of context size, batch size, thread count, GPU layers, flash attention, and sampling parameters (temperature, top-p, seed).

### 3.2. Configuration

The `LlamaCppProvider` is configured via the `llamacpp.Config` struct:

```go
type Config struct {
	Name           string `json:"name"`
	ModelPath      string `json:"model_path"`
	ContextSize    int    `json:"context_size"`
	BatchSize      int    `json:"batch_size"`
	Threads        int    `json:"threads"`
	GPULayers      int    `json:"gpu_layers"`
	FlashAttention bool   `json:"flash_attention"`
	KVCacheType    string `json:"kv_cache_type"`
	Seed           int    `json:"seed"`
}
```

## 4. Local Embedding with `LlamaCppEmbedder`

To complement the local inference provider, a local embedding provider has also been created.

- **File**: `core/memory/embeddings/llamacpp_embedder.go`
- **Goal**: To provide an `EmbeddingProvider` that uses a local GGUF embedding model.

### 4.1. Key Features

- **Implements `memory.EmbeddingProvider`**: Can be used directly by the `MilvusClient` or any other `CognitiveMemory` implementation.
- **Uses `LlamaCppProvider` Internally**: Wraps a `LlamaCppProvider` instance to handle model loading and embedding generation via the `GetEmbedding` method.
- **Dedicated Embedding Models**: Designed to work best with specialized GGUF embedding models (e.g., `nomic-embed-text`, `bge`).

## 5. Integration Workflow

To enable the hybrid architecture, both the `JanClient` and `LlamaCppProvider` are registered with the `ProviderManager`.

```go
// Path: core/autonomous/autonomous_consciousness.go (modification)
package autonomous

import (
	"github.com/EchoCog/echo9llama/core/llm"
	"github.com/EchoCog/echo9llama/core/llm/providers/jan"
	"github.com/EchoCog/echo9llama/core/llm/providers/llamacpp"
)

func NewAutonomousConsciousness(...) (*AutonomousConsciousness, error) {
	// --- LLM Provider Manager Setup ---
	llmManager := llm.NewProviderManager()

	// 1. Initialize JanClient (Remote Provider)
	janConfig := jan.Config{ Name: "janecho", URL: "http://localhost:8080" }
	janClient, err := jan.NewJanClient(janConfig)
	if err == nil && janClient.Available() {
		llmManager.RegisterProvider(janClient)
	}

	// 2. Initialize LlamaCppProvider (Local Provider)
	llamaConfig := llamacpp.DefaultConfig("/path/to/your/model.gguf")
	llamaProvider, err := llamacpp.NewLlamaCppProvider(llamaConfig)
	if err == nil && llamaProvider.Available() {
		llmManager.RegisterProvider(llamaProvider)
	}

	// 3. Set Fallback Chain: Try remote first, then local
	if err := llmManager.SetFallbackChain([]string{"janecho", "llamacpp"}); err != nil {
		// Handle error if no providers are available
	}

	// --- Memory System Setup with Local Embedder ---
	
	// Initialize LlamaCppEmbedder
	localEmbedder, err := embeddings.NewLlamaCppEmbedder("/path/to/embedding_model.gguf", 768)
	if err != nil {
		// Handle error
	}
	
	// Initialize MilvusClient with the local embedder
	milvusConfig := milvus.Config{ Address: "localhost:19530" }
	memoryBackend, err := milvus.NewMilvusClient(ctx, milvusConfig, localEmbedder)

	// ... rest of initialization
}
```

## 6. Building with `llama.cpp`

Since `llama.cpp` is a C++ library, `echo9llama` must be built with CGo enabled. The Go bindings in `llama/llama.go` include the necessary CGo directives.

```go
/*
#cgo CFLAGS: -std=c11
#cgo CXXFLAGS: -std=c++17
#cgo CPPFLAGS: -I${SRCDIR}/llama.cpp/include
// ... other flags
*/
import "C"
```

A `Makefile` or build script should handle the compilation of the C++ code and the Go program.

```makefile
# Example Makefile target
build:
	@echo "Building echo9llama with llama.cpp support..."
	@CGO_ENABLED=1 go build -tags "static,netgo" -o ./bin/echo9llama ./cmd/autonomous
```

## 7. Summary

By integrating `go-llama.cpp`, `echo9llama` gains a significant architectural enhancement. The hybrid LLM system provides a robust, flexible, and efficient foundation for autonomous operation.

- **`LlamaCppProvider`** delivers local, hardware-accelerated inference.
- **`LlamaCppEmbedder`** enables private, on-device embedding generation.
- The **`ProviderManager`** seamlessly orchestrates between remote and local models.

This architecture moves `echo9llama` closer to its goal of a fully autonomous AGI by providing it with both powerful external reasoning capabilities and a self-reliant internal thought generation and understanding and privacy-preserving local cognitive core.
