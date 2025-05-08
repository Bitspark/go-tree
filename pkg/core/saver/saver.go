// Package saver defines interfaces and implementations for saving Go modules.
package saver

import (
	"bitspark.dev/go-tree/pkg/core/module"
)

// SaveOptions defines options for module saving
type SaveOptions struct {
	// Whether to format the code
	Format bool

	// Whether to organize imports
	OrganizeImports bool

	// Whether to generate gofmt-compatible output
	Gofmt bool

	// Whether to use tabs (true) or spaces (false) for indentation
	UseTabs bool

	// The number of spaces per indentation level (if UseTabs=false)
	TabWidth int

	// Force overwrite existing files
	Force bool

	// Whether to create a backup of modified files
	CreateBackups bool

	// Save only modified files
	OnlyModified bool
}

// DefaultSaveOptions returns the default save options
func DefaultSaveOptions() SaveOptions {
	return SaveOptions{
		Format:          true,
		OrganizeImports: true,
		Gofmt:           true,
		UseTabs:         true,
		TabWidth:        8,
		Force:           false,
		CreateBackups:   false,
		OnlyModified:    true,
	}
}

// ModuleSaver saves a Go module to disk
type ModuleSaver interface {
	// Save writes a module back to its original location
	Save(module *module.Module) error

	// SaveTo writes a module to a new location
	SaveTo(module *module.Module, dir string) error

	// SaveWithOptions writes a module with custom options
	SaveWithOptions(module *module.Module, options SaveOptions) error

	// SaveToWithOptions writes a module to a new location with custom options
	SaveToWithOptions(module *module.Module, dir string, options SaveOptions) error
}
