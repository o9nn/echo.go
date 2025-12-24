# Deep Tree Echo - Sys6 Payload Specification

**Date:** 2025-12-24

**Version:** 1.0

## 1. Overview

This document details the structure of the concrete payloads that flow through the **sys6 cognitive architecture**. These payloads are the fundamental data structures that carry cognitive content, enabling the system to perceive, analyze, plan, act, and integrate knowledge. The design supports two primary forms of cognitive representation:

1.  **Cognitive Tokens:** Discrete, semantic units of information, such as percepts, thoughts, or goals.
2.  **Graph Messages:** Relational structures that represent connections, dependencies, and complex relationships between cognitive elements.

All payloads are wrapped in a `PayloadEnvelope` that manages their journey through the sys6 pipeline.

## 2. Payload Envelope

The `PayloadEnvelope` is the unified container for all data moving through the sys6 processor. It provides metadata for routing, prioritization, and state tracking.

| Field | Type | Description |
| :--- | :--- | :--- |
| `ID` | `string` | Unique identifier for the envelope. |
| `CreatedAt` | `time.Time` | Timestamp of creation. |
| `PayloadType` | `PayloadType` | The type of content: `token`, `graph`, `token_batch`, `graph_batch`. |
| `Token` / `Graph` | `*CognitiveToken` / `*GraphMessage` | The actual payload content (mutually exclusive). |
| `Priority` | `int` | Routing priority (higher is more urgent). |
| `Route` | `[]string` | A record of the components the payload has visited (e.g., `stage:analysis`, `c8:complete`). |
| `EntryStep` | `int` | The `Clock30` step when the payload entered the pipeline. |
| `CurrentStep` | `int` | The `Clock30` step the payload is currently being processed in. |
| `Errors` | `[]PayloadError` | A list of any errors encountered during processing. |

## 3. Cognitive Token (`CognitiveToken`)

A `CognitiveToken` is the atomic unit of thought or perception. It is designed to be processed in parallel by the C₈ and K₉ components to extract multi-faceted meaning.

### 3.1. Core Structure

| Field | Type | Description |
| :--- | :--- | :--- |
| `ID` | `string` | Unique identifier for the token. |
| `Content` | `TokenContent` | The core semantic content (text, data, source). |
| `ContentType` | `TokenType` | The cognitive category (e.g., `percept`, `thought`, `goal`). |
| `Salience` | `float64` | How attention-grabbing the token is (0.0-1.0). |
| `Relevance` | `float64` | How relevant the token is to current goals (0.0-1.0). |
| `Valence` | `float64` | The emotional tone of the content (-1.0 to 1.0). |
| `PipelineState` | `TokenPipelineState` | Tracks the token's position and status within the sys6 pipeline. |
| `Transformations` | `[]TokenTransformation` | A log of every processing step applied to the token. |

### 3.2. Pipeline Processing Results

As a token flows through the pipeline, it accumulates results from each major component:

-   **`C8Results` (`[8]*C8TokenResult`):** An array storing the output from each of the 8 parallel cognitive perspectives in the **Cubic Concurrency (C₈)** module. Each result contains the LLM-generated analysis from a unique angle (e.g., `Perception-Action-Learning`).

-   **`K9Results` (`[9]*K9TokenResult`):** An array storing the output from each of the 9 orthogonal phases in the **Triadic Convolution (K₉)** module. Each result contains the LLM-generated analysis from a specific temporal-scope lens (e.g., `Past-Universal`, `Present-Particular`).

-   **`PhiResult` (`*PhiTokenResult`):** The final integrated synthesis produced by the **Delay Fold (φ)**, which combines the C₈ and K₉ results into a single, coherent understanding.

### 3.3. Data Flow through Sys6

1.  **Ingestion (σ - Perception):** A new `CognitiveToken` is created, typically as a `percept` or `query`, and enters the pipeline via the `Sys6PayloadProcessor`.
2.  **Cubic Concurrency (C₈):** The token is broadcast to all 8 C₈ workers. Each worker processes the token in parallel, generating an analysis from its unique cognitive perspective. The 8 results are stored in `C8Results`.
3.  **Triadic Convolution (K₉):** The token is then processed by the 9 K₉ phases. Each phase analyzes the token from its temporal-scope perspective, and the 9 results are stored in `K9Results`.
4.  **Delay Fold (φ):** The `PhiCognitiveProcessor` takes the original token, the `C8Results`, and the `K9Results`. It synthesizes all 17 streams of analysis into a single, integrated output, which is stored in `PhiResult`.
5.  **Integration (σ - Integration):** The fully processed token, now enriched with multi-faceted analysis and an integrated summary, is passed to the Integration stage. Here, its insights can be used to update the agent's knowledge base, memories, or goals.

## 4. Graph Message (`GraphMessage`)

A `GraphMessage` represents complex relational knowledge. It is used for tasks that involve understanding systems, causality, and hierarchical structures.

### 4.1. Core Structure

| Field | Type | Description |
| :--- | :--- | :--- |
| `ID` | `string` | Unique identifier for the graph message. |
| `MessageType` | `GraphMessageType` | The type of relational structure (e.g., `causal`, `semantic`, `goal_tree`). |
| `Nodes` | `[]GraphNode` | The set of nodes in the graph (concepts, entities, events). |
| `Edges` | `[]GraphEdge` | The set of directed, weighted edges connecting the nodes. |
| `RootNodeID` | `string` | The primary entry point for graph traversal and processing. |
| `Coherence` | `float64` | A metric for the internal consistency of the graph's relationships. |
| `Complexity` | `float64` | A metric for the structural complexity of the graph. |

### 4.2. Data Flow through Sys6

The flow of a `GraphMessage` is analogous to that of a `CognitiveToken`, but the processing at each stage focuses on structural and relational properties rather than discrete semantic content.

1.  **Ingestion (σ):** A `GraphMessage` is created to represent a complex idea or system.
2.  **Cubic Concurrency (C₈):** Each of the 8 perspectives analyzes the graph. For example, the `Action` perspective might identify executable paths, while the `Reflection` perspective might identify cycles or inconsistencies. The analyses are stored in `C8GraphResults`.
3.  **Triadic Convolution (K₉):** Each of the 9 phases extracts or transforms the graph. The `Past-Universal` phase might identify historical patterns, while the `Future-Relational` phase might predict future changes in relationships. These results are stored in `K9GraphResults`.
4.  **Delay Fold (φ):** The `PhiCognitiveProcessor` synthesizes the various graph analyses and transformations, recommending a set of modifications (new nodes, edges, weight adjustments) to create a more complete and accurate relational model.
5.  **Integration (σ):** The transformed graph is used to update the agent's long-term knowledge graph.

## 5. Conclusion

This dual-payload system allows Deep Tree Echo to process both discrete thoughts and complex relational structures through the same powerful, multi-faceted sys6 pipeline. The detailed metadata and processing history stored within each payload provide rich data for introspection, learning, and the cultivation of wisdom.
