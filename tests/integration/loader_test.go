// Package integration contains integration tests that span multiple packages.
package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoaderWithSimpleModule focuses on testing just the loader functionality
func TestLoaderWithSimpleModule(t *testing.T) {
	// Create a test module with a minimal structure
	modDir, cleanup := setupSimpleTestModule(t)
	defer cleanup()

	// Print the directory structure to verify files were created
	fmt.Println("=== Test module structure ===")
	printDirContents(modDir, "")

	// List actual content of go.mod and main.go
	goModPath := filepath.Join(modDir, "go.mod")
	mainPath := filepath.Join(modDir, "main.go")

	goModContent, err := os.ReadFile(goModPath)
	require.NoError(t, err, "Failed to read go.mod")
	fmt.Println("\n=== go.mod content ===")
	fmt.Println(string(goModContent))

	mainContent, err := os.ReadFile(mainPath)
	require.NoError(t, err, "Failed to read main.go")
	fmt.Println("\n=== main.go content ===")
	fmt.Println(string(mainContent))

	// Now load the module
	fmt.Println("\n=== Loading module ===")
	module, err := loader.LoadModule(modDir, nil)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Print debug info about the loaded module
	fmt.Println("\n=== Loaded module info ===")
	fmt.Printf("Module path: %s\n", module.Path)
	fmt.Printf("Module directory: %s\n", module.Dir)
	fmt.Printf("Number of packages: %d\n", len(module.Packages))

	// Check the loaded packages
	for importPath, pkg := range module.Packages {
		fmt.Printf("\nPackage: %s\n", importPath)
		fmt.Printf("  Directory: %s\n", pkg.Dir)
		fmt.Printf("  Number of files: %d\n", len(pkg.Files))

		// Check each file
		for filePath, file := range pkg.Files {
			fmt.Printf("  File: %s\n", filePath)
			fmt.Printf("    Number of symbols: %d\n", len(file.Symbols))
			for _, sym := range file.Symbols {
				fmt.Printf("    Symbol: %s (Kind: %v)\n", sym.Name, sym.Kind)
			}
		}
	}

	// Verify the module was loaded correctly
	assert.Equal(t, "example.com/loadertest", module.Path, "Module path is incorrect")
	assert.Greater(t, len(module.Packages), 0, "No packages were loaded")

	// Find the main package
	mainPkg, ok := module.Packages["example.com/loadertest"]
	if assert.True(t, ok, "Main package not found") {
		assert.Greater(t, len(mainPkg.Files), 0, "No files in main package")
		assert.Greater(t, len(mainPkg.Symbols), 0, "No symbols in main package")
	}
}

// Helper function to recursively print directory contents
func printDirContents(dir, indent string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("%sError reading directory: %v\n", indent, err)
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			fmt.Printf("%s[DIR] %s\n", indent, entry.Name())
			printDirContents(path, indent+"  ")
		} else {
			info, err := entry.Info()
			if err != nil {
				fmt.Printf("%s[FILE] %s (error getting info)\n", indent, entry.Name())
			} else {
				fmt.Printf("%s[FILE] %s (%d bytes)\n", indent, entry.Name(), info.Size())
			}
		}
	}
}

// setupSimpleTestModule creates a minimal Go module for testing the loader
func setupSimpleTestModule(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "loader-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Print the created directory
	fmt.Printf("Created test directory at: %s\n", tempDir)

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create go.mod file
	goModContent := `module example.com/loadertest

go 1.20
`
	goModPath := filepath.Join(tempDir, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	require.NoError(t, err, "Failed to write go.mod")

	// Create main.go file
	mainContent := `package loadertest

import (
	"fmt"
)

// Person represents a person
type Person struct {
	Name string
	Age  int
}

// Greet returns a greeting
func (p *Person) Greet() string {
	return fmt.Sprintf("Hello, my name is %s and I am %d years old", p.Name, p.Age)
}

// NewPerson creates a new person
func NewPerson(name string, age int) *Person {
	return &Person{
		Name: name,
		Age:  age,
	}
}

func main() {
	person := NewPerson("Alice", 30)
	fmt.Println(person.Greet())
}
`
	mainPath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(mainPath, []byte(mainContent), 0644)
	require.NoError(t, err, "Failed to write main.go")

	return tempDir, cleanup
}
