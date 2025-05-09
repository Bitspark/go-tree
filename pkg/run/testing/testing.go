// Package testing provides functionality for generating and running tests
// based on the type system.
package testing

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
	"bitspark.dev/go-tree/pkg/run/testing/common"
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

// ExecuteTests generates and runs tests for a symbol.
// This is a convenience function that combines test generation and execution.
func ExecuteTests(mod *typesys.Module, sym *typesys.Symbol, verbose bool) (*common.TestResult, error) {
	// We'll implement this in terms of the generator and runner packages
	// For now, maintain backwards compatibility with the old implementation

	// Create a generator using DefaultTestGenerator
	gen := DefaultTestGenerator(mod)
	testSuite, err := gen.GenerateTests(sym)
	if err != nil {
		return nil, err
	}

	// TODO: Save the generated tests to the module
	_ = testSuite // Using the variable to avoid linter error until implementation is complete

	// Execute tests
	executor := execute.NewTmpExecutor()

	execResult, err := executor.ExecuteTest(mod, sym.Package.ImportPath, "-v")
	if err != nil {
		return nil, err
	}

	// Convert execute.TestResult to common.TestResult
	result := &common.TestResult{
		Package:       execResult.Package,
		Tests:         execResult.Tests,
		Passed:        execResult.Passed,
		Failed:        execResult.Failed,
		Output:        execResult.Output,
		Error:         execResult.Error,
		TestedSymbols: []*typesys.Symbol{sym},
		Coverage:      0.0, // We'd calculate this from coverage data
	}

	return result, nil
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
