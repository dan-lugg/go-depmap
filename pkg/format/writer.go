// Package format provides different output formatters for dependency graphs,
// including JSON and D3.js-compatible formats.
package format

import (
	"io"

	"go-depmap/pkg/graph"
)

// Writer is the interface for different output formatters
type Writer interface {
	// Write formats and writes the dependency graph to the given writer
	Write(w io.Writer, graph *graph.DependencyGraph) error
}

// GetFormatWriter returns a Writer for the given format name
func GetFormatWriter(format string) Writer {
	switch format {
	case "pretty-json":
		return &PrettyJSONWriter{}
	case "minify-json":
		return &MinifyJSONWriter{}
	case "d3js-json":
		return &D3JSJSONWriter{}
	default:
		// Default to pretty JSON
		return &PrettyJSONWriter{}
	}
}
