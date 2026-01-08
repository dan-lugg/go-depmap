package format

import (
	"bytes"
	"encoding/json"
	"testing"

	"go-depmap/pkg/graph"
)

func Test_D3JSJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name    string
		graph   *graph.DependencyGraph
		wantErr bool
	}{
		{
			name:    "empty graph",
			graph:   graph.NewDependencyGraph(),
			wantErr: false,
		},
		{
			name: "graph with single node",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
				},
				Edges: make(map[string][]string),
			},
			wantErr: false,
		},
		{
			name: "graph with multiple nodes and edges",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
					"test::method1": {
						ID:        "test::method1",
						Name:      "method1",
						Kind:      graph.KindMethod,
						Package:   "test",
						File:      "test.go",
						Line:      20,
						Signature: "func (t Type) method1()",
					},
					"test::Type1": {
						ID:        "test::Type1",
						Name:      "Type1",
						Kind:      graph.KindType,
						Package:   "test",
						File:      "test.go",
						Line:      5,
						Signature: "type Type1 struct{}",
					},
				},
				Edges: map[string][]string{
					"test::func1":   {"test::Type1"},
					"test::method1": {"test::func1", "test::Type1"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &D3JSJSONWriter{}
			var buf bytes.Buffer

			err := w.Write(&buf, tt.graph)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var result D3JSGraph
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Output is not valid D3JS JSON: %v", err)
					return
				}

				if len(result.Nodes) != len(tt.graph.Nodes) {
					t.Errorf("Node count mismatch: got %d, want %d",
						len(result.Nodes), len(tt.graph.Nodes))
				}

				expectedEdgeCount := 0
				for _, targets := range tt.graph.Edges {
					expectedEdgeCount += len(targets)
				}
				if len(result.Links) != expectedEdgeCount {
					t.Errorf("Link count mismatch: got %d, want %d",
						len(result.Links), expectedEdgeCount)
				}

				for _, node := range result.Nodes {
					if node.Group == 0 {
						t.Errorf("Node %s has group 0 (should be 1, 2, or 3)", node.ID)
					}
					if node.Group < 0 || node.Group > 3 {
						t.Errorf("Node %s has invalid group %d", node.ID, node.Group)
					}
				}

				for _, link := range result.Links {
					if link.Source == "" {
						t.Error("Link has empty source")
					}
					if link.Target == "" {
						t.Error("Link has empty target")
					}
					if link.Value != 1 {
						t.Errorf("Link has unexpected value %d, want 1", link.Value)
					}
				}
			}
		})
	}
}

func Test_ConvertToD3Format(t *testing.T) {
	tests := []struct {
		name          string
		graph         *graph.DependencyGraph
		expectedNodes int
		expectedLinks int
	}{
		{
			name:          "empty graph",
			graph:         graph.NewDependencyGraph(),
			expectedNodes: 0,
			expectedLinks: 0,
		},
		{
			name: "single node no edges",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
				},
				Edges: make(map[string][]string),
			},
			expectedNodes: 1,
			expectedLinks: 0,
		},
		{
			name: "multiple nodes with edges",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:   "test::func1",
						Name: "func1",
						Kind: graph.KindFunction,
					},
					"test::func2": {
						ID:   "test::func2",
						Name: "func2",
						Kind: graph.KindFunction,
					},
					"test::Type1": {
						ID:   "test::Type1",
						Name: "Type1",
						Kind: graph.KindType,
					},
				},
				Edges: map[string][]string{
					"test::func1": {"test::func2", "test::Type1"},
					"test::func2": {"test::Type1"},
				},
			},
			expectedNodes: 3,
			expectedLinks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToD3Format(tt.graph)

			if len(result.Nodes) != tt.expectedNodes {
				t.Errorf("Node count = %d, want %d", len(result.Nodes), tt.expectedNodes)
			}

			if len(result.Links) != tt.expectedLinks {
				t.Errorf("Link count = %d, want %d", len(result.Links), tt.expectedLinks)
			}
		})
	}
}

