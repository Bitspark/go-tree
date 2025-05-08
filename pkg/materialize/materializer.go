// Package materialize provides functionality for materializing Go modules to disk.
// It serves as the inverse operation to the resolve package, enabling serialization
// of in-memory modules back to filesystem with proper dependency structure.
package materialize

import (
	"bitspark.dev/go-tree/pkg/typesys"
)

// Materializer defines the interface for module materialization
type Materializer interface {
	// Materialize writes a module to disk with dependencies
	Materialize(module *typesys.Module, opts MaterializeOptions) (*Environment, error)

	// MaterializeForExecution prepares a module for running
	MaterializeForExecution(module *typesys.Module, opts MaterializeOptions) (*Environment, error)

	// MaterializeMultipleModules materializes multiple modules together
	MaterializeMultipleModules(modules []*typesys.Module, opts MaterializeOptions) (*Environment, error)
}

// MaterializationError represents an error during materialization
type MaterializationError struct {
	// Module path where the error occurred
	ModulePath string

	// Error message
	Message string

	// Original error
	Err error
}

// Error returns a string representation of the error
func (e *MaterializationError) Error() string {
	msg := "materialization error"
	if e.ModulePath != "" {
		msg += " for module " + e.ModulePath
	}
	if e.Message != "" {
		msg += ": " + e.Message
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}
