package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkg/core/parse"
	"bitspark.dev/go-tree/pkg/visual/html"
	"bitspark.dev/go-tree/pkg/visual/markdown"
)

// TestAgainstGoldenFiles tests formatters against predefined golden files
func TestAgainstGoldenFiles(t *testing.T) {
	testCases := []struct {
		name        string
		packagePath string
		goldenHTML  string
		goldenMD    string
	}{
		{
			name:        "Simple package",
			packagePath: "./testdata/packages/simple",
			goldenHTML:  "./testdata/golden/simple.html",
			goldenMD:    "./testdata/golden/simple.md",
		},
		{
			name:        "Complex package",
			packagePath: "./testdata/packages/complex",
			goldenHTML:  "./testdata/golden/complex.html",
			goldenMD:    "./testdata/golden/complex.md",
		},
		{
			name:        "Edge cases",
			packagePath: "./testdata/packages/edge_cases",
			goldenHTML:  "./testdata/golden/edge_cases.html",
			goldenMD:    "./testdata/golden/edge_cases.md",
		},
	}

	updateGolden := os.Getenv("UPDATE_GOLDEN") == "true"

	// Track temporary files for cleanup
	var tempFiles []string

	// Register cleanup function to remove temporary files
	t.Cleanup(func() {
		for _, file := range tempFiles {
			if _, err := os.Stat(file); err == nil {
				os.Remove(file)
				t.Logf("Cleaned up temporary file: %s", file)
			}
		}
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the test package
			pkg, err := parse.ParsePackage(tc.packagePath)
			if err != nil {
				t.Fatalf("Failed to parse package %s: %v", tc.packagePath, err)
			}

			// Test HTML generation
			htmlGen := html.NewGenerator(html.DefaultOptions())
			htmlOutput, err := htmlGen.Generate(pkg)
			if err != nil {
				t.Fatalf("HTML generation failed: %v", err)
			}

			// Test Markdown generation
			mdGen := markdown.NewGenerator(markdown.DefaultOptions())
			mdOutput, err := mdGen.Generate(pkg)
			if err != nil {
				t.Fatalf("Markdown generation failed: %v", err)
			}

			// Handle golden files
			if updateGolden {
				// Update golden files with current output
				if err := os.MkdirAll(filepath.Dir(tc.goldenHTML), 0755); err != nil {
					t.Fatalf("Failed to create golden directory: %v", err)
				}

				if err := os.WriteFile(tc.goldenHTML, []byte(htmlOutput), 0644); err != nil {
					t.Fatalf("Failed to update HTML golden file: %v", err)
				}

				if err := os.WriteFile(tc.goldenMD, []byte(mdOutput), 0644); err != nil {
					t.Fatalf("Failed to update Markdown golden file: %v", err)
				}

				t.Logf("Updated golden files for %s", tc.name)
			} else {
				// Compare with existing golden files
				compareWithGoldenFile(t, htmlOutput, tc.goldenHTML, "HTML", &tempFiles)
				compareWithGoldenFile(t, mdOutput, tc.goldenMD, "Markdown", &tempFiles)
			}
		})
	}
}

// compareWithGoldenFile compares output with a golden file
func compareWithGoldenFile(t *testing.T, output, goldenPath, format string, tempFiles *[]string) {
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read %s golden file: %v", format, err)
	}

	if output != string(golden) {
		t.Errorf("%s output doesn't match golden file %s", format, goldenPath)
		// For better diagnostics, you could:
		// 1. Save the current output to a temp file
		// 2. Use a diff library to highlight differences
		// 3. Show a limited context around the first difference

		// Save the current output to a temp file for debugging
		tempFile := goldenPath + ".current"
		if err := os.WriteFile(tempFile, []byte(output), 0644); err != nil {
			t.Logf("Failed to write current output to temp file: %v", err)
		} else {
			t.Logf("Current output saved to %s for comparison", tempFile)
			// Add to list of files to clean up
			*tempFiles = append(*tempFiles, tempFile)
		}
	}
}
