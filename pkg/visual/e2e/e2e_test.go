// Package e2e contains end-to-end tests for the Go-Tree formatters and visualizers.
package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
	"bitspark.dev/go-tree/pkg/core/parse"
	"bitspark.dev/go-tree/pkg/visual/html"
	"bitspark.dev/go-tree/pkg/visual/markdown"
)

// TestRealPackages tests our formatters with real Go packages of different complexity
func TestRealPackages(t *testing.T) {
	testCases := []struct {
		name        string
		packagePath string
		skipIfErr   bool
	}{
		// Simple test package
		{
			name:        "Simple test package",
			packagePath: "./testdata/packages/simple",
			skipIfErr:   false,
		},
		// Complex test package
		{
			name:        "Complex test package",
			packagePath: "./testdata/packages/complex",
			skipIfErr:   false,
		},
		// Edge cases test package
		{
			name:        "Edge cases test package",
			packagePath: "./testdata/packages/edge_cases",
			skipIfErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the package
			pkg, err := parse.ParsePackage(tc.packagePath)
			if err != nil {
				if tc.skipIfErr {
					t.Skipf("Skipping test for %s: %v", tc.packagePath, err)
					return
				}
				t.Fatalf("Failed to parse package %s: %v", tc.packagePath, err)
			}

			// Test HTML generation
			htmlGen := html.NewGenerator(html.DefaultOptions())
			htmlOutput, err := htmlGen.Generate(pkg)
			if err != nil {
				t.Errorf("HTML generation failed: %v", err)
			} else {
				validateHTMLOutput(t, pkg, htmlOutput)
			}

			// Test Markdown generation
			mdGen := markdown.NewGenerator(markdown.DefaultOptions())
			mdOutput, err := mdGen.Generate(pkg)
			if err != nil {
				t.Errorf("Markdown generation failed: %v", err)
			} else {
				validateMarkdownOutput(t, pkg, mdOutput)
			}

			// Save outputs for manual inspection if requested
			if os.Getenv("SAVE_TEST_OUTPUT") == "true" {
				saveOutputs(t, tc.name, htmlOutput, mdOutput)
			}
		})
	}
}

// validateHTMLOutput checks that HTML output contains expected elements
func validateHTMLOutput(t *testing.T, pkg *model.GoPackage, output string) {
	// Basic structure checks
	if !strings.Contains(output, "<title>") {
		t.Error("HTML output missing title element")
	}

	if !strings.Contains(output, "<h1>Package "+pkg.Name+"</h1>") {
		t.Error("HTML output missing package name heading")
	}

	// Check for package elements
	validatePackageElementsInHTML(t, pkg, output)
}

// validateMarkdownOutput checks that Markdown output contains expected elements
func validateMarkdownOutput(t *testing.T, pkg *model.GoPackage, output string) {
	// Basic structure checks
	if !strings.Contains(output, "# Package "+pkg.Name) {
		t.Error("Markdown output missing package name heading")
	}

	// Check for package elements
	validatePackageElementsInMarkdown(t, pkg, output)
}

// validatePackageElementsInHTML checks for package elements in HTML output
func validatePackageElementsInHTML(t *testing.T, pkg *model.GoPackage, output string) {
	// Check for a sample of types
	for i, typ := range pkg.Types {
		if i >= 3 { // Check first few to keep test reasonable
			break
		}
		if !strings.Contains(output, typ.Name) {
			t.Errorf("HTML output missing type: %s", typ.Name)
		}
	}

	// Check for a sample of functions
	for i, fn := range pkg.Functions {
		if i >= 3 {
			break
		}
		if !strings.Contains(output, fn.Name) {
			t.Errorf("HTML output missing function: %s", fn.Name)
		}
	}
}

// validatePackageElementsInMarkdown checks for package elements in Markdown
func validatePackageElementsInMarkdown(t *testing.T, pkg *model.GoPackage, output string) {
	// Check for a sample of types
	for i, typ := range pkg.Types {
		if i >= 3 {
			break
		}
		if !strings.Contains(output, "## Type: "+typ.Name) {
			t.Errorf("Markdown output missing type: %s", typ.Name)
		}
	}

	// Check for a sample of functions
	for i, fn := range pkg.Functions {
		if i >= 3 {
			break
		}

		if fn.Receiver != nil {
			// This is a method - should have format like: "## Method: (s *SimpleStruct) SimpleMethod"
			methodHeading := "## Method: ("
			if !strings.Contains(output, methodHeading) || !strings.Contains(output, fn.Name) {
				t.Errorf("Markdown output missing method: %s", fn.Name)
			}
		} else {
			// This is a function - should have format like: "## Function: SimpleFunction"
			functionHeading := "## Function: " + fn.Name
			if !strings.Contains(output, functionHeading) {
				t.Errorf("Markdown output missing function: %s", fn.Name)
			}
		}
	}
}

// saveOutputs saves the generated outputs for manual inspection
func saveOutputs(t *testing.T, testName string, htmlOutput, mdOutput string) {
	outDir := filepath.Join("testdata", "outputs")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Logf("Failed to create output directory: %v", err)
		return
	}

	// Clean test name for filename
	safeTestName := strings.ReplaceAll(testName, " ", "_")
	safeTestName = strings.ReplaceAll(safeTestName, "/", "_")

	// Save HTML output
	htmlPath := filepath.Join(outDir, safeTestName+".html")
	if err := os.WriteFile(htmlPath, []byte(htmlOutput), 0644); err != nil {
		t.Logf("Failed to save HTML output: %v", err)
	}

	// Save Markdown output
	mdPath := filepath.Join(outDir, safeTestName+".md")
	if err := os.WriteFile(mdPath, []byte(mdOutput), 0644); err != nil {
		t.Logf("Failed to save Markdown output: %v", err)
	}

	t.Logf("Saved outputs to %s and %s", htmlPath, mdPath)
}
