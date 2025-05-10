package saver

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// AST-based code generation utilities

// ASTGenerator generates code from AST
type ASTGenerator struct {
	// Configuration options
	options SaveOptions
}

// NewASTGenerator creates a new AST-based code generator
func NewASTGenerator(options SaveOptions) *ASTGenerator {
	return &ASTGenerator{
		options: options,
	}
}

// GenerateFromAST generates Go source code from an AST
func (g *ASTGenerator) GenerateFromAST(file *ast.File, fset *token.FileSet) ([]byte, error) {
	if file == nil || fset == nil {
		return nil, fmt.Errorf("AST file or FileSet is nil")
	}

	var buf bytes.Buffer

	// Choose the appropriate printer configuration based on options
	var config printer.Config
	mode := printer.TabIndent
	if !g.options.UseTabs {
		mode = 0
	}

	tabWidth := g.options.TabWidth
	if tabWidth <= 0 {
		tabWidth = 8
	}

	// For standard Go formatting
	if g.options.Gofmt {
		// Use the format package for standard Go formatting
		if err := format.Node(&buf, fset, file); err != nil {
			return nil, fmt.Errorf("failed to format AST: %w", err)
		}
		return buf.Bytes(), nil
	}

	// For custom formatting
	config.Mode = mode
	config.Tabwidth = tabWidth

	err := config.Fprint(&buf, fset, file)
	if err != nil {
		return nil, fmt.Errorf("failed to print AST: %w", err)
	}

	return buf.Bytes(), nil
}

// ModifyAST modifies an AST based on type-aware symbols
func (g *ASTGenerator) ModifyAST(file *ast.File, symbols []*typesys.Symbol) error {
	// Check if the file is nil
	if file == nil {
		return fmt.Errorf("AST file is nil")
	}

	// Implement AST modification strategies here
	// This would involve:
	// 1. Mapping symbols to AST nodes
	// 2. Updating AST nodes based on symbol changes
	// 3. Adding new declarations for new symbols
	// 4. Removing nodes for deleted symbols

	// For now, just return nil as a placeholder
	return nil
}

// GenerateSourceFile generates a complete source file using AST-based approach
func GenerateSourceFile(file *typesys.File, options SaveOptions) ([]byte, error) {
	// Check if we have AST available
	if file.AST == nil || file.FileSet == nil {
		return nil, fmt.Errorf("file doesn't have AST information")
	}

	// Create AST generator
	generator := NewASTGenerator(options)

	// Modify AST if needed based on the reconstruction mode
	if options.ASTMode != PreserveOriginal {
		if err := generator.ModifyAST(file.AST, file.Symbols); err != nil {
			return nil, fmt.Errorf("failed to modify AST: %w", err)
		}
	}

	// Generate code from AST
	return generator.GenerateFromAST(file.AST, file.FileSet)
}
