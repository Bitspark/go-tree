package execute

import (
	"reflect"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestDifferentFunctionTypes uses table-driven testing to verify support for different function types
func TestDifferentFunctionTypes(t *testing.T) {
	// Create the base function runner with mocks
	resolver := &MockResolver{
		Modules: map[string]*typesys.Module{},
	}
	materializer := &MockMaterializer{}
	baseRunner := NewFunctionRunner(resolver, materializer)

	// Create a module for testing
	module := createMockModule()
	addSymbol := typesys.NewSymbol("Add", typesys.KindFunction)
	addSymbol.Package = module.Packages["github.com/test/simplemath"]

	// Setup the mock executor for handling different function types
	mockExecutor := &MockExecutor{
		ExecuteResult: &ExecutionResult{
			StdOut:   "42",
			StdErr:   "",
			ExitCode: 0,
		},
	}
	baseRunner.WithExecutor(mockExecutor)

	// Get a mock processor to handle results
	mockProcessor := &MockResultProcessor{
		ProcessedResult: nil,
	}
	baseRunner.WithProcessor(mockProcessor)

	// Define the table of test cases
	tests := []struct {
		name        string
		returnValue interface{}
	}{
		{"Integer return", 42},
		{"String return", "hello world"},
		{"Boolean return", true},
		{"Float return", 3.14},
		{"Map return", map[string]interface{}{"name": "Alice"}},
		{"Array return", []interface{}{1, 2, 3}},
		{"Nil return", nil},
	}

	// Execute the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock processor to return the expected value
			mockProcessor.ProcessedResult = tt.returnValue

			// Execute the function
			result, err := baseRunner.ExecuteFunc(module, addSymbol, 2, 3)

			// Verify results
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			// Check that the result matches what the mock processor returned
			// Use type-specific comparisons
			switch v := tt.returnValue.(type) {
			case map[string]interface{}:
				// For maps, use reflect.DeepEqual
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Errorf("Expected map result, got %T", result)
					return
				}
				if !reflect.DeepEqual(resultMap, v) {
					t.Errorf("Expected %v, got %v", v, resultMap)
				}
			case []interface{}:
				// For slices, use reflect.DeepEqual
				resultSlice, ok := result.([]interface{})
				if !ok {
					t.Errorf("Expected slice result, got %T", result)
					return
				}
				if !reflect.DeepEqual(resultSlice, v) {
					t.Errorf("Expected %v, got %v", v, resultSlice)
				}
			default:
				// For primitive types, use direct comparison
				if result != tt.returnValue {
					t.Errorf("Expected %v, got %v", tt.returnValue, result)
				}
			}
		})
	}
}

// MockResultProcessor is a mock implementation of ResultProcessor
type MockResultProcessor struct {
	ProcessedResult interface{}
	ProcessedError  error
}

func (p *MockResultProcessor) ProcessFunctionResult(result *ExecutionResult, funcSymbol *typesys.Symbol) (interface{}, error) {
	return p.ProcessedResult, p.ProcessedError
}

func (p *MockResultProcessor) ProcessTestResult(result *ExecutionResult, testSymbol *typesys.Symbol) (*TestResult, error) {
	return &TestResult{
		Passed: 1,
		Failed: 0,
	}, nil
}
