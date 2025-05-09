package execute

import (
	"fmt"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// RetryPolicy defines how retries should be performed
type RetryPolicy struct {
	MaxRetries      int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Exponential backoff factor (delay increases by this factor after each attempt)
	JitterFactor    float64       // Random jitter factor (0-1) to add to delay to prevent thundering herd
	RetryableErrors []string      // Substring patterns of error messages that are retryable
}

// DefaultRetryPolicy returns a reasonable default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.2,
		RetryableErrors: []string{
			"connection reset",
			"timeout",
			"temporary",
			"deadline exceeded",
		},
	}
}

// RetryingFunctionRunner executes functions with automatic retries on failure
type RetryingFunctionRunner struct {
	*FunctionRunner // Embed the base FunctionRunner
	Policy          *RetryPolicy
	lastAttempts    int
	lastError       error
}

// NewRetryingFunctionRunner creates a new retrying function runner
func NewRetryingFunctionRunner(base *FunctionRunner) *RetryingFunctionRunner {
	return &RetryingFunctionRunner{
		FunctionRunner: base,
		Policy:         DefaultRetryPolicy(),
	}
}

// WithPolicy sets the retry policy
func (r *RetryingFunctionRunner) WithPolicy(policy *RetryPolicy) *RetryingFunctionRunner {
	r.Policy = policy
	return r
}

// WithMaxRetries sets the maximum number of retry attempts
func (r *RetryingFunctionRunner) WithMaxRetries(maxRetries int) *RetryingFunctionRunner {
	r.Policy.MaxRetries = maxRetries
	return r
}

// ExecuteFunc executes a function with retries according to the retry policy
func (r *RetryingFunctionRunner) ExecuteFunc(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	args ...interface{}) (interface{}, error) {

	var result interface{}
	var err error
	r.lastAttempts = 0
	r.lastError = nil

	delay := r.Policy.InitialDelay

	for attempt := 0; attempt <= r.Policy.MaxRetries; attempt++ {
		r.lastAttempts = attempt + 1

		// Execute the function
		result, err = r.FunctionRunner.ExecuteFunc(module, funcSymbol, args...)
		if err == nil {
			// Success!
			return result, nil
		}

		r.lastError = err

		// Check if we've exhausted retries
		if attempt >= r.Policy.MaxRetries {
			return nil, fmt.Errorf("function execution failed after %d attempts: %w", r.lastAttempts, err)
		}

		// Check if error is retryable
		if !r.isRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Add jitter to delay
		jitter := time.Duration(float64(delay) * r.Policy.JitterFactor * (2*r.randFloat() - 1))
		sleepTime := delay + jitter
		time.Sleep(sleepTime)

		// Exponential backoff for next attempt
		delay = time.Duration(float64(delay) * r.Policy.BackoffFactor)
		if delay > r.Policy.MaxDelay {
			delay = r.Policy.MaxDelay
		}
	}

	// We should never reach here, but just in case
	return nil, fmt.Errorf("unexpected error after %d attempts: %w", r.lastAttempts, err)
}

// isRetryableError checks if an error is retryable based on the policy
func (r *RetryingFunctionRunner) isRetryableError(err error) bool {
	// If no specific error patterns are defined, all errors are retryable
	if len(r.Policy.RetryableErrors) == 0 {
		return true
	}

	errMsg := err.Error()
	for _, pattern := range r.Policy.RetryableErrors {
		if pattern != "" && containsSubstring(errMsg, pattern) {
			return true
		}
	}

	return false
}

// LastAttempts returns the number of attempts made in the last execution
func (r *RetryingFunctionRunner) LastAttempts() int {
	return r.lastAttempts
}

// LastError returns the last error encountered in the last execution
func (r *RetryingFunctionRunner) LastError() error {
	return r.lastError
}

// Helper functions

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s != substr && contains(s, substr)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// randFloat returns a random float64 between 0 and 1
func (r *RetryingFunctionRunner) randFloat() float64 {
	// Simple implementation that doesn't require importing math/rand
	// In a real implementation, you'd use a proper random source
	return float64(time.Now().UnixNano()%1000) / 1000.0
}
