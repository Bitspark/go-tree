package generator

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

// TestGenerateTestTemplate tests basic test template generation
func TestGenerateTestTemplate(t *testing.T) {
	generator := NewGenerator()

	// Test for a function with no parameters and no return
	simpleFunc := model.GoFunction{
		Name:      "DoNothing",
		Signature: "()",
	}

	basicTemplate, err := generator.GenerateTestTemplate(simpleFunc, "basic")
	if err != nil {
		t.Fatalf("Failed to generate basic template: %v", err)
	}

	// Check that the template contains the function name
	if !strings.Contains(basicTemplate, "TestDoNothing") {
		t.Error("Generated template doesn't contain test function name")
	}

	if !strings.Contains(basicTemplate, "DoNothing()") {
		t.Error("Generated template doesn't reference the target function correctly")
	}

	// Test for a function with parameters and return value
	complexFunc := model.GoFunction{
		Name:      "ProcessUser",
		Signature: "(user *User, options Options) (bool, error)",
	}

	tableTemplate, err := generator.GenerateTestTemplate(complexFunc, "table")
	if err != nil {
		t.Fatalf("Failed to generate table-driven template: %v", err)
	}

	// Check that the template is for a table-driven test
	if !strings.Contains(tableTemplate, "testCases := []struct") {
		t.Error("Table-driven template doesn't contain test cases declaration")
	}

	if !strings.Contains(tableTemplate, "TestProcessUser") {
		t.Error("Generated template doesn't contain test function name")
	}

	// Test for a parallel test
	parallelTemplate, err := generator.GenerateTestTemplate(complexFunc, "parallel")
	if err != nil {
		t.Fatalf("Failed to generate parallel template: %v", err)
	}

	// Check that the template is for a parallel test
	if !strings.Contains(parallelTemplate, "t.Parallel()") {
		t.Error("Parallel template doesn't contain t.Parallel() call")
	}

	// Test with an invalid template type (should default to basic)
	invalidTemplate, err := generator.GenerateTestTemplate(simpleFunc, "nonexistent")
	if err != nil {
		t.Fatalf("Failed to generate template with invalid type: %v", err)
	}

	// Should fall back to basic template
	if !strings.Contains(invalidTemplate, "TestDoNothing") {
		t.Error("Invalid template type didn't fall back to basic template")
	}
}

// TestGenerateMissingTests tests generating templates for untested functions
func TestGenerateMissingTests(t *testing.T) {
	generator := NewGenerator()

	// Create a package with some functions
	pkg := &model.GoPackage{
		Name: "testpackage",
		Functions: []model.GoFunction{
			{Name: "Func1", Signature: "() error"},                              // No test
			{Name: "Func2", Signature: "(input string) (output string, error)"}, // No test
			{Name: "Func3", Signature: "(x, y int) int"},                        // Has test
			{Name: "TestFunc3"},      // Test function
			{Name: "BenchmarkFunc1"}, // Benchmark
		},
	}

	// Create test package with mapping
	testPkg := &TestPackage{
		PackageName: "testpackage",
		TestMap: TestMap{
			FunctionToTests: map[string][]TestFunction{
				"Func3": {{Name: "TestFunc3", TargetName: "Func3"}},
			},
			Unmapped: []TestFunction{},
		},
	}

	// Generate missing tests
	templates := generator.GenerateMissingTests(pkg, testPkg, "basic")

	// Check that we have templates for the untested functions
	if len(templates) != 2 {
		t.Errorf("Expected 2 missing test templates, got %d", len(templates))
	}

	if _, ok := templates["Func1"]; !ok {
		t.Error("Missing test template for Func1")
	}

	if _, ok := templates["Func2"]; !ok {
		t.Error("Missing test template for Func2")
	}

	if _, ok := templates["Func3"]; ok {
		t.Error("Generated test template for Func3 which already has a test")
	}

	// Check that the templates contain the appropriate signatures
	if tmpl, ok := templates["Func1"]; ok {
		if !strings.Contains(tmpl, "func TestFunc1(t *testing.T)") {
			t.Error("Template for Func1 doesn't have correct function signature")
		}
	}

	if tmpl, ok := templates["Func2"]; ok {
		if !strings.Contains(tmpl, "func TestFunc2(t *testing.T)") {
			t.Error("Template for Func2 doesn't have correct function signature")
		}
	}

	// Test with table-driven template
	tableTemplates := generator.GenerateMissingTests(pkg, testPkg, "table")
	if len(tableTemplates) != 2 {
		t.Errorf("Expected 2 missing table test templates, got %d", len(tableTemplates))
	}

	// Check that the templates are table-driven
	for _, tmpl := range tableTemplates {
		if !strings.Contains(tmpl, "testCases := []struct") {
			t.Error("Table template doesn't contain test cases declaration")
		}
	}
}

// TestVariousSignatureTypes tests template generation for different function signatures
func TestVariousSignatureTypes(t *testing.T) {
	generator := NewGenerator()

	// Test functions with various signature types
	testCases := []struct {
		name      string
		signature string
		hasParams bool
		hasReturn bool
	}{
		{"Empty", "", false, false},
		{"NoParamsNoReturn", "()", false, false},
		{"ParamsNoReturn", "(a, b int)", true, false},
		{"NoParamsReturn", "() error", false, true},
		{"ParamsReturn", "(name string) bool", true, true},
		{"ParamsMultipleReturns", "(x int) (int, error)", true, true},
		{"ComplexParams", "(ctx context.Context, opts ...Option)", true, false},
		{"NamedReturns", "(x, y float64) (sum, product float64)", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fn := model.GoFunction{
				Name:      "Test" + tc.name,
				Signature: tc.signature,
			}

			template, err := generator.GenerateTestTemplate(fn, "basic")
			if err != nil {
				t.Fatalf("Failed to generate template for %s: %v", tc.name, err)
			}

			// Check that the template was generated
			if !strings.Contains(template, "TestTest"+tc.name) {
				t.Errorf("Template for %s doesn't contain correct function name", tc.name)
			}

			// Additional checks could be performed for each case
		})
	}
}
