package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

func TestParseGoMod(t *testing.T) {
	tests := []struct {
		name                 string
		content              string
		expectedDeps         map[string]string
		expectedReplacements map[string]string
	}{
		{
			name: "simple dependencies",
			content: `module example.com/mymodule

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)
`,
			expectedDeps: map[string]string{
				"github.com/pkg/errors":       "v0.9.1",
				"github.com/stretchr/testify": "v1.7.0",
			},
			expectedReplacements: map[string]string{},
		},
		{
			name: "with local replacements",
			content: `module example.com/mymodule

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/pkg/errors => ./local/errors
`,
			expectedDeps: map[string]string{
				"github.com/pkg/errors":       "v0.9.1",
				"github.com/stretchr/testify": "v1.7.0",
			},
			expectedReplacements: map[string]string{
				"github.com/pkg/errors": "./local/errors",
			},
		},
		{
			name: "with remote replacements",
			content: `module example.com/mymodule

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/pkg/errors => github.com/my/errors v0.8.0
`,
			expectedDeps: map[string]string{
				"github.com/pkg/errors":       "v0.9.1",
				"github.com/stretchr/testify": "v1.7.0",
			},
			expectedReplacements: map[string]string{
				"github.com/pkg/errors": "github.com/my/errors",
			},
		},
		{
			name: "with mixed replacements",
			content: `module example.com/mymodule

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/pkg/errors => github.com/my/errors v0.8.0
	github.com/stretchr/testify => ../testify
)
`,
			expectedDeps: map[string]string{
				"github.com/pkg/errors":       "v0.9.1",
				"github.com/stretchr/testify": "v1.7.0",
			},
			expectedReplacements: map[string]string{
				"github.com/pkg/errors":       "github.com/my/errors",
				"github.com/stretchr/testify": "../testify",
			},
		},
		{
			name: "without v prefix",
			content: `module example.com/mymodule

go 1.16

require (
	github.com/pkg/errors 0.9.1
	github.com/stretchr/testify 1.7.0
)
`,
			expectedDeps: map[string]string{
				"github.com/pkg/errors":       "v0.9.1",
				"github.com/stretchr/testify": "v1.7.0",
			},
			expectedReplacements: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, replacements, err := parseGoMod(tt.content)
			if err != nil {
				t.Fatalf("parseGoMod() error = %v", err)
			}

			// Check dependencies
			if len(deps) != len(tt.expectedDeps) {
				t.Errorf("parseGoMod() got %d deps, want %d", len(deps), len(tt.expectedDeps))
			}

			for path, version := range tt.expectedDeps {
				if deps[path] != version {
					t.Errorf("parseGoMod() dep %s = %s, want %s", path, deps[path], version)
				}
			}

			// Check replacements
			if len(replacements) != len(tt.expectedReplacements) {
				t.Errorf("parseGoMod() got %d replacements, want %d",
					len(replacements), len(tt.expectedReplacements))
			}

			for path, replacement := range tt.expectedReplacements {
				if replacements[path] != replacement {
					t.Errorf("parseGoMod() replacement %s = %s, want %s",
						path, replacements[path], replacement)
				}
			}
		})
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	// Create a temporary directory for our test modules
	tempDir, err := os.MkdirTemp("", "go-tree-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Local helper function to create test modules
	createDirectModule := func(dir, modulePath string, deps []string) {
		// Create directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create module directory %s: %v", dir, err)
		}

		// Create go.mod content
		content := fmt.Sprintf("module %s\n\ngo 1.16\n", modulePath)

		// Add dependencies if specified
		if len(deps) > 0 {
			content += "\nrequire (\n"
			for _, dep := range deps {
				content += fmt.Sprintf("\t%s\n", dep)
			}
			content += ")\n"
		}

		// Write go.mod file
		goModPath := filepath.Join(dir, "go.mod")
		if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write go.mod file: %v", err)
		}
	}

	// Create test module directories and go.mod files
	mainModDir := filepath.Join(tempDir, "main")
	dep1ModDir := filepath.Join(tempDir, "dep1")
	dep2ModDir := filepath.Join(tempDir, "dep2")

	createDirectModule(mainModDir, "example.com/main", []string{
		"example.com/dep1 v1.0.0",
		"example.com/dep2 v1.0.0",
	})

	createDirectModule(dep1ModDir, "example.com/dep1", []string{
		"example.com/dep2 v1.0.0",
	})

	createDirectModule(dep2ModDir, "example.com/dep2", nil)

	// Create a mock service with mock modules
	mockMainModule := &typesys.Module{
		Path:     "example.com/main",
		Dir:      mainModDir,
		Packages: map[string]*typesys.Package{},
	}

	mockDep1Module := &typesys.Module{
		Path:     "example.com/dep1",
		Dir:      dep1ModDir,
		Packages: map[string]*typesys.Package{},
	}

	mockDep2Module := &typesys.Module{
		Path:     "example.com/dep2",
		Dir:      dep2ModDir,
		Packages: map[string]*typesys.Package{},
	}

	service := &Service{
		Modules: map[string]*typesys.Module{
			"example.com/main": mockMainModule,
			"example.com/dep1": mockDep1Module,
			"example.com/dep2": mockDep2Module,
		},
		MainModulePath: "example.com/main",
		Config:         &Config{}, // Initialize with empty config
	}

	// Test building the dependency graph
	depManager := NewDependencyManager(service)
	graph := depManager.BuildDependencyGraph()

	// Verify the graph
	if len(graph) != 3 {
		t.Errorf("Expected 3 modules in graph, got %d", len(graph))
	}

	// Check main module dependencies
	mainDeps := graph["example.com/main"]
	if len(mainDeps) != 2 {
		t.Errorf("Expected 2 dependencies for main module, got %d", len(mainDeps))
	}

	// Check dep1 module dependencies
	dep1Deps := graph["example.com/dep1"]
	if len(dep1Deps) != 1 {
		t.Errorf("Expected 1 dependency for dep1 module, got %d", len(dep1Deps))
	}
	if dep1Deps[0] != "example.com/dep2" {
		t.Errorf("Expected dep1 to depend on dep2, got %s", dep1Deps[0])
	}

	// Check dep2 module dependencies
	dep2Deps := graph["example.com/dep2"]
	if len(dep2Deps) != 0 {
		t.Errorf("Expected 0 dependencies for dep2 module, got %d", len(dep2Deps))
	}
}

