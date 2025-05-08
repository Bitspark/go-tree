package generator

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

// TestAnalyzeTestFunction tests the analysis of individual test functions
func TestAnalyzeTestFunction(t *testing.T) {
	analyzer := NewAnalyzer()

	// Test regular test function
	regularTest := model.GoFunction{
		Name: "TestCreateUser",
		Body: `
			user := CreateUser("test", "test@example.com")
			if user == nil {
				t.Error("Expected user, got nil")
			}
		`,
	}

	result := analyzer.analyzeTestFunction(regularTest)

	if result.Name != "TestCreateUser" {
		t.Errorf("Expected name 'TestCreateUser', got '%s'", result.Name)
	}

	if result.TargetName != "CreateUser" {
		t.Errorf("Expected target name 'CreateUser', got '%s'", result.TargetName)
	}

	if result.IsTableTest {
		t.Error("Regular test incorrectly identified as table test")
	}

	if result.IsParallel {
		t.Error("Regular test incorrectly identified as parallel test")
	}

	// Test table-driven test function
	tableTest := model.GoFunction{
		Name: "TestValidateInput",
		Body: `
			testCases := []struct {
				name     string
				input    string
				expected bool
			}{
				{"valid input", "valid", true},
				{"invalid input", "", false},
			}
			
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					result := ValidateInput(tc.input)
					if result != tc.expected {
						t.Errorf("Expected %v, got %v", tc.expected, result)
					}
				})
			}
		`,
	}

	tableResult := analyzer.analyzeTestFunction(tableTest)

	if !tableResult.IsTableTest {
		t.Error("Table test not correctly identified")
	}

	// Test parallel test function
	parallelTest := model.GoFunction{
		Name: "TestProcessData",
		Body: `
			t.Parallel()
			result := ProcessData("test")
			if result != "processed" {
				t.Errorf("Expected 'processed', got '%s'", result)
			}
		`,
	}

	parallelResult := analyzer.analyzeTestFunction(parallelTest)

	if !parallelResult.IsParallel {
		t.Error("Parallel test not correctly identified")
	}
}

// TestMapTestsToFunctions tests mapping tests to their target functions
func TestMapTestsToFunctions(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create some test functions and a package
	tests := []TestFunction{
		{Name: "TestCreateUser", TargetName: "CreateUser"},
		{Name: "TestValidateInput", TargetName: "ValidateInput"},
		{Name: "TestProcessDataSuccess", TargetName: "ProcessDataSuccess"},
		{Name: "TestGetUserByID", TargetName: "GetUserByID"},
	}

	pkg := &model.GoPackage{
		Functions: []model.GoFunction{
			{Name: "CreateUser"},
			{Name: "validateInput"}, // lowercase first letter
			{Name: "processData"},   // partial match
			{Name: "UnrelatedFunc"},
		},
	}

	// Map tests to functions
	testMap := analyzer.mapTestsToFunctions(tests, pkg)

	// Check direct match
	if len(testMap.FunctionToTests["CreateUser"]) != 1 {
		t.Error("Failed to map TestCreateUser to CreateUser")
	}

	// Check lowercase match
	if len(testMap.FunctionToTests["validateInput"]) != 1 {
		t.Error("Failed to map TestValidateInput to validateInput")
	}

	// Check partial match
	if len(testMap.FunctionToTests["processData"]) != 1 {
		t.Error("Failed to map TestProcessDataSuccess to processData")
	}

	// Check unmapped tests
	if len(testMap.Unmapped) != 1 || testMap.Unmapped[0].Name != "TestGetUserByID" {
		t.Error("Failed to correctly identify unmapped test")
	}
}

// TestCreateTestSummary tests the creation of test summary
func TestCreateTestSummary(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create test functions and test map
	tests := []TestFunction{
		{Name: "TestFunc1", TargetName: "Func1", IsTableTest: true},
		{Name: "TestFunc2", TargetName: "Func2", IsParallel: true},
		{Name: "TestFunc3", TargetName: "Func3", HasBenchmark: true},
	}

	benchmarks := []string{"BenchmarkFunc3", "BenchmarkOther"}

	testMap := TestMap{
		FunctionToTests: map[string][]TestFunction{
			"Func1": {tests[0]},
			"Func2": {tests[1]},
			"Func3": {tests[2]},
		},
		Unmapped: []TestFunction{},
	}

	pkg := &model.GoPackage{
		Functions: []model.GoFunction{
			{Name: "Func1"},
			{Name: "Func2"},
			{Name: "Func3"},
			{Name: "Func4"},          // No test for this function
			{Name: "TestFunc1"},      // Test function itself
			{Name: "BenchmarkFunc3"}, // Benchmark function
		},
	}

	// Create summary
	summary := analyzer.createTestSummary(tests, benchmarks, pkg, testMap)

	// Check counts
	if summary.TotalTests != 3 {
		t.Errorf("Expected 3 total tests, got %d", summary.TotalTests)
	}

	if summary.TotalTableTests != 1 {
		t.Errorf("Expected 1 table test, got %d", summary.TotalTableTests)
	}

	if summary.TotalParallelTests != 1 {
		t.Errorf("Expected 1 parallel test, got %d", summary.TotalParallelTests)
	}

	if summary.TotalBenchmarks != 2 {
		t.Errorf("Expected 2 benchmarks, got %d", summary.TotalBenchmarks)
	}

	// Check test coverage
	expectedCoverage := 75.0 // 3 tested out of 4 testable functions
	if summary.TestCoverage != expectedCoverage {
		t.Errorf("Expected coverage %.1f%%, got %.1f%%", expectedCoverage, summary.TestCoverage)
	}

	// Check tested functions
	for _, funcName := range []string{"Func1", "Func2", "Func3"} {
		if !summary.TestedFunctions[funcName] {
			t.Errorf("Function %s should be marked as tested", funcName)
		}
	}

	if summary.TestedFunctions["Func4"] {
		t.Error("Function Func4 should not be marked as tested")
	}
}

