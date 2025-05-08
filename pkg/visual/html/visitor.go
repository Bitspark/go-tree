package html

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
	"bitspark.dev/go-tree/pkg/visual/formatter"
)

// HTMLVisitor traverses the type system and builds HTML representations
type HTMLVisitor struct {
	// Output buffer for HTML content
	buffer *bytes.Buffer

	// Formatting options
	options *formatter.FormatOptions

	// Tracking state
	currentPackage *typesys.Package
	currentSymbol  *typesys.Symbol
	indentLevel    int

	// Contains all symbols we've already visited to avoid duplicates
	visitedSymbols map[string]bool
}

// NewHTMLVisitor creates a new HTML visitor with the given options
func NewHTMLVisitor(options *formatter.FormatOptions) *HTMLVisitor {
	if options == nil {
		options = &formatter.FormatOptions{
			DetailLevel: 3, // Medium detail by default
		}
	}

	return &HTMLVisitor{
		buffer:         bytes.NewBuffer(nil),
		options:        options,
		indentLevel:    0,
		visitedSymbols: make(map[string]bool),
	}
}

// Result returns the generated HTML content
func (v *HTMLVisitor) Result() (string, error) {
	return v.buffer.String(), nil
}

// Write adds content to the buffer
func (v *HTMLVisitor) Write(format string, args ...interface{}) {
	fmt.Fprintf(v.buffer, format, args...)
}

// Indent returns the current indentation string
func (v *HTMLVisitor) Indent() string {
	return strings.Repeat("    ", v.indentLevel)
}

// VisitModule processes a module
func (v *HTMLVisitor) VisitModule(mod *typesys.Module) error {
	v.Write("<div class=\"packages\">\n")

	// Modules don't need special processing - we'll handle packages individually
	return nil
}

// VisitPackage processes a package
func (v *HTMLVisitor) VisitPackage(pkg *typesys.Package) error {
	v.currentPackage = pkg

	v.Write("%s<div class=\"package\" id=\"pkg-%s\">\n", v.Indent(), template.HTMLEscapeString(pkg.Name))
	v.indentLevel++

	v.Write("%s<div class=\"package-header\">\n", v.Indent())
	v.Write("%s    <h2>Package %s</h2>\n", v.Indent(), template.HTMLEscapeString(pkg.Name))
	v.Write("%s    <div class=\"package-import\">%s</div>\n", v.Indent(), template.HTMLEscapeString(pkg.ImportPath))
	v.Write("%s</div>\n", v.Indent())

	// Package description could go here

	// Add symbols section
	v.Write("%s<div class=\"symbols\">\n", v.Indent())

	// First process types
	v.Write("%s    <h3>Types</h3>\n", v.Indent())
	v.Write("%s    <div class=\"type-list\">\n", v.Indent())

	// Types will be processed by the type visitor methods

	return nil
}

// AfterVisitPackage is called after all symbols in a package have been processed
func (v *HTMLVisitor) AfterVisitPackage(pkg *typesys.Package) error {
	v.Write("%s    </div>\n", v.Indent()) // Close type-list

	// Process functions
	v.Write("%s    <h3>Functions</h3>\n", v.Indent())
	v.Write("%s    <div class=\"function-list\">\n", v.Indent())

	// Functions will be processed by the function visitor method

	v.Write("%s    </div>\n", v.Indent()) // Close function-list

	// Process variables and constants
	v.Write("%s    <h3>Variables and Constants</h3>\n", v.Indent())
	v.Write("%s    <div class=\"var-const-list\">\n", v.Indent())

	// Variables and constants will be processed by their visitor methods

	v.Write("%s    </div>\n", v.Indent()) // Close var-const-list

	v.Write("%s</div>\n", v.Indent()) // Close symbols

	v.indentLevel--
	v.Write("%s</div>\n", v.Indent()) // Close package

	v.currentPackage = nil

	return nil
}

// getSymbolClass returns the CSS class for a symbol based on its kind
func (v *HTMLVisitor) getSymbolClass(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	var kindClass string
	switch sym.Kind {
	case typesys.KindFunction:
		kindClass = "symbol-fn"
	case typesys.KindType:
		kindClass = "symbol-type"
	case typesys.KindVariable:
		kindClass = "symbol-var"
	case typesys.KindConstant:
		kindClass = "symbol-const"
	case typesys.KindField:
		kindClass = "symbol-field"
	case typesys.KindPackage:
		kindClass = "symbol-pkg"
	default:
		kindClass = ""
	}

	var exportedClass string
	if sym.Exported {
		exportedClass = "exported"
	} else {
		exportedClass = "private"
	}

	return fmt.Sprintf("symbol %s %s", kindClass, exportedClass)
}

