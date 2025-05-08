package generator

import (
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/core/module"
)

var (
	// Regular expressions for identifying test patterns
	tableTestRegexp     = regexp.MustCompile(`(?i)(table|test(case|data)s?|cases|fixtures|scenarios|inputs|examples).*(\[\]|\bmap\b)`)
	parallelTestRegexp  = regexp.MustCompile(`t\.Parallel\(\)`)
	benchmarkTestRegexp = regexp.MustCompile(`^Benchmark`)
	testPrefixRegexp    = regexp.MustCompile(`^Test`)
)

// Analyzer provides functionality for analyzing tests in a package
type Analyzer struct{}

// NewAnalyzer creates a new test analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzePackage analyzes a package and extracts test information
func (a *Analyzer) AnalyzePackage(pkg *module.Package, includeTestFiles bool) *TestPackage {
	testPkg := &TestPackage{
		PackageName:   pkg.Name,
		TestFunctions: []TestFunction{},
		TestMap: TestMap{
			FunctionToTests: make(map[string][]TestFunction),
			Unmapped:        []TestFunction{},
		},
		Summary: TestSummary{
			TestedFunctions: make(map[string]bool),
		},
		Patterns: []TestPattern{},
	}

	// If there are no test files, there's nothing to analyze
	if !includeTestFiles {
		return testPkg
	}

	// Extract tests and benchmarks
	var tests []TestFunction
	var benchmarks []string

	for _, fn := range pkg.Functions {
		if benchmarkTestRegexp.MatchString(fn.Name) {
			// This is a benchmark
			benchmarks = append(benchmarks, fn.Name)
			continue
		}

		if testPrefixRegexp.MatchString(fn.Name) {
			// This is a test function
			test := a.analyzeTestFunction(fn)
			tests = append(tests, test)
		}
	}

	// Mark tests that have benchmarks
	for i := range tests {
		targetName := tests[i].TargetName
		for _, benchName := range benchmarks {
			if strings.HasPrefix(benchName, "Benchmark"+targetName) {
				tests[i].HasBenchmark = true
				break
			}
		}
	}

	// Store test functions in the test package
	testPkg.TestFunctions = tests

	// Map tests to their target functions
	testPkg.TestMap = a.mapTestsToFunctions(tests, pkg)

	// Calculate test summary
	testPkg.Summary = a.createTestSummary(tests, benchmarks, pkg, testPkg.TestMap)

	// Identify common test patterns
	testPkg.Patterns = a.identifyTestPatterns(tests)

	return testPkg
}

// analyzeTestFunction analyzes a single test function
func (a *Analyzer) analyzeTestFunction(fn *module.Function) TestFunction {
	test := TestFunction{
		Name:   fn.Name,
		Source: *fn,
	}

	// Extract the target function name from the test name
	if testPrefixRegexp.MatchString(fn.Name) {
		test.TargetName = fn.Name[4:] // Remove "Test" prefix
	}

	// Check if it's a table test
	if fn.Body != "" && tableTestRegexp.MatchString(fn.Body) {
		test.IsTableTest = true
	}

	// Check if it's a parallel test
	if fn.Body != "" && parallelTestRegexp.MatchString(fn.Body) {
		test.IsParallel = true
	}

	return test
}

// mapTestsToFunctions maps test functions to their target functions
func (a *Analyzer) mapTestsToFunctions(tests []TestFunction, pkg *module.Package) TestMap {
	result := TestMap{
		FunctionToTests: make(map[string][]TestFunction),
		Unmapped:        []TestFunction{},
	}

	// Get all function names
	functionNames := make(map[string]bool)
	for fnName := range pkg.Functions {
		functionNames[fnName] = true
	}

	// For each test, try to find a matching function
	for _, test := range tests {
		mapped := false

		// Direct match (TestFoo -> Foo)
		if functionNames[test.TargetName] {
			result.FunctionToTests[test.TargetName] = append(
				result.FunctionToTests[test.TargetName], test)
			mapped = true
			continue
		}

		// Try lowercase first letter (TestFoo -> foo)
		if len(test.TargetName) > 0 {
			lowerTarget := strings.ToLower(test.TargetName[:1]) + test.TargetName[1:]
			if functionNames[lowerTarget] {
				result.FunctionToTests[lowerTarget] = append(
					result.FunctionToTests[lowerTarget], test)
				mapped = true
				continue
			}
		}

		// Try package level functions that might be split across multiple tests
		// e.g., TestFooSuccess and TestFooError -> foo
		for fnName := range functionNames {
			if strings.HasPrefix(strings.ToLower(test.TargetName), strings.ToLower(fnName)) {
				result.FunctionToTests[fnName] = append(
					result.FunctionToTests[fnName], test)
				mapped = true
				break
			}
		}

		// If we couldn't map it, add to unmapped
		if !mapped {
			result.Unmapped = append(result.Unmapped, test)
		}
	}

	return result
}

