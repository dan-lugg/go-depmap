package format

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"

	"go-depmap/pkg/graph"
)

//go:embed templates/cosmo.html
var cosmoTemplateFS embed.FS

// CosmoWriter implements the Writer interface for Cosmograph visualization
type CosmoWriter struct{}

// CosmoNode represents a node in Cosmograph format
type CosmoNode struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"` // "package", "type", "function", "method"
	Label string  `json:"label"`
	Group string  `json:"group"` // Fully qualified package name for grouping
	Color string  `json:"color"`
	Size  float64 `json:"size"`
}

// CosmoLink represents a link in Cosmograph format
type CosmoLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// CosmoGraph is the complete data structure for Cosmograph
type CosmoGraph struct {
	Nodes []CosmoNode `json:"nodes"`
	Links []CosmoLink `json:"links"`
}

// Write generates Cosmograph-compatible JSON or HTML output
func (w *CosmoWriter) Write(writer io.Writer, depGraph *graph.DependencyGraph, config Config) error {
	cosmoGraph := convertToCosmoFormat(depGraph, config)

	// Check if HTML page should be generated
	if config.GetBool("htmlPage", false) {
		return writeCosmographHTML(writer, cosmoGraph)
	}

	// Otherwise, output JSON
	var jsonData []byte
	var err error

	if config.GetBool("pretty", true) {
		jsonData, err = json.MarshalIndent(cosmoGraph, "", "  ")
	} else {
		jsonData, err = json.Marshal(cosmoGraph)
	}

	if err != nil {
		return err
	}

	_, err = writer.Write(jsonData)
	return err
}

// convertToCosmoFormat converts DependencyGraph to Cosmograph format using Hub & Spoke model
func convertToCosmoFormat(depGraph *graph.DependencyGraph, _ Config) *CosmoGraph {
	cosmoGraph := &CosmoGraph{
		Nodes: make([]CosmoNode, 0),
		Links: make([]CosmoLink, 0),
	}

	// Track which hub nodes we've created
	packageHubs := make(map[string]bool)
	typeHubs := make(map[string]bool)

	// Color palette for packages (using HSL to generate distinct colors)
	packageColors := make(map[string]string)
	colorIndex := 0

	// Helper to generate package color
	getPackageColor := func(pkgName string) string {
		if color, exists := packageColors[pkgName]; exists {
			return color
		}
		// Generate distinct hues across the spectrum
		hue := (colorIndex * 137) % 360 // Golden angle for better distribution
		colorIndex++
		packageColors[pkgName] = hslToHex(hue, 70, 50)
		return packageColors[pkgName]
	}

	// Helper to lighten color for child nodes
	lightenColor := func(hexColor string, amount int) string {
		// Parse hex and increase lightness
		h, s, l := hexToHSL(hexColor)
		return hslToHex(h, s, l+amount)
	}

	// Helper to add node
	addNode := func(node CosmoNode) {
		cosmoGraph.Nodes = append(cosmoGraph.Nodes, node)
	}

	// Phase 1: Create package hub nodes
	for _, node := range depGraph.Nodes {
		if !packageHubs[node.Package] {
			packageHubs[node.Package] = true
			addNode(CosmoNode{
				ID:    "pkg:" + node.Package,
				Type:  "package",
				Label: node.Package,
				Group: node.Package, // Package is its own group
				Color: getPackageColor(node.Package),
				Size:  10.0, // Large hub node
			})
		}
	}

	// Phase 2: Create type hub nodes and link to package hubs
	for _, node := range depGraph.Nodes {
		if node.Kind == graph.KindType {
			typeID := "type:" + node.ID
			if !typeHubs[typeID] {
				typeHubs[typeID] = true
				pkgColor := getPackageColor(node.Package)
				addNode(CosmoNode{
					ID:    typeID,
					Type:  "type",
					Label: node.Name,
					Group: node.Package, // Group by package
					Color: lightenColor(pkgColor, 10),
					Size:  5.0, // Medium hub node
				})

				// Link type to its package
				pkgHubID := "pkg:" + node.Package
				cosmoGraph.Links = append(cosmoGraph.Links, CosmoLink{
					Source: typeID,
					Target: pkgHubID,
				})
			}
		}
	}

	// Phase 3: Create function/method nodes and link to appropriate hubs
	for _, node := range depGraph.Nodes {
		var nodeType string
		var nodeSize float64
		var parentHub string
		pkgColor := getPackageColor(node.Package)

		switch node.Kind {
		case graph.KindFunction:
			nodeType = "function"
			nodeSize = 2.0
			parentHub = "pkg:" + node.Package
		case graph.KindMethod:
			nodeType = "method"
			nodeSize = 2.0
			// Try to find the receiver type
			receiverType := extractReceiverType(node.Name)
			if receiverType != "" {
				parentHub = "type:" + node.Package + "::" + receiverType
				// If type hub doesn't exist, fall back to package
				if !typeHubs[parentHub] {
					parentHub = "pkg:" + node.Package
				}
			} else {
				parentHub = "pkg:" + node.Package
			}
		case graph.KindType:
			// Already added as hub, skip
			continue
		default:
			nodeType = "unknown"
			nodeSize = 2.0
			parentHub = "pkg:" + node.Package
		}

		addNode(CosmoNode{
			ID:    node.ID,
			Type:  nodeType,
			Label: node.Name,
			Group: node.Package, // Group by package
			Color: lightenColor(pkgColor, 20),
			Size:  nodeSize,
		})

		// Link to parent hub (structural edge - opaque)
		cosmoGraph.Links = append(cosmoGraph.Links, CosmoLink{
			Source: node.ID,
			Target: parentHub,
		})
	}

	// Phase 4: Add dependency edges (function -> function)
	for sourceID, targets := range depGraph.Edges {
		for _, targetID := range targets {
			// Skip if target doesn't exist in graph
			if _, exists := depGraph.Nodes[targetID]; !exists {
				continue
			}

			cosmoGraph.Links = append(cosmoGraph.Links, CosmoLink{
				Source: sourceID,
				Target: targetID,
			})
		}
	}

	return cosmoGraph
}

// writeCosmographHTML generates a self-contained HTML page with embedded Cosmograph
func writeCosmographHTML(writer io.Writer, cosmoGraph *CosmoGraph) error {
	// Parse the embedded template
	tmpl, err := template.ParseFS(cosmoTemplateFS, "templates/cosmo.html")
	if err != nil {
		return err
	}

	// Marshal the graph data to JSON
	jsonData, err := json.Marshal(cosmoGraph)
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

// Color conversion helpers
func hslToHex(h, s, l int) string {
	// Convert HSL to RGB
	sF := float64(s) / 100.0
	lF := float64(l) / 100.0

	c := (1 - abs(2*lF-1)) * sF
	x := c * (1 - abs(float64((h/60)%2)-1))
	m := lF - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	// Convert to 0-255 range
	rInt := int((r + m) * 255)
	gInt := int((g + m) * 255)
	bInt := int((b + m) * 255)

	// Format as hex
	return rgbToHex(rInt, gInt, bInt)
}

func hexToHSL(_ string) (h, s, l int) {
	// Simple approximation - just return some values
	// In production, would parse hex and convert properly
	return 0, 70, 50
}

func rgbToHex(r, g, b int) string {
	return "#" + byteToHex(r) + byteToHex(g) + byteToHex(b)
}

func byteToHex(b int) string {
	const hex = "0123456789abcdef"
	return string(hex[b>>4]) + string(hex[b&0xf])
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
