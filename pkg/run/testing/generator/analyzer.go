package generator

import (
	"fmt"
	"regexp"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Analyzer analyzes code to determine test needs and coverage
type Analyzer struct {
	// Module to analyze
	Module *typesys.Module
}

// NewAnalyzer creates a new code analyzer
func NewAnalyzer(mod *typesys.Module) *Analyzer {
	return &Analyzer{
		Module: mod,
	}
}

// AnalyzePackage analyzes a package to find test patterns and coverage
func (a *Analyzer) AnalyzePackage(pkg *typesys.Package) (*TestPackage, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package cannot be nil")
	}

	// Find all test functions in the package
	testFunctions, err := a.findTestFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to find test functions: %w", err)
	}

	// Map tests to functions
	testMap, err := a.MapTestsToFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to map tests to functions: %w", err)
	}

	// Find test patterns
	patterns, err := a.FindTestPatterns(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to find test patterns: %w", err)
	}

	// Calculate test coverage
	summary, err := a.CalculateTestCoverage(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate test coverage: %w", err)
	}

	// Create the test package
	testPkg := &TestPackage{
		Package:       pkg,
		TestFunctions: testFunctions,
		TestMap:       *testMap,
		Summary:       *summary,
		Patterns:      patterns,
	}

	return testPkg, nil
}

// findTestFunctions finds all test functions in a package
func (a *Analyzer) findTestFunctions(pkg *typesys.Package) ([]TestFunction, error) {
	testFunctions := make([]TestFunction, 0)

	// Regular expressions for detecting test patterns
	tableTestRe := regexp.MustCompile(`(?i)(table|test(?:case|ing)s|tc\s*:=|tc\s*=\s*test)`)
	parallelRe := regexp.MustCompile(`t\.Parallel\(\)`)

	// Find test files (files with _test.go suffix)
	for _, file := range pkg.Files {
		// Only consider test files
		if !strings.HasSuffix(file.Path, "_test.go") {
			continue
		}

		// Find test functions in the file
		for _, sym := range file.Symbols {
			if sym.Kind != typesys.KindFunction {
				continue
			}

			// Check if this is a test function
			if !strings.HasPrefix(sym.Name, "Test") {
				continue
			}

			// Create a test function
			testFunc := TestFunction{
				Name:       sym.Name,
				TargetName: strings.TrimPrefix(sym.Name, "Test"),
				Source:     sym,
			}

			// Get the function body to analyze patterns
			// This depends on having access to the function source code
			// In a real implementation, we'd access the AST or source code
			// For this simplified version, we'll use placeholder logic
			testFunc.IsTableTest = tableTestRe.MatchString("placeholder source code")
			testFunc.IsParallel = parallelRe.MatchString("placeholder source code")
			testFunc.HasBenchmark = false // We'd check if a benchmark exists for this function

			testFunctions = append(testFunctions, testFunc)
		}
	}

	return testFunctions, nil
}

// MapTestsToFunctions matches test functions to the functions they test
func (a *Analyzer) MapTestsToFunctions(pkg *typesys.Package) (*TestMap, error) {
	testMap := &TestMap{
		FunctionToTests: make(map[*typesys.Symbol][]TestFunction),
		Unmapped:        make([]TestFunction, 0),
	}

	// Find all test functions
	testFunctions, err := a.findTestFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to find test functions: %w", err)
	}

	// Map tests to functions
	for _, testFunc := range testFunctions {
		// Try to find the target function
		targetName := testFunc.TargetName
		mapped := false

		// Search in all files of the package
		for _, file := range pkg.Files {
			for _, sym := range file.Symbols {
				if sym.Kind == typesys.KindFunction && sym.Name == targetName {
					// Map this test to the function
					testMap.FunctionToTests[sym] = append(testMap.FunctionToTests[sym], testFunc)
					mapped = true
					break
				}
			}
			if mapped {
				break
			}
		}

		// If not mapped, add to unmapped
		if !mapped {
			testMap.Unmapped = append(testMap.Unmapped, testFunc)
		}
	}

	return testMap, nil
}

