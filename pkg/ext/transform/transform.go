// Package transform provides type-safe code transformation operations for Go modules.
// It builds on the typesys package to ensure transformations preserve type correctness.
package transform

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TransformResult contains information about the result of a transformation.
type TransformResult struct {
	// Summary of the transformation
	Summary string

	// Detailed description of changes made
	Details string

	// Number of files affected
	FilesAffected int

	// Whether the transformation was successful
	Success bool

	// Any error that occurred during transformation
	Error error

	// Whether this was a dry run (preview only)
	IsDryRun bool

	// List of affected file paths
	AffectedFiles []string

	// Specific changes that were made (or would be in dry run mode)
	Changes []Change
}

// Change represents a single change made to the code.
type Change struct {
	// File path where the change was made
	FilePath string

	// Position information
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int

	// Original code content
	Original string

	// New code content
	New string

	// Symbol that was affected, if applicable
	AffectedSymbol *typesys.Symbol
}

// Transformer defines the interface for code transformations.
type Transformer interface {
	// Transform applies a transformation to the module and returns the result
	Transform(ctx *Context) (*TransformResult, error)

	// Validate checks if the transformation would be valid without applying it
	Validate(ctx *Context) error

	// Name returns the name of the transformer
	Name() string

	// Description returns a description of what the transformer does
	Description() string
}

// Context provides context for a transformation, including the module,
// index, and any additional options needed for the transformation.
type Context struct {
	// The module to transform
	Module *typesys.Module

	// Index for fast symbol lookups
	Index *index.Index

	// Whether to perform a dry run (preview only)
	DryRun bool

	// Additional options for the transformer
	Options map[string]interface{}

	// Internal state used during transformation
	state map[string]interface{}
}

// NewContext creates a new transformation context.
func NewContext(mod *typesys.Module, idx *index.Index, dryRun bool) *Context {
	return &Context{
		Module:  mod,
		Index:   idx,
		DryRun:  dryRun,
		Options: make(map[string]interface{}),
		state:   make(map[string]interface{}),
	}
}

// SetOption sets an option for the transformation.
func (ctx *Context) SetOption(key string, value interface{}) {
	ctx.Options[key] = value
}

// ChainedTransformer chains multiple transformers together.
type ChainedTransformer struct {
	transformers []Transformer
	name         string
	description  string
}

// NewChainedTransformer creates a new transformer that applies multiple transformations in sequence.
func NewChainedTransformer(name, description string, transformers ...Transformer) *ChainedTransformer {
	return &ChainedTransformer{
		transformers: transformers,
		name:         name,
		description:  description,
	}
}

// Transform applies all transformers in sequence.
func (c *ChainedTransformer) Transform(ctx *Context) (*TransformResult, error) {
	result := &TransformResult{
		Summary:       "Chained transformation",
		Success:       true,
		IsDryRun:      ctx.DryRun,
		AffectedFiles: []string{},
		Changes:       []Change{},
	}

	for _, transformer := range c.transformers {
		tResult, err := transformer.Transform(ctx)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, err
		}

		// If any transformer fails, mark the chain as failed
		if !tResult.Success {
			result.Success = false
			result.Error = tResult.Error
			return result, tResult.Error
		}

		// Aggregate affected files
		for _, file := range tResult.AffectedFiles {
			// Check if already in the list
			found := false
			for _, existing := range result.AffectedFiles {
				if existing == file {
					found = true
					break
				}
			}
			if !found {
				result.AffectedFiles = append(result.AffectedFiles, file)
			}
		}

		// Aggregate changes
		result.Changes = append(result.Changes, tResult.Changes...)
	}

	result.FilesAffected = len(result.AffectedFiles)
	result.Details = fmt.Sprintf("Applied %d transformations", len(c.transformers))

	return result, nil
}

// Validate checks if all transformers in the chain would be valid.
func (c *ChainedTransformer) Validate(ctx *Context) error {
	for _, transformer := range c.transformers {
		if err := transformer.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Name returns the name of the chained transformer.
func (c *ChainedTransformer) Name() string {
	return c.name
}

// Description returns the description of the chained transformer.
func (c *ChainedTransformer) Description() string {
	return c.description
}
