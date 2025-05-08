package execute

import (
	"bytes"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// MockModuleExecutor implements ModuleExecutor for testing
type MockModuleExecutor struct {
	ExecuteFn     func(module *typesys.Module, args ...string) (ExecutionResult, error)
	ExecuteTestFn func(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error)
	ExecuteFuncFn func(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error)
}

func (m *MockModuleExecutor) Execute(module *typesys.Module, args ...string) (ExecutionResult, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(module, args...)
	}
	return ExecutionResult{}, nil
}

func (m *MockModuleExecutor) ExecuteTest(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
	if m.ExecuteTestFn != nil {
		return m.ExecuteTestFn(module, pkgPath, testFlags...)
	}
	return TestResult{}, nil
}

func (m *MockModuleExecutor) ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	if m.ExecuteFuncFn != nil {
		return m.ExecuteFuncFn(module, funcSymbol, args...)
	}
	return nil, nil
}

func TestNewExecutionContext(t *testing.T) {
	// Create a dummy module for testing
	module := &typesys.Module{
		Path: "test/module",
	}

	// Create a new execution context
	ctx := NewExecutionContext(module)

	// Verify the context was created correctly
	if ctx == nil {
		t.Fatal("NewExecutionContext returned nil")
	}

	if ctx.Module != module {
		t.Errorf("Expected module %v, got %v", module, ctx.Module)
	}

	if ctx.Files == nil {
		t.Error("Files map should not be nil")
	}

	if len(ctx.Files) != 0 {
		t.Errorf("Expected empty Files map, got %d entries", len(ctx.Files))
	}

	if ctx.Stdout != nil {
		t.Errorf("Expected nil Stdout, got %v", ctx.Stdout)
	}

	if ctx.Stderr != nil {
		t.Errorf("Expected nil Stderr, got %v", ctx.Stderr)
	}
}

func TestExecutionContext_WithOutputCapture(t *testing.T) {
	// Create a dummy module for testing
	module := &typesys.Module{
		Path: "test/module",
	}

	// Create a new execution context
	ctx := NewExecutionContext(module)

	// Set output capture
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ctx.Stdout = stdout
	ctx.Stderr = stderr

	// Verify the output capture was set correctly
	if ctx.Stdout != stdout {
		t.Errorf("Expected Stdout to be %v, got %v", stdout, ctx.Stdout)
	}

	if ctx.Stderr != stderr {
		t.Errorf("Expected Stderr to be %v, got %v", stderr, ctx.Stderr)
	}
}

func TestExecutionContext_Execute(t *testing.T) {
	// This is a placeholder test for the Execute method
	// Currently the implementation is a stub, so we're just testing the interface
	// Once implemented, this test should be expanded

	module := &typesys.Module{
		Path: "test/module",
	}

	ctx := NewExecutionContext(module)
	result, err := ctx.Execute("fmt.Println(\"Hello, World!\")")

	// Since the function is stubbed to return nil, nil
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Future implementation should test these behaviors:
	// 1. Code compilation
	// 2. Type checking
	// 3. Execution
	// 4. Result capturing
	// 5. Error handling
}

func TestExecutionContext_ExecuteInline(t *testing.T) {
	// This is a placeholder test for the ExecuteInline method
	// Currently the implementation is a stub, so we're just testing the interface
	// Once implemented, this test should be expanded

	module := &typesys.Module{
		Path: "test/module",
	}

	ctx := NewExecutionContext(module)
	result, err := ctx.ExecuteInline("fmt.Println(\"Hello, World!\")")

	// Since the function is stubbed to return nil, nil
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Future implementation should test these behaviors:
	// 1. Code execution in current context
	// 2. State preservation
	// 3. Output capturing
	// 4. Error handling
}

func TestExecutionResult(t *testing.T) {
	// Test creating and using ExecutionResult
	result := ExecutionResult{
		Command:  "go run main.go",
		StdOut:   "Hello, World!",
		StdErr:   "",
		ExitCode: 0,
		Error:    nil,
		TypeInfo: map[string]typesys.Symbol{
			"main": {Name: "main"},
		},
	}

	if result.Command != "go run main.go" {
		t.Errorf("Expected Command to be 'go run main.go', got '%s'", result.Command)
	}

	if result.StdOut != "Hello, World!" {
		t.Errorf("Expected StdOut to be 'Hello, World!', got '%s'", result.StdOut)
	}

	if result.StdErr != "" {
		t.Errorf("Expected empty StdErr, got '%s'", result.StdErr)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Expected nil Error, got %v", result.Error)
	}

	if len(result.TypeInfo) == 0 {
		t.Error("Expected non-empty TypeInfo")
	}
}

