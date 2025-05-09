package html

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestNewHTMLVisualizer(t *testing.T) {
	v := NewHTMLVisualizer()
	if v == nil {
		t.Fatal("NewHTMLVisualizer returned nil")
	}

	if v.template == nil {
		t.Fatal("HTMLVisualizer has nil template")
	}
}

func TestFormat(t *testing.T) {
	v := NewHTMLVisualizer()
	if v.Format() != "html" {
		t.Errorf("Expected format 'html', got '%s'", v.Format())
	}
}

func TestSupportsTypeAnnotations(t *testing.T) {
	v := NewHTMLVisualizer()
	if !v.SupportsTypeAnnotations() {
		t.Error("HTMLVisualizer should support type annotations")
	}
}

// createTestModule creates a simple test module structure
func createTestModule() *typesys.Module {
	// Create a module
	mod := &typesys.Module{
		Path:      "example.com/test",
		GoVersion: "1.18",
		Packages:  make(map[string]*typesys.Package),
	}

	// Add a package to the module
	pkg := &typesys.Package{
		Module:     mod,
		Name:       "test",
		ImportPath: "example.com/test",
		Symbols:    make(map[string]*typesys.Symbol),
		Files:      make(map[string]*typesys.File),
	}
	mod.Packages[pkg.ImportPath] = pkg

	// Add a file to the package
	file := &typesys.File{
		Package: pkg,
		Path:    "test.go",
		Symbols: []*typesys.Symbol{},
	}
	pkg.Files[file.Path] = file

	// Add a function to the file
	fn := &typesys.Symbol{
		ID:       "test.MyFunc",
		Package:  pkg,
		File:     file,
		Name:     "MyFunc",
		Kind:     typesys.KindFunction,
		Exported: true,
	}
	file.Symbols = append(file.Symbols, fn)

	// Connect the symbol to the package as well
	pkg.Symbols[fn.ID] = fn

	return mod
}

func TestVisualize(t *testing.T) {
	// Create a test module
	mod := createTestModule()

	// Visualize the module
	v := NewHTMLVisualizer()
	result, err := v.Visualize(mod, nil)

	if err != nil {
		t.Fatalf("Visualize returned error: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Visualize returned empty result")
	}

	// Convert result to string for easier assertions
	html := string(result)

	// Check for expected content
	expectedItems := []string{
		"Module Path:", "example.com/test",
		"Go Version:", "1.18",
		"Package test",
		"MyFunc",
		"tag-exported",
	}

	for _, item := range expectedItems {
		if !strings.Contains(html, item) {
			t.Errorf("Expected HTML to contain '%s', but it doesn't", item)
		}
	}
}

func TestVisualizeWithOptions(t *testing.T) {
	// Create a test module
	mod := createTestModule()

	// Add a private function
	pkg := mod.Packages["example.com/test"]
	file := pkg.Files["test.go"]

	privateFn := &typesys.Symbol{
		ID:       "test.myPrivateFunc",
		Package:  pkg,
		File:     file,
		Name:     "myPrivateFunc",
		Kind:     typesys.KindFunction,
		Exported: false,
	}
	file.Symbols = append(file.Symbols, privateFn)
	pkg.Symbols[privateFn.ID] = privateFn

	// Test with different options
	tests := []struct {
		name             string
		options          *VisualizationOptions
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "With custom title",
			options: &VisualizationOptions{
				Title: "Custom Title",
			},
			shouldContain: []string{"Custom Title"},
		},
		{
			name: "Include private = false",
			options: &VisualizationOptions{
				IncludePrivate: false,
			},
			shouldContain:    []string{"MyFunc"},
			shouldNotContain: []string{"myPrivateFunc"},
		},
		{
			name: "Include private = true",
			options: &VisualizationOptions{
				IncludePrivate: true,
			},
			shouldContain: []string{"MyFunc", "myPrivateFunc"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := NewHTMLVisualizer()
			result, err := v.Visualize(mod, tc.options)

			if err != nil {
				t.Fatalf("Visualize returned error: %v", err)
			}

			html := string(result)

			// Check for expected content
			for _, item := range tc.shouldContain {
				if !strings.Contains(html, item) {
					t.Errorf("Expected HTML to contain '%s', but it doesn't", item)
				}
			}

			// Check for content that should not be present
			for _, item := range tc.shouldNotContain {
				if strings.Contains(html, item) {
					t.Errorf("HTML should not contain '%s', but it does", item)
				}
			}
		})
	}
}
