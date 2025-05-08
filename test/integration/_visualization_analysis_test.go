package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"bitspark.dev/go-tree/pkgold/analysis"
	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/visual"
)

// TestVisualizationAnalysisWorkflow demonstrates a workflow combining:
// 1. Loading a Go module
// 2. Performing static analysis on its structure
// 3. Generating a dependency graph visualization
// 4. Exporting the analysis results and visualizations
func TestVisualizationAnalysisWorkflow(t *testing.T) {
	// Setup test directories
	testDir := filepath.Join("testdata", "visualization")
	outDir := filepath.Join(testDir, "output")

	// Ensure output directory exists
	if err := os.MkdirAll(outDir, 0750); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	defer os.RemoveAll(outDir) // Clean up after test

	// Step 1: Load the module
	modLoader := loader.NewGoModuleLoader()
	mod, err := modLoader.Load(testDir)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Step 2: Analyze the module structure
	// 2.1. Package dependencies analysis
	depAnalyzer := analysis.NewDependencyAnalyzer()
	deps, err := depAnalyzer.AnalyzePackageDependencies(mod)
	if err != nil {
		t.Fatalf("Failed to analyze package dependencies: %v", err)
	}

	// Save dependency analysis result
	depJSON, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal dependencies: %v", err)
	}

	depPath := filepath.Join(outDir, "package_dependencies.json")
	if err := os.WriteFile(depPath, depJSON, 0644); err != nil {
		t.Fatalf("Failed to write dependencies: %v", err)
	}

	// 2.2. Type analysis
	typeAnalyzer := analysis.NewTypeAnalyzer()
	typeInfo, err := typeAnalyzer.AnalyzeTypes(mod)
	if err != nil {
		t.Fatalf("Failed to analyze types: %v", err)
	}

	// Save type analysis result
	typeJSON, err := json.MarshalIndent(typeInfo, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal type info: %v", err)
	}

	typePath := filepath.Join(outDir, "type_analysis.json")
	if err := os.WriteFile(typePath, typeJSON, 0644); err != nil {
		t.Fatalf("Failed to write type info: %v", err)
	}

	// Step 3: Generate visualizations
	// 3.1. Package dependency graph
	depVisualizer := visual.NewDependencyVisualizer()
	dotGraphPath := filepath.Join(outDir, "package_deps.dot")
	svgGraphPath := filepath.Join(outDir, "package_deps.svg")

	err = depVisualizer.VisualizePackageDependencies(mod, deps, dotGraphPath)
	if err != nil {
		t.Fatalf("Failed to visualize package dependencies: %v", err)
	}

	// Check if Graphviz is available for SVG conversion
	_, err = os.Stat(dotGraphPath)
	if err == nil {
		// Convert DOT to SVG using Graphviz (if available)
		err = depVisualizer.ConvertDotToSVG(dotGraphPath, svgGraphPath)
		if err != nil {
			// Not failing the test if just the conversion fails
			t.Logf("Failed to convert DOT to SVG: %v", err)
		}
	}

	// 3.2. Module structure visualization
	moduleVisualizer := visual.NewModuleVisualizer()
	structurePath := filepath.Join(outDir, "module_structure.html")

	err = moduleVisualizer.VisualizeModuleStructure(mod, structurePath)
	if err != nil {
		t.Fatalf("Failed to visualize module structure: %v", err)
	}

	// Step 4: Verify outputs exist
	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("Failed to read output directory: %v", err)
	}

	expectedFiles := map[string]bool{
		"package_dependencies.json": false,
		"type_analysis.json":        false,
		"package_deps.dot":          false,
		"module_structure.html":     false,
	}

	for _, file := range files {
		expectedFiles[file.Name()] = true
	}

	for fileName, found := range expectedFiles {
		if !found {
			t.Errorf("Expected output file %s was not created", fileName)
		}
	}
}
