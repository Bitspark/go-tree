package env

import (
	"context"
	"fmt"
	"sync"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Context keys for passing data through middleware chain
type contextKey string

const (
	// For tracking resolution path
	contextKeyResolutionPath contextKey = "resolutionPath"
	// For tracking resolution depth
	contextKeyResolutionDepth contextKey = "resolutionDepth"
	// For tracking call chains in middleware
	contextKeyChainID contextKey = "chainID"
)

// ResolutionFunc represents the next resolver in the chain
type ResolutionFunc func() (*typesys.Module, error)

// ResolutionMiddleware intercepts module resolution requests
// The returned context should be used for subsequent calls to maintain state
type ResolutionMiddleware func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error)

// DepthLimitError represents an error when max depth is reached
type DepthLimitError struct {
	ImportPath string
	Version    string
	MaxDepth   int
	Path       []string
}

// Error returns a string representation of the error
func (e *DepthLimitError) Error() string {
	return fmt.Sprintf("max depth %d reached for module %s@%s in path: %v",
		e.MaxDepth, e.ImportPath, e.Version, e.Path)
}

// MiddlewareChain represents a chain of middleware
type MiddlewareChain struct {
	middlewares []ResolutionMiddleware
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]ResolutionMiddleware, 0),
	}
}

// Add appends middleware to the chain
func (c *MiddlewareChain) Add(middleware ...ResolutionMiddleware) {
	c.middlewares = append(c.middlewares, middleware...)
}

// Middlewares returns the current middleware chain
func (c *MiddlewareChain) Middlewares() []ResolutionMiddleware {
	return c.middlewares
}

// Execute runs the middleware chain
func (c *MiddlewareChain) Execute(ctx context.Context, importPath, version string, final ResolutionFunc) (*typesys.Module, error) {
	if len(c.middlewares) == 0 {
		return final()
	}

	// Create the middleware chain
	chain := final
	currentCtx := ctx

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		mw := c.middlewares[i]
		nextChain := chain
		chain = func() (*typesys.Module, error) {
			var module *typesys.Module
			var err error
			currentCtx, module, err = mw(currentCtx, importPath, version, nextChain)
			return module, err
		}
	}

	// Execute the chain
	return chain()
}

// WithChainID adds a chain ID to context for middleware tracking
func WithChainID(ctx context.Context, chainID uint64) context.Context {
	return context.WithValue(ctx, contextKeyChainID, chainID)
}

// GetChainID retrieves a chain ID from context
func GetChainID(ctx context.Context) (uint64, bool) {
	val := ctx.Value(contextKeyChainID)
	if val == nil {
		return 0, false
	}
	id, ok := val.(uint64)
	return id, ok
}

// NewDepthLimitingMiddleware creates a middleware that limits resolution depth
func NewDepthLimitingMiddleware(maxDepth int) ResolutionMiddleware {
	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
		// Get current depth from context, defaults to 0
		var currentDepth int
		if depthVal := ctx.Value(contextKeyResolutionDepth); depthVal != nil {
			if depth, ok := depthVal.(int); ok {
				currentDepth = depth
			}
		}

		// Check if we've reached the maximum depth
		if currentDepth >= maxDepth {
			// Extract current path from context
			var resolutionPath []string
			if pathVal := ctx.Value(contextKeyResolutionPath); pathVal != nil {
				if path, ok := pathVal.([]string); ok {
					resolutionPath = path
				}
			}

			// Construct and return a depth limit error
			return ctx, nil, &DepthLimitError{
				ImportPath: importPath,
				Version:    version,
				MaxDepth:   maxDepth,
				Path:       append(resolutionPath, importPath),
			}
		}

		// Create a new context with incremented depth
		newCtx := context.WithValue(ctx, contextKeyResolutionDepth, currentDepth+1)

		// Also add this import path to the resolution path if not already there
		var resolutionPath []string
		if pathVal := ctx.Value(contextKeyResolutionPath); pathVal != nil {
			if path, ok := pathVal.([]string); ok {
				resolutionPath = append([]string{}, path...) // Make a copy
			}
		}
		// Add current import path to resolution path
		resolutionPath = append(resolutionPath, importPath)
		newCtx = context.WithValue(newCtx, contextKeyResolutionPath, resolutionPath)

		// Call the next function with the new context
		module, err := next()

		// Return the new context along with the result
		return newCtx, module, err
	}
}

// NewCachingMiddleware creates a middleware that caches resolved modules
func NewCachingMiddleware() ResolutionMiddleware {
	cache := make(map[string]*typesys.Module)
	mu := &sync.RWMutex{}

	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
		cacheKey := importPath
		if version != "" {
			cacheKey += "@" + version
		}

		// Check cache first with read lock
		mu.RLock()
		cachedModule, found := cache[cacheKey]
		mu.RUnlock()

		if found {
			return ctx, cachedModule, nil
		}

		// Not in cache, need to acquire write lock before calling next
		// This prevents multiple goroutines from resolving the same module
		mu.Lock()

		// Check again after acquiring the write lock
		// Another goroutine might have populated the cache already
		if cachedModule, found := cache[cacheKey]; found {
			mu.Unlock()
			return ctx, cachedModule, nil
		}

		// Call next while holding the lock to prevent duplicate resolution
		module, err := next()

		// Only cache successful results
		if err == nil && module != nil {
			cache[cacheKey] = module
		}

		mu.Unlock()
		return ctx, module, err
	}
}

// NewErrorEnhancerMiddleware creates a middleware that enhances errors with context
func NewErrorEnhancerMiddleware() ResolutionMiddleware {
	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (context.Context, *typesys.Module, error) {
		// Get resolution path from context if available
		var resolutionPath []string
		if path, ok := ctx.Value(contextKeyResolutionPath).([]string); ok {
			resolutionPath = path
		}

		// Call next in chain
		module, err := next()

		// Enhance error with context if needed
		if err != nil {
			// Check if it's already a typed error we don't want to wrap
			switch err.(type) {
			case *DepthLimitError:
				return ctx, nil, err
			}

			// Create enhanced error with context
			return ctx, nil, fmt.Errorf("module resolution failed for %s@%s in path %v: %w",
				importPath, version, resolutionPath, err)
		}

		return ctx, module, nil
	}
}
