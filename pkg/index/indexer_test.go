package index

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/typesys"
)

// MockSymbol represents a simplified symbol for testing
type MockSymbol struct {
	ID       string
	Name     string
	Kind     string
	FilePath string
	Parent   *MockSymbol
}

// MockReference represents a simplified reference for testing
type MockReference struct {
	SymbolID string
	FilePath string
	Line     int
	Column   int
}

// MockIndexer simulates an Indexer for testing
type MockIndexer struct {
	symbols    map[string]*MockSymbol   // by ID
	symbolsMap map[string][]*MockSymbol // by name
	refs       map[string][]*MockReference
}

// NewMockIndexer creates a new mock indexer for testing
func NewMockIndexer() *MockIndexer {
	return &MockIndexer{
		symbols:    make(map[string]*MockSymbol),
		symbolsMap: make(map[string][]*MockSymbol),
		refs:       make(map[string][]*MockReference),
	}
}

// AddSymbol adds a symbol to the mock indexer
func (m *MockIndexer) AddSymbol(sym *MockSymbol) {
	m.symbols[sym.ID] = sym
	m.symbolsMap[sym.Name] = append(m.symbolsMap[sym.Name], sym)
}

// AddReference adds a reference to the mock indexer
func (m *MockIndexer) AddReference(ref *MockReference) {
	m.refs[ref.SymbolID] = append(m.refs[ref.SymbolID], ref)
}

// FindSymbolsByName looks up symbols by name
func (m *MockIndexer) FindSymbolsByName(name string) []*MockSymbol {
	return m.symbolsMap[name]
}

// FindReferences looks up references for a symbol
func (m *MockIndexer) FindReferences(symbolID string) []*MockReference {
	return m.refs[symbolID]
}

// TestMockIndexerBasic tests the basic functionality of our mock indexer
func TestMockIndexerBasic(t *testing.T) {
	indexer := NewMockIndexer()

	// Test that we can create an indexer
	if indexer == nil {
		t.Fatal("Failed to create mock indexer")
	}

	// Test adding a symbol
	sym := &MockSymbol{
		ID:       "sym1",
		Name:     "TestSymbol",
		Kind:     "function",
		FilePath: "test.go",
	}
	indexer.AddSymbol(sym)

	// Test finding symbols by name
	results := indexer.FindSymbolsByName("TestSymbol")
	if len(results) != 1 {
		t.Errorf("Expected to find 1 symbol named 'TestSymbol', got %d", len(results))
	} else if results[0].ID != "sym1" {
		t.Errorf("Expected to find symbol with ID 'sym1', got '%s'", results[0].ID)
	}

	// Test finding non-existent symbol
	results = indexer.FindSymbolsByName("NonExistentSymbol")
	if len(results) != 0 {
		t.Errorf("Expected to find 0 symbols named 'NonExistentSymbol', got %d", len(results))
	}
}

// TestMockIndexerReferences tests reference tracking in our mock indexer
func TestMockIndexerReferences(t *testing.T) {
	indexer := NewMockIndexer()

	// Add a symbol
	sym := &MockSymbol{
		ID:       "sym1",
		Name:     "TestSymbol",
		Kind:     "function",
		FilePath: "test.go",
	}
	indexer.AddSymbol(sym)

	// Add references to the symbol
	ref1 := &MockReference{
		SymbolID: "sym1",
		FilePath: "main.go",
		Line:     10,
		Column:   20,
	}
	indexer.AddReference(ref1)

	ref2 := &MockReference{
		SymbolID: "sym1",
		FilePath: "util.go",
		Line:     30,
		Column:   15,
	}
	indexer.AddReference(ref2)

	// Test finding references
	refs := indexer.FindReferences("sym1")
	if len(refs) != 2 {
		t.Errorf("Expected to find 2 references, got %d", len(refs))
	}

	// Test finding references for non-existent symbol
	refs = indexer.FindReferences("nonexistent")
	if len(refs) != 0 {
		t.Errorf("Expected to find 0 references for non-existent symbol, got %d", len(refs))
	}
}

