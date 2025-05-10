package materialize

import (
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/testutil"
)

// TestMaterializeRealModules tests materializing real modules with different layout strategies
func TestMaterializeRealModules(t *testing.T) {
	// Create a test module resolver
	resolver := testutil.NewTestModuleResolver()

	// Test with each of our test modules
	testModules := []string{"simplemath", "complexreturn", "errors"}

	for _, moduleName := range testModules {
		t.Run(moduleName, func(t *testing.T) {
			// Get module path
			modulePath, err := testutil.GetTestModulePath(moduleName)
			if err != nil {
				t.Fatalf("Failed to get test module path: %v", err)
			}

			// Resolve the module
			importPath := "github.com/test/" + moduleName
			module, err := resolver.ResolveModule(importPath, "", nil)
			if err != nil {
				t.Fatalf("Failed to resolve module: %v", err)
			}

			// Create materializer
			materializer := NewModuleMaterializer()

			// Set up options for different test cases
			layoutStrategies := []struct {
				name     string
				strategy LayoutStrategy
			}{
				{"flat", FlatLayout},
				{"hierarchical", HierarchicalLayout},
				{"gopath", GoPathLayout},
			}

			for _, layout := range layoutStrategies {
				t.Run(layout.name, func(t *testing.T) {
					// Create options with this layout
					opts := DefaultMaterializeOptions()
					opts.LayoutStrategy = layout.strategy
					opts.Registry = resolver.GetRegistry()

					// Materialize the module
					env, err := materializer.Materialize(module, opts)
					if err != nil {
						t.Fatalf("Failed to materialize module: %v", err)
					}
					defer env.Cleanup()

					// Verify correct layout was used
					verifyLayoutStrategy(t, env, module, layout.strategy)

					// Verify all files were materialized
					verifyFilesExist(t, env, module)
				})
			}
		})
	}
}

// verifyLayoutStrategy verifies that the correct layout strategy was used
func verifyLayoutStrategy(t *testing.T, env *Environment, module *typesys.Module, strategy LayoutStrategy) {
	modulePath, ok := env.ModulePaths[module.Path]
	if !ok {
		t.Fatalf("Module path %s missing from environment", module.Path)
	}

	switch strategy {
	case FlatLayout:
		// Expect module in a flat directory structure
		base := filepath.Base(modulePath)
		expected := strings.ReplaceAll(module.Path, "/", "_")
		if base != expected {
			t.Errorf("Expected base directory %s for flat layout, got %s", expected, base)
		}
	case HierarchicalLayout:
		// Expect module path to end with the full import path
		if !strings.HasSuffix(filepath.ToSlash(modulePath), module.Path) {
			t.Errorf("Expected hierarchical path to end with %s, got %s", module.Path, modulePath)
		}
	case GoPathLayout:
		// Expect GOPATH-like structure with src directory
		if !strings.Contains(filepath.ToSlash(modulePath), "src/"+module.Path) {
			t.Errorf("Expected GOPATH layout to contain src/%s, got %s", module.Path, modulePath)
		}
	}
}

// verifyFilesExist verifies that all expected files were materialized
func verifyFilesExist(t *testing.T, env *Environment, module *typesys.Module) {
	modulePath, ok := env.ModulePaths[module.Path]
	if !ok {
		t.Fatalf("Module path %s missing from environment", module.Path)
	}

	// Check go.mod exists
	goModPath := filepath.Join(modulePath, "go.mod")
	if !env.FileExists(module.Path, "go.mod") {
		t.Errorf("go.mod file not found at %s", goModPath)
	}

	// Check each source file exists
	for _, pkg := range module.Packages {
		for _, file := range pkg.Files {
			if file.Path == "" {
				continue // Skip files without paths
			}

			relativePath := strings.TrimPrefix(file.Path, module.Dir)
			relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))

			if !env.FileExists(module.Path, relativePath) {
				t.Errorf("File %s not found in materialized module", relativePath)
			}
		}
	}
}
