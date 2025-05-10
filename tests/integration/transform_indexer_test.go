//go:build integration

// Package integration contains integration tests that span multiple packages.
package integration

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"bitspark.dev/go-tree/pkg/io/loader"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIndexerWithInterfaces focuses on testing if the indexer properly identifies interface implementations
func TestIndexerWithInterfaces(t *testing.T) {
	// Create a test module with interfaces and implementations
	modDir, cleanup := setupIndexerInterfaceTestModule(t)
	defer cleanup()

	fmt.Println("Created test directory at:", modDir)

	// Print the directory structure and file contents
	fmt.Println("=== Test module structure ===")
	printDirContents(modDir, "")
	printFileContents(filepath.Join(modDir, "interfaces.go"))

	// Load the module with the loader
	module, err := loader.LoadModule(modDir, nil)
	require.NoError(t, err, "Failed to load module")

	// Verify module was loaded correctly
	assert.Equal(t, "example.com/indexertest", module.Path, "Module path is incorrect")

	// Create an index for the module
	idx := index.NewIndex(module)
	err = idx.Build()
	require.NoError(t, err, "Failed to build index")

	// Print all the symbols found by the indexer
	fmt.Println("\n=== Symbols in the index ===")
	printAllSymbols(idx)

	// Find the Reader interface
	readerInterfaceSymbols := idx.FindSymbolsByName("Reader")
	if assert.GreaterOrEqual(t, len(readerInterfaceSymbols), 1, "Reader interface should be found") {
		readerInterface := findSymbolOfKind(readerInterfaceSymbols, typesys.KindInterface)
		if assert.NotNil(t, readerInterface, "Reader symbol should be an interface") {
			fmt.Printf("Found Reader interface: %s (ID: %s)\n", readerInterface.Name, readerInterface.ID)

			// Try to directly check method presence on interface
			interfaceMethods := findMethodsForInterface(idx, readerInterface)
			for _, method := range interfaceMethods {
				fmt.Printf("Interface method: %s\n", method.Name)
			}
		}
	}

	// Find the implementations
	fileReaderSymbols := idx.FindSymbolsByName("FileReader")
	if assert.GreaterOrEqual(t, len(fileReaderSymbols), 1, "FileReader type should be found") {
		fileReader := findSymbolOfKind(fileReaderSymbols, typesys.KindStruct)
		if assert.NotNil(t, fileReader, "FileReader symbol should be a struct") {
			fmt.Printf("Found FileReader: %s (ID: %s)\n", fileReader.Name, fileReader.ID)

			// Find methods on the struct
			methods := findMethodsOnType(idx, fileReader)
			fmt.Printf("Found %d methods on FileReader:\n", len(methods))
			for _, method := range methods {
				fmt.Printf("  Method: %s\n", method.Name)
			}

			// Verify it has the Read method
			assert.Contains(t, extractNames(methods), "Read", "FileReader should have Read method")
		}
	}

	// Find StringReader and its methods
	stringReaderSymbols := idx.FindSymbolsByName("StringReader")
	if assert.GreaterOrEqual(t, len(stringReaderSymbols), 1, "StringReader type should be found") {
		stringReader := findSymbolOfKind(stringReaderSymbols, typesys.KindStruct)
		if assert.NotNil(t, stringReader, "StringReader symbol should be a struct") {
			fmt.Printf("Found StringReader: %s (ID: %s)\n", stringReader.Name, stringReader.ID)

			// Find methods on the struct
			methods := findMethodsOnType(idx, stringReader)
			fmt.Printf("Found %d methods on StringReader:\n", len(methods))
			for _, method := range methods {
				fmt.Printf("  Method: %s\n", method.Name)
			}

			// Verify it has the Read method
			assert.Contains(t, extractNames(methods), "Read", "StringReader should have Read method")
		}
	}

	// Test functionality for finding method implementations
	fmt.Println("\n=== Checking implementation relationships ===")

	// Find all Read methods
	readMethods := idx.FindSymbolsByName("Read")
	fmt.Printf("Found %d Read methods\n", len(readMethods))
	for i, method := range readMethods {
		parent := ""
		if method.Parent != nil {
			parent = method.Parent.Name
		}
		fmt.Printf("  Read method %d: Kind=%v, Parent=%s\n", i, method.Kind, parent)
	}

	// Print symbols by kind for interfaces and methods
	printSymbolsByKind(idx, typesys.KindInterface, "Interface")
	printSymbolsByKind(idx, typesys.KindMethod, "Method")
}

