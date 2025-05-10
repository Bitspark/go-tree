package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestModuleResolver_FindModuleLocation(t *testing.T) {
	// Create a new resolver with default options
	resolver := NewModuleResolver()

	// Test finding the standard library
	dir, err := resolver.FindModuleLocation("fmt", "")
	if err != nil {
		// It's okay if this fails in some environments, as the standard library may not be in the module cache
		t.Logf("Could not find standard library module: %v", err)
	} else {
		t.Logf("Found standard library at: %s", dir)
	}

	// Test finding a non-existent module
	_, err = resolver.FindModuleLocation("github.com/this/does/not/exist", "v1.0.0")
	if err == nil {
		t.Errorf("Expected error when looking for non-existent module, got nil")
	}
}

func TestModuleResolver_ParseGoMod(t *testing.T) {
	// Test parsing a simple go.mod file
	content := `module example.com/mymodule

go 1.16

require (
	golang.org/x/text v0.3.7
	golang.org/x/time v0.3.0
)

replace golang.org/x/text => golang.org/x/text v0.3.5
`

	deps, replacements, err := parseGoMod(content)
	if err != nil {
		t.Fatalf("Failed to parse go.mod: %v", err)
	}

	// Check dependencies
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	if v, ok := deps["golang.org/x/text"]; !ok || v != "v0.3.7" {
		t.Errorf("Expected golang.org/x/text@v0.3.7, got %s", v)
	}

	if v, ok := deps["golang.org/x/time"]; !ok || v != "v0.3.0" {
		t.Errorf("Expected golang.org/x/time@v0.3.0, got %s", v)
	}

	// Check replacements
	if len(replacements) != 1 {
		t.Errorf("Expected 1 replacement, got %d", len(replacements))
	}

	if v, ok := replacements["golang.org/x/text"]; !ok || v != "golang.org/x/text" {
		t.Errorf("Expected replacement golang.org/x/text => golang.org/x/text, got %s", v)
	}
}

func TestModuleResolver_BuildDependencyGraph(t *testing.T) {
	// Create a temporary test module
	tempDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Ignore error during cleanup
	}()

	// Create a simple go.mod file
	goModContent := `module example.com/testmodule

go 1.16

require (
	golang.org/x/text v0.3.7
	golang.org/x/time v0.3.0
)
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a simple module
	module := &typesys.Module{
		Path: "example.com/testmodule",
		Dir:  tempDir,
	}

	// Create a resolver
	resolver := NewModuleResolver()

	// Build the dependency graph
	graph, err := resolver.BuildDependencyGraph(module)
	if err != nil {
		t.Fatalf("Failed to build dependency graph: %v", err)
	}

	// Check the graph
	if len(graph) != 1 {
		t.Errorf("Expected 1 entry in graph, got %d", len(graph))
	}

	deps, ok := graph["example.com/testmodule"]
	if !ok {
		t.Fatalf("Module not found in graph")
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Check the dependencies
	expectedDeps := map[string]bool{
		"golang.org/x/text": true,
		"golang.org/x/time": true,
	}

	for _, dep := range deps {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
		delete(expectedDeps, dep)
	}

	if len(expectedDeps) > 0 {
		t.Errorf("Missing dependencies: %v", expectedDeps)
	}
}
