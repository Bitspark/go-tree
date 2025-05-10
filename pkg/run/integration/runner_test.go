// Package integration contains integration tests for the execute package
package integration

import (
	"testing"

	"bitspark.dev/go-tree/pkg/testutil"
)

// TestSimpleMathFunctions tests executing functions from the simplemath module
func TestSimpleMathFunctions(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a runner with real dependencies
	runner := testutil.CreateRunner()

	// Get the path to the test module
	modulePath, err := testutil.GetTestModulePath("simplemath")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Test table for different functions
	tests := []struct {
		name     string
		function string
		args     []interface{}
		want     interface{}
	}{
		{"Add", "Add", []interface{}{5, 3}, float64(8)},
		{"Subtract", "Subtract", []interface{}{10, 7}, float64(3)},
		{"Multiply", "Multiply", []interface{}{4, 3}, float64(12)},
		{"Divide", "Divide", []interface{}{10, 2}, float64(5)},
	}

	// Run all tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runner.ResolveAndExecuteFunc(
				modulePath,
				"github.com/test/simplemath",
				tt.function,
				tt.args...)

			if err != nil {
				t.Fatalf("Failed to execute %s: %v", tt.function, err)
			}

			// Debug output
			t.Logf("Result type: %T, value: %v", result, result)

			// Check if the result is what we expect
			// Results usually come as float64 due to JSON serialization
			if result != tt.want {
				t.Errorf("Expected %v, got %v", tt.want, result)
			}
		})
	}
}

// TestComplexReturnTypes tests functions that return complex types
func TestComplexReturnTypes(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a runner with real dependencies
	runner := testutil.CreateRunner()

	// Get the path to the test module
	modulePath, err := testutil.GetTestModulePath("complexreturn")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Test the GetPerson function which returns a struct
	result, err := runner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/complexreturn",
		"GetPerson",
		"Alice")

	if err != nil {
		t.Fatalf("Failed to execute GetPerson: %v", err)
	}

	// The result should be a map since structs are serialized to JSON
	personMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T: %v", result, result)
	}

	// Check that the name is correct
	name, ok := personMap["Name"].(string)
	if !ok || name != "Alice" {
		t.Errorf("Expected Name: Alice, got %v", personMap["Name"])
	}

	// Check that the age is correct (likely as float64 due to JSON)
	age, ok := personMap["Age"].(float64)
	if !ok || int(age) != 30 {
		t.Errorf("Expected Age: 30, got %v", personMap["Age"])
	}
}

// TestVerifyTestModulePaths verifies that the test module paths are correct
func TestVerifyTestModulePaths(t *testing.T) {
	// This test always runs, even in short mode

	// Try to get path to simplemath module
	simpleMathPath, err := testutil.GetTestModulePath("simplemath")
	if err != nil {
		t.Fatalf("Failed to get simplemath module path: %v", err)
	}

	// Check that the file exists
	t.Logf("Simplemath module path: %s", simpleMathPath)

	// Try to get path to errors module
	errorsPath, err := testutil.GetTestModulePath("errors")
	if err != nil {
		t.Fatalf("Failed to get errors module path: %v", err)
	}

	// Check that the file exists
	t.Logf("Errors module path: %s", errorsPath)

	// Try to get path to complexreturn module
	complexReturnPath, err := testutil.GetTestModulePath("complexreturn")
	if err != nil {
		t.Fatalf("Failed to get complexreturn module path: %v", err)
	}

	// Check that the file exists
	t.Logf("Complexreturn module path: %s", complexReturnPath)
}
