package format

import (
	"encoding/json"
	"io"

	"go-depmap/pkg/graph"
)

// PrettyJSONWriter writes the graph as pretty-printed JSON
type PrettyJSONWriter struct{}

func (w *PrettyJSONWriter) Write(writer io.Writer, graph *graph.DependencyGraph) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	return enc.Encode(graph)
}

// MinifyJSONWriter writes the graph as minified JSON
type MinifyJSONWriter struct{}

func (w *MinifyJSONWriter) Write(writer io.Writer, graph *graph.DependencyGraph) error {
	enc := json.NewEncoder(writer)
	return enc.Encode(graph)
}
