package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
	"bitspark.dev/go-tree/pkg/visual/formatter"
)

// MarkdownVisitor traverses the type system and builds Markdown representations
type MarkdownVisitor struct {
	// Output buffer for Markdown content
	buffer *bytes.Buffer

	// Formatting options
	options *formatter.FormatOptions

	// Tracking state
	currentPackage *typesys.Package
	currentSymbol  *typesys.Symbol

	// Contains all symbols we've already visited to avoid duplicates
	visitedSymbols map[string]bool
}

// NewMarkdownVisitor creates a new Markdown visitor with the given options
func NewMarkdownVisitor(options *formatter.FormatOptions) *MarkdownVisitor {
	if options == nil {
		options = &formatter.FormatOptions{
			DetailLevel: 3, // Medium detail by default
		}
	}

	return &MarkdownVisitor{
		buffer:         bytes.NewBuffer(nil),
		options:        options,
		visitedSymbols: make(map[string]bool),
	}
}

// Result returns the generated Markdown content
func (v *MarkdownVisitor) Result() (string, error) {
	return v.buffer.String(), nil
}

// Write adds content to the buffer
func (v *MarkdownVisitor) Write(format string, args ...interface{}) {
	fmt.Fprintf(v.buffer, format, args...)
}

// VisitModule processes a module
func (v *MarkdownVisitor) VisitModule(mod *typesys.Module) error {
	// Add module header
	v.Write("# Module: %s\n\n", mod.Path)
	v.Write("Go Version: %s\n\n", mod.GoVersion)

	// Add table of contents if we have enough packages
	if len(mod.Packages) > 3 {
		v.Write("## Table of Contents\n\n")

		for _, pkg := range mod.Packages {
			v.Write("- [Package %s](#package-%s)\n", pkg.Name, pkg.Name)
		}

		v.Write("\n")
	}

	return nil
}

// VisitFile processes a file
func (v *MarkdownVisitor) VisitFile(file *typesys.File) error {
	// We don't need special handling for files in the Markdown output
	// The symbols in the file will be processed individually
	return nil
}

// VisitSymbol is a generic method that handles any symbol
func (v *MarkdownVisitor) VisitSymbol(sym *typesys.Symbol) error {
	// We handle symbols in their specific visit methods
	return nil
}

// VisitPackage processes a package
func (v *MarkdownVisitor) VisitPackage(pkg *typesys.Package) error {
	v.currentPackage = pkg

	// Add package header
	v.Write("## Package: %s\n\n", pkg.Name)
	v.Write("Import Path: `%s`\n\n", pkg.ImportPath)

	// First process types
	v.Write("### Types\n\n")

	// Types will be processed by the type visitor methods

	return nil
}

// AfterVisitPackage is called after all symbols in a package have been processed
func (v *MarkdownVisitor) AfterVisitPackage(pkg *typesys.Package) error {
	// Add section for functions
	v.Write("\n### Functions\n\n")

	// Functions will be processed by the function visitor method

	// Add section for variables and constants
	v.Write("\n### Variables and Constants\n\n")

	// Variables and constants will be processed by their visitor methods

	v.Write("\n---\n\n") // Add separator between packages

	v.currentPackage = nil

	return nil
}

// getSymbolAnchor returns the anchor ID for a symbol
func (v *MarkdownVisitor) getSymbolAnchor(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	return strings.ToLower(strings.ReplaceAll(sym.Name, " ", "-"))
}

// renderSymbolHeader generates the Markdown for a symbol header
func (v *MarkdownVisitor) renderSymbolHeader(sym *typesys.Symbol) {
	if sym == nil {
		return
	}

	// Symbol header with anchor
	v.Write("<a id=\"%s\"></a>\n", v.getSymbolAnchor(sym))
	v.Write("#### %s\n\n", sym.Name)

	// Add visibility badge
	if sym.Exported {
		v.Write("**Exported** | ")
	} else {
		v.Write("**Private** | ")
	}

	// Add kind badge
	v.Write("**%s**", sym.Kind.String())

	// Add type-specific tags
	if sym.Kind == typesys.KindType && sym.TypeInfo != nil {
		typeStr := sym.TypeInfo.String()
		if strings.Contains(typeStr, "interface") {
			v.Write(" | **Interface**")
		} else if strings.Contains(typeStr, "struct") {
			v.Write(" | **Struct**")
		}
	}

	v.Write("\n\n")

	// Type information if available and requested
	if v.options.IncludeTypeAnnotations && sym.TypeInfo != nil {
		v.Write("```go\n%s\n```\n\n", sym.TypeInfo.String())
	}
}

