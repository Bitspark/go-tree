// Package markdown provides functionality for generating Markdown documentation
// from Go-Tree package model data.
package markdown

import (
	"bytes"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/model"
)

// MarkdownVisitor implements formatter.Visitor for Markdown output
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

// VisitPackage starts the Markdown document with the package info
func (v *MarkdownVisitor) VisitPackage(pkg *model.GoPackage) error {
	v.packageName = pkg.Name

	// Add package title
	v.buffer.WriteString("# Package " + pkg.Name + "\n\n")

	// Add package documentation if available
	if pkg.PackageDoc != "" {
		v.buffer.WriteString(pkg.PackageDoc + "\n\n")
	}

	return nil
}

// VisitType processes a type declaration
func (v *MarkdownVisitor) VisitType(typ model.GoType) error {
	// Add type header
	v.buffer.WriteString(fmt.Sprintf("## Type: %s (%s)\n\n", typ.Name, typ.Kind))

	// Add type documentation if available
	if typ.Doc != "" {
		v.buffer.WriteString(typ.Doc + "\n\n")
	}

	// Add code block with the type definition
	if v.options.IncludeCodeBlocks && typ.Code != "" {
		v.buffer.WriteString("```go\n")
		v.buffer.WriteString(typ.Code + "\n")
		v.buffer.WriteString("```\n\n")
	}

	// Add struct fields table if applicable
	if typ.Kind == "struct" && len(typ.Fields) > 0 {
		v.buffer.WriteString("### Fields\n\n")
		v.buffer.WriteString("| Name | Type | Tag | Comment |\n")
		v.buffer.WriteString("|------|------|-----|--------|\n")

		for _, field := range typ.Fields {
			name := field.Name
			if name == "" {
				name = "*embedded*"
			}

			v.buffer.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s |\n",
				name, field.Type, field.Tag, field.Comment))
		}
		v.buffer.WriteString("\n")
	}

	// Add interface methods table if applicable
	if typ.Kind == "interface" && len(typ.InterfaceMethods) > 0 {
		v.buffer.WriteString("### Methods\n\n")
		v.buffer.WriteString("| Name | Signature | Comment |\n")
		v.buffer.WriteString("|------|-----------|--------|\n")

		for _, method := range typ.InterfaceMethods {
			sig := method.Signature
			if sig == "" {
				sig = "*embedded interface*"
			}

			v.buffer.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				method.Name, sig, method.Comment))
		}
		v.buffer.WriteString("\n")
	}

	return nil
}

// VisitFunction processes a function declaration
func (v *MarkdownVisitor) VisitFunction(fn model.GoFunction) error {
	// Format function header
	var header string
	if fn.Receiver != nil {
		receiverStr := fn.Receiver.Type
		if fn.Receiver.Name != "" {
			receiverStr = fn.Receiver.Name + " " + receiverStr
		}
		header = fmt.Sprintf("## Method: (%s) %s", receiverStr, fn.Name)
	} else {
		header = "## Function: " + fn.Name
	}

	v.buffer.WriteString(header + "\n\n")

	// Add function documentation if available
	if fn.Doc != "" {
		v.buffer.WriteString(fn.Doc + "\n\n")
	}

	// Add signature
	if fn.Signature != "" {
		v.buffer.WriteString(fmt.Sprintf("**Signature:** `%s`\n\n", fn.Signature))
	}

	// Add code block with the function definition
	if v.options.IncludeCodeBlocks && fn.Code != "" {
		v.buffer.WriteString("```go\n")
		v.buffer.WriteString(fn.Code + "\n")
		v.buffer.WriteString("```\n\n")
	}

	return nil
}

// VisitConstant processes a constant declaration
func (v *MarkdownVisitor) VisitConstant(c model.GoConstant) error {
	// We'll collect constants and output them as a group later
	return nil
}

// VisitVariable processes a variable declaration
func (v *MarkdownVisitor) VisitVariable(vr model.GoVariable) error {
	// We'll collect variables and output them as a group later
	return nil
}

// VisitImport processes an import declaration
func (v *MarkdownVisitor) VisitImport(imp model.GoImport) error {
	// We'll collect imports and output them as a group later
	return nil
}

// Result returns the final markdown
func (v *MarkdownVisitor) Result() (string, error) {
	return v.buffer.String(), nil
}
