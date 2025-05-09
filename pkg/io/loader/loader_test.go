package loader

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestModuleLoading tests the basic module loading functionality
func TestModuleLoading(t *testing.T) {
	// Get the project root
	moduleDir, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Loading module from: %s", moduleDir)

	// Verify go.mod exists
	goModPath := filepath.Join(moduleDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatalf("go.mod not found at %s", goModPath)
	} else {
		t.Logf("Found go.mod at %s", goModPath)
	}

	// Load with default options
	module, err := LoadModule(moduleDir, nil)
	if err != nil {
		t.Fatalf("Failed to load module with default options: %v", err)
	}

	// Check module info
	t.Logf("Module path: %s", module.Path)
	t.Logf("Go version: %s", module.GoVersion)
	t.Logf("Loaded %d packages", len(module.Packages))

	if len(module.Packages) == 0 {
		t.Errorf("No packages loaded - this is the root issue!")
	}

	// Try with explicit options
	loadOpts := &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
		Trace:          true,
	}

	verboseModule, err := LoadModule(moduleDir, loadOpts)
	if err != nil {
		t.Fatalf("Failed to load module with verbose options: %v", err)
	}

	t.Logf("With verbose options: loaded %d packages", len(verboseModule.Packages))
}

// TestPackageLoading tests the package loading step specifically
func TestPackageLoading(t *testing.T) {
	// Get the project root
	moduleDir, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create module without loading packages
	module := typesys.NewModule(moduleDir)

	// Try to load packages directly
	opts := &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
		Trace:          true,
	}

	err = loadPackages(module, opts)
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	t.Logf("Loaded %d packages", len(module.Packages))

	if len(module.Packages) == 0 {
		// Let's inspect the directory structure to see what's there
		files, err := os.ReadDir(moduleDir)
		if err != nil {
			t.Logf("Error reading directory: %v", err)
		} else {
			t.Logf("Directory contents:")
			for _, file := range files {
				t.Logf("- %s (dir: %t)", file.Name(), file.IsDir())
			}
		}

		// Check a specific package we know should be there
		pkgDir := filepath.Join(moduleDir, "pkg", "typesys")
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			t.Errorf("typesys package directory not found at %s", pkgDir)
		} else {
			t.Logf("Found typesys directory at %s", pkgDir)

			// Check for Go files
			goFiles, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
			if err != nil {
				t.Logf("Error finding Go files: %v", err)
			} else {
				t.Logf("Go files in typesys package:")
				for _, file := range goFiles {
					t.Logf("- %s", filepath.Base(file))
				}
			}
		}
	}
}

// TestPackagesLoadDetails tests the detailed behavior of packages loading
func TestPackagesLoadDetails(t *testing.T) {
	// Get the project root
	moduleDir, err := filepath.Abs("../../../")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Test direct go/packages loading to see if that works
	t.Log("Testing direct use of golang.org/x/tools/go/packages")

	// Let's look at pkg/typesys specifically
	pkgPath := filepath.Join(moduleDir, "pkg", "typesys")
	basicTest(t, pkgPath)

	// Let's also try the whole project with ./...
	t.Log("\nTesting with ./... pattern")
	basicTest(t, moduleDir)
}

// Helper to test basic package loading
func basicTest(t *testing.T, dir string) {
	t.Logf("Testing in directory: %s", dir)

	// Use the direct package loading to diagnose
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedSyntax,
		Dir:   dir,
		Tests: true,
	}

	// Try with different patterns
	patterns := []string{
		".",     // current directory only
		"./...", // recursively
	}

	for _, pattern := range patterns {
		t.Logf("Loading with pattern: %s", pattern)
		pkgs, err := packages.Load(cfg, pattern)
		if err != nil {
			t.Errorf("Failed to load packages with pattern %s: %v", pattern, err)
			continue
		}

		t.Logf("Loaded %d packages with pattern %s", len(pkgs), pattern)

		// Count packages without errors
		validPkgs := 0
		for _, pkg := range pkgs {
			if len(pkg.Errors) == 0 {
				validPkgs++
			} else {
				t.Logf("Package %s has errors:", pkg.PkgPath)
				for _, err := range pkg.Errors {
					t.Logf("  - %v", err)
				}
			}
		}

		t.Logf("Valid packages (no errors): %d", validPkgs)

		// Check first few packages
		for i, pkg := range pkgs {
			if i >= 3 {
				t.Logf("... and %d more packages", len(pkgs)-i)
				break
			}

			t.Logf("Package[%d]: %s with %d files", i, pkg.PkgPath, len(pkg.CompiledGoFiles))
		}
	}
}

// TestGoModAndPathDetection specifically tests the go.mod detection logic
func TestGoModAndPathDetection(t *testing.T) {
	// Get the project root
	moduleDir, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Check go.mod exists explicitly
	goModPath := filepath.Join(moduleDir, "go.mod")
	if info, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatalf("go.mod not found at %s", goModPath)
	} else {
		t.Logf("Found go.mod at %s (size: %d bytes)", goModPath, info.Size())

		// Read and log go.mod content to verify it's correct
		content, err := os.ReadFile(goModPath)
		if err != nil {
			t.Errorf("Failed to read go.mod: %v", err)
		} else {
			t.Logf("go.mod content:\n%s", string(content))
		}
	}

	// Check the pattern used for packages.Load
	t.Log("Checking if directory can be properly loaded as a Go module")

	// Create a module without loading packages
	module := typesys.NewModule(moduleDir)

	// Extract module info
	if err := extractModuleInfo(module); err != nil {
		t.Errorf("Error extracting module info: %v", err)
	} else {
		t.Logf("Extracted module path: %s", module.Path)
		t.Logf("Extracted Go version: %s", module.GoVersion)
	}

	// Test directory structure and Go file presence
	pkgDir := filepath.Join(moduleDir, "pkg")
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		t.Errorf("pkg directory not found at %s", pkgDir)
	} else {
		subdirs, err := os.ReadDir(pkgDir)
		if err != nil {
			t.Errorf("Failed to read pkg subdirectories: %v", err)
		} else {
			t.Logf("Found %d subdirectories in pkg/", len(subdirs))
			for _, subdir := range subdirs {
				if subdir.IsDir() {
					t.Logf("- %s", subdir.Name())

					// Check for Go files in this package
					pkgPath := filepath.Join(pkgDir, subdir.Name())
					goFiles, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
					if err != nil {
						t.Logf("  Error finding Go files: %v", err)
					} else {
						t.Logf("  Found %d Go files", len(goFiles))
						for _, file := range goFiles[:minInt(3, len(goFiles))] {
							t.Logf("  - %s", filepath.Base(file))
						}
						if len(goFiles) > 3 {
							t.Logf("  - ... and %d more", len(goFiles)-3)
						}
					}
				}
			}
		}
	}
}

// Helper for min of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
