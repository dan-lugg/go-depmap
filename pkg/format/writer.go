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
	Write(w io.Writer, graph *graph.DependencyGraph, config Config) error
}

// GetFormatWriter returns a Writer for the given format name
func GetFormatWriter(format string) Writer {
	switch format {
	case "json":
		return &JSONWriter{}
	case "d3js":
		return &D3JSWriter{}
	case "cosmo":
		return &CosmoWriter{}
	case "antvg6":
		return &AntVG6Writer{}
	default:
		// Default to JSON
		return &JSONWriter{}
	}
}
