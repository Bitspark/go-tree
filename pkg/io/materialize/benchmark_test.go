package materialize

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/io/resolve"
)

// getTestModulePath returns the absolute path to a test module
func getTestModulePath(moduleName string) (string, error) {
	// First, check relative to the current directory (for running tests from IDE)
	path := filepath.Join("testdata", moduleName)
	if _, err := os.Stat(path); err == nil {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	// Otherwise, try relative to the io package root
	path = filepath.Join("..", "..", "testdata", moduleName)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

// createTestResolver creates a resolver for test modules
func createTestResolver() *resolve.ModuleResolver {
	registry := resolve.NewStandardModuleRegistry()
	resolver := resolve.NewModuleResolver().WithRegistry(registry)

	// Register test modules
	modulePaths := map[string]string{
		"simplemath":    "github.com/test/simplemath",
		"errors":        "github.com/test/errors",
		"complexreturn": "github.com/test/complexreturn",
	}

	for name, importPath := range modulePaths {
		modulePath, err := getTestModulePath(name)
		if err == nil {
			registry.RegisterModule(importPath, modulePath, true)
		}
	}

	return resolver
}

// BenchmarkMaterialize benchmarks the materialization process with different layout strategies
func BenchmarkMaterialize(b *testing.B) {
	// Create resolver with test modules
	resolver := createTestResolver()

	// Use simplemath module for benchmarking (small and simple)
	moduleName := "simplemath"

	// Get module path
	modulePath, err := getTestModulePath(moduleName)
	if err != nil {
		b.Fatalf("Failed to get test module path: %v", err)
	}

	// Resolve module
	resolveOpts := resolve.DefaultResolveOptions()
	resolveOpts.DownloadMissing = false
	module, err := resolver.ResolveModule(modulePath, "", resolveOpts)
	if err != nil {
		b.Fatalf("Failed to resolve module: %v", err)
	}

	// Test different layout strategies
	strategies := []struct {
		name     string
		strategy LayoutStrategy
	}{
		{"flat", FlatLayout},
		{"hierarchical", HierarchicalLayout},
		{"gopath", GoPathLayout},
	}

	// Test different dependency policies
	policies := []struct {
		name   string
		policy DependencyPolicy
	}{
		{"no-deps", NoDependencies},
		{"direct-deps", DirectDependenciesOnly},
	}

	// Run benchmarks for each combination
	for _, strategy := range strategies {
		for _, policy := range policies {
			benchName := fmt.Sprintf("%s/%s", strategy.name, policy.name)
			b.Run(benchName, func(b *testing.B) {
				// Create materializer
				materializer := NewModuleMaterializer()

				// Set up options
				opts := DefaultMaterializeOptions()
				opts.LayoutStrategy = strategy.strategy
				opts.DependencyPolicy = policy.policy
				opts.RunGoModTidy = false // Skip tidy for benchmarks

				// Reset timer before the loop
				b.ResetTimer()

				// Run the benchmark
				for i := 0; i < b.N; i++ {
					env, err := materializer.Materialize(module, opts)
					if err != nil {
						b.Fatalf("Failed to materialize module: %v", err)
					}
					// Clean up after each run
					env.Cleanup()
				}
			})
		}
	}
}

// BenchmarkMaterializeComplexModule benchmarks materializing a more complex module
func BenchmarkMaterializeComplexModule(b *testing.B) {
	// Create resolver with test modules
	resolver := createTestResolver()

	// Use complexreturn module for benchmarking (more complex)
	moduleName := "complexreturn"

	// Get module path
	modulePath, err := getTestModulePath(moduleName)
	if err != nil {
		b.Fatalf("Failed to get test module path: %v", err)
	}

	// Resolve module
	resolveOpts := resolve.DefaultResolveOptions()
	resolveOpts.DownloadMissing = false
	module, err := resolver.ResolveModule(modulePath, "", resolveOpts)
	if err != nil {
		b.Fatalf("Failed to resolve module: %v", err)
	}

	// Create materializer
	materializer := NewModuleMaterializer()

	// Set up options
	opts := DefaultMaterializeOptions()
	opts.RunGoModTidy = false // Skip tidy for benchmarks

	// Reset timer before the loop
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		env, err := materializer.Materialize(module, opts)
		if err != nil {
			b.Fatalf("Failed to materialize module: %v", err)
		}
		// Clean up after each run
		env.Cleanup()
	}
}
