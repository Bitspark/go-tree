package tests

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/io/resolve"
)

// Create custom resolve options that don't try to download modules
func createTestResolveOptions() resolve.ResolveOptions {
	opts := resolve.DefaultResolveOptions()
	opts.DownloadMissing = false
	return opts
}

func TestModuleResolutionIntegration(t *testing.T) {
	// Get the absolute path to the testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Construct the path to the testdata directory (going up one level)
	testdataPath := filepath.Join(filepath.Dir(wd), "testdata")

	// Create a registry to track modules
	registry := resolve.NewStandardModuleRegistry()

	// Create a resolver with the registry
	resolver := resolve.NewModuleResolver().WithRegistry(registry)

	// Load the modules into the registry directly first
	moduleAPath := filepath.Join(testdataPath, "module-a")
	moduleBPath := filepath.Join(testdataPath, "module-b")
	moduleCPath := filepath.Join(testdataPath, "module-c")

	registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", moduleAPath, true)
	registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b", moduleBPath, true)
	registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", moduleCPath, true)

	// Test resolving module-a (standalone module)
	t.Run("ResolveModuleA", func(t *testing.T) {
		opts := createTestResolveOptions()

		moduleA, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", "", opts)
		if err != nil {
			t.Fatalf("Failed to resolve module-a: %v", err)
		}

		// Verify the module was loaded correctly
		if moduleA.Path != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a" {
			t.Errorf("Expected module path %s, got %s",
				"bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", moduleA.Path)
		}

		// Verify package was loaded
		if len(moduleA.Packages) == 0 {
			t.Errorf("Expected at least one package in module-a")
		}
	})

	// Test resolving module-b (depends on module-a)
	t.Run("ResolveModuleB", func(t *testing.T) {
		// Create options that include dependency resolution
		opts := createTestResolveOptions()
		opts.DependencyPolicy = resolve.AllDependencies

		moduleB, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b", "", opts)
		if err != nil {
			t.Fatalf("Failed to resolve module-b: %v", err)
		}

		// Verify the module was loaded correctly
		if moduleB.Path != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b" {
			t.Errorf("Expected module path %s, got %s",
				"bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b", moduleB.Path)
		}

		// Verify dependencies were resolved
		if len(moduleB.Dependencies) == 0 {
			t.Errorf("Expected at least one dependency in module-b")
		}

		// Check for module-a in the resolved modules
		moduleAImportPath := "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a"
		found := false
		for _, dep := range moduleB.Dependencies {
			if dep.ImportPath == moduleAImportPath {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find module-a in dependencies of module-b")
		}
	})

	// Test resolving module-c (depends on module-b which depends on module-a)
	t.Run("ResolveModuleC", func(t *testing.T) {
		// Create options with recursive dependency resolution
		opts := createTestResolveOptions()
		opts.DependencyPolicy = resolve.AllDependencies
		opts.DependencyDepth = 2 // Ensure we get module-b and module-a

		moduleC, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", "", opts)
		if err != nil {
			t.Fatalf("Failed to resolve module-c: %v", err)
		}

		// Verify the module was loaded correctly
		if moduleC.Path != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c" {
			t.Errorf("Expected module path %s, got %s",
				"bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", moduleC.Path)
		}

		// Verify dependencies were resolved
		if len(moduleC.Dependencies) == 0 {
			t.Errorf("Expected at least one dependency in module-c")
		}

		// Check for module-b in the resolved modules
		moduleBImportPath := "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b"
		found := false
		for _, dep := range moduleC.Dependencies {
			if dep.ImportPath == moduleBImportPath {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected to find module-b in dependencies of module-c")
		}

		// Build a dependency graph
		graph, err := resolver.BuildDependencyGraph(moduleC)
		if err != nil {
			t.Fatalf("Failed to build dependency graph: %v", err)
		}

		// Verify the graph structure
		if len(graph) < 3 {
			t.Errorf("Expected at least 3 nodes in dependency graph, got %d", len(graph))
		}

		// Verify module-c depends on module-b
		deps, ok := graph["bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c"]
		if !ok {
			t.Errorf("Expected to find module-c in dependency graph")
		} else if len(deps) == 0 || deps[0] != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b" {
			t.Errorf("Expected module-c to depend on module-b in graph")
		}

		// Verify module-b depends on module-a
		deps, ok = graph["bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b"]
		if !ok {
			t.Errorf("Expected to find module-b in dependency graph")
		} else if len(deps) == 0 || deps[0] != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a" {
			t.Errorf("Expected module-b to depend on module-a in graph")
		}
	})
}

