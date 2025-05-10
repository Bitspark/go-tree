package toolkit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestDepthLimitingMiddleware tests the depth limiting middleware
func TestDepthLimitingMiddleware(t *testing.T) {
	// Create middleware with max depth of 2
	middleware := NewDepthLimitingMiddleware(2)

	// Create test module
	testModule := &typesys.Module{Path: "test/module"}

	// Create counters for testing
	callCount := 0

	// Create a next function that counts calls
	nextFunc := func() (*typesys.Module, error) {
		callCount++
		return testModule, nil
	}

	// Test successful execution (within depth limit)
	ctx := context.Background()
	importPath := "test/module"
	version := "v1.0.0"

	// First call - depth 0
	var module *typesys.Module
	var err error
	ctx, module, err = middleware(ctx, importPath, version, nextFunc)
	if err != nil {
		t.Errorf("First call: Expected no error, got: %v", err)
	}
	if module != testModule {
		t.Errorf("First call: Expected test module, got: %v", module)
	}

	// Verify depth has been incremented in the context
	depth := 0
	if depthVal := ctx.Value(contextKeyResolutionDepth); depthVal != nil {
		if depthFromCtx, ok := depthVal.(int); ok {
			depth = depthFromCtx
		}
	}
	if depth != 1 {
		t.Errorf("Expected depth of 1 after first call, got: %d", depth)
	}

	// Second call - depth 1
	ctx, module, err = middleware(ctx, importPath, version, nextFunc)
	if err != nil {
		t.Errorf("Second call: Expected no error, got: %v", err)
	}
	if module != testModule {
		t.Errorf("Second call: Expected test module, got: %v", module)
	}

	// Verify depth has been incremented again
	depth = 0
	if depthVal := ctx.Value(contextKeyResolutionDepth); depthVal != nil {
		if depthFromCtx, ok := depthVal.(int); ok {
			depth = depthFromCtx
		}
	}
	if depth != 2 {
		t.Errorf("Expected depth of 2 after second call, got: %d", depth)
	}

	// Verify call count
	if callCount != 2 {
		t.Errorf("Expected 2 calls to next function, got: %d", callCount)
	}

	// Third call (should hit depth limit - depth 2)
	_, _, err = middleware(ctx, importPath, version, nextFunc)
	if err == nil {
		t.Errorf("Third call: Expected depth limit error, got nil")
	}

	// Check if it's the right error type
	depthErr, ok := err.(*DepthLimitError)
	if !ok {
		t.Errorf("Third call: Expected *DepthLimitError, got: %T", err)
	} else {
		// Verify error fields
		if depthErr.MaxDepth != 2 {
			t.Errorf("Expected MaxDepth=2, got: %d", depthErr.MaxDepth)
		}
		if depthErr.ImportPath != importPath {
			t.Errorf("Expected ImportPath='%s', got: '%s'", importPath, depthErr.ImportPath)
		}
		if depthErr.Version != version {
			t.Errorf("Expected Version='%s', got: '%s'", version, depthErr.Version)
		}
	}

	// Verify call count didn't increase (no next() on error)
	if callCount != 2 {
		t.Errorf("Expected still 2 calls to next function, got: %d", callCount)
	}

	// Now, let's create a new context to test that depth is not carried over
	freshCtx := context.Background()

	// First call with fresh context should succeed
	// We don't use the returned context in this test
	_, _, err = middleware(freshCtx, importPath, version, nextFunc)
	if err != nil {
		t.Errorf("Fresh context call: Expected no error, got: %v", err)
	}
}

