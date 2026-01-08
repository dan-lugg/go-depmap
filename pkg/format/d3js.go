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
	Group     int    `json:"group"`      // For coloring by kind
	PackageID string `json:"package_id"` // Fully qualified package name for grouping
}

// D3JSLink represents an edge in D3.js force-directed graph format
type D3JSLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"` // Weight of the edge (can be used for styling)
}

// D3JSPackageGroup represents a package group for convex hull visualization
type D3JSPackageGroup struct {
	ID    string   `json:"id"`    // Fully qualified package name
	Label string   `json:"label"` // Display label for the package
	Nodes []string `json:"nodes"` // IDs of nodes belonging to this package
}

// D3JSGraph is the D3.js compatible graph structure with package grouping
type D3JSGraph struct {
	Nodes    []D3JSNode         `json:"nodes"`
	Links    []D3JSLink         `json:"links"`
	Packages []D3JSPackageGroup `json:"packages"` // Package groups for convex hull rendering
}

// D3JSJSONWriter writes the graph in D3.js force-directed graph format
type D3JSJSONWriter struct{}

func (w *D3JSJSONWriter) Write(writer io.Writer, graph *graph.DependencyGraph) error {
	d3Graph := convertToD3Format(graph)
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	return enc.Encode(d3Graph)
}

// convertToD3Format converts a DependencyGraph to D3.js format with package grouping
func convertToD3Format(graph *graph.DependencyGraph) *D3JSGraph {
	d3Graph := &D3JSGraph{
		Nodes:    make([]D3JSNode, 0, len(graph.Nodes)),
		Links:    make([]D3JSLink, 0),
		Packages: make([]D3JSPackageGroup, 0),
	}

	// Map to assign group numbers based on kind
	kindToGroup := map[string]int{
		"function": 1,
		"method":   2,
		"type":     3,
	}

	// Map to track packages and their nodes
	packageNodes := make(map[string][]string)

	// Convert nodes
	for _, node := range graph.Nodes {
		group := kindToGroup[string(node.Kind)]
		d3Node := D3JSNode{
			ID:        node.ID,
			Name:      node.Name,
			Kind:      string(node.Kind),
			Package:   node.Package,
			File:      node.File,
			Line:      node.Line,
			Signature: node.Signature,
			Group:     group,
			PackageID: node.Package, // Use fully qualified package name
		}
		d3Graph.Nodes = append(d3Graph.Nodes, d3Node)

		// Track which nodes belong to which package
		packageNodes[node.Package] = append(packageNodes[node.Package], node.ID)
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

	// Build package groups for convex hull rendering
	for pkgName, nodeIDs := range packageNodes {
		d3Graph.Packages = append(d3Graph.Packages, D3JSPackageGroup{
			ID:    pkgName,
			Label: pkgName,
			Nodes: nodeIDs,
		})
	}

	return d3Graph
}
