package service

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// MockSymbol creates a mock Symbol for testing
func mockSymbol(id, name string, kind typesys.SymbolKind) *typesys.Symbol {
	return &typesys.Symbol{
		ID:   id,
		Name: name,
		Kind: kind,
	}
}

// MockPackage creates a mock Package for testing
func mockPackage(importPath string) *typesys.Package {
	return &typesys.Package{
		ImportPath: importPath,
		Symbols:    make(map[string]*typesys.Symbol),
	}
}

// TestResolveImport tests cross-module package resolution
func TestResolveImport(t *testing.T) {
	// Create a service with mock modules
	service := &Service{
		Modules: map[string]*typesys.Module{
			"mod1": {
				Path: "mod1",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {ImportPath: "pkg/foo"},
					"pkg/bar": {ImportPath: "pkg/bar"},
				},
			},
			"mod2": {
				Path: "mod2",
				Packages: map[string]*typesys.Package{
					"pkg/baz": {ImportPath: "pkg/baz"},
				},
			},
		},
		MainModulePath: "mod1",
	}

	// Test resolving from mod1 to mod1
	pkg, err := service.ResolveImport("pkg/foo", "mod1")
	if err != nil {
		t.Errorf("ResolveImport() error = %v", err)
	}
	if pkg.ImportPath != "pkg/foo" {
		t.Errorf("ResolveImport() got %s, want pkg/foo", pkg.ImportPath)
	}

	// Test resolving from mod2 to mod1
	pkg, err = service.ResolveImport("pkg/bar", "mod2")
	if err != nil {
		t.Errorf("ResolveImport() error = %v", err)
	}
	if pkg.ImportPath != "pkg/bar" {
		t.Errorf("ResolveImport() got %s, want pkg/bar", pkg.ImportPath)
	}

	// Test resolving a non-existent package
	_, err = service.ResolveImport("pkg/nonexistent", "mod1")
	if err == nil {
		t.Errorf("ResolveImport() expected error for non-existent package")
	}
}

// TestAvailableModules tests the AvailableModules function
func TestAvailableModules(t *testing.T) {
	service := &Service{
		Modules: map[string]*typesys.Module{
			"mod1": {Path: "mod1"},
			"mod2": {Path: "mod2"},
			"mod3": {Path: "mod3"},
		},
	}

	modules := service.AvailableModules()
	if len(modules) != 3 {
		t.Errorf("AvailableModules() got %d modules, want 3", len(modules))
	}

	// Check all modules are included
	modulesSet := make(map[string]bool)
	for _, m := range modules {
		modulesSet[m] = true
	}

	if !modulesSet["mod1"] || !modulesSet["mod2"] || !modulesSet["mod3"] {
		t.Errorf("AvailableModules() missing some modules")
	}
}

// TestResolvePackage tests package resolution with versioning
func TestResolvePackage(t *testing.T) {
	// Create a service with mocked package versions
	service := &Service{
		Modules: map[string]*typesys.Module{
			"mod1": {
				Path: "mod1",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {ImportPath: "pkg/foo"},
				},
			},
		},
		PackageVersions: make(map[string]map[string]*ModulePackage),
	}

	// Add versioned packages
	service.PackageVersions["pkg/bar"] = map[string]*ModulePackage{
		"v1.0.0": {
			Module:     service.Modules["mod1"],
			Package:    &typesys.Package{ImportPath: "pkg/bar"},
			ImportPath: "pkg/bar",
			Version:    "v1.0.0",
		},
		"v2.0.0": {
			Module:     service.Modules["mod1"],
			Package:    &typesys.Package{ImportPath: "pkg/bar"},
			ImportPath: "pkg/bar",
			Version:    "v2.0.0",
		},
	}

	// Test resolving a non-versioned package
	pkg, err := service.ResolvePackage("pkg/foo", "")
	if err != nil {
		t.Errorf("ResolvePackage() error = %v", err)
	}
	if pkg.Package.ImportPath != "pkg/foo" {
		t.Errorf("ResolvePackage() got wrong package: %s", pkg.Package.ImportPath)
	}

	// Test resolving a versioned package with preferred version
	pkg, err = service.ResolvePackage("pkg/bar", "v1.0.0")
	if err != nil {
		t.Errorf("ResolvePackage() error = %v", err)
	}
	if pkg.Version != "v1.0.0" {
		t.Errorf("ResolvePackage() got version %s, want v1.0.0", pkg.Version)
	}

	// Test resolving a versioned package with no preferred version
	pkg, err = service.ResolvePackage("pkg/bar", "")
	if err != nil {
		t.Errorf("ResolvePackage() error = %v", err)
	}
	if pkg.Version == "" {
		t.Errorf("ResolvePackage() got empty version")
	}

	// Test resolving a non-existent package
	_, err = service.ResolvePackage("pkg/nonexistent", "")
	if err == nil {
		t.Errorf("ResolvePackage() expected error for non-existent package")
	}
}

// TestFindTypeAcrossModules tests finding types across modules
func TestFindTypeAcrossModules(t *testing.T) {
	service := &Service{
		Modules: map[string]*typesys.Module{
			"mod1": {
				Path: "mod1",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {
						ImportPath: "pkg/foo",
						Symbols: map[string]*typesys.Symbol{
							"sym1": mockSymbol("sym1", "MyType", typesys.KindStruct),
						},
					},
				},
			},
			"mod2": {
				Path: "mod2",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {
						ImportPath: "pkg/foo",
						Symbols: map[string]*typesys.Symbol{
							"sym2": mockSymbol("sym2", "MyType", typesys.KindStruct),
						},
					},
				},
			},
		},
	}

	// Test finding a type across modules
	typeVersions := service.FindTypeAcrossModules("pkg/foo", "MyType")
	if len(typeVersions) != 2 {
		t.Errorf("FindTypeAcrossModules() got %d versions, want 2", len(typeVersions))
	}

	if typeVersions["mod1"] == nil || typeVersions["mod2"] == nil {
		t.Errorf("FindTypeAcrossModules() missing versions from some modules")
	}
}

// Helper function to create a test module with a go.mod file
func createTestModule(t *testing.T, dir string, modPath string, deps []string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create module directory %s: %v", dir, err)
	}

	// Create go.mod content
	content := "module " + modPath + "\n\ngo 1.16\n\n"

	// Add dependencies if any
	if len(deps) > 0 {
		content += "require (\n"
		for _, dep := range deps {
			content += "\t" + dep + "\n"
		}
		content += ")\n"
	}

	// Write go.mod file
	goModPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write go.mod file: %v", err)
	}
}
