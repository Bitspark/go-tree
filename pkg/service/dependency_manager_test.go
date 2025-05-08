package service

import (
	"os"
	"path/filepath"
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

	// Create test module directories and go.mod files
	mainModDir := filepath.Join(tempDir, "main")
	dep1ModDir := filepath.Join(tempDir, "dep1")
	dep2ModDir := filepath.Join(tempDir, "dep2")

	createTestModule(t, mainModDir, "example.com/main", []string{
		"example.com/dep1 v1.0.0",
		"example.com/dep2 v1.0.0",
	})

	createTestModule(t, dep1ModDir, "example.com/dep1", []string{
		"example.com/dep2 v1.0.0",
	})

	createTestModule(t, dep2ModDir, "example.com/dep2", nil)

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
