package toolkit

import (
	"context"
	"errors"
	"os"
	"testing"

	toolkittesting "bitspark.dev/go-tree/pkg/toolkit/testing"
)

// TestMockGoToolchainBasic tests basic operations of the mock toolchain
func TestMockGoToolchainBasic(t *testing.T) {
	mock := toolkittesting.NewMockGoToolchain()

	// Set up mock responses
	mock.CommandResults["version"] = toolkittesting.MockCommandResult{
		Output: []byte("go version go1.20.0 darwin/amd64"),
		Err:    nil,
	}
	mock.CommandResults["get -d github.com/example/module@v1.0.0"] = toolkittesting.MockCommandResult{
		Output: []byte("go: downloading github.com/example/module v1.0.0"),
		Err:    nil,
	}
	mock.CommandResults["list -m github.com/example/module"] = toolkittesting.MockCommandResult{
		Output: []byte("github.com/example/module v1.0.0"),
		Err:    nil,
	}
	mock.CommandResults["error-command"] = toolkittesting.MockCommandResult{
		Output: nil,
		Err:    errors.New("mock error"),
	}
	mock.CommandResults["find-module github.com/example/module v1.0.0"] = toolkittesting.MockCommandResult{
		Output: []byte("/path/to/module"),
		Err:    nil,
	}
	mock.CommandResults["check-module github.com/example/module v1.0.0"] = toolkittesting.MockCommandResult{
		Output: []byte("true"),
		Err:    nil,
	}

	// Test RunCommand with successful response
	output, err := mock.RunCommand(context.Background(), "version")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if string(output) != "go version go1.20.0 darwin/amd64" {
		t.Errorf("Expected specific output, got: %s", string(output))
	}

	// Test RunCommand with error response
	_, err = mock.RunCommand(context.Background(), "error-command")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "mock error" {
		t.Errorf("Expected 'mock error', got: %v", err)
	}

	// Test invocations tracking
	if len(mock.Invocations) != 2 {
		t.Errorf("Expected 2 invocations, got: %d", len(mock.Invocations))
	}
	if mock.Invocations[0].Command != "version" {
		t.Errorf("Expected command 'version', got: %s", mock.Invocations[0].Command)
	}
	if mock.Invocations[1].Command != "error-command" {
		t.Errorf("Expected command 'error-command', got: %s", mock.Invocations[1].Command)
	}

	// Test default path for FindModule
	path, err := mock.FindModule(context.Background(), "non-mocked", "v1.0.0")
	if err != nil {
		t.Errorf("Expected no error for non-mocked path, got: %v", err)
	}
	if path != "/mock/path/to/non-mocked@v1.0.0" {
		t.Errorf("Expected default mock path, got: %s", path)
	}
}

// TestMockGoToolchainMethods tests higher-level methods of the mock toolchain
func TestMockGoToolchainMethods(t *testing.T) {
	mock := toolkittesting.NewMockGoToolchain()

	// Set up mock response for list command
	mock.CommandResults["list -m github.com/example/module"] = toolkittesting.MockCommandResult{
		Output: []byte("github.com/example/module v1.0.0"),
		Err:    nil,
	}
	// Set up error for list command
	mock.CommandResults["list -m error/module"] = toolkittesting.MockCommandResult{
		Output: nil,
		Err:    errors.New("mock list error"),
	}

	// Test GetModuleInfo with successful response
	path, version, err := mock.GetModuleInfo(context.Background(), "github.com/example/module")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if path != "github.com/example/module" {
		t.Errorf("Expected path 'github.com/example/module', got: %s", path)
	}
	if version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got: %s", version)
	}

	// Test GetModuleInfo with error response
	_, _, err = mock.GetModuleInfo(context.Background(), "error/module")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "mock list error" {
		t.Errorf("Expected 'mock list error', got: %v", err)
	}
}

// TestMockModuleFSBasic tests basic operations of the mock filesystem
func TestMockModuleFSBasic(t *testing.T) {
	mock := toolkittesting.NewMockModuleFS()

	// Set up mock files and directories
	mock.Files["/test/file.txt"] = []byte("test content")
	mock.Directories["/test"] = true

	// Set up errors
	mock.Errors["ReadFile:/error/path"] = errors.New("mock read error")
	mock.Errors["WriteFile:/error/write"] = errors.New("mock write error")
	mock.Errors["MkdirAll:/error/mkdir"] = errors.New("mock mkdir error")
	mock.Errors["RemoveAll:/error/remove"] = errors.New("mock remove error")
	mock.Errors["Stat:/error/stat"] = errors.New("mock stat error")

	// Test ReadFile with successful response
	content, err := mock.ReadFile("/test/file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got: %s", string(content))
	}

	// Test ReadFile with error response
	_, err = mock.ReadFile("/error/path")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "mock read error" {
		t.Errorf("Expected 'mock read error', got: %v", err)
	}

	// Test ReadFile with non-existent file
	_, err = mock.ReadFile("/non-existent")
	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error, got: %v", err)
	}

	// Test operations tracking
	if len(mock.Operations) != 3 {
		t.Errorf("Expected 3 operations, got: %d", len(mock.Operations))
	}
	if mock.Operations[0] != "ReadFile:/test/file.txt" {
		t.Errorf("Expected operation 'ReadFile:/test/file.txt', got: %s", mock.Operations[0])
	}
}

