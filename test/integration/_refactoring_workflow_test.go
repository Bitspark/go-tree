// Package integration contains end-to-end tests that combine multiple features
// of the go-tree library in real-world scenarios.
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/saver"
	"bitspark.dev/go-tree/pkgold/transform/extract"
	"bitspark.dev/go-tree/pkgold/transform/rename"
)

// TestRefactoringWorkflow demonstrates a complete refactoring workflow:
// 1. Load a Go module
// 2. Analyze its structure
// 3. Perform code transformations (renaming and interface extraction)
// 4. Save the transformed module
func TestRefactoringWorkflow(t *testing.T) {
	// Setup test directories
	testDir := filepath.Join("testdata", "refactoring")
	outDir := filepath.Join(testDir, "output")

	// Ensure output directory exists
	if err := os.MkdirAll(outDir, 0750); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer os.RemoveAll(outDir) // Clean up after test

	// Step 1: Load the module
	modLoader := loader.NewGoModuleLoader()
	mod, err := modLoader.Load(testDir)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Step 2: Analyze the module structure
	if len(mod.Packages) == 0 {
		t.Fatal("Expected at least one package in the module")
	}

	// Find a package to transform (main package or the first non-test package)
	var pkg *module.Package
	if mod.MainPackage != nil {
		pkg = mod.MainPackage
	} else {
		// Get the first package
		for _, p := range mod.Packages {
			// Check if the package is a test package (ends with _test)
			if !(len(p.ImportPath) > 5 && p.ImportPath[len(p.ImportPath)-5:] == "_test") {
				pkg = p
				break
			}
		}
	}

	if pkg == nil {
		t.Fatal("Could not find a suitable package to transform")
	}

	// Step 3: Apply transformations
	// 3.1 Rename a type (if exists)
	renamer := rename.NewTypeRenamer(pkg.ImportPath, "", "", false)
	if len(pkg.Types) > 0 {
		// Get the first type
		var typeName string
		for name := range pkg.Types {
			typeName = name
			break
		}

		// Rename the type
		newName := typeName + "Refactored"
		renamer = rename.NewTypeRenamer(pkg.ImportPath, typeName, newName, false)
		result := renamer.Transform(mod)
		if !result.Success {
			t.Fatalf("Failed to rename type: %v", result.Error)
		}

		// Verify the type was renamed
		if _, exists := pkg.Types[newName]; !exists {
			t.Errorf("Expected renamed type %s to exist", newName)
		}
		if _, exists := pkg.Types[typeName]; exists {
			t.Errorf("Original type %s should not exist after renaming", typeName)
		}
	}

	// 3.2 Extract an interface (if methods exist)
	options := extract.DefaultOptions()
	extractor := extract.NewInterfaceExtractor(options)

	// Find a type with methods to extract an interface from
	var methodReceiverType string
	for _, fn := range pkg.Functions {
		if fn.IsMethod && fn.Receiver != nil && fn.Receiver.Type != "" {
			methodReceiverType = fn.Receiver.Type
			break
		}
	}

	if methodReceiverType != "" {
		// Extract interface from the type's methods
		interfaceName := "I" + methodReceiverType

		// Create a custom extractor that just extracts from one type
		options.MinimumTypes = 1 // Allow extraction from a single type
		options.TargetPackage = pkg.ImportPath
		extractor = extract.NewInterfaceExtractor(options)

		// Apply the transformation
		err := extractor.Transform(mod)
		if err != nil {
			t.Fatalf("Failed to extract interface: %v", err)
		}

		// Verify interface was created
		if _, exists := pkg.Types[interfaceName]; !exists {
			t.Errorf("Expected extracted interface %s to exist", interfaceName)
		}
	}

	// Step 4: Save the transformed module
	modSaver := saver.NewGoModuleSaver()
	err = modSaver.SaveTo(mod, outDir)
	if err != nil {
		t.Fatalf("Failed to save transformed module: %v", err)
	}

	// Verify output files exist
	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected output files after saving module")
	}
}