// TestMockIndexerHierarchy tests parent-child relationships in our mock indexer
func TestMockIndexerHierarchy(t *testing.T) {
	indexer := NewMockIndexer()

	// Create a type and method
	typeSymbol := &MockSymbol{
		ID:       "type1",
		Name:     "Person",
		Kind:     "type",
		FilePath: "person.go",
	}

	methodSymbol := &MockSymbol{
		ID:       "method1",
		Name:     "GetName",
		Kind:     "method",
		FilePath: "person.go",
		Parent:   typeSymbol,
	}

	// Add to indexer
	indexer.AddSymbol(typeSymbol)
	indexer.AddSymbol(methodSymbol)

	// Find the method
	methods := indexer.FindSymbolsByName("GetName")
	if len(methods) != 1 {
		t.Errorf("Expected to find 1 method named 'GetName', got %d", len(methods))
	} else {
		// Check parent relationship
		method := methods[0]
		if method.Parent == nil {
			t.Error("Method should have a parent")
		} else if method.Parent.Name != "Person" {
			t.Errorf("Method's parent should be 'Person', got '%s'", method.Parent.Name)
		}
	}
}

// TestMockFileOperations tests file-related operations
func TestMockFileOperations(t *testing.T) {
	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "indexer_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	})

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	content := []byte(`package test

type Person struct {
	Name string
}

func (p *Person) GetName() string {
	return p.Name
}
`)

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create mock indexer with symbols from our "file"
	indexer := NewMockIndexer()

	// Add symbols to simulate indexing the file
	typeSymbol := &MockSymbol{
		ID:       "type1",
		Name:     "Person",
		Kind:     "type",
		FilePath: testFile,
	}

	fieldSymbol := &MockSymbol{
		ID:       "field1",
		Name:     "Name",
		Kind:     "field",
		FilePath: testFile,
		Parent:   typeSymbol,
	}

	methodSymbol := &MockSymbol{
		ID:       "method1",
		Name:     "GetName",
		Kind:     "method",
		FilePath: testFile,
		Parent:   typeSymbol,
	}

	indexer.AddSymbol(typeSymbol)
	indexer.AddSymbol(fieldSymbol)
	indexer.AddSymbol(methodSymbol)

	// Test our mock "indexing" - we should be able to find the symbols
	symbols := indexer.FindSymbolsByName("Person")
	if len(symbols) != 1 {
		t.Errorf("Expected to find 1 symbol for Person, got %d", len(symbols))
	}

	symbols = indexer.FindSymbolsByName("GetName")
	if len(symbols) != 1 {
		t.Errorf("Expected to find 1 symbol for GetName, got %d", len(symbols))
	}

	// Verify the file exists and we can read it
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Test file does not exist")
	} else {
		// Read the file to verify content
		readContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Errorf("Failed to read test file: %v", err)
		} else if len(readContent) == 0 {
			t.Error("Test file is empty")
		}
	}
}

// Helper function to load a test module
func loadTestModuleFromPath(t *testing.T) (*typesys.Module, error) {
	moduleDir := "../../" // Root of the Go-Tree project
	absPath, err := filepath.Abs(moduleDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Load the module
	return loader.LoadModule(absPath, &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
}

func TestNewIndexer(t *testing.T) {
	// Load test module
	module, err := loadTestModuleFromPath(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})

	if indexer == nil {
		t.Fatal("NewIndexer should return a non-nil indexer")
	}

	if indexer.Module == nil {
		t.Error("Indexer should have a module")
	}
}

func TestBuildAndGetIndex(t *testing.T) {
	// Load test module
	module, err := loadTestModuleFromPath(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})

	// Build the index
	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Check that index has been built
	if indexer.Index == nil {
		t.Fatal("Index should be non-nil after BuildIndex")
	}

	// Verify index has our module
	if indexer.Index.Module != module {
		t.Error("Index doesn't have the correct module reference")
	}
}

