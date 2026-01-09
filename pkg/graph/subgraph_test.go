package graph

import (
	"testing"
)

func TestComputeSubgraphs_SingleComponent(t *testing.T) {
	// Create a fully connected graph
	g := NewDependencyGraph()
	g.Nodes["A"] = &Node{ID: "A", Name: "A"}
	g.Nodes["B"] = &Node{ID: "B", Name: "B"}
	g.Nodes["C"] = &Node{ID: "C", Name: "C"}
	g.Edges["A"] = []string{"B"}
	g.Edges["B"] = []string{"C"}
	g.Edges["C"] = []string{"A"}

	g.ComputeSubgraphs()

	if len(g.Subgraphs) != 1 {
		t.Errorf("Expected 1 subgraph, got %d", len(g.Subgraphs))
	}

	subgraph := g.Subgraphs[0]
	if len(subgraph.NodeIDs) != 3 {
		t.Errorf("Expected 3 nodes in subgraph, got %d", len(subgraph.NodeIDs))
	}

	if subgraph.EdgeCount != 3 {
		t.Errorf("Expected 3 edges in subgraph, got %d", subgraph.EdgeCount)
	}

	if subgraph.Score <= 0 {
		t.Errorf("Expected positive score, got %f", subgraph.Score)
	}

	// Verify nodes have subgraph metadata
	for _, node := range g.Nodes {
		if node.SubgraphID != 0 {
			t.Errorf("Expected node %s to be in subgraph 0, got %d", node.ID, node.SubgraphID)
		}
		if node.SubgraphScore != subgraph.Score {
			t.Errorf("Expected node %s score %f, got %f", node.ID, subgraph.Score, node.SubgraphScore)
		}
	}
}

func TestComputeSubgraphs_MultipleComponents(t *testing.T) {
	// Create a graph with 2 disconnected components
	g := NewDependencyGraph()

	// Component 1: A -> B -> C (3 nodes, 2 edges)
	g.Nodes["A"] = &Node{ID: "A", Name: "A"}
	g.Nodes["B"] = &Node{ID: "B", Name: "B"}
	g.Nodes["C"] = &Node{ID: "C", Name: "C"}
	g.Edges["A"] = []string{"B"}
	g.Edges["B"] = []string{"C"}

	// Component 2: D -> E (2 nodes, 1 edge)
	g.Nodes["D"] = &Node{ID: "D", Name: "D"}
	g.Nodes["E"] = &Node{ID: "E", Name: "E"}
	g.Edges["D"] = []string{"E"}

	g.ComputeSubgraphs()

	if len(g.Subgraphs) != 2 {
		t.Errorf("Expected 2 subgraphs, got %d", len(g.Subgraphs))
	}

	// Verify larger component (A-B-C) has higher score than smaller (D-E)
	// Subgraphs should be sorted by score
	largestSubgraph := g.Subgraphs[0]
	if len(largestSubgraph.NodeIDs) != 3 {
		t.Errorf("Expected largest subgraph to have 3 nodes, got %d", len(largestSubgraph.NodeIDs))
	}

	smallestSubgraph := g.Subgraphs[1]
	if len(smallestSubgraph.NodeIDs) != 2 {
		t.Errorf("Expected smallest subgraph to have 2 nodes, got %d", len(smallestSubgraph.NodeIDs))
	}

	if largestSubgraph.Score <= smallestSubgraph.Score {
		t.Errorf("Expected larger subgraph score (%f) > smaller subgraph score (%f)",
			largestSubgraph.Score, smallestSubgraph.Score)
	}
}

func TestComputeSubgraphs_IsolatedNodes(t *testing.T) {
	// Create a graph with isolated nodes
	g := NewDependencyGraph()
	g.Nodes["A"] = &Node{ID: "A", Name: "A"}
	g.Nodes["B"] = &Node{ID: "B", Name: "B"}
	g.Nodes["C"] = &Node{ID: "C", Name: "C"}
	// No edges - all isolated

	g.ComputeSubgraphs()

	if len(g.Subgraphs) != 3 {
		t.Errorf("Expected 3 subgraphs (one per isolated node), got %d", len(g.Subgraphs))
	}

	// All isolated nodes should have equal scores
	for _, subgraph := range g.Subgraphs {
		if len(subgraph.NodeIDs) != 1 {
			t.Errorf("Expected isolated node subgraph to have 1 node, got %d", len(subgraph.NodeIDs))
		}
		if subgraph.EdgeCount != 0 {
			t.Errorf("Expected isolated node to have 0 edges, got %d", subgraph.EdgeCount)
		}
	}
}

func TestComputeSubgraphs_EmptyGraph(t *testing.T) {
	g := NewDependencyGraph()
	g.ComputeSubgraphs()

	if len(g.Subgraphs) != 0 {
		t.Errorf("Expected 0 subgraphs for empty graph, got %d", len(g.Subgraphs))
	}
}