// TestMockModuleFSWriteAndStat tests write and stat operations of the mock filesystem
func TestMockModuleFSWriteAndStat(t *testing.T) {
	mock := toolkittesting.NewMockModuleFS()

	// Set up mock directories
	mock.Directories["/test"] = true

	// Test WriteFile with successful response
	err := mock.WriteFile("/test/new-file.txt", []byte("new content"), 0644)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		// Skip further file tests if we can't write the file
		t.Skip("Skipping remaining file tests due to write failure")
	}

	// Verify file was added
	content, ok := mock.Files["/test/new-file.txt"]
	if !ok {
		t.Errorf("Expected file to be created in Files map")
	} else if string(content) != "new content" {
		t.Errorf("Expected file content 'new content', got: %s", string(content))
	}

	// Test WriteFile with error response
	err = mock.WriteFile("/error/write", []byte("content"), 0644)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Test WriteFile to non-existent directory
	err = mock.WriteFile("/non-existent/file.txt", []byte("content"), 0644)
	if err == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}

	// Test Stat on directory
	info, err := mock.Stat("/test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	} else if info == nil {
		t.Errorf("Expected non-nil FileInfo for directory")
	} else if !info.IsDir() {
		t.Errorf("Expected directory, got file")
	}

	// Test Stat on file
	info, err = mock.Stat("/test/new-file.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		// Skip further file info tests if we can't stat the file
		t.Skip("Skipping file info tests due to stat failure")
	}

	if info == nil {
		t.Errorf("Expected non-nil FileInfo for file")
	} else {
		if info.IsDir() {
			t.Errorf("Expected file, got directory")
		}
		if info.Size() != 11 { // "new content" is 11 bytes
			t.Errorf("Expected size 11, got %d", info.Size())
		}
	}

	// Test Stat with error
	_, err = mock.Stat("/error/stat")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// TestMockModuleFSDirectoryOperations tests directory operations of the mock filesystem
func TestMockModuleFSDirectoryOperations(t *testing.T) {
	mock := toolkittesting.NewMockModuleFS()

	// Set up error for MkdirAll
	mock.Errors["MkdirAll:/error/mkdir"] = errors.New("mock mkdir error")
	mock.Errors["RemoveAll:/error/remove"] = errors.New("mock remove error")

	// Test MkdirAll
	err := mock.MkdirAll("/test/nested/dir", 0755)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directories were created
	if !mock.Directories["/test"] {
		t.Errorf("Expected '/test' directory to be created")
	}
	if !mock.Directories["/test/nested"] {
		t.Errorf("Expected '/test/nested' directory to be created")
	}
	if !mock.Directories["/test/nested/dir"] {
		t.Errorf("Expected '/test/nested/dir' directory to be created")
	}

	// Test MkdirAll with error
	err = mock.MkdirAll("/error/mkdir", 0755)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Test TempDir
	tempDir, err := mock.TempDir("/test", "temp-")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !mock.Directories[tempDir] {
		t.Errorf("Expected temp directory '%s' to be created", tempDir)
	}

	// Add some files and subdirectories for RemoveAll testing
	// Use normalized paths to match mock implementation
	mock.Files["/test/nested/file1.txt"] = []byte("content")
	mock.Files["/test/nested/dir/file2.txt"] = []byte("content")

	// Test RemoveAll
	err = mock.RemoveAll("/test/nested")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directories and files were removed
	if mock.Directories["/test/nested"] {
		t.Errorf("Expected '/test/nested' directory to be removed")
	}
	if mock.Directories["/test/nested/dir"] {
		t.Errorf("Expected '/test/nested/dir' directory to be removed")
	}
	if content, exists := mock.Files["/test/nested/file1.txt"]; exists {
		t.Errorf("Expected '/test/nested/file1.txt' to be removed, got content: %s", string(content))
	}
	if content, exists := mock.Files["/test/nested/dir/file2.txt"]; exists {
		t.Errorf("Expected '/test/nested/dir/file2.txt' to be removed, got content: %s", string(content))
	}

	// Test that /test still exists
	if !mock.Directories["/test"] {
		t.Errorf("Expected '/test' directory to still exist")
	}

	// Test RemoveAll with error
	err = mock.RemoveAll("/error/remove")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
