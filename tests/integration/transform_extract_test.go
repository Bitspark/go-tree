//go:build integration

// Package integration contains integration tests that span multiple packages.
package integration

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"bitspark.dev/go-tree/pkg/ext/transform"
	extract2 "bitspark.dev/go-tree/pkg/ext/transform/extract"
	"bitspark.dev/go-tree/pkg/io/loader"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractTransformer focuses specifically on the extract transformer
func TestExtractTransformer(t *testing.T) {
	// Create a test module with structs that have common method patterns
	modDir, cleanup := setupExtractTestModule(t)
	defer cleanup()

	fmt.Println("Created test module at:", modDir)

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

	// Create transformer context
	ctx := transform.NewContext(module, idx, true) // Dry run mode

	// Create an interface extractor with extremely permissive options
	options := extract2.DefaultOptions()
	options.MinimumTypes = 2      // Only require 2 types to have a common pattern
	options.MinimumMethods = 1    // Only require 1 common method
	options.MethodThreshold = 0.1 // Very low threshold

	extractor := extract2.NewInterfaceExtractor(options)

	// Validate the transformer
	err = extractor.Validate(ctx)
	require.NoError(t, err, "Extract validation failed")

	// Execute the transformer
	result, err := extractor.Transform(ctx)
	require.NoError(t, err, "Extract transformation failed")

	// Print the transformation result
	fmt.Println("\n=== Extract Result ===")
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
	}

	// Verify extract was successful
	assert.True(t, result.Success, "Extract should succeed")
	assert.Greater(t, len(result.Changes), 0, "Should find at least one interface")

	// Look for expected interfaces by checking for the word "interface" in the New field
	found := false
	for _, change := range result.Changes {
		if change.New != "" && (contains(change.New, "interface") ||
			contains(change.New, "Reader") ||
			contains(change.New, "Executor")) {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find at least one interface definition")
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s == substr || (len(s) >= len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr)
}

// setupExtractTestModule creates a test module for interface extraction tests
func setupExtractTestModule(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "extract-test-*")
	require.NoError(t, err, "Failed to create temp directory")

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Create go.mod file
	goModContent := `module example.com/extracttest

go 1.20
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err, "Failed to write go.mod")

	// Create readers.go with types that have a common Read method
	readersContent := `package extracttest

// FileReader reads from a file
type FileReader struct {
	path string
}

// Read reads from the file into the buffer
func (r *FileReader) Read(buf []byte) (int, error) {
	return 0, nil
}

// StringReader reads from a string
type StringReader struct {
	data string
	pos  int
}

// Read reads from the string into the buffer
func (s *StringReader) Read(buf []byte) (int, error) {
	return 0, nil
}
`
	err = os.WriteFile(filepath.Join(tempDir, "readers.go"), []byte(readersContent), 0644)
	require.NoError(t, err, "Failed to write readers.go")

	// Create executors.go with types that have a common Execute method
	executorsContent := `package extracttest

// Task represents a task that can be executed
type Task struct {
	name string
}

// Execute runs the task
func (t *Task) Execute() error {
	return nil
}

// Job represents a background job
type Job struct {
	id string
}

// Execute runs the job
func (j *Job) Execute() error {
	return nil
}
`
	err = os.WriteFile(filepath.Join(tempDir, "executors.go"), []byte(executorsContent), 0644)
	require.NoError(t, err, "Failed to write executors.go")

	// Create main.go
	mainContent := `package extracttest

func main() {
	// Use readers
	fr := &FileReader{path: "test.txt"}
	sr := &StringReader{data: "test data"}
	
	buf := make([]byte, 10)
	fr.Read(buf)
	sr.Read(buf)
	
	// Use executors
	task := &Task{name: "sample task"}
	job := &Job{id: "job-1"}
	
	task.Execute()
	job.Execute()
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	require.NoError(t, err, "Failed to write main.go")

	return tempDir, cleanup
}
