package saver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

func TestGoModuleSaver_Save(t *testing.T) {
	// Create a programmatic test module instead of loading from testdata
	mod := module.NewModule("testmodule", "/test")
	mod.GoVersion = "1.18"

	// Create a simple package
	pkg := module.NewPackage("samplepackage", "testmodule/samplepackage", "/test/samplepackage")
	mod.AddPackage(pkg)

	// Create a simple Go file with valid Go code
	file := module.NewFile("/test/samplepackage/sample.go", "sample.go", false)
	file.SourceCode = `package samplepackage

import (
	"fmt"
)

// SampleType is a test struct
type SampleType struct {
	Name string
	ID   int
}

// SampleFunc is a test function
func SampleFunc() {
	fmt.Println("Sample function")
}
`
	pkg.AddFile(file)

	// Add a type
	sampleType := module.NewType("SampleType", "struct", true)
	sampleType.Doc = "SampleType is a test struct"
	sampleType.AddField("Name", "string", "", false, "")
	sampleType.AddField("ID", "int", "", false, "")
	file.AddType(sampleType)
	pkg.AddType(sampleType)

	// Add a function
	sampleFunc := module.NewFunction("SampleFunc", true, false)
	sampleFunc.Doc = "SampleFunc is a test function"
	sampleFunc.Signature = "()"
	file.AddFunction(sampleFunc)
	pkg.AddFunction(sampleFunc)

	// Create a temp directory for saving
	tempDir, err := os.MkdirTemp("", "gosaver-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create saver
	saver := NewGoModuleSaver()

	// Test Save with default options (it should use module.Dir)
	mod.Dir = tempDir // Set this to make Save() work
	err = saver.Save(mod)
	if err != nil {
		t.Fatalf("Failed to save module with Save(): %v", err)
	}

	// Test SaveTo with default options to a new directory
	newTempDir, err := os.MkdirTemp("", "gosaver-saveto-test-*")
	if err != nil {
		t.Fatalf("Failed to create second temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(newTempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	err = saver.SaveTo(mod, newTempDir)
	if err != nil {
		t.Fatalf("Failed to save module with SaveTo(): %v", err)
	}

	// Verify files were saved in the second directory
	goModPath := filepath.Join(newTempDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod file was not created in %s", newTempDir)
	}

	// Check if package directory was created
	samplePkgDir := filepath.Join(newTempDir, "samplepackage")
	if _, err := os.Stat(samplePkgDir); os.IsNotExist(err) {
		t.Errorf("Sample package directory was not created at %s", samplePkgDir)
	}

	// Check if Go file was created
	sampleFile := filepath.Join(samplePkgDir, "sample.go")
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Errorf("sample.go file was not created at %s", sampleFile)
	}

	// Read the content of the go.mod file to verify it's correct
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module testmodule") {
		t.Errorf("go.mod does not contain expected module declaration, got: %s", content)
	}

	// Read the saved Go file to verify it contains the expected content
	fileContent, err := os.ReadFile(sampleFile)
	if err != nil {
		t.Fatalf("Failed to read sample.go: %v", err)
	}

	// Check that the file contains expected elements
	sampleFileStr := string(fileContent)
	if !strings.Contains(sampleFileStr, "package samplepackage") {
		t.Errorf("sample.go does not contain package declaration")
	}
	if !strings.Contains(sampleFileStr, "type SampleType struct") {
		t.Errorf("sample.go does not contain SampleType struct")
	}
	if !strings.Contains(sampleFileStr, "func SampleFunc") {
		t.Errorf("sample.go does not contain SampleFunc function")
	}
}

func TestGoModuleSaver_SaveWithOptions(t *testing.T) {
	// Create a simple module programmatically instead of loading from testdata
	mod := module.NewModule("testmodule", "/test")
	mod.GoVersion = "1.18"

	// Create a package
	pkg := module.NewPackage("testpkg", "testmodule/testpkg", "/test/testpkg")
	mod.AddPackage(pkg)

	// Create a file
	file := module.NewFile("/test/testpkg/main.go", "main.go", false)
	pkg.AddFile(file)

	// Add a simple type
	typ := module.NewType("TestType", "struct", true)
	file.AddType(typ)
	pkg.AddType(typ)

	// Add a field to the type
	typ.AddField("Name", "string", "", false, "")

	// Create a temp directory for saving
	tempDir, err := os.MkdirTemp("", "gosaver-options-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create saver
	saver := NewGoModuleSaver()

	// Create custom options
	options := SaveOptions{
		Format:          true,
		OrganizeImports: true,
		CreateBackups:   true,
	}

	// Test SaveToWithOptions
	err = saver.SaveToWithOptions(mod, tempDir, options)
	if err != nil {
		t.Fatalf("Failed to save module with options: %v", err)
	}

	// Verify files were saved
	goModPath := filepath.Join(tempDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod file was not created with custom options")
	}

	// Check if package directory was created
	pkgDir := filepath.Join(tempDir, "testpkg")
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		t.Errorf("Package directory was not created")
	}

	// Check if file was created
	filePathInTempDir := filepath.Join(pkgDir, "main.go")
	if _, err := os.Stat(filePathInTempDir); os.IsNotExist(err) {
		t.Errorf("File main.go was not created")
	}
}

func TestDefaultSaveOptions(t *testing.T) {
	options := DefaultSaveOptions()

	if !options.Format {
		t.Errorf("Expected Format to be true in default options")
	}

	if !options.OrganizeImports {
		t.Errorf("Expected OrganizeImports to be true in default options")
	}

	if options.CreateBackups {
		t.Errorf("Expected CreateBackups to be false in default options")
	}
}

func TestSaveWithModifiedModule(t *testing.T) {
	// Create a module programmatically
	mod := module.NewModule("testmodule", "/test")
	mod.GoVersion = "1.18"

	// Add a dependency
	mod.Dependencies = append(mod.Dependencies, &module.ModuleDependency{
		Path:    "github.com/example/testdep",
		Version: "v1.0.0",
	})

	// Add a replacement
	mod.Replace = append(mod.Replace, &module.ModuleReplace{
		Old: &module.ModuleDependency{Path: "github.com/example/testdep"},
		New: &module.ModuleDependency{Path: "../testdep", Version: ""},
	})

	// Create a temp directory for saving
	tempDir, err := os.MkdirTemp("", "gosaver-modified-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create saver
	saver := NewGoModuleSaver()

	// Save the modified module
	err = saver.SaveTo(mod, tempDir)
	if err != nil {
		t.Fatalf("Failed to save modified module: %v", err)
	}

	// Read the content of the go.mod file to verify modifications
	content, err := os.ReadFile(filepath.Join(tempDir, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Verify dependency was added
	if !strings.Contains(string(content), "github.com/example/testdep v1.0.0") {
		t.Errorf("go.mod does not contain added dependency")
	}

	// Verify replacement was added
	if !strings.Contains(string(content), "github.com/example/testdep => ../testdep") {
		t.Errorf("go.mod does not contain added replacement")
	}
}
