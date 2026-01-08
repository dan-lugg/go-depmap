package format

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

	"go-depmap/pkg/graph"
)

//go:embed templates/d3js.html
var templateFS embed.FS

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

// D3JSGroup represents a hierarchical group for WebCola constraint-based layout
type D3JSGroup struct {
	ID      string `json:"id"`               // Unique identifier for the group
	Label   string `json:"label"`            // Display label
	Leaves  []int  `json:"leaves,omitempty"` // Indices of nodes in this group
	Groups  []int  `json:"groups,omitempty"` // Indices of nested groups
	Level   string `json:"level"`            // "package" or "type"
	Padding int    `json:"padding"`          // Padding around the group in pixels
}

// D3JSGraph is the D3.js compatible graph structure with hierarchical grouping
type D3JSGraph struct {
	Nodes  []D3JSNode  `json:"nodes"`
	Links  []D3JSLink  `json:"links"`
	Groups []D3JSGroup `json:"groups,omitempty"` // Hierarchical groups for WebCola layout
}

// D3JSWriter writes the graph in D3.js force-directed graph format
type D3JSWriter struct{}

func (w *D3JSWriter) Write(writer io.Writer, depGraph *graph.DependencyGraph, config Config) error {
	// Check grouping options (all default to true)
	groupByPackage := config.GetBool("groupByPackage", true) // WebCola package grouping
	groupByType := config.GetBool("groupByType", true)       // WebCola type-level grouping

	d3Graph := convertToD3Format(depGraph, groupByPackage, groupByType)

	// Check if HTML page output is requested
	if config.GetBool("htmlPage", false) {
		return writeHTMLPage(writer, d3Graph)
	}

	// Otherwise output JSON
	enc := json.NewEncoder(writer)

	// Check if pretty printing is enabled (defaults to true)
	if config.GetBool("pretty", true) {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(d3Graph)
}

// convertToD3Format converts a DependencyGraph to D3.js format with optional package grouping
func convertToD3Format(depGraph *graph.DependencyGraph, groupByPackage bool, groupByType bool) *D3JSGraph {
	d3Graph := &D3JSGraph{
		Nodes:  make([]D3JSNode, 0, len(depGraph.Nodes)),
		Links:  make([]D3JSLink, 0),
		Groups: make([]D3JSGroup, 0),
	}

	// Map to assign group numbers based on kind
	kindToGroup := map[string]int{
		"function": 1,
		"method":   2,
		"type":     3,
	}

	// Maps for tracking grouping
	packageNodes := make(map[string][]string)                // package -> node IDs
	nodeIndexMap := make(map[string]int)                     // node ID -> array index
	packageTypeNodes := make(map[string]map[string][]string) // package -> type -> node IDs
	typeToPackage := make(map[string]string)                 // type -> package

	// Convert nodes and build index maps
	for _, node := range depGraph.Nodes {
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
			PackageID: node.Package,
		}

		nodeIndex := len(d3Graph.Nodes)
		d3Graph.Nodes = append(d3Graph.Nodes, d3Node)
		nodeIndexMap[node.ID] = nodeIndex

		// Track nodes by package
		packageNodes[node.Package] = append(packageNodes[node.Package], node.ID)

		// Track methods by their receiver type
		if node.Kind == graph.KindMethod {
			// Extract receiver type from method name (format: "(*Type).method" or "Type.method")
			receiverType := extractReceiverType(node.Name)
			if receiverType != "" {
				if packageTypeNodes[node.Package] == nil {
					packageTypeNodes[node.Package] = make(map[string][]string)
				}
				packageTypeNodes[node.Package][receiverType] = append(packageTypeNodes[node.Package][receiverType], node.ID)
				typeToPackage[receiverType] = node.Package
			}
		}

		// Track type declarations
		if node.Kind == graph.KindType {
			typeToPackage[node.Name] = node.Package
		}
	}

	// Convert edges
	for sourceID, targets := range depGraph.Edges {
		for _, targetID := range targets {
			d3Graph.Links = append(d3Graph.Links, D3JSLink{
				Source: sourceID,
				Target: targetID,
				Value:  1,
			})
		}
	}

	// Build WebCola-compatible hierarchical groups
	if groupByPackage {
		for pkgName, nodeIDs := range packageNodes {
			// Collect leaf nodes (non-method nodes or methods without type grouping)
			var packageLeaves []int
			var nestedTypeGroupIndices []int

			// If type grouping is enabled, separate methods by type
			if groupByType && packageTypeNodes[pkgName] != nil {
				// Add non-method nodes as direct leaves
				for _, nodeID := range nodeIDs {
					idx := nodeIndexMap[nodeID]
					node := d3Graph.Nodes[idx]
					if node.Kind != "method" {
						packageLeaves = append(packageLeaves, idx)
					}
				}

				// Create type groups for methods
				for typeName, methodIDs := range packageTypeNodes[pkgName] {
					if len(methodIDs) > 0 {
						// Get indices for methods
						var typeLeaves []int
						for _, methodID := range methodIDs {
							if idx, ok := nodeIndexMap[methodID]; ok {
								typeLeaves = append(typeLeaves, idx)
							}
						}

						// Store the index where this type group will be added
						typeGroupIndex := len(d3Graph.Groups)

						// Add type group
						d3Graph.Groups = append(d3Graph.Groups, D3JSGroup{
							ID:      pkgName + "::" + typeName,
							Label:   typeName,
							Leaves:  typeLeaves,
							Level:   "type",
							Padding: 50, // Increased from 30 to 50 for even better spacing
						})

						nestedTypeGroupIndices = append(nestedTypeGroupIndices, typeGroupIndex)
					}
				}
			} else {
				// No type grouping - all nodes are direct leaves
				for _, nodeID := range nodeIDs {
					if idx, ok := nodeIndexMap[nodeID]; ok {
						packageLeaves = append(packageLeaves, idx)
					}
				}
			}

			// Add package group
			packageGroup := D3JSGroup{
				ID:      pkgName,
				Label:   pkgName,
				Leaves:  packageLeaves,
				Groups:  nestedTypeGroupIndices,
				Level:   "package",
				Padding: 80, // Increased from 50 to 80 for even better spacing
			}
			d3Graph.Groups = append(d3Graph.Groups, packageGroup)
		}
	}

	return d3Graph
}

