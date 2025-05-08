package common

import (
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

func TestTestSuite(t *testing.T) {
	// Create a test suite
	suite := &TestSuite{
		PackageName: "testpkg",
		Tests:       []*Test{},
		SourceCode:  "// Test source code",
	}

	if suite.PackageName != "testpkg" {
		t.Errorf("Expected PackageName 'testpkg', got '%s'", suite.PackageName)
	}

	if len(suite.Tests) != 0 {
		t.Errorf("Expected empty Tests slice, got %d tests", len(suite.Tests))
	}

	if suite.SourceCode != "// Test source code" {
		t.Errorf("Expected SourceCode comment, got '%s'", suite.SourceCode)
	}
}

func TestTest(t *testing.T) {
	// Create a symbol
	sym := &typesys.Symbol{
		Name: "TestSymbol",
		Package: &typesys.Package{
			Name: "testpkg",
		},
	}

	// Create a test
	test := &Test{
		Name:       "TestFunction",
		Target:     sym,
		Type:       "unit",
		SourceCode: "func TestFunction(t *testing.T) {}",
	}

	if test.Name != "TestFunction" {
		t.Errorf("Expected Name 'TestFunction', got '%s'", test.Name)
	}

	if test.Target != sym {
		t.Errorf("Expected Target to be the test symbol")
	}

	if test.Type != "unit" {
		t.Errorf("Expected Type 'unit', got '%s'", test.Type)
	}

	if test.SourceCode != "func TestFunction(t *testing.T) {}" {
		t.Errorf("Expected SourceCode to match, got '%s'", test.SourceCode)
	}
}

func TestRunOptions(t *testing.T) {
	// Create run options
	opts := &RunOptions{
		Verbose:    true,
		Parallel:   true,
		Benchmarks: false,
		Tests:      []string{"Test1", "Test2"},
	}

	if !opts.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if !opts.Parallel {
		t.Error("Expected Parallel to be true")
	}

	if opts.Benchmarks {
		t.Error("Expected Benchmarks to be false")
	}

	if len(opts.Tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(opts.Tests))
	}

	if opts.Tests[0] != "Test1" || opts.Tests[1] != "Test2" {
		t.Errorf("Expected Tests to be ['Test1', 'Test2'], got %v", opts.Tests)
	}
}

func TestTestResult(t *testing.T) {
	// Create a symbol
	sym := &typesys.Symbol{
		Name: "TestSymbol",
		Package: &typesys.Package{
			Name: "testpkg",
		},
	}

	// Create a test result
	result := &TestResult{
		Package:       "testpkg",
		Tests:         []string{"Test1", "Test2"},
		Passed:        1,
		Failed:        1,
		Output:        "test output",
		Error:         nil,
		TestedSymbols: []*typesys.Symbol{sym},
		Coverage:      0.75,
	}

	if result.Package != "testpkg" {
		t.Errorf("Expected Package 'testpkg', got '%s'", result.Package)
	}

	if len(result.Tests) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(result.Tests))
	}

	if result.Passed != 1 {
		t.Errorf("Expected Passed 1, got %d", result.Passed)
	}

	if result.Failed != 1 {
		t.Errorf("Expected Failed 1, got %d", result.Failed)
	}

	if result.Output != "test output" {
		t.Errorf("Expected Output 'test output', got '%s'", result.Output)
	}

	if result.Error != nil {
		t.Errorf("Expected Error nil, got %v", result.Error)
	}

	if len(result.TestedSymbols) != 1 || result.TestedSymbols[0] != sym {
		t.Error("Expected TestedSymbols to contain the test symbol")
	}

	if result.Coverage != 0.75 {
		t.Errorf("Expected Coverage 0.75, got %f", result.Coverage)
	}
}

func TestCoverageResult(t *testing.T) {
	// Create a symbol
	sym := &typesys.Symbol{
		Name: "TestSymbol",
		Package: &typesys.Package{
			Name: "testpkg",
		},
	}

	// Create a coverage result
	coverage := &CoverageResult{
		Percentage:         0.85,
		Files:              map[string]float64{"file1.go": 0.9, "file2.go": 0.8},
		Functions:          map[string]float64{"func1": 1.0, "func2": 0.7},
		UncoveredFunctions: []*typesys.Symbol{sym},
	}

	if coverage.Percentage != 0.85 {
		t.Errorf("Expected Percentage 0.85, got %f", coverage.Percentage)
	}

	if len(coverage.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(coverage.Files))
	}

	if coverage.Files["file1.go"] != 0.9 {
		t.Errorf("Expected file1.go coverage 0.9, got %f", coverage.Files["file1.go"])
	}

	if coverage.Files["file2.go"] != 0.8 {
		t.Errorf("Expected file2.go coverage 0.8, got %f", coverage.Files["file2.go"])
	}

	if len(coverage.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(coverage.Functions))
	}

	if coverage.Functions["func1"] != 1.0 {
		t.Errorf("Expected func1 coverage 1.0, got %f", coverage.Functions["func1"])
	}

	if coverage.Functions["func2"] != 0.7 {
		t.Errorf("Expected func2 coverage 0.7, got %f", coverage.Functions["func2"])
	}

	if len(coverage.UncoveredFunctions) != 1 || coverage.UncoveredFunctions[0] != sym {
		t.Error("Expected UncoveredFunctions to contain the test symbol")
	}
}
