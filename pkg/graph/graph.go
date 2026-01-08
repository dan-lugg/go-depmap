package graph

import (
	"fmt"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// CreateNode creates a Node from a types.Object
func CreateNode(pkg *packages.Package, obj types.Object, name string, kind NodeKind, signature string) *Node {
	fset := pkg.Fset
	pos := fset.Position(obj.Pos())

	// Create a unique ID. PkgPath + Name is usually sufficient,
	// but for methods we want PkgPath + Receiver + Name.
	// We use the full String() representation of the object as the base for uniqueness,
	// hashed or cleaned if necessary. Here we use a composite key.
	id := fmt.Sprintf("%s::%s", pkg.PkgPath, name)

	return &Node{
		ID:        id,
		Name:      name,
		Kind:      kind,
		Package:   pkg.PkgPath,
		File:      filepath.Base(pos.Filename),
		Line:      pos.Line,
		Signature: signature,
	}
}
