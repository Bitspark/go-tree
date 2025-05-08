// Package markdown provides functionality for generating Markdown documentation
// from Go-Tree module data.
package markdown

import (
	"bytes"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/module"
)

// MarkdownVisitor implements the visitor interface for Markdown output
type MarkdownVisitor struct {
	options     Options
	buffer      bytes.Buffer
	packageName string
}

// NewMarkdownVisitor creates a new Markdown visitor
func NewMarkdownVisitor(options Options) *MarkdownVisitor {
	return &MarkdownVisitor{
		options: options,
	}
}

// VisitModule processes a module
func (v *MarkdownVisitor) VisitModule(mod *module.Module) error {
	// Add module title
	v.buffer.WriteString(fmt.Sprintf("# Module %s\n\n", mod.Path))

	// Module doesn't have a Doc field, so we won't add module documentation

	return nil
}

// VisitPackage processes a package
func (v *MarkdownVisitor) VisitPackage(pkg *module.Package) error {
	v.packageName = pkg.Name

	// Add package title
	v.buffer.WriteString("## Package " + pkg.Name + "\n\n")

	// Add package documentation if available
	if pkg.Documentation != "" {
		v.buffer.WriteString(pkg.Documentation + "\n\n")
	}

	return nil
}

// VisitFile processes a file
func (v *MarkdownVisitor) VisitFile(file *module.File) error {
	// Files are not typically represented in Markdown documentation
	// at the file level, so we'll just ignore this visit.
	return nil
}

// VisitType processes a type declaration
func (v *MarkdownVisitor) VisitType(typ *module.Type) error {
	// Add type header
	v.buffer.WriteString(fmt.Sprintf("### Type: %s (%s)\n\n", typ.Name, typ.Kind))

	// Add type documentation if available
	if typ.Doc != "" {
		v.buffer.WriteString(typ.Doc + "\n\n")
	}

	// Type doesn't have a Code field, so we'll just include a placeholder for the code block
	if v.options.IncludeCodeBlocks {
		v.buffer.WriteString("```go\n")
		// Show type definition based on available fields
		v.buffer.WriteString(fmt.Sprintf("type %s %s\n", typ.Name, typ.Kind))
		v.buffer.WriteString("```\n\n")
	}

	// For structs, fields will be processed by VisitField

	return nil
}

// VisitFunction processes a function declaration
func (v *MarkdownVisitor) VisitFunction(fn *module.Function) error {
	// Add function header
	v.buffer.WriteString(fmt.Sprintf("### Function: %s\n\n", fn.Name))

	// Add function documentation if available
	if fn.Doc != "" {
		v.buffer.WriteString(fn.Doc + "\n\n")
	}

	// Add signature if available
	if fn.Signature != "" {
		v.buffer.WriteString(fmt.Sprintf("**Signature:** `%s`\n\n", fn.Signature))
	}

	// Function doesn't have a Code field, so we'll just include the signature in the code block
	if v.options.IncludeCodeBlocks && fn.Signature != "" {
		v.buffer.WriteString("```go\n")
		v.buffer.WriteString(fmt.Sprintf("func %s%s\n", fn.Name, fn.Signature))
		v.buffer.WriteString("```\n\n")
	}

	return nil
}

// VisitMethod processes a method declaration
func (v *MarkdownVisitor) VisitMethod(method *module.Method) error {
	// Method doesn't have a Function field, it's a standalone entity
	// Format method header with receiver type from the method's parent type
	receiverStr := ""
	if method.Parent != nil {
		receiverStr = method.Parent.Name
		v.buffer.WriteString(fmt.Sprintf("### Method: (%s) %s\n\n", receiverStr, method.Name))
	} else {
		v.buffer.WriteString(fmt.Sprintf("### Method: %s\n\n", method.Name))
	}

	// Add method documentation if available
	if method.Doc != "" {
		v.buffer.WriteString(method.Doc + "\n\n")
	}

	// Add signature if available
	if method.Signature != "" {
		v.buffer.WriteString(fmt.Sprintf("**Signature:** `%s`\n\n", method.Signature))
	}

	// Method doesn't have a Code field, so we'll just include the signature in the code block
	if v.options.IncludeCodeBlocks && method.Signature != "" {
		v.buffer.WriteString("```go\n")
		if receiverStr != "" {
			v.buffer.WriteString(fmt.Sprintf("func (%s) %s%s\n", receiverStr, method.Name, method.Signature))
		} else {
			v.buffer.WriteString(fmt.Sprintf("func %s%s\n", method.Name, method.Signature))
		}
		v.buffer.WriteString("```\n\n")
	}

	return nil
}

// VisitField processes a struct field
func (v *MarkdownVisitor) VisitField(field *module.Field) error {
	// Fields are usually processed as part of the struct type
	// We could accumulate them here and output them when we've seen all fields
	// For now, we'll just ignore individual fields
	return nil
}

// VisitVariable processes a variable declaration
func (v *MarkdownVisitor) VisitVariable(variable *module.Variable) error {
	// Add variable information
	v.buffer.WriteString(fmt.Sprintf("### Variable: %s\n\n", variable.Name))

	if variable.Doc != "" {
		v.buffer.WriteString(variable.Doc + "\n\n")
	}

	v.buffer.WriteString(fmt.Sprintf("**Type:** %s\n\n", variable.Type))

	// Variable doesn't have a Code field, so we'll just show a simplified declaration
	if v.options.IncludeCodeBlocks {
		v.buffer.WriteString("```go\n")
		v.buffer.WriteString(fmt.Sprintf("var %s %s\n", variable.Name, variable.Type))
		v.buffer.WriteString("```\n\n")
	}

	return nil
}

// VisitConstant processes a constant declaration
func (v *MarkdownVisitor) VisitConstant(constant *module.Constant) error {
	// Add constant information
	v.buffer.WriteString(fmt.Sprintf("### Constant: %s\n\n", constant.Name))

	if constant.Doc != "" {
		v.buffer.WriteString(constant.Doc + "\n\n")
	}

	v.buffer.WriteString(fmt.Sprintf("**Type:** %s\n\n", constant.Type))
	v.buffer.WriteString(fmt.Sprintf("**Value:** %s\n\n", constant.Value))

	// Constant doesn't have a Code field, so we'll just show a simplified declaration
	if v.options.IncludeCodeBlocks {
		v.buffer.WriteString("```go\n")
		v.buffer.WriteString(fmt.Sprintf("const %s %s = %s\n", constant.Name, constant.Type, constant.Value))
		v.buffer.WriteString("```\n\n")
	}

	return nil
}

// VisitImport processes an import declaration
func (v *MarkdownVisitor) VisitImport(imp *module.Import) error {
	// Imports are usually not documented individually in Markdown
	return nil
}

// Result returns the final markdown
func (v *MarkdownVisitor) Result() (string, error) {
	return v.buffer.String(), nil
}
