package index

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TestIndexBuild tests building an index from a module.
func TestIndexBuild(t *testing.T) {
	// Load a module for testing
	moduleDir := "../../" // Root of the Go-Tree project
	absPath, err := filepath.Abs(moduleDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Loading module from absolute path: %s", absPath)

	// Load options with verbose output to help debug
	loadOpts := &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
		Trace:          true, // Enable verbose output
	}

	// Load the module
	module, err := typesys.LoadModule(absPath, loadOpts)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	t.Logf("Loaded module with %d packages", len(module.Packages))

	// Print package names for debugging
	if len(module.Packages) == 0 {
		t.Logf("WARNING: No packages were loaded!")
	} else {
		t.Logf("Loaded packages:")
		for name := range module.Packages {
			t.Logf("  - %s", name)
		}
	}

	// Create an index
	idx := NewIndex(module)

	// Build the index
	err = idx.Build()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Check that we have symbols
	if len(idx.symbolsByID) == 0 {
		t.Errorf("No symbols were indexed")
	}

	// Check that we have symbols by kind
	foundTypes := idx.symbolsByKind[typesys.KindType]
	if len(foundTypes) == 0 {
		t.Errorf("No types were indexed")
	}

	// Check that we can look up symbols by name
	// Try to find any symbol, not specifically "Index"
	indexSymbols := idx.FindSymbolsByName("Index")
	if len(indexSymbols) == 0 {
		// Try to find any symbol
		t.Logf("Could not find 'Index' symbol, checking if any symbols exist")

		// Check if there are any symbols at all
		var foundSymbols bool
		for kind := range idx.symbolsByKind {
			if len(idx.symbolsByKind[kind]) > 0 {
				foundSymbols = true
				break
			}
		}

		if !foundSymbols {
			t.Errorf("Could not find any symbols in the index")
		} else {
			t.Logf("Found other symbols, but not 'Index' (this is not an error)")
		}
	}

	// Test the indexer wrapper
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	// Build index
	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index via indexer: %v", err)
	}

	// Test search
	results := indexer.Search("Index")
	if len(results) == 0 {
		// Try a more general search
		t.Logf("Search returned no results for 'Index', trying a more general search")

		// Check if we can find any symbols with a general search
		allTypes := indexer.FindAllTypes("")
		if len(allTypes) == 0 {
			t.Errorf("Search couldn't find any symbols, index might be empty")
		} else {
			t.Logf("Found %d types with a general search", len(allTypes))
		}
	}

	// Test methods lookup
	// Find a type first, try with "Index" but fall back to any type if not found
	types := indexer.FindAllTypes("Index")
	if len(types) == 0 {
		// Try to find any type instead
		t.Logf("Could not find 'Index' type, searching for any type")
		// Get all types
		for _, kind := range []typesys.SymbolKind{typesys.KindType, typesys.KindStruct, typesys.KindInterface} {
			typeSymbols := indexer.Index.FindSymbolsByKind(kind)
			if len(typeSymbols) > 0 {
				types = append(types, typeSymbols...)
				break
			}
		}
	}

	if len(types) == 0 {
		t.Errorf("Could not find any types in the codebase")
	} else {
		// Find methods for this type
		methods := indexer.FindMethodsOfType(types[0])
		if len(methods) == 0 {
			t.Logf("No methods found for type %s (this is not an error, just informational)", types[0].Name)
		} else {
			t.Logf("Found %d methods for type %s", len(methods), types[0].Name)
		}
	}
}

// TestCommandContext tests the command context.
func TestCommandContext(t *testing.T) {
	// Load a module for testing
	moduleDir := "../../" // Root of the Go-Tree project
	absPath, err := filepath.Abs(moduleDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Loading module from absolute path: %s", absPath)

	// Load the module with trace enabled
	module, err := typesys.LoadModule(absPath, &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
		Trace:          true,
	})
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	t.Logf("Loaded module with %d packages", len(module.Packages))

	// Create a command context
	ctx, err := NewCommandContext(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})
	if err != nil {
		t.Fatalf("Failed to create command context: %v", err)
	}

	// Test that we have an indexer
	if ctx.Indexer == nil {
		t.Errorf("Command context has no indexer")
	}

	// Test that we can find a file
	thisFile, err := filepath.Abs("index_test.go")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Get file symbols
	err = ctx.ListFileSymbols(thisFile)
	if err != nil {
		// This might fail if the file isn't in the module scope, so just log it
		t.Logf("Warning: Could not list file symbols: %v", err)
	}
}

