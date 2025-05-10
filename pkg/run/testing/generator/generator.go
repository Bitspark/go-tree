package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"strings"
	"text/template"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/testing"
)

// Generator provides functionality for generating test code
type Generator struct {
	// Templates for different test types
	templates map[string]*template.Template

	// Module containing the code to test
	Module *typesys.Module

	// Analyzer for analyzing code
	Analyzer *Analyzer
}

// NewGenerator creates a new test generator
func NewGenerator(mod *typesys.Module) *Generator {
	g := &Generator{
		templates: make(map[string]*template.Template),
		Module:    mod,
		Analyzer:  NewAnalyzer(mod),
	}

	// Initialize the standard templates
	g.templates["basic"] = template.Must(template.New("basic").Parse(basicTestTemplate))
	g.templates["table"] = template.Must(template.New("table").Parse(tableTestTemplate))
	g.templates["parallel"] = template.Must(template.New("parallel").Parse(parallelTestTemplate))
	g.templates["mock"] = template.Must(template.New("mock").Parse(mockTemplate))

	return g
}

// GenerateTests generates tests for a symbol
func (g *Generator) GenerateTests(sym *typesys.Symbol) (*testing.TestSuite, error) {
	if sym == nil {
		return nil, fmt.Errorf("symbol cannot be nil")
	}

	if sym.Kind != typesys.KindFunction && sym.Kind != typesys.KindMethod {
		return nil, fmt.Errorf("can only generate tests for functions and methods, got %s", sym.Kind)
	}

	// Determine the test type to use
	testType := "basic"
	if g.shouldUseTableTest(sym) {
		testType = "table"
	}

	// Generate the test
	testSource, err := g.GenerateTestTemplate(sym, testType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate test template: %w", err)
	}

	// Create the test
	test := &testing.Test{
		Name:       "Test" + sym.Name,
		Target:     sym,
		Type:       testType,
		SourceCode: testSource,
	}

	// Create the suite
	suite := &testing.TestSuite{
		PackageName: sym.Package.Name,
		Tests:       []*testing.Test{test},
		SourceCode:  testSource,
	}

	return suite, nil
}

// GenerateMock generates a mock implementation of an interface
func (g *Generator) GenerateMock(iface *typesys.Symbol) (string, error) {
	if iface == nil {
		return "", fmt.Errorf("interface symbol cannot be nil")
	}

	if iface.Kind != typesys.KindInterface {
		return "", fmt.Errorf("symbol is not an interface: %s", iface.Kind)
	}

	// Extract methods from the interface
	methods, err := g.extractInterfaceMethods(iface)
	if err != nil {
		return "", fmt.Errorf("failed to extract interface methods: %w", err)
	}

	// Create a mock generator
	mockGen := &MockGenerator{
		Interface: iface,
		Methods:   methods,
		MockName:  "Mock" + iface.Name,
	}

	// Generate the mock implementation
	return g.generateMockImpl(mockGen)
}

// GenerateTestData generates test data with correct types
func (g *Generator) GenerateTestData(sym *typesys.Symbol) (interface{}, error) {
	if sym == nil {
		return nil, fmt.Errorf("symbol cannot be nil")
	}

	// Determine the type of the symbol
	typeObj := sym.TypeObj
	if typeObj == nil {
		return nil, fmt.Errorf("symbol has no type information")
	}

	// Generate appropriate test data based on the type
	return g.generateTestDataForType(typeObj, sym.Kind)
}

// shouldUseTableTest determines if a table-driven test is appropriate
func (g *Generator) shouldUseTableTest(sym *typesys.Symbol) bool {
	// Use table tests for functions with parameters
	if funcObj, ok := sym.TypeObj.(*types.Func); ok {
		sig := funcObj.Type().(*types.Signature)
		return sig.Params().Len() > 0
	}
	return false
}

