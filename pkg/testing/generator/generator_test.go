package generator

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/module"
)

// createTestFunction creates a module.Function for testing
func createTestFunction(name string, signature string) *module.Function {
	fn := module.NewFunction(name, true, false)

	// Very basic signature parsing - this is just for tests
	if signature == "" || signature == "()" {
		return fn
	}

	// Parse parameters
	paramsEnd := strings.Index(signature, ")")
	if paramsEnd > 0 {
		paramsStr := signature[1:paramsEnd]
		if paramsStr != "" {
			params := strings.Split(paramsStr, ",")
			for _, param := range params {
				param = strings.TrimSpace(param)
				parts := strings.Split(param, " ")

				name := ""
				typeName := parts[len(parts)-1]
				if len(parts) > 1 {
					name = parts[0]
				}

				isVariadic := strings.HasPrefix(typeName, "...")
				if isVariadic {
					typeName = typeName[3:] // Remove "..."
				}

				fn.AddParameter(name, typeName, isVariadic)
			}
		}
	}

	// Parse returns
	if len(signature) > paramsEnd+1 {
		returnStr := strings.TrimSpace(signature[paramsEnd+1:])
		if returnStr != "" {
			// Check if multiple returns in parentheses
			if strings.HasPrefix(returnStr, "(") && strings.HasSuffix(returnStr, ")") {
				returnStr = returnStr[1 : len(returnStr)-1]
				returns := strings.Split(returnStr, ",")
				for _, ret := range returns {
					ret = strings.TrimSpace(ret)
					parts := strings.Split(ret, " ")

					name := ""
					typeName := parts[len(parts)-1]
					if len(parts) > 1 {
						name = parts[0]
					}

					fn.AddResult(name, typeName)
				}
			} else {
				// Single return
				fn.AddResult("", returnStr)
			}
		}
	}

	return fn
}

// TestGenerateTestTemplate tests basic test template generation
func TestGenerateTestTemplate(t *testing.T) {
	generator := NewGenerator()

	// Test for a function with no parameters and no return
	simpleFunc := createTestFunction("DoNothing", "()")

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
	complexFunc := createTestFunction("ProcessUser", "(user *User, options Options) (bool, error)")

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
	pkg := &module.Package{
		Name:      "testpackage",
		Functions: make(map[string]*module.Function),
	}

	// Add functions to the package
	func1 := createTestFunction("Func1", "() error")
	func2 := createTestFunction("Func2", "(input string) (output string, error)")
	func3 := createTestFunction("Func3", "(x, y int) int")
	testFunc3 := createTestFunction("TestFunc3", "")
	benchFunc1 := createTestFunction("BenchmarkFunc1", "")

	pkg.Functions[func1.Name] = func1
	pkg.Functions[func2.Name] = func2
	pkg.Functions[func3.Name] = func3
	pkg.Functions[testFunc3.Name] = testFunc3
	pkg.Functions[benchFunc1.Name] = benchFunc1

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
			fn := createTestFunction("Test"+tc.name, tc.signature)

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
