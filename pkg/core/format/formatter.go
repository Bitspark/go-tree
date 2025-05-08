// Package format provides functionality for formatting Go packages.
// This package is part of the public API for go-tree.
package format

import (
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/core/model"
)

// FormatPackage formats a GoPackage model into a single Go source file.
func FormatPackage(pkg *model.GoPackage) (string, error) {
	var out strings.Builder

	// Write license header comments if present
	if pkg.LicenseHeader != "" {
		out.WriteString(pkg.LicenseHeader)
		if !strings.HasSuffix(pkg.LicenseHeader, "\n\n") {
			// ensure a blank line after license block
			out.WriteString("\n")
		}
	}

	// Write package documentation comment if present
	if pkg.PackageDoc != "" {
		for _, line := range strings.Split(pkg.PackageDoc, "\n") {
			// Reconstruct as line comments (assuming PackageDoc was stripped of comment markers by parser)
			if line == "" {
				out.WriteString("//\n")
			} else {
				out.WriteString("// " + line + "\n")
			}
		}
	}

	// Package declaration
	out.WriteString("package " + pkg.Name + "\n\n")

	// Imports (combined unique imports)
	if len(pkg.Imports) > 0 {
		if len(pkg.Imports) == 1 {
			imp := pkg.Imports[0]
			out.WriteString("import ")
			if imp.Alias != "" {
				out.WriteString(imp.Alias + " ")
			}
			out.WriteString(fmt.Sprintf("%q", imp.Path))
			if imp.Comment != "" {
				out.WriteString(" // " + imp.Comment)
			}
			out.WriteString("\n\n")
		} else {
			out.WriteString("import (\n")
			for _, imp := range pkg.Imports {
				// Write doc comment for import if present
				if imp.Doc != "" {
					for _, line := range strings.Split(imp.Doc, "\n") {
						out.WriteString("\t// " + line + "\n")
					}
				}

				out.WriteString("\t")
				if imp.Alias != "" {
					out.WriteString(imp.Alias + " ")
				}
				out.WriteString(fmt.Sprintf("%q", imp.Path))
				if imp.Comment != "" {
					out.WriteString(" // " + imp.Comment)
				}
				out.WriteString("\n")
			}
			out.WriteString(")\n\n")
		}
	}

	// Constants
	if len(pkg.Constants) > 0 {
		formatConstants(&out, pkg.Constants)
		out.WriteString("\n")
	}

	// Variables
	if len(pkg.Variables) > 0 {
		formatVariables(&out, pkg.Variables)
		out.WriteString("\n")
	}

	// Types
	for _, t := range pkg.Types {
		formatType(&out, t)
		out.WriteString("\n")
	}

	// Functions
	for _, f := range pkg.Functions {
		formatFunction(&out, f)
		out.WriteString("\n")
	}

	return out.String(), nil
}

// formatConstants formats constant declarations
func formatConstants(out *strings.Builder, constants []model.GoConstant) {
	if len(constants) == 0 {
		return
	}

	// Check if we can group constants with the same type (for iota sequences)
	groups := groupConstants(constants)

	for _, group := range groups {
		if len(group) == 1 {
			// Single constant
			c := group[0]
			if c.Doc != "" {
				writeDocComment(out, c.Doc)
			}
			out.WriteString("const ")
			out.WriteString(c.Name)
			if c.Type != "" {
				out.WriteString(" " + c.Type)
			}
			if c.Value != "" {
				out.WriteString(" = " + c.Value)
			}
			if c.Comment != "" {
				out.WriteString(" // " + c.Comment)
			}
			out.WriteString("\n")
		} else {
			// Group of constants
			first := group[0]
			if first.Doc != "" {
				writeDocComment(out, first.Doc)
			}
			out.WriteString("const (\n")
			for _, c := range group {
				out.WriteString("\t" + c.Name)
				if c.Type != "" {
					out.WriteString(" " + c.Type)
				}
				if c.Value != "" {
					out.WriteString(" = " + c.Value)
				}
				if c.Comment != "" {
					out.WriteString(" // " + c.Comment)
				}
				out.WriteString("\n")
			}
			out.WriteString(")\n")
		}
	}
}

// groupConstants groups constants that should be in the same block (based on docs and sequential declaration)
func groupConstants(constants []model.GoConstant) [][]model.GoConstant {
	if len(constants) == 0 {
		return nil
	}

	var result [][]model.GoConstant
	var currentGroup []model.GoConstant

	for i, c := range constants {
		// Start a new group on the first constant or if this constant has a doc comment (suggests separation)
		if i == 0 || (i > 0 && c.Doc != "") {
			if len(currentGroup) > 0 {
				result = append(result, currentGroup)
			}
			currentGroup = []model.GoConstant{c}
		} else {
			// Add to current group
			currentGroup = append(currentGroup, c)
		}
	}

	// Add the last group if not empty
	if len(currentGroup) > 0 {
		result = append(result, currentGroup)
	}

	return result
}

