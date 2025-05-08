package typesys

import (
	"testing"
)

func TestPackageCreation(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	if pkg.Name != "testpkg" {
		t.Errorf("Package.Name = %q, want %q", pkg.Name, "testpkg")
	}

	if pkg.ImportPath != "github.com/example/testpkg" {
		t.Errorf("Package.ImportPath = %q, want %q", pkg.ImportPath, "github.com/example/testpkg")
	}

	if pkg.Module != module {
		t.Errorf("Package.Module not set correctly")
	}

	if len(pkg.Symbols) != 0 {
		t.Errorf("New package should have no symbols, got %d", len(pkg.Symbols))
	}

	if len(pkg.Files) != 0 {
		t.Errorf("New package should have no files, got %d", len(pkg.Files))
	}

	if len(pkg.Imports) != 0 {
		t.Errorf("New package should have no imports, got %d", len(pkg.Imports))
	}
}

func TestAddFile(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	// Create and add a file
	file := NewFile("/test/module/file.go", nil) // nil package to avoid circular reference
	pkg.AddFile(file)

	// Verify file was added
	if len(pkg.Files) != 1 {
		t.Errorf("Package should have 1 file, got %d", len(pkg.Files))
	}

	if pkg.Files["/test/module/file.go"] != file {
		t.Errorf("File not correctly added to package")
	}

	// Verify package reference in file
	if file.Package != pkg {
		t.Errorf("File.Package not set to the package")
	}
}

func TestPackageAddSymbol(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	// Create and add a symbol
	sym := NewSymbol("TestSymbol", KindFunction)
	pkg.AddSymbol(sym)

	// Verify symbol was added to the package
	if len(pkg.Symbols) != 1 {
		t.Errorf("Package should have 1 symbol, got %d", len(pkg.Symbols))
	}

	// Use the actual symbol ID instead of hardcoding
	if pkg.Symbols[sym.ID] != sym {
		t.Errorf("Symbol not correctly added to package")
	}

	// Verify package reference in symbol
	if sym.Package != pkg {
		t.Errorf("Symbol.Package not set to the package")
	}
}

func TestSymbolByName(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	// Create and add several symbols
	funcSym := NewSymbol("TestFunction", KindFunction)
	varSym := NewSymbol("TestVariable", KindVariable)
	structSym := NewSymbol("TestStruct", KindStruct)
	constSym := NewSymbol("TestConstant", KindConstant)

	pkg.AddSymbol(funcSym)
	pkg.AddSymbol(varSym)
	pkg.AddSymbol(structSym)
	pkg.AddSymbol(constSym)

	// Test SymbolByName with single kind
	funcs := pkg.SymbolByName("TestFunction", KindFunction)
	if len(funcs) != 1 || funcs[0] != funcSym {
		t.Errorf("SymbolByName for function returned wrong symbols")
	}

	// Test SymbolByName with multiple kinds
	varsAndConsts := pkg.SymbolByName("Test", KindVariable, KindConstant)
	if len(varsAndConsts) != 2 {
		t.Errorf("SymbolByName for variables and constants returned %d symbols, want 2",
			len(varsAndConsts))
	}

	// Test SymbolByName with no match
	noMatches := pkg.SymbolByName("NonExistent", KindFunction)
	if len(noMatches) != 0 {
		t.Errorf("SymbolByName for non-existent name returned %d symbols, want 0",
			len(noMatches))
	}

	// Test SymbolByName with wrong kind
	wrongKind := pkg.SymbolByName("TestFunction", KindVariable)
	if len(wrongKind) != 0 {
		t.Errorf("SymbolByName with wrong kind returned %d symbols, want 0",
			len(wrongKind))
	}
}
