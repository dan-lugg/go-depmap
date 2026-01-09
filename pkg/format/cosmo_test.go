package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go-depmap/pkg/graph"
)

func TestCosmoWriter_Write_JSON(t *testing.T) {
	// Create a simple test graph
	g := &graph.DependencyGraph{
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
			"pkg1::(*Type1).Method1": {
				ID:      "pkg1::(*Type1).Method1",
				Name:    "(*Type1).Method1",
				Kind:    graph.KindMethod,
				Package: "example.com/pkg1",
			},
		},
		Edges: map[string][]string{
			"pkg1::func1": {"pkg1::Type1"},
		},
	}

	w := &CosmoWriter{}
	var buf bytes.Buffer
	config := Config{"pretty": true}

	err := w.Write(&buf, g, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Parse the output
	var result CosmoGraph
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify we have nodes
	if len(result.Nodes) == 0 {
		t.Error("Expected nodes in output")
	}

	// Verify we have links
	if len(result.Links) == 0 {
		t.Error("Expected links in output")
	}

	// Verify hub nodes were created
	foundPackageHub := false
	foundTypeHub := false
	for _, node := range result.Nodes {
		if node.Type == "package" && strings.Contains(node.ID, "pkg:") {
			foundPackageHub = true
		}
		if node.Type == "type" && strings.Contains(node.ID, "type:") {
			foundTypeHub = true
		}
	}

	if !foundPackageHub {
		t.Error("Expected package hub node")
	}
	if !foundTypeHub {
		t.Error("Expected type hub node")
	}
}

func TestCosmoWriter_Write_HTML(t *testing.T) {
	g := &graph.DependencyGraph{
		Nodes: map[string]*graph.Node{
			"pkg1::func1": {
				ID:      "pkg1::func1",
				Name:    "func1",
				Kind:    graph.KindFunction,
				Package: "example.com/pkg1",
			},
		},
		Edges: map[string][]string{},
	}

	w := &CosmoWriter{}
	var buf bytes.Buffer
	config := Config{"htmlPage": true}

	err := w.Write(&buf, g, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	html := buf.String()

	// Verify it's HTML
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Expected HTML doctype")
	}

	// Verify Cosmograph is loaded
	if !strings.Contains(html, "cosmograph") {
		t.Error("Expected Cosmograph library reference")
	}

	// Verify data is embedded
	if !strings.Contains(html, "const data") {
		t.Error("Expected embedded data")
	}

	// Verify nodes and links arrays exist in embedded data
	if !strings.Contains(html, "\"nodes\"") {
		t.Error("Expected nodes array in embedded data")
	}
	if !strings.Contains(html, "\"links\"") {
		t.Error("Expected links array in embedded data")
	}
}

func TestCosmoWriter_HubSpoke_Topology(t *testing.T) {
	// Test the Hub & Spoke topology creation
	g := &graph.DependencyGraph{
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
			"pkg1::(*Type1).Method1": {
				ID:      "pkg1::(*Type1).Method1",
				Name:    "(*Type1).Method1",
				Kind:    graph.KindMethod,
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

	w := &CosmoWriter{}
	var buf bytes.Buffer
	config := Config{"pretty": true}

	err := w.Write(&buf, g, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var result CosmoGraph
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Count different node types
	var packageHubs, typeHubs, funcNodes, methodNodes int
	for _, node := range result.Nodes {
		switch node.Type {
		case "package":
			packageHubs++
		case "type":
			typeHubs++
		case "function":
			funcNodes++
		case "method":
			methodNodes++
		}
	}

	// Should have 2 packages
	if packageHubs != 2 {
		t.Errorf("Expected 2 package hubs, got %d", packageHubs)
	}

	// Should have 1 type hub
	if typeHubs != 1 {
		t.Errorf("Expected 1 type hub, got %d", typeHubs)
	}

	// Should have 2 functions
	if funcNodes != 2 {
		t.Errorf("Expected 2 function nodes, got %d", funcNodes)
	}

	// Should have 1 method
	if methodNodes != 1 {
		t.Errorf("Expected 1 method node, got %d", methodNodes)
	}

	// Verify structural links (nodes to hubs)
	// Each non-hub node should link to its parent hub
	foundStructuralLinks := 0
	for _, link := range result.Links {
		if strings.HasPrefix(link.Target, "pkg:") || strings.HasPrefix(link.Target, "type:") {
			foundStructuralLinks++
		}
	}

	if foundStructuralLinks < 3 {
		t.Errorf("Expected at least 3 structural links to hubs, got %d", foundStructuralLinks)
	}
}

func TestCosmoWriter_NodeSizing(t *testing.T) {
	g := &graph.DependencyGraph{
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
		},
		Edges: map[string][]string{},
	}

	w := &CosmoWriter{}
	var buf bytes.Buffer
	config := Config{}

	err := w.Write(&buf, g, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var result CosmoGraph
	json.Unmarshal(buf.Bytes(), &result)

	// Verify node sizes follow the hub & spoke pattern
	for _, node := range result.Nodes {
		switch node.Type {
		case "package":
			if node.Size != 10.0 {
				t.Errorf("Package hub should have size 10.0, got %f", node.Size)
			}
		case "type":
			if node.Size != 5.0 {
				t.Errorf("Type hub should have size 5.0, got %f", node.Size)
			}
		case "function", "method":
			if node.Size != 2.0 {
				t.Errorf("Function/Method should have size 2.0, got %f", node.Size)
			}
		}
	}
}

func TestCosmoWriter_ColorCoding(t *testing.T) {
	g := &graph.DependencyGraph{
		Nodes: map[string]*graph.Node{
			"pkg1::func1": {
				ID:      "pkg1::func1",
				Name:    "func1",
				Kind:    graph.KindFunction,
				Package: "example.com/pkg1",
			},
		},
		Edges: map[string][]string{},
	}

	w := &CosmoWriter{}
	var buf bytes.Buffer
	config := Config{}

	err := w.Write(&buf, g, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var result CosmoGraph
	json.Unmarshal(buf.Bytes(), &result)

	// Verify all nodes have colors
	for _, node := range result.Nodes {
		if node.Color == "" {
			t.Errorf("Node %s should have a color", node.ID)
		}
		// Verify color is in hex format
		if !strings.HasPrefix(node.Color, "#") {
			t.Errorf("Node color should be hex format, got %s", node.Color)
		}
	}
}
