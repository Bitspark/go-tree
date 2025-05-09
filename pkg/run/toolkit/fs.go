package toolkit

import (
	"os"
)

// ModuleFS defines filesystem operations for modules
type ModuleFS interface {
	// ReadFile reads a file from the filesystem
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file
	WriteFile(path string, data []byte, perm os.FileMode) error

	// MkdirAll creates a directory with all necessary parents
	MkdirAll(path string, perm os.FileMode) error

	// RemoveAll removes a path and any children
	RemoveAll(path string) error

	// Stat returns file info
	Stat(path string) (os.FileInfo, error)

	// TempDir creates a temporary directory
	TempDir(dir, pattern string) (string, error)
}