// renderSymbolHeader generates the HTML for a symbol header
func (v *HTMLVisitor) renderSymbolHeader(sym *typesys.Symbol) {
	if sym == nil {
		return
	}

	symClass := v.getSymbolClass(sym)
	highlightClass := ""
	if v.options.HighlightSymbol != nil && sym.ID == v.options.HighlightSymbol.ID {
		highlightClass = "highlight"
	}

	v.Write("%s<div class=\"%s %s\" id=\"sym-%s\">\n", v.Indent(), symClass, highlightClass, template.HTMLEscapeString(sym.ID))
	v.indentLevel++

	// Symbol name and tags
	v.Write("%s<div class=\"symbol-header\">\n", v.Indent())
	v.Write("%s    <span class=\"symbol-name\">%s</span>\n", v.Indent(), template.HTMLEscapeString(sym.Name))

	// Add visibility tag
	if sym.Exported {
		v.Write("%s    <span class=\"tag tag-exported\">exported</span>\n", v.Indent())
	} else {
		v.Write("%s    <span class=\"tag tag-private\">private</span>\n", v.Indent())
	}

	// Add kind-specific tags
	switch sym.Kind {
	case typesys.KindType:
		// Add type-specific tag if we can determine it
		if sym.TypeInfo != nil {
			typeStr := sym.TypeInfo.String()
			if strings.Contains(typeStr, "interface") {
				v.Write("%s    <span class=\"tag tag-interface\">interface</span>\n", v.Indent())
			} else if strings.Contains(typeStr, "struct") {
				v.Write("%s    <span class=\"tag tag-struct\">struct</span>\n", v.Indent())
			}
		}
	}

	v.Write("%s</div>\n", v.Indent())

	// Type information if available and requested
	if v.options.IncludeTypeAnnotations && sym.TypeInfo != nil {
		v.Write("%s<div class=\"type-info\">%s</div>\n", v.Indent(), template.HTMLEscapeString(sym.TypeInfo.String()))
	}
}

// renderSymbolFooter closes a symbol div
func (v *HTMLVisitor) renderSymbolFooter() {
	// Add references section if we're showing relationships and at sufficient detail level
	if v.options.DetailLevel >= 3 && v.currentSymbol != nil {
		refs, err := v.currentSymbol.Package.Module.FindAllReferences(v.currentSymbol)
		if err == nil && len(refs) > 0 {
			v.Write("%s<div class=\"references\">\n", v.Indent())
			v.Write("%s    <div class=\"references-title\">References (%d)</div>\n", v.Indent(), len(refs))

			// Only show a limited number of references based on detail level
			maxRefs := 5
			if v.options.DetailLevel >= 4 {
				maxRefs = 10
			}
			if v.options.DetailLevel >= 5 {
				maxRefs = len(refs) // Show all
			}

			for i, ref := range refs {
				if i >= maxRefs {
					v.Write("%s    <div class=\"reference-more\">... and %d more</div>\n", v.Indent(), len(refs)-maxRefs)
					break
				}

				// Format the reference location
				if ref.File != nil {
					if pos := ref.GetPosition(); pos != nil {
						v.Write("%s    <div class=\"reference\">%s:%d</div>\n", v.Indent(),
							template.HTMLEscapeString(ref.File.Path),
							pos.LineStart,
						)
					}
				}
			}

			v.Write("%s</div>\n", v.Indent())
		}
	}

	v.indentLevel--
	v.Write("%s</div>\n", v.Indent()) // Close symbol
}

// VisitType processes a type symbol
func (v *HTMLVisitor) VisitType(sym *typesys.Symbol) error {
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

	v.renderSymbolFooter()
	v.currentSymbol = nil

	return nil
}

// VisitFunction processes a function symbol
func (v *HTMLVisitor) VisitFunction(sym *typesys.Symbol) error {
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

	v.renderSymbolFooter()
	v.currentSymbol = nil

	return nil
}

// VisitVariable processes a variable symbol
func (v *HTMLVisitor) VisitVariable(sym *typesys.Symbol) error {
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

	v.renderSymbolFooter()
	v.currentSymbol = nil

	return nil
}

// VisitConstant processes a constant symbol
func (v *HTMLVisitor) VisitConstant(sym *typesys.Symbol) error {
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

	v.renderSymbolFooter()
	v.currentSymbol = nil

	return nil
}

// VisitImport processes an import
func (v *HTMLVisitor) VisitImport(imp *typesys.Import) error {
	// Imports are typically shown as part of the package, not individually
	return nil
}

// VisitInterface processes an interface type
func (v *HTMLVisitor) VisitInterface(sym *typesys.Symbol) error {
	// This is called after VisitType for interface types
	// We could add interface-specific details here
	return nil
}

// VisitStruct processes a struct type
func (v *HTMLVisitor) VisitStruct(sym *typesys.Symbol) error {
	// This is called after VisitType for struct types
	// We could add struct-specific details here
	return nil
}

// VisitMethod processes a method
func (v *HTMLVisitor) VisitMethod(sym *typesys.Symbol) error {
	// Similar to VisitFunction, but for methods
	// VisitMethod is called for methods on types
	return v.VisitFunction(sym)
}

// VisitField processes a field symbol
func (v *HTMLVisitor) VisitField(sym *typesys.Symbol) error {
	// Similar to VisitVariable, but for struct fields
	return v.VisitVariable(sym)
}

// VisitGenericType processes a generic type
func (v *HTMLVisitor) VisitGenericType(sym *typesys.Symbol) error {
	// This is called for generic types (Go 1.18+)
	return v.VisitType(sym)
}

// VisitTypeParameter processes a type parameter
func (v *HTMLVisitor) VisitTypeParameter(sym *typesys.Symbol) error {
	// This is called for type parameters in generic types
	return nil
}

// VisitFile processes a file
func (v *HTMLVisitor) VisitFile(file *typesys.File) error {
	// We don't need special handling for files in the HTML output
	// The symbols in the file will be processed individually
	return nil
}

// VisitSymbol is a generic method that handles any symbol
func (v *HTMLVisitor) VisitSymbol(sym *typesys.Symbol) error {
	// We handle symbols in their specific visit methods
	// This is called before the specific methods like VisitType, VisitFunction, etc.
	return nil
}

// VisitParameter processes a parameter symbol
func (v *HTMLVisitor) VisitParameter(sym *typesys.Symbol) error {
	// Parameters are typically shown as part of their function, not individually
	return nil
}
