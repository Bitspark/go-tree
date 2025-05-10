// Package formatter provides base interfaces and functionality for
// formatting and visualizing Go package data into different output formats.
package formatter

import (
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Formatter defines the interface for different visualization formats
type Formatter interface {
	// Format converts a module to a formatted representation
	Format(mod *typesys.Module, opts *FormatOptions) (string, error)
}

// FormatOptions provides configuration for formatting
type FormatOptions struct {
	// Whether to include type annotations
	IncludeTypeAnnotations bool

	// Whether to include private (unexported) elements
	IncludePrivate bool

	// Whether to include test files
	IncludeTests bool

	// Whether to include generated files
	IncludeGenerated bool

	// Level of detail (1=minimal, 5=complete)
	DetailLevel int

	// Symbol to highlight (if any)
	HighlightSymbol *typesys.Symbol
}

// FormatVisitor implements typesys.TypeSystemVisitor to build formatted output
type FormatVisitor interface {
	typesys.TypeSystemVisitor

	// Result returns the final formatted output
	Result() (string, error)
}

// BaseFormatter provides common functionality for formatters
type BaseFormatter struct {
	visitor FormatVisitor
	options *FormatOptions
}

// NewBaseFormatter creates a new formatter with the given visitor
func NewBaseFormatter(visitor FormatVisitor, options *FormatOptions) *BaseFormatter {
	if options == nil {
		options = &FormatOptions{
			DetailLevel: 3, // Medium detail by default
		}
	}
	return &BaseFormatter{
		visitor: visitor,
		options: options,
	}
}

// Format applies the visitor to a module and returns the formatted result
func (f *BaseFormatter) Format(mod *typesys.Module, opts *FormatOptions) (string, error) {
	// Use provided options or default to the formatter's options
	if opts == nil {
		opts = f.options
	}

	// Store the effective options for use by the visitor
	f.options = opts

	// Walk the module with our visitor
	if err := typesys.Walk(f.visitor, mod); err != nil {
		return "", err
	}

	// Get the result from the visitor
	return f.visitor.Result()
}

// FormatTypeSignature returns a formatted type signature with options for detail level
func FormatTypeSignature(typ typesys.Symbol, includeTypes bool, detailLevel int) string {
	// Just a basic implementation - real one would be more sophisticated
	name := typ.Name

	if includeTypes {
		if typ.TypeInfo != nil {
			// Add type information based on detail level
			switch detailLevel {
			case 1:
				// Just the basic type name
				name += " " + typ.TypeInfo.String()
			case 2, 3:
				// More detailed type info
				name += " " + typ.TypeInfo.String()
			case 4, 5:
				// Full type information
				name += " " + typ.TypeInfo.String()
			}
		}
	}

	return name
}

// FormatSymbolName returns a formatted symbol name with optional qualifiers
func FormatSymbolName(sym *typesys.Symbol, showPackage bool) string {
	if sym == nil {
		return ""
	}

	if showPackage && sym.Package != nil {
		return sym.Package.Name + "." + sym.Name
	}

	return sym.Name
}

// BuildQualifiedName builds a fully qualified name for a symbol
func BuildQualifiedName(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	parts := []string{sym.Name}

	// Add parent names if any
	parent := sym.Parent
	for parent != nil {
		parts = append([]string{parent.Name}, parts...)
		parent = parent.Parent
	}

	// Add package name
	if sym.Package != nil {
		parts = append([]string{sym.Package.Name}, parts...)
	}

	return strings.Join(parts, ".")
}

// ShouldIncludeSymbol determines if a symbol should be included based on options
func ShouldIncludeSymbol(sym *typesys.Symbol, opts *FormatOptions) bool {
	if sym == nil {
		return false
	}

	// Check if we should include private symbols
	if !opts.IncludePrivate && !sym.Exported {
		return false
	}

	// Add more filtering based on options as needed

	return true
}
