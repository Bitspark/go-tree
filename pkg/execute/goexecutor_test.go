package execute

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

func TestGoExecutor_Execute(t *testing.T) {
	// Create a test module
	mod := &module.Module{
		Path:      "example.com/testmodule",
		GoVersion: "1.18",
		Dir:       os.TempDir(), // Use temp dir for the test
	}

	// Create executor
	executor := NewGoExecutor()

	// Test a simple version command
	result, err := executor.Execute(mod, "version")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check if version command worked
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.StdOut == "" {
		t.Error("Expected stdout to contain Go version info, got empty string")
	}
}

func TestGoExecutor_ExecuteTest(t *testing.T) {
	// Skip this test if running in CI environment without a complete Go environment
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment")
	}

	// Create a temporary go module for testing
	testDir, err := createTestModule(t)
	if err != nil {
		t.Fatalf("Failed to create test module: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Warning: failed to remove test directory %s: %v", testDir, err)
		}
	}()

	// Create module representation
	mod := &module.Module{
		Path:      "example.com/testmod",
		GoVersion: "1.18",
		Dir:       testDir,
	}

	// Create executor
	executor := NewGoExecutor()

	// Run tests on the module
	result, err := executor.ExecuteTest(mod, "./...", "-v")

	// We're expecting the test to pass
	if err != nil {
		t.Fatalf("ExecuteTest failed: %v", err)
	}

	// Check test results
	if result.Failed > 0 {
		t.Errorf("Expected 0 failed tests, got %d", result.Failed)
	}

	if len(result.Tests) == 0 {
		t.Error("Expected to find at least one test, got none")
	}
}

// Helper function to create a temporary Go module with a simple test
func createTestModule(t *testing.T) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "goexecutor-test-*")
	if err != nil {
		return "", err
	}

	// Initialize Go module
	initCmd := exec.Command("go", "mod", "init", "example.com/testmod")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		if cleanErr := os.RemoveAll(tempDir); cleanErr != nil {
			return "", fmt.Errorf("failed to clean up temp dir: %v (after: %v)", cleanErr, err)
		}
		return "", err
	}

	// Create a simple Go file with a test
	mainFile := filepath.Join(tempDir, "main.go")
	mainContent := []byte(`package main

func main() {
	println("Hello, world!")
}

func Add(a, b int) int {
	return a + b
}
`)
	if err := os.WriteFile(mainFile, mainContent, 0644); err != nil {
		if cleanErr := os.RemoveAll(tempDir); cleanErr != nil {
			return "", fmt.Errorf("failed to clean up temp dir: %v (after: %v)", cleanErr, err)
		}
		return "", err
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "main_test.go")
	testContent := []byte(`package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Errorf("Expected Add(2, 3) to be 5")
	}
}
`)
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		if cleanErr := os.RemoveAll(tempDir); cleanErr != nil {
			return "", fmt.Errorf("failed to clean up temp dir: %v (after: %v)", cleanErr, err)
		}
		return "", err
	}

	return tempDir, nil
}
