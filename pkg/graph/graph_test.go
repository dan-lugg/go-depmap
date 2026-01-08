package graph

import (
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func Test_CreateNode(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("example.com/test", "test")
	obj := types.NewFunc(pos, pkg, "TestFunc", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	testPkg := &packages.Package{
		PkgPath: "example.com/test",
		Fset:    fset,
	}

	tests := []struct {
		name         string
		pkg          *packages.Package
		obj          types.Object
		nodeName     string
		kind         NodeKind
		signature    string
		expectedID   string
		expectedName string
		expectedKind NodeKind
		expectedPkg  string
		expectedFile string
	}{
		{
			name:         "function node",
			pkg:          testPkg,
			obj:          obj,
			nodeName:     "TestFunc",
			kind:         KindFunction,
			signature:    "func TestFunc()",
			expectedID:   "example.com/test::TestFunc",
			expectedName: "TestFunc",
			expectedKind: KindFunction,
			expectedPkg:  "example.com/test",
			expectedFile: "test.go",
		},
		{
			name:         "method node",
			pkg:          testPkg,
			obj:          obj,
			nodeName:     "(*MyType).Method",
			kind:         KindMethod,
			signature:    "func (m *MyType) Method()",
			expectedID:   "example.com/test::(*MyType).Method",
			expectedName: "(*MyType).Method",
			expectedKind: KindMethod,
			expectedPkg:  "example.com/test",
			expectedFile: "test.go",
		},
		{
			name:         "type node",
			pkg:          testPkg,
			obj:          obj,
			nodeName:     "MyType",
			kind:         KindType,
			signature:    "type MyType struct{}",
			expectedID:   "example.com/test::MyType",
			expectedName: "MyType",
			expectedKind: KindType,
			expectedPkg:  "example.com/test",
			expectedFile: "test.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := CreateNode(tt.pkg, tt.obj, tt.nodeName, tt.kind, tt.signature)

			if node == nil {
				t.Fatal("CreateNode returned nil")
			}

			if node.ID != tt.expectedID {
				t.Errorf("ID = %s, want %s", node.ID, tt.expectedID)
			}

			if node.Name != tt.expectedName {
				t.Errorf("Name = %s, want %s", node.Name, tt.expectedName)
			}

			if node.Kind != tt.expectedKind {
				t.Errorf("Kind = %s, want %s", node.Kind, tt.expectedKind)
			}

			if node.Package != tt.expectedPkg {
				t.Errorf("Package = %s, want %s", node.Package, tt.expectedPkg)
			}

			if node.File != tt.expectedFile {
				t.Errorf("File = %s, want %s", node.File, tt.expectedFile)
			}

			if node.Signature != tt.signature {
				t.Errorf("Signature = %s, want %s", node.Signature, tt.signature)
			}

			if node.Line == 0 {
				t.Error("Line should not be 0")
			}
		})
	}
}

func Test_CreateNode_IDFormat(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("example.com/myapp/utils", "utils")
	obj := types.NewFunc(pos, pkg, "Helper", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	testPkg := &packages.Package{
		PkgPath: "example.com/myapp/utils",
		Fset:    fset,
	}

	tests := []struct {
		name       string
		nodeName   string
		expectedID string
	}{
		{
			name:       "simple function",
			nodeName:   "Helper",
			expectedID: "example.com/myapp/utils::Helper",
		},
		{
			name:       "pointer receiver method",
			nodeName:   "(*Config).Load",
			expectedID: "example.com/myapp/utils::(*Config).Load",
		},
		{
			name:       "value receiver method",
			nodeName:   "Config.Save",
			expectedID: "example.com/myapp/utils::Config.Save",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := CreateNode(testPkg, obj, tt.nodeName, KindFunction, "func signature")

			if node.ID != tt.expectedID {
				t.Errorf("ID = %s, want %s", node.ID, tt.expectedID)
			}
		})
	}
}

func Test_CreateNode_FilePathHandling(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("/full/path/to/project/pkg/test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("test", "test")
	obj := types.NewFunc(pos, pkg, "TestFunc", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	testPkg := &packages.Package{
		PkgPath: "test",
		Fset:    fset,
	}

	node := CreateNode(testPkg, obj, "TestFunc", KindFunction, "func TestFunc()")

	if node.File != "test.go" {
		t.Errorf("File = %s, want test.go (basename only)", node.File)
	}
}

func Test_CreateNode_DifferentKinds(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("test", "test")
	obj := types.NewFunc(pos, pkg, "TestFunc", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	testPkg := &packages.Package{
		PkgPath: "test",
		Fset:    fset,
	}

	kinds := []NodeKind{KindFunction, KindMethod, KindType}

	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			node := CreateNode(testPkg, obj, "TestItem", kind, "signature")

			if node.Kind != kind {
				t.Errorf("Kind = %s, want %s", node.Kind, kind)
			}
		})
	}
}

func Test_CreateNode_SignaturePreserved(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("test.go", -1, 100)
	pos := file.Pos(10)

	pkg := types.NewPackage("test", "test")
	obj := types.NewFunc(pos, pkg, "TestFunc", types.NewSignatureType(nil, nil, nil, nil, nil, false))

	testPkg := &packages.Package{
		PkgPath: "test",
		Fset:    fset,
	}

	signatures := []string{
		"func Simple()",
		"func WithParams(a int, b string) error",
		"func Complex[T any](x T) (T, error)",
		"func (r *Receiver) Method() error",
	}

	for _, sig := range signatures {
		t.Run(sig, func(t *testing.T) {
			node := CreateNode(testPkg, obj, "TestFunc", KindFunction, sig)

			if node.Signature != sig {
				t.Errorf("Signature = %s, want %s", node.Signature, sig)
			}
		})
	}
}
