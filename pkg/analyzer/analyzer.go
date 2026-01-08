// Package analyzer provides functionality for analyzing Go code dependencies
// and building dependency graphs from parsed Go packages.
package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"go-depmap/pkg/graph"

	"golang.org/x/tools/go/packages"
)

// Analyzer performs dependency analysis on Go packages
type Analyzer struct {
	packages       []*packages.Package
	projectObjects map[types.Object]*graph.Node
	graph          *graph.DependencyGraph
}

// New creates a new Analyzer for the given packages
func New(pkgs []*packages.Package) *Analyzer {
	return &Analyzer{
		packages:       pkgs,
		projectObjects: make(map[types.Object]*graph.Node),
		graph:          graph.NewDependencyGraph(),
	}
}

// Analyze performs the full dependency analysis
func (a *Analyzer) Analyze() *graph.DependencyGraph {
	a.collectDefinitions()
	a.analyzeDependencies()
	return a.graph
}

// collectDefinitions scans all packages and collects function and type definitions
func (a *Analyzer) collectDefinitions() {
	log.Println("Scanning definitions...")

	for _, pkg := range a.packages {
		// Skip if it's not part of the main module being analyzed
		if pkg.Module == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {

				// Case A: Function Declarations
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.Defs[x.Name]
					if obj == nil {
						return true
					}

					kind := graph.KindFunction
					name := x.Name.Name
					sig := obj.Type().String()

					// Check if it is a method
					if x.Recv != nil {
						kind = graph.KindMethod
						// Format: (Receiver).Method
						recvType := x.Recv.List[0].Type
						// We try to get the raw type name for the ID
						if star, ok := recvType.(*ast.StarExpr); ok {
							if ident, ok := star.X.(*ast.Ident); ok {
								name = fmt.Sprintf("(*%s).%s", ident.Name, name)
							}
						} else if ident, ok := recvType.(*ast.Ident); ok {
							name = fmt.Sprintf("%s.%s", ident.Name, name)
						}
					}

					node := graph.CreateNode(pkg, obj, name, kind, sig)
					a.projectObjects[obj] = node
					a.graph.Nodes[node.ID] = node

				// Case B: Type Declarations (GenDecl with TypeSpec)
				case *ast.GenDecl:
					if x.Tok == token.TYPE {
						for _, spec := range x.Specs {
							typeSpec, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							obj := pkg.TypesInfo.Defs[typeSpec.Name]
							if obj == nil {
								continue
							}

							node := graph.CreateNode(pkg, obj, typeSpec.Name.Name, graph.KindType, obj.Type().String())
							a.projectObjects[obj] = node
							a.graph.Nodes[node.ID] = node
						}
					}
				}
				return true
			})
		}
	}

	log.Printf("Found %d definitions inside the project.", len(a.projectObjects))
}

// analyzeDependencies analyzes function bodies to find dependencies
func (a *Analyzer) analyzeDependencies() {
	log.Println("Analyzing function dependencies...")

	for _, pkg := range a.packages {
		if pkg.Module == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				// Get the Node for this function
				fnObj := pkg.TypesInfo.Defs[fn.Name]
				sourceNode, exists := a.projectObjects[fnObj]
				if !exists {
					return true
				}

				// Track unique dependencies to avoid duplicates
				seenDeps := make(map[string]bool)

				// Helper to record a dependency
				addDep := func(targetObj types.Object) {
					// Ignore if target is not in our project definitions
					// This automatically filters out stdlib, vendor, etc.
					if targetNode, isLocal := a.projectObjects[targetObj]; isLocal {
						// Don't depend on self
						if targetNode.ID == sourceNode.ID {
							return
						}
						if !seenDeps[targetNode.ID] {
							a.graph.Edges[sourceNode.ID] = append(a.graph.Edges[sourceNode.ID], targetNode.ID)
							seenDeps[targetNode.ID] = true
						}
					}
				}

				// Walk the function body and signature
				ast.Inspect(fn, func(subNode ast.Node) bool {
					ident, ok := subNode.(*ast.Ident)
					if !ok {
						return true
					}

					// Resolve the identifier using TypeInfo
					// Uses maps identifiers to the objects they denote
					if usedObj, ok := pkg.TypesInfo.Uses[ident]; ok {
						addDep(usedObj)
					}
					return true
				})

				return true
			})
		}
	}
}