func TestComputeSubgraphScore(t *testing.T) {
	tests := []struct {
		name      string
		nodeCount int
		edgeCount int
		wantZero  bool
	}{
		{"empty", 0, 0, true},
		{"single node", 1, 0, false},
		{"two nodes one edge", 2, 1, false},
		{"three nodes three edges", 3, 3, false},
		{"large component", 10, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := computeSubgraphScore(tt.nodeCount, tt.edgeCount)
			if tt.wantZero && score != 0 {
				t.Errorf("Expected score 0, got %f", score)
			}
			if !tt.wantZero && score <= 0 {
				t.Errorf("Expected positive score, got %f", score)
			}
		})
	}

	// Verify score increases with more nodes
	score1 := computeSubgraphScore(2, 1)
	score2 := computeSubgraphScore(3, 2)
	score3 := computeSubgraphScore(5, 4)

	if score1 >= score2 || score2 >= score3 {
		t.Errorf("Expected scores to increase with size: %f, %f, %f", score1, score2, score3)
	}
}

func TestGetSubgraphByID(t *testing.T) {
	g := NewDependencyGraph()
	g.Nodes["A"] = &Node{ID: "A"}
	g.Nodes["B"] = &Node{ID: "B"}
	g.Edges["A"] = []string{"B"}
	g.ComputeSubgraphs()

	subgraph := g.GetSubgraphByID(0)
	if subgraph == nil {
		t.Error("Expected to find subgraph 0")
	}

	notFound := g.GetSubgraphByID(999)
	if notFound != nil {
		t.Error("Expected nil for non-existent subgraph")
	}
}

func TestGetNodeSubgraph(t *testing.T) {
	g := NewDependencyGraph()
	g.Nodes["A"] = &Node{ID: "A"}
	g.Nodes["B"] = &Node{ID: "B"}
	g.Nodes["C"] = &Node{ID: "C"}
	g.Edges["A"] = []string{"B"}
	// C is isolated
	g.ComputeSubgraphs()

	subgraphA := g.GetNodeSubgraph("A")
	subgraphB := g.GetNodeSubgraph("B")
	if subgraphA == nil || subgraphB == nil {
		t.Error("Expected to find subgraph for A and B")
	}
	if subgraphA.ID != subgraphB.ID {
		t.Error("Expected A and B to be in same subgraph")
	}

	subgraphC := g.GetNodeSubgraph("C")
	if subgraphC == nil {
		t.Error("Expected to find subgraph for C")
	}
	if subgraphC.ID == subgraphA.ID {
		t.Error("Expected C to be in different subgraph from A")
	}

	notFound := g.GetNodeSubgraph("NonExistent")
	if notFound != nil {
		t.Error("Expected nil for non-existent node")
	}
}

func TestGetLargestSubgraph(t *testing.T) {
	g := NewDependencyGraph()

	// Large component
	g.Nodes["A"] = &Node{ID: "A"}
	g.Nodes["B"] = &Node{ID: "B"}
	g.Nodes["C"] = &Node{ID: "C"}
	g.Edges["A"] = []string{"B"}
	g.Edges["B"] = []string{"C"}

	// Small component
	g.Nodes["D"] = &Node{ID: "D"}
	g.Edges["D"] = []string{}

	g.ComputeSubgraphs()

	largest := g.GetLargestSubgraph()
	if largest == nil {
		t.Fatal("Expected to find largest subgraph")
	}

	if len(largest.NodeIDs) != 3 {
		t.Errorf("Expected largest subgraph to have 3 nodes, got %d", len(largest.NodeIDs))
	}

	// Empty graph
	emptyGraph := NewDependencyGraph()
	emptyGraph.ComputeSubgraphs()
	if emptyGraph.GetLargestSubgraph() != nil {
		t.Error("Expected nil for empty graph")
	}
}

func TestSubgraphSorting(t *testing.T) {
	g := NewDependencyGraph()

	// Create multiple components of different sizes
	// Component 1: Large (4 nodes, 3 edges)
	g.Nodes["A1"] = &Node{ID: "A1"}
	g.Nodes["A2"] = &Node{ID: "A2"}
	g.Nodes["A3"] = &Node{ID: "A3"}
	g.Nodes["A4"] = &Node{ID: "A4"}
	g.Edges["A1"] = []string{"A2"}
	g.Edges["A2"] = []string{"A3"}
	g.Edges["A3"] = []string{"A4"}

	// Component 2: Medium (3 nodes, 2 edges)
	g.Nodes["B1"] = &Node{ID: "B1"}
	g.Nodes["B2"] = &Node{ID: "B2"}
	g.Nodes["B3"] = &Node{ID: "B3"}
	g.Edges["B1"] = []string{"B2"}
	g.Edges["B2"] = []string{"B3"}

	// Component 3: Small (2 nodes, 1 edge)
	g.Nodes["C1"] = &Node{ID: "C1"}
	g.Nodes["C2"] = &Node{ID: "C2"}
	g.Edges["C1"] = []string{"C2"}

	g.ComputeSubgraphs()

	if len(g.Subgraphs) != 3 {
		t.Fatalf("Expected 3 subgraphs, got %d", len(g.Subgraphs))
	}

	// Verify sorting: scores should be in descending order
	for i := 0; i < len(g.Subgraphs)-1; i++ {
		if g.Subgraphs[i].Score < g.Subgraphs[i+1].Score {
			t.Errorf("Subgraphs not sorted: subgraph %d (score %f) < subgraph %d (score %f)",
				i, g.Subgraphs[i].Score, i+1, g.Subgraphs[i+1].Score)
		}
	}

	// Verify largest component is first
	if len(g.Subgraphs[0].NodeIDs) != 4 {
		t.Errorf("Expected largest subgraph first, got %d nodes", len(g.Subgraphs[0].NodeIDs))
	}
}