// FindTestPatterns identifies common test patterns in a package
func (a *Analyzer) FindTestPatterns(pkg *typesys.Package) ([]TestPattern, error) {
	patterns := make([]TestPattern, 0)

	// Find all test functions
	testFunctions, err := a.findTestFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to find test functions: %w", err)
	}

	// Count pattern occurrences
	tableDriven := TestPattern{
		Name:     "TableDriven",
		Count:    0,
		Examples: make([]string, 0),
	}

	parallel := TestPattern{
		Name:     "Parallel",
		Count:    0,
		Examples: make([]string, 0),
	}

	for _, testFunc := range testFunctions {
		if testFunc.IsTableTest {
			tableDriven.Count++
			if len(tableDriven.Examples) < 3 { // Limit examples to 3
				tableDriven.Examples = append(tableDriven.Examples, testFunc.Name)
			}
		}

		if testFunc.IsParallel {
			parallel.Count++
			if len(parallel.Examples) < 3 { // Limit examples to 3
				parallel.Examples = append(parallel.Examples, testFunc.Name)
			}
		}
	}

	// Add patterns if they occur
	if tableDriven.Count > 0 {
		patterns = append(patterns, tableDriven)
	}

	if parallel.Count > 0 {
		patterns = append(patterns, parallel)
	}

	return patterns, nil
}

// CalculateTestCoverage calculates test coverage for a package
func (a *Analyzer) CalculateTestCoverage(pkg *typesys.Package) (*TestSummary, error) {
	summary := &TestSummary{
		TestedFunctions: make(map[string]bool),
	}

	// Find all test functions
	testFunctions, err := a.findTestFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to find test functions: %w", err)
	}

	// Count test functions
	summary.TotalTests = len(testFunctions)

	// Count table-driven tests
	for _, testFunc := range testFunctions {
		if testFunc.IsTableTest {
			summary.TotalTableTests++
		}
		if testFunc.IsParallel {
			summary.TotalParallelTests++
		}
	}

	// Map tests to functions
	testMap, err := a.MapTestsToFunctions(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to map tests to functions: %w", err)
	}

	// Count functions that have tests
	testedFunctionCount := 0
	totalFunctions := 0

	// Count all functions in the package (excluding test functions)
	for _, file := range pkg.Files {
		// Skip test files
		if strings.HasSuffix(file.Path, "_test.go") {
			continue
		}

		for _, sym := range file.Symbols {
			if sym.Kind == typesys.KindFunction || sym.Kind == typesys.KindMethod {
				totalFunctions++

				// Check if this function has tests
				if _, ok := testMap.FunctionToTests[sym]; ok {
					testedFunctionCount++
					summary.TestedFunctions[sym.Name] = true
				} else {
					summary.TestedFunctions[sym.Name] = false
				}
			}
		}
	}

	// Calculate coverage percentage
	if totalFunctions > 0 {
		summary.TestCoverage = float64(testedFunctionCount) / float64(totalFunctions) * 100.0
	} else {
		summary.TestCoverage = 0.0
	}

	return summary, nil
}

// IdentifyTestedFunctions finds which functions have tests
func (a *Analyzer) IdentifyTestedFunctions(pkg *typesys.Package) (map[string]bool, error) {
	// This is just a wrapper around CalculateTestCoverage to get the TestedFunctions map
	summary, err := a.CalculateTestCoverage(pkg)
	if err != nil {
		return nil, err
	}

	return summary.TestedFunctions, nil
}

// FunctionNeedsTests determines if a function should have tests
func (a *Analyzer) FunctionNeedsTests(sym *typesys.Symbol) bool {
	if sym == nil {
		return false
	}

	// Skip certain function types
	if sym.Kind != typesys.KindFunction && sym.Kind != typesys.KindMethod {
		return false
	}

	// Skip test functions and benchmarks
	if strings.HasPrefix(sym.Name, "Test") || strings.HasPrefix(sym.Name, "Benchmark") {
		return false
	}

	// Skip very simple getters/setters (could be expanded with more logic)
	if a.isSimpleAccessor(sym) {
		return false
	}

	return true
}

// isSimpleAccessor determines if a function is a simple getter or setter
func (a *Analyzer) isSimpleAccessor(sym *typesys.Symbol) bool {
	// This needs access to the function body which depends on AST/source access
	// This is a simplified placeholder implementation

	// Simple length check on name (real implementation would analyze the function body)
	return len(sym.Name) <= 5 && (strings.HasPrefix(sym.Name, "Get") || strings.HasPrefix(sym.Name, "Set"))
}
