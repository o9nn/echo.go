# Echo9llama Integration Roadmap: Dgraph, Goakt, Cog

**Version:** 1.0  
**Date:** December 19, 2025  
**Purpose:** Detail the next steps for integrating Dgraph, Goakt, and Cog into the `core/` and `server/` directories of the echo9llama repository.

---

## 1. Dgraph Integration: Hypergraph Persistence

**Objective:** Replace the current in-memory hypergraph in `core/memory/hypergraph.go` with Dgraph to provide a persistent, scalable, and high-performance knowledge graph, enabling true wisdom accumulation.

### Target Packages & Files:

- **`core/memory/`**: The entire package will be refactored to use Dgraph as the backend.
- **`core/persistence/`**: A new `dgraph_client.go` will be created to manage the connection to Dgraph.
- **`core/hgql/`**: The HyperGraphQL engine will be updated to generate Dgraph GraphQL+- queries.
- **`server/autonomous/main.go`**: The Dgraph client will be initialized and injected into the memory system.
- **`docker-compose.yml`**: Dgraph services (Zero and Alpha) will be added for local development.

### Integration Steps:

1.  **Environment Setup:**
    - Add Dgraph Zero and Alpha services to the project's `docker-compose.yml`.
    - Define a new Dgraph schema in `core/memory/schema.dql` that maps the existing `MemoryNode`, `MemoryEdge`, and `HyperEdge` structs to Dgraph predicates.

2.  **Dgraph Client Implementation:**
    - Create `core/persistence/dgraph_client.go` to handle the connection to the Dgraph Alpha instance.
    - Implement functions for setting the schema and handling transactions.

3.  **Refactor `core/memory`:**
    - Create a new `dgraph_hypergraph.go` file with a `DgraphHypergraph` struct that implements the existing `Hypergraph` interface.
    - Rewrite the `AddNode`, `AddEdge`, `GetNode`, `Traverse` and other methods to use the Dgraph client for all CRUD operations.
    - This ensures that the rest of the cognitive architecture can interact with the memory system without changes to the interface.

4.  **Update `core/hgql`:**
    - Modify the `HGQLEngine` to translate HyperGraphQL queries into Dgraph's GraphQL+- query language.
    - This will allow the system to leverage Dgraph's native graph traversal and query optimization capabilities.

5.  **Server Integration:**
    - In `server/autonomous/main.go`, initialize the Dgraph client and pass it to the `NewHypergraphMemory` constructor.
    - The server will be responsible for managing the lifecycle of the Dgraph connection.

### Benefits:

- **Persistence:** Knowledge is no longer ephemeral and will persist across server restarts.
- **Scalability:** The memory system can scale far beyond the limitations of in-memory storage.
- **Performance:** Dgraph is optimized for graph traversals and complex queries, which will improve the performance of relevance and association searches.
- **Queryability:** Dgraph's GraphQL+- API provides a powerful and flexible way to query the knowledge graph.

---

## 2. Goakt Integration: Concurrent Cognitive Engines

**Objective:** Refactor the three concurrent inference engines in `core/echobeats/concurrent_engines.go` to use the `goakt` actor framework, improving concurrency, distribution, and fault tolerance.

### Target Packages & Files:

- **`core/echobeats/`**: The core package for the 12-step cognitive loop.
- **`core/echobeats/concurrent_engines.go`**: The current implementation of the three engines will be replaced with `goakt` actors.
- **`server/autonomous/main.go`**: The `goakt` actor system will be initialized.

### Integration Steps:

1.  **Define Actor Messages:**
    - Using protocol buffers, define the messages that will be passed between the cognitive engine actors (e.g., `Step`, `StateUpdate`, `PivotalSync`).

2.  **Implement Actors:**
    - Create three `goakt` actors: `AffordanceActor`, `RelevanceActor`, and `SalienceActor`.
    - Each actor will encapsulate the logic of one of the current inference engines.
    - The actor's `Receive` method will handle the incoming messages and manage the actor's state.

3.  **Refactor `concurrent_engines.go`:**
    - Replace the current `ConcurrentInferenceSystem` with a `GoAktCognitiveSystem`.
    - This new system will be responsible for creating the actor system, spawning the three engine actors, and orchestrating the 12-step cognitive loop by sending messages to the actors.
    - The `PhaseSynchronizer` will be replaced with `goakt`'s built-in messaging and synchronization mechanisms.

4.  **Server Integration:**
    - In `server/autonomous/main.go`, initialize the `goakt` actor system and the `GoAktCognitiveSystem`.

### Benefits:

- **Improved Concurrency:** `goakt` provides a more structured and robust model for concurrency than the current mix of goroutines and mutexes.
- **Distribution:** The actor model makes it easier to distribute the cognitive engines across multiple machines in the future.
- **Fault Tolerance:** `goakt` provides built-in supervision and fault tolerance mechanisms, making the cognitive loop more resilient.
- **Clarity:** The actor model can make the complex interactions between the cognitive engines easier to understand and maintain.

---

## 3. Cog Integration: Cognitive Component Containerization

**Objective:** Use `cog` to containerize the `echoself` autonomous agent, creating a portable, reproducible, and isolated environment for deployment and execution.

### Target Packages & Files:

- **`cog.yaml`**: A new configuration file will be created in the root of the repository.
- **`predict.go`**: A new file that defines the interface for `cog`.
- **`Dockerfile`**: The existing Dockerfile will be adapted to work with `cog`.

### Integration Steps:

1.  **Create `cog.yaml`:**
    - Define the build environment, including the Go version and any system dependencies.
    - Specify the `predict` interface, defining the inputs (e.g., a prompt or a goal) and outputs (e.g., a thought or an action).

2.  **Implement `predict.go`:**
    - Create a `Predictor` struct that encapsulates the `echoself` autonomous agent.
    - The `Setup` method will initialize the agent (LLM providers, memory system, etc.).
    - The `Predict` method will be the entry point for `cog`, taking an input and returning the agent's output.

3.  **Adapt Dockerfile:**
    - Modify the existing `Dockerfile` to be compatible with `cog`'s build process.
    - This will involve ensuring that the Go toolchain is correctly installed and that the `echoself` binary is built in the right location.

4.  **Build and Run with Cog:**
    - Use the `cog build` command to create the container image.
    - Use `cog run` to execute the `echoself` agent in the containerized environment.

### Benefits:

- **Reproducibility:** `cog` ensures that the `echoself` agent runs in a consistent and reproducible environment.
- **Portability:** The `cog` container can be deployed anywhere that supports Docker/OCI images.
- **Isolation:** The agent runs in an isolated environment, which improves security and stability.
- **Simplified Deployment:** `cog` simplifies the process of deploying the `echoself` agent, especially in cloud environments.
