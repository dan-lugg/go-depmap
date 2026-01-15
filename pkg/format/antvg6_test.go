package format

import (
	"bytes"
	"testing"

	"go-depmap/pkg/graph"
)

func TestAntVG6Writer_Write_JSON(t *testing.T) {
	// Create a simple test graph
	depGraph := graph.NewDependencyGraph()
	depGraph.Nodes["test.Package::TestFunc"] = &graph.Node{
		ID:      "test.Package::TestFunc",
		Name:    "TestFunc",
		Package: "test.Package",
		Kind:    graph.KindFunction,
	}

	writer := &AntVG6Writer{}
	var buf bytes.Buffer
	config := Config{"pretty": true}

	err := writer.Write(&buf, depGraph, config)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// Check for key elements
	if !bytes.Contains([]byte(output), []byte("nodes")) {
		t.Error("Output should contain 'nodes'")
	}
	if !bytes.Contains([]byte(output), []byte("edges")) {
		t.Error("Output should contain 'edges'")
	}
	if !bytes.Contains([]byte(output), []byte("combos")) {
		t.Error("Output should contain 'combos'")
	}
}

func TestAntVG6Writer_Write_HTML(t *testing.T) {
	depGraph := graph.NewDependencyGraph()
	depGraph.Nodes["test.Package::TestFunc"] = &graph.Node{
		ID:      "test.Package::TestFunc",
		Name:    "TestFunc",
		Package: "test.Package",
		Kind:    graph.KindFunction,
	}

	writer := &AntVG6Writer{}
	var buf bytes.Buffer
	config := Config{"htmlPage": true}

	err := writer.Write(&buf, depGraph, config)
	if err != nil {
		t.Fatalf("Write HTML failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected non-empty HTML output")
	}

	// Check for HTML elements
	if !bytes.Contains([]byte(output), []byte("<!DOCTYPE html>")) {
		t.Error("Output should contain HTML doctype")
	}
	if !bytes.Contains([]byte(output), []byte("AntV G6")) {
		t.Error("Output should reference AntV G6")
	}
	if !bytes.Contains([]byte(output), []byte("@antv/g6")) {
		t.Error("Output should import @antv/g6 library")
	}
}

func TestConvertToAntVG6Format(t *testing.T) {
	depGraph := graph.NewDependencyGraph()

	// Add a type
	depGraph.Nodes["pkg.Type"] = &graph.Node{
		ID:      "pkg.Type",
		Name:    "Type",
		Package: "pkg",
		Kind:    graph.KindType,
	}

	// Add a function
	depGraph.Nodes["pkg.Func"] = &graph.Node{
		ID:      "pkg.Func",
		Name:    "Func",
		Package: "pkg",
		Kind:    graph.KindFunction,
	}

	// Add a dependency
	depGraph.Edges["pkg.Func"] = []string{"pkg.Type"}

	result := convertToAntVG6Format(depGraph, Config{})

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes in output")
	}
	if len(result.Edges) == 0 {
		t.Error("Expected edges in output")
	}
	if len(result.Combos) == 0 {
		t.Error("Expected combos (package containers) in output")
	}

	// Verify combo exists for package
	foundCombo := false
	for _, combo := range result.Combos {
		if combo.ID == "pkg:pkg" {
			foundCombo = true
			break
		}
	}
	if !foundCombo {
		t.Error("Expected combo for package 'pkg'")
	}

	// Verify nodes have comboId
	for _, node := range result.Nodes {
		if comboID, ok := node.Data["comboId"].(string); !ok || comboID == "" {
			t.Errorf("Node %s should have comboId", node.ID)
		}
	}
}
