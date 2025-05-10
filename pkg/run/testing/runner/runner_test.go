package runner

import (
	"bitspark.dev/go-tree/pkg/env"
	"bitspark.dev/go-tree/pkg/run/common"
	"errors"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// MockExecutor implements execute.Executor for testing
type MockExecutor struct {
	ExecuteResult     *execute.ExecutionResult
	ExecuteError      error
	ExecuteFuncResult interface{}
	ExecuteFuncError  error
	ExecuteCalled     bool
	ExecuteFuncCalled bool
	Args              []string
	// We'll store test command info here instead
	LastCommand []string
}

func (m *MockExecutor) Execute(env *env.Environment, command []string) (*execute.ExecutionResult, error) {
	m.ExecuteCalled = true
	m.Args = command
	m.LastCommand = command

	// For tests that expect ExecuteTest, create appropriate output based on command
	if len(command) > 0 && command[0] == "go" && len(command) > 1 && command[1] == "test" {
		// Set up a standard output for test commands
		testOutput := "=== RUN   Test1\n--- PASS: Test1 (0.00s)\nPASS\n"

		// If we're running with coverage, add coverage info
		for _, arg := range command {
			if arg == "-cover" {
				testOutput += "coverage: 75.0% of statements\n"
				break
			}
		}

		// Return as part of ExecutionResult
		return &execute.ExecutionResult{
			Command:  "go test " + command[len(command)-1],
			StdOut:   testOutput,
			StdErr:   "",
			ExitCode: 0,
			Error:    m.ExecuteError,
		}, m.ExecuteError
	}

	return m.ExecuteResult, m.ExecuteError
}

func (m *MockExecutor) ExecuteFunc(env *env.Environment, module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	m.ExecuteFuncCalled = true
	return m.ExecuteFuncResult, m.ExecuteFuncError
}

func TestNewRunner(t *testing.T) {
	// Test with nil executor
	runner := NewRunner(nil)
	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}
	if runner.unifiedRunner == nil || runner.unifiedRunner.Executor == nil {
		t.Error("NewRunner should create default executor when nil is provided")
	}

	// Test with mock executor
	mockExecutor := &MockExecutor{}
	runner = NewRunner(mockExecutor)
	if runner.unifiedRunner.Executor != mockExecutor {
		t.Error("NewRunner did not use provided executor")
	}
}

func TestRunTests(t *testing.T) {
	// Test with nil module
	mockExecutor := &MockExecutor{}
	runner := NewRunner(mockExecutor)

	// Verify runner exists before using it
	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}

	result, err := runner.RunTests(nil, "test/pkg", nil)
	if err == nil {
		t.Error("RunTests should return error for nil module")
	}
	if result != nil {
		t.Error("RunTests should return nil result for nil module")
	}

	// Test with empty package path
	mod := &typesys.Module{Path: "test-module"}
	// Set up appropriate mock response
	mockExecutor.ExecuteResult = &execute.ExecutionResult{
		Command:  "go test ./...",
		StdOut:   "=== RUN   Test1\n--- PASS: Test1 (0.00s)\nPASS\n",
		StdErr:   "",
		ExitCode: 0,
	}

	result, err = runner.RunTests(mod, "", nil)
	if err != nil {
		t.Errorf("RunTests returned error: %v", err)
	}

	// Verify results match expectations
	if result == nil {
		t.Fatal("RunTests returned nil result")
	}
	if result.Package != "./..." {
		t.Errorf("Expected package path './...', got '%s'", result.Package)
	}

	// Check if the right command was executed
	if !mockExecutor.ExecuteCalled {
		t.Error("Execute not called")
	}
	foundPackagePath := false
	for _, arg := range mockExecutor.LastCommand {
		if arg == "./..." {
			foundPackagePath = true
			break
		}
	}
	if !foundPackagePath {
		t.Errorf("Expected package path './...' in command, got: %v", mockExecutor.LastCommand)
	}

	// Test with run options
	mockExecutor.ExecuteCalled = false
	opts := &common.RunOptions{
		Verbose:  true,
		Parallel: true,
		Tests:    []string{"TestFunc1", "TestFunc2"},
	}
	result, _ = runner.RunTests(mod, "test/pkg", opts)

	// Verify results
	if result == nil {
		t.Fatal("RunTests returned nil result")
	}

	if !mockExecutor.ExecuteCalled {
		t.Error("Execute not called")
	}

	// Check flags
	hasVerbose := false
	hasParallel := false
	hasRun := false
	hasPackage := false
	for _, arg := range mockExecutor.LastCommand {
		if arg == "-v" {
			hasVerbose = true
		}
		if arg == "-parallel=4" {
			hasParallel = true
		}
		if arg == "-run=TestFunc1|TestFunc2" {
			hasRun = true
		}
		if arg == "test/pkg" {
			hasPackage = true
		}
	}
	if !hasVerbose {
		t.Error("Expected -v flag")
	}
	if !hasParallel {
		t.Error("Expected -parallel flag")
	}
	if !hasRun {
		t.Error("Expected -run flag with tests")
	}
	if !hasPackage {
		t.Error("Expected test/pkg in command")
	}

	// Test execution error
	mockExecutor.ExecuteError = errors.New("execution error")
	result, err = runner.RunTests(mod, "test/pkg", nil)
	if err == nil {
		t.Error("RunTests should return executor error")
	}
}

