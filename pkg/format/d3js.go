package format

import (
	"encoding/json"
	"io"

	"go-depmap/pkg/graph"
)

// D3JSNode represents a node in D3.js force-directed graph format
type D3JSNode struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Package   string `json:"package"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Signature string `json:"signature"`
	Group     int    `json:"group"` // For coloring by kind
}

// D3JSLink represents an edge in D3.js force-directed graph format
type D3JSLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"` // Weight of the edge (can be used for styling)
}

// D3JSGraph is the D3.js compatible graph structure
type D3JSGraph struct {
	Nodes []D3JSNode `json:"nodes"`
	Links []D3JSLink `json:"links"`
}

// D3JSJSONWriter writes the graph in D3.js force-directed graph format
type D3JSJSONWriter struct{}

func (w *D3JSJSONWriter) Write(writer io.Writer, graph *graph.DependencyGraph) error {
	d3Graph := convertToD3Format(graph)
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	return enc.Encode(d3Graph)
}

// convertToD3Format converts a DependencyGraph to D3.js format
func convertToD3Format(graph *graph.DependencyGraph) *D3JSGraph {
	d3Graph := &D3JSGraph{
		Nodes: make([]D3JSNode, 0, len(graph.Nodes)),
		Links: make([]D3JSLink, 0),
	}

	// Map to assign group numbers based on kind
	kindToGroup := map[string]int{
		"function": 1,
		"method":   2,
		"type":     3,
	}

	// Convert nodes
	for _, node := range graph.Nodes {
		group := kindToGroup[string(node.Kind)]
		d3Graph.Nodes = append(d3Graph.Nodes, D3JSNode{
			ID:        node.ID,
			Name:      node.Name,
			Kind:      string(node.Kind),
			Package:   node.Package,
			File:      node.File,
			Line:      node.Line,
			Signature: node.Signature,
			Group:     group,
		})
	}

	// Convert edges
	for sourceID, targets := range graph.Edges {
		for _, targetID := range targets {
			d3Graph.Links = append(d3Graph.Links, D3JSLink{
				Source: sourceID,
				Target: targetID,
				Value:  1, // Default weight
			})
		}
	}

	return d3Graph
}
