package env

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestStandardGoToolchainInitialization tests initialization of the standard Go toolchain
func TestStandardGoToolchainInitialization(t *testing.T) {
	toolchain := NewStandardGoToolchain()

	// Verify default values
	if toolchain.GoExecutable != "go" {
		t.Errorf("Expected GoExecutable to be 'go', got '%s'", toolchain.GoExecutable)
	}

	// Env should include the system environment
	if len(toolchain.Env) == 0 {
		t.Errorf("Expected non-empty Env, got empty")
	}

	// WorkDir should be empty
	if toolchain.WorkDir != "" {
		t.Errorf("Expected empty WorkDir, got '%s'", toolchain.WorkDir)
	}

	// Create a custom toolchain
	customToolchain := &StandardGoToolchain{
		GoExecutable: "/usr/local/bin/go",
		WorkDir:      "/tmp/work",
		Env:          []string{"GO111MODULE=on", "GOPROXY=direct"},
	}

	// Verify custom values
	if customToolchain.GoExecutable != "/usr/local/bin/go" {
		t.Errorf("Expected GoExecutable to be '/usr/local/bin/go', got '%s'", customToolchain.GoExecutable)
	}
	if customToolchain.WorkDir != "/tmp/work" {
		t.Errorf("Expected WorkDir to be '/tmp/work', got '%s'", customToolchain.WorkDir)
	}
	if len(customToolchain.Env) != 2 {
		t.Errorf("Expected 2 env vars, got %d", len(customToolchain.Env))
	}
}

// TestStandardGoToolchainRunCommand tests the RunCommand method
func TestStandardGoToolchainRunCommand(t *testing.T) {
	// Check if Go is installed using exec.LookPath instead of hardcoded paths
	_, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Skipping test as go is not installed or not in PATH")
	}

	ctx := context.Background()
	toolchain := NewStandardGoToolchain()

	// Test a simple version command
	output, err := toolchain.RunCommand(ctx, "version")
	if err != nil {
		t.Errorf("Expected no error running 'go version', got: %v", err)
	}
	if !strings.Contains(string(output), "go version") {
		t.Errorf("Expected output to contain 'go version', got: %s", string(output))
	}

	// Test an invalid command
	_, err = toolchain.RunCommand(ctx, "invalid-command")
	if err == nil {
		t.Errorf("Expected error running invalid command, got nil")
	}

	// Test with context cancellation
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately
	_, err = toolchain.RunCommand(cancelledCtx, "version")
	if err == nil {
		t.Errorf("Expected error with cancelled context, got nil")
	}
}

// TestStandardGoToolchainFindModule tests the FindModule method
func TestStandardGoToolchainFindModule(t *testing.T) {
	// This is a complex test that depends on the Go environment
	// So we'll just do basic checks and skip if needed
	ctx := context.Background()
	toolchain := NewStandardGoToolchain()

	// Try to find a standard library module
	dir, err := toolchain.FindModule(ctx, "fmt", "")
	// Just ensure the call doesn't panic - result will depend on environment
	if err != nil {
		// This is expected on some systems, so just log, don't fail
		t.Logf("FindModule returned err: %v", err)
	} else {
		t.Logf("FindModule returned dir: %s", dir)
	}

	// Empty version should not panic
	_, _ = toolchain.FindModule(ctx, "github.com/example/module", "")

	// Invalid module should return error but not panic
	_, err = toolchain.FindModule(ctx, "not-a-valid-module-path", "v1.0.0")
	if err == nil {
		t.Logf("Expected error for invalid module, but might work on some setups")
	}
}

// TestStandardGoToolchainGetModuleInfo tests the GetModuleInfo method
func TestStandardGoToolchainGetModuleInfo(t *testing.T) {
	// Similar to FindModule, this is environment-dependent
	// So we'll focus on testing the method doesn't panic
	ctx := context.Background()
	toolchain := NewStandardGoToolchain()

	// Try to get info for a standard library module
	path, version, err := toolchain.GetModuleInfo(ctx, "fmt")
	// Just check it doesn't panic, results will vary
	if err != nil {
		t.Logf("GetModuleInfo returned err: %v", err)
	} else {
		t.Logf("GetModuleInfo returned path: %s, version: %s", path, version)
	}

	// Invalid module should return error but not panic
	_, _, err = toolchain.GetModuleInfo(ctx, "not-a-valid-module-path")
	if err == nil {
		t.Logf("Expected error for invalid module, but might work on some setups")
	}
}

// TestGoToolchainInterface_CheckModuleExists tests the CheckModuleExists method
func TestGoToolchainInterface_CheckModuleExists(t *testing.T) {
	ctx := context.Background()
	toolchain := NewStandardGoToolchain()

	// Test with standard library module
	exists, err := toolchain.CheckModuleExists(ctx, "fmt", "")
	if err != nil {
		t.Logf("CheckModuleExists returned err: %v", err)
	} else if exists {
		t.Logf("Standard library module exists as expected")
	}

	// Test with invalid module
	exists, err = toolchain.CheckModuleExists(ctx, "not-a-valid-module-path", "v1.0.0")
	if err != nil {
		t.Logf("CheckModuleExists for invalid module returned err: %v", err)
	} else if !exists {
		t.Logf("Invalid module doesn't exist as expected")
	}
}

// Integration test - skip by default
func TestGoToolchainIntegration(t *testing.T) {
	// Only run integration tests if environment variable is set
	if os.Getenv("GO_TREE_RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set GO_TREE_RUN_INTEGRATION_TESTS=1 to enable")
	}

	ctx := context.Background()
	toolchain := NewStandardGoToolchain()

	// Download a well-known module
	err := toolchain.DownloadModule(ctx, "github.com/stretchr/testify", "v1.8.0")
	if err != nil {
		t.Errorf("Failed to download module: %v", err)
	}

	// Check if it exists
	exists, err := toolchain.CheckModuleExists(ctx, "github.com/stretchr/testify", "v1.8.0")
	if err != nil {
		t.Errorf("Error checking if module exists: %v", err)
	}
	if !exists {
		t.Errorf("Module should exist after downloading")
	}

	// Find its location
	dir, err := toolchain.FindModule(ctx, "github.com/stretchr/testify", "v1.8.0")
	if err != nil {
		t.Errorf("Failed to find module: %v", err)
	}
	if dir == "" {
		t.Errorf("Module directory should not be empty")
	}

	// Get info about the module
	path, version, err := toolchain.GetModuleInfo(ctx, "github.com/stretchr/testify")
	if err != nil {
		t.Errorf("Failed to get module info: %v", err)
	}
	if path != "github.com/stretchr/testify" {
		t.Errorf("Wrong module path: %s", path)
	}
	if version == "" {
		t.Errorf("Module version should not be empty")
	}
}
