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
