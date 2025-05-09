package typesys

import (
	"go/ast"
	"go/token"
	"path/filepath"
)

// File represents a Go source file with type information.
type File struct {
	// Basic information
	Path    string   // Absolute file path
	Name    string   // File name (without directory)
	Package *Package // Parent package
	IsTest  bool     // Whether this is a test file

	// AST information
	AST     *ast.File      // Go AST
	FileSet *token.FileSet // FileSet for position information

	// Symbols in this file
	Symbols []*Symbol // All symbols defined in this file
	Imports []*Import // All imports in this file
}

// NewFile creates a new file with the given path.
func NewFile(path string, pkg *Package) *File {
	return &File{
		Path:    path,
		Name:    filepath.Base(path),
		Package: pkg,
		IsTest:  isTestFile(path),
		Symbols: make([]*Symbol, 0),
		Imports: make([]*Import, 0),
	}
}

// AddSymbol adds a symbol to the file.
func (f *File) AddSymbol(sym *Symbol) {
	f.Symbols = append(f.Symbols, sym)
	sym.File = f

	// Also add to package
	if f.Package != nil {
		f.Package.AddSymbol(sym)
	}
}

// RemoveSymbol removes a symbol from the file.
func (f *File) RemoveSymbol(sym *Symbol) {
	// Find and remove the symbol from the Symbols slice
	for i, s := range f.Symbols {
		if s == sym {
			// Remove by swapping with the last element and truncating
			f.Symbols[i] = f.Symbols[len(f.Symbols)-1]
			f.Symbols = f.Symbols[:len(f.Symbols)-1]
			break
		}
	}
}

// AddImport adds an import to the file.
func (f *File) AddImport(imp *Import) {
	f.Imports = append(f.Imports, imp)
	imp.File = f

	// Also add to package
	if f.Package != nil {
		f.Package.Imports[imp.Path] = imp
	}
}

// GetPositionInfo returns line and column information for a token position range.
func (f *File) GetPositionInfo(start, end token.Pos) *PositionInfo {
	if f.FileSet == nil {
		return nil
	}

	// Validate positions first
	if !start.IsValid() || !end.IsValid() {
		return nil
	}

	// Make sure start is before end
	if start > end {
		start, end = end, start
	}

	startPos := f.FileSet.Position(start)
	endPos := f.FileSet.Position(end)

	// Ensure positions are valid and in the correct file
	if !startPos.IsValid() || !endPos.IsValid() {
		return nil
	}

	// We no longer need to log warnings here since we fix the mismatches in createSymbol function
	// Setting the correct file there is better than just warning here

	// Calculate length safely
	length := 0
	if endPos.Offset >= startPos.Offset {
		length = endPos.Offset - startPos.Offset
	}

	return &PositionInfo{
		LineStart:   startPos.Line,
		LineEnd:     endPos.Line,
		ColumnStart: startPos.Column,
		ColumnEnd:   endPos.Column,
		Offset:      startPos.Offset,
		Length:      length,
		Filename:    startPos.Filename,
	}
}

// PositionInfo contains line and column information for a symbol or reference.
type PositionInfo struct {
	LineStart   int // Starting line (1-based)
	LineEnd     int // Ending line (1-based)
	ColumnStart int // Starting column (1-based)
	ColumnEnd   int // Ending column (1-based)
	Offset      int // Byte offset in file
	Length      int // Length in bytes
	Filename    string
}

// Helper function to check if a file is a test file
func isTestFile(path string) bool {
	name := filepath.Base(path)
	return len(name) > 8 && (name[len(name)-8:] == "_test.go")
}
