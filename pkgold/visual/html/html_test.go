package html

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkgold/core/module"
)

func TestHTMLVisualizer_Visualize(t *testing.T) {
	// Create a test module
	mod := createTestModule()

	// Create visualizer with default options
	visualizer := NewHTMLVisualizer(DefaultOptions())

	// Generate HTML
	html, err := visualizer.Visualize(mod)
	if err != nil {
		t.Fatalf("Visualize failed: %v", err)
	}

	// Basic checks
	htmlStr := string(html)

	// Check module path
	if !strings.Contains(htmlStr, "example.com/testmodule") {
		t.Error("HTML output doesn't contain module path")
	}

	// Check package name
	if !strings.Contains(htmlStr, "Package main") {
		t.Error("HTML output doesn't contain package name")
	}

	// Check function
	if !strings.Contains(htmlStr, "ExportedFunc") {
		t.Error("HTML output doesn't contain exported function")
	}

	// Check type
	if !strings.Contains(htmlStr, "TestStruct") {
		t.Error("HTML output doesn't contain type definition")
	}
}

func TestHTMLVisualizer_CustomTitle(t *testing.T) {
	// Create a test module
	mod := createTestModule()

	// Create visualizer with custom title
	options := DefaultOptions()
	options.Title = "Custom Module Documentation"
	visualizer := NewHTMLVisualizer(options)

	// Generate HTML
	html, err := visualizer.Visualize(mod)
	if err != nil {
		t.Fatalf("Visualize failed: %v", err)
	}

	// Check title
	htmlStr := string(html)
	if !strings.Contains(htmlStr, "Custom Module Documentation") {
		t.Error("HTML output doesn't contain custom title")
	}
}

func TestHTMLVisualizer_PrivateElements(t *testing.T) {
	// Create a test module
	mod := createTestModule()

	// Test with private elements hidden (default)
	defaultVisualizer := NewHTMLVisualizer(DefaultOptions())
	defaultHTML, err := defaultVisualizer.Visualize(mod)
	if err != nil {
		t.Fatalf("Visualize failed: %v", err)
	}

	// Private elements should not be visible
	if strings.Contains(string(defaultHTML), "privateFunc") {
		t.Error("Private function should not be visible with default options")
	}

	// Test with private elements shown
	options := DefaultOptions()
	options.IncludePrivate = true
	includePrivateVisualizer := NewHTMLVisualizer(options)
	includePrivateHTML, err := includePrivateVisualizer.Visualize(mod)
	if err != nil {
		t.Fatalf("Visualize failed: %v", err)
	}

	// Private elements should be visible
	if !strings.Contains(string(includePrivateHTML), "privateFunc") {
		t.Error("Private function should be visible when IncludePrivate is true")
	}
}

// createTestModule creates a test module for use in tests
func createTestModule() *module.Module {
	// Create a module
	mod := &module.Module{
		Path:      "example.com/testmodule",
		GoVersion: "1.18",
		Dir:       "/path/to/module",
		Packages:  make(map[string]*module.Package),
	}

	// Create a package
	pkg := &module.Package{
		Name:          "main",
		ImportPath:    "example.com/testmodule",
		Module:        mod,
		Documentation: "Package main is a test package.",
		Files:         make(map[string]*module.File),
		Types:         make(map[string]*module.Type),
		Functions:     make(map[string]*module.Function),
		Variables:     make(map[string]*module.Variable),
		Constants:     make(map[string]*module.Constant),
	}

	// Add the package to the module
	mod.Packages["example.com/testmodule"] = pkg
	mod.MainPackage = pkg

	// Create a file
	file := &module.File{
		Path:    "/path/to/module/main.go",
		Name:    "main.go",
		Package: pkg,
	}
	pkg.Files["main.go"] = file

	// Create a type
	structType := &module.Type{
		Name:       "TestStruct",
		File:       file,
		Package:    pkg,
		Kind:       "struct",
		IsExported: true,
		Doc:        "TestStruct is a test struct.",
		Fields: []*module.Field{
			{
				Name:   "Field1",
				Type:   "string",
				Tag:    `json:"field1"`,
				Doc:    "Field1 is a string field.",
				Parent: nil,
			},
			{
				Name:   "field2",
				Type:   "int",
				Tag:    `json:"field2"`,
				Doc:    "field2 is a private int field.",
				Parent: nil,
			},
		},
	}
	pkg.Types["TestStruct"] = structType

	// Create exported function
	exportedFunc := &module.Function{
		Name:       "ExportedFunc",
		File:       file,
		Package:    pkg,
		Signature:  "func ExportedFunc(arg string) error",
		IsExported: true,
		Doc:        "ExportedFunc is an exported function.",
	}
	pkg.Functions["ExportedFunc"] = exportedFunc

	// Create private function
	privateFunc := &module.Function{
		Name:       "privateFunc",
		File:       file,
		Package:    pkg,
		Signature:  "func privateFunc(arg int) bool",
		IsExported: false,
		Doc:        "privateFunc is a private function.",
	}
	pkg.Functions["privateFunc"] = privateFunc

	// Create a constant
	constant := &module.Constant{
		Name:       "VERSION",
		File:       file,
		Package:    pkg,
		Type:       "string",
		Value:      `"1.0.0"`,
		IsExported: true,
		Doc:        "VERSION is the version constant.",
	}
	pkg.Constants["VERSION"] = constant

	// Create a variable
	variable := &module.Variable{
		Name:       "config",
		File:       file,
		Package:    pkg,
		Type:       "map[string]string",
		IsExported: false,
		Doc:        "config is a private variable.",
	}
	pkg.Variables["config"] = variable

	return mod
}
