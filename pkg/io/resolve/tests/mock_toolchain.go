package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MockGoToolchain is a mock implementation of the Go toolchain for testing
type MockGoToolchain struct {
	// Map of module paths to filesystem paths
	ModulePaths map[string]string
}

// NewMockGoToolchain creates a new mock Go toolchain
func NewMockGoToolchain() *MockGoToolchain {
	return &MockGoToolchain{
		ModulePaths: make(map[string]string),
	}
}

// RegisterModule registers a module path to filesystem path mapping
func (t *MockGoToolchain) RegisterModule(importPath, fsPath string) {
	t.ModulePaths[importPath] = fsPath
}

// FindModule finds a module's filesystem path
func (t *MockGoToolchain) FindModule(ctx context.Context, importPath, version string) (string, error) {
	if path, ok := t.ModulePaths[importPath]; ok {
		return path, nil
	}
	return "", fmt.Errorf("module %s not found", importPath)
}

// DownloadModule downloads a module (mock implementation)
func (t *MockGoToolchain) DownloadModule(ctx context.Context, importPath, version string) error {
	if _, ok := t.ModulePaths[importPath]; ok {
		return nil
	}
	return fmt.Errorf("failed to download module %s", importPath)
}

// ListVersions lists available versions for a module
func (t *MockGoToolchain) ListVersions(ctx context.Context, importPath string) ([]string, error) {
	if _, ok := t.ModulePaths[importPath]; ok {
		return []string{"v0.0.0"}, nil
	}
	return nil, fmt.Errorf("module %s not found", importPath)
}

// ListGoModules lists Go modules in a directory
func (t *MockGoToolchain) ListGoModules(ctx context.Context, dir string) ([]string, error) {
	var result []string

	for importPath, path := range t.ModulePaths {
		if filepath.Dir(path) == dir {
			result = append(result, importPath)
		}
	}

	return result, nil
}

// MockModuleFS is a mock implementation of the module filesystem for testing
type MockModuleFS struct {
	// Map of file paths to contents
	Files map[string][]byte
}

// NewMockModuleFS creates a new mock module filesystem
func NewMockModuleFS() *MockModuleFS {
	return &MockModuleFS{
		Files: make(map[string][]byte),
	}
}

// AddFile adds a file to the mock filesystem
func (fs *MockModuleFS) AddFile(path string, content []byte) {
	fs.Files[path] = content
}

// ReadFile reads a file from the mock filesystem
func (fs *MockModuleFS) ReadFile(path string) ([]byte, error) {
	if content, ok := fs.Files[path]; ok {
		return content, nil
	}
	return nil, fmt.Errorf("file %s not found", path)
}

// WriteFile writes a file to the mock filesystem
func (fs *MockModuleFS) WriteFile(path string, content []byte, perm os.FileMode) error {
	fs.Files[path] = content
	return nil
}

// FileExists checks if a file exists in the mock filesystem
func (fs *MockModuleFS) FileExists(path string) bool {
	_, ok := fs.Files[path]
	return ok
}

// DirExists checks if a directory exists in the mock filesystem
func (fs *MockModuleFS) DirExists(path string) bool {
	// For simplicity, we'll just check if any file has this path as a prefix
	for filePath := range fs.Files {
		if filepath.Dir(filePath) == path {
			return true
		}
	}
	return false
}

// CheckModuleExists checks if a module exists
func (t *MockGoToolchain) CheckModuleExists(ctx context.Context, importPath, version string) (bool, error) {
	_, ok := t.ModulePaths[importPath]
	return ok, nil
}

// GetModuleInfo gets information about a module
func (t *MockGoToolchain) GetModuleInfo(ctx context.Context, importPath string) (string, string, error) {
	if _, ok := t.ModulePaths[importPath]; ok {
		return importPath, "v0.0.0", nil
	}
	return "", "", fmt.Errorf("module %s not found", importPath)
}

// RunCommand runs a Go command
func (t *MockGoToolchain) RunCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	// Just return empty bytes for mock implementation
	return []byte{}, nil
}

// MkdirAll creates a directory and all parent directories if they don't exist
func (fs *MockModuleFS) MkdirAll(path string, perm os.FileMode) error {
	// For simplicity in the mock, just return nil
	return nil
}

// RemoveAll removes a directory and all its contents
func (fs *MockModuleFS) RemoveAll(path string) error {
	// For simplicity in the mock, just return nil
	return nil
}

// MockFileInfo is a mock implementation of os.FileInfo
type MockFileInfo struct {
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func (fi MockFileInfo) Name() string       { return fi.name }
func (fi MockFileInfo) Size() int64        { return fi.size }
func (fi MockFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi MockFileInfo) ModTime() time.Time { return time.Now() }
func (fi MockFileInfo) IsDir() bool        { return fi.isDir }
func (fi MockFileInfo) Sys() interface{}   { return nil }

// Stat returns file info for a path
func (fs *MockModuleFS) Stat(path string) (os.FileInfo, error) {
	if content, ok := fs.Files[path]; ok {
		// It's a file
		return MockFileInfo{
			name:  filepath.Base(path),
			size:  int64(len(content)),
			mode:  0644,
			isDir: false,
		}, nil
	}

	// Check if it's a directory
	for filePath := range fs.Files {
		if filepath.Dir(filePath) == path {
			return MockFileInfo{
				name:  filepath.Base(path),
				size:  0,
				mode:  0755,
				isDir: true,
			}, nil
		}
	}

	return nil, fmt.Errorf("file or directory %s not found", path)
}

// TempDir creates a temporary directory
func (fs *MockModuleFS) TempDir(dir, pattern string) (string, error) {
	// For simplicity, just return a fixed temp path
	return filepath.Join(dir, "mock-"+pattern), nil
}
