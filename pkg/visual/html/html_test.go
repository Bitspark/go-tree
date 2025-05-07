package html

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
	"bitspark.dev/go-tree/pkg/core/parse"
)

// TestGenerate tests the basic HTML generation functionality
func TestGenerate(t *testing.T) {
	// Create a simple model package for testing
	pkg := &model.GoPackage{
		Name:       "testpkg",
		PackageDoc: "This is a test package",
		Imports: []model.GoImport{
			{Path: "fmt", Alias: ""},
			{Path: "os", Alias: ""},
		},
		Types: []model.GoType{
			{
				Name: "Person",
				Kind: "struct",
				Fields: []model.GoField{
					{Name: "Name", Type: "string", Tag: "`json:\"name\"`"},
					{Name: "Age", Type: "int", Tag: "`json:\"age\"`"},
				},
				Code: "type Person struct {\n\tName string `json:\"name\"`\n\tAge int `json:\"age\"`\n}",
				Doc:  "Person represents a person",
			},
		},
		Functions: []model.GoFunction{
			{
				Name:      "NewPerson",
				Signature: "(name string, age int) *Person",
				Body:      "\treturn &Person{Name: name, Age: age}\n",
				Code:      "func NewPerson(name string, age int) *Person {\n\treturn &Person{Name: name, Age: age}\n}",
				Doc:       "NewPerson creates a new Person",
			},
		},
		Constants: []model.GoConstant{
			{Name: "MaxAge", Type: "int", Value: "120"},
		},
		Variables: []model.GoVariable{
			{Name: "DefaultAge", Type: "int", Value: "30"},
		},
	}

	// Create HTML generator with default options
	generator := NewGenerator(DefaultOptions())

	// Generate HTML
	html, err := generator.Generate(pkg)
	if err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Check that the HTML contains expected elements
	expectedElements := []string{
		"<title>Go Package Documentation - testpkg</title>",
		"<h1>Package testpkg</h1>",
		"Person represents a person",
		"NewPerson creates a new Person",
		"MaxAge",
		"DefaultAge",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(html, expected) {
			t.Errorf("Generated HTML doesn't contain expected element: %s", expected)
		}
	}

	// Test with custom options
	customOptions := Options{
		Title:              "Custom Title",
		SyntaxHighlighting: true,
		IncludeCSS:         true,
		CustomCSS:          ".custom { color: red; }",
	}

	customGenerator := NewGenerator(customOptions)
	customHTML, err := customGenerator.Generate(pkg)
	if err != nil {
		t.Fatalf("Failed to generate HTML with custom options: %v", err)
	}

	// Inspect the template data being passed
	t.Logf("Custom CSS being passed: '%s'", customOptions.CustomCSS)

	// Debug: Check if style tag contains the custom CSS
	styleTagStart := strings.Index(customHTML, "<style>")
	styleTagEnd := strings.Index(customHTML, "</style>")
	if styleTagStart != -1 && styleTagEnd != -1 && styleTagEnd > styleTagStart {
		styleContent := customHTML[styleTagStart+7 : styleTagEnd]
		t.Logf("Style tag content (truncated): %s", styleContent[:100])

		// Check if the style tag contains the custom CSS
		if !strings.Contains(styleContent, ".custom { color: red; }") {
			t.Logf("Custom CSS not found in style tag content")
		}
	} else {
		t.Logf("Couldn't locate style tag in output HTML")
	}

	// Check custom elements
	if !strings.Contains(customHTML, "<title>Custom Title - testpkg</title>") {
		t.Error("Custom title not applied")
	}

	if !strings.Contains(customHTML, ".custom { color: red; }") {
		t.Error("Custom CSS not included")
	}
}

// TestGenerateFromRealPackage tests HTML generation with a real parsed package
func TestGenerateFromRealPackage(t *testing.T) {
	// Skip this test if we can't find the test package
	pkg, err := parse.ParsePackage("../../../test/samplepackage")
	if err != nil {
		t.Skipf("Skipping real package test: %v", err)
	}

	// Create HTML generator
	generator := NewGenerator(DefaultOptions())

	// Generate HTML
	html, err := generator.Generate(pkg)
	if err != nil {
		t.Fatalf("Failed to generate HTML from real package: %v", err)
	}

	// Basic checks
	if !strings.Contains(html, "<h1>Package samplepackage</h1>") {
		t.Error("Generated HTML doesn't contain package name")
	}

	// Just check that the output looks reasonably large
	if len(html) < 1000 {
		t.Error("Generated HTML seems too short")
	}
}
