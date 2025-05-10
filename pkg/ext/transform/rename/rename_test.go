package rename

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"bitspark.dev/go-tree/pkg/ext/transform"
	"fmt"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
)

// createTestModule creates a test module with symbols for renaming tests
func createTestModule() *typesys.Module {
	module := &typesys.Module{
		Path:     "test/module",
		Dir:      "/test/module",
		Packages: make(map[string]*typesys.Package),
		FileSet:  nil, // In a real test, would initialize this
	}

	// Create a package
	pkg := &typesys.Package{
		Name:       "testpkg",
		ImportPath: "test/module/testpkg",
		Dir:        "/test/module/testpkg",
		Module:     module,
		Files:      make(map[string]*typesys.File),
		Symbols:    make(map[string]*typesys.Symbol),
	}
	module.Packages[pkg.ImportPath] = pkg

	// Create a file
	file := &typesys.File{
		Path:    "/test/module/testpkg/file.go",
		Name:    "file.go",
		Package: pkg,
		Symbols: []*typesys.Symbol{}, // Will add symbols here
	}
	pkg.Files[file.Path] = file

	// Create a variable symbol
	varSymbol := &typesys.Symbol{
		ID:      "var_oldName",
		Name:    "oldName",
		Kind:    typesys.KindVariable,
		File:    file,
		Package: pkg,
		// In a real test, would set positions
	}

	// Create a function symbol
	funcSymbol := &typesys.Symbol{
		ID:      "func_doSomething",
		Name:    "doSomething",
		Kind:    typesys.KindFunction,
		File:    file,
		Package: pkg,
		// In a real test, would set positions
	}

	// Create a type symbol
	typeSymbol := &typesys.Symbol{
		ID:      "type_TestType",
		Name:    "TestType",
		Kind:    typesys.KindStruct,
		File:    file,
		Package: pkg,
		// In a real test, would set positions
	}

	// Add symbols to file
	file.Symbols = append(file.Symbols, varSymbol, funcSymbol, typeSymbol)

	// Add symbols to package
	pkg.Symbols[varSymbol.ID] = varSymbol
	pkg.Symbols[funcSymbol.ID] = funcSymbol
	pkg.Symbols[typeSymbol.ID] = typeSymbol

	// Create some references
	ref1 := &typesys.Reference{
		Symbol:  varSymbol,
		File:    file,
		Context: funcSymbol, // Reference inside the function
		// In a real test, would set positions
	}

	ref2 := &typesys.Reference{
		Symbol:  varSymbol,
		File:    file,
		Context: typeSymbol, // Reference inside the type
		// In a real test, would set positions
	}

	// Add references to symbols
	varSymbol.References = []*typesys.Reference{ref1, ref2}

	return module
}

// createTestIndex creates an index for the test module
func createTestIndex(module *typesys.Module) *index.Index {
	// Create a new index
	idx := index.NewIndex(module)

	// Instead of trying to access private fields, we'll manually add each symbol
	// to the original module, then register them through the public Build() method

	// First, ensure the module's file objects have their symbols properly registered
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			// Make sure each file has references to all its symbols
			for _, sym := range pkg.Symbols {
				if sym.File != nil && sym.File.Path == file.Path {
					file.Symbols = append(file.Symbols, sym)
				}
			}
		}
	}

	// Now build the index properly
	err := idx.Build()
	if err != nil {
		// If the build fails, this is a test setup issue
		panic(fmt.Sprintf("Failed to build index in test: %v", err))
	}

	return idx
}

// TestSymbolRenamerTransform tests the Symbol renamer's Transform method
func TestSymbolRenamerTransform(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := createTestIndex(module)

	// Find the symbol to rename
	varSymbol := module.Packages["test/module/testpkg"].Symbols["var_oldName"]
	assert.NotNil(t, varSymbol)

	// Create the symbol renamer
	renamer := NewSymbolRenamer(varSymbol.ID, "newName")

	// Create transformation context
	ctx := transform.NewContext(module, idx, false)

	// Apply the transformation
	result, err := renamer.Transform(ctx)

	// Verify transformation result
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "newName", varSymbol.Name)
	assert.Len(t, result.AffectedFiles, 1)
	assert.Len(t, result.Changes, 3) // One for definition, two for references
}

// TestSymbolRenamerDryRun tests the SymbolRenamer in dry run mode
func TestSymbolRenamerDryRun(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := createTestIndex(module)

	// Find the symbol to rename
	varSymbol := module.Packages["test/module/testpkg"].Symbols["var_oldName"]
	assert.NotNil(t, varSymbol)

	// Create the symbol renamer
	renamer := NewSymbolRenamer(varSymbol.ID, "newName")

	// Create transformation context with dry run enabled
	ctx := transform.NewContext(module, idx, true)

	// Apply the transformation
	result, err := renamer.Transform(ctx)

	// Verify transformation result
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.IsDryRun)

	// In dry run mode, the original symbol should not be changed
	assert.Equal(t, "oldName", varSymbol.Name)

	// But we should have changes in the result
	assert.Len(t, result.Changes, 3) // One for definition, two for references
}

// TestSymbolRenamerValidate tests the SymbolRenamer's validation
func TestSymbolRenamerValidate(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := createTestIndex(module)

	// Find the symbol to rename
	varSymbol := module.Packages["test/module/testpkg"].Symbols["var_oldName"]
	assert.NotNil(t, varSymbol)

	// Create a valid renamer
	validRenamer := NewSymbolRenamer(varSymbol.ID, "validName")

	// Create transformation context
	ctx := transform.NewContext(module, idx, false)

	// Test valid rename
	err := validRenamer.Validate(ctx)
	assert.NoError(t, err)

	// Test invalid rename (empty name)
	invalidRenamer := NewSymbolRenamer(varSymbol.ID, "")
	err = invalidRenamer.Validate(ctx)
	assert.Error(t, err)

	// Test non-existent symbol
	nonExistentRenamer := NewSymbolRenamer("non_existent_id", "newName")
	err = nonExistentRenamer.Validate(ctx)
	assert.Error(t, err)
}

// TestSymbolRenamerNameAndDescription tests the Name and Description methods
func TestSymbolRenamerNameAndDescription(t *testing.T) {
	// Create the symbol renamer
	renamer := NewSymbolRenamer("some_id", "newName")

	// Test Name method
	assert.Equal(t, "SymbolRenamer", renamer.Name())

	// Test Description method
	assert.Contains(t, renamer.Description(), "newName")
}
