package execute

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestTestRunner_ExecuteModuleTests tests executing all tests in a module
func TestTestRunner_ExecuteModuleTests(t *testing.T) {
	// Create mocks
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{},
	}
	materializer := &MockMaterializer{}

	// Create a test runner with the mocks
	runner := NewTestRunner(resolver, materializer)

	// Use a mock executor that returns a known test result
	mockExecutor := &MockExecutor{
		TestResult: &TestResult{
			Package: "github.com/test/simplemath",
			Tests:   []string{"TestAdd", "TestSubtract"},
			Passed:  2,
			Failed:  0,
			Output:  "ok\ngithub.com/test/simplemath\n",
		},
	}
	runner.WithExecutor(mockExecutor)

	// Get a mock module
	module := createMockModule()

	// Execute tests on the module
	result, err := runner.ExecuteModuleTests(module)

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Passed != 2 || result.Failed != 0 {
		t.Errorf("Expected 2 passed tests, 0 failed tests, got: %d passed, %d failed",
			result.Passed, result.Failed)
	}
}

// TestTestRunner_ExecuteSpecificTest tests executing a specific test function
func TestTestRunner_ExecuteSpecificTest(t *testing.T) {
	// Skip this test for now
	t.Skip("Skipping TestTestRunner_ExecuteSpecificTest until implementation is complete")
}

// TestTestRunner_ResolveAndExecuteModuleTests tests resolving a module and running its tests
func TestTestRunner_ResolveAndExecuteModuleTests(t *testing.T) {
	// Skip this test for now
	t.Skip("Skipping TestTestRunner_ResolveAndExecuteModuleTests until implementation is complete")
}
