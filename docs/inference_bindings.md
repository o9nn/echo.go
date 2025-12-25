# Deep Tree Echo Inference Bindings

This document describes the Go bindings for GGML and llama.cpp inference libraries, designed to power the 3 concurrent inference engines in the echobeats cognitive loop.

## Overview

The inference package provides:

1. **GGML Bindings** (`core/inference/ggml/`) - Low-level tensor operations
2. **Llama Bindings** (`core/inference/llama/`) - LLM inference API
3. **Echobeats Engine** (`core/inference/`) - 3-stream concurrent inference

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Echobeats Engine                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Stream    │  │   Stream    │  │   Stream    │             │
│  │   Alpha     │  │   Beta      │  │   Gamma     │             │
│  │  {1,5,9}    │  │  {2,6,10}   │  │  {3,7,11}   │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                     │
│         └────────────────┼────────────────┘                     │
│                          │                                      │
│                ┌─────────▼─────────┐                           │
│                │  Inference Engine │                           │
│                │   (llama.cpp)     │                           │
│                └─────────┬─────────┘                           │
│                          │                                      │
│                ┌─────────▼─────────┐                           │
│                │   GGML Backend    │                           │
│                │  (CPU / Vulkan)   │                           │
│                └───────────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

## GGML Package

### Types

| Type | Description |
|------|-------------|
| `Type` | Tensor data types (F32, F16, Q4_0, etc.) |
| `BackendType` | Backend types (CPU, GPU, GPU_SPLIT) |
| `Op` | Tensor operations (Add, Mul, MulMat, etc.) |
| `Context` | GGML computation context |
| `Tensor` | Multi-dimensional tensor |
| `Graph` | Computation graph |
| `Backend` | Compute backend |
| `Buffer` | Backend memory buffer |

### Key Functions

```go
// Create a new context
ctx, err := ggml.NewContext(ggml.InitParams{
    MemSize: 1024 * 1024 * 512, // 512MB
    NoAlloc: false,
})
defer ctx.Free()

// Create tensors
a := ctx.NewTensor2D(ggml.TypeF32, 768, 768)
b := ctx.NewTensor2D(ggml.TypeF32, 768, 768)

// Perform operations
c := ctx.MulMat(a, b)
c = ctx.RMSNorm(c, 1e-5)
c = ctx.Silu(c)

// Build computation graph
graph := ctx.NewGraph()
graph.BuildForward(c)
```

### Backend Loading

```go
// Load Vulkan backend
err := ggml.LoadBackend("/path/to/libggml-vulkan.so")

// Or load all available backends
n := ggml.LoadAllBackendsFromPath("/path/to/libs/")

// Initialize best available backend
backend, err := ggml.InitBestBackend()
```

## Llama Package

### Types

| Type | Description |
|------|-------------|
| `Model` | Loaded LLM model |
| `Context` | Inference context |
| `Vocab` | Model vocabulary |
| `Batch` | Token batch for processing |
| `Sampler` | Token sampling strategy |
| `SamplerChain` | Chain of samplers |

### Model Loading

```go
// Initialize backend
llama.BackendInit()
defer llama.BackendFree()

// Load model
params := llama.DefaultModelParams()
params.NGPULayers = 35 // Offload layers to GPU
params.UseMmap = true

model, err := llama.LoadModel("/path/to/model.gguf", params)
defer model.Free()

// Create context
ctxParams := llama.DefaultContextParams()
ctxParams.NCtx = 4096
ctxParams.NBatch = 512
ctxParams.FlashAttn = true

ctx, err := model.NewContext(ctxParams)
defer ctx.Free()
```

### Tokenization

```go
vocab := model.GetVocab()

// Tokenize text
tokens, err := vocab.Tokenize("Hello, world!", true, true)

// Detokenize
text := vocab.Detokenize(tokens, false, false)

// Get special tokens
bos := vocab.BOS()
eos := vocab.EOS()
```

### Inference

```go
// Create batch
batch := llama.NewBatch(512, 0, 1)
defer batch.Free()

// Add tokens to batch
for i, token := range tokens {
    batch.AddToken(token, int32(i), 0, i == len(tokens)-1)
}

// Decode
err := ctx.Decode(batch)

// Get logits
logits := ctx.GetLogits()
```

### Sampling

```go
// Create sampler chain
chain := llama.NewSamplerChain(false)
defer chain.Free()

// Add samplers
chain.Add(llama.NewSamplerTemp(0.7))
chain.Add(llama.NewSamplerTopK(40))
chain.Add(llama.NewSamplerTopP(0.9, 1))
chain.Add(llama.NewSamplerDist(42))

// Sample token
token := chain.Sample(ctx, -1)
chain.Accept(token)
```

## Echobeats Engine

### 12-Step Cognitive Loop

The echobeats engine implements a 12-step cognitive loop with 3 concurrent streams:

