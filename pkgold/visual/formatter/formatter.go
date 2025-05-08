// Package formatter provides base interfaces and functionality for
// formatting and visualizing Go package data into different output formats.
package formatter

import (
	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/visitor"
)

// Formatter defines the interface for different visualization formats
type Formatter interface {
	// Format converts a module to a formatted representation
	Format(mod *module.Module) (string, error)
}

// FormatVisitor implements visitor.ModuleVisitor to format modules
// into different output formats
type FormatVisitor interface {
	visitor.ModuleVisitor

	// Result returns the final formatted output
	Result() (string, error)
}

// BaseFormatter provides common functionality for formatters
type BaseFormatter struct {
	visitor FormatVisitor
}

// NewBaseFormatter creates a new formatter with the given visitor
func NewBaseFormatter(visitor FormatVisitor) *BaseFormatter {
	return &BaseFormatter{visitor: visitor}
}

// Format applies the visitor to a module and returns the formatted result
func (f *BaseFormatter) Format(mod *module.Module) (string, error) {
	// Create a walker to traverse the module structure
	walker := visitor.NewModuleWalker(f.visitor)

	// Walk the module
	if err := walker.Walk(mod); err != nil {
		return "", err
	}

	// Get the result from the visitor
	return f.visitor.Result()
}
