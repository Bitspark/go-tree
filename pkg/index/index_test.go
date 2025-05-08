package index

import (
	"path/filepath"
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
	// Use "Index" since we know that exists in our codebase
	indexSymbols := idx.FindSymbolsByName("Index")
	if len(indexSymbols) == 0 {
		t.Errorf("Could not find Index symbol")
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
		t.Errorf("Search returned no results")
	}

	// Test methods lookup
	// Find a type first
	types := indexer.FindAllTypes("Index")
	if len(types) == 0 {
		t.Errorf("Could not find any types matching 'Index'")
	} else {
		// Find methods for this type
		methods := indexer.FindMethodsOfType(types[0])
		// We might not have methods on every type, so just log it
		t.Logf("Found %d methods for type %s", len(methods), types[0].Name)
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
