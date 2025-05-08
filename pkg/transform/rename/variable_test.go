package rename

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/loader"
)

// TestVariableRenamer tests renaming a variable and verifies position tracking
func TestVariableRenamer(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get sample package
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	// Check that DefaultTimeout variable exists
	defaultTimeout, ok := samplePkg.Variables["DefaultTimeout"]
	if !ok {
		t.Fatalf("Expected to find variable 'DefaultTimeout'")
	}

	// Store original position
	originalPos := defaultTimeout.Pos
	originalEnd := defaultTimeout.End
	originalPosition := defaultTimeout.GetPosition()

	if originalPosition == nil {
		t.Fatal("Expected DefaultTimeout to have position information")
	}

	// Create a transformer to rename DefaultTimeout to GlobalTimeout
	renamer := NewVariableRenamer("DefaultTimeout", "GlobalTimeout", false)

	// Apply the transformation
	result := renamer.Transform(mod)
	if !result.Success {
		t.Fatalf("Failed to apply transformation: %v", result.Error)
	}

	// Verify the old variable no longer exists
	_, ok = samplePkg.Variables["DefaultTimeout"]
	if ok {
		t.Error("Expected 'DefaultTimeout' to be removed")
	}

	// Verify the new variable exists
	globalTimeout, ok := samplePkg.Variables["GlobalTimeout"]
	if !ok {
		t.Fatalf("Expected to find variable 'GlobalTimeout'")
	}

	// Verify package and file are marked as modified
	if !samplePkg.IsModified {
		t.Error("Expected package to be marked as modified")
	}

	if !globalTimeout.File.IsModified {
		t.Error("Expected file to be marked as modified")
	}

	// Verify positions were preserved
	if globalTimeout.Pos != originalPos {
		t.Errorf("Expected Pos to be preserved: wanted %v, got %v",
			originalPos, globalTimeout.Pos)
	}

	if globalTimeout.End != originalEnd {
		t.Errorf("Expected End to be preserved: wanted %v, got %v",
			originalEnd, globalTimeout.End)
	}

	// Verify GetPosition returns the same information
	newPosition := globalTimeout.GetPosition()
	if newPosition == nil {
		t.Fatal("Expected GlobalTimeout to have position information")
	}

	// Verify line/column information is preserved
	if newPosition.LineStart != originalPosition.LineStart {
		t.Errorf("Expected line start to be preserved: wanted %d, got %d",
			originalPosition.LineStart, newPosition.LineStart)
	}

	if newPosition.ColStart != originalPosition.ColStart {
		t.Errorf("Expected column start to be preserved: wanted %d, got %d",
			originalPosition.ColStart, newPosition.ColStart)
	}
}

// TestVariableReferenceUpdates tests that references to the variable are updated
func TestVariableReferenceUpdates(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get sample package
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	// Create a transformer to rename a variable that has references
	renamer := NewVariableRenamer("VariableWithReferences", "RenamedVar", false)

	// Apply the transformation
	result := renamer.Transform(mod)
	if !result.Success {
		t.Fatalf("Failed to apply transformation: %v", result.Error)
	}

	// Check that references were updated in functions
	for _, fn := range samplePkg.Functions {
		// This will check for variable references in functions
		for _, ref := range fn.References {
			if ref.Name == "VariableWithReferences" {
				t.Errorf("Found unchanged reference to 'VariableWithReferences' in function %s", fn.Name)
			}
		}
	}

	// Check for correct change previews - should include all reference updates
	if len(result.Changes) < 2 {
		t.Errorf("Expected multiple changes (declaration + references), got %d", len(result.Changes))
	}
}