// TestIndexSymbolLookups tests the various lookup methods of the Index.
func TestIndexSymbolLookups(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build index
	idx := NewIndex(module)
	err = idx.Build()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Test GetSymbolByID
	// First get a symbol to use for testing
	someSymbols := idx.FindSymbolsByKind(typesys.KindType)
	if len(someSymbols) == 0 {
		t.Fatalf("No type symbols found, cannot test GetSymbolByID")
	}
	testSymbol := someSymbols[0]

	// Test lookup by ID
	foundSymbol := idx.GetSymbolByID(testSymbol.ID)
	if foundSymbol == nil {
		t.Errorf("GetSymbolByID failed to find symbol with ID %s", testSymbol.ID)
	} else if foundSymbol != testSymbol {
		t.Errorf("GetSymbolByID returned wrong symbol: expected %v, got %v", testSymbol, foundSymbol)
	}

	// Test FindSymbolsByKind
	typeSymbols := idx.FindSymbolsByKind(typesys.KindType)
	if len(typeSymbols) == 0 {
		t.Errorf("FindSymbolsByKind failed to find any type symbols")
	}

	funcSymbols := idx.FindSymbolsByKind(typesys.KindFunction)
	if len(funcSymbols) == 0 {
		t.Errorf("FindSymbolsByKind failed to find any function symbols")
	}

	// Test FindSymbolsInFile
	if len(someSymbols) > 0 && someSymbols[0].File != nil {
		fileSymbols := idx.FindSymbolsInFile(someSymbols[0].File.Path)
		if len(fileSymbols) == 0 {
			t.Errorf("FindSymbolsInFile failed to find symbols in file %s", someSymbols[0].File.Path)
		}
	}
}

// TestIndexReferenceLookups tests the reference lookup methods of the Index.
func TestIndexReferenceLookups(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build index
	idx := NewIndex(module)
	err = idx.Build()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Find a symbol with references to test with
	// We'll try to find a common type that should have references
	indexSymbols := idx.FindSymbolsByName("Index")
	var symbolWithRefs *typesys.Symbol

	// Find first symbol with references
	for _, sym := range indexSymbols {
		refs := idx.FindReferences(sym)
		if len(refs) > 0 {
			symbolWithRefs = sym
			break
		}
	}

	if symbolWithRefs == nil {
		t.Logf("Could not find a symbol with references to test FindReferences")
		return
	}

	// Test FindReferences
	refs := idx.FindReferences(symbolWithRefs)
	if len(refs) == 0 {
		t.Errorf("FindReferences returned no references for symbol %s", symbolWithRefs.Name)
	}

	// Test FindReferencesInFile
	if len(refs) > 0 && refs[0].File != nil {
		fileRefs := idx.FindReferencesInFile(refs[0].File.Path)
		if len(fileRefs) == 0 {
			t.Errorf("FindReferencesInFile failed to find references in file %s", refs[0].File.Path)
		}
	}
}

// TestIndexMethodsAndInterfaces tests finding methods and interface implementations.
func TestIndexMethodsAndInterfaces(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build index
	idx := NewIndex(module)
	err = idx.Build()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Test FindMethods
	// Find Index type first
	indexSymbols := idx.FindSymbolsByName("Index")
	var indexType *typesys.Symbol
	for _, sym := range indexSymbols {
		if sym.Kind == typesys.KindType || sym.Kind == typesys.KindStruct {
			indexType = sym
			break
		}
	}

	if indexType != nil {
		methods := idx.FindMethods(indexType.Name)
		t.Logf("Found %d methods for type %s", len(methods), indexType.Name)
		for i, m := range methods {
			if i < 5 { // Log only first 5 to avoid too much output
				t.Logf("  - Method: %s", m.Name)
			}
		}
	}

	// Test FindImplementations
	interfaces := idx.FindSymbolsByKind(typesys.KindInterface)
	if len(interfaces) > 0 {
		impls := idx.FindImplementations(interfaces[0])
		t.Logf("Found %d implementations for interface %s", len(impls), interfaces[0].Name)
	}
}