// formatVariables formats variable declarations
func formatVariables(out *strings.Builder, variables []model.GoVariable) {
	if len(variables) == 0 {
		return
	}

	// Group variables similarly to constants
	groups := groupVariables(variables)

	for _, group := range groups {
		if len(group) == 1 {
			// Single variable
			v := group[0]
			if v.Doc != "" {
				writeDocComment(out, v.Doc)
			}
			out.WriteString("var ")
			out.WriteString(v.Name)
			if v.Type != "" {
				out.WriteString(" " + v.Type)
			}
			if v.Value != "" {
				out.WriteString(" = " + v.Value)
			}
			if v.Comment != "" {
				out.WriteString(" // " + v.Comment)
			}
			out.WriteString("\n")
		} else {
			// Group of variables
			first := group[0]
			if first.Doc != "" {
				writeDocComment(out, first.Doc)
			}
			out.WriteString("var (\n")
			for _, v := range group {
				out.WriteString("\t" + v.Name)
				if v.Type != "" {
					out.WriteString(" " + v.Type)
				}
				if v.Value != "" {
					out.WriteString(" = " + v.Value)
				}
				if v.Comment != "" {
					out.WriteString(" // " + v.Comment)
				}
				out.WriteString("\n")
			}
			out.WriteString(")\n")
		}
	}
}

// groupVariables groups variables that should be in the same block
func groupVariables(variables []model.GoVariable) [][]model.GoVariable {
	if len(variables) == 0 {
		return nil
	}

	var result [][]model.GoVariable
	var currentGroup []model.GoVariable

	for i, v := range variables {
		// Start a new group on the first variable or if this variable has a doc comment
		if i == 0 || (i > 0 && v.Doc != "") {
			if len(currentGroup) > 0 {
				result = append(result, currentGroup)
			}
			currentGroup = []model.GoVariable{v}
		} else {
			// Add to current group
			currentGroup = append(currentGroup, v)
		}
	}

	// Add the last group if not empty
	if len(currentGroup) > 0 {
		result = append(result, currentGroup)
	}

	return result
}

// formatType formats a type declaration
func formatType(out *strings.Builder, t model.GoType) {
	if t.Doc != "" {
		writeDocComment(out, t.Doc)
	}

	// Use the saved original code if available
	if t.Code != "" {
		out.WriteString(t.Code)
		return
	}

	// Otherwise reconstruct the type declaration
	out.WriteString("type " + t.Name + " ")

	switch t.Kind {
	case "alias":
		out.WriteString("= " + t.AliasOf)
	case "struct":
		out.WriteString("struct {\n")
		for _, field := range t.Fields {
			if field.Doc != "" {
				writeIndentedDocComment(out, field.Doc, "\t")
			}
			out.WriteString("\t")
			if field.Name == "" {
				// Embedded field
				out.WriteString(field.Type)
			} else {
				out.WriteString(field.Name + " " + field.Type)
			}
			if field.Tag != "" {
				out.WriteString(" " + field.Tag)
			}
			if field.Comment != "" {
				out.WriteString(" // " + field.Comment)
			}
			out.WriteString("\n")
		}
		out.WriteString("}")
	case "interface":
		out.WriteString("interface {\n")
		for _, method := range t.InterfaceMethods {
			if method.Doc != "" {
				writeIndentedDocComment(out, method.Doc, "\t")
			}
			out.WriteString("\t")
			if method.Signature == "" {
				// Embedded interface
				out.WriteString(method.Name)
			} else {
				out.WriteString(method.Name + method.Signature)
			}
			if method.Comment != "" {
				out.WriteString(" // " + method.Comment)
			}
			out.WriteString("\n")
		}
		out.WriteString("}")
	default:
		// Regular type
		out.WriteString(t.UnderlyingType)
	}

	out.WriteString("\n")
}

// formatFunction formats a function or method declaration
func formatFunction(out *strings.Builder, f model.GoFunction) {
	// Use the saved original code if available
	if f.Code != "" {
		out.WriteString(f.Code)
		return
	}

	// Otherwise reconstruct the function
	if f.Doc != "" {
		writeDocComment(out, f.Doc)
	}

	out.WriteString("func ")

	// Add receiver for methods
	if f.Receiver != nil {
		out.WriteString("(")
		if f.Receiver.Name != "" {
			out.WriteString(f.Receiver.Name + " ")
		}
		out.WriteString(f.Receiver.Type)
		out.WriteString(") ")
	}

	out.WriteString(f.Name)
	out.WriteString(f.Signature)
	out.WriteString(" {\n")
	out.WriteString(f.Body)
	out.WriteString("}\n")
}

// writeDocComment writes a documentation comment
func writeDocComment(out *strings.Builder, doc string) {
	for _, line := range strings.Split(doc, "\n") {
		out.WriteString("// " + line + "\n")
	}
}

// writeIndentedDocComment writes an indented documentation comment
func writeIndentedDocComment(out *strings.Builder, doc string, indent string) {
	for _, line := range strings.Split(doc, "\n") {
		out.WriteString(indent + "// " + line + "\n")
	}
}
