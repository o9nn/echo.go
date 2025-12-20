# Real Integration Testing Framework

This document outlines the real integration testing framework for the echo9llama project, which validates the system's interaction with live external services.

## Overview

Unlike unit tests that use mocks, these integration tests connect to actual instances of Dgraph, Supabase, and LLM providers (Anthropic, OpenRouter) to ensure production readiness.

**To run these tests, you must have the following environment variables set:**

- `DGRAPH_ENDPOINT`: The address of your Dgraph Alpha instance (e.g., `localhost:9080`)
- `SUPABASE_URL`: Your Supabase project URL
- `SUPABASE_KEY`: Your Supabase service role key
- `ANTHROPIC_API_KEY`: Your Anthropic API key
- `OPENROUTER_API_KEY`: Your OpenRouter API key

## Test Suites

### Dgraph Integration (`dgraph_integration_test.go`)

This suite validates the full lifecycle of memory operations against a real Dgraph instance.

| Test Case | Description |
|:---|:---|
| `TestDgraphConnection` | Verifies that a connection can be established to the Dgraph endpoint. |
| `TestDgraphSchemaSetup` | Sets up the hypergraph schema defined in `core/memory/schema.dql`. |
| `TestDgraphNodeCRUD` | Performs Create, Read, Update, and Delete operations on individual memory nodes. |
| `TestDgraphEdgeOperations` | Creates two nodes and an edge between them, then traverses the edge. |
| `TestDgraphHypergraphMemory` | Tests the `DgraphHypergraph` abstraction for adding and retrieving nodes. |
| `TestDgraphBulkOperations` | Benchmarks bulk insertion and querying of 100 nodes. |

### Supabase Integration (`supabase_integration_test.go`)

This suite tests the persistence of various memory types to a real Supabase project.

| Test Case | Description |
|:---|:---|
| `TestSupabaseConnection` | Verifies a basic connection by querying a table. |
| `TestSupabaseMemoryNodeCRUD` | Full CRUD lifecycle for the `memory_nodes` table. |
| `TestSupabaseMemoryEdgeCRUD` | Full CRUD lifecycle for the `memory_edges` table. |
| `TestSupabaseEpisodeStorage` | Tests storage and retrieval from the `episodes` table. |
| `TestSupabaseIdentitySnapshot` | Tests storage and retrieval from the `identity_snapshots` table. |
| `TestSupabaseDreamJournal` | Tests storage and retrieval from the `dream_journals` table. |
| `TestSupabaseBulkOperations` | Benchmarks sequential insertion of 50 nodes. |
| `TestSupabaseRPC` | Placeholder for testing custom RPC functions. |

### LLM Provider Integration (`llm_integration_test.go`)

This suite validates the functionality of each LLM provider against their live APIs.

| Test Case | Description |
|:---|:---|
| `TestAnthropicProvider` | Tests simple generation, system prompts, long generation, and streaming with the Anthropic API. |
| `TestOpenRouterProvider` | Tests simple generation with multiple models available through OpenRouter. |
| `TestOpenAIProvider` | Tests simple generation with the OpenAI API. |
| `TestProviderManager` | Verifies that the provider manager can register and fall back between multiple live providers. |
| `TestCognitiveThoughtGeneration` | Simulates the cognitive loop by generating thoughts for Relevance, Affordance, Salience, and Meta-Reflection steps. |
| `TestStreamOfConsciousness` | Generates a continuous, multi-turn stream of consciousness to test coherent thought flow. |

## Running the Tests

To run the full integration test suite, use the following command:

```bash
export PATH=$PATH:/usr/local/go/bin
cd /home/ubuntu/echo9llama
go test -tags=integration -v -timeout 300s ./test/integration/...
```

Individual test suites or cases can be run using the `-run` flag.
