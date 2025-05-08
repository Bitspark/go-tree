package index

import (
	"testing"

	"bitspark.dev/go-tree/pkgold/core/loader"
)

// TestBuildIndex tests that we can successfully build an index from a module.
func TestBuildIndex(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Create indexer and build the index
	indexer := NewIndexer(mod)
	idx, err := indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	// Verify index was created and contains symbols
	if idx == nil {
		t.Fatal("Expected index to be created")
	}

	// Check that we've got symbols
	if len(idx.SymbolsByName) == 0 {
		t.Error("Expected index to contain symbols by name")
	}

	if len(idx.SymbolsByFile) == 0 {
		t.Error("Expected index to contain symbols by file")
	}
}

// TestFindSymbolsByName tests finding symbols by their name.
func TestFindSymbolsByName(t *testing.T) {
	// Load and index the test module
	idx := buildTestIndex(t)

	// Test finding common symbols
	userSymbols := idx.FindSymbolsByName("User")
	if len(userSymbols) == 0 {
		t.Fatal("Expected to find User type")
	}

	// Verify the symbol's properties
	userSymbol := userSymbols[0]
	if userSymbol.Kind != KindType {
		t.Errorf("Expected User to be a type, got %v", userSymbol.Kind)
	}

	// Test finding a function
	newUserSymbols := idx.FindSymbolsByName("NewUser")
	if len(newUserSymbols) == 0 {
		t.Fatal("Expected to find NewUser function")
	}

	newUserSymbol := newUserSymbols[0]
	if newUserSymbol.Kind != KindFunction {
		t.Errorf("Expected NewUser to be a function, got %v", newUserSymbol.Kind)
	}

	// Test finding a variable
	defaultTimeoutSymbols := idx.FindSymbolsByName("DefaultTimeout")
	if len(defaultTimeoutSymbols) == 0 {
		t.Fatal("Expected to find DefaultTimeout variable")
	}

	defaultTimeoutSymbol := defaultTimeoutSymbols[0]
	if defaultTimeoutSymbol.Kind != KindVariable {
		t.Errorf("Expected DefaultTimeout to be a variable, got %v", defaultTimeoutSymbol.Kind)
	}
}

// TestFindSymbolsForType tests finding symbols related to a specific type.
func TestFindSymbolsForType(t *testing.T) {
	idx := buildTestIndex(t)

	// Find methods and fields for the User type
	userSymbols := idx.FindSymbolsForType("User")
	if len(userSymbols) == 0 {
		t.Fatal("Expected to find symbols for User type")
	}

	// Check that we found methods
	methodCount := 0
	fieldCount := 0
	for _, sym := range userSymbols {
		if sym.Kind == KindMethod {
			methodCount++
		} else if sym.Kind == KindField {
			fieldCount++
		}
	}

	// The sample package should have at least some methods and fields for User
	if methodCount == 0 {
		t.Error("Expected to find methods for User type")
	}

	if fieldCount == 0 {
		t.Error("Expected to find fields for User type")
	}
}

// TestFindReferences tests finding references to a symbol.
func TestFindReferences(t *testing.T) {
	// Skip this test for now as reference detection needs more work
	t.Skip("Reference detection is not fully implemented yet")

	idx := buildTestIndex(t)

	// Find a symbol first - use ErrInvalidCredentials which is referenced in the Login method
	errCredentialsSymbols := idx.FindSymbolsByName("ErrInvalidCredentials")
	if len(errCredentialsSymbols) == 0 {
		t.Fatal("Expected to find ErrInvalidCredentials variable")
	}

	// Find references to that symbol
	references := idx.FindReferences(errCredentialsSymbols[0])

	// There should be at least one reference to ErrInvalidCredentials in the Login function
	if len(references) == 0 {
		t.Error("Expected to find at least one reference to ErrInvalidCredentials")
	}
}

// TestSymbolKindCounts tests that we index different kinds of symbols correctly.
func TestSymbolKindCounts(t *testing.T) {
	idx := buildTestIndex(t)

	// Count symbols by kind
	kindCounts := make(map[SymbolKind]int)
	for _, symbols := range idx.SymbolsByName {
		for _, symbol := range symbols {
			kindCounts[symbol.Kind]++
		}
	}

	// We expect to find at least one of each kind (except maybe parameters)
	expectedKinds := []SymbolKind{
		KindFunction,
		KindMethod,
		KindType,
		KindVariable,
		KindConstant,
		KindField,
		KindImport,
	}

	for _, kind := range expectedKinds {
		if kindCounts[kind] == 0 {
			t.Errorf("Expected to find at least one symbol of kind %v", kind)
		}
	}
}

// TestFindSymbolAtPosition tests finding a symbol at a specific position.
func TestFindSymbolAtPosition(t *testing.T) {
	// This is a bit trickier because we need specific position information
	// from a known file. Let's find a symbol first and then use its position.
	idx := buildTestIndex(t)

	// Find a symbol with position info
	userSymbols := idx.FindSymbolsByName("User")
	if len(userSymbols) == 0 {
		t.Fatal("Expected to find User type")
	}

	userSymbol := userSymbols[0]
	if userSymbol.Pos == 0 {
		t.Skip("Symbol position information not available, skipping position lookup test")
	}

	// Try to find a symbol at the User type's position
	foundSymbol := idx.FindSymbolAtPosition(userSymbol.File, userSymbol.Pos)
	if foundSymbol == nil {
		t.Fatal("Expected to find a symbol at User's position")
	}

	// It should be the User type
	if foundSymbol.Name != "User" {
		t.Errorf("Expected to find User at position, got %s", foundSymbol.Name)
	}
}

// Helper function to build an index from test data
func buildTestIndex(t *testing.T) *Index {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Create indexer with all features enabled
	indexer := NewIndexer(mod).
		WithPrivate(true).
		WithTests(true)

	idx, err := indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build index: %v", err)
	}

	return idx
}
