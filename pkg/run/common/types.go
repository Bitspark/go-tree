// Package common provides shared types for the testing packages
package common

import "bitspark.dev/go-tree/pkg/core/typesys"

// TestSuite represents a suite of generated tests
type TestSuite struct {
	// Package name
	PackageName string

	// Tests generated
	Tests []*Test

	// Source code of the test file
	SourceCode string
}

// Test represents a single generated test
type Test struct {
	// Name of the test
	Name string

	// Symbol being tested
	Target *typesys.Symbol

	// Type of test (unit, integration, etc.)
	Type string

	// Source code of the test
	SourceCode string
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
