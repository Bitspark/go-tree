// Package runner provides functionality for running tests with type awareness.
package runner

import (
	"bitspark.dev/go-tree/pkg/testing/common"
	"bitspark.dev/go-tree/pkg/typesys"
)

// TestRunner runs tests for Go code
type TestRunner interface {
	// RunTests runs tests for a module
	RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error)

	// AnalyzeCoverage analyzes test coverage for a module
	AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error)
}

// RunOptions specifies options for running tests
type RunOptions struct {
	// Verbose output
	Verbose bool

	// Run tests in parallel
	Parallel bool

	// Include benchmarks
	Benchmarks bool

	// Specific tests to run
	Tests []string
}

// TestResult contains the result of running tests
type TestResult struct {
	// Package that was tested
	Package string

	// Tests that were run
	Tests []string

	// Tests that passed
	Passed int

	// Tests that failed
	Failed int

	// Test output
	Output string

	// Error if any occurred during execution
	Error error

	// Symbols that were tested
	TestedSymbols []*typesys.Symbol

	// Test coverage information
	Coverage float64
}

// CoverageResult contains coverage analysis results
type CoverageResult struct {
	// Overall coverage percentage
	Percentage float64

	// Coverage by file
	Files map[string]float64

	// Coverage by function
	Functions map[string]float64

	// Uncovered functions
	UncoveredFunctions []*typesys.Symbol
}
