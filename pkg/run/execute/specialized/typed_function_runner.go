package specialized

import (
	"fmt"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// IntegerFunction is a type alias for a function that takes and returns integers
type IntegerFunction func(a, b int) (int, error)

// StringFunction is a type alias for a function that takes and returns strings
type StringFunction func(a string) (string, error)

// MapFunction is a type alias for a function that works with maps
type MapFunction func(data map[string]interface{}) (map[string]interface{}, error)

// TypedFunctionRunner provides type-safe execution for specific function signatures
type TypedFunctionRunner struct {
	*execute.FunctionRunner // Embed the base FunctionRunner
}

// NewTypedFunctionRunner creates a new typed function runner
func NewTypedFunctionRunner(base *execute.FunctionRunner) *TypedFunctionRunner {
	return &TypedFunctionRunner{
		FunctionRunner: base,
	}
}

// ExecuteIntegerFunction executes a function that takes two integers and returns an integer
func (r *TypedFunctionRunner) ExecuteIntegerFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	a, b int) (int, error) {

	result, err := r.ExecuteFunc(module, funcSymbol, a, b)
	if err != nil {
		return 0, err
	}

	// Convert result to integer
	intResult, ok := result.(int)
	if !ok {
		floatResult, ok := result.(float64)
		if !ok {
			return 0, fmt.Errorf("expected integer result, got %T", result)
		}
		intResult = int(floatResult)
	}

	return intResult, nil
}

// ExecuteStringFunction executes a function that takes a string and returns a string
func (r *TypedFunctionRunner) ExecuteStringFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	input string) (string, error) {

	result, err := r.ExecuteFunc(module, funcSymbol, input)
	if err != nil {
		return "", err
	}

	// Convert result to string
	strResult, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("expected string result, got %T", result)
	}

	return strResult, nil
}

// ExecuteMapFunction executes a function that takes a map and returns a map
func (r *TypedFunctionRunner) ExecuteMapFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	input map[string]interface{}) (map[string]interface{}, error) {

	result, err := r.ExecuteFunc(module, funcSymbol, input)
	if err != nil {
		return nil, err
	}

	// Convert result to map
	mapResult, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map result, got %T", result)
	}

	return mapResult, nil
}

// WrapIntegerFunction returns a strongly typed function that executes a Go function
func (r *TypedFunctionRunner) WrapIntegerFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol) IntegerFunction {

	return func(a, b int) (int, error) {
		return r.ExecuteIntegerFunction(module, funcSymbol, a, b)
	}
}

// WrapStringFunction returns a strongly typed function that executes a Go function
func (r *TypedFunctionRunner) WrapStringFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol) StringFunction {

	return func(a string) (string, error) {
		return r.ExecuteStringFunction(module, funcSymbol, a)
	}
}

// WrapMapFunction returns a strongly typed function that executes a Go function
func (r *TypedFunctionRunner) WrapMapFunction(
	module *typesys.Module,
	funcSymbol *typesys.Symbol) MapFunction {

	return func(data map[string]interface{}) (map[string]interface{}, error) {
		return r.ExecuteMapFunction(module, funcSymbol, data)
	}
}

// ResolveAndWrapIntegerFunction resolves a function and returns a strongly typed wrapper
func (r *TypedFunctionRunner) ResolveAndWrapIntegerFunction(
	modulePath, pkgPath, funcName string) (IntegerFunction, error) {

	// Resolve the module and function
	module, err := r.Resolver.ResolveModule(modulePath, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module: %w", err)
	}

	// Find the function symbol
	pkg, ok := module.Packages[pkgPath]
	if !ok {
		return nil, fmt.Errorf("package %s not found", pkgPath)
	}

	var funcSymbol *typesys.Symbol
	for _, sym := range pkg.Symbols {
		if sym.Kind == typesys.KindFunction && sym.Name == funcName {
			funcSymbol = sym
			break
		}
	}

	if funcSymbol == nil {
		return nil, fmt.Errorf("function %s not found in package %s", funcName, pkgPath)
	}

	// Return the wrapped function
	return r.WrapIntegerFunction(module, funcSymbol), nil
}
