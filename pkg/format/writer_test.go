package format

import (
	"bytes"
	"testing"

	"go-depmap/pkg/graph"
)

func Test_GetFormatWriter(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		expectedType string
	}{
		{
			name:         "pretty-json format",
			format:       "pretty-json",
			expectedType: "*format.PrettyJSONWriter",
		},
		{
			name:         "minify-json format",
			format:       "minify-json",
			expectedType: "*format.MinifyJSONWriter",
		},
		{
			name:         "d3js-json format",
			format:       "d3js-json",
			expectedType: "*format.D3JSJSONWriter",
		},
		{
			name:         "unknown format defaults to pretty-json",
			format:       "unknown",
			expectedType: "*format.PrettyJSONWriter",
		},
		{
			name:         "empty format defaults to pretty-json",
			format:       "",
			expectedType: "*format.PrettyJSONWriter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := GetFormatWriter(tt.format)
			if writer == nil {
				t.Fatal("GetFormatWriter returned nil")
			}

			writerType := ""
			switch writer.(type) {
			case *PrettyJSONWriter:
				writerType = "*format.PrettyJSONWriter"
			case *MinifyJSONWriter:
				writerType = "*format.MinifyJSONWriter"
			case *D3JSJSONWriter:
				writerType = "*format.D3JSJSONWriter"
			}

			if writerType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, writerType)
			}
		})
	}
}

func Test_GetFormatWriter_ImplementsInterface(t *testing.T) {
	formats := []string{"pretty-json", "minify-json", "d3js-json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			writer := GetFormatWriter(format)

			var _ Writer = writer

			g := graph.NewDependencyGraph()
			var buf bytes.Buffer
			err := writer.Write(&buf, g)
			if err != nil {
				t.Errorf("Write() error = %v", err)
			}
		})
	}
}
