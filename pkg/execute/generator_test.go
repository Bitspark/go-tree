package execute

import (
	"go/types"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// mockFunction creates a mock function symbol with type information for testing
func mockFunction(t *testing.T, name string, params int, returns int) *typesys.Symbol {
	// Create a basic symbol
	sym := &typesys.Symbol{
		ID:       name,
		Name:     name,
		Kind:     typesys.KindFunction,
		Exported: true,
		Package: &typesys.Package{
			ImportPath: "example.com/test",
			Name:       "test",
		},
	}

	// Create a simple mock function type
	paramVars := createTupleType(params)
	resultVars := createTupleType(returns)
	signature := types.NewSignature(nil, paramVars, resultVars, false)

	objFunc := types.NewFunc(0, nil, name, signature)
	sym.TypeObj = objFunc

	return sym
}

// createTupleType creates a simple tuple with n string parameters for testing
func createTupleType(n int) *types.Tuple {
	vars := make([]*types.Var, n)
	strType := types.Typ[types.String]

	for i := 0; i < n; i++ {
		vars[i] = types.NewParam(0, nil, "", strType)
	}

	return types.NewTuple(vars...)
}

// TestNewTypeAwareCodeGenerator tests creation of a new code generator
func TestNewTypeAwareCodeGenerator(t *testing.T) {
	module := &typesys.Module{
		Path: "example.com/test",
	}

	generator := NewTypeAwareCodeGenerator(module)

	if generator == nil {
		t.Fatal("NewTypeAwareCodeGenerator returned nil")
	}

	if generator.Module != module {
		t.Errorf("Expected module to be set correctly")
	}
}

// TestGenerateExecWrapper tests generation of function execution wrapper code
func TestGenerateExecWrapper(t *testing.T) {
	module := &typesys.Module{
		Path: "example.com/test",
	}

	generator := NewTypeAwareCodeGenerator(module)

	// Test with nil symbol
	_, err := generator.GenerateExecWrapper(nil)
	if err == nil {
		t.Error("Expected error for nil symbol, got nil")
	}

	// Test with non-function symbol
	nonFuncSymbol := &typesys.Symbol{
		Name: "NotAFunction",
		Kind: typesys.KindStruct,
	}

	_, err = generator.GenerateExecWrapper(nonFuncSymbol)
	if err == nil {
		t.Error("Expected error for non-function symbol, got nil")
	}

	// Test with function symbol but no type information
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
		Package: &typesys.Package{
			ImportPath: "example.com/test",
			Name:       "test",
		},
	}

	_, err = generator.GenerateExecWrapper(funcSymbol)
	if err == nil {
		t.Error("Expected error for function without type info, got nil")
	}

	// Test with a properly mocked function symbol
	mockFuncSymbol := mockFunction(t, "TestFunc", 2, 1)

	// Provide the required arguments to match the function signature
	code, err := generator.GenerateExecWrapper(mockFuncSymbol, "test1", "test2")
	if err != nil {
		t.Errorf("GenerateExecWrapper returned error: %v", err)
	}

	// Check that the generated code contains important elements
	expectedParts := []string{
		"package main",
		"import",
		"func main",
		"TestFunc",
	}

	for _, part := range expectedParts {
		if !strings.Contains(code, part) {
			t.Errorf("Generated code missing expected part '%s'", part)
		}
	}
}

// TestValidateArguments tests argument validation for functions
func TestValidateArguments(t *testing.T) {
	module := &typesys.Module{
		Path: "example.com/test",
	}

	generator := NewTypeAwareCodeGenerator(module)

	// Test with nil type object
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
	}

	err := generator.ValidateArguments(funcSymbol, "arg1", "arg2")
	if err == nil {
		t.Error("Expected error for nil type object, got nil")
	}

	// Test with mismatched argument count (too few)
	mockFuncSymbol := mockFunction(t, "TestFunc", 2, 1)

	err = generator.ValidateArguments(mockFuncSymbol, "arg1") // Only 1 arg, needs 2
	if err == nil {
		t.Error("Expected error for too few arguments, got nil")
	}

	// Test with mismatched argument count (too many, non-variadic)
	err = generator.ValidateArguments(mockFuncSymbol, "arg1", "arg2", "arg3") // 3 args, needs 2
	if err == nil {
		t.Error("Expected error for too many arguments, got nil")
	}

	// Test with correct argument count
	err = generator.ValidateArguments(mockFuncSymbol, "arg1", "arg2")
	if err != nil {
		t.Errorf("ValidateArguments returned error for correct arguments: %v", err)
	}
}

// TestGenerateArgumentConversions tests generation of argument conversion code
func TestGenerateArgumentConversions(t *testing.T) {
	module := &typesys.Module{
		Path: "example.com/test",
	}

	generator := NewTypeAwareCodeGenerator(module)

	// Test with nil type object
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
	}

	_, err := generator.GenerateArgumentConversions(funcSymbol, "arg1")
	if err == nil {
		t.Error("Expected error for nil type object, got nil")
	}

	// Test with valid function symbol
	mockFuncSymbol := mockFunction(t, "TestFunc", 2, 1)

	conversions, err := generator.GenerateArgumentConversions(mockFuncSymbol, "arg1", "arg2")
	if err != nil {
		t.Errorf("GenerateArgumentConversions returned error: %v", err)
	}

	// Check that the conversions code contains references to arguments
	expectedParts := []string{
		"arg0", "arg1", "args",
	}

	// Depending on the implementation, not all parts might be present
	// but we should see at least one argument reference
	foundArgReference := false
	for _, part := range expectedParts {
		if strings.Contains(conversions, part) {
			foundArgReference = true
			break
		}
	}

	if !foundArgReference {
		t.Errorf("Generated conversions code doesn't contain any argument references:\n%s", conversions)
	}
}

// TestExecWrapperTemplate tests the template used for generating wrapper code
func TestExecWrapperTemplate(t *testing.T) {
	// Just verify that the template exists and has the expected structure
	if !strings.Contains(execWrapperTemplate, "package main") {
		t.Error("Template should contain 'package main'")
	}

	if !strings.Contains(execWrapperTemplate, "import") {
		t.Error("Template should contain import statements")
	}

	if !strings.Contains(execWrapperTemplate, "func main") {
		t.Error("Template should contain a main function")
	}
}