// extractInterfaceMethods extracts method information from an interface
func (g *Generator) extractInterfaceMethods(iface *typesys.Symbol) ([]MockMethod, error) {
	methods := []MockMethod{}

	// Get the interface type
	ifaceType, ok := iface.TypeInfo.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("symbol does not have an interface type")
	}

	// Extract each method
	for i := 0; i < ifaceType.NumMethods(); i++ {
		method := ifaceType.Method(i)
		sig := method.Type().(*types.Signature)

		// Create a mock method
		mockMethod := MockMethod{
			Name:       method.Name(),
			IsVariadic: sig.Variadic(),
			Parameters: []MockParameter{},
			Returns:    []MockReturn{},
		}

		// Extract parameters
		params := sig.Params()
		for j := 0; j < params.Len(); j++ {
			param := params.At(j)
			mockParam := MockParameter{
				Name:       param.Name(),
				Type:       param.Type().String(),
				IsVariadic: sig.Variadic() && j == params.Len()-1,
			}
			mockMethod.Parameters = append(mockMethod.Parameters, mockParam)
		}

		// Extract return values
		results := sig.Results()
		for j := 0; j < results.Len(); j++ {
			result := results.At(j)
			mockReturn := MockReturn{
				Name: result.Name(),
				Type: result.Type().String(),
			}
			mockMethod.Returns = append(mockMethod.Returns, mockReturn)
		}

		methods = append(methods, mockMethod)
	}

	return methods, nil
}

// generateMockImpl generates the mock implementation code
func (g *Generator) generateMockImpl(mockGen *MockGenerator) (string, error) {
	// Prepare template data
	data := struct {
		Package   string
		MockName  string
		Interface string
		Methods   []MockMethod
	}{
		Package:   mockGen.Interface.Package.Name,
		MockName:  mockGen.MockName,
		Interface: mockGen.Interface.Name,
		Methods:   mockGen.Methods,
	}

	// Execute the template
	var buf bytes.Buffer
	if err := g.templates["mock"].Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute mock template: %w", err)
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code if formatting fails
		return buf.String(), fmt.Errorf("failed to format mock code: %w", err)
	}

	return string(formatted), nil
}

// generateTestDataForType generates appropriate test data for a type
func (g *Generator) generateTestDataForType(typeObj types.Object, kind typesys.SymbolKind) (interface{}, error) {
	// This is a simplified implementation that would need to be expanded
	// based on the actual type

	switch t := typeObj.Type().(type) {
	case *types.Basic:
		// Generate data for basic types (int, string, etc.)
		return g.generateBasicTypeTestData(t)
	case *types.Struct:
		// Generate data for structs
		return "struct{}", nil
	case *types.Slice:
		// Generate data for slices
		return "[]T{}", nil
	case *types.Map:
		// Generate data for maps
		return "map[K]V{}", nil
	default:
		// Default placeholder
		return "nil", nil
	}
}

// generateBasicTypeTestData generates test data for basic types
func (g *Generator) generateBasicTypeTestData(t *types.Basic) (string, error) {
	switch t.Kind() {
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		return "42", nil
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return "42", nil
	case types.Float32, types.Float64:
		return "3.14", nil
	case types.Bool:
		return "true", nil
	case types.String:
		return "\"test string\"", nil
	default:
		return "nil", nil
	}
}

// GenerateParameterValues generates values for function parameters
func (g *Generator) GenerateParameterValues(funcSymbol *typesys.Symbol) ([]string, error) {
	if funcSymbol == nil {
		return nil, fmt.Errorf("function symbol cannot be nil")
	}

	funcObj, ok := funcSymbol.TypeObj.(*types.Func)
	if !ok {
		return nil, fmt.Errorf("symbol is not a function")
	}

	sig := funcObj.Type().(*types.Signature)
	params := sig.Params()

	values := make([]string, 0, params.Len())

	for i := 0; i < params.Len(); i++ {
		param := params.At(i)

		// Generate a value based on the parameter type
		value, err := g.generateTestDataForType(param, typesys.KindParameter)
		if err != nil {
			return nil, fmt.Errorf("failed to generate test data for parameter %s: %w", param.Name(), err)
		}

		values = append(values, fmt.Sprintf("%v", value))
	}

	return values, nil
}

// GenerateAssertions generates assertions for function results
func (g *Generator) GenerateAssertions(funcSymbol *typesys.Symbol) (string, error) {
	if funcSymbol == nil {
		return "", fmt.Errorf("function symbol cannot be nil")
	}

	funcObj, ok := funcSymbol.TypeObj.(*types.Func)
	if !ok {
		return "", fmt.Errorf("symbol is not a function")
	}

	sig := funcObj.Type().(*types.Signature)
	results := sig.Results()

	if results.Len() == 0 {
		return "// No assertions needed for void function", nil
	}

	// For a single result, use a direct assertion
	if results.Len() == 1 {
		return "if result != expected {\n\t\tt.Errorf(\"Expected %v, got %v\", expected, result)\n\t}", nil
	}

	// For multiple results, use reflect.DeepEqual
	return "if !reflect.DeepEqual(result, expected) {\n\t\tt.Errorf(\"Expected %v, got %v\", expected, result)\n\t}", nil
}

