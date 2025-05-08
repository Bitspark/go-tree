// Package toolkit provides abstractions for external dependencies like the Go toolchain and filesystem.
package toolkit

import (
	"context"
)

// GoToolchain defines operations for interacting with the Go toolchain
type GoToolchain interface {
	// RunCommand executes a Go command with arguments
	RunCommand(ctx context.Context, command string, args ...string) ([]byte, error)

	// GetModuleInfo retrieves information about a module
	GetModuleInfo(ctx context.Context, importPath string) (path string, version string, err error)

	// DownloadModule downloads a module
	DownloadModule(ctx context.Context, importPath string, version string) error

	// FindModule locates a module in the module cache
	FindModule(ctx context.Context, importPath string, version string) (dir string, err error)

	// CheckModuleExists verifies a module exists and is accessible
	CheckModuleExists(ctx context.Context, importPath string, version string) (bool, error)
}
