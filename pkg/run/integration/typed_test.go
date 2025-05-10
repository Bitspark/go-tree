package integration

import (
	"testing"

	"bitspark.dev/go-tree/pkg/testutil"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestTypedFunctionRunner tests the typed function runner with real functions
func TestTypedFunctionRunner(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get path to the test module
	modulePath, err := testutil.GetTestModulePath("simplemath")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Create the typed function runner
	typedRunner := testutil.CreateTypedRunner()

	// Resolve the module to get symbols
	baseRunner := testutil.CreateRunner()
	rawModule, err := baseRunner.Resolver.ResolveModule(modulePath, "", nil)
	if err != nil {
		t.Fatalf("Failed to resolve module: %v", err)
	}

	// Type assertion to convert from interface{} to *typesys.Module
	module, ok := rawModule.(*typesys.Module)
	if !ok {
		t.Fatalf("Failed to convert module: got %T, expected *typesys.Module", rawModule)
	}

	// Find the Add function
	var addFunc *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			addFunc = sym
			break
		}
	}

	if addFunc == nil {
		t.Fatal("Failed to find Add function in module")
	}

	// Test typed integer function execution
	result, err := typedRunner.ExecuteIntegerFunction(
		module,
		addFunc,
		7, 3)

	if err != nil {
		t.Fatalf("Failed to execute integer function: %v", err)
	}

	// Verify result
	expected := 10
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}

	// Test with a wrapped function
	addWrapper := typedRunner.WrapIntegerFunction(module, addFunc)

	// Call the wrapper function
	wrapperResult, err := addWrapper(12, 8)

	if err != nil {
		t.Fatalf("Failed to execute wrapped function: %v", err)
	}

	// Verify wrapper result
	expected = 20
	if wrapperResult != expected {
		t.Errorf("Expected %d, got %d", expected, wrapperResult)
	}
}
