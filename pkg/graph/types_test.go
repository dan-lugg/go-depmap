package graph

import "testing"

func Test_NewDependencyGraph(t *testing.T) {
	g := NewDependencyGraph()

	if g == nil {
		t.Fatal("NewDependencyGraph returned nil")
	}

	if g.Nodes == nil {
		t.Error("Nodes map is nil")
	}

	if g.Edges == nil {
		t.Error("Edges map is nil")
	}

	if len(g.Nodes) != 0 {
		t.Errorf("Expected empty Nodes map, got %d entries", len(g.Nodes))
	}

	if len(g.Edges) != 0 {
		t.Errorf("Expected empty Edges map, got %d entries", len(g.Edges))
	}
}

func Test_DependencyGraph_CountEdges(t *testing.T) {
	tests := []struct {
		name     string
		edges    map[string][]string
		expected int
	}{
		{
			name:     "empty graph",
			edges:    map[string][]string{},
			expected: 0,
		},
		{
			name: "single edge",
			edges: map[string][]string{
				"node1": {"node2"},
			},
			expected: 1,
		},
		{
			name: "multiple edges from one node",
			edges: map[string][]string{
				"node1": {"node2", "node3", "node4"},
			},
			expected: 3,
		},
		{
			name: "multiple nodes with edges",
			edges: map[string][]string{
				"node1": {"node2", "node3"},
				"node2": {"node3"},
				"node4": {"node1", "node2", "node3"},
			},
			expected: 6,
		},
		{
			name: "node with empty edge list",
			edges: map[string][]string{
				"node1": {},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &DependencyGraph{
				Nodes: make(map[string]*Node),
				Edges: tt.edges,
			}

			count := g.CountEdges()
			if count != tt.expected {
				t.Errorf("CountEdges() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func Test_NodeKind_Constants(t *testing.T) {
	tests := []struct {
		kind     NodeKind
		expected string
	}{
		{KindFunction, "function"},
		{KindMethod, "method"},
		{KindType, "type"},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			if string(tt.kind) != tt.expected {
				t.Errorf("NodeKind %s = %s, want %s", tt.kind, string(tt.kind), tt.expected)
			}
		})
	}
}

func Test_Node_Structure(t *testing.T) {
	node := &Node{
		ID:        "test::func1",
		Name:      "func1",
		Kind:      KindFunction,
		Package:   "test",
		File:      "test.go",
		Line:      42,
		Signature: "func func1() error",
	}

	if node.ID != "test::func1" {
		t.Errorf("ID = %s, want test::func1", node.ID)
	}

	if node.Name != "func1" {
		t.Errorf("Name = %s, want func1", node.Name)
	}

	if node.Kind != KindFunction {
		t.Errorf("Kind = %s, want %s", node.Kind, KindFunction)
	}

	if node.Package != "test" {
		t.Errorf("Package = %s, want test", node.Package)
	}

	if node.File != "test.go" {
		t.Errorf("File = %s, want test.go", node.File)
	}

	if node.Line != 42 {
		t.Errorf("Line = %d, want 42", node.Line)
	}

	if node.Signature != "func func1() error" {
		t.Errorf("Signature = %s, want func func1() error", node.Signature)
	}
}

func Test_DependencyGraph_AddNodesAndEdges(t *testing.T) {
	g := NewDependencyGraph()

	node1 := &Node{
		ID:   "test::func1",
		Name: "func1",
		Kind: KindFunction,
	}
	node2 := &Node{
		ID:   "test::func2",
		Name: "func2",
		Kind: KindFunction,
	}

	g.Nodes[node1.ID] = node1
	g.Nodes[node2.ID] = node2

	if len(g.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g.Nodes))
	}

	g.Edges[node1.ID] = []string{node2.ID}

	if len(g.Edges) != 1 {
		t.Errorf("Expected 1 edge entry, got %d", len(g.Edges))
	}

	if g.CountEdges() != 1 {
		t.Errorf("Expected 1 total edge, got %d", g.CountEdges())
	}

	targets, exists := g.Edges[node1.ID]
	if !exists {
		t.Error("Edge from node1 doesn't exist")
	}

	if len(targets) != 1 || targets[0] != node2.ID {
		t.Errorf("Expected edge to node2, got %v", targets)
	}
}

func Test_DependencyGraph_NilEdges(t *testing.T) {
	g := &DependencyGraph{
		Nodes: make(map[string]*Node),
		Edges: nil,
	}

	count := g.CountEdges()
	if count != 0 {
		t.Errorf("Expected 0 edges for nil Edges map, got %d", count)
	}
}
