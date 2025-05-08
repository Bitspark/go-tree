// Package execute defines interfaces and implementations for executing code in Go modules.
package execute

import (
	"bitspark.dev/go-tree/pkg/core/module"
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
}

// ModuleExecutor runs code from a module
type ModuleExecutor interface {
	// Execute runs a command on a module
	Execute(module *module.Module, args ...string) (ExecutionResult, error)

	// ExecuteTest runs tests in a module
	ExecuteTest(module *module.Module, pkgPath string, testFlags ...string) (TestResult, error)

	// ExecuteFunc calls a specific function in the module
	ExecuteFunc(module *module.Module, funcPath string, args ...interface{}) (interface{}, error)
}
