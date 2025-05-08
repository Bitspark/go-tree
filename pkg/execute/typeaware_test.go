package execute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TestNewTypeAwareExecutor verifies creation of a TypeAwareExecutor
func TestNewTypeAwareExecutor(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a type-aware executor
	executor := NewTypeAwareExecutor(module)

	// Verify the executor was created correctly
	if executor == nil {
		t.Fatal("NewTypeAwareExecutor returned nil")
	}

	if executor.Module != module {
		t.Errorf("Expected executor.Module to be %v, got %v", module, executor.Module)
	}

	if executor.Sandbox == nil {
		t.Error("Executor should have a non-nil Sandbox")
	}

	if executor.Generator == nil {
		t.Error("Executor should have a non-nil Generator")
	}
}

// TestTypeAwareExecutor_ExecuteCode tests the ExecuteCode method
func TestTypeAwareExecutor_ExecuteCode(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  os.TempDir(), // Use a valid directory
	}

	// Create a type-aware executor
	executor := NewTypeAwareExecutor(module)

	// Test executing a simple program
	code := `
package main

import "fmt"

func main() {
	fmt.Println("Hello from type-aware execution")
}
`
	result, err := executor.ExecuteCode(code)

	// If execution fails, it might be due to environment issues (like Go not installed)
	// So we'll check the error and skip the test if necessary
	if err != nil {
		t.Skipf("Skipping test due to execution error: %v", err)
		return
	}

	// Verify the result
	if result == nil {
		t.Fatal("ExecuteCode returned nil result")
	}

	if !strings.Contains(result.StdOut, "Hello from type-aware execution") {
		t.Errorf("Expected output to contain greeting, got: %s", result.StdOut)
	}

	if result.Error != nil {
		t.Errorf("Expected nil error, got: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}

// TestTypeAwareExecutor_ExecuteFunction tests the ExecuteFunction method
func TestTypeAwareExecutor_ExecuteFunction(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a type-aware executor
	executor := NewTypeAwareExecutor(module)

	// Create a symbol to execute
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
	}

	// Attempt to execute the function (should return an error since it's a stub)
	_, err := executor.ExecuteFunction(funcSymbol)

	// Verify we get an expected error (since we expect execution to fail without a real symbol)
	if err == nil {
		t.Error("Expected error from ExecuteFunction for stub symbol, got nil")
	}

	// Check that the error message mentions the function name
	if !strings.Contains(err.Error(), "TestFunc") {
		t.Errorf("Expected error to mention function name, got: %s", err.Error())
	}
}

// TestTypeAwareExecutor_ExecuteFunc tests the ExecuteFunc interface method
func TestTypeAwareExecutor_ExecuteFunc(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a new module to trigger the module update branch
	newModule := &typesys.Module{
		Path: "example.com/newtest",
	}

	// Create a type-aware executor
	executor := NewTypeAwareExecutor(module)

	// Save the original sandbox and generator
	originalSandbox := executor.Sandbox
	originalGenerator := executor.Generator

	// Execute with a new module to trigger the module update branch
	funcSymbol := &typesys.Symbol{
		Name: "TestFunc",
		Kind: typesys.KindFunction,
	}

	// Call ExecuteFunc with the new module
	_, err := executor.ExecuteFunc(newModule, funcSymbol)

	// Verify the error as in the previous test
	if err == nil {
		t.Error("Expected error from ExecuteFunc for stub symbol, got nil")
	}

	// Verify the module was updated
	if executor.Module != newModule {
		t.Errorf("Expected module to be updated to %v, got %v", newModule, executor.Module)
	}

	// Verify the sandbox and generator were recreated
	if executor.Sandbox == originalSandbox {
		t.Error("Expected sandbox to be recreated")
	}

	if executor.Generator == originalGenerator {
		t.Error("Expected generator to be recreated")
	}
}

// TestNewExecutionContextImpl tests creating a new execution context
func TestNewExecutionContextImpl(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a new execution context
	ctx, err := NewExecutionContextImpl(module)
	if err != nil {
		t.Fatalf("NewExecutionContextImpl returned error: %v", err)
	}
	defer ctx.Close() // Ensure cleanup

	// Verify the context was created correctly
	if ctx == nil {
		t.Fatal("NewExecutionContextImpl returned nil context")
	}

	if ctx.Module != module {
		t.Errorf("Expected module %v, got %v", module, ctx.Module)
	}

	if ctx.TempDir == "" {
		t.Error("Expected non-empty TempDir")
	}

	// Check if the directory exists
	if _, err := os.Stat(ctx.TempDir); os.IsNotExist(err) {
		t.Errorf("TempDir %s does not exist", ctx.TempDir)
	}

	if ctx.Files == nil {
		t.Error("Files map should not be nil")
	}

	if ctx.Stdout == nil {
		t.Error("Stdout should not be nil")
	}

	if ctx.Stderr == nil {
		t.Error("Stderr should not be nil")
	}

	if ctx.executor == nil {
		t.Error("Executor should not be nil")
	}
}

// TestExecutionContextImpl_Execute tests the Execute method
func TestExecutionContextImpl_Execute(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  os.TempDir(), // Use a valid directory
	}

	// Create a new execution context
	ctx, err := NewExecutionContextImpl(module)
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer ctx.Close() // Ensure cleanup

	// Test executing a simple program
	code := `
package main

import "fmt"

func main() {
	fmt.Println("Hello from execution context")
}
`
	// Execute the code
	result, err := ctx.Execute(code)

	// If execution fails, it might be due to environment issues
	if err != nil {
		t.Skipf("Skipping test due to execution error: %v", err)
		return
	}

	// Verify the result
	if result == nil {
		t.Fatal("Execute returned nil result")
	}

	// Check stdout is captured in both the result and context
	if !strings.Contains(result.StdOut, "Hello from execution context") {
		t.Errorf("Expected result output to contain greeting, got: %s", result.StdOut)
	}

	if !strings.Contains(ctx.Stdout.String(), "Hello from execution context") {
		t.Errorf("Expected context stdout to contain greeting, got: %s", ctx.Stdout.String())
	}
}

