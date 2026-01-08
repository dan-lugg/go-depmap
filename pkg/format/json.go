package format

import (
	"encoding/json"
	"io"

	"go-depmap/pkg/graph"
)

// JSONWriter writes the graph as JSON (pretty-printed or minified based on config)
type JSONWriter struct{}

func (w *JSONWriter) Write(writer io.Writer, graph *graph.DependencyGraph, config Config) error {
	enc := json.NewEncoder(writer)

	// Check if pretty printing is enabled (defaults to true)
	if config.GetBool("pretty", true) {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(graph)
}
