package integration

import (
	"testing"
	"time"

	"bitspark.dev/go-tree/pkg/testutil"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute/specialized"
)

// TestRetryingFunctionRunner tests the retrying function runner with real error functions
func TestRetryingFunctionRunner(t *testing.T) {
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

	// Setup a policy with only 2 retries to keep test times reasonable
	retryPolicy := &specialized.RetryPolicy{
		MaxRetries:    2,
		InitialDelay:  50 * time.Millisecond, // Longer delay for more reliable timing tests
		MaxDelay:      200 * time.Millisecond,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"temporary failure", // This should match our RetryableError function
		},
	}
	retryRunner.WithPolicy(retryPolicy)

	// ---------- Test 1: The RetryingFunctionRunner should properly retry based on error pattern matching ----------
	// Measure execution time for the retryable error (should be slower due to retries)
	startRetryable := time.Now()
	_, errRetryable := retryRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"RetryableError")
	durationRetryable := time.Since(startRetryable)

	// This should eventually fail but should have retried (taking longer)
	if errRetryable == nil {
		t.Error("RetryableError should eventually fail")
	}

	// Measure execution time for the non-retryable error (should be faster, no retries)
	startNonRetryable := time.Now()
	_, errNonRetryable := retryRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"NonRetryableError")
	durationNonRetryable := time.Since(startNonRetryable)

	// Should fail without retries
	if errNonRetryable == nil {
		t.Error("Expected NonRetryableError to fail")
	}

	// Log the timings for diagnosis
	t.Logf("RetryableError duration: %v, NonRetryableError duration: %v",
		durationRetryable, durationNonRetryable)

	// The key test: RetryableError should take significantly longer than NonRetryableError
	// because it's being retried multiple times, while NonRetryableError fails immediately

	// Allow for some system variance - RetryableError should be at least 30% longer
	// This is a more reliable test than a fixed time difference
	if float64(durationRetryable) < float64(durationNonRetryable)*1.3 {
		t.Errorf("RetryableError (%v) didn't take significantly longer than NonRetryableError (%v)",
			durationRetryable, durationNonRetryable)
	}

	// ---------- Test 2: The base runner doesn't retry ----------
	// We'll just check that it fails as expected (can't test performance reliably)
	_, baseErr := baseRunner.ResolveAndExecuteFunc(
		modulePath,
		"github.com/test/errors",
		"RetryableError")

	if baseErr == nil {
		t.Error("Expected base runner to fail with RetryableError")
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