// extractReceiverType extracts the receiver type name from a method name
// Handles formats: "(*Type).method" or "Type.method"
func extractReceiverType(methodName string) string {
	// Look for pattern: (optional *)(Type).method
	if idx := len(methodName); idx > 0 {
		// Find the first dot (method separator)
		dotIdx := -1
		parenDepth := 0
		for i, ch := range methodName {
			if ch == '(' {
				parenDepth++
			} else if ch == ')' {
				parenDepth--
			} else if ch == '.' && parenDepth == 0 {
				dotIdx = i
				break
			}
		}

		if dotIdx > 0 {
			receiver := methodName[:dotIdx]
			// Remove (*  or (
			receiver = trimReceiverParens(receiver)
			return receiver
		}
	}
	return ""
}

// trimReceiverParens removes parentheses and pointer markers from receiver
func trimReceiverParens(receiver string) string {
	// Remove leading (* or (
	if len(receiver) > 0 && receiver[0] == '(' {
		receiver = receiver[1:]
		if len(receiver) > 0 && receiver[0] == '*' {
			receiver = receiver[1:]
		}
	}
	// Remove trailing )
	if len(receiver) > 0 && receiver[len(receiver)-1] == ')' {
		receiver = receiver[:len(receiver)-1]
	}
	return receiver
}

// writeHTMLPage generates a self-contained HTML page with embedded D3.js/WebCola visualization
func writeHTMLPage(writer io.Writer, d3Graph *D3JSGraph) error {
	// Parse the embedded template
	tmpl, err := template.ParseFS(templateFS, "templates/d3js.html")
	if err != nil {
		return err
	}

	// Marshal the graph data to JSON
	jsonData, err := json.Marshal(d3Graph)
	if err != nil {
		return err
	}

	// Prepare template data
	data := struct {
		Data template.JS
	}{
		Data: template.JS(jsonData),
	}

	// Execute the template
	return tmpl.Execute(writer, data)
}