func TestQueryFunctions(t *testing.T) {
	// Load test module
	module, err := loadTestModuleFromPath(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})

	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Test finding symbols by name
	symbols := indexer.Search("Person")
	if len(symbols) > 0 {
		t.Logf("Found %d symbols matching 'Person'", len(symbols))
	}

	// Test finding functions
	functions := indexer.FindAllFunctions("")
	if len(functions) > 0 {
		t.Logf("Found %d functions", len(functions))
	}

	// Test finding types
	types := indexer.FindAllTypes("Person")
	if len(types) > 0 {
		// Test finding methods for a type
		for _, typ := range types {
			methods := indexer.FindMethodsOfType(typ)
			t.Logf("Found %d methods for type %s", len(methods), typ.Name)
		}
	}
}

func TestFindSymbolAtPosition(t *testing.T) {
	// Load test module
	module, err := loadTestModuleFromPath(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})

	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Find a file with symbols
	var filePath string
	var line, column int

	// Try to find a file with symbols
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			symbols := indexer.Index.FindSymbolsInFile(file.Path)
			if len(symbols) > 0 && symbols[0].GetPosition() != nil {
				filePath = file.Path
				pos := symbols[0].GetPosition()
				line = pos.LineStart
				column = pos.ColumnStart
				break
			}
		}
		if filePath != "" {
			break
		}
	}

	if filePath != "" {
		// Test finding a symbol at position
		sym := indexer.FindSymbolAtPosition(filePath, line, column)
		if sym != nil {
			t.Logf("Found symbol %s at position %s:%d:%d", sym.Name, filePath, line, column)
		} else {
			t.Logf("No symbol found at position %s:%d:%d", filePath, line, column)
		}
	} else {
		t.Skip("No suitable file with symbol positions found for testing")
	}
}

func TestSearch(t *testing.T) {
	// Load test module
	module, err := loadTestModuleFromPath(t)
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build indexer
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})

	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	// Search for common Go terms
	terms := []string{"type", "struct", "func", "interface", "string"}
	for _, term := range terms {
		results := indexer.Search(term)
		t.Logf("Search for '%s' found %d results", term, len(results))

		if len(results) > 0 {
			// Successfully found some results
			break
		}
	}
}

func TestUpdateIndex(t *testing.T) {
	// Create a temporary directory for our test module
	tempDir, err := os.MkdirTemp("", "indexer-update-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	})

	// Create a simple Go module structure
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/indextest\n\ngo 1.18\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a sample Go file
	initialContent := `package indextest

// Person represents a person entity
type Person struct {
	Name string
	Age  int
}

// GetName returns the person's name
func (p *Person) GetName() string {
	return p.Name
}
`
	mainFile := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(mainFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Load the module
	module, err := loader.LoadModule(tempDir, &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
	})
	if err != nil {
		t.Fatalf("Failed to load test module: %v", err)
	}

	// Create and build the initial index
	indexer := NewIndexer(module, IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	err = indexer.BuildIndex()
	if err != nil {
		t.Fatalf("Failed to build initial index: %v", err)
	}

	// Verify that the Person type exists in the index
	personSymbols := indexer.Search("Person")
	if len(personSymbols) == 0 {
		t.Fatal("Person type not found in initial index")
	}

	// Verify GetName method exists
	getNameSymbols := indexer.Search("GetName")
	if len(getNameSymbols) == 0 {
		t.Fatal("GetName method not found in initial index")
	}

	// Now modify the file to add a new method
	updatedContent := initialContent + `
// GetAge returns the person's age
func (p *Person) GetAge() int {
	return p.Age
}
`
	err = os.WriteFile(mainFile, []byte(updatedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update main.go: %v", err)
	}

	// The indexer's UpdateIndex method takes a list of changed files
	changedFiles := []string{mainFile}

	// Update the index with the modified files
	err = indexer.UpdateIndex(changedFiles)
	if err != nil {
		t.Fatalf("Failed to update index: %v", err)
	}

	// Verify that the new GetAge method exists in the updated index
	getAgeSymbols := indexer.Search("GetAge")
	if len(getAgeSymbols) == 0 {
		t.Fatal("GetAge method not found in updated index")
	}

	// Make sure the original symbols still exist
	if len(indexer.Search("Person")) == 0 {
		t.Fatal("Person type lost during index update")
	}

	if len(indexer.Search("GetName")) == 0 {
		t.Fatal("GetName method lost during index update")
	}
}
