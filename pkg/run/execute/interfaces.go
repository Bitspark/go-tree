// Package execute provides a redesigned approach to executing Go code with type awareness.
// It integrates with the resolve and materialize packages for improved functionality.
package execute

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/env"
	"bitspark.dev/go-tree/pkg/io/materialize"
)

// Alias the interfaces from materialize for convenience
type ModuleMaterializer = *materialize.ModuleMaterializer
type Environment = *env.Environment

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

// ModuleResolver resolves modules by import path
type ModuleResolver interface {
	// ResolveModule resolves a module by import path and version
	ResolveModule(path, version string, opts interface{}) (*typesys.Module, error)

	// ResolveDependencies resolves dependencies for a module
	ResolveDependencies(module interface{}, depth int) error
}

// Executor executes commands in an environment
type Executor interface {
	// Execute executes a command in an environment
	Execute(env Environment, command []string) (*ExecutionResult, error)

	// ExecuteFunc executes a function in a materialized environment
	ExecuteFunc(env Environment, module *typesys.Module,
		funcSymbol *typesys.Symbol, args ...interface{}) (interface{}, error)
}

// ExecutionResult represents the result of executing a command
type ExecutionResult struct {
	// Exit code of the command
	ExitCode int

	// Standard output
	StdOut string

	// Standard error
	StdErr string

	// Command that was executed
	Command string

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

// SecurityPolicy defines a security policy for code execution
type SecurityPolicy interface {
	// ApplyToEnvironment applies the security policy to an environment
	ApplyToEnvironment(env *env.Environment) error

	// Apply security constraints to command execution
	ApplyToExecution(command []string) []string

	// Get environment variables for execution
	GetEnvironmentVariables() map[string]string
}
