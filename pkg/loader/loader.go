// Package loader provides functionality for loading Go modules with full type information.
// It integrates with the typesys package to extract and organize types, symbols, and references.
package loader

import (
	"fmt"

	"bitspark.dev/go-tree/pkg/typesys"
)

// LoadModule loads a Go module with full type checking.
func LoadModule(dir string, opts *typesys.LoadOptions) (*typesys.Module, error) {
	if opts == nil {
		opts = &typesys.LoadOptions{
			IncludeTests:   false,
			IncludePrivate: true,
		}
	}

	// Normalize and make directory path absolute
	dir = ensureAbsolutePath(normalizePath(dir))

	// Create a new module
	module := typesys.NewModule(dir)

	// Load packages
	if err := loadPackages(module, opts); err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	return module, nil
}
