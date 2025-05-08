package execute

import (
	"fmt"
	"go/format"
	"go/types"
	"strings"
	"text/template"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TypeAwareCodeGenerator generates code with type checking
type TypeAwareCodeGenerator struct {
	// Module containing the code to execute
	Module *typesys.Module
}

// NewTypeAwareCodeGenerator creates a new code generator for the given module
func NewTypeAwareCodeGenerator(module *typesys.Module) *TypeAwareCodeGenerator {
	return &TypeAwareCodeGenerator{
		Module: module,
	}
}

// GenerateExecWrapper generates code to call a function with proper type checking
func (g *TypeAwareCodeGenerator) GenerateExecWrapper(funcSymbol *typesys.Symbol, args ...interface{}) (string, error) {
	if funcSymbol == nil {
		return "", fmt.Errorf("function symbol cannot be nil")
	}

	if funcSymbol.Kind != typesys.KindFunction && funcSymbol.Kind != typesys.KindMethod {
		return "", fmt.Errorf("symbol %s is not a function or method", funcSymbol.Name)
	}

	// Validate arguments match parameter types
	if err := g.ValidateArguments(funcSymbol, args...); err != nil {
		return "", err
	}

	// Generate argument conversions
	argConversions, err := g.GenerateArgumentConversions(funcSymbol, args...)
	if err != nil {
		return "", err
	}

	// Build the wrapper program template data
	data := struct {
		PackagePath     string
		PackageName     string
		FunctionName    string
		ReceiverType    string
		IsMethod        bool
		ArgConversions  string
		ParamCount      []int
		HasReturnValues bool
		ReturnTypes     string
	}{
		PackagePath:     funcSymbol.Package.ImportPath,
		PackageName:     funcSymbol.Package.Name,
		FunctionName:    funcSymbol.Name,
		IsMethod:        funcSymbol.Kind == typesys.KindMethod,
		ArgConversions:  argConversions,
		ParamCount:      make([]int, len(args)), // Initialize with the number of arguments
		HasReturnValues: false,                  // Will be set below
		ReturnTypes:     "",                     // Will be set below
	}

	// Fill the ParamCount with indices (0, 1, 2, etc.)
	for i := range data.ParamCount {
		data.ParamCount[i] = i
	}

	// Handle method receiver if this is a method
	if data.IsMethod {
		// This is a placeholder - we would need to get actual receiver type from the type info
		// In a real implementation, this would use funcSymbol.TypeObj and type info
		data.ReceiverType = "ReceiverType" // Need to extract from TypeObj
	}

	// Build return type information
	if funcTypeObj, ok := funcSymbol.TypeObj.(*types.Func); ok {
		sig := funcTypeObj.Type().(*types.Signature)

		// Check if the function has return values
		if sig.Results().Len() > 0 {
			data.HasReturnValues = true

			// Build return type string
			var returnTypes []string
			for i := 0; i < sig.Results().Len(); i++ {
				returnTypes = append(returnTypes, sig.Results().At(i).Type().String())
			}

			// If multiple return values, wrap in parentheses
			if len(returnTypes) > 1 {
				data.ReturnTypes = "(" + strings.Join(returnTypes, ", ") + ")"
			} else {
				data.ReturnTypes = returnTypes[0]
			}
		}
	}

	// Create template with a custom function to check if an index is the last one
	funcMap := template.FuncMap{
		"isLast": func(index int, arr []int) bool {
			return index == len(arr)-1
		},
	}

	// Apply the template
	tmpl, err := template.New("execWrapper").Funcs(funcMap).Parse(execWrapperTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	source := buf.String()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		// If formatting fails, return the unformatted code
		return source, fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// ValidateArguments verifies that the provided arguments match the function's parameter types
func (g *TypeAwareCodeGenerator) ValidateArguments(funcSymbol *typesys.Symbol, args ...interface{}) error {
	if funcSymbol.TypeObj == nil {
		return fmt.Errorf("function %s has no type information", funcSymbol.Name)
	}

	// Get the function signature
	funcTypeObj, ok := funcSymbol.TypeObj.(*types.Func)
	if !ok {
		return fmt.Errorf("symbol %s is not a function", funcSymbol.Name)
	}

	sig := funcTypeObj.Type().(*types.Signature)
	params := sig.Params()

	// Check if the number of arguments matches (accounting for variadic functions)
	isVariadic := sig.Variadic()
	minArgs := params.Len()
	if isVariadic {
		minArgs--
	}

	if len(args) < minArgs {
		return fmt.Errorf("not enough arguments: expected at least %d, got %d", minArgs, len(args))
	}

	if !isVariadic && len(args) > params.Len() {
		return fmt.Errorf("too many arguments: expected %d, got %d", params.Len(), len(args))
	}

	// Type checking for individual arguments would go here
	// This is a simplified version that just performs count checking
	// A real implementation would do more sophisticated type compatibility checks

	return nil
}

// GenerateArgumentConversions creates code to convert runtime values to the expected types
func (g *TypeAwareCodeGenerator) GenerateArgumentConversions(funcSymbol *typesys.Symbol, args ...interface{}) (string, error) {
	if funcSymbol.TypeObj == nil {
		return "", fmt.Errorf("function %s has no type information", funcSymbol.Name)
	}

	// Get the function signature
	funcTypeObj, ok := funcSymbol.TypeObj.(*types.Func)
	if !ok {
		return "", fmt.Errorf("symbol %s is not a function", funcSymbol.Name)
	}

	sig := funcTypeObj.Type().(*types.Signature)
	params := sig.Params()
	isVariadic := sig.Variadic()

	var conversions []string

	// Generate conversions for each argument
	// This is a simplified implementation - a real one would generate proper conversion code
	// based on the actual types of the arguments and parameters
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		paramType := param.Type().String()

		if isVariadic && i == params.Len()-1 {
			// Handle variadic parameter
			variadicType := strings.TrimPrefix(paramType, "...") // Remove "..." prefix

			// Generate code to collect remaining arguments into a slice
			conversions = append(conversions, fmt.Sprintf("// Convert variadic arguments to %s", paramType))
			conversions = append(conversions, fmt.Sprintf("var arg%d []%s", i, variadicType))
			conversions = append(conversions, fmt.Sprintf("for _, v := range args[%d:] {", i))
			conversions = append(conversions, fmt.Sprintf("    arg%d = append(arg%d, v.(%s))", i, i, variadicType))
			conversions = append(conversions, "}")

			break // We've handled all remaining arguments as variadic
		} else if i < len(args) {
			// Regular parameter - generate type assertion or conversion
			conversions = append(conversions, fmt.Sprintf("// Convert argument %d to %s", i, paramType))
			conversions = append(conversions, fmt.Sprintf("arg%d := args[%d].(%s)", i, i, paramType))
		}
	}

	return strings.Join(conversions, "\n"), nil
}

// execWrapperTemplate is the template for the function execution wrapper
const execWrapperTemplate = `package main

import (
	"encoding/json"
	"fmt"
	"os"
	
	// Import the package containing the function
	pkg "{{.PackagePath}}"
)

// main function that will call the target function and output the results
func main() {
	// Convert arguments to the proper types
	{{.ArgConversions}}
	
	{{if .HasReturnValues}}
	// Call the function
	{{if .IsMethod}}
	// Need to initialize a receiver of the proper type
	var receiver {{.ReceiverType}}
	result := receiver.{{.FunctionName}}({{range $i := .ParamCount}}arg{{$i}}{{if not (isLast $i $.ParamCount)}}, {{end}}{{end}})
	{{else}}
	result := pkg.{{.FunctionName}}({{range $i := .ParamCount}}arg{{$i}}{{if not (isLast $i $.ParamCount)}}, {{end}}{{end}})
	{{end}}
	
	// Encode the result to JSON and print it
	jsonResult, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonResult))
	{{else}}
	// Call the function with no return values
	{{if .IsMethod}}
	// Need to initialize a receiver of the proper type
	var receiver {{.ReceiverType}}
	receiver.{{.FunctionName}}({{range $i := .ParamCount}}arg{{$i}}{{if not (isLast $i $.ParamCount)}}, {{end}}{{end}})
	{{else}}
	pkg.{{.FunctionName}}({{range $i := .ParamCount}}arg{{$i}}{{if not (isLast $i $.ParamCount)}}, {{end}}{{end}})
	{{end}}
	
	// Signal successful completion
	fmt.Println("{\"success\":true}")
	{{end}}
}`
