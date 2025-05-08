// Package runner provides functionality for running tests with type awareness.
package runner

import (
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/execute"
	"bitspark.dev/go-tree/pkg/testing/common"
	"bitspark.dev/go-tree/pkg/typesys"
)

// Runner implements the TestRunner interface
type Runner struct {
	// Executor for running tests
	Executor execute.ModuleExecutor
}

// NewRunner creates a new test runner
func NewRunner(executor execute.ModuleExecutor) *Runner {
	if executor == nil {
		executor = execute.NewGoExecutor()
	}
	return &Runner{
		Executor: executor,
	}
}

// RunTests runs tests for a module
func (r *Runner) RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error) {
	if mod == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Default to all packages if no path is specified
	if pkgPath == "" {
		pkgPath = "./..."
	}

	// Prepare test flags
	testFlags := make([]string, 0)
	if opts != nil {
		if opts.Verbose {
			testFlags = append(testFlags, "-v")
		}

		if opts.Parallel {
			// This doesn't actually start tests in parallel, but allows them to run
			// in parallel if they call t.Parallel()
			testFlags = append(testFlags, "-parallel=4")
		}

		if len(opts.Tests) > 0 {
			testFlags = append(testFlags, "-run="+strings.Join(opts.Tests, "|"))
		}
	}

	// Execute tests
	execResult, err := r.Executor.ExecuteTest(mod, pkgPath, testFlags...)
	if err != nil {
		// Don't return error here, as it might just indicate test failures
		// The error is already recorded in the result
	}

	// Convert execute.TestResult to TestResult
	result := &common.TestResult{
		Package:       execResult.Package,
		Tests:         execResult.Tests,
		Passed:        execResult.Passed,
		Failed:        execResult.Failed,
		Output:        execResult.Output,
		Error:         execResult.Error,
		TestedSymbols: execResult.TestedSymbols,
		Coverage:      0.0, // We'll calculate this if coverage analysis is requested
	}

	// Calculate coverage if requested
	if r.shouldCalculateCoverage(opts) {
		coverageResult, err := r.AnalyzeCoverage(mod, pkgPath)
		if err == nil && coverageResult != nil {
			result.Coverage = coverageResult.Percentage
		}
	}

	return result, nil
}

// AnalyzeCoverage analyzes test coverage for a module
func (r *Runner) AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error) {
	if mod == nil {
		return nil, fmt.Errorf("module cannot be nil")
	}

	// Default to all packages if no path is specified
	if pkgPath == "" {
		pkgPath = "./..."
	}

	// Run tests with coverage
	testFlags := []string{"-cover", "-coverprofile=coverage.out"}
	execResult, err := r.Executor.ExecuteTest(mod, pkgPath, testFlags...)
	if err != nil {
		// Don't fail completely if tests failed, we might still have partial coverage
		// The error is already in the result
	}

	// Parse coverage output
	coverageResult, err := r.ParseCoverageOutput(execResult.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage output: %w", err)
	}

	// Map coverage data to symbols in the module
	if err := r.MapCoverageToSymbols(mod, coverageResult); err != nil {
		// Just log the error, don't fail completely
		fmt.Printf("Warning: failed to map coverage to symbols: %v\n", err)
	}

	return coverageResult, nil
}

// ParseCoverageOutput parses the output of go test -cover
func (r *Runner) ParseCoverageOutput(output string) (*common.CoverageResult, error) {
	// Initialize coverage result
	result := &common.CoverageResult{
		Files:              make(map[string]float64),
		Functions:          make(map[string]float64),
		UncoveredFunctions: make([]*typesys.Symbol, 0),
	}

	// Look for coverage percentage in the output
	// Example: "coverage: 75.0% of statements"
	coverageRegex := strings.NewReader(`coverage: ([0-9.]+)% of statements`)
	var coveragePercentage float64
	if _, err := fmt.Fscanf(coverageRegex, "coverage: %f%% of statements", &coveragePercentage); err == nil {
		result.Percentage = coveragePercentage
	} else {
		// If we can't parse the overall percentage, default to 0
		result.Percentage = 0.0
	}

	// TODO: Parse more detailed coverage information from the coverage.out file
	// This would involve reading and parsing the file format

	return result, nil
}

// MapCoverageToSymbols maps coverage data to symbols in the module
func (r *Runner) MapCoverageToSymbols(mod *typesys.Module, coverageData *common.CoverageResult) error {
	// This is a placeholder implementation that would be expanded in practice
	// To properly implement this, we'd need to:
	// 1. Parse the coverage.out file to get line-by-line coverage data
	// 2. Map those lines to symbols in the module
	// 3. Calculate per-function coverage percentages
	// 4. Identify uncovered functions

	// For now, we'll just do some basic validation
	if mod == nil || coverageData == nil {
		return fmt.Errorf("module and coverage data must not be nil")
	}

	return nil
}

// shouldCalculateCoverage determines if coverage analysis should be performed
func (r *Runner) shouldCalculateCoverage(opts *common.RunOptions) bool {
	// In a real implementation, we'd check user options to see if coverage is requested
	// For this simplified implementation, we'll just return false
	return false
}

// DefaultRunner creates a test runner with default settings
func DefaultRunner() TestRunner {
	// Use a GoExecutor for now - in a real implementation, we might choose
	// a more appropriate executor based on the environment
	return NewRunner(execute.NewGoExecutor())
}
