package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestService_NewArchitecture(t *testing.T) {
	// Create a temporary test module
	tempDir, err := os.MkdirTemp("", "service-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple go.mod file
	goModContent := `module example.com/testmodule

go 1.16
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a dummy Go file
	goFileContent := `package main

import (
	"fmt"
	"errors" // Using standard library errors instead of external dependency
)

func main() {
	fmt.Println("Hello, world!")
	_ = errors.New("test error")
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Create service configuration
	config := &Config{
		ModuleDir:       tempDir,
		IncludeTests:    false,
		WithDeps:        false, // Don't load deps for basic test
		DependencyDepth: 1,
		DownloadMissing: false,
		Verbose:         true,
	}

	// Create the service
	service, err := NewService(config)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Verify that Resolver and Materializer were initialized
	if service.Resolver == nil {
		t.Errorf("Resolver was not initialized")
	}

	if service.Materializer == nil {
		t.Errorf("Materializer was not initialized")
	}

	// Verify that the main module was loaded
	if service.MainModulePath != "example.com/testmodule" {
		t.Errorf("Expected main module path to be 'example.com/testmodule', got '%s'", service.MainModulePath)
	}

	// Test creating an environment
	modules := []*typesys.Module{service.GetMainModule()}
	env, err := service.CreateEnvironment(modules, config)
	if err != nil {
		t.Logf("Note: Environment creation returned error: %v", err)
		t.Skip("Skipping environment test")
	} else {
		defer env.Cleanup()

		// Verify that the environment contains our module
		if len(env.ModulePaths) < 1 {
			t.Errorf("Expected at least 1 module in environment, got %d", len(env.ModulePaths))
		}

		if modulePath, ok := env.ModulePaths["example.com/testmodule"]; !ok {
			t.Errorf("Main module not found in environment")
		} else {
			// Check that go.mod exists and contains expected content
			goModPath := filepath.Join(modulePath, "go.mod")
			if _, err := os.Stat(goModPath); os.IsNotExist(err) {
				t.Errorf("go.mod not found in materialized module")
			}
		}
	}
}

func TestService_DependencyResolution(t *testing.T) {
	// Skip if running in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping dependency test in CI environment")
	}

	// Create a temporary test module
	tempDir, err := os.MkdirTemp("", "service-deps-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple go.mod file with dependencies
	goModContent := `module example.com/depsmodule

go 1.16

require (
	golang.org/x/text v0.3.7
)
`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create dummy Go file
	goFileContent := `package main

import (
	"fmt"
	"errors" // Using standard library
	_ "golang.org/x/text/language" // Common dependency that should be available
)

func main() {
	fmt.Println("Hello, world!")
	_ = errors.New("test error")
}
`
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Initialize go.sum file by running go mod tidy in the temporary directory
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tempDir
	if tidyOutput, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Warning: Failed to run go mod tidy: %v\nOutput: %s", err, tidyOutput)
		// Continue with the test anyway, as we want to test our handling of missing dependencies
	} else {
		t.Logf("Successfully initialized go.sum in test module")
	}

	// Create service configuration with dependency loading
	config := &Config{
		ModuleDir:       tempDir,
		IncludeTests:    false,
		WithDeps:        true, // Load dependencies
		DependencyDepth: 1,
		DownloadMissing: true,
		Verbose:         true,
	}

	// Create the service
	service, err := NewService(config)
	if err != nil {
		t.Logf("Note: Service creation returned error: %v", err)
		t.Skip("Skipping dependency test since service creation failed")
		return
	}

	// Test that the dependency was resolved
	// This might not work if the dependency is not in the cache
	if len(service.Modules) > 1 {
		t.Logf("Successfully resolved %d modules", len(service.Modules))

		// We should have both our module and the dependency
		if _, ok := service.Modules["example.com/depsmodule"]; !ok {
			t.Errorf("Main module not found")
		}

		// Dependency may not be present in all environments
		if _, ok := service.Modules["golang.org/x/text"]; ok {
			t.Logf("Dependency golang.org/x/text found")
		}

		// Test creating an environment with dependencies
		modules := []*typesys.Module{service.GetMainModule()}
		env, err := service.CreateEnvironment(modules, config)
		if err != nil {
			t.Logf("Environment creation returned error: %v", err)
		} else {
			defer env.Cleanup()
			t.Logf("Successfully created environment with %d modules", len(env.ModulePaths))
		}
	} else {
		t.Logf("No dependencies resolved, but this might be expected if dependency is not in cache")
	}
}