func TestFindModuleByDir(t *testing.T) {
	// Create a simple service with mock modules
	service := &Service{
		Modules: map[string]*typesys.Module{
			"example.com/mod1": {
				Path: "example.com/mod1",
				Dir:  "/path/to/mod1",
			},
			"example.com/mod2": {
				Path: "example.com/mod2",
				Dir:  "/path/to/mod2",
			},
		},
	}

	depManager := NewDependencyManager(service)

	// Test finding an existing module
	mod, found := depManager.FindModuleByDir("/path/to/mod1")
	if !found {
		t.Errorf("Expected to find module at /path/to/mod1")
	}
	if mod == nil || mod.Path != "example.com/mod1" {
		t.Errorf("Found incorrect module: %v", mod)
	}

	// Test finding a non-existent module
	_, found = depManager.FindModuleByDir("/path/to/nonexistent")
	if found {
		t.Errorf("Expected not to find module at /path/to/nonexistent")
	}
}

// TestDependencyManagerDepth tests the configurable depth feature
func TestDependencyManagerDepth(t *testing.T) {
	// Set up test modules with known dependencies
	testDir := setupTestModules(t)
	defer os.RemoveAll(testDir)

	// Create service with depth 0 (only direct dependencies)
	serviceDepth0, err := NewService(&Config{
		ModuleDir:       filepath.Join(testDir, "main"),
		WithDeps:        true,
		DependencyDepth: 0,
		Verbose:         true,
	})
	if err != nil {
		t.Fatalf("Failed to create service with depth 0: %v", err)
	}

	// Should only have loaded main module and its direct dependencies
	if len(serviceDepth0.Modules) != 2 {
		t.Errorf("Expected 2 modules (main + dep1), got %d", len(serviceDepth0.Modules))
	}
	if _, ok := serviceDepth0.Modules["example.com/main"]; !ok {
		t.Errorf("Main module not loaded")
	}
	if _, ok := serviceDepth0.Modules["example.com/dep1"]; !ok {
		t.Errorf("Direct dependency not loaded")
	}
	if _, ok := serviceDepth0.Modules["example.com/dep2"]; ok {
		t.Errorf("Transitive dependency loaded despite depth=0")
	}

	// Create service with depth 1 (direct dependencies and their dependencies)
	serviceDepth1, err := NewService(&Config{
		ModuleDir:       filepath.Join(testDir, "main"),
		WithDeps:        true,
		DependencyDepth: 1,
		Verbose:         true,
	})
	if err != nil {
		t.Fatalf("Failed to create service with depth 1: %v", err)
	}

	// Should have loaded main module and all dependencies
	if len(serviceDepth1.Modules) != 3 {
		t.Errorf("Expected 3 modules (main + dep1 + dep2), got %d", len(serviceDepth1.Modules))
	}
	if _, ok := serviceDepth1.Modules["example.com/main"]; !ok {
		t.Errorf("Main module not loaded")
	}
	if _, ok := serviceDepth1.Modules["example.com/dep1"]; !ok {
		t.Errorf("Direct dependency not loaded")
	}
	if _, ok := serviceDepth1.Modules["example.com/dep2"]; !ok {
		t.Errorf("Transitive dependency not loaded despite depth=1")
	}
}

