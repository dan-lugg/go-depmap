package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"go-depmap/pkg/graph"

	"golang.org/x/tools/go/packages"
)

func Test_New(t *testing.T) {
	tests := []struct {
		name     string
		packages []*packages.Package
	}{
		{
			name:     "empty packages",
			packages: []*packages.Package{},
		},
		{
			name: "single package",
			packages: []*packages.Package{
				{PkgPath: "test"},
			},
		},
		{
			name: "multiple packages",
			packages: []*packages.Package{
				{PkgPath: "test1"},
				{PkgPath: "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(tt.packages)

			if a == nil {
				t.Fatal("New returned nil")
			}

			if a.packages == nil {
				t.Error("packages field is nil")
			}

			if a.projectObjects == nil {
				t.Error("projectObjects field is nil")
			}

			if a.graph == nil {
				t.Error("graph field is nil")
			}

			if len(a.projectObjects) != 0 {
				t.Errorf("Expected empty projectObjects, got %d", len(a.projectObjects))
			}
		})
	}
}

func Test_Analyzer_Analyze_EmptyPackages(t *testing.T) {
	a := New([]*packages.Package{})
	result := a.Analyze()

	if result == nil {
		t.Fatal("Analyze returned nil")
	}

	if len(result.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(result.Nodes))
	}

	if len(result.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(result.Edges))
	}
}

func Test_Analyzer_Analyze_ReturnsGraph(t *testing.T) {
	a := New([]*packages.Package{})
	result := a.Analyze()

	if result == nil {
		t.Fatal("Analyze returned nil graph")
	}

	if result.Nodes == nil {
		t.Error("Graph Nodes is nil")
	}

	if result.Edges == nil {
		t.Error("Graph Edges is nil")
	}
}

func Test_Analyzer_CollectDefinitions_SkipsNonModulePackages(t *testing.T) {
	pkgs := []*packages.Package{
		{
			PkgPath: "test",
			Module:  nil,
		},
	}

	a := New(pkgs)
	a.collectDefinitions()

	if len(a.projectObjects) != 0 {
		t.Errorf("Expected 0 objects from non-module package, got %d", len(a.projectObjects))
	}
}

func Test_Analyzer_ProjectObjects_Initialization(t *testing.T) {
	a := New(nil)

	if a.projectObjects == nil {
		t.Fatal("projectObjects not initialized")
	}

	if len(a.projectObjects) != 0 {
		t.Errorf("Expected empty projectObjects, got %d entries", len(a.projectObjects))
	}
}

func Test_Analyzer_Graph_Initialization(t *testing.T) {
	a := New(nil)

	if a.graph == nil {
		t.Fatal("graph not initialized")
	}

	if a.graph.Nodes == nil {
		t.Error("graph.Nodes not initialized")
	}

	if a.graph.Edges == nil {
		t.Error("graph.Edges not initialized")
	}
}

func Test_Analyzer_AnalyzeCallsCollectAndAnalyze(t *testing.T) {
	a := New([]*packages.Package{})

	if len(a.graph.Nodes) != 0 {
		t.Error("Graph should be empty before Analyze")
	}

	result := a.Analyze()

	if result != a.graph {
		t.Error("Analyze should return the analyzer's graph")
	}
}

func Test_Analyzer_MultipleAnalyzeCalls(t *testing.T) {
	a := New([]*packages.Package{})

	result1 := a.Analyze()
	result2 := a.Analyze()

	if result1 != result2 {
		t.Error("Multiple Analyze calls should return the same graph instance")
	}
}

func Test_Analyzer_PackagesField(t *testing.T) {
	testPkgs := []*packages.Package{
		{PkgPath: "test1"},
		{PkgPath: "test2"},
		{PkgPath: "test3"},
	}

	a := New(testPkgs)

	if len(a.packages) != len(testPkgs) {
		t.Errorf("Expected %d packages, got %d", len(testPkgs), len(a.packages))
	}

	for i, pkg := range a.packages {
		if pkg.PkgPath != testPkgs[i].PkgPath {
			t.Errorf("Package[%d] PkgPath = %s, want %s", i, pkg.PkgPath, testPkgs[i].PkgPath)
		}
	}
}

func Test_Analyzer_MinimalIntegration(t *testing.T) {
	fset := token.NewFileSet()

	pkg := &packages.Package{
		PkgPath: "test",
		Fset:    fset,
		Module:  &packages.Module{Path: "test"},
		TypesInfo: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
			Uses: make(map[*ast.Ident]types.Object),
		},
		Syntax: nil,
	}

	a := New([]*packages.Package{pkg})
	result := a.Analyze()

	if result == nil {
		t.Fatal("Analyze returned nil")
	}

	if len(result.Nodes) != 0 {
		t.Errorf("Expected 0 nodes from package with no syntax trees, got %d", len(result.Nodes))
	}
}

func Test_Analyzer_GraphStructure(t *testing.T) {
	a := New([]*packages.Package{})

	testNode := &graph.Node{
		ID:   "test::func1",
		Name: "func1",
		Kind: graph.KindFunction,
	}

	a.graph.Nodes[testNode.ID] = testNode

	if len(a.graph.Nodes) != 1 {
		t.Errorf("Expected 1 node in graph, got %d", len(a.graph.Nodes))
	}

	retrievedNode, exists := a.graph.Nodes[testNode.ID]
	if !exists {
		t.Error("Node not found in graph")
	}

	if retrievedNode.ID != testNode.ID {
		t.Errorf("Retrieved node ID = %s, want %s", retrievedNode.ID, testNode.ID)
	}
}

func Test_Analyzer_EmptyAnalysis(t *testing.T) {
	a := New([]*packages.Package{})
	result := a.Analyze()

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Nodes == nil {
		t.Error("Nodes map should not be nil")
	}

	if result.Edges == nil {
		t.Error("Edges map should not be nil")
	}

	if len(result.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(result.Nodes))
	}

	if result.CountEdges() != 0 {
		t.Errorf("Expected 0 edges, got %d", result.CountEdges())
	}
}