// renderSymbolFooter adds closing elements for a symbol
func (v *MarkdownVisitor) renderSymbolFooter(sym *typesys.Symbol) {
	// Add references section if we're showing relationships and at sufficient detail level
	if v.options.DetailLevel >= 3 && sym != nil {
		refs, err := sym.Package.Module.FindAllReferences(sym)
		if err == nil && len(refs) > 0 {
			v.Write("**References:** ")

			// Only show a limited number of references based on detail level
			maxRefs := 3
			if v.options.DetailLevel >= 4 {
				maxRefs = 5
			}
			if v.options.DetailLevel >= 5 {
				maxRefs = len(refs) // Show all
			}

			for i, ref := range refs {
				if i >= maxRefs {
					v.Write(" and %d more", len(refs)-maxRefs)
					break
				}

				// Format the reference location
				if ref.File != nil {
					if pos := ref.GetPosition(); pos != nil {
						if i > 0 {
							v.Write(", ")
						}
						v.Write("`%s:%d`", ref.File.Path, pos.LineStart)
					}
				}
			}

			v.Write("\n\n")
		}
	}

	v.Write("\n")
}

// VisitType processes a type symbol
func (v *MarkdownVisitor) VisitType(sym *typesys.Symbol) error {
	if !formatter.ShouldIncludeSymbol(sym, v.options) {
		return nil
	}

	if v.visitedSymbols[sym.ID] {
		return nil // Already processed this symbol
	}
	v.visitedSymbols[sym.ID] = true

	v.currentSymbol = sym
	v.renderSymbolHeader(sym)

	// Type-specific content would go here
	// For example, showing struct fields or interface methods

	v.renderSymbolFooter(sym)
	v.currentSymbol = nil

	return nil
}

// VisitFunction processes a function symbol
func (v *MarkdownVisitor) VisitFunction(sym *typesys.Symbol) error {
	if !formatter.ShouldIncludeSymbol(sym, v.options) {
		return nil
	}

	if v.visitedSymbols[sym.ID] {
		return nil // Already processed this symbol
	}
	v.visitedSymbols[sym.ID] = true

	v.currentSymbol = sym
	v.renderSymbolHeader(sym)

	// Function-specific content would go here
	// For example, showing parameter and return types

	v.renderSymbolFooter(sym)
	v.currentSymbol = nil

	return nil
}

// VisitVariable processes a variable symbol
func (v *MarkdownVisitor) VisitVariable(sym *typesys.Symbol) error {
	if !formatter.ShouldIncludeSymbol(sym, v.options) {
		return nil
	}

	if v.visitedSymbols[sym.ID] {
		return nil // Already processed this symbol
	}
	v.visitedSymbols[sym.ID] = true

	v.currentSymbol = sym
	v.renderSymbolHeader(sym)

	// Variable-specific content would go here

	v.renderSymbolFooter(sym)
	v.currentSymbol = nil

	return nil
}

// VisitConstant processes a constant symbol
func (v *MarkdownVisitor) VisitConstant(sym *typesys.Symbol) error {
	if !formatter.ShouldIncludeSymbol(sym, v.options) {
		return nil
	}

	if v.visitedSymbols[sym.ID] {
		return nil // Already processed this symbol
	}
	v.visitedSymbols[sym.ID] = true

	v.currentSymbol = sym
	v.renderSymbolHeader(sym)

	// Constant-specific content would go here

	v.renderSymbolFooter(sym)
	v.currentSymbol = nil

	return nil
}

// VisitImport processes an import
func (v *MarkdownVisitor) VisitImport(imp *typesys.Import) error {
	// Imports are typically shown as part of the package, not individually
	return nil
}

// VisitInterface processes an interface type
func (v *MarkdownVisitor) VisitInterface(sym *typesys.Symbol) error {
	// This is called after VisitType for interface types
	// We could add interface-specific details here
	return nil
}

// VisitStruct processes a struct type
func (v *MarkdownVisitor) VisitStruct(sym *typesys.Symbol) error {
	// This is called after VisitType for struct types
	// We could add struct-specific details here
	return nil
}

// VisitMethod processes a method
func (v *MarkdownVisitor) VisitMethod(sym *typesys.Symbol) error {
	// Similar to VisitFunction, but for methods
	return v.VisitFunction(sym)
}

// VisitField processes a field symbol
func (v *MarkdownVisitor) VisitField(sym *typesys.Symbol) error {
	// Similar to VisitVariable, but for struct fields
	return v.VisitVariable(sym)
}

// VisitParameter processes a parameter symbol
func (v *MarkdownVisitor) VisitParameter(sym *typesys.Symbol) error {
	// Parameters are typically shown as part of their function, not individually
	return nil
}

// VisitGenericType processes a generic type
func (v *MarkdownVisitor) VisitGenericType(sym *typesys.Symbol) error {
	// This is called for generic types (Go 1.18+)
	return v.VisitType(sym)
}

// VisitTypeParameter processes a type parameter
func (v *MarkdownVisitor) VisitTypeParameter(sym *typesys.Symbol) error {
	// This is called for type parameters in generic types
	return nil
}

// AfterVisitModule is called after all packages in a module have been processed
func (v *MarkdownVisitor) AfterVisitModule(mod *typesys.Module) error {
	return nil
}
