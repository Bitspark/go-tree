package env

import (
	"os"
)

// StandardModuleFS provides filesystem operations using the standard library
type StandardModuleFS struct {
	// No configuration needed for the standard implementation
}

// NewStandardModuleFS creates a new standard filesystem implementation
func NewStandardModuleFS() *StandardModuleFS {
	return &StandardModuleFS{}
}

// ReadFile reads a file from the filesystem
func (fs *StandardModuleFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file
func (fs *StandardModuleFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// MkdirAll creates a directory with all necessary parents
func (fs *StandardModuleFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// RemoveAll removes a path and any children
func (fs *StandardModuleFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Stat returns file info
func (fs *StandardModuleFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// TempDir creates a temporary directory
func (fs *StandardModuleFS) TempDir(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}