// Helper function to print all symbols in the index by kind
func printAllSymbols(idx *index.Index) {
	fmt.Println("Interfaces:")
	printSymbolsByKind(idx, typesys.KindInterface, "")

	fmt.Println("\nStructs:")
	printSymbolsByKind(idx, typesys.KindStruct, "")

	fmt.Println("\nMethods:")
	printSymbolsByKind(idx, typesys.KindMethod, "")

	fmt.Println("\nFunctions:")
	printSymbolsByKind(idx, typesys.KindFunction, "")

	fmt.Println("\nFields:")
	printSymbolsByKind(idx, typesys.KindField, "")
}

// Helper function to print symbols of a certain kind
func printSymbolsByKind(idx *index.Index, kind typesys.SymbolKind, prefix string) {
	symbols := idx.FindSymbolsByKind(kind)
	for _, s := range symbols {
		parent := ""
		if s.Parent != nil {
			parent = fmt.Sprintf(" (Parent: %s)", s.Parent.Name)
		}
		fmt.Printf("%s%s%s (ID: %s)\n", prefix, s.Name, parent, s.ID)
	}
}

// Helper function to find methods defined on a type
func findMethodsOnType(idx *index.Index, typeSymbol *typesys.Symbol) []*typesys.Symbol {
	var methods []*typesys.Symbol

	allMethods := idx.FindSymbolsByKind(typesys.KindMethod)
	for _, method := range allMethods {
		if method.Parent != nil && method.Parent.ID == typeSymbol.ID {
			methods = append(methods, method)
		}
	}

	return methods
}

// Helper function to find methods for an interface
func findMethodsForInterface(idx *index.Index, interfaceSymbol *typesys.Symbol) []*typesys.Symbol {
	var methods []*typesys.Symbol

	// In a real implementation, this would use the type system's interface methods API
	// For now, we just look for methods with the same package and no parent
	allMethods := idx.FindSymbolsByKind(typesys.KindMethod)
	for _, method := range allMethods {
		if method.Package == interfaceSymbol.Package && method.Parent == nil {
			methods = append(methods, method)
		}
	}

	return methods
}

// Helper function to find a symbol of a specific kind from a list
func findSymbolOfKind(symbols []*typesys.Symbol, kind typesys.SymbolKind) *typesys.Symbol {
	for _, sym := range symbols {
		if sym.Kind == kind {
			return sym
		}
	}
	return nil
}

// Helper function to extract names from a list of symbols
func extractNames(symbols []*typesys.Symbol) []string {
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	return names
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

// setupIndexerInterfaceTestModule creates a Go module with interfaces and implementations
func setupIndexerInterfaceTestModule(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create go.mod file
	goModContent := `module example.com/indexertest

go 1.20
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err, "Failed to write go.mod")

	// Create a single file with interfaces and implementations
	interfacesContent := `package indexertest

// Reader is an interface for types that can read data
type Reader interface {
	Read(p []byte) (int, error)
}

// FileReader implements Reader for files
type FileReader struct {
	path string
}

// Read reads data from a file into p
func (fr *FileReader) Read(p []byte) (int, error) {
	return 0, nil
}

// StringReader implements Reader for strings
type StringReader struct {
	data string
	pos  int
}

// Read reads data from a string into p
func (sr *StringReader) Read(p []byte) (int, error) {
	return 0, nil
}

// main is the entry point
func main() {
	// Create readers
	fr := &FileReader{path: "test.txt"}
	sr := &StringReader{data: "test data"}
	
	// Use them
	buf := make([]byte, 10)
	fr.Read(buf)
	sr.Read(buf)
}
`
	err = os.WriteFile(filepath.Join(tempDir, "interfaces.go"), []byte(interfacesContent), 0644)
	require.NoError(t, err, "Failed to write interfaces.go")

	return tempDir, cleanup
}
