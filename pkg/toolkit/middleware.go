package toolkit

import (
	"context"
	"fmt"
	"sync"

	"bitspark.dev/go-tree/pkg/typesys"
)

// Context keys for passing data through middleware chain
type contextKey string

const (
	// For tracking resolution path
	contextKeyResolutionPath contextKey = "resolutionPath"
	// For tracking resolution depth
	contextKeyResolutionDepth contextKey = "resolutionDepth"
)

// ResolutionFunc represents the next resolver in the chain
type ResolutionFunc func() (*typesys.Module, error)

// ResolutionMiddleware intercepts module resolution requests
type ResolutionMiddleware func(ctx context.Context, importPath, version string, next ResolutionFunc) (*typesys.Module, error)

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
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		mw := c.middlewares[i]
		nextChain := chain
		chain = func() (*typesys.Module, error) {
			return mw(ctx, importPath, version, nextChain)
		}
	}

	// Execute the chain
	return chain()
}

// NewDepthLimitingMiddleware creates a middleware that limits resolution depth
func NewDepthLimitingMiddleware(maxDepth int) ResolutionMiddleware {
	depthMap := make(map[string]int) // Keep track of depth per import path
	mu := &sync.RWMutex{}

	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (*typesys.Module, error) {
		// Extract current path from context or create new path
		var resolutionPath []string
		if path, ok := ctx.Value(contextKeyResolutionPath).([]string); ok {
			resolutionPath = path
		} else {
			resolutionPath = []string{}
		}

		// Check current depth for this import path
		mu.RLock()
		currentDepth := depthMap[importPath]
		mu.RUnlock()

		if currentDepth >= maxDepth {
			return nil, &DepthLimitError{
				ImportPath: importPath,
				Version:    version,
				MaxDepth:   maxDepth,
				Path:       append(resolutionPath, importPath),
			}
		}

		// Update depth and path for next calls
		mu.Lock()
		depthMap[importPath] = currentDepth + 1
		mu.Unlock()

		// The context will be passed implicitly to the next middlewares
		// but we can't directly change the context for the current function.
		// This is a limitation of the middleware design - we accept it for simplicity

		// Call next middleware/resolver
		module, err := next()

		// Reset depth after completion
		mu.Lock()
		depthMap[importPath] = currentDepth
		mu.Unlock()

		return module, err
	}
}

// NewCachingMiddleware creates a middleware that caches resolved modules
func NewCachingMiddleware() ResolutionMiddleware {
	cache := make(map[string]*typesys.Module)
	mu := &sync.RWMutex{}

	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (*typesys.Module, error) {
		cacheKey := importPath
		if version != "" {
			cacheKey += "@" + version
		}

		// Check cache first with read lock
		mu.RLock()
		if cachedModule, ok := cache[cacheKey]; ok {
			mu.RUnlock()
			return cachedModule, nil
		}
		mu.RUnlock()

		// Not in cache, proceed with resolution
		module, err := next()
		if err != nil {
			return nil, err
		}

		// Cache the result with write lock
		if module != nil {
			mu.Lock()
			cache[cacheKey] = module
			mu.Unlock()
		}

		return module, nil
	}
}

// NewErrorEnhancerMiddleware creates a middleware that enhances errors with context
func NewErrorEnhancerMiddleware() ResolutionMiddleware {
	return func(ctx context.Context, importPath, version string, next ResolutionFunc) (*typesys.Module, error) {
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
				return nil, err
			}

			// Create enhanced error with context
			return nil, fmt.Errorf("module resolution failed for %s@%s in path %v: %w",
				importPath, version, resolutionPath, err)
		}

		return module, nil
	}
}
