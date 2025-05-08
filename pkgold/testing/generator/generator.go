package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"bitspark.dev/go-tree/pkgold/core/module"
)

// Generator provides functionality for generating test code
type Generator struct {
	// Templates for different test types
	templates map[string]*template.Template
}

// NewGenerator creates a new test generator
func NewGenerator() *Generator {
	g := &Generator{
		templates: make(map[string]*template.Template),
	}

	// Initialize the standard templates
	g.templates["basic"] = template.Must(template.New("basic").Parse(basicTestTemplate))
	g.templates["table"] = template.Must(template.New("table").Parse(tableTestTemplate))
	g.templates["parallel"] = template.Must(template.New("parallel").Parse(parallelTestTemplate))

	return g
}

// buildFunctionSignature builds a function signature string from a Function object
func buildFunctionSignature(fn *module.Function) string {
	var signature strings.Builder

	// Parameters
	signature.WriteString("(")
	for i, param := range fn.Parameters {
		if i > 0 {
			signature.WriteString(", ")
		}
		if param.Name != "" {
			signature.WriteString(param.Name + " ")
		}
		if param.IsVariadic {
			signature.WriteString("...")
		}
		signature.WriteString(param.Type)
	}
	signature.WriteString(")")

	// Results
	if len(fn.Results) > 0 {
		if len(fn.Results) > 1 {
			signature.WriteString(" (")
			for i, result := range fn.Results {
				if i > 0 {
					signature.WriteString(", ")
				}
				if result.Name != "" {
					signature.WriteString(result.Name + " ")
				}
				signature.WriteString(result.Type)
			}
			signature.WriteString(")")
		} else {
			// Single return value
			if fn.Results[0].Name != "" {
				signature.WriteString(" " + fn.Results[0].Name + " ")
			} else {
				signature.WriteString(" ")
			}
			signature.WriteString(fn.Results[0].Type)
		}
	}

	return signature.String()
}

// GenerateTestTemplate creates a test template for a function
func (g *Generator) GenerateTestTemplate(fn *module.Function, testType string) (string, error) {
	// Default to basic template if not specified or invalid
	tmpl, exists := g.templates[testType]
	if !exists {
		tmpl = g.templates["basic"]
	}

	// Build function signature
	signature := buildFunctionSignature(fn)

	// Prepare template data
	data := struct {
		FunctionName string
		TestName     string
		ReturnType   string
		HasParams    bool
		HasReturn    bool
		Signature    string
	}{
		FunctionName: fn.Name,
		TestName:     "Test" + fn.Name,
		Signature:    signature,
	}

	// Analyze the function parameters and results
	data.HasParams = len(fn.Parameters) > 0
	data.HasReturn = len(fn.Results) > 0

	// Format return type information for template
	if data.HasReturn {
		if len(fn.Results) > 1 {
			var returnBuilder strings.Builder
			returnBuilder.WriteString("(")
			for i, result := range fn.Results {
				if i > 0 {
					returnBuilder.WriteString(", ")
				}
				if result.Name != "" {
					returnBuilder.WriteString(result.Name + " ")
				}
				returnBuilder.WriteString(result.Type)
			}
			returnBuilder.WriteString(")")
			data.ReturnType = returnBuilder.String()
		} else {
			data.ReturnType = fn.Results[0].Type
		}
	}

	// Generate the test template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code if formatting fails
		return buf.String(), fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formattedCode), nil
}

// GenerateMissingTests generates test templates for untested functions
func (g *Generator) GenerateMissingTests(pkg *module.Package, testPkg *TestPackage, testType string) map[string]string {
	templates := make(map[string]string)

	// Get already tested functions
	testedFunctions := make(map[string]bool)
	for fnName := range testPkg.TestMap.FunctionToTests {
		testedFunctions[fnName] = true
	}

	// Generate templates for untested functions
	for _, fn := range pkg.Functions {
		// Skip test functions, benchmarks and functions that already have tests
		if strings.HasPrefix(fn.Name, "Test") ||
			strings.HasPrefix(fn.Name, "Benchmark") ||
			testedFunctions[fn.Name] {
			continue
		}

		// Skip methods (functions with receivers)
		if fn.Receiver != nil {
			continue
		}

		// Generate test template
		testTemplate, err := g.GenerateTestTemplate(fn, testType)
		if err != nil {
			// Skip functions that fail template generation
			continue
		}

		templates[fn.Name] = testTemplate
	}

	return templates
}

// Template for a basic test
const basicTestTemplate = `
func {{.TestName}}(t *testing.T) {
	// TODO: Implement test for {{.FunctionName}}
	{{if .HasParams}}
	// Example usage:
	// result := {{.FunctionName}}(...)
	{{if .HasReturn}}
	// if result != expected {
	//     t.Errorf("Expected %v, got %v", expected, result)
	// }
	{{end}}
	{{else}}
	// Example usage:
	// {{.FunctionName}}()
	{{end}}
	
	t.Error("Test not implemented")
}
`

// Template for a table-driven test
const tableTestTemplate = `
func {{.TestName}}(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		{{if .HasParams}}
		input    interface{} // TODO: Replace with actual input type(s)
		{{end}}
		{{if .HasReturn}}
		expected interface{} // TODO: Replace with actual return type(s)
		{{end}}
		wantErr  bool
	}{
		{
			name:     "basic test case",
			{{if .HasParams}}
			input:    nil, // TODO: Add actual test input
			{{end}}
			{{if .HasReturn}}
			expected: nil, // TODO: Add expected output
			{{end}}
			wantErr:  false,
		},
		// TODO: Add more test cases
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			{{if .HasParams}}
			// TODO: Convert tc.input to appropriate type(s)
			{{end}}
			
			{{if .HasReturn}}
			// TODO: Call function and check results
			// result := {{.FunctionName}}(...)
			// if !reflect.DeepEqual(result, tc.expected) {
			//     t.Errorf("Expected %v, got %v", tc.expected, result)
			// }
			{{else}}
			// TODO: Call function
			// {{.FunctionName}}(...)
			{{end}}
		})
	}
}
`

// Template for a parallel test
const parallelTestTemplate = `
func {{.TestName}}(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		{{if .HasParams}}
		input    interface{} // TODO: Replace with actual input type(s)
		{{end}}
		{{if .HasReturn}}
		expected interface{} // TODO: Replace with actual return type(s)
		{{end}}
	}{
		{
			name:     "basic test case",
			{{if .HasParams}}
			input:    nil, // TODO: Add actual test input
			{{end}}
			{{if .HasReturn}}
			expected: nil, // TODO: Add expected output
			{{end}}
		},
		// TODO: Add more test cases
	}

	// Run test cases in parallel
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run this test case in parallel with others
			
			{{if .HasParams}}
			// TODO: Convert tc.input to appropriate type(s)
			{{end}}
			
			{{if .HasReturn}}
			// TODO: Call function and check results
			// result := {{.FunctionName}}(...)
			// if !reflect.DeepEqual(result, tc.expected) {
			//     t.Errorf("Expected %v, got %v", tc.expected, result)
			// }
			{{else}}
			// TODO: Call function
			// {{.FunctionName}}(...)
			{{end}}
		})
	}
}
`