func Test_D3JSNode_GroupAssignment(t *testing.T) {
	tests := []struct {
		kind          graph.NodeKind
		expectedGroup int
	}{
		{graph.KindFunction, 1},
		{graph.KindMethod, 2},
		{graph.KindType, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			g := &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::item": {
						ID:   "test::item",
						Name: "item",
						Kind: tt.kind,
					},
				},
				Edges: make(map[string][]string),
			}

			result := convertToD3Format(g)

			if len(result.Nodes) != 1 {
				t.Fatalf("Expected 1 node, got %d", len(result.Nodes))
			}

			if result.Nodes[0].Group != tt.expectedGroup {
				t.Errorf("Group = %d, want %d", result.Nodes[0].Group, tt.expectedGroup)
			}
		})
	}
}

func Test_D3JSGraph_JSONStructure(t *testing.T) {
	testGraph := D3JSGraph{
		Nodes: []D3JSNode{
			{
				ID:        "test::func1",
				Name:      "func1",
				Kind:      "function",
				Package:   "test",
				File:      "test.go",
				Line:      10,
				Signature: "func func1()",
				Group:     1,
			},
		},
		Links: []D3JSLink{
			{
				Source: "test::func1",
				Target: "test::Type1",
				Value:  1,
			},
		},
	}

	data, err := json.Marshal(testGraph)
	if err != nil {
		t.Fatalf("Failed to marshal D3JSGraph: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal D3JSGraph: %v", err)
	}

	if _, ok := result["nodes"]; !ok {
		t.Error("Missing 'nodes' field")
	}
	if _, ok := result["links"]; !ok {
		t.Error("Missing 'links' field")
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok || len(nodes) != 1 {
		t.Errorf("Expected 1 node in array, got %v", result["nodes"])
	}

	links, ok := result["links"].([]interface{})
	if !ok || len(links) != 1 {
		t.Errorf("Expected 1 link in array, got %v", result["links"])
	}
}

func Test_ConvertToD3Format_PackageGrouping(t *testing.T) {
	graph := &graph.DependencyGraph{
		Nodes: map[string]*graph.Node{
			"pkg1::func1": {
				ID:      "pkg1::func1",
				Name:    "func1",
				Kind:    graph.KindFunction,
				Package: "example.com/pkg1",
			},
			"pkg1::Type1": {
				ID:      "pkg1::Type1",
				Name:    "Type1",
				Kind:    graph.KindType,
				Package: "example.com/pkg1",
			},
			"pkg2::func2": {
				ID:      "pkg2::func2",
				Name:    "func2",
				Kind:    graph.KindFunction,
				Package: "example.com/pkg2",
			},
		},
		Edges: map[string][]string{
			"pkg1::func1": {"pkg1::Type1"},
		},
	}

	result := convertToD3Format(graph)

	// Verify packages array exists
	if result.Packages == nil {
		t.Fatal("Packages array is nil")
	}

	// Should have 2 packages
	if len(result.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(result.Packages))
	}

	// Verify package grouping
	packageMap := make(map[string]D3JSPackageGroup)
	for _, pkg := range result.Packages {
		packageMap[pkg.ID] = pkg
	}

	pkg1, ok := packageMap["example.com/pkg1"]
	if !ok {
		t.Error("Package example.com/pkg1 not found")
	}
	if ok && len(pkg1.Nodes) != 2 {
		t.Errorf("Package pkg1 should have 2 nodes, got %d", len(pkg1.Nodes))
	}
	if ok && pkg1.Label != "example.com/pkg1" {
		t.Errorf("Package label mismatch: got %s, want example.com/pkg1", pkg1.Label)
	}

	pkg2, ok := packageMap["example.com/pkg2"]
	if !ok {
		t.Error("Package example.com/pkg2 not found")
	}
	if ok && len(pkg2.Nodes) != 1 {
		t.Errorf("Package pkg2 should have 1 node, got %d", len(pkg2.Nodes))
	}

	// Verify all nodes have package_id set
	for _, node := range result.Nodes {
		if node.PackageID == "" {
			t.Errorf("Node %s has empty package_id", node.ID)
		}
		if node.PackageID != node.Package {
			t.Errorf("Node %s package_id (%s) doesn't match package (%s)",
				node.ID, node.PackageID, node.Package)
		}
	}
}
