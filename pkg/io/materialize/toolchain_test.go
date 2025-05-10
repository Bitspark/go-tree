package materialize

import (
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	toolkitTesting "bitspark.dev/go-tree/pkg/toolkit/testing"
)

// TestMaterializeWithCustomToolchain tests materialization with a custom toolchain and filesystem
func TestMaterializeWithCustomToolchain(t *testing.T) {
	// Create mock toolchain that logs operations
	mockToolchain := toolkitTesting.NewMockGoToolchain()

	// Configure mock for finding modules
	mockToolchain.CommandResults["find-module github.com/test/simplemath"] = toolkitTesting.MockCommandResult{
		Output: []byte("/mock/path/to/simplemath"),
	}

	// Create mock filesystem
	mockFS := toolkitTesting.NewMockModuleFS()

	// Add mock files for the simplemath module
	mockFS.AddFile("/mock/path/to/simplemath/go.mod", []byte(`module github.com/test/simplemath

go 1.19`))
	mockFS.AddFile("/mock/path/to/simplemath/math.go", []byte(`package simplemath

// Add returns the sum of two integers
func Add(a, b int) int { 
	return a + b 
}`))

	// Create a module to materialize
	module := &typesys.Module{
		Path:      "github.com/test/simplemath",
		Dir:       "/mock/path/to/simplemath",
		GoVersion: "1.19",
		Packages:  make(map[string]*typesys.Package),
	}

	// Create test package within the module
	pkg := typesys.NewPackage(module, "simplemath", "github.com/test/simplemath")
	module.Packages[pkg.ImportPath] = pkg

	// Add file to the package
	file := &typesys.File{
		Path:    "/mock/path/to/simplemath/math.go",
		Name:    "math.go",
		Package: pkg,
	}
	pkg.Files = map[string]*typesys.File{file.Path: file}

	// Create materializer with mocks
	materializer := NewModuleMaterializer().
		WithToolchain(mockToolchain).
		WithFS(mockFS)

	// Materialize the module
	opts := DefaultMaterializeOptions()
	env, err := materializer.Materialize(module, opts)
	if err != nil {
		t.Fatalf("Failed to materialize module: %v", err)
	}

	// Verify the module was materialized in the environment
	modulePath, ok := env.ModulePaths[module.Path]
	if !ok {
		t.Fatalf("Module path not found in environment")
	}

	// Verify the mock filesystem was used
	if len(mockFS.Operations) == 0 {
		t.Errorf("No filesystem operations recorded")
	}

	// Verify mock toolchain was used
	if len(mockToolchain.Invocations) == 0 {
		t.Errorf("No toolchain operations recorded")
	}

	// Verify files were written to the mock filesystem
	goModPath := filepath.Join(modulePath, "go.mod")
	if !mockFS.FileExists(goModPath) {
		t.Errorf("go.mod not found at %s", goModPath)
	}

	mathGoPath := filepath.Join(modulePath, "math.go")
	if !mockFS.FileExists(mathGoPath) {
		t.Errorf("math.go not found at %s", mathGoPath)
	}

	// Verify the file content was written correctly
	goModContent, err := mockFS.ReadFile(goModPath)
	if err != nil {
		t.Errorf("Failed to read go.mod: %v", err)
	}
	if !contains(string(goModContent), "module github.com/test/simplemath") {
		t.Errorf("go.mod doesn't contain module declaration: %s", string(goModContent))
	}
}

// TestMaterializeWithErrorHandling tests error handling during materialization
func TestMaterializeWithErrorHandling(t *testing.T) {
	// Create mock filesystem that will return errors
	mockFS := toolkitTesting.NewMockModuleFS()

	// Configure mock to return error for WriteFile operations
	mockFS.Errors["WriteFile:/some/path/go.mod"] = &materialPlaceholderError{msg: "write error"}

	// Create a simple module
	module := &typesys.Module{
		Path:      "example.com/errortest",
		Dir:       "/some/path",
		GoVersion: "1.19",
		Packages:  make(map[string]*typesys.Package),
	}

	// Create materializer with mock
	materializer := NewModuleMaterializer().
		WithFS(mockFS)

	// Try a few different error scenarios

	// 1. Error during go.mod file creation
	opts := DefaultMaterializeOptions()
	opts.TargetDir = "/some/path"

	_, err := materializer.Materialize(module, opts)

	// This might or might not fail depending on the exact implementation
	// since we're only mocking one specific file path
	if err == nil {
		// Verify that at least some operations were attempted
		if len(mockFS.Operations) == 0 {
			t.Errorf("No filesystem operations recorded")
		}
	} else {
		// If it failed, it should be with our error
		if !contains(err.Error(), "write error") {
			t.Errorf("Expected 'write error' in error message, got: %v", err)
		}
	}

	// 2. Error due to target directory creation
	mockFS.Errors["MkdirAll:/error/path"] = &materialPlaceholderError{msg: "mkdir error"}

	opts = DefaultMaterializeOptions()
	opts.TargetDir = "/error/path"

	_, err = materializer.Materialize(module, opts)

	if err == nil {
		t.Errorf("Expected error for MkdirAll but got none")
	} else if !contains(err.Error(), "mkdir error") && !contains(err.Error(), "failed to create") {
		t.Errorf("Expected directory creation error, got: %v", err)
	}
}

// Helper type to simulate errors
type materialPlaceholderError struct {
	msg string
}

func (e *materialPlaceholderError) Error() string {
	return e.msg
}
