package materialize

import (
	"fmt"
)

// MaterializeError represents an error during materialization
type MaterializeError struct {
	Message string
	Err     error
}

// Error returns a string representation of the error
func (e *MaterializeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("materialization error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("materialization error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *MaterializeError) Unwrap() error {
	return e.Err
}

// MaterializationError represents an error during materialization of a specific module
type MaterializationError struct {
	ModulePath string
	Message    string
	Err        error
}

// Error returns a string representation of the error
func (e *MaterializationError) Error() string {
	if e.ModulePath != "" {
		if e.Err != nil {
			return fmt.Sprintf("materialization error for %s: %s: %v", e.ModulePath, e.Message, e.Err)
		}
		return fmt.Sprintf("materialization error for %s: %s", e.ModulePath, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("materialization error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("materialization error: %s", e.Message)
}

// Unwrap returns the underlying error
func (e *MaterializationError) Unwrap() error {
	return e.Err
}