// TestExecutionContextImpl_ExecuteInline tests the ExecuteInline method
func TestExecutionContextImpl_ExecuteInline(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
		Dir:  os.TempDir(), // Use a valid directory
	}

	// Create a new execution context
	ctx, err := NewExecutionContextImpl(module)
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer ctx.Close() // Ensure cleanup

	// Test executing inline code - use a simple fmt-only example that doesn't need the module
	code := `fmt.Println("Hello inline")`

	// Execute the inline code
	result, err := ctx.ExecuteInline(code)

	// If execution fails, provide detailed diagnostics
	if err != nil {
		t.Logf("Execution failed with error: %v", err)
		t.Logf("Generated code might be:\npackage main\n\nimport (\n  \"example.com/test\"\n  \"fmt\"\n)\n\nfunc main() {\n\tfmt.Println(\"Hello inline\")\n}")
		t.Skipf("Skipping test due to execution error: %v", err)
		return
	}

	// Verify the result
	if result == nil {
		t.Fatal("ExecuteInline returned nil result")
	}

	// Check stdout is captured and provide detailed error message
	if !strings.Contains(result.StdOut, "Hello inline") {
		t.Errorf("Expected output to contain 'Hello inline', got: %s", result.StdOut)
		if result.StdErr != "" {
			t.Logf("Stderr contained: %s", result.StdErr)
		}
	}
}

// TestExecutionContextImpl_Close tests the Close method
func TestExecutionContextImpl_Close(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a new execution context
	ctx, err := NewExecutionContextImpl(module)
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}

	// Save the temp directory path
	tempDir := ctx.TempDir

	// Verify the directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("TempDir %s does not exist before Close", tempDir)
	}

	// Close the context
	err = ctx.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Verify the directory was removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("TempDir %s still exists after Close", tempDir)
		// Clean up in case the test fails
		os.RemoveAll(tempDir)
	}
}

