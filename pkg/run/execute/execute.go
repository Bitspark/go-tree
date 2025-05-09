// Package execute defines interfaces and implementations for executing code in Go modules
// with full type awareness.
package execute

import (
	"io"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// ExecutionResult contains the result of executing a command
type ExecutionResult struct {
	// Command that was executed
	Command string

	// StdOut from the command
	StdOut string

	// StdErr from the command
	StdErr string

	// Exit code
	ExitCode int

	// Error if any occurred during execution
	Error error

	// Type information about the result (new in type-aware system)
	TypeInfo map[string]typesys.Symbol
}

// TestResult contains the result of running tests
type TestResult struct {
	// Package that was tested
	Package string

	// Tests that were run
	Tests []string

	// Tests that passed
	Passed int

	// Tests that failed
	Failed int

	// Test output
	Output string

	// Error if any occurred during execution
	Error error

	// Symbols that were tested (new in type-aware system)
	TestedSymbols []*typesys.Symbol

	// Test coverage information (new in type-aware system)
	Coverage float64
}

// ModuleExecutor runs code from a module
type ModuleExecutor interface {
	// Execute runs a command on a module
	Execute(module *typesys.Module, args ...string) (ExecutionResult, error)

	// ExecuteTest runs tests in a module
	ExecuteTest(module *typesys.Module, pkgPath string, testFlags ...string) (TestResult, error)

	// ExecuteFunc calls a specific function in the module with type checking
	// This is enhanced in the new system to leverage type information
	ExecuteFunc(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error)
}

// ExecutionContext manages code execution with type awareness
type ExecutionContext struct {
	// Module being executed
	Module *typesys.Module

	// Execution state
	TempDir string
	Files   map[string]*typesys.File

	// Output capture
	Stdout io.Writer
	Stderr io.Writer
}

// NewExecutionContext creates a new execution context for the given module
func NewExecutionContext(module *typesys.Module) *ExecutionContext {
	return &ExecutionContext{
		Module: module,
		Files:  make(map[string]*typesys.File),
	}
}

// Execute compiles and runs a piece of code with type checking
func (ctx *ExecutionContext) Execute(code string, args ...interface{}) (*ExecutionResult, error) {
	// Will be implemented in typeaware.go
	return nil, nil
}

// ExecuteInline executes code inline with the current context
func (ctx *ExecutionContext) ExecuteInline(code string) (*ExecutionResult, error) {
	// Will be implemented in typeaware.go
	return nil, nil
}