// TestCircularDependencyDetection tests that circular dependencies are properly detected
func TestCircularDependencyDetection(t *testing.T) {
	// Set up test modules with circular dependencies
	testDir := setupCircularTestModules(t)
	defer os.RemoveAll(testDir)

	// Create service
	service, err := NewService(&Config{
		ModuleDir:       filepath.Join(testDir, "main"),
		WithDeps:        true,
		DependencyDepth: 5, // Deep enough to detect circularity
		Verbose:         true,
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Should have loaded all modules despite the circular dependency
	if len(service.Modules) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(service.Modules))
	}

	// Check for specific modules
	if _, ok := service.Modules["example.com/main"]; !ok {
		t.Errorf("Main module not loaded")
	}
	if _, ok := service.Modules["example.com/dep1"]; !ok {
		t.Errorf("Dep1 module not loaded")
	}
	if _, ok := service.Modules["example.com/dep2"]; !ok {
		t.Errorf("Dep2 module not loaded")
	}
}

// TestDependencyCaching tests that dependency resolution caching works
func TestDependencyCaching(t *testing.T) {
	// Set up test modules
	testDir := setupTestModules(t)
	defer os.RemoveAll(testDir)

	// Create service
	service, err := NewService(&Config{
		ModuleDir:       filepath.Join(testDir, "main"),
		WithDeps:        true,
		DependencyDepth: 1,
		Verbose:         true,
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Get the dependency manager
	depManager := service.DependencyManager

	// First call should populate the cache
	startCacheSize := len(depManager.dirCache)

	// Call findDependencyDir to make sure it's cached
	dir, err := depManager.findDependencyDir("example.com/dep1", "v1.0.0")
	if err != nil {
		t.Fatalf("Failed to find dependency dir: %v", err)
	}
	if dir == "" {
		t.Fatalf("Empty dependency dir returned")
	}

	// Call it again, should use cache
	dir2, err := depManager.findDependencyDir("example.com/dep1", "v1.0.0")
	if err != nil {
		t.Fatalf("Failed to find dependency dir on second call: %v", err)
	}

	// Verify both calls returned the same directory
	if dir != dir2 {
		t.Errorf("Cache inconsistency: first call returned %s, second call returned %s", dir, dir2)
	}

	// Verify cache grew
	endCacheSize := len(depManager.dirCache)
	if endCacheSize <= startCacheSize {
		t.Errorf("Cache did not grow after dependency resolution: %d -> %d", startCacheSize, endCacheSize)
	}
}

// TestDependencyErrorReporting tests that dependency errors are properly reported
func TestDependencyErrorReporting(t *testing.T) {
	// Create a non-existent dependency error
	err := &DependencyError{
		ImportPath: "example.com/nonexistent",
		Version:    "v1.0.0",
		Module:     "example.com/main",
		Reason:     "could not locate dependency",
		Err:        os.ErrNotExist,
	}

	// Check error message
	errMsg := err.Error()
	if !strings.Contains(errMsg, "example.com/nonexistent") {
		t.Errorf("Error message missing import path: %s", errMsg)
	}
	if !strings.Contains(errMsg, "v1.0.0") {
		t.Errorf("Error message missing version: %s", errMsg)
	}
	if !strings.Contains(errMsg, "example.com/main") {
		t.Errorf("Error message missing module: %s", errMsg)
	}
	if !strings.Contains(errMsg, "could not locate dependency") {
		t.Errorf("Error message missing reason: %s", errMsg)
	}
	if !strings.Contains(errMsg, os.ErrNotExist.Error()) {
		t.Errorf("Error message missing underlying error: %s", errMsg)
	}
}

// Helper function to set up test modules
func setupTestModules(t *testing.T) string {
	// Create temporary directory
	testDir, err := os.MkdirTemp("", "deptest")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create main module
	mainDir := filepath.Join(testDir, "main")
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main module directory: %v", err)
	}

	// Create go.mod for main module
	mainGoMod := filepath.Join(mainDir, "go.mod")
	mainGoModContent := `module example.com/main

go 1.20

require example.com/dep1 v1.0.0

replace example.com/dep1 => ../dep1
replace example.com/dep2 => ../dep2
`
	if err := os.WriteFile(mainGoMod, []byte(mainGoModContent), 0644); err != nil {
		t.Fatalf("Failed to create main go.mod: %v", err)
	}

	// Create main.go
	mainGo := filepath.Join(mainDir, "main.go")
	mainGoContent := `package main

import "example.com/dep1"

func main() {
	dep1.Func()
}
`
	if err := os.WriteFile(mainGo, []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create dep1 module
	dep1Dir := filepath.Join(testDir, "dep1")
	if err := os.Mkdir(dep1Dir, 0755); err != nil {
		t.Fatalf("Failed to create dep1 module directory: %v", err)
	}

	// Create go.mod for dep1 module
	dep1GoMod := filepath.Join(dep1Dir, "go.mod")
	dep1GoModContent := `module example.com/dep1

go 1.20

require example.com/dep2 v1.0.0
`
	if err := os.WriteFile(dep1GoMod, []byte(dep1GoModContent), 0644); err != nil {
		t.Fatalf("Failed to create dep1 go.mod: %v", err)
	}

	// Create dep1.go
	dep1Go := filepath.Join(dep1Dir, "dep1.go")
	dep1GoContent := `package dep1

import "example.com/dep2"

func Func() {
	dep2.Func()
}
`
	if err := os.WriteFile(dep1Go, []byte(dep1GoContent), 0644); err != nil {
		t.Fatalf("Failed to create dep1.go: %v", err)
	}

	// Create dep2 module
	dep2Dir := filepath.Join(testDir, "dep2")
	if err := os.Mkdir(dep2Dir, 0755); err != nil {
		t.Fatalf("Failed to create dep2 module directory: %v", err)
	}

	// Create go.mod for dep2 module
	dep2GoMod := filepath.Join(dep2Dir, "go.mod")
	dep2GoModContent := `module example.com/dep2

go 1.20
`
	if err := os.WriteFile(dep2GoMod, []byte(dep2GoModContent), 0644); err != nil {
		t.Fatalf("Failed to create dep2 go.mod: %v", err)
	}

	// Create dep2.go
	dep2Go := filepath.Join(dep2Dir, "dep2.go")
	dep2GoContent := `package dep2

func Func() {
	// Empty function
}
`
	if err := os.WriteFile(dep2Go, []byte(dep2GoContent), 0644); err != nil {
		t.Fatalf("Failed to create dep2.go: %v", err)
	}

	return testDir
}

// Helper function to set up circular test modules
func setupCircularTestModules(t *testing.T) string {
	// Create temporary directory
	testDir, err := os.MkdirTemp("", "circulardeptest")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create main module
	mainDir := filepath.Join(testDir, "main")
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main module directory: %v", err)
	}

	// Create go.mod for main module
	mainGoMod := filepath.Join(mainDir, "go.mod")
	mainGoModContent := `module example.com/main

go 1.20

require example.com/dep1 v1.0.0

replace example.com/dep1 => ../dep1
replace example.com/dep2 => ../dep2
`
	if err := os.WriteFile(mainGoMod, []byte(mainGoModContent), 0644); err != nil {
		t.Fatalf("Failed to create main go.mod: %v", err)
	}

	// Create main.go
	mainGo := filepath.Join(mainDir, "main.go")
	mainGoContent := `package main

import "example.com/dep1"

func main() {
	dep1.Func()
}
`
	if err := os.WriteFile(mainGo, []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create dep1 module
	dep1Dir := filepath.Join(testDir, "dep1")
	if err := os.Mkdir(dep1Dir, 0755); err != nil {
		t.Fatalf("Failed to create dep1 module directory: %v", err)
	}

	// Create go.mod for dep1 module with circular dependency to dep2
	dep1GoMod := filepath.Join(dep1Dir, "go.mod")
	dep1GoModContent := `module example.com/dep1

go 1.20

require example.com/dep2 v1.0.0
`
	if err := os.WriteFile(dep1GoMod, []byte(dep1GoModContent), 0644); err != nil {
		t.Fatalf("Failed to create dep1 go.mod: %v", err)
	}

	// Create dep1.go
	dep1Go := filepath.Join(dep1Dir, "dep1.go")
	dep1GoContent := `package dep1

import "example.com/dep2"

func Func() {
	dep2.Func()
}
`
	if err := os.WriteFile(dep1Go, []byte(dep1GoContent), 0644); err != nil {
		t.Fatalf("Failed to create dep1.go: %v", err)
	}

	// Create dep2 module
	dep2Dir := filepath.Join(testDir, "dep2")
	if err := os.Mkdir(dep2Dir, 0755); err != nil {
		t.Fatalf("Failed to create dep2 module directory: %v", err)
	}

	// Create go.mod for dep2 module with circular dependency back to dep1
	dep2GoMod := filepath.Join(dep2Dir, "go.mod")
	dep2GoModContent := `module example.com/dep2

go 1.20
`
	if err := os.WriteFile(dep2GoMod, []byte(dep2GoModContent), 0644); err != nil {
		t.Fatalf("Failed to create dep2 go.mod: %v", err)
	}

	return testDir
}
