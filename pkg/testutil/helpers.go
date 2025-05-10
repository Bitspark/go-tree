// Package testutil provides helper functions for execute package integration tests
package testutil

import (
	"bitspark.dev/go-tree/pkg/io/materialize"
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/resolve"
	"bitspark.dev/go-tree/pkg/run/execute"
	"bitspark.dev/go-tree/pkg/run/execute/specialized"
)

// TestModuleResolver is a resolver specifically for tests that can handle test modules
type TestModuleResolver struct {
	baseResolver *resolve.ModuleResolver
	moduleCache  map[string]*typesys.Module
	pathMappings map[string]string // Maps import path to filesystem path
	registry     *resolve.StandardModuleRegistry
}

// NewTestModuleResolver creates a new resolver for tests
func NewTestModuleResolver() *TestModuleResolver {
	registry := resolve.NewStandardModuleRegistry()

	r := &TestModuleResolver{
		baseResolver: resolve.NewModuleResolver().WithRegistry(registry),
		moduleCache:  make(map[string]*typesys.Module),
		pathMappings: make(map[string]string),
		registry:     registry,
	}

	// Pre-register the standard test modules
	registerTestModules(r)

	return r
}

// MapModule registers a filesystem path to be used for a specific import path
func (r *TestModuleResolver) MapModule(importPath, fsPath string) {
	r.pathMappings[importPath] = fsPath

	// Also register with the registry
	r.registry.RegisterModule(importPath, fsPath, true)
}

// ResolveModule implements the execute.ModuleResolver interface
func (r *TestModuleResolver) ResolveModule(path, version string, opts interface{}) (*typesys.Module, error) {
	// Check if this is a filesystem path first
	if _, err := os.Stat(path); err == nil {
		// This is a filesystem path, load it directly
		resolveOpts := toResolveOptions(opts)
		module, err := r.baseResolver.ResolveModule(path, "", resolveOpts)
		if err != nil {
			return nil, err
		}

		// Cache by both filesystem path and import path (from go.mod)
		r.moduleCache[path] = module
		if module.Path != "" {
			r.moduleCache[module.Path] = module
			r.pathMappings[module.Path] = path

			// Register with the registry
			r.registry.RegisterModule(module.Path, path, true)
		}

		return module, nil
	}

	// Check if we have a mapping for this import path
	if fsPath, ok := r.pathMappings[path]; ok {
		// Check cache first
		if module, ok := r.moduleCache[path]; ok {
			return module, nil
		}

		// Load from the mapped filesystem path
		resolveOpts := toResolveOptions(opts)
		module, err := r.baseResolver.ResolveModule(fsPath, "", resolveOpts)
		if err != nil {
			return nil, err
		}

		// Cache the result
		r.moduleCache[path] = module
		r.moduleCache[fsPath] = module

		return module, nil
	}

	// Fall back to standard resolver
	return r.baseResolver.ResolveModule(path, version, toResolveOptions(opts))
}

// GetRegistry returns the module registry
func (r *TestModuleResolver) GetRegistry() interface{} {
	return r.registry
}

// ResolveDependencies implements the execute.ModuleResolver interface
func (r *TestModuleResolver) ResolveDependencies(module interface{}, depth int) error {
	// For test modules, we don't need to resolve dependencies
	return nil
}

// Helper to convert interface{} to resolve.ResolveOptions
func toResolveOptions(opts interface{}) resolve.ResolveOptions {
	if opts == nil {
		return resolve.ResolveOptions{
			DownloadMissing: false, // Disable auto-download for tests
		}
	}

	if resolveOpts, ok := opts.(resolve.ResolveOptions); ok {
		// Make sure auto-download is disabled for tests
		resolveOpts.DownloadMissing = false
		return resolveOpts
	}

	return resolve.ResolveOptions{
		DownloadMissing: false, // Disable auto-download for tests
	}
}

// GetTestModulePath returns the absolute path to a test module
func GetTestModulePath(moduleName string) (string, error) {
	// First, check relative to the current directory (for running tests from IDE)
	path := filepath.Join("testdata", moduleName)
	if _, err := os.Stat(path); err == nil {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	// Otherwise, try relative to the execute package root
	path = filepath.Join("..", "..", "testdata", moduleName)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// CreateRunner creates a function runner with real dependencies
func CreateRunner() *execute.FunctionRunner {
	// Create a test resolver that can handle local modules
	resolver := NewTestModuleResolver()

	// Pre-register the common test modules
	registerTestModules(resolver)

	materializer := materialize.NewModuleMaterializer()

	return execute.NewFunctionRunner(resolver, materializer)
}

// registerTestModules registers all test modules with the resolver
func registerTestModules(resolver *TestModuleResolver) {
	// Register the standard test modules
	registerModule(resolver, "simplemath", "github.com/test/simplemath")
	registerModule(resolver, "errors", "github.com/test/errors")
	registerModule(resolver, "complexreturn", "github.com/test/complexreturn")
}

// registerModule registers a single test module
func registerModule(resolver *TestModuleResolver, moduleName, importPath string) {
	modulePath, err := GetTestModulePath(moduleName)
	if err == nil {
		resolver.MapModule(importPath, modulePath)
	}
}

// CreateRetryingRunner creates a retrying function runner with real dependencies
func CreateRetryingRunner() *specialized.RetryingFunctionRunner {
	baseRunner := CreateRunner()
	return specialized.NewRetryingFunctionRunner(baseRunner)
}

// CreateBatchRunner creates a batch function runner with real dependencies
func CreateBatchRunner() *specialized.BatchFunctionRunner {
	baseRunner := CreateRunner()
	return specialized.NewBatchFunctionRunner(baseRunner)
}

// CreateTypedRunner creates a typed function runner with real dependencies
func CreateTypedRunner() *specialized.TypedFunctionRunner {
	baseRunner := CreateRunner()
	return specialized.NewTypedFunctionRunner(baseRunner)
}

// CreateCachedRunner creates a cached function runner with real dependencies
func CreateCachedRunner() *specialized.CachedFunctionRunner {
	baseRunner := CreateRunner()
	return specialized.NewCachedFunctionRunner(baseRunner)
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(prefix string) (string, error) {
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	return tempDir, nil
}