// TestDepthLimitingMiddlewareThreadSafety tests thread safety of depth limiting middleware
func TestDepthLimitingMiddlewareThreadSafety(t *testing.T) {
	// Create middleware with max depth of 3
	middleware := NewDepthLimitingMiddleware(3)

	// Create test module
	testModule := &typesys.Module{Path: "test/module"}

	// Next function that just returns the test module
	nextFunc := func() (*typesys.Module, error) {
		return testModule, nil
	}

	// Run multiple goroutines concurrently to test thread safety
	wg := sync.WaitGroup{}
	errChan := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Each goroutine starts with a fresh context
			ctx := context.Background()
			importPath := fmt.Sprintf("test/module/%d", goroutineID)
			version := "v1.0.0"

			var module *typesys.Module
			var err error

			// First 3 calls should succeed (depth 0, 1, 2)
			for call := 0; call < 3; call++ {
				ctx, module, err = middleware(ctx, importPath, version, nextFunc)
				if err != nil {
					errChan <- fmt.Errorf("goroutine %d: unexpected error on call %d: %v",
						goroutineID, call+1, err)
					return
				}

				// Verify the module was returned correctly
				if module != testModule {
					errChan <- fmt.Errorf("goroutine %d: expected testModule on call %d",
						goroutineID, call+1)
					return
				}

				// Verify depth in context
				depth := 0
				if depthVal := ctx.Value(contextKeyResolutionDepth); depthVal != nil {
					if d, ok := depthVal.(int); ok {
						depth = d
					}
				}
				if depth != call+1 {
					errChan <- fmt.Errorf("goroutine %d: expected depth %d after call %d, got %d",
						goroutineID, call+1, call+1, depth)
					return
				}
			}

			// 4th call should hit depth limit (depth 3)
			_, _, err = middleware(ctx, importPath, version, nextFunc)
			if err == nil {
				errChan <- fmt.Errorf("goroutine %d: expected depth limit error on 4th call, got nil",
					goroutineID)
				return
			}

			// Verify it's the right error type
			depthErr, ok := err.(*DepthLimitError)
			if !ok {
				errChan <- fmt.Errorf("goroutine %d: expected DepthLimitError, got %T",
					goroutineID, err)
				return
			}

			// Verify error fields
			if depthErr.MaxDepth != 3 {
				errChan <- fmt.Errorf("goroutine %d: expected MaxDepth=3, got %d",
					goroutineID, depthErr.MaxDepth)
			}
			if depthErr.ImportPath != importPath {
				errChan <- fmt.Errorf("goroutine %d: expected ImportPath=%s, got %s",
					goroutineID, importPath, depthErr.ImportPath)
			}

		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent test error: %v", err)
	}
}

// TestCachingMiddleware tests the caching middleware
func TestCachingMiddleware(t *testing.T) {
	// Create caching middleware
	middleware := NewCachingMiddleware()

	// Create a unique test module for each call
	moduleCounter := 0
	nextFunc := func() (*typesys.Module, error) {
		moduleCounter++
		return &typesys.Module{Path: "test/module", Dir: string(rune('a' + moduleCounter - 1))}, nil
	}

	ctx := context.Background()

	// First call should use the next function
	ctx, module1, err := middleware(ctx, "test/module", "v1.0.0", nextFunc)
	if err != nil {
		t.Errorf("First call: Expected no error, got: %v", err)
	}
	if module1.Dir != "a" {
		t.Errorf("First call: Expected module.Dir='a', got: '%s'", module1.Dir)
	}

	// Second call with same path and version should use cache
	ctx, module2, err := middleware(ctx, "test/module", "v1.0.0", nextFunc)
	if err != nil {
		t.Errorf("Second call: Expected no error, got: %v", err)
	}
	if module2.Dir != "a" {
		t.Errorf("Second call: Expected cached module.Dir='a', got: '%s'", module2.Dir)
	}

	// Different path should call next function
	ctx, module3, err := middleware(ctx, "other/module", "v1.0.0", nextFunc)
	if err != nil {
		t.Errorf("Third call: Expected no error, got: %v", err)
	}
	if module3.Dir != "b" {
		t.Errorf("Third call: Expected module.Dir='b', got: '%s'", module3.Dir)
	}

	// Different version should call next function
	_, module4, err := middleware(ctx, "test/module", "v2.0.0", nextFunc)
	if err != nil {
		t.Errorf("Fourth call: Expected no error, got: %v", err)
	}
	if module4.Dir != "c" {
		t.Errorf("Fourth call: Expected module.Dir='c', got: '%s'", module4.Dir)
	}

	// Verify the moduleCounter
	if moduleCounter != 3 {
		t.Errorf("Expected 3 unique modules created, got: %d", moduleCounter)
	}
}

// TestCachingMiddlewareWithErrors tests caching middleware with errors
func TestCachingMiddlewareWithErrors(t *testing.T) {
	// Create caching middleware
	middleware := NewCachingMiddleware()

	// Create a function that returns errors for certain modules
	callCount := 0
	nextFunc := func() (*typesys.Module, error) {
		callCount++
		return nil, errors.New("test error")
	}

	ctx := context.Background()

	// First call should return error
	ctx, _, err := middleware(ctx, "error/module", "v1.0.0", nextFunc)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Second call should still call next function since errors aren't cached
	_, _, err = middleware(ctx, "error/module", "v1.0.0", nextFunc)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Verify call count
	if callCount != 2 {
		t.Errorf("Expected 2 calls to next function, got: %d", callCount)
	}
}