| Step | Type | Active Streams |
|------|------|----------------|
| 1 | Relevance Realization | Alpha |
| 2 | Affordance Interaction | Beta |
| 3 | Affordance Interaction | Gamma |
| 4 | Affordance Interaction | (sync) |
| 5 | Affordance Interaction | Alpha |
| 6 | Affordance Interaction | Beta |
| 7 | Relevance Realization | Gamma |
| 8 | Salience Simulation | (sync) |
| 9 | Salience Simulation | Alpha |
| 10 | Salience Simulation | Beta |
| 11 | Salience Simulation | Gamma |
| 12 | Salience Simulation | (sync) |

### Stream Triads

- **Alpha (Perception)**: Steps {1, 5, 9}
- **Beta (Action)**: Steps {2, 6, 10}
- **Gamma (Simulation)**: Steps {3, 7, 11}

### Usage

```go
// Create engine
engine := inference.NewEchobeatsEngine()

// Initialize with model
config := inference.DefaultEngineConfig()
config.ContextSize = 4096
config.Threads = 8
config.GPULayers = 35

err := engine.Initialize("/path/to/model.gguf", config)
defer engine.Close()

// Execute single step
responses, err := engine.ExecuteCognitiveStep(ctx, 1, "Initial prompt")

// Execute full 12-step cycle
results, err := engine.ExecuteFullCycle(ctx, "Start of cognitive cycle")

// Get metrics
metrics := engine.GetMetrics()
fmt.Printf("Total inferences: %d\n", metrics.TotalInferences)
fmt.Printf("Tokens generated: %d\n", metrics.TotalTokens)
```

### Concurrent Inference

```go
// Prepare requests for all 3 streams
requests := [3]*inference.InferenceRequest{
    {Step: 1, Prompt: "Alpha perception", MaxTokens: 256},
    {Step: 2, Prompt: "Beta action", MaxTokens: 256},
    {Step: 3, Prompt: "Gamma simulation", MaxTokens: 256},
}

// Execute concurrently
responses, err := engine.InferConcurrent(ctx, requests)
```

## Native Library Integration

### Directory Structure

```
echo.go/
├── core/
│   └── inference/
│       ├── ggml/
│       │   ├── ggml.h      # C header
│       │   └── ggml.go     # Go bindings
│       ├── llama/
│       │   ├── llama.h     # C header
│       │   └── llama.go    # Go bindings
│       ├── echobeats_engine.go
│       └── llama_engine.go
└── libs/
    └── arm64-v8a/
        ├── libggml.so
        ├── libggml-base.so
        ├── libggml-cpu.so
        ├── libggml-vulkan.so
        ├── libllama.so
        └── ...
```

### Build Configuration

```bash
# Set library path
export LD_LIBRARY_PATH=/path/to/libs/arm64-v8a:$LD_LIBRARY_PATH

# Build with CGO
CGO_ENABLED=1 go build -tags=cgo ./...

# Cross-compile for ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
    CC=aarch64-linux-gnu-gcc \
    go build -tags=cgo ./...
```

### CGO Flags

The bindings use the following CGO configuration:

```go
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR}/../../../libs -lggml-base -lggml -lggml-cpu -lm -lpthread
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../libs/arm64-v8a
```

## Backend Support

### CPU Backend

The CPU backend is always available and uses optimized SIMD instructions:

- ARM NEON (ARM64)
- AVX/AVX2/AVX512 (x86_64)

### Vulkan Backend

For GPU acceleration on Android/Linux:

```go
// Load Vulkan backend
err := ggml.LoadBackend("/path/to/libggml-vulkan.so")

// Initialize Vulkan engine
engine := inference.NewVulkanEngine(inference.StreamAlpha, 0)
engine.SetMemoryLimit(4 * 1024 * 1024 * 1024) // 4GB
```

### OpenCL Backend

Alternative GPU backend:

```go
err := ggml.LoadBackend("/path/to/libggml-opencl.so")
```

## Performance Considerations

1. **Context Reuse**: Reuse contexts across inferences to avoid allocation overhead
2. **Batch Processing**: Use larger batch sizes for prompt processing
3. **KV Cache**: Manage KV cache to enable efficient continuation
4. **Memory Mapping**: Enable mmap for faster model loading
5. **Flash Attention**: Enable for reduced memory usage on long contexts

## Error Handling

All functions return errors that should be checked:

```go
model, err := llama.LoadModel(path, params)
if err != nil {
    log.Fatalf("Failed to load model: %v", err)
}
```

## Thread Safety

- `Model` is thread-safe for read operations
- `Context` should be used from a single goroutine
- `EchobeatsEngine` handles synchronization internally

## Future Enhancements

1. **Quantization Support**: Dynamic quantization for memory efficiency
2. **Speculative Decoding**: Multi-model speculation for faster generation
3. **Continuous Batching**: Dynamic batch management for throughput
4. **State Persistence**: Save/restore cognitive state across sessions
5. **Distributed Inference**: Multi-node inference for large models
