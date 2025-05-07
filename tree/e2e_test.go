package tree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestEndToEnd performs a complete end-to-end test:
// 1. Parse a package
// 2. Format it to a temp directory
// 3. Verify it compiles without errors
func TestEndToEnd(t *testing.T) {
	// Skip this test if go executable is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go executable not found in PATH, skipping end-to-end test")
	}

	// Create a temporary directory for the output package
	tempDir, err := os.MkdirTemp("", "golm-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Create directories for the sample package and the main program
	pkgDir := filepath.Join(tempDir, "samplepackage")
	mainDir := filepath.Join(tempDir, "main")

	if err := os.Mkdir(pkgDir, 0755); err != nil {
		t.Fatalf("Failed to create package directory: %v", err)
	}
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatalf("Failed to create main directory: %v", err)
	}

	// Parse the sample package
	pkg, err := Parse("../testdata/samplepackage")
	if err != nil {
		t.Fatalf("Failed to parse package: %v", err)
	}

	// Format the package with the original package name
	formatted, err := pkg.Format()
	if err != nil {
		t.Fatalf("Failed to format package: %v", err)
	}

	// Write formatted code to a file in the package directory
	outputFile := filepath.Join(pkgDir, "samplepackage.go")
	if err := os.WriteFile(outputFile, []byte(formatted), 0644); err != nil {
		t.Fatalf("Failed to write formatted output: %v", err)
	}

	// Create a go.mod file in the temp root directory
	rootModContent := "module example.com\n\ngo 1.18\n"
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(rootModContent), 0644); err != nil {
		t.Fatalf("Failed to create root go.mod file: %v", err)
	}

	// Create a simple test program that uses the package
	testMainFile := filepath.Join(mainDir, "main.go")
	testMainContent := `package main

import (
	"fmt"
	"example.com/samplepackage"
)

// Just use types from the package to verify it compiles
func main() {
	// Create a new role
	var role samplepackage.Role = samplepackage.RoleUser

	// Create a new user
	user := samplepackage.User{
		ID:   1,
		Name: "Test User",
	}

	fmt.Printf("Created user %s with ID %d and role %s\n", 
		user.Name, user.ID, role)
}
`
	if err := os.WriteFile(testMainFile, []byte(testMainContent), 0644); err != nil {
		t.Fatalf("Failed to create test main file: %v", err)
	}

	// Run go vet on the sample package
	vetCmd := exec.Command("go", "vet", "./samplepackage/...")
	vetCmd.Dir = tempDir
	if output, err := vetCmd.CombinedOutput(); err != nil {
		t.Fatalf("go vet failed on samplepackage: %v\nOutput: %s", err, output)
	}

	// Try to build the main program
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "test_binary"), "./main")
	buildCmd.Dir = tempDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", err, output)
	}

	t.Log("Successfully compiled the formatted package")
}

// TestRoundTrip tests parsing a package, formatting it, then parsing it again
// to ensure we don't lose information in the process
func TestRoundTrip(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "golm-roundtrip-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}()

	// Parse original package
	originalPkg, err := Parse("../testdata/samplepackage")
	if err != nil {
		t.Fatalf("Failed to parse original package: %v", err)
	}

	// Format the package
	formatted, err := originalPkg.Format()
	if err != nil {
		t.Fatalf("Failed to format package: %v", err)
	}

	// Write to temporary file
	outputFile := filepath.Join(tempDir, "formatted.go")
	if err := os.WriteFile(outputFile, []byte(formatted), 0644); err != nil {
		t.Fatalf("Failed to write formatted output: %v", err)
	}

	// Create a temp package directory
	tmpPkgDir := filepath.Join(tempDir, "tmppackage")
	if err := os.Mkdir(tmpPkgDir, 0755); err != nil {
		t.Fatalf("Failed to create temp package directory: %v", err)
	}

	// Copy the formatted file to the temp package directory
	if err := os.WriteFile(filepath.Join(tmpPkgDir, "package.go"), []byte(formatted), 0644); err != nil {
		t.Fatalf("Failed to copy formatted file: %v", err)
	}

	// Parse the formatted package
	reparsedPkg, err := Parse(tmpPkgDir)
	if err != nil {
		t.Fatalf("Failed to reparse formatted package: %v", err)
	}

	// Compare key metrics between original and reparsed package
	comparisons := []struct {
		name     string
		original int
		reparsed int
	}{
		{"types", len(originalPkg.Model.Types), len(reparsedPkg.Model.Types)},
		{"functions", len(originalPkg.Model.Functions), len(reparsedPkg.Model.Functions)},
		{"constants", len(originalPkg.Model.Constants), len(reparsedPkg.Model.Constants)},
		{"variables", len(originalPkg.Model.Variables), len(reparsedPkg.Model.Variables)},
		{"imports", len(originalPkg.Model.Imports), len(reparsedPkg.Model.Imports)},
	}

	for _, c := range comparisons {
		if c.original != c.reparsed {
			t.Errorf("Count mismatch for %s: original=%d, reparsed=%d",
				c.name, c.original, c.reparsed)
		}
	}

	// Compare specific elements (types, functions, etc.)
	// Original type names
	originalTypeNames := make(map[string]bool)
	for _, typ := range originalPkg.Model.Types {
		originalTypeNames[typ.Name] = true
	}

	// Check if all original types exist in reparsed package
	for _, typ := range reparsedPkg.Model.Types {
		if !originalTypeNames[typ.Name] {
			t.Errorf("Type %s found in reparsed package but not in original", typ.Name)
		}
	}

	// Similar checks could be done for functions, constants, etc.

	t.Log("Round-trip parsing succeeded with no significant data loss")
}
