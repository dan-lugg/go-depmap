# Go Dependency Mapper

[![CI](https://github.com/dan-lugg/go-depmap/workflows/CI/badge.svg)](https://github.com/dan-lugg/go-depmap/actions?query=workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/dan-lugg/go-depmap)](https://goreportcard.com/report/github.com/dan-lugg/go-depmap)
[![codecov](https://codecov.io/gh/dan-lugg/go-depmap/branch/main/graph/badge.svg)](https://codecov.io/gh/dan-lugg/go-depmap)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dan-lugg/go-depmap)](https://github.com/dan-lugg/go-depmap)
[![Release](https://img.shields.io/github/v/release/dan-lugg/go-depmap)](https://github.com/dan-lugg/go-depmap/releases)

A powerful Golang tool that analyzes Go projects to create comprehensive dependency mappings between functions, methods,
and types.

## Features

- **Accurate Symbol Resolution**: Uses Go's AST (Abstract Syntax Tree) and Type Checker for precise dependency tracking
- **Project-Scoped Analysis**: Automatically filters out standard library and vendor dependencies
- **Comprehensive Coverage**: Tracks functions, methods, and type declarations
- **Flexible Output**: Generates JSON output that can be easily transformed for visualization

## How It Works

The tool performs a two-pass analysis:

1. **Pass 1 (Discovery)**: Scans the project to collect all function, method, and type definitions
2. **Pass 2 (Analysis)**: Analyzes function bodies to identify dependencies on other project symbols

The tool correctly resolves all identifiers using Go's type system, ensuring that references are accurate (e.g.,
distinguishing between different variables with the same name).

## Installation

```bash
# Clone or navigate to the project directory
cd go-depmap

# Build the tool
go build -o go-depmap ./cmd/depmap
```

## Usage

### Basic Usage

Analyze the current directory and output pretty JSON to STDOUT:

```bash
./go-depmap
```

### Options

- `-source <path>`: Specify the directory of the Go project to analyze (default: ".")
- `-format <format>`: Specify the output format (default: "pretty-json")
    - `pretty-json`: Pretty-printed JSON output
    - `minify-json`: Compact JSON output
    - `d3js-json`: D3.js force-directed graph format

### Examples

Analyze a specific project:

```bash
./go-depmap -source=/path/to/your/go/project
```

Generate minified JSON output:

```bash
./go-depmap -format=minify-json
```

Generate D3.js-compatible output and save to file:

```bash
./go-depmap -format=d3js-json > graph.json
```

Pipe output to other tools:

```bash
./go-depmap -format=minify-json | jq '.nodes | length'
```

## Output Formats

The tool outputs to STDOUT (log messages go to STDERR), making it easy to pipe to other tools or redirect to files.

### Standard Format (pretty-json / minify-json)

The default format with two main sections:

**Nodes**: Contains metadata about each function, method, or type definition:

```json
{
  "nodes": {
    "example.com/myapp/utils::Helper": {
      "id": "example.com/myapp/utils::Helper",
      "name": "Helper",
      "kind": "function",
      "package": "example.com/myapp/utils",
      "file": "utils.go",
      "line": 10,
      "signature": "func() string"
    }
  },
  "edges": {
    "example.com/myapp/main::main": [
      "example.com/myapp/utils::Helper",
      "example.com/myapp/types::Config"
    ]
  }
}
```

### D3.js Format (d3js-json)

Compatible with D3.js force-directed graph visualizations:

```json
{
  "nodes": [
    {
      "id": "example.com/myapp/utils::Helper",
      "name": "Helper",
      "kind": "function",
      "package": "example.com/myapp/utils",
      "file": "utils.go",
      "line": 10,
      "signature": "func() string",
      "group": 1
    }
  ],
  "links": [
    {
      "source": "example.com/myapp/main::main",
      "target": "example.com/myapp/utils::Helper",
      "value": 1
    }
  ]
}
```

Node groups: function=1, method=2, type=3 (useful for coloring in visualizations)

## Use Cases

- **Dependency Visualization**: Generate interactive graphs using D3.js or other visualization libraries
- **Code Analysis**: Understand how functions and types interact within your project
- **Refactoring Support**: Identify impact of changes to functions or types
- **Documentation**: Auto-generate dependency documentation for your project

## Technical Details

- Uses `golang.org/x/tools/go/packages` for robust Go code loading and analysis
- Handles Go modules, build tags, and complex project structures
- Filters dependencies based on module boundaries (excludes stdlib and vendor code)
- Provides accurate symbol resolution through Go's type checker

## Example: Analyzing This Tool

Running the tool on itself:

```bash
./go-depmap -source=.
```

Output (to STDERR):

```
2026/01/07 21:16:46 Analyzing project in: .
2026/01/07 21:16:47 Scanning definitions...
2026/01/07 21:16:47 Found 7 definitions inside the project.
2026/01/07 21:16:47 Analyzing function dependencies...
2026/01/07 21:16:47 Analysis complete.
2026/01/07 21:16:47   Nodes: 7
2026/01/07 21:16:47   Edges: 7
```

The JSON output goes to STDOUT and can be redirected to a file or piped to another tool.

## License

This tool is provided as-is for dependency analysis purposes.

