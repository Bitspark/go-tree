// Package errors provides functions that return errors for testing
package errors

import (
	"errors"
	"fmt"
)

// DivideWithError returns the quotient of two integers
// Returns an error if b is 0
func DivideWithError(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// NotFoundError returns an error with a not found message
func NotFoundError(id string) error {
	return fmt.Errorf("resource with ID %s not found", id)
}

// FetchData returns data or an error
func FetchData(shouldFail bool) (string, error) {
	if shouldFail {
		return "", errors.New("failed to fetch data")
	}
	return "data", nil
}

// Global counter to track attempts across function calls
var temporaryFailureCounter int

// TemporaryFailure simulates a function that fails temporarily
// It will fail the given number of times, then succeed
func TemporaryFailure(failCount int) (string, error) {
	temporaryFailureCounter++

	if temporaryFailureCounter <= failCount {
		return "", fmt.Errorf("temporary failure (attempt %d of %d)", temporaryFailureCounter, failCount)
	}

	// Reset counter for next test run
	defer func() {
		temporaryFailureCounter = 0
	}()

	return "success after retries", nil
}

// PermanentFailure always fails with a permanent error
func PermanentFailure(unused int) (string, error) {
	return "", errors.New("permanent failure that should not be retried")
}
