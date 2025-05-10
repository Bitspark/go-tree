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
	module := createFunctionRunnerMockModule()

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
	// Create mocks
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{},
	}
	materializer := &MockMaterializer{}

	// Create a test runner with the mocks
	runner := NewTestRunner(resolver, materializer)

	// Use a mock executor that returns a known test result for a specific test
	mockExecutor := &MockExecutor{
		TestResult: &TestResult{
			Package: "github.com/test/simplemath",
			Tests:   []string{"TestAdd"},
			Passed:  1,
			Failed:  0,
			Output:  "=== RUN   TestAdd\n--- PASS: TestAdd (0.00s)\nPASS\n",
		},
	}
	runner.WithExecutor(mockExecutor)

	// Get a mock module
	module := createFunctionRunnerMockModule()

	// We need to add a test symbol to the mock module
	testSymbol := &typesys.Symbol{
		Name:    "TestAdd",
		Kind:    typesys.KindFunction,
		Package: module.Packages["github.com/test/simplemath"],
	}
	module.Packages["github.com/test/simplemath"].Symbols["TestAdd"] = testSymbol

	// Execute a specific test on the module
	result, err := runner.ExecuteSpecificTest(
		module,
		"github.com/test/simplemath",
		"TestAdd")

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Passed != 1 || result.Failed != 0 {
		t.Errorf("Expected 1 passed test, 0 failed tests, got: %d passed, %d failed",
			result.Passed, result.Failed)
	}

	if len(result.Tests) != 1 || result.Tests[0] != "TestAdd" {
		t.Errorf("Expected test 'TestAdd', got: %v", result.Tests)
	}
}

// TestTestRunner_ResolveAndExecuteModuleTests tests resolving a module and running its tests
func TestTestRunner_ResolveAndExecuteModuleTests(t *testing.T) {
	// Create a mock module
	mockModule := createFunctionRunnerMockModule()

	// Create a mock resolver that returns our mock module
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{
			"github.com/test/simplemath": mockModule,
		},
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

	// Resolve and execute the module tests
	result, err := runner.ResolveAndExecuteModuleTests("github.com/test/simplemath")

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Passed != 2 || result.Failed != 0 {
		t.Errorf("Expected 2 passed tests, 0 failed tests, got: %d passed, %d failed",
			result.Passed, result.Failed)
	}

	if len(result.Tests) != 2 {
		t.Errorf("Expected 2 tests, got: %d", len(result.Tests))
	}
}