func TestAnalyzeCoverage(t *testing.T) {
	// Test with nil module
	mockExecutor := &MockExecutor{}
	runner := NewRunner(mockExecutor)
	result, err := runner.AnalyzeCoverage(nil, "test/pkg")
	if err == nil {
		t.Error("AnalyzeCoverage should return error for nil module")
	}
	if result != nil {
		t.Error("AnalyzeCoverage should return nil result for nil module")
	}

	// Test with empty package path
	mod := &typesys.Module{Path: "test-module"}
	mockExecutor.ExecuteResult = &execute.ExecutionResult{
		Command:  "go test -cover -coverprofile=coverage.out ./...",
		StdOut:   "=== RUN   Test1\n--- PASS: Test1 (0.00s)\nPASS\ncoverage: 75.0% of statements",
		StdErr:   "",
		ExitCode: 0,
	}

	result, err = runner.AnalyzeCoverage(mod, "")
	if err != nil {
		t.Errorf("AnalyzeCoverage returned error: %v", err)
	}

	foundPackagePath := false
	for _, arg := range mockExecutor.LastCommand {
		if arg == "./..." {
			foundPackagePath = true
			break
		}
	}
	if !foundPackagePath {
		t.Errorf("Expected package path './...' in command, got: %v", mockExecutor.LastCommand)
	}

	// Verify the result
	if result == nil {
		t.Fatal("AnalyzeCoverage should return non-nil result")
	}
	if result.Percentage != 75.0 {
		t.Errorf("Expected coverage percentage to be 75.0, got %f", result.Percentage)
	}

	// Check coverage flags
	hasCoverFlag := false
	hasCoverProfileFlag := false
	for _, arg := range mockExecutor.LastCommand {
		if arg == "-cover" {
			hasCoverFlag = true
		}
		if arg == "-coverprofile=coverage.out" {
			hasCoverProfileFlag = true
		}
	}
	if !hasCoverFlag {
		t.Error("Expected -cover flag")
	}
	if !hasCoverProfileFlag {
		t.Error("Expected -coverprofile flag")
	}
}

func TestParseCoverageOutput(t *testing.T) {
	runner := NewRunner(nil)

	// Test with valid coverage output
	output := "coverage: 75.0% of statements"
	result, err := runner.ParseCoverageOutput(output)
	if err != nil {
		t.Errorf("ParseCoverageOutput returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseCoverageOutput returned nil result")
	}
	if result.Percentage != 75.0 {
		t.Errorf("Expected coverage 75.0%%, got %f%%", result.Percentage)
	}

	// Test with no coverage information
	output = "No test files"
	result, err = runner.ParseCoverageOutput(output)
	if err != nil {
		t.Errorf("ParseCoverageOutput returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseCoverageOutput returned nil result")
	}
	if result.Percentage != 0.0 {
		t.Errorf("Expected coverage 0.0%%, got %f%%", result.Percentage)
	}
}

func TestMapCoverageToSymbols(t *testing.T) {
	runner := NewRunner(nil)

	// Test with nil parameters
	err := runner.MapCoverageToSymbols(nil, nil)
	if err == nil {
		t.Error("MapCoverageToSymbols should return error for nil parameters")
	}

	// Test with valid parameters
	mod := &typesys.Module{Path: "test-module"}
	coverage := &common.CoverageResult{
		Percentage: 75.0,
		Files:      make(map[string]float64),
		Functions:  make(map[string]float64),
	}
	err = runner.MapCoverageToSymbols(mod, coverage)
	if err != nil {
		t.Errorf("MapCoverageToSymbols returned error: %v", err)
	}
}

func TestDefaultRunner(t *testing.T) {
	runner := DefaultRunner()
	if runner == nil {
		t.Error("DefaultRunner returned nil")
	}

	// Just verify we get an implementation of TestRunner
	_, ok := runner.(TestRunner)
	if !ok {
		t.Errorf("DefaultRunner returned unexpected type: %T", runner)
	}
}
