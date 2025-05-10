package typesys

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestFileCreation(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	testCases := []struct {
		name     string
		path     string
		isTest   bool
		fileName string
	}{
		{
			name:     "regular file",
			path:     "/test/module/file.go",
			isTest:   false,
			fileName: "file.go",
		},
		{
			name:     "test file",
			path:     "/test/module/file_test.go",
			isTest:   true,
			fileName: "file_test.go",
		},
		{
			name:     "nested path",
			path:     "/test/module/pkg/subpkg/file.go",
			isTest:   false,
			fileName: "file.go",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := NewFile(tc.path, pkg)

			if file.Path != tc.path {
				t.Errorf("File path = %q, want %q", file.Path, tc.path)
			}

			if file.Name != tc.fileName {
				t.Errorf("File name = %q, want %q", file.Name, tc.fileName)
			}

			if file.IsTest != tc.isTest {
				t.Errorf("File.IsTest = %v, want %v", file.IsTest, tc.isTest)
			}

			if file.Package != pkg {
				t.Errorf("File.Package not set correctly")
			}

			if len(file.Symbols) != 0 {
				t.Errorf("New file should have no symbols, got %d", len(file.Symbols))
			}

			if len(file.Imports) != 0 {
				t.Errorf("New file should have no imports, got %d", len(file.Imports))
			}
		})
	}
}

func TestAddSymbol(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)

	// Create and add a symbol
	sym := NewSymbol("TestSymbol", KindFunction)
	file.AddSymbol(sym)

	// Verify the symbol was added to the file
	if len(file.Symbols) != 1 {
		t.Errorf("File should have 1 symbol, got %d", len(file.Symbols))
	}

	if file.Symbols[0] != sym {
		t.Errorf("File.Symbols[0] is not the added symbol")
	}

	// Verify the symbol's file reference was set
	if sym.File != file {
		t.Errorf("Symbol.File not set to the file")
	}

	// Verify the symbol was added to the package
	if len(pkg.Symbols) != 1 {
		t.Errorf("Package should have 1 symbol, got %d", len(pkg.Symbols))
	}
}

func TestAddImport(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)

	// Create and add an import
	imp := &Import{
		Path: "fmt",
		Name: "",
		Pos:  token.Pos(10),
		End:  token.Pos(20),
	}

	file.AddImport(imp)

	// Verify the import was added to the file
	if len(file.Imports) != 1 {
		t.Errorf("File should have 1 import, got %d", len(file.Imports))
	}

	if file.Imports[0] != imp {
		t.Errorf("File.Imports[0] is not the added import")
	}

	// Verify the import's file reference was set
	if imp.File != file {
		t.Errorf("Import.File not set to the file")
	}

	// Verify the import was added to the package
	if len(pkg.Imports) != 1 {
		t.Errorf("Package should have 1 import, got %d", len(pkg.Imports))
	}

	if pkg.Imports["fmt"] != imp {
		t.Errorf("Package.Imports[\"fmt\"] is not the added import")
	}
}

func TestGetPositionInfo(t *testing.T) {
	// Create a FileSet and parse some Go code to get real token.Pos values
	fset := token.NewFileSet()
	src := `package test

func main() {
	println("Hello, world!")
}
`

	// Parse the source code
	f, err := parser.ParseFile(fset, "test.go", src, parser.AllErrors)
	if err != nil {
		t.Fatalf("Failed to parse test code: %v", err)
	}

	// Create our test file
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/test.go", pkg)
	file.FileSet = fset
	file.AST = f

	// Find a function declaration to use for positions
	var funcDecl *ast.FuncDecl
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			funcDecl = fd
			return false
		}
		return true
	})

	if funcDecl == nil {
		t.Fatalf("Failed to find function declaration in test code")
	}

	// Test GetPositionInfo with valid positions
	posInfo := file.GetPositionInfo(funcDecl.Pos(), funcDecl.End())

	if posInfo == nil {
		t.Fatalf("GetPositionInfo returned nil for valid positions")
	}

	// The line numbers should be valid (function spans lines 3-5)
	if posInfo.LineStart < 3 || posInfo.LineEnd > 5 {
		t.Errorf("Position line numbers out of expected range, got %d-%d",
			posInfo.LineStart, posInfo.LineEnd)
	}

	// The length should be positive
	if posInfo.Length <= 0 {
		t.Errorf("Position length should be positive, got %d", posInfo.Length)
	}

	// Test with invalid positions
	posInfo = file.GetPositionInfo(token.NoPos, token.NoPos)
	if posInfo != nil {
		t.Errorf("GetPositionInfo should return nil for invalid positions")
	}

	// Test with reversed positions (end before start)
	posInfo = file.GetPositionInfo(funcDecl.End(), funcDecl.Pos())
	if posInfo == nil {
		t.Errorf("GetPositionInfo should handle reversed positions")
	}

	// Verify it swapped them correctly - length should still be positive
	if posInfo == nil || posInfo.Length <= 0 {
		t.Errorf("Position length should be positive for swapped positions, got %d", posInfo.Length)
	}
}
