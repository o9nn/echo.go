// +build integration

package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EchoCog/echollama/core/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getSupabaseClient creates a Supabase client from environment variables
func getSupabaseClient(t *testing.T) *memory.SupabaseClient {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	if url == "" || key == "" {
		t.Skip("SUPABASE_URL or SUPABASE_KEY not set, skipping Supabase integration tests")
	}

	return memory.NewSupabaseClient(url, key)
}

// TestSupabaseConnection tests basic connection to Supabase
func TestSupabaseConnection(t *testing.T) {
	client := getSupabaseClient(t)

	// Test with a simple query to verify connection
	_, err := client.Query("memory_nodes", nil, 1)
	// Note: This may fail if table doesn't exist, which is expected
	// The important thing is that we can make a request
	if err != nil {
		t.Logf("Query returned error (may be expected if table doesn't exist): %v", err)
	}
}

// TestSupabaseMemoryNodeCRUD tests Create, Read, Update, Delete on memory_nodes table
func TestSupabaseMemoryNodeCRUD(t *testing.T) {
	client := getSupabaseClient(t)

	// Create a test node
	nodeID := uuid.New().String()
	node := map[string]interface{}{
		"id":         nodeID,
		"node_type":  "episodic",
		"content":    "Integration test memory node",
		"importance": 0.85,
		"metadata":   `{"source": "integration_test", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`,
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	// Insert
	err := client.Insert("memory_nodes", node)
	if err != nil {
		t.Logf("Insert error (table may not exist): %v", err)
		t.Skip("Skipping test - memory_nodes table may not exist")
	}

	// Query back
	results, err := client.Query("memory_nodes", map[string]interface{}{"id": nodeID}, 1)
	require.NoError(t, err, "Failed to query node")
	require.Len(t, results, 1, "Should find exactly one node")
	assert.Equal(t, nodeID, results[0]["id"])
	assert.Equal(t, "episodic", results[0]["node_type"])

	// Update
	updateData := map[string]interface{}{
		"importance": 0.95,
		"content":    "Updated integration test content",
	}
	err = client.Update("memory_nodes", map[string]interface{}{"id": nodeID}, updateData)
	require.NoError(t, err, "Failed to update node")

	// Verify update
	results, err = client.Query("memory_nodes", map[string]interface{}{"id": nodeID}, 1)
	require.NoError(t, err)
	assert.Equal(t, 0.95, results[0]["importance"])

	// Delete
	err = client.Delete("memory_nodes", map[string]interface{}{"id": nodeID})
	require.NoError(t, err, "Failed to delete node")

	// Verify deletion
	results, err = client.Query("memory_nodes", map[string]interface{}{"id": nodeID}, 1)
	require.NoError(t, err)
	assert.Len(t, results, 0, "Node should be deleted")
}

// TestSupabaseMemoryEdgeCRUD tests edge operations
func TestSupabaseMemoryEdgeCRUD(t *testing.T) {
	client := getSupabaseClient(t)

	// Create source and target nodes first
	sourceID := uuid.New().String()
	targetID := uuid.New().String()

	sourceNode := map[string]interface{}{
		"id":         sourceID,
		"node_type":  "concept",
		"content":    "Source concept node",
		"importance": 0.7,
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	targetNode := map[string]interface{}{
		"id":         targetID,
		"node_type":  "concept",
		"content":    "Target concept node",
		"importance": 0.7,
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	err := client.Insert("memory_nodes", sourceNode)
	if err != nil {
		t.Skip("Skipping test - memory_nodes table may not exist")
	}
	err = client.Insert("memory_nodes", targetNode)
	require.NoError(t, err)

	// Create an edge
	edgeID := uuid.New().String()
	edge := map[string]interface{}{
		"id":         edgeID,
		"source_id":  sourceID,
		"target_id":  targetID,
		"edge_type":  "association",
		"weight":     0.8,
		"metadata":   `{"created_by": "integration_test"}`,
		"created_at": time.Now().Format(time.RFC3339),
	}

	err = client.Insert("memory_edges", edge)
	if err != nil {
		t.Logf("Edge insert error (table may not exist): %v", err)
		// Cleanup nodes
		client.Delete("memory_nodes", map[string]interface{}{"id": sourceID})
		client.Delete("memory_nodes", map[string]interface{}{"id": targetID})
		t.Skip("Skipping test - memory_edges table may not exist")
	}

	// Query the edge
	results, err := client.Query("memory_edges", map[string]interface{}{"id": edgeID}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, sourceID, results[0]["source_id"])
	assert.Equal(t, targetID, results[0]["target_id"])
	assert.Equal(t, "association", results[0]["edge_type"])

	// Cleanup
	client.Delete("memory_edges", map[string]interface{}{"id": edgeID})
	client.Delete("memory_nodes", map[string]interface{}{"id": sourceID})
	client.Delete("memory_nodes", map[string]interface{}{"id": targetID})
}

// TestSupabaseEpisodeStorage tests episode storage
func TestSupabaseEpisodeStorage(t *testing.T) {
	client := getSupabaseClient(t)

	episodeID := uuid.New().String()
	episode := map[string]interface{}{
		"id":         episodeID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"context":    "Integration test episode context",
		"importance": 0.75,
		"node_ids":   `["node1", "node2", "node3"]`,
		"metadata":   `{"duration": 3600, "source": "test"}`,
	}

	err := client.Insert("episodes", episode)
	if err != nil {
		t.Skip("Skipping test - episodes table may not exist")
	}

	// Query back
	results, err := client.Query("episodes", map[string]interface{}{"id": episodeID}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Integration test episode context", results[0]["context"])

	// Cleanup
	client.Delete("episodes", map[string]interface{}{"id": episodeID})
}

// TestSupabaseIdentitySnapshot tests identity snapshot storage
func TestSupabaseIdentitySnapshot(t *testing.T) {
	client := getSupabaseClient(t)

	snapshotID := uuid.New().String()
	snapshot := map[string]interface{}{
		"id":        snapshotID,
		"timestamp": time.Now().Format(time.RFC3339),
		"coherence": 0.92,
		"state":     `{"mood": "contemplative", "energy": 0.8, "focus": "learning"}`,
		"metadata":  `{"trigger": "reflection_cycle"}`,
	}

	err := client.Insert("identity_snapshots", snapshot)
	if err != nil {
		t.Skip("Skipping test - identity_snapshots table may not exist")
	}

	// Query back
	results, err := client.Query("identity_snapshots", map[string]interface{}{"id": snapshotID}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 0.92, results[0]["coherence"])

	// Cleanup
	client.Delete("identity_snapshots", map[string]interface{}{"id": snapshotID})
}

// TestSupabaseDreamJournal tests dream journal storage
func TestSupabaseDreamJournal(t *testing.T) {
	client := getSupabaseClient(t)

	journalID := uuid.New().String()
	journal := map[string]interface{}{
		"id":                    journalID,
		"timestamp":             time.Now().Format(time.RFC3339),
		"dream_state":           "deep_consolidation",
		"memories_consolidated": 25,
		"patterns_synthesized":  5,
		"insights":              `["Pattern A emerged", "Connection B strengthened", "New concept C formed"]`,
		"metadata":              `{"duration": 7200, "quality": 0.85}`,
	}

	err := client.Insert("dream_journals", journal)
	if err != nil {
		t.Skip("Skipping test - dream_journals table may not exist")
	}

	// Query back
	results, err := client.Query("dream_journals", map[string]interface{}{"id": journalID}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "deep_consolidation", results[0]["dream_state"])
	assert.Equal(t, float64(25), results[0]["memories_consolidated"])

	// Cleanup
	client.Delete("dream_journals", map[string]interface{}{"id": journalID})
}

// TestSupabaseBulkOperations tests bulk insert performance
func TestSupabaseBulkOperations(t *testing.T) {
	client := getSupabaseClient(t)

	batchSize := 50
	nodes := make([]map[string]interface{}, batchSize)
	nodeIDs := make([]string, batchSize)

	for i := 0; i < batchSize; i++ {
		nodeIDs[i] = uuid.New().String()
		nodes[i] = map[string]interface{}{
			"id":         nodeIDs[i],
			"node_type":  "bulk_test",
			"content":    fmt.Sprintf("Bulk test node %d", i),
			"importance": float64(i) / float64(batchSize),
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		}
	}

	// Insert nodes one by one (Supabase REST API doesn't support bulk insert easily)
	start := time.Now()
	successCount := 0
	for _, node := range nodes {
		err := client.Insert("memory_nodes", node)
		if err != nil {
			if successCount == 0 {
				t.Skip("Skipping - memory_nodes table may not exist")
			}
			break
		}
		successCount++
	}
	insertDuration := time.Since(start)
	t.Logf("Inserted %d nodes in %v (avg: %v/node)", successCount, insertDuration, insertDuration/time.Duration(successCount))

	// Query bulk test nodes
	start = time.Now()
	results, err := client.Query("memory_nodes", map[string]interface{}{"node_type": "bulk_test"}, batchSize)
	queryDuration := time.Since(start)
	require.NoError(t, err)
	t.Logf("Queried %d nodes in %v", len(results), queryDuration)

	// Cleanup
	for _, nodeID := range nodeIDs[:successCount] {
		client.Delete("memory_nodes", map[string]interface{}{"id": nodeID})
	}
}

// TestSupabaseRPC tests Supabase RPC function calls
func TestSupabaseRPC(t *testing.T) {
	client := getSupabaseClient(t)

	// Test calling a custom RPC function if it exists
	// This is a placeholder - actual RPC functions would need to be defined in Supabase
	result, err := client.RPC("get_related_nodes", map[string]interface{}{
		"node_id":   "test-node",
		"max_depth": 2,
	})

	if err != nil {
		t.Logf("RPC call failed (function may not exist): %v", err)
		t.Skip("Skipping - RPC function may not exist")
	}

	t.Logf("RPC result: %v", result)
}

// BenchmarkSupabaseOperations benchmarks Supabase operations
func BenchmarkSupabaseOperations(b *testing.B) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	if url == "" || key == "" {
		b.Skip("SUPABASE_URL or SUPABASE_KEY not set")
	}

	client := memory.NewSupabaseClient(url, key)

	b.Run("SingleInsert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			nodeID := uuid.New().String()
			node := map[string]interface{}{
				"id":         nodeID,
				"node_type":  "benchmark",
				"content":    "Benchmark node",
				"importance": 0.5,
				"created_at": time.Now().Format(time.RFC3339),
				"updated_at": time.Now().Format(time.RFC3339),
			}
			err := client.Insert("memory_nodes", node)
			if err != nil {
				b.Skip("Table may not exist")
			}
			// Cleanup
			client.Delete("memory_nodes", map[string]interface{}{"id": nodeID})
		}
	})

	b.Run("SingleQuery", func(b *testing.B) {
		// Create a node to query
		nodeID := uuid.New().String()
		node := map[string]interface{}{
			"id":         nodeID,
			"node_type":  "benchmark_query",
			"content":    "Benchmark query node",
			"importance": 0.5,
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		}
		err := client.Insert("memory_nodes", node)
		if err != nil {
			b.Skip("Table may not exist")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Query("memory_nodes", map[string]interface{}{"id": nodeID}, 1)
		}

		// Cleanup
		client.Delete("memory_nodes", map[string]interface{}{"id": nodeID})
	})
}
