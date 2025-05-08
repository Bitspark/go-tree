// Package transform defines interfaces and implementations for transforming Go modules.
package transform

import (
	"bitspark.dev/go-tree/pkg/core/module"
)

// ModuleTransformer defines an interface for transforming a Go module
type ModuleTransformer interface {
	// Transform applies transformations to a module
	Transform(mod *module.Module) error

	// Name returns the name of the transformer
	Name() string

	// Description returns a description of what the transformer does
	Description() string
}

// TransformationResult contains information about a transformation
type TransformationResult struct {
	// Summary of changes made
	Summary string

	// Details of the transformation
	Details string

	// Number of files affected
	FilesAffected int

	// Whether the transformation was successful
	Success bool

	// Any error that occurred during transformation
	Error error
}

// ChainedTransformer chains multiple transformers together
type ChainedTransformer struct {
	transformers []ModuleTransformer
	name         string
	description  string
}

// NewChainedTransformer creates a new transformer that applies multiple transformations in sequence
func NewChainedTransformer(name, description string, transformers ...ModuleTransformer) *ChainedTransformer {
	return &ChainedTransformer{
		transformers: transformers,
		name:         name,
		description:  description,
	}
}

// Transform applies all transformers in sequence
func (c *ChainedTransformer) Transform(mod *module.Module) error {
	for _, transformer := range c.transformers {
		if err := transformer.Transform(mod); err != nil {
			return err
		}
	}
	return nil
}

// Name returns the name of the chained transformer
func (c *ChainedTransformer) Name() string {
	return c.name
}

// Description returns the description of the chained transformer
func (c *ChainedTransformer) Description() string {
	return c.description
}
