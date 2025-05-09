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
