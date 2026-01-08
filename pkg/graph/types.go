// Package graph provides types and utilities for representing code dependency graphs.
package graph

// NodeKind represents the type of a code element (function, method, or type)
type NodeKind string

// Node kind constants define the different types of code elements that can appear in the dependency graph.
const (
	KindFunction NodeKind = "function"
	KindMethod   NodeKind = "method"
	KindType     NodeKind = "type"
)

// Node represents a code element in the dependency graph
type Node struct {
	ID        string   `json:"id"`        // Unique signature
	Name      string   `json:"name"`      // Short name
	Kind      NodeKind `json:"kind"`      // function, method, or type
	Package   string   `json:"package"`   // Import path
	File      string   `json:"file"`      // Source filename
	Line      int      `json:"line"`      // Line number
	Signature string   `json:"signature"` // Human readable signature
}

// DependencyGraph represents the complete dependency graph with nodes and edges
type DependencyGraph struct {
	Nodes map[string]*Node    `json:"nodes"`
	Edges map[string][]string `json:"edges"` // SourceID -> []TargetIDs
}

// NewDependencyGraph creates a new empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
	}
}

// CountEdges returns the total number of edges in the graph
func (g *DependencyGraph) CountEdges() int {
	count := 0
	for _, targets := range g.Edges {
		count += len(targets)
	}
	return count
}