func TestTestResult(t *testing.T) {
	// Test creating and using TestResult
	symbol := &typesys.Symbol{Name: "TestFunc"}
	result := TestResult{
		Package:       "example/pkg",
		Tests:         []string{"TestFunc1", "TestFunc2"},
		Passed:        1,
		Failed:        1,
		Output:        "PASS: TestFunc1\nFAIL: TestFunc2",
		Error:         nil,
		TestedSymbols: []*typesys.Symbol{symbol},
		Coverage:      75.5,
	}

	if result.Package != "example/pkg" {
		t.Errorf("Expected Package to be 'example/pkg', got '%s'", result.Package)
	}

	expectedTests := []string{"TestFunc1", "TestFunc2"}
	if len(result.Tests) != len(expectedTests) {
		t.Errorf("Expected %d tests, got %d", len(expectedTests), len(result.Tests))
	}

	for i, test := range expectedTests {
		if i >= len(result.Tests) || result.Tests[i] != test {
			t.Errorf("Expected test %d to be '%s', got '%s'", i, test, result.Tests[i])
		}
	}

	if result.Passed != 1 {
		t.Errorf("Expected Passed to be 1, got %d", result.Passed)
	}

	if result.Failed != 1 {
		t.Errorf("Expected Failed to be 1, got %d", result.Failed)
	}

	if !bytes.Contains([]byte(result.Output), []byte("PASS: TestFunc1")) {
		t.Errorf("Expected Output to contain 'PASS: TestFunc1', got '%s'", result.Output)
	}

	if !bytes.Contains([]byte(result.Output), []byte("FAIL: TestFunc2")) {
		t.Errorf("Expected Output to contain 'FAIL: TestFunc2', got '%s'", result.Output)
	}

	if result.Error != nil {
		t.Errorf("Expected nil Error, got %v", result.Error)
	}

	if len(result.TestedSymbols) != 1 || result.TestedSymbols[0] != symbol {
		t.Errorf("Expected TestedSymbols to contain symbol, got %v", result.TestedSymbols)
	}

	if result.Coverage != 75.5 {
		t.Errorf("Expected Coverage to be 75.5, got %f", result.Coverage)
	}
}

func TestModuleExecutor_Interface(t *testing.T) {
	// Create mock executor with custom implementations
	executor := &MockModuleExecutor{}

	// Create dummy module and symbol
	module := &typesys.Module{Path: "test/module"}
	symbol := &typesys.Symbol{Name: "TestFunc"}

	// Setup mock implementations
	expectedResult := ExecutionResult{
		Command:  "go run main.go",
		StdOut:   "Hello, World!",
		ExitCode: 0,
	}

	executor.ExecuteFn = func(m *typesys.Module, args ...string) (ExecutionResult, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if len(args) != 2 || args[0] != "run" || args[1] != "main.go" {
			t.Errorf("Expected args [run main.go], got %v", args)
		}

		return expectedResult, nil
	}

	expectedTestResult := TestResult{
		Package: "test/module",
		Tests:   []string{"TestFunc"},
		Passed:  1,
		Failed:  0,
	}

	executor.ExecuteTestFn = func(m *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if pkgPath != "test/module" {
			t.Errorf("Expected pkgPath 'test/module', got '%s'", pkgPath)
		}

		if len(testFlags) != 1 || testFlags[0] != "-v" {
			t.Errorf("Expected testFlags [-v], got %v", testFlags)
		}

		return expectedTestResult, nil
	}

	executor.ExecuteFuncFn = func(m *typesys.Module, funcSym *typesys.Symbol, args ...interface{}) (interface{}, error) {
		if m != module {
			t.Errorf("Expected module %v, got %v", module, m)
		}

		if funcSym != symbol {
			t.Errorf("Expected symbol %v, got %v", symbol, funcSym)
		}

		if len(args) != 1 || args[0] != "arg1" {
			t.Errorf("Expected args [arg1], got %v", args)
		}

		return "result", nil
	}

	// Execute and verify
	result, err := executor.Execute(module, "run", "main.go")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Command != expectedResult.Command ||
		result.StdOut != expectedResult.StdOut ||
		result.ExitCode != expectedResult.ExitCode {
		t.Errorf("Expected result %v, got %v", expectedResult, result)
	}

	testResult, err := executor.ExecuteTest(module, "test/module", "-v")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if testResult.Package != expectedTestResult.Package ||
		len(testResult.Tests) != len(expectedTestResult.Tests) ||
		testResult.Passed != expectedTestResult.Passed ||
		testResult.Failed != expectedTestResult.Failed {
		t.Errorf("Expected test result %v, got %v", expectedTestResult, testResult)
	}

	funcResult, err := executor.ExecuteFunc(module, symbol, "arg1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if funcResult != "result" {
		t.Errorf("Expected func result 'result', got %v", funcResult)
	}
}