// TestCachingMiddlewareThreadSafety tests thread safety of caching middleware
func TestCachingMiddlewareThreadSafety(t *testing.T) {
	// Create caching middleware
	middleware := NewCachingMiddleware()

	// Create a function that returns a unique module each time
	var mu sync.Mutex
	moduleCounter := 0
	nextFunc := func() (*typesys.Module, error) {
		mu.Lock()
		moduleCounter++
		count := moduleCounter
		mu.Unlock()

		// Simulate some work
		time.Sleep(time.Millisecond)

		return &typesys.Module{Path: "test/module", Dir: string(rune('a' + count - 1))}, nil
	}

	// Run multiple goroutines concurrently accessing the same key
	wg := sync.WaitGroup{}
	resultChan := make(chan string, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := context.Background()
			_, module, err := middleware(ctx, "test/module", "v1.0.0", nextFunc)
			if err != nil {
				resultChan <- "error:" + err.Error()
				return
			}

			resultChan <- module.Dir
		}()
	}

	wg.Wait()
	close(resultChan)

	// We should get the same module.Dir for all goroutines
	expectedDir := ""
	for dir := range resultChan {
		if expectedDir == "" {
			expectedDir = dir
		} else if dir != expectedDir {
			t.Errorf("Cache inconsistency: got both '%s' and '%s'", expectedDir, dir)
		}
	}

	// Verify moduleCounter is 1 (only one call to next function)
	if moduleCounter != 1 {
		t.Errorf("Expected 1 call to next function, got: %d", moduleCounter)
	}
}

// TestErrorEnhancerMiddleware tests the error enhancer middleware
func TestErrorEnhancerMiddleware(t *testing.T) {
	// Create error enhancer middleware
	middleware := NewErrorEnhancerMiddleware()

	// Test with a function that returns error
	errNextFunc := func() (*typesys.Module, error) {
		return nil, errors.New("original error")
	}

	ctx := context.Background()

	// Call with error
	ctx, _, err := middleware(ctx, "test/module", "v1.0.0", errNextFunc)
	if err == nil {
		t.Errorf("Expected enhanced error, got nil")
	}

	// Error should contain the module info
	errStr := err.Error()
	if !strings.Contains(errStr, "test/module") {
		t.Errorf("Expected error to contain module path, got: %s", errStr)
	}
	if !strings.Contains(errStr, "v1.0.0") {
		t.Errorf("Expected error to contain version, got: %s", errStr)
	}
	if !strings.Contains(errStr, "original error") {
		t.Errorf("Expected error to contain original error message, got: %s", errStr)
	}

	// Test with a function that returns a typed error
	depthErrNextFunc := func() (*typesys.Module, error) {
		return nil, &DepthLimitError{
			ImportPath: "test/module",
			Version:    "v1.0.0",
			MaxDepth:   3,
			Path:       []string{"a", "b", "c"},
		}
	}

	// Call with depth error - should not be wrapped
	ctx, _, err = middleware(ctx, "test/module", "v1.0.0", depthErrNextFunc)
	if err == nil {
		t.Errorf("Expected depth error, got nil")
	}

	// Verify it's the right error type (not wrapped)
	_, ok := err.(*DepthLimitError)
	if !ok {
		t.Errorf("Expected unwrapped *DepthLimitError, got: %T", err)
	}

	// Test with a function that returns success
	successNextFunc := func() (*typesys.Module, error) {
		return &typesys.Module{Path: "test/module"}, nil
	}

	// Call with success
	_, module, err := middleware(ctx, "test/module", "v1.0.0", successNextFunc)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if module == nil {
		t.Errorf("Expected module, got nil")
	}
}

// TestMiddlewareChainComplex tests a complex middleware chain
func TestMiddlewareChainComplex(t *testing.T) {
	chain := NewMiddlewareChain()

	// Add multiple middleware types
	chain.Add(
		// Depth limiting middleware
		NewDepthLimitingMiddleware(2),
		// Caching middleware
		NewCachingMiddleware(),
		// Error enhancer middleware
		NewErrorEnhancerMiddleware(),
	)

	// Create counters
	callCount := 0

	// Create final function
	finalFunc := func() (*typesys.Module, error) {
		callCount++
		return &typesys.Module{Path: "test/module", Dir: string(rune('a' + callCount - 1))}, nil
	}

	ctx := context.Background()

	// First call
	module1, err := chain.Execute(ctx, "test/module", "v1.0.0", finalFunc)
	if err != nil {
		t.Errorf("First call: Expected no error, got: %v", err)
	}
	if module1.Dir != "a" {
		t.Errorf("First call: Expected module.Dir='a', got: '%s'", module1.Dir)
	}

	// Second call with same args should use cache
	module2, err := chain.Execute(ctx, "test/module", "v1.0.0", finalFunc)
	if err != nil {
		t.Errorf("Second call: Expected no error, got: %v", err)
	}
	if module2.Dir != "a" {
		t.Errorf("Second call: Expected cached module.Dir='a', got: '%s'", module2.Dir)
	}

	// Third call with different path should not hit cache but still be within depth limit
	module3, err := chain.Execute(ctx, "other/module", "v1.0.0", finalFunc)
	if err != nil {
		t.Errorf("Third call: Expected no error, got: %v", err)
	}
	if module3.Dir != "b" {
		t.Errorf("Third call: Expected module.Dir='b', got: '%s'", module3.Dir)
	}

	// Count should be 2 due to caching
	if callCount != 2 {
		t.Errorf("Expected 2 calls to final function, got: %d", callCount)
	}
}
