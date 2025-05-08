// Package loader defines interfaces and implementations for loading Go modules.
package loader

import (
	"bitspark.dev/go-tree/pkg/core/module"
)

// LoadOptions defines options for module loading
type LoadOptions struct {
	// Include test files in the loaded module
	IncludeTests bool

	// Include generated files in the loaded module
	IncludeGenerated bool

	// Include build tags to control which files are loaded
	BuildTags []string

	// Load only specific packages (empty means all packages)
	PackagePaths []string

	// Maximum depth for loading dependencies (0 means only direct dependencies)
	DependencyDepth int

	// Whether to load documentation comments
	LoadDocs bool

	// Whether to include AST nodes in the module
	IncludeAST bool
}

// DefaultLoadOptions returns the default load options
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		IncludeTests:     false,
		IncludeGenerated: false,
		BuildTags:        []string{},
		PackagePaths:     []string{},
		DependencyDepth:  0,
		LoadDocs:         true,
		IncludeAST:       false,
	}
}

// ModuleLoader loads a Go module into memory
type ModuleLoader interface {
	// Load parses a Go module and returns its representation
	Load(dir string) (*module.Module, error)

	// LoadWithOptions parses a Go module with custom options
	LoadWithOptions(dir string, options LoadOptions) (*module.Module, error)
}
