package execute

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TypeAwareExecutor provides type-aware execution of code
type TypeAwareExecutor struct {
	// Module being executed
	Module *typesys.Module

	// Sandbox for secure execution
	Sandbox *Sandbox

	// Code generator for creating wrapper code
	Generator *TypeAwareCodeGenerator
}

// NewTypeAwareExecutor creates a new type-aware executor
func NewTypeAwareExecutor(module *typesys.Module) *TypeAwareExecutor {
	return &TypeAwareExecutor{
		Module:    module,
		Sandbox:   NewSandbox(module),
		Generator: NewTypeAwareCodeGenerator(module),
	}
}

// ExecuteCode executes a piece of code with type awareness
func (e *TypeAwareExecutor) ExecuteCode(code string) (*ExecutionResult, error) {
	return e.Sandbox.Execute(code)
}

// ExecuteFunction executes a function with proper type checking
func (e *TypeAwareExecutor) ExecuteFunction(funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	return e.Sandbox.ExecuteFunction(funcSymbol, args...)
}

// Execute implements the ModuleExecutor.ExecuteFunc interface
func (e *TypeAwareExecutor) ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error) {
	// Update the module and sandbox if needed
	if module != e.Module {
		e.Module = module
		e.Sandbox = NewSandbox(module)
		e.Generator = NewTypeAwareCodeGenerator(module)
	}

	return e.ExecuteFunction(funcSymbol, args...)
}

// ExecutionContextImpl provides a concrete implementation of ExecutionContext
type ExecutionContextImpl struct {
	// Module being executed
	Module *typesys.Module

	// Execution state
	TempDir string
	Files   map[string]*typesys.File

	// Output capture
	Stdout *strings.Builder
	Stderr *strings.Builder

	// Executor for running code
	executor *TypeAwareExecutor
}

// NewExecutionContextImpl creates a new execution context
func NewExecutionContextImpl(module *typesys.Module) (*ExecutionContextImpl, error) {
	// Create a temporary directory for execution
	tempDir, err := ioutil.TempDir("", "goexec-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return &ExecutionContextImpl{
		Module:   module,
		TempDir:  tempDir,
		Files:    make(map[string]*typesys.File),
		Stdout:   &strings.Builder{},
		Stderr:   &strings.Builder{},
		executor: NewTypeAwareExecutor(module),
	}, nil
}

// Execute compiles and runs a piece of code with type checking
func (ctx *ExecutionContextImpl) Execute(code string, args ...interface{}) (*ExecutionResult, error) {
	// Save the code to a temporary file
	filename := "execute.go"
	filePath := filepath.Join(ctx.TempDir, filename)

	if err := ioutil.WriteFile(filePath, []byte(code), 0600); err != nil {
		return nil, fmt.Errorf("failed to write code to file: %w", err)
	}

	// Configure the sandbox to capture output
	ctx.executor.Sandbox.AllowFileIO = true // Allow file access within the temp directory

	// Execute the code
	result, err := ctx.executor.ExecuteCode(code)
	if err != nil {
		return nil, err
	}

	// Append output to context's stdout/stderr
	if result.StdOut != "" {
		ctx.Stdout.WriteString(result.StdOut)
	}
	if result.StdErr != "" {
		ctx.Stderr.WriteString(result.StdErr)
	}

	return result, nil
}

// ExecuteInline executes code inline with the current context
func (ctx *ExecutionContextImpl) ExecuteInline(code string) (*ExecutionResult, error) {
	// For inline execution, we'll enhance the code with imports for the current module
	// and wrap it in a function that can be executed

	packageImport := fmt.Sprintf("import \"%s\"\n", ctx.Module.Path)
	wrappedCode := fmt.Sprintf(`
package main

%s
import "fmt"

func main() {
%s
}
`, packageImport, code)

	return ctx.Execute(wrappedCode)
}

// Close cleans up the execution context
func (ctx *ExecutionContextImpl) Close() error {
	if ctx.TempDir != "" {
		if err := os.RemoveAll(ctx.TempDir); err != nil {
			return fmt.Errorf("failed to remove temporary directory: %w", err)
		}
	}
	return nil
}

// ParseExecutionResult attempts to parse the result of an execution into a typed value
func ParseExecutionResult(result string, target interface{}) error {
	result = strings.TrimSpace(result)
	if result == "" {
		return fmt.Errorf("empty execution result")
	}

	return json.Unmarshal([]byte(result), target)
}
