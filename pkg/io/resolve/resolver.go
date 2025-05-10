// Package resolve provides module resolution and dependency handling capabilities.
// It handles locating modules on the filesystem, resolving dependencies, and managing module versions.
package resolve

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Resolver defines the interface for module resolution
type Resolver interface {
	// ResolveModule resolves a module by path and version
	ResolveModule(path, version string, opts ResolveOptions) (*typesys.Module, error)

	// ResolveDependencies resolves dependencies for a module
	ResolveDependencies(module *typesys.Module, depth int) error

	// FindModuleLocation finds a module's location in the filesystem
	FindModuleLocation(importPath, version string) (string, error)

	// EnsureModuleAvailable ensures a module is available, downloading if necessary
	EnsureModuleAvailable(importPath, version string) (string, error)

	// FindModuleVersion finds the latest version of a module
	FindModuleVersion(importPath string) (string, error)

	// BuildDependencyGraph builds a dependency graph for visualization
	BuildDependencyGraph(module *typesys.Module) (map[string][]string, error)

	// AddDependency adds a dependency to a module and loads it
	AddDependency(module *typesys.Module, importPath, version string) error

	// RemoveDependency removes a dependency from a module
	RemoveDependency(module *typesys.Module, importPath string) error

	// FindModuleByDir finds a module by its directory
	FindModuleByDir(dir string) (*typesys.Module, bool)
}

// ResolutionError represents a specific resolution-related error with context
type ResolutionError struct {
	ImportPath string
	Version    string
	Module     string
	Reason     string
	Err        error
}

// Error returns a string representation of the error
func (e *ResolutionError) Error() string {
	msg := "module resolution error"
	if e.ImportPath != "" {
		msg += " for " + e.ImportPath
		if e.Version != "" {
			msg += "@" + e.Version
		}
	}
	if e.Module != "" {
		msg += " in module " + e.Module
	}
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}
