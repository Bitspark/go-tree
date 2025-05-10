package execute

import (
	"errors"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func TestJsonResultProcessor_ProcessFunctionResult(t *testing.T) {
	// Create a processor
	processor := NewJsonResultProcessor()

	// Set up test cases
	testCases := []struct {
		name           string
		result         *ExecutionResult
		expectedValue  interface{}
		expectError    bool
		errorSubstring string
	}{
		{
			name: "integer result",
			result: &ExecutionResult{
				StdOut:   "42",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  float64(42), // JSON unmarshals numbers as float64
			expectError:    false,
			errorSubstring: "",
		},
		{
			name: "string result",
			result: &ExecutionResult{
				StdOut:   "\"hello world\"",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  "hello world",
			expectError:    false,
			errorSubstring: "",
		},
		{
			name: "boolean result",
			result: &ExecutionResult{
				StdOut:   "true",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  true,
			expectError:    false,
			errorSubstring: "",
		},
		{
			name: "array result",
			result: &ExecutionResult{
				StdOut:   "[1, 2, 3]",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  []interface{}{float64(1), float64(2), float64(3)},
			expectError:    false,
			errorSubstring: "",
		},
		{
			name: "object result",
			result: &ExecutionResult{
				StdOut:   "{\"name\":\"Alice\", \"age\":30}",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue: map[string]interface{}{
				"name": "Alice",
				"age":  float64(30),
			},
			expectError:    false,
			errorSubstring: "",
		},
		{
			name: "execution error",
			result: &ExecutionResult{
				StdOut:   "",
				StdErr:   "Error: something went wrong",
				ExitCode: 1,
				Error:    errors.New("execution failed"),
			},
			expectedValue:  nil,
			expectError:    true,
			errorSubstring: "execution failed",
		},
		{
			name: "invalid JSON",
			result: &ExecutionResult{
				StdOut:   "{invalid json",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  nil,
			expectError:    true,
			errorSubstring: "unmarshal",
		},
		{
			name: "void function success",
			result: &ExecutionResult{
				StdOut:   "{\"success\":true}",
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedValue:  nil,
			expectError:    false,
			errorSubstring: "",
		},
	}

	// Create a mock symbol
	module := createMockModule()
	var addFunc *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			addFunc = sym
			break
		}
	}

	if addFunc == nil {
		t.Fatal("Failed to set up test: could not find Add function")
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := processor.ProcessFunctionResult(tc.result, addFunc)

			// Check error expectations
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.errorSubstring)
				} else if tc.errorSubstring != "" && !strings.Contains(err.Error(), tc.errorSubstring) {
					// We just check if the error message contains the substring
					t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}

			// Check result value
			if !tc.expectError && !deepEqual(result, tc.expectedValue) {
				t.Errorf("Expected result %v, got %v", tc.expectedValue, result)
			}
		})
	}
}

func TestJsonResultProcessor_ProcessTestResult(t *testing.T) {
	// Create a processor
	processor := NewJsonResultProcessor()

	// Set up test cases
	testCases := []struct {
		name           string
		result         *ExecutionResult
		expectedPassed int
		expectedFailed int
		expectError    bool
	}{
		{
			name: "all tests pass",
			result: &ExecutionResult{
				StdOut: `
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestSubtract
--- PASS: TestSubtract (0.00s)
PASS
ok  	github.com/test/simplemath	0.005s	coverage: 75.0% of statements
`,
				StdErr:   "",
				ExitCode: 0,
				Error:    nil,
			},
			expectedPassed: 2,
			expectedFailed: 0,
			expectError:    false,
		},
		{
			name: "some tests fail",
			result: &ExecutionResult{
				StdOut: `
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestSubtract
--- FAIL: TestSubtract (0.00s)
    math_test.go:15: Subtract(5, 3) = 1; want 2
FAIL
exit status 1
FAIL	github.com/test/simplemath	0.005s
`,
				StdErr:   "",
				ExitCode: 1,
				Error:    nil,
			},
			expectedPassed: 1,
			expectedFailed: 1,
			expectError:    false,
		},
		{
			name: "compilation error",
			result: &ExecutionResult{
				StdOut:   "",
				StdErr:   "math.go:10:15: undefined: someFunction",
				ExitCode: 2,
				Error:    errors.New("exit status 2"),
			},
			expectedPassed: 0,
			expectedFailed: 0,
			expectError:    false,
		},
	}

	// Create a mock test symbol
	module := createMockModule()
	var testSymbol *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			testSymbol = sym
			break
		}
	}

	if testSymbol == nil {
		t.Fatal("Failed to set up test: could not find test symbol")
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testResult, err := processor.ProcessTestResult(tc.result, testSymbol)

			// Check error handling
			if tc.expectError {
				if err != nil {
					return // Test passes
				}
				t.Errorf("Expected error, got nil")
			} else if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Verify test counts
			if testResult.Passed != tc.expectedPassed {
				t.Errorf("Expected %d passed tests, got %d", tc.expectedPassed, testResult.Passed)
			}
			if testResult.Failed != tc.expectedFailed {
				t.Errorf("Expected %d failed tests, got %d", tc.expectedFailed, testResult.Failed)
			}
		})
	}
}

// Helper functions for testing

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[0:len(substr)] == substr
}

// deepEqual compares two values for deep equality
// This is a simplified version for tests
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Type-specific comparisons
	switch va := a.(type) {
	case map[string]interface{}:
		// Map comparison
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if bv, ok := vb[k]; !ok || !deepEqual(v, bv) {
				return false
			}
		}
		return true

	case []interface{}:
		// Slice comparison
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i, v := range va {
			if !deepEqual(v, vb[i]) {
				return false
			}
		}
		return true

	default:
		// Simple value comparison (strings, numbers, booleans)
		return a == b
	}
}
