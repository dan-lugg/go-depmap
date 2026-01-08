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
- `-format <format>`: Specify the output format (default: "json")
    - `json`: JSON output with configurable formatting
    - `d3js`: D3.js force-directed graph format
- `-config <json>`: JSON configuration object for the formatter (default: "{}")
    - Available config options:
        - `pretty` (bool): Enable pretty-printed output (default: true)
        - `groupPackages` (bool): Group nodes by package in D3.js format (default: true)

### Examples

Analyze a specific project:

```bash
./go-depmap -source=/path/to/your/go/project
```

Generate minified JSON output:

```bash
./go-depmap -format=json -config='{"pretty":false}'
```

Generate D3.js output without package grouping:

```bash
./go-depmap -format=d3js -config='{"groupPackages":false}'
```

Generate minified D3.js output:

```bash
./go-depmap -format=d3js -config='{"pretty":false,"groupPackages":true}'
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

### D3.js Format (d3js)

Compatible with D3.js force-directed graph visualizations with **convex hull package grouping**:

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
      "group": 1,
      "package_id": "example.com/myapp/utils"
    }
  ],
  "links": [
    {
      "source": "example.com/myapp/main::main",
      "target": "example.com/myapp/utils::Helper",
      "value": 1
    }
  ],
  "packages": [
    {
      "id": "example.com/myapp/utils",
      "label": "example.com/myapp/utils",
      "nodes": [
        "example.com/myapp/utils::Helper",
        "example.com/myapp/utils::Config"
      ]
    }
  ]
}
```

**Features:**
- **Node groups**: function=1, method=2, type=3 (useful for coloring in visualizations)
- **Package grouping**: Nodes are grouped by their fully qualified package name
- **Convex hull support**: The `packages` array enables visual grouping with convex hulls
- **Interactive visualization**: Includes `index.html` for interactive D3.js visualization with package boundaries

See [PACKAGE_GROUPING.md](PACKAGE_GROUPING.md) for detailed information about the package grouping feature.

## Visualization

The tool includes an interactive D3.js visualization (`index.html`) that displays your dependency graph with package grouping:

### Quick Start

1. Generate the dependency graph:
   ```bash
   ./go-depmap -source=./pkg -format=d3js > graph.json
   ```

2. Start a local web server:
   ```bash
   python3 -m http.server 8000
   ```

3. Open your browser to `http://localhost:8000/index.html`

### Features

- **Package Grouping**: Visual boundaries (convex hulls) around nodes from the same package
- **Interactive Controls**: Adjust force simulation, node size, and toggle labels
- **Color Coding**: Functions (orange), Methods (blue), Types (green)
- **Drag & Zoom**: Rearrange nodes and explore large graphs
- **Tooltips**: Hover over nodes for detailed information

See [PACKAGE_GROUPING.md](PACKAGE_GROUPING.md) for complete visualization documentation.

## Use Cases

- **Dependency Visualization**: Generate interactive graphs using D3.js or other visualization libraries
- **Code Analysis**: Understand how functions and types interact within your project
- **Refactoring Support**: Identify impact of changes to functions or types
- **Documentation**: Auto-generate dependency documentation for your project
- **Architecture Review**: Visualize package structure and cross-package dependencies

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

