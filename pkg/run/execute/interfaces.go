// Package execute2 provides a redesigned approach to executing Go code with type awareness.
// It integrates with the resolve and materialize packages for improved functionality.
package execute

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/io/materialize"
)

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

	// Symbols that were tested
	TestedSymbols []*typesys.Symbol

	// Test coverage information
	Coverage float64
}

// Executor defines the core execution capabilities
type Executor interface {
	// Execute a command in a materialized environment
	Execute(env *materialize.Environment, command []string) (*ExecutionResult, error)

	// Execute a function in a materialized environment
	ExecuteFunc(env *materialize.Environment, module *typesys.Module,
		funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error)
}

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
}

// CodeGenerator generates executable code
type CodeGenerator interface {
	// Generate a complete executable program for a function
	GenerateFunctionWrapper(module *typesys.Module, funcSymbol *typesys.Symbol,
		args ...interface{}) (string, error)

	// Generate a test driver for a specific test function
	GenerateTestWrapper(module *typesys.Module, testSymbol *typesys.Symbol) (string, error)
}

// ResultProcessor handles processing execution output
type ResultProcessor interface {
	// Process raw execution result into a typed value
	ProcessFunctionResult(result *ExecutionResult, funcSymbol *typesys.Symbol) (interface{}, error)
}

// SecurityPolicy defines constraints for code execution
type SecurityPolicy interface {
	// Apply security constraints to an environment
	ApplyToEnvironment(env *materialize.Environment) error

	// Apply security constraints to command execution
	ApplyToExecution(command []string) []string

	// Get environment variables for execution
	GetEnvironmentVariables() map[string]string
}
