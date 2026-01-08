package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go-depmap/pkg/graph"
)

func Test_PrettyJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name    string
		graph   *graph.DependencyGraph
		wantErr bool
	}{
		{
			name:    "empty graph",
			graph:   graph.NewDependencyGraph(),
			wantErr: false,
		},
		{
			name: "graph with nodes",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
				},
				Edges: make(map[string][]string),
			},
			wantErr: false,
		},
		{
			name: "graph with nodes and edges",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
					"test::Type1": {
						ID:        "test::Type1",
						Name:      "Type1",
						Kind:      graph.KindType,
						Package:   "test",
						File:      "test.go",
						Line:      5,
						Signature: "type Type1 struct{}",
					},
				},
				Edges: map[string][]string{
					"test::func1": {"test::Type1"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &PrettyJSONWriter{}
			var buf bytes.Buffer

			err := w.Write(&buf, tt.graph)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
				}

				output := buf.String()
				if !strings.Contains(output, "  ") {
					t.Error("Output does not appear to be pretty-printed (no indentation found)")
				}

				if _, ok := result["nodes"]; !ok {
					t.Error("Output missing 'nodes' field")
				}
				if _, ok := result["edges"]; !ok {
					t.Error("Output missing 'edges' field")
				}
			}
		})
	}
}

func Test_MinifyJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name    string
		graph   *graph.DependencyGraph
		wantErr bool
	}{
		{
			name:    "empty graph",
			graph:   graph.NewDependencyGraph(),
			wantErr: false,
		},
		{
			name: "graph with nodes",
			graph: &graph.DependencyGraph{
				Nodes: map[string]*graph.Node{
					"test::func1": {
						ID:        "test::func1",
						Name:      "func1",
						Kind:      graph.KindFunction,
						Package:   "test",
						File:      "test.go",
						Line:      10,
						Signature: "func func1()",
					},
				},
				Edges: make(map[string][]string),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &MinifyJSONWriter{}
			var buf bytes.Buffer

			err := w.Write(&buf, tt.graph)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var result map[string]interface{}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Errorf("Output is not valid JSON: %v", err)
				}
			}
		})
	}
}

func Test_JSONWriters_Comparison(t *testing.T) {
	testGraph := &graph.DependencyGraph{
		Nodes: map[string]*graph.Node{
			"test::func1": {
				ID:        "test::func1",
				Name:      "func1",
				Kind:      graph.KindFunction,
				Package:   "test",
				File:      "test.go",
				Line:      10,
				Signature: "func func1()",
			},
		},
		Edges: map[string][]string{
			"test::func1": {"test::Type1"},
		},
	}

	prettyWriter := &PrettyJSONWriter{}
	minifyWriter := &MinifyJSONWriter{}

	var prettyBuf, minifyBuf bytes.Buffer

	if err := prettyWriter.Write(&prettyBuf, testGraph); err != nil {
		t.Fatalf("PrettyJSONWriter.Write() error = %v", err)
	}

	if err := minifyWriter.Write(&minifyBuf, testGraph); err != nil {
		t.Fatalf("MinifyJSONWriter.Write() error = %v", err)
	}

	var prettyResult, minifyResult map[string]interface{}

	if err := json.Unmarshal(prettyBuf.Bytes(), &prettyResult); err != nil {
		t.Fatalf("Failed to parse pretty JSON: %v", err)
	}

	if err := json.Unmarshal(minifyBuf.Bytes(), &minifyResult); err != nil {
		t.Fatalf("Failed to parse minified JSON: %v", err)
	}

	if len(prettyResult) != len(minifyResult) {
		t.Errorf("Different number of top-level keys: pretty=%d, minify=%d",
			len(prettyResult), len(minifyResult))
	}
}