// TestPackageTargeting tests that variable renaming can be restricted to a specific package
func TestPackageTargeting(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Create a transformer with package targeting
	renamer := NewVariableRenamerWithPackage("test/samplepackage", "DefaultTimeout", "GlobalTimeout", false)

	// Apply the transformation
	result := renamer.Transform(mod)
	if !result.Success {
		t.Fatalf("Failed to apply transformation: %v", result.Error)
	}

	// Create a transformer with an incorrect package path
	badRenamer := NewVariableRenamerWithPackage("non/existent/package", "DefaultTimeout", "GlobalTimeout", false)
	badResult := badRenamer.Transform(mod)

	// Should fail with proper error message
	if badResult.Success {
		t.Errorf("Expected transformation to fail with non-existent package")
	}
	if badResult.Error == nil || badResult.Details == "" {
		t.Errorf("Expected error and details about non-existent package")
	}
}

// TestLineNumberTracking tests that change previews include correct line numbers
func TestLineNumberTracking(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Create a transformer
	renamer := NewVariableRenamer("DefaultTimeout", "GlobalTimeout", false)

	// Apply the transformation
	result := renamer.Transform(mod)
	if !result.Success {
		t.Fatalf("Failed to apply transformation: %v", result.Error)
	}

	// Verify line numbers are included in change previews
	for _, change := range result.Changes {
		if change.LineNumber <= 0 {
			t.Errorf("Expected valid line number in change preview, got %d", change.LineNumber)
		}
	}
}

// TestVariableValidation tests handling of invalid identifiers and name conflicts
func TestVariableValidation(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Test invalid Go identifier
	invalidRenamer := NewVariableRenamer("DefaultTimeout", "123-Invalid-Name", false)
	invalidResult := invalidRenamer.Transform(mod)

	if invalidResult.Success {
		t.Errorf("Expected transformation to fail with invalid Go identifier")
	}

	// Test name conflict
	// Assuming "ExistingVar" already exists in the package
	conflictRenamer := NewVariableRenamer("DefaultTimeout", "ExistingVar", false)
	conflictResult := conflictRenamer.Transform(mod)

	if conflictResult.Success {
		t.Errorf("Expected transformation to fail due to name conflict")
	}
}

// TestDocCommentUpdates tests that documentation references to the variable are updated
func TestDocCommentUpdates(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get sample package
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	// Check that DefaultTimeout variable exists and has doc comments
	defaultTimeout, ok := samplePkg.Variables["DefaultTimeout"]
	if !ok || defaultTimeout.Doc == nil {
		t.Skip("Test requires DefaultTimeout variable with doc comments")
	}

	// Create a transformer
	renamer := NewVariableRenamer("DefaultTimeout", "GlobalTimeout", false)

	// Apply the transformation
	result := renamer.Transform(mod)
	if !result.Success {
		t.Fatalf("Failed to apply transformation: %v", result.Error)
	}

	// Get the renamed variable
	globalTimeout, ok := samplePkg.Variables["GlobalTimeout"]
	if !ok {
		t.Fatalf("Expected to find variable 'GlobalTimeout'")
	}

	// Check that doc comments were updated
	if globalTimeout.Doc != nil {
		docText := globalTimeout.Doc.Text()
		if contains(docText, "DefaultTimeout") {
			t.Errorf("Doc comments still contain reference to old name: %s", docText)
		}
	}

	// Check for related function documentation
	for _, fn := range samplePkg.Functions {
		if fn.Doc != nil {
			docText := fn.Doc.Text()
			if contains(docText, "DefaultTimeout") {
				t.Errorf("Function %s doc still contains reference to old name: %s", fn.Name, docText)
			}
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && s != s[len(substr):]
}

// TestRenameConvenienceMethod tests the Rename convenience method
func TestRenameConvenienceMethod(t *testing.T) {
	// Load test module
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Create a transformer
	renamer := NewVariableRenamer("DefaultTimeout", "GlobalTimeout", false)

	// Call the Rename method
	err = renamer.Rename(mod)
	if err != nil {
		t.Fatalf("Rename method failed: %v", err)
	}

	// Verify the variable was renamed
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	_, ok = samplePkg.Variables["DefaultTimeout"]
	if ok {
		t.Error("Expected 'DefaultTimeout' to be removed")
	}

	_, ok = samplePkg.Variables["GlobalTimeout"]
	if !ok {
		t.Fatalf("Expected to find variable 'GlobalTimeout'")
	}
}