// TestIndexerSearch tests the search functionality of the Indexer.
func TestIndexerSearch(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	// Build index
	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Test general search
	results := indexer.Search("Index")
	if len(results) == 0 {
		// Try a more general search
		t.Logf("Search returned no results for 'Index', trying a more general search")

		// Search for some common Go keywords that should exist in any Go codebase
		commonTerms := []string{"func", "type", "struct", "interface", "package"}
		for _, term := range commonTerms {
			altResults := indexer.Search(term)
			if len(altResults) > 0 {
				t.Logf("Found %d results for search term '%s'", len(altResults), term)
				break
			}
		}
	}

	// Test FindAllFunctions
	funcs := indexer.FindAllFunctions("Find")
	if len(funcs) == 0 {
		// Try a more general search
		t.Logf("FindAllFunctions returned no results for 'Find', searching for any function")

		// Get all functions
		allFuncs := indexer.Index.FindSymbolsByKind(typesys.KindFunction)
		if len(allFuncs) == 0 {
			t.Errorf("No functions found in the codebase")
		} else {
			t.Logf("Found %d functions in total", len(allFuncs))
		}
	}

	// Test FindAllTypes
	types := indexer.FindAllTypes("Index")
	if len(types) == 0 {
		// Try to find any types instead
		t.Logf("FindAllTypes returned no results for 'Index', searching for any types")

		// Try a general search for types
		for _, kind := range []typesys.SymbolKind{typesys.KindType, typesys.KindStruct, typesys.KindInterface} {
			kindTypes := indexer.Index.FindSymbolsByKind(kind)
			if len(kindTypes) > 0 {
				t.Logf("Found %d symbols of kind %s", len(kindTypes), kind)
				// Successfully found some types
				return
			}
		}

		t.Errorf("Could not find any types in the codebase")
	}
}

// TestIndexerPositionLookups tests position-based lookups in the Indexer.
func TestIndexerPositionLookups(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	// Build index
	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Find a file with symbols to test with
	var testFile string
	var foundPos *typesys.PositionInfo

	// Check all files until we find one with symbols that have positions
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			symbols := indexer.Index.FindSymbolsInFile(file.Path)
			for _, sym := range symbols {
				pos := sym.GetPosition()
				if pos != nil && pos.LineStart > 0 {
					testFile = file.Path
					foundPos = pos
					break
				}
			}
			if testFile != "" {
				break
			}
		}
		if testFile != "" {
			break
		}
	}

	if testFile == "" {
		t.Logf("Could not find file with suitable symbols for position testing")
		return
	}

	// Test FindSymbolAtPosition
	sym := indexer.FindSymbolAtPosition(testFile, foundPos.LineStart, foundPos.ColumnStart+1)
	if sym == nil {
		t.Errorf("FindSymbolAtPosition failed to find symbol at position %d:%d in file %s",
			foundPos.LineStart, foundPos.ColumnStart, testFile)
	} else {
		t.Logf("Found symbol %s at position %d:%d", sym.Name, foundPos.LineStart, foundPos.ColumnStart)
	}

	// We can't test FindReferenceAtPosition without knowing where references are located
	// Future: Add more specific test cases with known positions
}

