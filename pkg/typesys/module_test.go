package typesys

import (
	"testing"
)

func TestModuleCreation(t *testing.T) {
	// Create a new module
	module := NewModule("/test/module")

	if module.Dir != "/test/module" {
		t.Errorf("Module.Dir = %q, want %q", module.Dir, "/test/module")
	}

	if module.Path != "" {
		t.Errorf("Module.Path should be empty initially, got %q", module.Path)
	}

	if module.GoVersion != "" {
		t.Errorf("Module.GoVersion should be empty initially, got %q", module.GoVersion)
	}

	if module.FileSet == nil {
		t.Errorf("Module.FileSet should be initialized")
	}

	if len(module.Packages) != 0 {
		t.Errorf("New module should have no packages, got %d", len(module.Packages))
	}

	if module.pkgCache == nil {
		t.Errorf("Module.pkgCache should be initialized")
	}
}

func TestModuleSetPath(t *testing.T) {
	module := NewModule("/test/module")

	// Set the path
	module.Path = "github.com/example/testmodule"

	if module.Path != "github.com/example/testmodule" {
		t.Errorf("Module.Path = %q, want %q", module.Path, "github.com/example/testmodule")
	}
}

func TestModuleAddPackage(t *testing.T) {
	module := NewModule("/test/module")

	// Create a package
	pkg := NewPackage(module, "testpkg", "github.com/example/testmodule/testpkg")

	// Add the package to the module
	module.Packages[pkg.ImportPath] = pkg

	// Verify the package was added
	if len(module.Packages) != 1 {
		t.Errorf("Module should have 1 package, got %d", len(module.Packages))
	}

	if module.Packages["github.com/example/testmodule/testpkg"] != pkg {
		t.Errorf("Package not correctly added to module")
	}

	// Verify the module reference in the package
	if pkg.Module != module {
		t.Errorf("Package.Module not set to the module")
	}
}

func TestModuleFileSet(t *testing.T) {
	module := NewModule("/test/module")

	// The FileSet should be initialized
	if module.FileSet == nil {
		t.Errorf("Module.FileSet should be initialized")
	}

	// Check that it's a valid token.FileSet
	pos := module.FileSet.AddFile("test.go", -1, 100).Pos(0)
	if !pos.IsValid() {
		t.Errorf("FileSet should create valid token.Pos values")
	}

	// Get the position back
	position := module.FileSet.Position(pos)
	if position.Filename != "test.go" {
		t.Errorf("FileSet position filename = %q, want %q", position.Filename, "test.go")
	}
}
