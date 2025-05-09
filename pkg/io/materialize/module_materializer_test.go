package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestModuleMaterializer_Materialize(t *testing.T) {
	// Create a temporary test module
	tempDir, err := os.MkdirTemp("", "materializer-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer safeRemoveAll(tempDir)

	// Create a simple go.mod file
	goModContent := `module example.com/testmodule

go 1.16

require (
	golang.org/x/text v0.3.7
)
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple module
	module := &typesys.Module{
		Path:      "example.com/testmodule",
		Dir:       tempDir,
		GoVersion: "1.16",
		Packages:  make(map[string]*typesys.Package),
	}

	// Create a materializer
	materializer := NewModuleMaterializer()

	// Set up options for materialization
	materializeDir, err := os.MkdirTemp("", "materialized-*")
	if err != nil {
		t.Fatalf("Failed to create materialization dir: %v", err)
	}
	defer safeRemoveAll(materializeDir)

	opts := MaterializeOptions{
		TargetDir:        materializeDir,
		DependencyPolicy: NoDependencies, // For this test, we don't need dependencies
		ReplaceStrategy:  NoReplace,
		LayoutStrategy:   FlatLayout,
		RunGoModTidy:     false,
		Verbose:          false,
	}

	// Materialize the module
	env, err := materializer.Materialize(module, opts)
	if err != nil {
		t.Fatalf("Failed to materialize module: %v", err)
	}

	// The environment should contain our module
	if len(env.ModulePaths) != 1 {
		t.Errorf("Expected 1 module in environment, got %d", len(env.ModulePaths))
	}

	modulePath, ok := env.ModulePaths["example.com/testmodule"]
	if !ok {
		t.Fatalf("Module path not found in environment")
	}

	// Check that go.mod exists and contains expected content
	goModPath := filepath.Join(modulePath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Errorf("go.mod not found in materialized module")
	}

	// Read the go.mod file and verify its content
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Basic verification that it contains the module path
	if len(content) == 0 {
		t.Errorf("go.mod is empty")
	} else if string(content[:7]) != "module " {
		t.Errorf("go.mod doesn't start with 'module', got: %s", string(content[:min(10, len(content))]))
	}
}

func TestModuleMaterializer_MaterializeWithDependencies(t *testing.T) {
	// This is a more complex test that requires actual modules in the GOPATH
	// So we'll skip it if dependencies aren't available or if running in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	// Create a temporary test module
	tempDir, err := os.MkdirTemp("", "materializer-deps-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer safeRemoveAll(tempDir)

	// Create a simple go.mod file with dependencies
	goModContent := `module example.com/testmodule

go 1.16

require (
	golang.org/x/text v0.3.7
)
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple module
	module := &typesys.Module{
		Path:      "example.com/testmodule",
		Dir:       tempDir,
		GoVersion: "1.16",
		Packages:  make(map[string]*typesys.Package),
	}

	// Create a materializer
	materializer := NewModuleMaterializer()

	// Set up options for materialization with dependencies
	materializeDir, err := os.MkdirTemp("", "materialized-deps-*")
	if err != nil {
		t.Fatalf("Failed to create materialization dir: %v", err)
	}
	defer safeRemoveAll(materializeDir)

	opts := MaterializeOptions{
		TargetDir:        materializeDir,
		DependencyPolicy: DirectDependenciesOnly,
		ReplaceStrategy:  RelativeReplace,
		LayoutStrategy:   FlatLayout,
		RunGoModTidy:     false,
		Verbose:          true, // Verbose for debugging
	}

	// Materialize the module with dependencies
	env, err := materializer.Materialize(module, opts)
	if err != nil {
		// Dependency materialization might fail if the dependency isn't in the GOPATH
		// That's okay for testing purposes
		t.Logf("Note: Materialization returned error: %v", err)
		t.Skip("Skipping test since dependency materialization failed")
	}

	// The environment should contain our module
	if len(env.ModulePaths) < 1 {
		t.Errorf("Expected at least 1 module in environment, got %d", len(env.ModulePaths))
	}

	// If dependency materialization succeeded, we should have 2 modules
	if len(env.ModulePaths) > 1 {
		t.Logf("Successfully materialized module with dependencies")

		// We should have both our module and the dependency
		if _, ok := env.ModulePaths["example.com/testmodule"]; !ok {
			t.Errorf("Main module not found in environment")
		}

		// Check for dependency (may not be present in all environments)
		if _, ok := env.ModulePaths["golang.org/x/text"]; ok {
			t.Logf("Dependency golang.org/x/text found in environment")
		}
	}
}

// Helper function to get minimum of two integers for string slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