// TestIdentifyTestPatterns tests pattern identification in test functions
func TestIdentifyTestPatterns(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create test functions with different patterns
	tests := []TestFunction{
		{
			Name:        "TestFunc1",
			IsTableTest: true,
			Source:      model.GoFunction{Body: "testCases := []struct{}"},
		},
		{
			Name:       "TestFunc2",
			IsParallel: true,
			Source:     model.GoFunction{Body: "t.Parallel()"},
		},
		{
			Name:         "TestFunc3",
			HasBenchmark: true,
		},
		{
			Name:   "TestFunc4",
			Source: model.GoFunction{Body: "// Given a valid user\n// When we call the function\n// Then it should return true"},
		},
	}

	// Identify patterns
	patterns := analyzer.identifyTestPatterns(tests)

	// Check if all patterns were identified
	expectedPatterns := map[string]bool{
		"Table-Driven Tests":        false,
		"Parallel Tests":            false,
		"Functions with Benchmarks": false,
		"BDD-Style Tests":           false,
	}

	for _, pattern := range patterns {
		expectedPatterns[pattern.Name] = true

		// Check counts
		if pattern.Count != 1 {
			t.Errorf("Expected pattern %s to have count 1, got %d", pattern.Name, pattern.Count)
		}

		// Check that examples were added
		if len(pattern.Examples) != 1 {
			t.Errorf("Expected pattern %s to have 1 example, got %d", pattern.Name, len(pattern.Examples))
		}
	}

	// Check that all patterns were found
	for patternName, found := range expectedPatterns {
		if !found {
			t.Errorf("Pattern %s was not identified", patternName)
		}
	}
}

// TestAnalyzePackage tests the full analysis of a package
func TestAnalyzePackage(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create a test package
	pkg := &model.GoPackage{
		Name: "testpackage",
		Functions: []model.GoFunction{
			{Name: "CreateUser", Signature: "(name string, email string) *User"},
			{Name: "ValidateEmail", Signature: "(email string) bool"},
			{Name: "ProcessData", Signature: "(data []byte) error"},

			// Test functions
			{
				Name: "TestCreateUser",
				Body: "user := CreateUser(\"test\", \"test@example.com\")\nif user == nil {\n\tt.Error(\"Expected user, got nil\")\n}",
			},
			{
				Name: "TestValidateEmail",
				Body: "testCases := []struct{\n\temail string\n\tvalid bool\n}{\n\t{\"test@example.com\", true},\n\t{\"\", false},\n}\nfor _, tc := range testCases {\n\tresult := ValidateEmail(tc.email)\n\tif result != tc.valid {\n\t\tt.Errorf(\"Expected %v, got %v\", tc.valid, result)\n\t}\n}",
			},

			// Benchmark
			{Name: "BenchmarkValidateEmail", Body: "for i := 0; i < b.N; i++ {\n\tValidateEmail(\"test@example.com\")\n}"},
		},
	}

	// Analyze the package
	testPkg := analyzer.AnalyzePackage(pkg, true)

	// Check package name
	if testPkg.PackageName != "testpackage" {
		t.Errorf("Expected package name 'testpackage', got '%s'", testPkg.PackageName)
	}

	// Check test functions
	if len(testPkg.TestFunctions) != 2 {
		t.Errorf("Expected 2 test functions, got %d", len(testPkg.TestFunctions))
	}

	// Check test map
	if len(testPkg.TestMap.FunctionToTests) != 2 {
		t.Errorf("Expected 2 mapped functions, got %d", len(testPkg.TestMap.FunctionToTests))
	}

	if len(testPkg.TestMap.FunctionToTests["CreateUser"]) != 1 {
		t.Error("TestCreateUser not properly mapped")
	}

	if len(testPkg.TestMap.FunctionToTests["ValidateEmail"]) != 1 {
		t.Error("TestValidateEmail not properly mapped")
	}

	if len(testPkg.TestMap.Unmapped) != 0 {
		t.Errorf("Expected 0 unmapped tests, got %d", len(testPkg.TestMap.Unmapped))
	}

	// Check summary
	if testPkg.Summary.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", testPkg.Summary.TotalTests)
	}

	if testPkg.Summary.TotalBenchmarks != 1 {
		t.Errorf("Expected 1 benchmark, got %d", testPkg.Summary.TotalBenchmarks)
	}

	if testPkg.Summary.TotalTableTests != 1 {
		t.Errorf("Expected 1 table test, got %d", testPkg.Summary.TotalTableTests)
	}

	// Check test coverage
	expectedCoverage := 66.67 // 2 tested out of 3 testable functions, rounded
	if testPkg.Summary.TestCoverage < 66.0 || testPkg.Summary.TestCoverage > 67.0 {
		t.Errorf("Expected coverage around %.2f%%, got %.2f%%", expectedCoverage, testPkg.Summary.TestCoverage)
	}

	// Check test patterns
	if len(testPkg.Patterns) < 1 {
		t.Error("Expected at least 1 identified test pattern")
	}

	tablePatternFound := false
	for _, pattern := range testPkg.Patterns {
		if pattern.Name == "Table-Driven Tests" {
			tablePatternFound = true
			break
		}
	}

	if !tablePatternFound {
		t.Error("Table-driven test pattern not identified")
	}
}
