package typesys

import (
	"go/token"
	"testing"
)

func TestNewReference(t *testing.T) {
	// Setup test data
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	symbol := NewSymbol("TestSymbol", KindFunction)

	// Create a new reference
	ref := NewReference(symbol, file, token.Pos(10), token.Pos(20))

	// Verify reference properties
	if ref.Symbol != symbol {
		t.Errorf("Reference.Symbol = %v, want %v", ref.Symbol, symbol)
	}

	if ref.File != file {
		t.Errorf("Reference.File = %v, want %v", ref.File, file)
	}

	if ref.Pos != token.Pos(10) {
		t.Errorf("Reference.Pos = %v, want %v", ref.Pos, token.Pos(10))
	}

	if ref.End != token.Pos(20) {
		t.Errorf("Reference.End = %v, want %v", ref.End, token.Pos(20))
	}

	// Verify the reference was added to the symbol
	if len(symbol.References) != 1 || symbol.References[0] != ref {
		t.Errorf("Reference not added to symbol.References")
	}
}

func TestSetReferenceContext(t *testing.T) {
	// Setup test data
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	symbol := NewSymbol("TestSymbol", KindFunction)
	contextSym := NewSymbol("ContextFunction", KindFunction)

	// Create a new reference
	ref := NewReference(symbol, file, token.Pos(10), token.Pos(20))

	// Set the reference context
	ref.SetContext(contextSym)

	// Verify context was set
	if ref.Context != contextSym {
		t.Errorf("Reference.Context = %v, want %v", ref.Context, contextSym)
	}
}

func TestSetIsWrite(t *testing.T) {
	// Setup test data
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	symbol := NewSymbol("TestVariable", KindVariable)

	// Create a new reference
	ref := NewReference(symbol, file, token.Pos(10), token.Pos(20))

	// Default should be read (IsWrite = false)
	if ref.IsWrite != false {
		t.Errorf("Default Reference.IsWrite = %v, want false", ref.IsWrite)
	}

	// Set to write
	ref.SetIsWrite(true)

	// Verify IsWrite was set
	if ref.IsWrite != true {
		t.Errorf("After SetIsWrite(true), Reference.IsWrite = %v, want true", ref.IsWrite)
	}
}

func TestGetReferencePosition(t *testing.T) {
	// Setup test data
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	file.FileSet = token.NewFileSet()
	symbol := NewSymbol("TestSymbol", KindFunction)

	// Create a new reference
	ref := NewReference(symbol, file, token.Pos(10), token.Pos(20))

	// Test GetPosition
	// Since we're not using a real FileSet with a real file, this should return nil
	posInfo := ref.GetPosition()
	if posInfo != nil {
		t.Errorf("GetPosition should return nil with mock FileSet")
	}

	// Test with nil file
	ref.File = nil
	posInfo = ref.GetPosition()
	if posInfo != nil {
		t.Errorf("GetPosition should return nil with nil file")
	}
}

func TestReferencesFinder(t *testing.T) {
	// Setup test data
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	symbol := NewSymbol("TestSymbol", KindFunction)
	ref := NewReference(symbol, file, token.Pos(10), token.Pos(20))

	// Add the package to the module
	module.Packages[pkg.ImportPath] = pkg

	// Add the symbol to the package
	pkg.AddSymbol(symbol)

	// Create a references finder
	finder := &TypeAwareReferencesFinder{Module: module}

	// Test FindReferences
	refs, err := finder.FindReferences(symbol)
	if err != nil {
		t.Errorf("FindReferences returned error: %v", err)
	}

	if len(refs) != 1 || refs[0] != ref {
		t.Errorf("FindReferences returned %v, want [%v]", refs, ref)
	}

	// Test FindReferencesByName
	refsByName, err := finder.FindReferencesByName("TestSymbol")
	if err != nil {
		t.Errorf("FindReferencesByName returned error: %v", err)
	}

	if len(refsByName) != 1 || refsByName[0] != ref {
		t.Errorf("FindReferencesByName returned %v, want [%v]", refsByName, ref)
	}
}
