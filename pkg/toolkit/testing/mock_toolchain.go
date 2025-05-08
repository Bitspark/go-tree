// Package testing provides mock implementations for testing
package testing

import (
	"context"
	"fmt"
	"strings"
)

// MockCommandResult holds the response for a mocked command
type MockCommandResult struct {
	Output []byte
	Err    error
}

// MockInvocation records information about a command invocation
type MockInvocation struct {
	Command string
	Args    []string
}

// MockGoToolchain implements toolkit.GoToolchain for testing
type MockGoToolchain struct {
	// Mock responses for different commands
	CommandResults map[string]MockCommandResult

	// Track command invocations
	Invocations []MockInvocation
}

// NewMockGoToolchain creates a new mock toolchain
func NewMockGoToolchain() *MockGoToolchain {
	return &MockGoToolchain{
		CommandResults: make(map[string]MockCommandResult),
		Invocations:    make([]MockInvocation, 0),
	}
}

// RunCommand executes a Go command with arguments
func (t *MockGoToolchain) RunCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	// Record the invocation
	t.Invocations = append(t.Invocations, MockInvocation{
		Command: command,
		Args:    args,
	})

	// Build the command key
	cmdKey := command
	if len(args) > 0 {
		cmdKey += " " + strings.Join(args, " ")
	}

	// Look for an exact match
	if result, ok := t.CommandResults[cmdKey]; ok {
		return result.Output, result.Err
	}

	// Look for a prefix match
	for k, result := range t.CommandResults {
		if strings.HasPrefix(cmdKey, k) {
			return result.Output, result.Err
		}
	}

	return nil, fmt.Errorf("no mock response found for command: %s", cmdKey)
}

// GetModuleInfo retrieves information about a module
func (t *MockGoToolchain) GetModuleInfo(ctx context.Context, importPath string) (path string, version string, err error) {
	output, err := t.RunCommand(ctx, "list", "-m", importPath)
	if err != nil {
		return "", "", err
	}

	// Parse output (format: "path version")
	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected output format from mock go list -m: %s", output)
	}

	return parts[0], parts[1], nil
}

// DownloadModule downloads a module
func (t *MockGoToolchain) DownloadModule(ctx context.Context, importPath string, version string) error {
	versionSpec := importPath
	if version != "" {
		versionSpec += "@" + version
	}

	_, err := t.RunCommand(ctx, "get", "-d", versionSpec)
	return err
}

// FindModule locates a module in the module cache
func (t *MockGoToolchain) FindModule(ctx context.Context, importPath string, version string) (string, error) {
	cmdKey := fmt.Sprintf("find-module %s %s", importPath, version)

	// Check if we have a mock for this specific query
	if result, ok := t.CommandResults[cmdKey]; ok {
		if result.Err != nil {
			return "", result.Err
		}
		return string(result.Output), nil
	}

	// If there's no specific mock, just invent a path
	return fmt.Sprintf("/mock/path/to/%s@%s", importPath, version), nil
}

// CheckModuleExists verifies a module exists and is accessible
func (t *MockGoToolchain) CheckModuleExists(ctx context.Context, importPath string, version string) (bool, error) {
	cmdKey := fmt.Sprintf("check-module %s %s", importPath, version)

	// Check if we have a mock for this specific query
	if result, ok := t.CommandResults[cmdKey]; ok {
		if result.Err != nil {
			return false, result.Err
		}
		return string(result.Output) == "true", nil
	}

	// Look for a FindModule mock
	_, err := t.FindModule(ctx, importPath, version)
	return err == nil, nil
}
