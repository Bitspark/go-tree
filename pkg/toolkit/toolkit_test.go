package toolkit

import (
	"context"
	"errors"
	"testing"

	toolkittesting "bitspark.dev/go-tree/pkg/toolkit/testing"
	"bitspark.dev/go-tree/pkg/typesys"
)

func TestStandardGoToolchain(t *testing.T) {
	toolchain := NewStandardGoToolchain()

	// Just verify it doesn't panic
	if toolchain.GoExecutable != "go" {
		t.Errorf("Expected GoExecutable to be 'go', got '%s'", toolchain.GoExecutable)
	}
}

func TestMockGoToolchain(t *testing.T) {
	mock := toolkittesting.NewMockGoToolchain()

	// Set up a mock response
	mock.CommandResults["list -m test/module"] = toolkittesting.MockCommandResult{
		Output: []byte("test/module v1.0.0"),
		Err:    nil,
	}

	// Test GetModuleInfo
	path, version, err := mock.GetModuleInfo(context.Background(), "test/module")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if path != "test/module" {
		t.Errorf("Expected path 'test/module', got '%s'", path)
	}

	if version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", version)
	}

	// Test error condition
	mock.CommandResults["list -m error/module"] = toolkittesting.MockCommandResult{
		Output: nil,
		Err:    errors.New("mock error"),
	}

	_, _, err = mock.GetModuleInfo(context.Background(), "error/module")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Verify invocations
	if len(mock.Invocations) != 2 {
		t.Errorf("Expected 2 invocations, got %d", len(mock.Invocations))
	}
}

func TestStandardModuleFS(t *testing.T) {
	fs := NewStandardModuleFS()

	// Just verify it doesn't panic
	_ = fs
}

func TestMockModuleFS(t *testing.T) {
	mock := toolkittesting.NewMockModuleFS()

	// Set up some mock files and directories
	mock.Files["/test/file.txt"] = []byte("test content")
	mock.Directories["/test"] = true

	// Test ReadFile
	content, err := mock.ReadFile("/test/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got '%s'", string(content))
	}

	// Test error condition
	mock.Errors["ReadFile:/error/file.txt"] = errors.New("mock read error")

	_, err = mock.ReadFile("/error/file.txt")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Test Stat on directory
	info, err := mock.Stat("/test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("Expected directory, got file")
	}

	// Test Stat on file
	info, err = mock.Stat("/test/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if info.IsDir() {
		t.Errorf("Expected file, got directory")
	}

	if info.Size() != 12 { // "test content" is 12 bytes
		t.Errorf("Expected size 12, got %d", info.Size())
	}

	// Verify operations were tracked
	if len(mock.Operations) != 4 {
		t.Errorf("Expected 4 operations, got %d", len(mock.Operations))
	}

	if mock.Operations[0] != "ReadFile:/test/file.txt" {
		t.Errorf("Expected 'ReadFile:/test/file.txt', got '%s'", mock.Operations[0])
	}
}

func TestMiddlewareChain(t *testing.T) {
	chain := NewMiddlewareChain()

	// Create some test middleware
	callOrder := []string{}

	middleware1 := func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
		callOrder = append(callOrder, "middleware1")
		module, err := next()
		return ctx, module, err
	}

	middleware2 := func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
		callOrder = append(callOrder, "middleware2")
		module, err := next()
		return ctx, module, err
	}

	// Add middleware to the chain
	chain.Add(middleware1, middleware2)

	// Create a final function
	final := func() (*typesys.Module, error) {
		callOrder = append(callOrder, "final")
		return nil, nil
	}

	// Execute the chain
	_, err := chain.Execute(context.Background(), "test/module", "v1.0.0", final)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify call order
	if len(callOrder) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(callOrder))
	}

	if callOrder[0] != "middleware1" {
		t.Errorf("Expected first call to be middleware1, got %s", callOrder[0])
	}

	if callOrder[1] != "middleware2" {
		t.Errorf("Expected second call to be middleware2, got %s", callOrder[1])
	}

	if callOrder[2] != "final" {
		t.Errorf("Expected third call to be final, got %s", callOrder[2])
	}
}
