// Package generator provides functionality for analyzing and generating tests
// and test-related metrics for Go packages.
package generator

import (
	"bitspark.dev/go-tree/pkgold/core/module"
)

// TestFunction represents a test function with metadata
type TestFunction struct {
	// Name is the full name of the test function (e.g., "TestCreateUser")
	Name string

	// TargetName is the derived name of the target function being tested (e.g., "CreateUser")
	TargetName string

	// IsTableTest indicates whether this is likely a table-driven test
	IsTableTest bool

	// IsParallel indicates whether this test runs in parallel
	IsParallel bool

	// HasBenchmark indicates whether a benchmark exists for the same function
	HasBenchmark bool

	// Source contains the full function definition
	Source module.Function
}

// TestSummary provides summary information about tests in a package
type TestSummary struct {
	// TotalTests is the total number of test functions
	TotalTests int

	// TotalTableTests is the number of table-driven tests
	TotalTableTests int

	// TotalParallelTests is the number of parallel tests
	TotalParallelTests int

	// TotalBenchmarks is the number of benchmark functions
	TotalBenchmarks int

	// TestedFunctions is a map of function names to a boolean indicating whether they have tests
	TestedFunctions map[string]bool

	// TestCoverage is the percentage of functions that have tests (0-100)
	TestCoverage float64
}

// TestPattern represents a recognized test pattern
type TestPattern struct {
	// Name is the name of the pattern (e.g., "TableDriven", "Parallel")
	Name string

	// Count is the number of tests using this pattern
	Count int

	// Examples are function names that use this pattern
	Examples []string
}

// TestMap maps regular functions to their corresponding test functions
type TestMap struct {
	// FunctionToTests maps function names to their test functions
	FunctionToTests map[string][]TestFunction

	// Unmapped contains test functions that couldn't be mapped to a specific function
	Unmapped []TestFunction
}

// TestPackage represents the test analysis for a package
type TestPackage struct {
	// PackageName is the name of the analyzed package
	PackageName string

	// TestFunctions is a list of all test functions in the package
	TestFunctions []TestFunction

	// TestMap maps functions to their tests
	TestMap TestMap

	// Summary contains test metrics and summary information
	Summary TestSummary

	// Patterns contains identified test patterns
	Patterns []TestPattern
}
