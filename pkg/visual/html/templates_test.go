package html

import (
	"html/template"
	"strings"
	"testing"
)

func TestBaseTemplateParses(t *testing.T) {
	// Verify that the BaseTemplate can be parsed without errors
	tmpl, err := template.New("html").Parse(BaseTemplate)
	if err != nil {
		t.Fatalf("Failed to parse BaseTemplate: %v", err)
	}

	if tmpl == nil {
		t.Fatal("Parsed template is nil")
	}
}

func TestBaseTemplateRenders(t *testing.T) {
	// Test that the template renders with expected values
	tmpl, err := template.New("html").Parse(BaseTemplate)
	if err != nil {
		t.Fatalf("Failed to parse BaseTemplate: %v", err)
	}

	// Create test data to render
	data := map[string]interface{}{
		"Title":        "Test Title",
		"ModulePath":   "example.com/test",
		"GoVersion":    "1.18",
		"PackageCount": 5,
		"Content":      template.HTML("<div>Test Content</div>"),
	}

	// Execute the template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	result := buf.String()

	// Check for expected content
	expectedItems := []string{
		"<title>Test Title</title>",
		"<h1>Test Title</h1>",
		"<strong>Module Path:</strong> example.com/test",
		"<strong>Go Version:</strong> 1.18",
		"<strong>Packages:</strong> 5",
		"<div>Test Content</div>",
	}

	for _, item := range expectedItems {
		if !strings.Contains(result, item) {
			t.Errorf("Expected rendered template to contain '%s', but it doesn't", item)
		}
	}
}

func TestBaseTemplateStyles(t *testing.T) {
	// Test that the template contains essential styling elements
	essentialStyles := []string{
		"--primary-color:",
		"--background-color:",
		"--text-color:",
		".symbol-fn {",
		".symbol-type {",
		".symbol-var {",
		".symbol-const {",
		".tag-exported {",
		".tag-private {",
		"@media (prefers-color-scheme: dark) {", // Check for dark mode support
	}

	for _, style := range essentialStyles {
		if !strings.Contains(BaseTemplate, style) {
			t.Errorf("Expected BaseTemplate to contain '%s' style, but it doesn't", style)
		}
	}
}

func TestBaseTemplateStructure(t *testing.T) {
	// Test that the template has the expected HTML structure
	essentialTags := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"<meta name=\"viewport\"",
		"<body>",
		"<div class=\"container\">",
		"<div class=\"module-info\">",
	}

	for _, tag := range essentialTags {
		if !strings.Contains(BaseTemplate, tag) {
			t.Errorf("Expected BaseTemplate to contain '%s' HTML element, but it doesn't", tag)
		}
	}
}
