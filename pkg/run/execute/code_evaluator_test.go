package execute

import (
	"testing"
)

// TestCodeEvaluator_EvaluateGoCode tests evaluating a simple Go code snippet
func TestCodeEvaluator_EvaluateGoCode(t *testing.T) {
	// Create mocks
	materializer := &MockMaterializer{}

	// Create a code evaluator with the mock
	evaluator := NewCodeEvaluator(materializer)

	// Use a mock executor that returns a known result
	mockExecutor := &MockExecutor{
		ExecuteResult: &ExecutionResult{
			StdOut:   "Hello, World!",
			StdErr:   "",
			ExitCode: 0,
		},
	}
	evaluator.WithExecutor(mockExecutor)

	// Evaluate a simple Go code snippet
	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	result, err := evaluator.EvaluateGoCode(code)

	// Check the result
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.StdOut != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!' output, got: %s", result.StdOut)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}
