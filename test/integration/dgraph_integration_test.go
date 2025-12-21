// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cogpy/echo9llama/core/memory"
	"github.com/cogpy/echo9llama/core/persistence"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDgraphConnection tests real connection to Dgraph
func TestDgraphConnection(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	config := persistence.DefaultDgraphConfig()
	client, err := persistence.NewDgraphClient(config)
	require.NoError(t, err, "Failed to connect to Dgraph")
	defer client.Close()

	assert.True(t, client.IsConnected(), "Client should be connected")
}

// TestDgraphSchemaSetup tests setting up the hypergraph schema
func TestDgraphSchemaSetup(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	client, err := persistence.NewDgraphClient(nil)
	require.NoError(t, err)
	defer client.Close()

	// Read schema from file
	schemaPath := "../../core/memory/schema.dql"
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		// Use inline schema if file not found
		schemaBytes = []byte(`
			# Memory Node Schema
			node_id: string @index(exact) .
			node_type: string @index(exact) .
			content: string @index(fulltext) .
			embedding: [float] .
			importance: float .
			created_at: datetime .
			updated_at: datetime .
			metadata: string .

			# Memory Edge Schema
			edge_id: string @index(exact) .
			edge_type: string @index(exact) .
			weight: float .
			source: uid @reverse .
			target: uid @reverse .

			# HyperEdge Schema
			hyperedge_id: string @index(exact) .
			hyperedge_type: string @index(exact) .
			members: [uid] @reverse .

			# Type definitions
			type MemoryNode {
				node_id
				node_type
				content
				embedding
				importance
				created_at
				updated_at
				metadata
			}

			type MemoryEdge {
				edge_id
				edge_type
				weight
				source
				target
			}

			type HyperEdge {
				hyperedge_id
				hyperedge_type
				members
			}
		`)
	}

	err = client.SetSchema(string(schemaBytes))
	require.NoError(t, err, "Failed to set schema")
}

