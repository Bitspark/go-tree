package integration

import (
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkgold/analysis"
	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/transform"
)

// TestDocumentationGenerationWorkflow demonstrates a workflow for:
// 1. Loading a Go module
// 2. Extracting documentation from code comments and structure
// 3. Transforming the documentation into different formats
// 4. Generating organized documentation output
func TestDocumentationGenerationWorkflow(t *testing.T) {
	// Setup test directories
	testDir := filepath.Join("testdata", "docgen")
	outDir := filepath.Join(testDir, "output")

	// Ensure output directory exists
	if err := os.MkdirAll(outDir, 0750); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer os.RemoveAll(outDir) // Clean up after test

	// Step 1: Load the module
	modLoader := loader.NewGoModuleLoader()
	loadOptions := loader.DefaultLoadOptions()
	loadOptions.LoadDocs = true // Ensure we load documentation comments

	mod, err := modLoader.LoadWithOptions(testDir, loadOptions)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Step 2: Extract and analyze documentation
	// 2.1. Extract doc comments and structure
	docExtractor := analysis.NewDocumentationExtractor()
	docs, err := docExtractor.ExtractDocs(mod)
	if err != nil {
		t.Fatalf("Failed to extract documentation: %v", err)
	}

	// 2.2. Analyze documentation coverage
	coverageAnalyzer := analysis.NewCoverageAnalyzer()
	coverage, err := coverageAnalyzer.AnalyzeDocCoverage(mod, docs)
	if err != nil {
		t.Fatalf("Failed to analyze documentation coverage: %v", err)
	}

	// Save coverage report
	coveragePath := filepath.Join(outDir, "doc_coverage.json")
	coverageReporter := analysis.NewCoverageReporter()
	err = coverageReporter.ExportJSON(coverage, coveragePath)
	if err != nil {
		t.Fatalf("Failed to export coverage report: %v", err)
	}

	// Step 3: Generate documentation in multiple formats
	// 3.1. Generate Markdown docs
	mdGenerator := transform.NewMarkdownGenerator()
	err = mdGenerator.GeneratePackageDocs(mod, docs, outDir)
	if err != nil {
		t.Fatalf("Failed to generate Markdown docs: %v", err)
	}

	// 3.2. Generate HTML docs
	htmlGenerator := transform.NewHTMLGenerator()
	err = htmlGenerator.GenerateModuleDocs(mod, docs, filepath.Join(outDir, "html"))
	if err != nil {
		t.Fatalf("Failed to generate HTML docs: %v", err)
	}

	// 3.3. Generate examples from doc tests
	exampleGenerator := transform.NewExampleGenerator()
	err = exampleGenerator.GenerateExamples(mod, docs, filepath.Join(outDir, "examples"))
	if err != nil {
		t.Fatalf("Failed to generate examples: %v", err)
	}

	// Step 4: Generate a documentation index
	indexGenerator := transform.NewIndexGenerator()
	err = indexGenerator.GenerateIndex(mod, docs, filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("Failed to generate documentation index: %v", err)
	}

	// Step 5: Verify outputs exist
	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	// Check that we have the expected directories and files
	expectedFiles := []string{
		"doc_coverage.json",
		"index.html",
		"html",
		"examples",
	}

	foundFiles := make(map[string]bool)
	for _, file := range files {
		foundFiles[file.Name()] = true
	}

	for _, fileName := range expectedFiles {
		if !foundFiles[fileName] {
			t.Errorf("Expected output file/directory %s was not created", fileName)
		}
	}

	// Verify we have at least one markdown file generated
	mdFiles, err := filepath.Glob(filepath.Join(outDir, "*.md"))
	if err != nil {
		t.Fatalf("Failed to find Markdown files: %v", err)
	}

	if len(mdFiles) == 0 {
		t.Error("Expected at least one Markdown file to be generated")
	}
}
