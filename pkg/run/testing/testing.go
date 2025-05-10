// Package testing provides functionality for generating and running tests
// based on the type system.
package testing

import (
	"bitspark.dev/go-tree/pkg/run/common"
	"fmt"
	"regexp"
	"strconv"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// Re-export common types for backward compatibility
type TestSuite = common.TestSuite
type Test = common.Test
type RunOptions = common.RunOptions
type TestResult = common.TestResult
type CoverageResult = common.CoverageResult

// TestGenerator is temporarily here for backward compatibility
type TestGenerator interface {
	// GenerateTests generates tests for a symbol
	GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error)
	// GenerateMock generates a mock implementation of an interface
	GenerateMock(iface *typesys.Symbol) (string, error)
	// GenerateTestData generates test data with correct types
	GenerateTestData(typ *typesys.Symbol) (interface{}, error)
}

// TestRunner runs tests for Go code
type TestRunner interface {
	// RunTests runs tests for a module
	RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error)

	// AnalyzeCoverage analyzes test coverage for a module
	AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error)
}

// TestExecutor abstracts test execution to avoid import cycles
// The real implementation will be set by the runner package
var testExecutor func(env *materialize.Environment, module *typesys.Module,
	pkgPath string, testFlags ...string) (*common.TestResult, error)

// RegisterTestExecutor sets the implementation for test execution
func RegisterTestExecutor(executor func(env *materialize.Environment, module *typesys.Module,
	pkgPath string, testFlags ...string) (*common.TestResult, error)) {
	testExecutor = executor
}

// RegisterGeneratorFactory registers a factory function for creating test generators.
// This allows the generator package to provide implementations without creating
// import cycles.
func RegisterGeneratorFactory(factory func(*typesys.Module) TestGenerator) {
	generatorFactory = factory
}

// RegisterRunnerFactory registers a factory function for creating test runners.
// This allows the runner package to provide implementations without creating
// import cycles.
func RegisterRunnerFactory(factory func() TestRunner) {
	runnerFactory = factory
}

// processTestOutput parses go test output and builds a test result
func processTestOutput(stdOut, stdErr string, pkgPath string, sym *typesys.Symbol) *common.TestResult {
	output := stdOut + stdErr
	testResult := &common.TestResult{
		Package:       pkgPath,
		Tests:         []string{},
		Passed:        0,
		Failed:        0,
		Output:        output,
		Error:         nil,
		TestedSymbols: []*typesys.Symbol{},
		Coverage:      0.0,
	}

	if sym != nil {
		testResult.TestedSymbols = append(testResult.TestedSymbols, sym)
	}

	// Parse test output to identify tests and results
	testNameRegex := regexp.MustCompile(`--- (PASS|FAIL): (Test\w+) \(`)
	testMatches := testNameRegex.FindAllStringSubmatch(output, -1)

	for _, match := range testMatches {
		status := match[1]
		testName := match[2]

		// Add test to the list
		testResult.Tests = append(testResult.Tests, testName)

		// Update pass/fail counts
		if status == "PASS" {
			testResult.Passed++
		} else {
			testResult.Failed++
		}
	}

	// Try to extract coverage information if present
	coverageRegex := regexp.MustCompile(`coverage: (\d+\.\d+)% of statements`)
	coverageMatch := coverageRegex.FindStringSubmatch(output)
	if len(coverageMatch) > 1 {
		coverage, err := strconv.ParseFloat(coverageMatch[1], 64)
		if err == nil {
			testResult.Coverage = coverage
		}
	}

	return testResult
}

// ExecuteTests generates and runs tests for a symbol.
// This is a convenience function that combines test generation and execution.
func ExecuteTests(mod *typesys.Module, sym *typesys.Symbol, verbose bool) (*common.TestResult, error) {
	// Create a generator using DefaultTestGenerator
	gen := DefaultTestGenerator(mod)
	_, err := gen.GenerateTests(sym)
	if err != nil {
		return nil, err
	}

	// In the future, we would use the generated test suite
	// For now we just verify we can generate tests

	// Create a simple environment for test execution
	env := &materialize.Environment{}

	// Prepare test flags
	testFlags := []string{}
	if verbose {
		testFlags = append(testFlags, "-v")
	}

	// Execute tests using the registered executor
	if testExecutor != nil {
		return testExecutor(env, mod, sym.Package.ImportPath, testFlags...)
	}

	// Fallback to direct execution if no executor is registered
	// Since we can't use ExecuteTest directly anymore, we'll use Execute and process the output
	executor := execute.NewGoExecutor()

	// Prepare the test command
	cmd := append([]string{"go", "test"}, testFlags...)
	if sym.Package.ImportPath != "" {
		cmd = append(cmd, sym.Package.ImportPath)
	}

	// Execute the command
	execResult, err := executor.Execute(env, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tests: %w", err)
	}

	// Process the output to create a TestResult
	return processTestOutput(execResult.StdOut, execResult.StdErr, sym.Package.ImportPath, sym), nil
}

// DefaultTestGenerator provides a factory method for creating a test generator.
func DefaultTestGenerator(mod *typesys.Module) TestGenerator {
	// Create a generator via the adapter pattern.
	// This uses function injection to avoid import cycles.
	// The actual initialization logic is in pkg/testing/init.go
	if generatorFactory != nil {
		return generatorFactory(mod)
	}

	// Fallback implementation if the real factory isn't registered
	return &nullGenerator{mod: mod}
}

// DefaultTestRunner provides a factory method for creating a test runner.
func DefaultTestRunner() TestRunner {
	// Create a runner via the adapter pattern.
	// This uses function injection to avoid import cycles.
	if runnerFactory != nil {
		return runnerFactory()
	}

	// Fallback implementation if the real factory isn't registered
	return &nullRunner{}
}

// GenerateTestsWithDefaults generates tests using the default test generator
func GenerateTestsWithDefaults(mod *typesys.Module, sym *typesys.Symbol) (*common.TestSuite, error) {
	generator := DefaultTestGenerator(mod)
	return generator.GenerateTests(sym)
}

// Internal factory function for creating test generators
var generatorFactory func(*typesys.Module) TestGenerator

// Internal factory function for creating test runners
var runnerFactory func() TestRunner

// nullGenerator is a placeholder implementation of TestGenerator
type nullGenerator struct {
	mod *typesys.Module
}

func (g *nullGenerator) GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error) {
	return &common.TestSuite{
		PackageName: sym.Package.Name,
		Tests:       []*common.Test{},
		SourceCode:  "// Not implemented",
	}, nil
}

func (g *nullGenerator) GenerateMock(iface *typesys.Symbol) (string, error) {
	return "// Not implemented", nil
}

func (g *nullGenerator) GenerateTestData(typ *typesys.Symbol) (interface{}, error) {
	return nil, nil
}

// nullRunner is a placeholder implementation of TestRunner
type nullRunner struct{}

func (r *nullRunner) RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error) {
	return &common.TestResult{
		Package: pkgPath,
		Tests:   []string{},
		Passed:  0,
		Failed:  0,
		Output:  "// Not implemented",
		Error:   nil,
	}, nil
}

func (r *nullRunner) AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error) {
	return &common.CoverageResult{
		Percentage:         0.0,
		Files:              make(map[string]float64),
		Functions:          make(map[string]float64),
		UncoveredFunctions: []*typesys.Symbol{},
	}, nil
}
