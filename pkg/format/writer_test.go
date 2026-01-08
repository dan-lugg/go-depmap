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
			name:         "json format",
			format:       "json",
			expectedType: "*format.JSONWriter",
		},
		{
			name:         "d3js format",
			format:       "d3js",
			expectedType: "*format.D3JSWriter",
		},
		{
			name:         "unknown format defaults to json",
			format:       "unknown",
			expectedType: "*format.JSONWriter",
		},
		{
			name:         "empty format defaults to json",
			format:       "",
			expectedType: "*format.JSONWriter",
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
			case *JSONWriter:
				writerType = "*format.JSONWriter"
			case *D3JSWriter:
				writerType = "*format.D3JSWriter"
			}

			if writerType != tt.expectedType {
				t.Errorf("expected type %s, got %s", tt.expectedType, writerType)
			}
		})
	}
}

func Test_GetFormatWriter_ImplementsInterface(t *testing.T) {
	formats := []string{"json", "d3js"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			writer := GetFormatWriter(format)

			var _ Writer = writer

			g := graph.NewDependencyGraph()
			var buf bytes.Buffer
			config := Config{"pretty": true}
			err := writer.Write(&buf, g, config)
			if err != nil {
				t.Errorf("Write() error = %v", err)
			}
		})
	}
}
