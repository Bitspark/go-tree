// Package integration contains integration tests that span multiple packages.
package integration

import (
	"bitspark.dev/go-tree/pkg/io/loader"
	"bitspark.dev/go-tree/pkg/io/saver"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestLoaderSaverRoundTrip tests the roundtrip from loader to saver and back.
func TestLoaderSaverRoundTrip(t *testing.T) {
	// Create a simple Go module for testing
	modDir, cleanup := setupTestModule(t)
	defer cleanup()

	// Load the module with the loader
	module, err := loader.LoadModule(modDir, nil)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify module was loaded correctly
	if module.Path != "example.com/testmod" {
		t.Errorf("Expected module path 'example.com/testmod', got '%s'", module.Path)
	}

	// Find the main.go file for modification
	var mainFile *typesys.File
	var mainPkg *typesys.Package
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			if file.Name == "main.go" {
				mainFile = file
				mainPkg = pkg
				break
			}
		}
		if mainFile != nil {
			break
		}
	}

	if mainFile == nil {
		t.Fatal("main.go file not found in loaded module")
	}

	// Add a new function to the file
	newFunc := &typesys.Symbol{
		ID:       "newFuncID",
		Name:     "NewFunction",
		Kind:     typesys.KindFunction,
		Exported: true,
		Package:  mainPkg,
		File:     mainFile,
	}
	mainFile.Symbols = append(mainFile.Symbols, newFunc)

	// Create a directory to save the modified module
	outDir, err := os.MkdirTemp("", "integration-savedir-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to clean up output directory: %v", err)
		}
	}()

	// Save the modified module
	moduleSaver := saver.NewGoModuleSaver()
	err = moduleSaver.SaveTo(module, outDir)
	if err != nil {
		t.Fatalf("Failed to save module: %v", err)
	}

	// Verify the saved file contains our changes
	mainPath := filepath.Join(outDir, "main.go")
	content, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("Failed to read saved main.go: %v", err)
	}

	if !strings.Contains(string(content), "func NewFunction") {
		t.Error("Saved file doesn't contain the new function we added")
	}

	// Reload the saved module to verify it can be processed correctly
	reloadedModule, err := loader.LoadModule(outDir, nil)
	if err != nil {
		t.Fatalf("Failed to reload saved module: %v", err)
	}

	// Verify the reloaded module has our changes
	var foundNewFunc bool
	for _, pkg := range reloadedModule.Packages {
		for _, file := range pkg.Files {
			for _, sym := range file.Symbols {
				if sym.Kind == typesys.KindFunction && sym.Name == "NewFunction" {
					foundNewFunc = true
					break
				}
			}
			if foundNewFunc {
				break
			}
		}
		if foundNewFunc {
			break
		}
	}

	if !foundNewFunc {
		t.Error("Reloaded module doesn't contain the new function we added")
	}
}

// TestModifyAndSave tests modifying a loaded module and saving the changes.
func TestModifyAndSave(t *testing.T) {
	// Create a simple Go module for testing
	modDir, cleanup := setupTestModule(t)
	defer cleanup()

	// Load the module
	module, err := loader.LoadModule(modDir, nil)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Find a package to modify
	var mainPkg *typesys.Package
	for importPath, pkg := range module.Packages {
		if strings.HasSuffix(importPath, "testmod") {
			mainPkg = pkg
			break
		}
	}

	if mainPkg == nil {
		t.Fatal("Main package not found in loaded module")
	}

	// Create a new file in the package
	newFilePath := filepath.Join(mainPkg.Module.Dir, "newfile.go")
	newFile := &typesys.File{
		Path:    newFilePath,
		Name:    "newfile.go",
		Package: mainPkg,
		Symbols: make([]*typesys.Symbol, 0),
	}

	// Add a type to the new file
	newType := &typesys.Symbol{
		ID:       "newTypeID",
		Name:     "NewType",
		Kind:     typesys.KindType,
		Exported: true,
		Package:  mainPkg,
		File:     newFile,
	}
	newFile.Symbols = append(newFile.Symbols, newType)

	// Add the file to the package
	mainPkg.Files[newFilePath] = newFile

	// Create an output directory
	outDir, err := os.MkdirTemp("", "integration-modifysave-*")
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Logf("Failed to clean up output directory: %v", err)
		}
	}()

	// Save the modified module
	moduleSaver := saver.NewGoModuleSaver()
	err = moduleSaver.SaveTo(module, outDir)
	if err != nil {
		t.Fatalf("Failed to save module: %v", err)
	}

	// Verify the new file was created
	newFileSavedPath := filepath.Join(outDir, "newfile.go")
	if _, err := os.Stat(newFileSavedPath); os.IsNotExist(err) {
		t.Error("New file was not created in the saved module")
	}

	// Check the contents of the new file
	content, err := os.ReadFile(newFileSavedPath)
	if err != nil {
		t.Fatalf("Failed to read new file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "package testmod") {
		t.Error("New file does not contain correct package declaration")
	}

	if !strings.Contains(contentStr, "type NewType") {
		t.Error("New file does not contain the type we added")
	}
}

// setupTestModule creates a temporary Go module for testing.
func setupTestModule(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create cleanup function
	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}

	// Create go.mod file
	goModContent := `module example.com/testmod

go 1.18
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create main.go file
	mainContent := `package testmod

// TestFunc is a test function
func TestFunc() string {
	return "test"
}

// ExampleType is a test type
type ExampleType struct {
	Name string
	ID   int
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainContent), 0644)
	if err != nil {
		cleanup()
		t.Fatalf("Failed to write main.go: %v", err)
	}

	return tempDir, cleanup
}
