package resolve

import (
	"path/filepath"
	"testing"
)

func TestStandardModuleRegistry(t *testing.T) {
	registry := NewStandardModuleRegistry()

	testPath := "/path/to/module"
	normalizedTestPath := filepath.Clean(testPath)

	// Test registering a module
	err := registry.RegisterModule("github.com/test/module", testPath, true)
	if err != nil {
		t.Errorf("Failed to register module: %v", err)
	}

	// Test finding a module by import path
	module, ok := registry.FindModule("github.com/test/module")
	if !ok {
		t.Error("Failed to find module by import path")
	} else if module.FilesystemPath != normalizedTestPath {
		t.Errorf("Expected path %s, got %s", normalizedTestPath, module.FilesystemPath)
	}

	// Test finding a module by filesystem path
	module, ok = registry.FindByPath(testPath)
	if !ok {
		t.Error("Failed to find module by filesystem path")
	} else if module.ImportPath != "github.com/test/module" {
		t.Errorf("Expected import path %s, got %s", "github.com/test/module", module.ImportPath)
	}

	// Test registering a duplicate with same path (should succeed)
	err = registry.RegisterModule("github.com/test/module", testPath, true)
	if err != nil {
		t.Errorf("Failed to register duplicate module with same path: %v", err)
	}

	// Test registering a duplicate with different path (should fail)
	err = registry.RegisterModule("github.com/test/module", "/different/path", true)
	if err != ErrModuleAlreadyRegistered {
		t.Errorf("Expected ErrModuleAlreadyRegistered, got %v", err)
	}

	// Test listing modules
	modules := registry.ListModules()
	if len(modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(modules))
	}
}