// GenerateTestTemplate creates a test template for a function
func (g *Generator) GenerateTestTemplate(fn *typesys.Symbol, testType string) (string, error) {
	// Default to basic template if not specified or invalid
	tmpl, exists := g.templates[testType]
	if !exists {
		tmpl = g.templates["basic"]
	}

	// Get function signature information
	funcObj, ok := fn.TypeObj.(*types.Func)
	if !ok {
		return "", fmt.Errorf("symbol is not a function")
	}

	sig := funcObj.Type().(*types.Signature)

	// Generate parameter values
	paramValues, err := g.GenerateParameterValues(fn)
	if err != nil {
		return "", fmt.Errorf("failed to generate parameter values: %w", err)
	}

	// Generate assertions
	assertions, err := g.GenerateAssertions(fn)
	if err != nil {
		return "", fmt.Errorf("failed to generate assertions: %w", err)
	}

	// Prepare template data
	data := struct {
		FunctionName string
		TestName     string
		PackageName  string
		HasParams    bool
		HasReturn    bool
		ParamValues  []string
		Assertions   string
		IsMethod     bool
		ReceiverType string
	}{
		FunctionName: fn.Name,
		TestName:     "Test" + fn.Name,
		PackageName:  fn.Package.Name,
		HasParams:    sig.Params().Len() > 0,
		HasReturn:    sig.Results().Len() > 0,
		ParamValues:  paramValues,
		Assertions:   assertions,
		IsMethod:     fn.Kind == typesys.KindMethod,
	}

	// Handle method receiver if this is a method
	if data.IsMethod {
		// This is a placeholder - we would need to extract the actual receiver type
		data.ReceiverType = "ReceiverType" // Would be extracted from TypeObj
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
func (g *Generator) GenerateMissingTests(pkg *typesys.Package) (map[string]string, error) {
	// Analyze the package to find which functions already have tests
	testPkg, err := g.Analyzer.AnalyzePackage(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze package: %w", err)
	}

	// Get already tested functions
	testedFunctions := make(map[string]bool)
	for fnName := range testPkg.TestMap.FunctionToTests {
		testedFunctions[fnName.Name] = true
	}

	templates := make(map[string]string)

	// Find all functions in the package
	for _, file := range pkg.Files {
		for _, sym := range file.Symbols {
			// Skip non-functions and already tested functions
			if (sym.Kind != typesys.KindFunction && sym.Kind != typesys.KindMethod) ||
				strings.HasPrefix(sym.Name, "Test") ||
				strings.HasPrefix(sym.Name, "Benchmark") ||
				testedFunctions[sym.Name] {
				continue
			}

			// Generate test template
			testTemplate, err := g.GenerateTestTemplate(sym, "basic")
			if err != nil {
				// Skip functions that fail template generation
				continue
			}

			templates[sym.Name] = testTemplate
		}
	}

	return templates, nil
}

// Template for a basic test
const basicTestTemplate = `package {{.PackageName}}_test

import (
	"testing"
	{{if .HasReturn}}
	"reflect"
	{{end}}
	
	"{{.Package.ImportPath}}"
)

func {{.TestName}}(t *testing.T) {
	// Test setup
	{{if .IsMethod}}
	var receiver {{.ReceiverType}}
	{{end}}
	
	{{if .HasParams}}
	// Provide test inputs
	{{range $i, $val := .ParamValues}}
	param{{$i}} := {{$val}}
	{{end}}
	{{end}}
	
	{{if .HasReturn}}
	// Define expected output
	var expected interface{} // TODO: set expected output
	
	// Call function
	{{if .IsMethod}}
	result := receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}param{{$i}}, {{end}})
	{{else}}
	result := {{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}param{{$i}}, {{end}})
	{{end}}
	
	// Verify result
	{{.Assertions}}
	{{else}}
	// Call function
	{{if .IsMethod}}
	receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}param{{$i}}, {{end}})
	{{else}}
	{{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}param{{$i}}, {{end}})
	{{end}}
	
	// Verify expected side effects
	// t.Error("Test not implemented")
	{{end}}
}
`

// Template for a table-driven test
const tableTestTemplate = `package {{.PackageName}}_test

import (
	"testing"
	{{if .HasReturn}}
	"reflect"
	{{end}}
	
	"{{.Package.ImportPath}}"
)

func {{.TestName}}(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		{{if .HasParams}}
		// Input parameters
		{{range $i, $_ := .ParamValues}}
		param{{$i}} interface{}
		{{end}}
		{{end}}
		{{if .HasReturn}}
		expected interface{}
		{{end}}
		wantErr  bool
	}{
		{
			name:     "basic test case",
			{{if .HasParams}}
			// TODO: Add actual test inputs
			{{range $i, $val := .ParamValues}}
			param{{$i}}: {{$val}},
			{{end}}
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
			{{if .IsMethod}}
			var receiver {{.ReceiverType}}
			{{end}}
			
			{{if .HasReturn}}
			// Call function
			{{if .IsMethod}}
			result := receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{else}}
			result := {{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{end}}
			
			// Verify result
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
			{{else}}
			// Call function
			{{if .IsMethod}}
			receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{else}}
			{{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{end}}
			
			// Verify expected side effects
			{{end}}
		})
	}
}
`

// Template for a parallel test
const parallelTestTemplate = `package {{.PackageName}}_test

import (
	"testing"
	{{if .HasReturn}}
	"reflect"
	{{end}}
	
	"{{.Package.ImportPath}}"
)

func {{.TestName}}(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		{{if .HasParams}}
		// Input parameters
		{{range $i, $_ := .ParamValues}}
		param{{$i}} interface{}
		{{end}}
		{{end}}
		{{if .HasReturn}}
		expected interface{}
		{{end}}
	}{
		{
			name:     "basic test case",
			{{if .HasParams}}
			// TODO: Add actual test inputs
			{{range $i, $val := .ParamValues}}
			param{{$i}}: {{$val}},
			{{end}}
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
			
			{{if .IsMethod}}
			var receiver {{.ReceiverType}}
			{{end}}
			
			{{if .HasReturn}}
			// Call function
			{{if .IsMethod}}
			result := receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{else}}
			result := {{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{end}}
			
			// Verify result
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
			{{else}}
			// Call function
			{{if .IsMethod}}
			receiver.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{else}}
			{{.PackageName}}.{{.FunctionName}}({{range $i, $_ := .ParamValues}}tc.param{{$i}}, {{end}})
			{{end}}
			
			// Verify expected side effects
			{{end}}
		})
	}
}
`

// Template for a mock implementation
const mockTemplate = `package {{.Package}}

import (
	"sync"
)

// {{.MockName}} is a mock implementation of the {{.Interface}} interface
type {{.MockName}} struct {
	// Mutex for thread safety
	mu sync.Mutex
	
	{{range .Methods}}
	// Fields to record calls to {{.Name}}
	{{.Name}}Calls     int
	{{.Name}}Called    bool
	{{.Name}}Arguments []struct {
		{{range $i, $param := .Parameters}}
		Param{{$i}} {{$param.Type}}
		{{end}}
	}
	{{if .Returns}}
	// Fields to control return values for {{.Name}}
	{{.Name}}Returns struct {
		{{range $i, $ret := .Returns}}
		Ret{{$i}} {{$ret.Type}}
		{{end}}
	}
	{{end}}
	
	{{end}}
}

// New{{.MockName}} creates a new mock of the {{.Interface}} interface
func New{{.MockName}}() *{{.MockName}} {
	return &{{.MockName}}{}
}

{{range .Methods}}
// {{.Name}} implements the {{$.Interface}} interface
func (m *{{$.MockName}}) {{.Name}}({{range $i, $param := .Parameters}}{{if $i}}, {{end}}{{if $param.Name}}{{$param.Name}} {{end}}{{if $param.IsVariadic}}...{{end}}{{$param.Type}}{{end}}) {{if .Returns}}({{range $i, $ret := .Returns}}{{if $i}}, {{end}}{{if $ret.Name}}{{$ret.Name}} {{end}}{{$ret.Type}}{{end}}){{end}} {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record the call
	m.{{.Name}}Called = true
	m.{{.Name}}Calls++
	
	// Record the arguments
	m.{{.Name}}Arguments = append(m.{{.Name}}Arguments, struct {
		{{range $i, $param := .Parameters}}
		Param{{$i}} {{$param.Type}}
		{{end}}
	}{
		{{range $i, $param := .Parameters}}
		{{if $param.Name}}Param{{$i}}: {{$param.Name}}{{else}}Param{{$i}}: param{{$i}}{{end}},
		{{end}}
	})
	
	{{if .Returns}}
	// Return the configured return values
	return {{range $i, $ret := .Returns}}{{if $i}}, {{end}}m.{{$.Name}}Returns.Ret{{$i}}{{end}}
	{{end}}
}

{{end}}
`
