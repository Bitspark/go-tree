// Package saver provides functionality for saving Go modules with type information.
// It serves as the inverse operation to the loader package, enabling serialization
// of in-memory typesys representations back to Go source files.
package saver

import (
	"bytes"
	"fmt"
	"io"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// ModuleSaver defines the interface for saving type-aware modules
type ModuleSaver interface {
	// Save writes a module back to its original location
	Save(module *typesys.Module) error

	// SaveTo writes a module to a new location
	SaveTo(module *typesys.Module, dir string) error

	// SaveWithOptions writes a module with custom options
	SaveWithOptions(module *typesys.Module, options SaveOptions) error

	// SaveToWithOptions writes a module to a new location with custom options
	SaveToWithOptions(module *typesys.Module, dir string, options SaveOptions) error
}

// FileContentGenerator generates Go source code from type-aware representations
type FileContentGenerator interface {
	// GenerateFileContent produces source code for a file
	GenerateFileContent(file *typesys.File, options SaveOptions) ([]byte, error)
}

// SymbolWriter writes Go code for specific symbol types
type SymbolWriter interface {
	// WriteSymbol generates code for a symbol
	WriteSymbol(sym *typesys.Symbol, dst *bytes.Buffer) error
}

// ModificationTracker tracks modifications to typesys elements
type ModificationTracker interface {
	// IsModified checks if an element has been modified
	IsModified(element interface{}) bool

	// MarkModified marks an element as modified
	MarkModified(element interface{})
}

// DefaultFileContentGenerator provides a simple implementation of FileContentGenerator
type DefaultFileContentGenerator struct {
	// Symbol writers for different symbol kinds
	symbolWriters map[typesys.SymbolKind]SymbolWriter
}

// NewDefaultFileContentGenerator creates a new file content generator with default settings
func NewDefaultFileContentGenerator() *DefaultFileContentGenerator {
	gen := &DefaultFileContentGenerator{
		symbolWriters: make(map[typesys.SymbolKind]SymbolWriter),
	}

	gen.RegisterSymbolWriter(typesys.KindFunction, &FunctionWriter{})
	gen.RegisterSymbolWriter(typesys.KindType, &TypeWriter{})
	gen.RegisterSymbolWriter(typesys.KindVariable, &VarWriter{})
	gen.RegisterSymbolWriter(typesys.KindConstant, &ConstWriter{})

	return gen
}

// RegisterSymbolWriter registers a symbol writer for a specific symbol kind
func (g *DefaultFileContentGenerator) RegisterSymbolWriter(kind typesys.SymbolKind, writer SymbolWriter) {
	g.symbolWriters[kind] = writer
}

// GenerateFileContent produces source code for a file
func (g *DefaultFileContentGenerator) GenerateFileContent(file *typesys.File, options SaveOptions) ([]byte, error) {
	if file == nil {
		return nil, fmt.Errorf("cannot generate content for nil file")
	}

	// If we have an AST and want to preserve formatting, use AST-based generation
	if file.AST != nil && options.ASTMode == PreserveOriginal {
		return GenerateSourceFile(file, options)
	}

	// Otherwise, generate from symbols
	return g.generateFromSymbols(file, options)
}

// generateFromSymbols generates file content using symbols
func (g *DefaultFileContentGenerator) generateFromSymbols(file *typesys.File, options SaveOptions) ([]byte, error) {
	var buf bytes.Buffer

	// Write package declaration
	if file.Package != nil {
		buf.WriteString(fmt.Sprintf("package %s\n\n", file.Package.Name))
	} else {
		buf.WriteString("package unknown\n\n")
	}

	// Write imports
	if len(file.Imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range file.Imports {
			if imp.Name != "" {
				buf.WriteString(fmt.Sprintf("\t%s \"%s\"\n", imp.Name, imp.Path))
			} else {
				buf.WriteString(fmt.Sprintf("\t\"%s\"\n", imp.Path))
			}
		}
		buf.WriteString(")\n\n")
	}

	// Group symbols by kind for proper ordering
	// We'll define placeholder constants for each kind
	functionKind := typesys.KindFunction
	typeKind := typesys.KindType
	variableKind := typesys.KindVariable
	constantKind := typesys.KindConstant

	var constants, types, vars, funcs []*typesys.Symbol

	for _, sym := range file.Symbols {
		switch sym.Kind {
		case constantKind:
			constants = append(constants, sym)
		case typeKind:
			types = append(types, sym)
		case variableKind:
			vars = append(vars, sym)
		case functionKind:
			funcs = append(funcs, sym)
		}
	}

	// Write constants
	if len(constants) > 0 {
		for _, sym := range constants {
			if writer, ok := g.symbolWriters[sym.Kind]; ok {
				if err := writer.WriteSymbol(sym, &buf); err != nil {
					return nil, fmt.Errorf("error writing constant %s: %w", sym.Name, err)
				}
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Write types
	if len(types) > 0 {
		for _, sym := range types {
			if writer, ok := g.symbolWriters[sym.Kind]; ok {
				if err := writer.WriteSymbol(sym, &buf); err != nil {
					return nil, fmt.Errorf("error writing type %s: %w", sym.Name, err)
				}
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Write variables
	if len(vars) > 0 {
		for _, sym := range vars {
			if writer, ok := g.symbolWriters[sym.Kind]; ok {
				if err := writer.WriteSymbol(sym, &buf); err != nil {
					return nil, fmt.Errorf("error writing variable %s: %w", sym.Name, err)
				}
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Write functions
	if len(funcs) > 0 {
		for _, sym := range funcs {
			if writer, ok := g.symbolWriters[sym.Kind]; ok {
				if err := writer.WriteSymbol(sym, &buf); err != nil {
					return nil, fmt.Errorf("error writing function %s: %w", sym.Name, err)
				}
			}
			buf.WriteString("\n\n")
		}
	}

	return buf.Bytes(), nil
}

// WriteTo writes the file content to a writer
func WriteTo(content []byte, w io.Writer) error {
	_, err := w.Write(content)
	return err
}