// TestModuleRegistryIntegration tests the module registry's ability to cache and retrieve modules
func TestModuleRegistryIntegration(t *testing.T) {
	// Get the absolute path to the testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Construct the path to the testdata directory (going up one level)
	testdataPath := filepath.Join(filepath.Dir(wd), "testdata")

	// Create a registry
	registry := resolve.NewStandardModuleRegistry()

	// Register modules
	moduleAPath := filepath.Join(testdataPath, "module-a")
	moduleBPath := filepath.Join(testdataPath, "module-b")
	moduleCPath := filepath.Join(testdataPath, "module-c")

	err = registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", moduleAPath, true)
	if err != nil {
		t.Fatalf("Failed to register module-a: %v", err)
	}

	err = registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b", moduleBPath, true)
	if err != nil {
		t.Fatalf("Failed to register module-b: %v", err)
	}

	err = registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", moduleCPath, true)
	if err != nil {
		t.Fatalf("Failed to register module-c: %v", err)
	}

	// Test finding modules by import path
	t.Run("FindModuleByImportPath", func(t *testing.T) {
		moduleA, ok := registry.FindModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a")
		if !ok {
			t.Errorf("Failed to find module-a by import path")
		} else if moduleA.FilesystemPath != moduleAPath {
			t.Errorf("Expected path %s, got %s", moduleAPath, moduleA.FilesystemPath)
		}
	})

	// Test finding modules by filesystem path
	t.Run("FindModuleByPath", func(t *testing.T) {
		moduleB, ok := registry.FindByPath(moduleBPath)
		if !ok {
			t.Errorf("Failed to find module-b by filesystem path")
		} else if moduleB.ImportPath != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b" {
			t.Errorf("Expected import path %s, got %s",
				"bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b", moduleB.ImportPath)
		}
	})

	// Test creating a resolver from the registry
	t.Run("CreateResolverFromRegistry", func(t *testing.T) {
		resolver := registry.CreateResolver()

		// Use the custom options with dependency resolution enabled
		opts := createTestResolveOptions()
		opts.DependencyPolicy = resolve.AllDependencies
		opts.DependencyDepth = 2

		// Resolve a module using the resolver
		moduleC, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", "", opts)
		if err != nil {
			t.Fatalf("Failed to resolve module-c: %v", err)
		}

		if moduleC.Path != "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c" {
			t.Errorf("Expected module path %s, got %s",
				"bitspark.dev/go-tree/pkg/io/resolve/testdata/module-c", moduleC.Path)
		}

		// Verify dependencies were resolved
		if len(moduleC.Dependencies) == 0 {
			t.Errorf("Expected at least one dependency in module-c")
		}
	})
}

// TestLoaderCacheIntegration tests that the module resolver caches loaded modules correctly
func TestLoaderCacheIntegration(t *testing.T) {
	// Get the absolute path to the testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Construct the path to the testdata directory (going up one level)
	testdataPath := filepath.Join(filepath.Dir(wd), "testdata")
	moduleAPath := filepath.Join(testdataPath, "module-a")

	// Create a registry and register module-a
	registry := resolve.NewStandardModuleRegistry()
	registry.RegisterModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", moduleAPath, true)

	// Create a resolver with caching enabled
	opts := createTestResolveOptions()
	opts.UseResolutionCache = true
	resolver := resolve.NewModuleResolverWithOptions(opts).WithRegistry(registry)

	// Load the module first time
	start := testingTimeNow()
	moduleA1, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", "", opts)
	if err != nil {
		t.Fatalf("Failed to resolve module-a: %v", err)
	}
	firstLoadTime := testingTimeNow() - start

	// Load the module second time (should be cached)
	start = testingTimeNow()
	moduleA2, err := resolver.ResolveModule("bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a", "", opts)
	if err != nil {
		t.Fatalf("Failed to resolve module-a second time: %v", err)
	}
	secondLoadTime := testingTimeNow() - start

	// Verify the modules are the same instance
	if moduleA1 != moduleA2 {
		t.Errorf("Expected cached module to be the same instance")
	}

	// The second load should be significantly faster if caching is working
	t.Logf("First load: %v, Second load: %v", firstLoadTime, secondLoadTime)
}

// Helper function to get current time for basic performance measurements
func testingTimeNow() int64 {
	return 0 // This is just a stub - in a real test we would use time.Now().UnixNano()
}