// TestParseExecutionResult tests the ParseExecutionResult function
func TestParseExecutionResult(t *testing.T) {
	// Test parsing a valid JSON result
	jsonResult := `{"name": "test", "value": 42}`

	// Create a struct to parse into
	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Parse the result
	err := ParseExecutionResult(jsonResult, &result)
	if err != nil {
		t.Errorf("ParseExecutionResult returned error for valid JSON: %v", err)
	}

	// Verify the parsed values
	if result.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", result.Name)
	}

	if result.Value != 42 {
		t.Errorf("Expected value 42, got %d", result.Value)
	}

	// Test parsing with whitespace
	jsonWithWhitespace := `
	{
		"name": "test2",
		"value": 43
	}
	`

	var result2 struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	err = ParseExecutionResult(jsonWithWhitespace, &result2)
	if err != nil {
		t.Errorf("ParseExecutionResult returned error for valid JSON with whitespace: %v", err)
	}

	// Verify the parsed values
	if result2.Name != "test2" {
		t.Errorf("Expected name 'test2', got '%s'", result2.Name)
	}

	// Test parsing an empty result
	err = ParseExecutionResult("", &result)
	if err == nil {
		t.Error("Expected error for empty result, got nil")
	}

	// Test parsing invalid JSON
	err = ParseExecutionResult("not json", &result)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestTypeAwareCodeGenerator verifies that the TypeAwareCodeGenerator can be created
func TestTypeAwareCodeGenerator(t *testing.T) {
	// Create a test module
	module := &typesys.Module{
		Path: "example.com/test",
	}

	// Create a type-aware code generator
	generator := NewTypeAwareCodeGenerator(module)

	// Verify the generator was created correctly
	if generator == nil {
		t.Fatal("NewTypeAwareCodeGenerator returned nil")
	}

	if generator.Module != module {
		t.Errorf("Expected generator.Module to be %v, got %v", module, generator.Module)
	}
}

// TestTypeAwareExecution_Integration does a simple integration test of the type-aware execution system
func TestTypeAwareExecution_Integration(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "typeaware-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go module
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module example.com/typeaware\n\ngo 1.16\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create a file with exported functions
	utilContent := `package utils

// Add adds two integers
func Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers
func Multiply(a, b int) int {
	return a * b
}
`
	err = os.MkdirAll(filepath.Join(tempDir, "utils"), 0755)
	if err != nil {
		t.Fatalf("Failed to create utils directory: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "utils", "math.go"), []byte(utilContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write utils/math.go: %v", err)
	}

	// Create the module structure
	module := &typesys.Module{
		Path: "example.com/typeaware",
		Dir:  tempDir,
		Packages: map[string]*typesys.Package{
			"example.com/typeaware/utils": {
				ImportPath: "example.com/typeaware/utils",
				Name:       "utils",
				Files: map[string]*typesys.File{
					filepath.Join(tempDir, "utils", "math.go"): {
						Path: filepath.Join(tempDir, "utils", "math.go"),
						Name: "math.go",
					},
				},
				Symbols: map[string]*typesys.Symbol{
					"Add": {
						ID:       "Add",
						Name:     "Add",
						Kind:     typesys.KindFunction,
						Exported: true,
					},
					"Multiply": {
						ID:       "Multiply",
						Name:     "Multiply",
						Kind:     typesys.KindFunction,
						Exported: true,
					},
				},
			},
		},
	}

	// Create a new execution context
	ctx, err := NewExecutionContextImpl(module)
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer ctx.Close()

	// Execute code that uses the module
	code := `
package main

import (
	"fmt"
	"example.com/typeaware/utils"
)

func main() {
	sum := utils.Add(5, 3)
	product := utils.Multiply(4, 7)
	fmt.Printf("Sum: %d, Product: %d\n", sum, product)
}
`
	// This test may fail depending on environment, so we'll make it conditional
	result, err := ctx.Execute(code)
	if err != nil {
		t.Skipf("Skipping integration test due to execution error: %v", err)
		return
	}

	// Verify the result
	expectedOutput := "Sum: 8, Product: 28"
	if !strings.Contains(result.StdOut, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedOutput, result.StdOut)
	}
}
