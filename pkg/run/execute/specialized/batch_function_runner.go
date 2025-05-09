package specialized

import (
	"fmt"
	"sync"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// FunctionExecution represents a single function to be executed
type FunctionExecution struct {
	Module      *typesys.Module
	FuncSymbol  *typesys.Symbol
	Args        []interface{}
	Result      interface{}
	Error       error
	Description string // Optional description for the function execution
}

// BatchFunctionRunner executes multiple functions in sequence or parallel
type BatchFunctionRunner struct {
	*execute.FunctionRunner      // Embed the base FunctionRunner
	Parallel                bool // Whether to execute functions in parallel
	MaxConcurrent           int  // Maximum number of concurrent executions (0 = unlimited)
	Results                 []*FunctionExecution
}

// NewBatchFunctionRunner creates a new batch function runner
func NewBatchFunctionRunner(base *execute.FunctionRunner) *BatchFunctionRunner {
	return &BatchFunctionRunner{
		FunctionRunner: base,
		Parallel:       false,
		MaxConcurrent:  0,
		Results:        make([]*FunctionExecution, 0),
	}
}

// WithParallel sets whether functions should be executed in parallel
func (r *BatchFunctionRunner) WithParallel(parallel bool) *BatchFunctionRunner {
	r.Parallel = parallel
	return r
}

// WithMaxConcurrent sets the maximum number of concurrent executions
func (r *BatchFunctionRunner) WithMaxConcurrent(max int) *BatchFunctionRunner {
	r.MaxConcurrent = max
	return r
}

// Add adds a function execution to the batch
func (r *BatchFunctionRunner) Add(module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) *FunctionExecution {
	execution := &FunctionExecution{
		Module:     module,
		FuncSymbol: funcSymbol,
		Args:       args,
	}
	r.Results = append(r.Results, execution)
	return execution
}

// AddWithDescription adds a function execution with a description
func (r *BatchFunctionRunner) AddWithDescription(description string, module *typesys.Module, funcSymbol *typesys.Symbol, args ...interface{}) *FunctionExecution {
	execution := &FunctionExecution{
		Module:      module,
		FuncSymbol:  funcSymbol,
		Args:        args,
		Description: description,
	}
	r.Results = append(r.Results, execution)
	return execution
}

// Execute executes all functions in the batch
func (r *BatchFunctionRunner) Execute() error {
	if r.Parallel {
		return r.executeParallel()
	}
	return r.executeSequential()
}

// executeSequential executes functions one after another
func (r *BatchFunctionRunner) executeSequential() error {
	var lastError error
	for _, execution := range r.Results {
		result, err := r.ExecuteFunc(execution.Module, execution.FuncSymbol, execution.Args...)
		execution.Result = result
		execution.Error = err
		if err != nil {
			lastError = err
		}
	}
	return lastError
}

// executeParallel executes functions in parallel with an optional concurrency limit
func (r *BatchFunctionRunner) executeParallel() error {
	var wg sync.WaitGroup
	var errMutex sync.Mutex
	var lastError error

	// Create a semaphore if we have a concurrency limit
	var sem chan struct{}
	if r.MaxConcurrent > 0 {
		sem = make(chan struct{}, r.MaxConcurrent)
	}

	for _, execution := range r.Results {
		wg.Add(1)
		go func(e *FunctionExecution) {
			defer wg.Done()

			// Acquire semaphore if using concurrency limiting
			if sem != nil {
				sem <- struct{}{}
				defer func() { <-sem }()
			}

			result, err := r.ExecuteFunc(e.Module, e.FuncSymbol, e.Args...)
			e.Result = result
			e.Error = err

			if err != nil {
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
			}
		}(execution)
	}

	wg.Wait()
	return lastError
}

// GetResults returns all function execution results
func (r *BatchFunctionRunner) GetResults() []*FunctionExecution {
	return r.Results
}

// Successful returns true if all executions were successful
func (r *BatchFunctionRunner) Successful() bool {
	for _, execution := range r.Results {
		if execution.Error != nil {
			return false
		}
	}
	return true
}

// FirstError returns the first error encountered, or nil if all executions were successful
func (r *BatchFunctionRunner) FirstError() error {
	for _, execution := range r.Results {
		if execution.Error != nil {
			return execution.Error
		}
	}
	return nil
}

// ResultsWithErrors returns all executions that encountered errors
func (r *BatchFunctionRunner) ResultsWithErrors() []*FunctionExecution {
	var results []*FunctionExecution
	for _, execution := range r.Results {
		if execution.Error != nil {
			results = append(results, execution)
		}
	}
	return results
}

// ResultsWithoutErrors returns all executions that completed successfully
func (r *BatchFunctionRunner) ResultsWithoutErrors() []*FunctionExecution {
	var results []*FunctionExecution
	for _, execution := range r.Results {
		if execution.Error == nil {
			results = append(results, execution)
		}
	}
	return results
}

// Summary returns a summary of the batch execution
func (r *BatchFunctionRunner) Summary() string {
	total := len(r.Results)
	successful := len(r.ResultsWithoutErrors())
	failed := len(r.ResultsWithErrors())

	return fmt.Sprintf("Batch execution summary: %d total, %d successful, %d failed",
		total, successful, failed)
}
