package materialize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironment_Execute(t *testing.T) {
	// Create a temporary directory for the environment
	tempDir, err := os.MkdirTemp("", "environment-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an environment
	env := NewEnvironment(tempDir, true)

	// Add a module path
	moduleDir := filepath.Join(tempDir, "mymodule")
	if err := os.Mkdir(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}
	env.ModulePaths["example.com/mymodule"] = moduleDir

	// Test executing a command in the environment
	cmd, err := env.Execute([]string{"pwd"}, "")
	if err != nil {
		t.Fatalf("Failed to create command: %v", err)
	}

	// The command should be targeting the root directory
	if cmd.Dir != tempDir {
		t.Errorf("Expected command directory to be %s, got %s", tempDir, cmd.Dir)
	}

	// Test executing a command in a module
	cmd, err = env.Execute([]string{"ls"}, "example.com/mymodule")
	if err != nil {
		t.Fatalf("Failed to create command in module: %v", err)
	}

	// The command should be targeting the module directory
	if cmd.Dir != moduleDir {
		t.Errorf("Expected command directory to be %s, got %s", moduleDir, cmd.Dir)
	}

	// Test invalid command
	_, err = env.Execute([]string{}, "")
	if err == nil {
		t.Errorf("Expected error for empty command, got nil")
	}
}

func TestEnvironment_EnvironmentVariables(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "env-vars-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an environment
	env := NewEnvironment(tempDir, true)

	// Set environment variables
	env.SetEnvVar("TEST_VAR1", "value1")
	env.SetEnvVar("TEST_VAR2", "value2")

	// Check getting environment variables
	value, ok := env.GetEnvVar("TEST_VAR1")
	if !ok {
		t.Errorf("Expected to find TEST_VAR1 but it's missing")
	}
	if value != "value1" {
		t.Errorf("Expected TEST_VAR1 to be 'value1', got '%s'", value)
	}

	// Check for non-existent variable
	_, ok = env.GetEnvVar("NONEXISTENT")
	if ok {
		t.Errorf("Expected NONEXISTENT to be missing, but it was found")
	}

	// Test clearing environment variables
	env.ClearEnvVars()
	_, ok = env.GetEnvVar("TEST_VAR1")
	if ok {
		t.Errorf("Expected TEST_VAR1 to be cleared, but it still exists")
	}
}

func TestEnvironment_Cleanup(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a file to check removal
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create environment with isTemporary=true
	env := NewEnvironment(tempDir, true)

	// Cleanup should remove the directory
	if err := env.Cleanup(); err != nil {
		t.Errorf("Failed to cleanup: %v", err)
	}

	// Check that the directory is gone
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("Temporary directory still exists after cleanup")
		// If the test fails, cleanup manually to avoid leaving temp files
		os.RemoveAll(tempDir)
	}

	// Create a non-temporary environment
	permanentDir, err := os.MkdirTemp("", "permanent-test-*")
	if err != nil {
		t.Fatalf("Failed to create permanent dir: %v", err)
	}
	defer os.RemoveAll(permanentDir)

	permanentEnv := NewEnvironment(permanentDir, false)

	// Cleanup should not remove the directory
	if err := permanentEnv.Cleanup(); err != nil {
		t.Errorf("Failed during cleanup of permanent environment: %v", err)
	}

	// Check that the directory still exists
	if _, err := os.Stat(permanentDir); os.IsNotExist(err) {
		t.Errorf("Permanent directory was removed during cleanup")
	}
}

func TestEnvironment_FileExists(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "file-exists-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an environment
	env := NewEnvironment(tempDir, true)

	// Create a module directory
	moduleDir := filepath.Join(tempDir, "mymodule")
	if err := os.Mkdir(moduleDir, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}
	env.ModulePaths["example.com/mymodule"] = moduleDir

	// Create a test file
	testFile := filepath.Join(moduleDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !env.FileExists("example.com/mymodule", "testfile.txt") {
		t.Errorf("Expected testfile.txt to exist, but it was not found")
	}

	// Test non-existent file
	if env.FileExists("example.com/mymodule", "nonexistent.txt") {
		t.Errorf("Expected nonexistent.txt to be missing, but it was found")
	}

	// Test non-existent module
	if env.FileExists("example.com/nonexistent", "testfile.txt") {
		t.Errorf("Expected file in non-existent module to be missing, but it was found")
	}
}

func TestEnvironment_AllModulePaths(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "module-paths-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an environment
	env := NewEnvironment(tempDir, true)

	// Add some module paths
	env.ModulePaths["example.com/module1"] = filepath.Join(tempDir, "module1")
	env.ModulePaths["example.com/module2"] = filepath.Join(tempDir, "module2")
	env.ModulePaths["example.com/module3"] = filepath.Join(tempDir, "module3")

	// Get all module paths
	paths := env.AllModulePaths()

	// Check that we have the expected number of paths
	if len(paths) != 3 {
		t.Errorf("Expected 3 module paths, got %d", len(paths))
	}

	// Check that all modules are included
	expectedModules := map[string]bool{
		"example.com/module1": true,
		"example.com/module2": true,
		"example.com/module3": true,
	}

	for _, path := range paths {
		if !expectedModules[path] {
			t.Errorf("Unexpected module path: %s", path)
		}
		delete(expectedModules, path)
	}

	if len(expectedModules) > 0 {
		t.Errorf("Missing module paths: %v", expectedModules)
	}
}
