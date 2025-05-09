// Package generator provides functionality for analyzing and generating tests
// and test-related metrics for Go packages with full type awareness.
package generator

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
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
	Source *typesys.Symbol
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
	// FunctionToTests maps function symbols to their test functions
	FunctionToTests map[*typesys.Symbol][]TestFunction

	// Unmapped contains test functions that couldn't be mapped to a specific function
	Unmapped []TestFunction
}

// TestPackage represents the test analysis for a package
type TestPackage struct {
	// Package is the analyzed package
	Package *typesys.Package

	// TestFunctions is a list of all test functions in the package
	TestFunctions []TestFunction

	// TestMap maps functions to their tests
	TestMap TestMap

	// Summary contains test metrics and summary information
	Summary TestSummary

	// Patterns contains identified test patterns
	Patterns []TestPattern
}

// MockMethod represents a method to be mocked
type MockMethod struct {
	// Name of the method
	Name string

	// Parameters of the method
	Parameters []MockParameter

	// Return values of the method
	Returns []MockReturn

	// Whether the method is variadic
	IsVariadic bool

	// Source symbol
	Source *typesys.Symbol
}

// MockParameter represents a parameter in a mocked method
type MockParameter struct {
	// Name of the parameter
	Name string

	// Type of the parameter
	Type string

	// Whether this is a variadic parameter
	IsVariadic bool
}

// MockReturn represents a return value in a mocked method
type MockReturn struct {
	// Name of the return value (if named)
	Name string

	// Type of the return value
	Type string
}

// MockGenerator handles generation of mock implementations
type MockGenerator struct {
	// Original interface being mocked
	Interface *typesys.Symbol

	// Methods to mock
	Methods []MockMethod

	// Name of the mock struct
	MockName string
}

// TestData represents generated test data for a type
type TestData struct {
	// Original type
	Type *typesys.Symbol

	// Generated data value (as a string representation)
	Value string

	// Whether the data is a zero value
	IsZero bool

	// For struct types, field values
	Fields map[string]TestData

	// For slice/array types, element values
	Elements []TestData
}

// TestTemplate represents a template for a test function
type TestTemplate struct {
	// Name of the test function
	Name string

	// Function being tested
	Target *typesys.Symbol

	// Type of test (basic, table, parallel)
	Type string

	// Test data to use
	TestData []TestData

	// Expected results for test cases
	ExpectedResults []TestData

	// Template text
	Template string
}