// TestDgraphNodeCRUD tests Create, Read, Update, Delete operations on nodes
func TestDgraphNodeCRUD(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	client, err := persistence.NewDgraphClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// Create a test node
	nodeID := uuid.New().String()
	node := map[string]interface{}{
		"uid":        "_:newnode",
		"dgraph.type": "MemoryNode",
		"node_id":    nodeID,
		"node_type":  "episodic",
		"content":    "Test memory content for integration test",
		"importance": 0.85,
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	nodeJSON, err := json.Marshal(node)
	require.NoError(t, err)

	mu := &api.Mutation{
		SetJson:   nodeJSON,
		CommitNow: true,
	}

	resp, err := client.Mutate(ctx, mu)
	require.NoError(t, err, "Failed to create node")
	assert.NotEmpty(t, resp.Uids, "Should return UIDs for created node")

	// Read the node back
	query := fmt.Sprintf(`{
		node(func: eq(node_id, "%s")) {
			uid
			node_id
			node_type
			content
			importance
			created_at
		}
	}`, nodeID)

	readResp, err := client.Query(ctx, query, nil)
	require.NoError(t, err, "Failed to query node")

	var result struct {
		Node []struct {
			UID        string  `json:"uid"`
			NodeID     string  `json:"node_id"`
			NodeType   string  `json:"node_type"`
			Content    string  `json:"content"`
			Importance float64 `json:"importance"`
		} `json:"node"`
	}

	err = json.Unmarshal(readResp.Json, &result)
	require.NoError(t, err)
	require.Len(t, result.Node, 1, "Should find exactly one node")
	assert.Equal(t, nodeID, result.Node[0].NodeID)
	assert.Equal(t, "episodic", result.Node[0].NodeType)
	assert.Equal(t, 0.85, result.Node[0].Importance)

	// Update the node
	updateMu := &api.Mutation{
		SetJson: []byte(fmt.Sprintf(`{
			"uid": "%s",
			"importance": 0.95,
			"content": "Updated content"
		}`, result.Node[0].UID)),
		CommitNow: true,
	}

	_, err = client.Mutate(ctx, updateMu)
	require.NoError(t, err, "Failed to update node")

	// Verify update
	readResp, err = client.Query(ctx, query, nil)
	require.NoError(t, err)
	err = json.Unmarshal(readResp.Json, &result)
	require.NoError(t, err)
	assert.Equal(t, 0.95, result.Node[0].Importance)

	// Delete the node
	deleteMu := &api.Mutation{
		DeleteJson: []byte(fmt.Sprintf(`{"uid": "%s"}`, result.Node[0].UID)),
		CommitNow:  true,
	}

	_, err = client.Mutate(ctx, deleteMu)
	require.NoError(t, err, "Failed to delete node")

	// Verify deletion
	readResp, err = client.Query(ctx, query, nil)
	require.NoError(t, err)
	err = json.Unmarshal(readResp.Json, &result)
	require.NoError(t, err)
	assert.Len(t, result.Node, 0, "Node should be deleted")
}

// TestDgraphEdgeOperations tests edge creation and traversal
func TestDgraphEdgeOperations(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	client, err := persistence.NewDgraphClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// Create two nodes
	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	nodes := []map[string]interface{}{
		{
			"uid":         "_:source",
			"dgraph.type": "MemoryNode",
			"node_id":     sourceID,
			"node_type":   "concept",
			"content":     "Source concept",
		},
		{
			"uid":         "_:target",
			"dgraph.type": "MemoryNode",
			"node_id":     targetID,
			"node_type":   "concept",
			"content":     "Target concept",
		},
	}

	nodesJSON, _ := json.Marshal(nodes)
	resp, err := client.Mutate(ctx, &api.Mutation{SetJson: nodesJSON, CommitNow: true})
	require.NoError(t, err)

	sourceUID := resp.Uids["source"]
	targetUID := resp.Uids["target"]

	// Create an edge between them
	edge := map[string]interface{}{
		"uid":         sourceUID,
		"source": map[string]interface{}{
			"uid":         "_:edge",
			"dgraph.type": "MemoryEdge",
			"edge_id":     uuid.New().String(),
			"edge_type":   "association",
			"weight":      0.75,
			"target": map[string]string{
				"uid": targetUID,
			},
		},
	}

	edgeJSON, _ := json.Marshal(edge)
	_, err = client.Mutate(ctx, &api.Mutation{SetJson: edgeJSON, CommitNow: true})
	require.NoError(t, err, "Failed to create edge")

	// Query to traverse the edge
	query := fmt.Sprintf(`{
		source(func: eq(node_id, "%s")) {
			node_id
			content
			source {
				edge_type
				weight
				target {
					node_id
					content
				}
			}
		}
	}`, sourceID)

	readResp, err := client.Query(ctx, query, nil)
	require.NoError(t, err)

	t.Logf("Edge traversal result: %s", string(readResp.Json))

	// Cleanup
	client.Mutate(ctx, &api.Mutation{
		DeleteJson: []byte(fmt.Sprintf(`[{"uid": "%s"}, {"uid": "%s"}]`, sourceUID, targetUID)),
		CommitNow:  true,
	})
}

// TestDgraphHypergraphMemory tests the DgraphHypergraph implementation
func TestDgraphHypergraphMemory(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	config := persistence.DefaultDgraphConfig()
	client, err := persistence.NewDgraphClient(config)
	require.NoError(t, err)
	defer client.Close()

	hg, err := memory.NewDgraphHypergraph(client)
	require.NoError(t, err)
	defer hg.Close()

	ctx := context.Background()

	// Test AddNode
	node := &memory.MemoryNode{
		ID:         uuid.New().String(),
		Type:       memory.NodeConcept,
		Content:    "Integration test concept",
		Importance: 0.9,
		Metadata: map[string]interface{}{
			"source": "integration_test",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = hg.AddNode(node)
	require.NoError(t, err, "Failed to add node to hypergraph")

	// Test GetNode
	retrieved, err := hg.GetNode(node.ID)
	require.NoError(t, err, "Failed to get node from hypergraph")
	assert.Equal(t, node.ID, retrieved.ID)
	assert.Equal(t, node.Content, retrieved.Content)

	// Test GetNodesByType
	nodes, err := hg.GetNodesByType(memory.NodeConcept, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)

	// Note: Cleanup would be done via direct Dgraph client
	_ = ctx // ctx used for future cleanup operations
}

// TestDgraphBulkOperations tests bulk insert and query performance
func TestDgraphBulkOperations(t *testing.T) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		t.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph integration tests")
	}

	client, err := persistence.NewDgraphClient(nil)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	batchSize := 100

	// Create batch of nodes
	nodes := make([]map[string]interface{}, batchSize)
	nodeIDs := make([]string, batchSize)

	for i := 0; i < batchSize; i++ {
		nodeIDs[i] = uuid.New().String()
		nodes[i] = map[string]interface{}{
			"uid":         fmt.Sprintf("_:node%d", i),
			"dgraph.type": "MemoryNode",
			"node_id":     nodeIDs[i],
			"node_type":   "bulk_test",
			"content":     fmt.Sprintf("Bulk test node %d", i),
			"importance":  float64(i) / float64(batchSize),
		}
	}

	start := time.Now()
	nodesJSON, _ := json.Marshal(nodes)
	resp, err := client.Mutate(ctx, &api.Mutation{SetJson: nodesJSON, CommitNow: true})
	insertDuration := time.Since(start)

	require.NoError(t, err, "Failed bulk insert")
	assert.Len(t, resp.Uids, batchSize)
	t.Logf("Bulk insert of %d nodes took %v", batchSize, insertDuration)

	// Query all bulk test nodes
	start = time.Now()
	query := `{
		nodes(func: eq(node_type, "bulk_test"), first: 100) {
			node_id
			content
			importance
		}
	}`

	readResp, err := client.Query(ctx, query, nil)
	queryDuration := time.Since(start)

	require.NoError(t, err)
	t.Logf("Query of %d nodes took %v", batchSize, queryDuration)

	var result struct {
		Nodes []struct {
			NodeID string `json:"node_id"`
		} `json:"nodes"`
	}
	json.Unmarshal(readResp.Json, &result)
	assert.Len(t, result.Nodes, batchSize)

	// Cleanup - delete all bulk test nodes using upsert
	client.Upsert(ctx, `{ nodes as var(func: eq(node_type, "bulk_test")) }`, &api.Mutation{
		DelNquads: []byte(`uid(nodes) * * .`),
		Cond:      "@if(gt(len(nodes), 0))",
	})
}

// BenchmarkDgraphOperations benchmarks Dgraph operations
func BenchmarkDgraphOperations(b *testing.B) {
	if os.Getenv("DGRAPH_ENDPOINT") == "" {
		b.Skip("DGRAPH_ENDPOINT not set, skipping Dgraph benchmarks")
	}

	client, err := persistence.NewDgraphClient(nil)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	b.Run("SingleNodeInsert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			node := map[string]interface{}{
				"uid":         "_:node",
				"dgraph.type": "MemoryNode",
				"node_id":     uuid.New().String(),
				"node_type":   "benchmark",
				"content":     "Benchmark node",
			}
			nodeJSON, _ := json.Marshal(node)
			client.Mutate(ctx, &api.Mutation{SetJson: nodeJSON, CommitNow: true})
		}
	})

	b.Run("SingleNodeQuery", func(b *testing.B) {
		// Create a node to query
		nodeID := uuid.New().String()
		node := map[string]interface{}{
			"uid":         "_:node",
			"dgraph.type": "MemoryNode",
			"node_id":     nodeID,
			"node_type":   "benchmark",
			"content":     "Benchmark query node",
		}
		nodeJSON, _ := json.Marshal(node)
		client.Mutate(ctx, &api.Mutation{SetJson: nodeJSON, CommitNow: true})

		query := fmt.Sprintf(`{ node(func: eq(node_id, "%s")) { node_id content } }`, nodeID)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Query(ctx, query, nil)
		}
	})
}
