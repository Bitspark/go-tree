// Package runner provides functionality for running tests with type awareness.
package runner

import (
	"bitspark.dev/go-tree/pkg/env"
	"bitspark.dev/go-tree/pkg/run/common"
	"fmt"
	"strconv"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// Runner implements the TestRunner interface
type Runner struct {
	// Unified test runner for internal use
	unifiedRunner *UnifiedTestRunner
}

// NewRunner creates a new test runner
func NewRunner(executor execute.Executor) *Runner {
	if executor == nil {
		executor = execute.NewGoExecutor()
	}
	return &Runner{
		unifiedRunner: NewUnifiedTestRunner(executor, nil, nil),
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

	// Create a simple environment for test execution
	env := &env.Environment{}

	// Execute tests using the unified test runner instead of directly calling executor
	return r.unifiedRunner.ExecuteTest(env, mod, pkgPath, testFlags...)
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

	// Create a simple environment for test execution
	env := &env.Environment{}

	// Run tests with coverage
	testFlags := []string{"-cover", "-coverprofile=coverage.out"}

	// Use unified runner to execute the tests
	execResult, err := r.unifiedRunner.ExecuteTest(env, mod, pkgPath, testFlags...)
	if err != nil {
		// Don't fail completely if tests failed, we might still have partial coverage
		fmt.Printf("Warning: tests failed but continuing with coverage analysis: %v\n", err)
		// Still proceed with the coverage analysis using the partial results
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
	coverageRegex := strings.Index(output, "coverage: ")
	if coverageRegex >= 0 {
		// Extract the substring that contains the coverage info
		subStr := output[coverageRegex:]
		endPercentage := strings.Index(subStr, "%")

		if endPercentage > 0 {
			// Extract just the number part (after "coverage: " and before "%")
			coverageStr := subStr[len("coverage: "):endPercentage]
			// Parse it as a float
			if percentage, err := parseFloat(coverageStr); err == nil {
				result.Percentage = percentage
				return result, nil
			}
		}
	}

	// If we couldn't parse with the specific method above, try the original implementation
	var coveragePercentage float64
	_, err := fmt.Sscanf(output, "coverage: %f%% of statements", &coveragePercentage)
	if err == nil {
		result.Percentage = coveragePercentage
	} else {
		// If we can't parse the overall percentage, try with a substring search
		index := strings.Index(output, "coverage: ")
		if index >= 0 {
			substr := output[index:]
			_, err = fmt.Sscanf(substr, "coverage: %f%% of statements", &coveragePercentage)
			if err == nil {
				result.Percentage = coveragePercentage
			}
		}
	}

	return result, nil
}

// parseFloat is a helper to parse a string to float, handling error cases
func parseFloat(s string) (float64, error) {
	// Check if the string is a valid format before trying to parse
	if !strings.Contains(s, ".") || len(s) == 0 || s[0] == '.' || s[len(s)-1] == '.' {
		return 0.0, fmt.Errorf("invalid float format: %s", s)
	}

	// Use standard string to float conversion
	return strconv.ParseFloat(s, 64)
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
