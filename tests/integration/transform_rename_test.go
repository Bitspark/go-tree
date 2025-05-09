//go:build integration

// Package integration contains integration tests that span multiple packages.
package integration

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"bitspark.dev/go-tree/pkg/ext/transform"
	"bitspark.dev/go-tree/pkg/ext/transform/rename"
	"bitspark.dev/go-tree/pkg/io/loader"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenameTransformer focuses specifically on the rename transformer
func TestRenameTransformer(t *testing.T) {
	// Create a simple test module with just a few types
	modDir, cleanup := setupSimpleRenameTestModule(t)
	defer cleanup()

	fmt.Println("Created test module at:", modDir)
	printFileContents(filepath.Join(modDir, "simple.go"))

	// Load the module
	module, err := loader.LoadModule(modDir, nil)
	require.NoError(t, err, "Failed to load module")

	// Create an index
	idx := index.NewIndex(module)
	err = idx.Build()
	require.NoError(t, err, "Failed to build index")

	// Print indexed symbols
	fmt.Println("\n=== Indexed Symbols ===")
	printSymbolsByKind(idx, typesys.KindInterface, "Interface")
	printSymbolsByKind(idx, typesys.KindStruct, "Struct")
	printSymbolsByKind(idx, typesys.KindMethod, "Method")
	printSymbolsByKind(idx, typesys.KindFunction, "Function")
	printSymbolsByKind(idx, typesys.KindField, "Field")

	// Find the Person struct
	personSymbols := idx.FindSymbolsByName("Person")
	if assert.GreaterOrEqual(t, len(personSymbols), 1, "Person type should be found") {
		// Get the first Person symbol
		personSymbol := personSymbols[0]
		fmt.Printf("\nFound Person: %s (ID: %s)\n", personSymbol.Name, personSymbol.ID)

		// Create a transformer context
		ctx := transform.NewContext(module, idx, true) // Dry run mode

		// Create a renamer to change Person to Individual
		renamer := rename.NewSymbolRenamer(personSymbol.ID, "Individual")

		// Validate the transformer
		err = renamer.Validate(ctx)
		require.NoError(t, err, "Rename validation failed")

		// Execute the transformer
		result, err := renamer.Transform(ctx)
		require.NoError(t, err, "Rename transformation failed")

		// Print the transformation result
		fmt.Println("\n=== Rename Result ===")
		fmt.Printf("Success: %v\n", result.Success)
		fmt.Printf("Summary: %s\n", result.Summary)
		fmt.Printf("Details: %s\n", result.Details)
		fmt.Printf("Files affected: %d\n", result.FilesAffected)
		for i, file := range result.AffectedFiles {
			fmt.Printf("  Affected file %d: %s\n", i+1, file)
		}

		// Print changes
		fmt.Printf("Changes: %d\n", len(result.Changes))
		for i, change := range result.Changes {
			fmt.Printf("  Change %d: '%s' -> '%s'\n", i+1, change.Original, change.New)
			if change.AffectedSymbol != nil {
				fmt.Printf("    Symbol: %s (ID: %s)\n", change.AffectedSymbol.Name, change.AffectedSymbol.ID)
			}
		}

		// Verify rename was successful
		assert.True(t, result.Success, "Rename should succeed")
		assert.Greater(t, len(result.Changes), 0, "Should have at least one change")
		assert.Equal(t, "Person", result.Changes[0].Original, "Original name should be Person")
		assert.Equal(t, "Individual", result.Changes[0].New, "New name should be Individual")
	}
}

// Helper function to print file contents
func printFileContents(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
		return
	}

	fmt.Printf("\n=== Content of %s ===\n", filepath.Base(path))
	fmt.Println(string(content))
}

// Helper function to print symbols of a certain kind
func printSymbolsByKind(idx *index.Index, kind typesys.SymbolKind, prefix string) {
	symbols := idx.FindSymbolsByKind(kind)
	fmt.Printf("%ss (%d):\n", prefix, len(symbols))
	for _, s := range symbols {
		parent := ""
		if s.Parent != nil {
			parent = fmt.Sprintf(" (Parent: %s)", s.Parent.Name)
		}
		fmt.Printf("  %s%s (ID: %s)\n", s.Name, parent, s.ID)
	}
}

// setupSimpleRenameTestModule creates a minimal Go module for testing renaming
func setupSimpleRenameTestModule(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "rename-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create go.mod file
	goModContent := `module example.com/renametest

go 1.20
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err, "Failed to write go.mod")

	// Create a simple file with types that can be renamed
	simpleContent := `package renametest

import (
	"fmt"
)

// Person represents a person in the system
type Person struct {
	Name    string
	Age     int
	Address string
}

// Greet returns a greeting for the person
func (p *Person) Greet() string {
	return fmt.Sprintf("Hello, my name is %s", p.Name)
}

// CreatePerson creates a new person with the given name and age
func CreatePerson(name string, age int) *Person {
	return &Person{
		Name: name,
		Age:  age,
	}
}

func main() {
	// Create a new person
	person := CreatePerson("Alice", 30)
	
	// Greet the person
	fmt.Println(person.Greet())
}
`
	err = os.WriteFile(filepath.Join(tempDir, "simple.go"), []byte(simpleContent), 0644)
	require.NoError(t, err, "Failed to write simple.go")

	return tempDir, cleanup
}
