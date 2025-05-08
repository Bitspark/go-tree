package toolkit

import (
	"context"

	"bitspark.dev/go-tree/pkg/typesys"
)

// ResolutionFunc represents the next resolver in the chain
type ResolutionFunc func() (*typesys.Module, error)

// ResolutionMiddleware intercepts module resolution requests
type ResolutionMiddleware func(ctx context.Context, importPath, version string, next ResolutionFunc) (*typesys.Module, error)

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
