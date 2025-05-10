package integration

import (
	"testing"
	"time"

	"bitspark.dev/go-tree/pkg/run/execute/specialized"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute/integration/testutil"
)

// TestRetryingFunctionRunner tests the retrying function runner with real error functions
func TestRetryingFunctionRunner(t *testing.T) {
	t.Skip("Skipping for now - implement AttemptNetworkAccess in complexreturn test module to fully test")

	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get path to the error test module
	modulePath, err := testutil.GetTestModulePath("errors")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Create a retrying runner with real dependencies
	baseRunner := testutil.CreateRunner()
	retryRunner := testutil.CreateRetryingRunner()

	// Setup a policy with 3 max retries
	retryRunner.WithPolicy(&specialized.RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Millisecond, // Use small delays for tests
		MaxDelay:      50 * time.Millisecond,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"temporary failure", // This should match our test module's error message
		},
	})

	// Execute a function that should succeed after retries
	result, err := retryRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"TemporaryFailure", // This function in our test module should fail temporarily
		2)                  // Value indicating how many times to fail before succeeding

	if err != nil {
		t.Fatalf("Expected success after retries: %v", err)
	}

	// Check the result
	expectedResult := "success after retries"
	if result != expectedResult {
		t.Errorf("Expected '%s', got: %v", expectedResult, result)
	}

	// Check that we get an error when using the base runner without retries
	_, baseErr := baseRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"TemporaryFailure",
		1) // Should fail on first attempt

	if baseErr == nil {
		t.Errorf("Expected base runner to fail without retries")
	}

	// Try a function that returns a non-retryable error
	_, nonRetryableErr := retryRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"PermanentFailure",
		0)

	if nonRetryableErr == nil {
		t.Errorf("Expected error for non-retryable function")
	}
}

// TestBatchFunctionRunner tests the batch function runner with real functions
func TestBatchFunctionRunner(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get path to the test module
	modulePath, err := testutil.GetTestModulePath("simplemath")
	if err != nil {
		t.Fatalf("Failed to get test module path: %v", err)
	}

	// Create a batch runner
	batchRunner := testutil.CreateBatchRunner()

	// Resolve the module to get symbols
	baseRunner := testutil.CreateRunner()
	module, err := baseRunner.Resolver.ResolveModule(modulePath, "", nil)
	if err != nil {
		t.Fatalf("Failed to resolve module: %v", err)
	}

	// Get the package
	pkg, ok := module.Packages["github.com/test/simplemath"]
	if !ok {
		t.Fatalf("Package 'github.com/test/simplemath' not found in module")
	}

	// Find the functions
	var addFunc, subtractFunc, multiplyFunc *typesys.Symbol
	for _, sym := range pkg.Symbols {
		switch sym.Name {
		case "Add":
			addFunc = sym
		case "Subtract":
			subtractFunc = sym
		case "Multiply":
			multiplyFunc = sym
		}
	}

	if addFunc == nil || subtractFunc == nil || multiplyFunc == nil {
		t.Fatal("Failed to find required functions in module")
	}

	// Add functions to the batch
	batchRunner.Add(module, addFunc, 5, 3)
	batchRunner.AddWithDescription("Subtraction", module, subtractFunc, 10, 4)
	batchRunner.Add(module, multiplyFunc, 2, 6)

	// Execute the batch
	err = batchRunner.Execute()
	if err != nil {
		t.Fatalf("Failed to execute batch: %v", err)
	}

	// Check all results
	results := batchRunner.GetResults()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Check individual results
	expectedValues := []float64{8, 6, 12} // Add, Subtract, Multiply
	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Result %d had error: %v", i, result.Error)
		}

		// Results come as float64 due to JSON serialization
		value, ok := result.Result.(float64)
		if !ok {
			t.Errorf("Result %d: Expected float64, got %T", i, result.Result)
			continue
		}

		if value != expectedValues[i] {
			t.Errorf("Result %d: Expected %v, got %v", i, expectedValues[i], value)
		}
	}

	// Test parallel execution
	parallelBatchRunner := testutil.CreateBatchRunner()
	parallelBatchRunner.WithParallel(true)

	// Add the same functions again
	parallelBatchRunner.Add(module, addFunc, 5, 3)
	parallelBatchRunner.Add(module, subtractFunc, 10, 4)
	parallelBatchRunner.Add(module, multiplyFunc, 2, 6)

	// Execute in parallel and time it
	start := time.Now()
	err = parallelBatchRunner.Execute()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute parallel batch: %v", err)
	}

	// Check that all functions succeeded
	if !parallelBatchRunner.Successful() {
		t.Error("Expected all functions to succeed in parallel execution")
	}

	t.Logf("Parallel execution took %v", duration)
}
