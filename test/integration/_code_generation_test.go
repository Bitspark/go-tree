package integration

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/saver"
	"bitspark.dev/go-tree/pkgold/testing/generator"
)

// TestCodeGenerationWorkflow demonstrates a workflow for:
// 1. Loading a Go module
// 2. Analyzing its structure
// 3. Generating complementary code (tests, interface implementations)
// 4. Saving the extended module
func TestCodeGenerationWorkflow(t *testing.T) {
	// Setup test directories
	testDir := filepath.Join("testdata", "codegen")
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

	// Step 2: Analyze module for code generation opportunities
	// Find a package to extend with generated code
	var targetPkg *module.Package
	for _, pkg := range mod.Packages {
		if !isTestPackage(pkg.ImportPath) {
			targetPkg = pkg
			break
		}
	}

	if targetPkg == nil {
		t.Fatal("Could not find a suitable package for code generation")
	}

	// Step 3: Generate test code
	testGen := generator.NewGenerator()

	// Find functions without tests
	testFiles := make(map[string]string)
	for fnName, fn := range targetPkg.Functions {
		if fn.IsExported && !fn.IsMethod {
			// Generate a table-driven test for this function
			testCode, err := testGen.GenerateTestTemplate(fn, "table")
			if err != nil {
				t.Fatalf("Failed to generate test for %s: %v", fnName, err)
			}

			// Store the generated test
			testFileName := fnName + "_test.go"
			testFiles[testFileName] = testCode
		}
	}

	// Save generated tests
	testPkgDir := filepath.Join(outDir, targetPkg.Name+"_test")
	if err := os.MkdirAll(testPkgDir, 0750); err != nil {
		t.Fatalf("Failed to create test package directory: %v", err)
	}

	for fileName, fileContent := range testFiles {
		testFilePath := filepath.Join(testPkgDir, fileName)
		if err := os.WriteFile(testFilePath, []byte(fileContent), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", fileName, err)
		}
	}

	// Step 4: Generate interface implementations
	// Find an interface to implement
	var interfacePkg *module.Package
	var interfaceName string
	var interfaceType *module.Type

	for _, pkg := range mod.Packages {
		for typeName, typeObj := range pkg.Types {
			if typeObj.Kind == "interface" && typeObj.IsExported {
				interfacePkg = pkg
				interfaceName = typeName
				interfaceType = typeObj
				break
			}
		}
		if interfaceName != "" {
			break
		}
	}

	if interfaceName != "" {
		// Create a new package for the implementation
		implPkg := &module.Package{
			Name:       interfaceName + "Impl",
			ImportPath: mod.Path + "/impl/" + interfaceName,
			Functions:  make(map[string]*module.Function),
			Types:      make(map[string]*module.Type),
			Constants:  make(map[string]*module.Variable),
			Variables:  make(map[string]*module.Variable),
			Imports:    []*module.Import{},
		}

		// Add an import for the package containing the interface
		implPkg.Imports = append(implPkg.Imports, &module.Import{
			Path: interfacePkg.ImportPath,
			Name: interfacePkg.Name,
		})

		// Generate a struct that implements the interface
		implType := &module.Type{
			Name:       interfaceName + "Impl",
			Kind:       "struct",
			IsExported: true,
			Fields:     []*module.Field{},
		}
		implPkg.Types[implType.Name] = implType

		// Generate method implementations for each interface method
		// This is simplified; a real implementation would need to analyze
		// the interface methods more thoroughly
		methodPrefix := interfaceName + "Impl"

		// Add the package to the module
		mod.AddPackage(implPkg)

		// Save the enhanced module
		implDir := filepath.Join(outDir, "impl")
		modSaver := saver.NewGoModuleSaver()
		if err := modSaver.SaveTo(mod, implDir); err != nil {
			t.Fatalf("Failed to save implementation: %v", err)
		}
	}

	// Step 5: Verify outputs exist
	// Check that test files were generated
	testFiles, err = os.ReadDir(testPkgDir)
	if err != nil {
		t.Fatalf("Failed to read test package directory: %v", err)
	}

	if len(testFiles) == 0 {
		t.Error("Expected at least one test file to be generated")
	}

	// Check for struct implementation if an interface was found
	if interfaceName != "" {
		implDir := filepath.Join(outDir, "impl", interfaceName)
		_, err = os.Stat(implDir)
		if os.IsNotExist(err) {
			t.Errorf("Expected implementation directory %s was not created", implDir)
		}
	}
}

// Helper function to determine if a package is a test package
func isTestPackage(importPath string) bool {
	return len(importPath) > 5 && importPath[len(importPath)-5:] == "_test"
}
