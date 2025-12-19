package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryNode tests the MemoryNode struct and its methods
func TestMemoryNode(t *testing.T) {
	t.Run("CreateNode", func(t *testing.T) {
		node := &MemoryNode{
			ID:        "test-node-1",
			Type:      NodeTypeEpisodic,
			Content:   "Test memory content",
			Embedding: []float64{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"source": "test",
			},
		}

		assert.Equal(t, "test-node-1", node.ID)
		assert.Equal(t, NodeTypeEpisodic, node.Type)
		assert.Equal(t, "Test memory content", node.Content)
		assert.Len(t, node.Embedding, 3)
	})

	t.Run("NodeTypes", func(t *testing.T) {
		types := []NodeType{
			NodeTypeEpisodic,
			NodeTypeSemantic,
			NodeTypeProcedural,
			NodeTypeWorking,
		}

		for _, nodeType := range types {
			assert.NotEmpty(t, string(nodeType))
		}
	})
}

// TestMemoryEdge tests the MemoryEdge struct
func TestMemoryEdge(t *testing.T) {
	t.Run("CreateEdge", func(t *testing.T) {
		edge := &MemoryEdge{
			ID:     "test-edge-1",
			Type:   EdgeTypeAssociation,
			Source: "node-1",
			Target: "node-2",
			Weight: 0.8,
			Metadata: map[string]interface{}{
				"created": time.Now(),
			},
		}

		assert.Equal(t, "test-edge-1", edge.ID)
		assert.Equal(t, EdgeTypeAssociation, edge.Type)
		assert.Equal(t, "node-1", edge.Source)
		assert.Equal(t, "node-2", edge.Target)
		assert.Equal(t, 0.8, edge.Weight)
	})

	t.Run("EdgeTypes", func(t *testing.T) {
		types := []EdgeType{
			EdgeTypeAssociation,
			EdgeTypeCausal,
			EdgeTypeTemporal,
			EdgeTypeSimilarity,
		}

		for _, edgeType := range types {
			assert.NotEmpty(t, string(edgeType))
		}
	})
}

// TestHyperEdge tests the HyperEdge struct
func TestHyperEdge(t *testing.T) {
	t.Run("CreateHyperEdge", func(t *testing.T) {
		hyperEdge := &HyperEdge{
			ID:      "test-hyperedge-1",
			Type:    HyperEdgeTypeContext,
			NodeIDs: []string{"node-1", "node-2", "node-3"},
			Weight:  0.9,
			Metadata: map[string]interface{}{
				"context": "test-context",
			},
		}

		assert.Equal(t, "test-hyperedge-1", hyperEdge.ID)
		assert.Equal(t, HyperEdgeTypeContext, hyperEdge.Type)
		assert.Len(t, hyperEdge.NodeIDs, 3)
		assert.Equal(t, 0.9, hyperEdge.Weight)
	})
}

// TestHypergraphMemory tests the in-memory hypergraph implementation
func TestHypergraphMemory(t *testing.T) {
	t.Run("NewHypergraphMemory", func(t *testing.T) {
		hg := NewHypergraphMemory()
		require.NotNil(t, hg)
	})

	t.Run("AddAndGetNode", func(t *testing.T) {
		hg := NewHypergraphMemory()

		node := &MemoryNode{
			ID:      "test-node",
			Type:    NodeTypeEpisodic,
			Content: "Test content",
		}

		err := hg.AddNode(node)
		require.NoError(t, err)

		retrieved, err := hg.GetNode("test-node")
		require.NoError(t, err)
		assert.Equal(t, node.ID, retrieved.ID)
		assert.Equal(t, node.Content, retrieved.Content)
	})

	t.Run("AddAndGetEdge", func(t *testing.T) {
		hg := NewHypergraphMemory()

		// Add source and target nodes first
		sourceNode := &MemoryNode{ID: "source", Type: NodeTypeEpisodic, Content: "Source"}
		targetNode := &MemoryNode{ID: "target", Type: NodeTypeEpisodic, Content: "Target"}
		require.NoError(t, hg.AddNode(sourceNode))
		require.NoError(t, hg.AddNode(targetNode))

		edge := &MemoryEdge{
			ID:     "test-edge",
			Type:   EdgeTypeAssociation,
			Source: "source",
			Target: "target",
			Weight: 0.5,
		}

		err := hg.AddEdge(edge)
		require.NoError(t, err)
	})

	t.Run("AddHyperEdge", func(t *testing.T) {
		hg := NewHypergraphMemory()

		// Add member nodes
		for i := 1; i <= 3; i++ {
			node := &MemoryNode{
				ID:      fmt.Sprintf("node-%d", i),
				Type:    NodeTypeEpisodic,
				Content: fmt.Sprintf("Content %d", i),
			}
			require.NoError(t, hg.AddNode(node))
		}

		hyperEdge := &HyperEdge{
			ID:      "test-hyperedge",
			Type:    HyperEdgeTypeContext,
			NodeIDs: []string{"node-1", "node-2", "node-3"},
			Weight:  0.7,
		}

		err := hg.AddHyperEdge(hyperEdge)
		require.NoError(t, err)
	})

	t.Run("GetNodeNotFound", func(t *testing.T) {
		hg := NewHypergraphMemory()

		_, err := hg.GetNode("nonexistent")
		assert.Error(t, err)
	})

	t.Run("Traverse", func(t *testing.T) {
		hg := NewHypergraphMemory()

		// Create a simple graph: A -> B -> C
		nodes := []*MemoryNode{
			{ID: "A", Type: NodeTypeEpisodic, Content: "Node A"},
			{ID: "B", Type: NodeTypeEpisodic, Content: "Node B"},
			{ID: "C", Type: NodeTypeEpisodic, Content: "Node C"},
		}
		for _, node := range nodes {
			require.NoError(t, hg.AddNode(node))
		}

		edges := []*MemoryEdge{
			{ID: "A-B", Type: EdgeTypeAssociation, Source: "A", Target: "B", Weight: 1.0},
			{ID: "B-C", Type: EdgeTypeAssociation, Source: "B", Target: "C", Weight: 1.0},
		}
		for _, edge := range edges {
			require.NoError(t, hg.AddEdge(edge))
		}

		// Traverse from A with depth 2
		result, err := hg.Traverse("A", 2, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

// BenchmarkHypergraphMemory benchmarks the hypergraph operations
func BenchmarkHypergraphMemory(b *testing.B) {
	b.Run("AddNode", func(b *testing.B) {
		hg := NewHypergraphMemory()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			node := &MemoryNode{
				ID:      fmt.Sprintf("node-%d", i),
				Type:    NodeTypeEpisodic,
				Content: "Benchmark content",
			}
			hg.AddNode(node)
		}
	})

	b.Run("GetNode", func(b *testing.B) {
		hg := NewHypergraphMemory()
		// Pre-populate
		for i := 0; i < 1000; i++ {
			node := &MemoryNode{
				ID:      fmt.Sprintf("node-%d", i),
				Type:    NodeTypeEpisodic,
				Content: "Benchmark content",
			}
			hg.AddNode(node)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			hg.GetNode(fmt.Sprintf("node-%d", i%1000))
		}
	})
}
