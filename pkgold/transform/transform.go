// Package transform defines interfaces and implementations for transforming Go modules.
package transform

import (
	"bitspark.dev/go-tree/pkgold/core/module"
)

// ModuleTransformer defines an interface for transforming a Go module
type ModuleTransformer interface {
	// Transform applies transformations to a module and returns the result
	Transform(mod *module.Module) *TransformationResult

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

	// Whether this was a dry run (preview only)
	IsDryRun bool

	// List of affected file paths
	AffectedFiles []string

	// Specific changes that would be made (used in dry run mode)
	Changes []ChangePreview
}

// ChangePreview represents a single change that would be made
type ChangePreview struct {
	// File path relative to module root
	FilePath string

	// Line number where the change occurs
	LineNumber int

	// Original text
	Original string

	// New text that would replace the original
	New string
}

// ChainedTransformer chains multiple transformers together
type ChainedTransformer struct {
	transformers []ModuleTransformer
	name         string
	description  string
	dryRun       bool
}

// NewChainedTransformer creates a new transformer that applies multiple transformations in sequence
func NewChainedTransformer(name, description string, dryRun bool, transformers ...ModuleTransformer) *ChainedTransformer {
	return &ChainedTransformer{
		transformers: transformers,
		name:         name,
		description:  description,
		dryRun:       dryRun,
	}
}

// Transform applies all transformers in sequence
func (c *ChainedTransformer) Transform(mod *module.Module) *TransformationResult {
	result := &TransformationResult{
		Summary:       "Chained transformation",
		Success:       true,
		IsDryRun:      c.dryRun,
		AffectedFiles: []string{},
		Changes:       []ChangePreview{},
	}

	for _, transformer := range c.transformers {
		tResult := transformer.Transform(mod)

		// If any transformer fails, mark the chain as failed
		if !tResult.Success {
			result.Success = false
			result.Error = tResult.Error
			return result
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

		// Aggregate changes for dry run
		if c.dryRun {
			result.Changes = append(result.Changes, tResult.Changes...)
		}
	}

	result.FilesAffected = len(result.AffectedFiles)
	return result
}

// Name returns the name of the chained transformer
func (c *ChainedTransformer) Name() string {
	return c.name
}

// Description returns the description of the chained transformer
func (c *ChainedTransformer) Description() string {
	return c.description
}
