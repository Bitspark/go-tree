package runner

import (
	"errors"
	"testing"

	"bitspark.dev/go-tree/pkg/execute"
	"bitspark.dev/go-tree/pkg/testing/common"
	"bitspark.dev/go-tree/pkg/typesys"
)

// MockExecutor implements execute.ModuleExecutor for testing
type MockExecutor struct {
	ExecuteResult     execute.ExecutionResult
	ExecuteError      error
	ExecuteTestResult execute.TestResult
	ExecuteTestError  error
	ExecuteFuncResult interface{}
	ExecuteFuncError  error
	ExecuteCalled     bool
	ExecuteTestCalled bool
	ExecuteFuncCalled bool
	Args              []string
	PkgPath           string
	TestFlags         []string
}

func (m *MockExecutor) Execute(module *typesys.Module, args ...string) (execute.ExecutionResult, error) {
	m.ExecuteCalled = true
	m.Args = args
	return m.ExecuteResult, m.ExecuteError
}

func (m *MockExecutor) ExecuteTest(module *typesys.Module, pkgPath string, testFlags ...string) (execute.TestResult, error) {
	m.ExecuteTestCalled = true
	m.PkgPath = pkgPath
	m.TestFlags = testFlags
	return m.ExecuteTestResult, m.ExecuteTestError
}

func (m *MockExecutor) ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	m.ExecuteFuncCalled = true
	return m.ExecuteFuncResult, m.ExecuteFuncError
}

func TestNewRunner(t *testing.T) {
	// Test with nil executor
	runner := NewRunner(nil)
	if runner == nil {
		t.Error("NewRunner returned nil")
	}
	if runner.Executor == nil {
		t.Error("NewRunner should create default executor when nil is provided")
	}

	// Test with mock executor
	mockExecutor := &MockExecutor{}
	runner = NewRunner(mockExecutor)
	if runner.Executor != mockExecutor {
		t.Error("NewRunner did not use provided executor")
	}
}

func TestRunTests(t *testing.T) {
	// Test with nil module
	mockExecutor := &MockExecutor{}
	runner := NewRunner(mockExecutor)
	result, err := runner.RunTests(nil, "test/pkg", nil)
	if err == nil {
		t.Error("RunTests should return error for nil module")
	}
	if result != nil {
		t.Error("RunTests should return nil result for nil module")
	}

	// Test with empty package path
	mod := &typesys.Module{Path: "test-module"}
	mockExecutor.ExecuteTestResult = execute.TestResult{
		Package: "./...",
		Tests:   []string{"Test1"},
		Passed:  1,
		Failed:  0,
	}
	result, err = runner.RunTests(mod, "", nil)
	if err != nil {
		t.Errorf("RunTests returned error: %v", err)
	}
	if mockExecutor.PkgPath != "./..." {
		t.Errorf("Expected package path './...', got '%s'", mockExecutor.PkgPath)
	}

	// Test with run options
	mockExecutor.ExecuteTestCalled = false
	opts := &common.RunOptions{
		Verbose:  true,
		Parallel: true,
		Tests:    []string{"TestFunc1", "TestFunc2"},
	}
	_, _ = runner.RunTests(mod, "test/pkg", opts)
	if !mockExecutor.ExecuteTestCalled {
		t.Error("Executor.ExecuteTest not called")
	}
	if mockExecutor.PkgPath != "test/pkg" {
		t.Errorf("Expected package path 'test/pkg', got '%s'", mockExecutor.PkgPath)
	}
	// Check flags
	hasVerbose := false
	hasParallel := false
	hasRun := false
	for _, flag := range mockExecutor.TestFlags {
		if flag == "-v" {
			hasVerbose = true
		}
		if flag == "-parallel=4" {
			hasParallel = true
		}
		if flag == "-run=TestFunc1|TestFunc2" {
			hasRun = true
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

	// Test execution error
	mockExecutor.ExecuteTestError = errors.New("execution error")
	result, err = runner.RunTests(mod, "test/pkg", nil)
	if err != nil {
		t.Errorf("RunTests should not return executor error: %v", err)
	}
	if result == nil {
		t.Error("RunTests should return result even when execution fails")
	}
	if result.Error == nil {
		t.Error("Result should contain executor error")
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
	mockExecutor.ExecuteTestResult = execute.TestResult{
		Package: "./...",
		Output:  "coverage: 75.0% of statements",
	}
	result, err = runner.AnalyzeCoverage(mod, "")
	if err != nil {
		t.Errorf("AnalyzeCoverage returned error: %v", err)
	}
	if mockExecutor.PkgPath != "./..." {
		t.Errorf("Expected package path './...', got '%s'", mockExecutor.PkgPath)
	}

	// Check coverage flags
	hasCoverFlag := false
	hasCoverProfileFlag := false
	for _, flag := range mockExecutor.TestFlags {
		if flag == "-cover" {
			hasCoverFlag = true
		}
		if flag == "-coverprofile=coverage.out" {
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
		t.Error("ParseCoverageOutput returned nil result")
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
		t.Error("ParseCoverageOutput returned nil result")
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

func TestShouldCalculateCoverage(t *testing.T) {
	runner := NewRunner(nil)

	// Test with nil options
	should := runner.shouldCalculateCoverage(nil)
	if should {
		t.Error("shouldCalculateCoverage should return false for nil options")
	}

	// Test with options
	opts := &common.RunOptions{
		Verbose: true,
	}
	should = runner.shouldCalculateCoverage(opts)
	if should {
		t.Error("shouldCalculateCoverage should return false in this implementation")
	}
}

func TestDefaultRunner(t *testing.T) {
	runner := DefaultRunner()
	if runner == nil {
		t.Error("DefaultRunner returned nil")
	}

	// Check if it's the expected type
	_, ok := runner.(*Runner)
	if !ok {
		t.Errorf("DefaultRunner returned unexpected type: %T", runner)
	}
}