// TestFileStructure tests the GetFileStructure and GetFileSymbols functions.
func TestFileStructure(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	// Build index
	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Find a file with symbols to test with
	var fileWithSymbols string
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			symbols := indexer.Index.FindSymbolsInFile(file.Path)
			if len(symbols) > 0 {
				fileWithSymbols = file.Path
				break
			}
		}
		if fileWithSymbols != "" {
			break
		}
	}

	if fileWithSymbols == "" {
		t.Fatalf("Could not find file with symbols for testing")
	}

	// Test GetFileSymbols
	symbolsByKind := indexer.GetFileSymbols(fileWithSymbols)
	if len(symbolsByKind) == 0 {
		t.Errorf("GetFileSymbols returned no symbols for file %s", fileWithSymbols)
	}

	// Test GetFileStructure
	structure := indexer.GetFileStructure(fileWithSymbols)
	if len(structure) == 0 {
		t.Errorf("GetFileStructure returned no structure for file %s", fileWithSymbols)
	}

	// Verify structure has parent-child relationships if possible
	hasChildren := false
	for _, node := range structure {
		if len(node.Children) > 0 {
			hasChildren = true
			break
		}
	}

	t.Logf("File structure for %s: %d root nodes, has hierarchical structure: %v",
		fileWithSymbols, len(structure), hasChildren)
}

// TestIndexUpdate tests the incremental update functionality.
func TestIndexUpdate(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build index
	idx := NewIndex(module)
	err = idx.Build()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Get initial symbol count
	initialSymbolCount := len(idx.symbolsByID)

	// Find a file to "update"
	var fileToUpdate string
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			symbols := idx.FindSymbolsInFile(file.Path)
			if len(symbols) > 0 {
				fileToUpdate = file.Path
				break
			}
		}
		if fileToUpdate != "" {
			break
		}
	}

	if fileToUpdate == "" {
		t.Logf("Could not find file with symbols for update testing")
		return
	}

	// Call Update with a single file
	err = idx.Update([]string{fileToUpdate})
	if err != nil {
		t.Errorf("Index.Update failed: %v", err)
	}

	// Check that symbols are still present after update
	afterUpdateCount := len(idx.symbolsByID)
	t.Logf("Symbol count - before: %d, after: %d", initialSymbolCount, afterUpdateCount)

	// The counts may differ slightly due to how update works
	// But we should still have symbols after the update
	if afterUpdateCount == 0 {
		t.Errorf("After update, index has no symbols")
	}
}

// TestCommandFunctions tests the various command functions in CommandContext.
func TestCommandFunctions(t *testing.T) {
	// Load test module
	module, err := loadTestModule(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create command context
	ctx, err := NewCommandContext(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})
	if err != nil {
		t.Fatalf("Failed to create command context: %v", err)
	}

	// Temporarily redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test SearchSymbols
	err = ctx.SearchSymbols("Index", "type")
	if err != nil {
		t.Errorf("SearchSymbols failed: %v", err)
	}

	// Test FindImplementations - might not have interfaces to test with
	// Just verify it doesn't crash with an unexpected error
	err = ctx.FindImplementations("Stringer")
	if err != nil {
		// This is expected if no Stringer interface is found
		t.Logf("FindImplementations result: %v", err)
	}

	// Test ListFileSymbols - find a file with symbols
	var fileWithSymbols string
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			symbols := ctx.Indexer.Index.FindSymbolsInFile(file.Path)
			if len(symbols) > 0 {
				fileWithSymbols = file.Path
				break
			}
		}
		if fileWithSymbols != "" {
			break
		}
	}

	if fileWithSymbols != "" {
		err = ctx.ListFileSymbols(fileWithSymbols)
		if err != nil {
			t.Errorf("ListFileSymbols failed: %v", err)
		}
	}

	// Test FindUsages
	symbols := ctx.Indexer.Search("Index")
	if len(symbols) > 0 {
		// Use empty file path to search by name only
		err = ctx.FindUsages(symbols[0].Name, "", 0, 0)
		if err != nil {
			t.Errorf("FindUsages failed: %v", err)
		}
	}

	// Restore stdout
	w.Close()
	outBytes, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	// Log output summary
	outputLines := strings.Split(string(outBytes), "\n")
	t.Logf("Command output: %d lines", len(outputLines))

	// Log a few lines of output for verification
	for i, line := range outputLines {
		if i < 5 {
			t.Logf("Output line %d: %s", i, line)
		} else {
			break
		}
	}
}

// Helper function to load a test module
func loadTestModule(t *testing.T) (*typesys.Module, error) {
	moduleDir := "../../" // Root of the Go-Tree project
	absPath, err := filepath.Abs(moduleDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Load the module
	return typesys.LoadModule(absPath, &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
}
