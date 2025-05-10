package execute

import (
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TypeAwareGenerator generates type-aware code for function execution
type TypeAwareGenerator struct{}

// NewTypeAwareGenerator creates a new type-aware generator
func NewTypeAwareGenerator() *TypeAwareGenerator {
	return &TypeAwareGenerator{}
}

// GenerateFunctionWrapper generates a wrapper program to execute a function
func (g *TypeAwareGenerator) GenerateFunctionWrapper(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	args ...interface{}) (string, error) {

	if module == nil || funcSymbol == nil {
		return "", fmt.Errorf("module and function symbol cannot be nil")
	}

	if funcSymbol.Kind != typesys.KindFunction && funcSymbol.Kind != typesys.KindMethod {
		return "", fmt.Errorf("symbol is not a function or method: %s", funcSymbol.Name)
	}

	// Get package information
	pkgPath := funcSymbol.Package.ImportPath
	pkgName := funcSymbol.Package.Name

	// Generate argument conversion
	argValues, argTypes, err := generateArguments(args)
	if err != nil {
		return "", fmt.Errorf("failed to generate arguments: %w", err)
	}

	// Determine return type handling
	hasReturnValues, returnTypes := analyzeReturnTypes(funcSymbol)

	// Prepare template data
	data := struct {
		PackagePath     string
		PackageName     string
		FunctionName    string
		ArgValues       string
		ArgTypes        string
		HasReturnValues bool
		ReturnTypes     string
		IsMethod        bool
		ReceiverType    string
		ModulePath      string
	}{
		PackagePath:     pkgPath,
		PackageName:     pkgName,
		FunctionName:    funcSymbol.Name,
		ArgValues:       argValues,
		ArgTypes:        argTypes,
		HasReturnValues: hasReturnValues,
		ReturnTypes:     returnTypes,
		IsMethod:        funcSymbol.Kind == typesys.KindMethod,
		ReceiverType:    "", // Will be populated if it's a method
		ModulePath:      module.Path,
	}

	// Apply template
	var buf strings.Builder
	tmpl, err := template.New("wrapper").Parse(functionWrapperTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the code
	formattedCode, err := format.Source([]byte(buf.String()))
	if err != nil {
		// If formatting fails, return the unformatted code with a warning
		return buf.String(), fmt.Errorf("generated valid but unformatted code: %w", err)
	}

	return string(formattedCode), nil
}

// GenerateTestWrapper generates a test driver for a specific test function
func (g *TypeAwareGenerator) GenerateTestWrapper(module *typesys.Module, testSymbol *typesys.Symbol) (string, error) {
	if module == nil || testSymbol == nil {
		return "", fmt.Errorf("module and test symbol cannot be nil")
	}

	// This is a placeholder implementation
	// A real implementation would generate a test driver specific to the test function
	return "", fmt.Errorf("test wrapper generation not implemented yet")
}

// Helper functions

// generateArguments converts the provided arguments to Go code strings
func generateArguments(args []interface{}) (string, string, error) {
	var argValues []string
	var argTypes []string

	for i, arg := range args {
		switch v := arg.(type) {
		case string:
			argValues = append(argValues, fmt.Sprintf("%q", v))
			argTypes = append(argTypes, "string")
		case int:
			argValues = append(argValues, fmt.Sprintf("%d", v))
			argTypes = append(argTypes, "int")
		case float64:
			argValues = append(argValues, fmt.Sprintf("%f", v))
			argTypes = append(argTypes, "float64")
		case bool:
			argValues = append(argValues, fmt.Sprintf("%t", v))
			argTypes = append(argTypes, "bool")
		default:
			// For more complex types, use fmt.Sprintf("%#v", v)
			argValues = append(argValues, fmt.Sprintf("%#v", v))
			argTypes = append(argTypes, fmt.Sprintf("interface{} /* arg %d */", i))
		}
	}

	return strings.Join(argValues, ", "), strings.Join(argTypes, ", "), nil
}

// analyzeReturnTypes examines a function symbol to determine its return types
func analyzeReturnTypes(funcSymbol *typesys.Symbol) (bool, string) {
	// This is a simplified implementation
	// A real implementation would extract the return types from the symbol's type information

	// For now, we'll assume all functions return a generic interface{}
	return true, "interface{}"
}

// Template for the function wrapper
const functionWrapperTemplate = `// Generated wrapper for executing {{.FunctionName}}
package main

import (
	"encoding/json"
	"fmt"
	"os"
	
	// Import the package containing the function
	pkg "{{.PackagePath}}"
)

func main() {
	// Call the function
	{{if .HasReturnValues}}
	{{if .IsMethod}}
	// Method execution not fully implemented yet
	fmt.Fprintf(os.Stderr, "Method execution not implemented")
	os.Exit(1)
	{{else}}
	result := pkg.{{.FunctionName}}({{.ArgValues}})
	
	// Encode the result to JSON and print it
	jsonResult, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(string(jsonResult))
	{{end}}
	{{else}}
	// Function has no return values
	pkg.{{.FunctionName}}({{.ArgValues}})
	fmt.Println("{\"success\":true}")
	{{end}}
}
`
