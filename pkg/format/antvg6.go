package format

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

	"go-depmap/pkg/graph"
)

//go:embed templates/antvg6.html
var antvg6TemplateFS embed.FS

// AntVG6Writer implements the Writer interface for AntV G6 visualization
type AntVG6Writer struct{}

// AntVG6Node represents a node in AntV G6 v4 format
type AntVG6Node struct {
	ID      string                 `json:"id"`
	Label   string                 `json:"label,omitempty"`
	ComboID string                 `json:"comboId,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// AntVG6Edge represents an edge in AntV G6 format
type AntVG6Edge struct {
	ID     string                 `json:"id"`
	Source string                 `json:"source"`
	Target string                 `json:"target"`
	Data   map[string]interface{} `json:"data"`
}

// AntVG6Combo represents a combo (package container) in AntV G6 v4 format
type AntVG6Combo struct {
	ID    string                 `json:"id"`
	Label string                 `json:"label,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// AntVG6Graph is the complete data structure for AntV G6
type AntVG6Graph struct {
	Nodes  []AntVG6Node  `json:"nodes"`
	Edges  []AntVG6Edge  `json:"edges"`
	Combos []AntVG6Combo `json:"combos,omitempty"`
}

// Write generates AntV G6-compatible JSON or HTML output
func (w *AntVG6Writer) Write(writer io.Writer, depGraph *graph.DependencyGraph, config Config) error {
	antvg6Graph := convertToAntVG6Format(depGraph, config)

	// Check if HTML page should be generated
	if config.GetBool("htmlPage", false) {
		return writeAntVG6HTML(writer, antvg6Graph)
	}

	// Otherwise, output JSON
	var jsonData []byte
	var err error

	if config.GetBool("pretty", true) {
		jsonData, err = json.MarshalIndent(antvg6Graph, "", "  ")
	} else {
		jsonData, err = json.Marshal(antvg6Graph)
	}

	if err != nil {
		return err
	}

	_, err = writer.Write(jsonData)
	return err
}

// convertToAntVG6Format converts DependencyGraph to AntV G6 format with package combos
func convertToAntVG6Format(depGraph *graph.DependencyGraph, _ Config) *AntVG6Graph {
	antvg6Graph := &AntVG6Graph{
		Nodes:  make([]AntVG6Node, 0),
		Edges:  make([]AntVG6Edge, 0),
		Combos: make([]AntVG6Combo, 0),
	}

	// Track which package combos we've created
	packageCombos := make(map[string]bool)
	typeHubs := make(map[string]bool)

	// Color palette for packages
	packageColors := make(map[string]string)
	colorIndex := 0

	// Helper to generate package color
	getPackageColor := func(pkgName string) string {
		if color, exists := packageColors[pkgName]; exists {
			return color
		}
		hue := (colorIndex * 137) % 360
		colorIndex++
		packageColors[pkgName] = hslToHex(hue, 70, 50)
		return packageColors[pkgName]
	}

	// Helper to lighten color
	lightenColor := func(hexColor string, amount int) string {
		h, s, l := hexToHSL(hexColor)
		return hslToHex(h, s, l+amount)
	}

	// Phase 1: Create package combos (containers)
	for _, node := range depGraph.Nodes {
		if !packageCombos[node.Package] {
			packageCombos[node.Package] = true
			pkgColor := getPackageColor(node.Package)
			antvg6Graph.Combos = append(antvg6Graph.Combos, AntVG6Combo{
				ID:    "pkg:" + node.Package,
				Label: node.Package,
				Data: map[string]interface{}{
					"color":       "rgba(100, 100, 200, 0.05)",
					"strokeColor": lightenColor(pkgColor, 20),
				},
			})
		}
	}

	// Phase 2: Create type nodes (not as combos, but as regular nodes)
	for _, node := range depGraph.Nodes {
		if node.Kind == graph.KindType {
			typeID := "type:" + node.ID
			if !typeHubs[typeID] {
				typeHubs[typeID] = true
				pkgColor := getPackageColor(node.Package)
				antvg6Graph.Nodes = append(antvg6Graph.Nodes, AntVG6Node{
					ID:      typeID,
					Label:   node.Name,
					ComboID: "pkg:" + node.Package,
					Data: map[string]interface{}{
						"type":  "type",
						"group": node.Package,
						"color": lightenColor(pkgColor, 15),
						"size":  8.0,
					},
				})
				// Note: No structural edge to package - combo provides visual grouping
			}
		}
	}

	// Phase 3: Create function/method nodes
	for _, node := range depGraph.Nodes {
		var nodeType string
		var nodeSize float64
		pkgColor := getPackageColor(node.Package)

		switch node.Kind {
		case graph.KindFunction:
			nodeType = "function"
			nodeSize = 4.0
		case graph.KindMethod:
			nodeType = "method"
			nodeSize = 4.0
		case graph.KindType:
			// Already added, skip
			continue
		default:
			nodeType = "unknown"
			nodeSize = 4.0
		}

		antvg6Graph.Nodes = append(antvg6Graph.Nodes, AntVG6Node{
			ID:      node.ID,
			Label:   node.Name,
			ComboID: "pkg:" + node.Package,
			Data: map[string]interface{}{
				"type":  nodeType,
				"group": node.Package,
				"color": pkgColor,
				"size":  nodeSize,
			},
		})
		// Note: No structural edges - combo provides visual grouping
	}

	// Phase 4: Add dependency edges (only between actual nodes that exist)
	nodeExists := make(map[string]bool)
	for _, node := range antvg6Graph.Nodes {
		nodeExists[node.ID] = true
	}

	// Track edges to prevent duplicates
	edgeExists := make(map[string]bool)

	for sourceID, targets := range depGraph.Edges {
		// Check if source exists in our node list
		if !nodeExists[sourceID] {
			continue
		}

		for _, targetID := range targets {
			// Check if target exists in our node list
			if !nodeExists[targetID] {
				continue
			}

			// Create edge ID and check if it already exists
			edgeID := sourceID + "->" + targetID
			if edgeExists[edgeID] {
				continue // Skip duplicate edge
			}
			edgeExists[edgeID] = true

			antvg6Graph.Edges = append(antvg6Graph.Edges, AntVG6Edge{
				ID:     edgeID,
				Source: sourceID,
				Target: targetID,
				Data: map[string]interface{}{
					"linkType": "dependency",
				},
			})
		}
	}

	return antvg6Graph
}

// writeAntVG6HTML generates a self-contained HTML page with embedded AntV G6
func writeAntVG6HTML(writer io.Writer, antvg6Graph *AntVG6Graph) error {
	// Parse the embedded template
	tmpl, err := template.ParseFS(antvg6TemplateFS, "templates/antvg6.html")
	if err != nil {
		return err
	}

	// Marshal the graph data to JSON
	jsonData, err := json.Marshal(antvg6Graph)
	if err != nil {
		return err
	}

	// Prepare template data
	data := struct {
		Data template.JS
	}{
		Data: template.JS(jsonData), // #nosec G203 - JSON data is safe, we control the marshaling
	}

	// Execute the template
	return tmpl.Execute(writer, data)
}
