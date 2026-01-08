// Package main provides the command-line interface for the go-depmap tool,
// which analyzes Go code dependencies and generates dependency graphs.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"

	"go-depmap/pkg/analyzer"
	"go-depmap/pkg/format"

	"golang.org/x/tools/go/packages"
)

func main() {
	// CLI Flags
	sourcePtr := flag.String("source", ".", "The directory of the Go project to analyze")
	formatPtr := flag.String("format", "json", "Output format: json, d3js")
	configPtr := flag.String("config", "{}", "JSON configuration object for the formatter (e.g., {\"pretty\":true,\"groupPackages\":true})")
	flag.Parse()

	log.Printf("Analyzing project in: %s", *sourcePtr)

	// Parse config JSON
	var configMap map[string]any
	if err := json.Unmarshal([]byte(*configPtr), &configMap); err != nil {
		log.Fatalf("Failed to parse config JSON: %v", err)
	}
	config := format.Config(configMap)

	// Load the packages using go/packages
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedModule,
		Dir:   *sourcePtr,
		Tests: false, // Set to true if you want to include test files
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Fatalf("Packages contained errors")
	}

	// Analyze the packages
	a := analyzer.New(pkgs)
	graph := a.Analyze()

	// Get the appropriate format writer
	writer := format.GetFormatWriter(*formatPtr)
	writerType := reflect.TypeOf(writer).Elem().Name()
	log.Printf("Using writer: %s", writerType)

	// Write to STDOUT
	if err := writer.Write(os.Stdout, graph, config); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	log.Printf("Analysis complete.")
	log.Printf("  Nodes: %d", len(graph.Nodes))
	log.Printf("  Edges: %d", graph.CountEdges())
}
