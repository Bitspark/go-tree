package env

import (
	"context"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// RegistryAwareMiddleware provides module resolution based on a registry
type RegistryAwareMiddleware struct {
	registry interface{} // Registry implementing FindByPath and FindModule methods
}

// NewRegistryAwareMiddleware creates a new registry-aware middleware
func NewRegistryAwareMiddleware(registry interface{}) *RegistryAwareMiddleware {
	return &RegistryAwareMiddleware{
		registry: registry,
	}
}

// Execute implements the ResolutionMiddleware interface
func (m *RegistryAwareMiddleware) Execute(ctx context.Context, path, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
	// If we have no registry, just call next
	if m.registry == nil {
		module, err := next()
		return ctx, module, err
	}

	// Check if this is a filesystem path
	if filepath.IsAbs(path) {
		// This is an absolute filesystem path, check if we have it in the registry
		// Use type assertion for registry methods
		if finder, ok := m.registry.(interface {
			FindByPath(string) (interface{}, bool)
		}); ok {
			if resolvedModule, ok := finder.FindByPath(path); ok {
				// Extract the module using reflection
				if moduleGetter, ok := resolvedModule.(interface {
					GetModule() *typesys.Module
				}); ok && moduleGetter.GetModule() != nil {
					return ctx, moduleGetter.GetModule(), nil
				}
			}
		}
	} else {
		// Check if we have this import path in the registry
		if finder, ok := m.registry.(interface {
			FindModule(string) (interface{}, bool)
		}); ok {
			if resolvedModule, ok := finder.FindModule(path); ok {
				// Extract the module and filesystem path using reflection
				var fsPath string
				var module *typesys.Module

				if pathGetter, ok := resolvedModule.(interface {
					GetFilesystemPath() string
				}); ok {
					fsPath = pathGetter.GetFilesystemPath()
				}

				if moduleGetter, ok := resolvedModule.(interface {
					GetModule() *typesys.Module
				}); ok {
					module = moduleGetter.GetModule()
				}

				if module != nil {
					return ctx, module, nil
				} else if fsPath != "" {
					// Module not loaded yet, update the path to the filesystem path
					// This is just a hint for the resolver, which may or may not use it
					path = fsPath
				}
			}
		}
	}

	// Continue with normal resolution
	module, err := next()
	return ctx, module, err
}