// createTestSummary calculates test coverage and other statistics
func (a *Analyzer) createTestSummary(tests []TestFunction, benchmarks []string, pkg *module.Package, testMap TestMap) TestSummary {
	summary := TestSummary{
		TotalTests:      len(tests),
		TotalBenchmarks: len(benchmarks),
		TestedFunctions: make(map[string]bool),
	}

	// Count table tests and parallel tests
	for _, test := range tests {
		if test.IsTableTest {
			summary.TotalTableTests++
		}
		if test.IsParallel {
			summary.TotalParallelTests++
		}
	}

	// Mark which functions have tests
	for fnName := range testMap.FunctionToTests {
		summary.TestedFunctions[fnName] = true
	}

	// Count testable functions (excluding tests and benchmarks)
	var testableCount int
	for fnName, fn := range pkg.Functions {
		if !testPrefixRegexp.MatchString(fnName) && !benchmarkTestRegexp.MatchString(fnName) && !fn.IsMethod {
			testableCount++
		}
	}

	// Calculate test coverage
	if testableCount > 0 {
		summary.TestCoverage = float64(len(summary.TestedFunctions)) / float64(testableCount) * 100
	}

	return summary
}

// identifyTestPatterns identifies common test patterns in the package
func (a *Analyzer) identifyTestPatterns(tests []TestFunction) []TestPattern {
	patterns := make(map[string]*TestPattern)

	// Check for table-driven tests
	if tableTests := countPatternTests(tests, func(t TestFunction) bool { return t.IsTableTest }); tableTests > 0 {
		patterns["TableDriven"] = &TestPattern{
			Name:  "Table-Driven Tests",
			Count: tableTests,
		}
	}

	// Check for parallel tests
	if parallelTests := countPatternTests(tests, func(t TestFunction) bool { return t.IsParallel }); parallelTests > 0 {
		patterns["Parallel"] = &TestPattern{
			Name:  "Parallel Tests",
			Count: parallelTests,
		}
	}

	// Check for benchmark coverage
	if benchmarkedTests := countPatternTests(tests, func(t TestFunction) bool { return t.HasBenchmark }); benchmarkedTests > 0 {
		patterns["Benchmarked"] = &TestPattern{
			Name:  "Functions with Benchmarks",
			Count: benchmarkedTests,
		}
	}

	// Check for BDD-style tests (Given/When/Then or similar)
	bddRegex := regexp.MustCompile(`(?i)(given|when|then|should|expect|assert)`)
	if bddTests := countPatternTests(tests, func(t TestFunction) bool {
		return t.Source.Body != "" && bddRegex.MatchString(t.Source.Body)
	}); bddTests > 0 {
		patterns["BDD"] = &TestPattern{
			Name:  "BDD-Style Tests",
			Count: bddTests,
		}
	}

	// Add examples for each pattern
	for _, test := range tests {
		if test.IsTableTest && patterns["TableDriven"] != nil {
			if len(patterns["TableDriven"].Examples) < 3 {
				patterns["TableDriven"].Examples = append(patterns["TableDriven"].Examples, test.Name)
			}
		}
		if test.IsParallel && patterns["Parallel"] != nil {
			if len(patterns["Parallel"].Examples) < 3 {
				patterns["Parallel"].Examples = append(patterns["Parallel"].Examples, test.Name)
			}
		}
		if test.HasBenchmark && patterns["Benchmarked"] != nil {
			if len(patterns["Benchmarked"].Examples) < 3 {
				patterns["Benchmarked"].Examples = append(patterns["Benchmarked"].Examples, test.Name)
			}
		}
		if patterns["BDD"] != nil && test.Source.Body != "" && bddRegex.MatchString(test.Source.Body) {
			if len(patterns["BDD"].Examples) < 3 {
				patterns["BDD"].Examples = append(patterns["BDD"].Examples, test.Name)
			}
		}
	}

	// Convert map to slice
	var result []TestPattern
	for _, pattern := range patterns {
		result = append(result, *pattern)
	}

	return result
}

// countPatternTests counts the number of tests that match a pattern
func countPatternTests(tests []TestFunction, matcher func(TestFunction) bool) int {
	count := 0
	for _, test := range tests {
		if matcher(test) {
			count++
		}
	}
	return count
}
