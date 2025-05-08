// Package module defines file-related types for the module data model.
package module

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
)

// File represents a Go source file
type File struct {
	// File identity
	Path    string   // Absolute path to file
	Name    string   // File name
	Package *Package // Package this file belongs to

	// File content
	Imports   []*Import   // Imports in this file
	Types     []*Type     // Types defined in this file
	Functions []*Function // Functions defined in this file
	Variables []*Variable // Variables defined in this file
	Constants []*Constant // Constants defined in this file

	// Source information
	SourceCode string         // Original source code (preserved)
	AST        *ast.File      // AST representation (optional, may be nil)
	FileSet    *token.FileSet // FileSet used to parse this file (for position information)
	TokenFile  *token.File    // Token file for precise position mapping

	// Build information
	BuildTags   []string // Build constraints
	IsTest      bool     // Whether this is a test file
	IsGenerated bool     // Whether this file is generated

	// Tracking
	IsModified bool // Whether this file has been modified since loading
}

// Position represents a position in the source code
type Position struct {
	File      *File     // File containing this position
	Pos       token.Pos // Position in the file
	End       token.Pos // End position (for spans)
	LineStart int       // Line number start (1-based)
	ColStart  int       // Column start (1-based)
	LineEnd   int       // Line number end (1-based)
	ColEnd    int       // Column end (1-based)
}

// NewFile creates a new empty file
func NewFile(path, name string, isTest bool) *File {
	return &File{
		Path:       path,
		Name:       name,
		IsTest:     isTest,
		Imports:    make([]*Import, 0),
		Types:      make([]*Type, 0),
		Functions:  make([]*Function, 0),
		Variables:  make([]*Variable, 0),
		Constants:  make([]*Constant, 0),
		BuildTags:  make([]string, 0),
		FileSet:    token.NewFileSet(),
		IsModified: false,
	}
}

// AddImport adds an import to the file
func (f *File) AddImport(i *Import) {
	f.Imports = append(f.Imports, i)
	f.IsModified = true
	i.File = f
}

// AddType adds a type to the file
func (f *File) AddType(t *Type) {
	f.Types = append(f.Types, t)
	f.IsModified = true
	t.File = f
}

// AddFunction adds a function to the file
func (f *File) AddFunction(fn *Function) {
	f.Functions = append(f.Functions, fn)
	f.IsModified = true
	fn.File = f
}

// AddVariable adds a variable to the file
func (f *File) AddVariable(v *Variable) {
	f.Variables = append(f.Variables, v)
	f.IsModified = true
	v.File = f
}

// AddConstant adds a constant to the file
func (f *File) AddConstant(c *Constant) {
	f.Constants = append(f.Constants, c)
	f.IsModified = true
	c.File = f
}

// GetPositionInfo converts a token.Pos to a Position structure with file and line information
func (f *File) GetPositionInfo(pos token.Pos, end token.Pos) *Position {
	if f.FileSet == nil || pos == token.NoPos {
		return nil
	}

	// Get position information
	startPos := f.FileSet.Position(pos)
	endPos := f.FileSet.Position(end)

	return &Position{
		File:      f,
		Pos:       pos,
		End:       end,
		LineStart: startPos.Line,
		ColStart:  startPos.Column,
		LineEnd:   endPos.Line,
		ColEnd:    endPos.Column,
	}
}

// FindElementAtPosition finds the element that contains the specified position
func (f *File) FindElementAtPosition(pos token.Pos) interface{} {
	// Check if the position is within this file
	if f.FileSet == nil || pos == token.NoPos {
		// DEBUG
		fmt.Printf("DEBUG: FindElementAtPosition: FileSet is nil or position is NoPos\n")
		return nil
	}

	// Convert token.Pos to a Position for easier comparison
	posInfo := f.FileSet.Position(pos)
	filePath := posInfo.Filename

	// Check if this position is in this file
	if filepath.Base(filePath) != f.Name {
		// Different file
		fmt.Printf("DEBUG: FindElementAtPosition: Position is in file %s, not %s\n",
			filepath.Base(filePath), f.Name)
		return nil
	}

	// DEBUG: Print all types with positions
	fmt.Printf("DEBUG: FindElementAtPosition: Checking %d types in file %s\n",
		len(f.Types), f.Name)

	for _, t := range f.Types {
		if t.Pos == token.NoPos || t.End == token.NoPos {
			continue
		}

		// Convert type positions to Position for accurate comparison
		typeStartPos := f.FileSet.Position(t.Pos)
		typeEndPos := f.FileSet.Position(t.End)

		fmt.Printf("DEBUG: Type %s: Pos=%v (line %d), End=%v (line %d)\n",
			t.Name, t.Pos, typeStartPos.Line, t.End, typeEndPos.Line)

		// Check if the position is within the type's range
		if typeStartPos.Filename == posInfo.Filename &&
			typeStartPos.Line <= posInfo.Line && posInfo.Line <= typeEndPos.Line {
			fmt.Printf("DEBUG: FindElementAtPosition: Found type %s (line match)\n", t.Name)
			return t
		}
	}

	// Check functions
	fmt.Printf("DEBUG: FindElementAtPosition: Checking %d functions\n", len(f.Functions))
	for _, fn := range f.Functions {
		if fn.Pos == token.NoPos || fn.End == token.NoPos {
			continue
		}

		// Convert function positions to Position for accurate comparison
		fnStartPos := f.FileSet.Position(fn.Pos)
		fnEndPos := f.FileSet.Position(fn.End)

		// Check if the position is within the function's range
		if fnStartPos.Filename == posInfo.Filename &&
			fnStartPos.Line <= posInfo.Line && posInfo.Line <= fnEndPos.Line {
			fmt.Printf("DEBUG: FindElementAtPosition: Found function %s\n", fn.Name)
			return fn
		}
	}

	// Check variables
	fmt.Printf("DEBUG: FindElementAtPosition: Checking %d variables\n", len(f.Variables))
	for _, v := range f.Variables {
		if v.Pos == token.NoPos || v.End == token.NoPos {
			continue
		}

		// Convert variable positions to Position for accurate comparison
		varStartPos := f.FileSet.Position(v.Pos)
		varEndPos := f.FileSet.Position(v.End)

		// Check if the position is within the variable's range
		if varStartPos.Filename == posInfo.Filename &&
			varStartPos.Line <= posInfo.Line && posInfo.Line <= varEndPos.Line {
			fmt.Printf("DEBUG: FindElementAtPosition: Found variable %s\n", v.Name)
			return v
		}
	}

	// Check constants
	fmt.Printf("DEBUG: FindElementAtPosition: Checking %d constants\n", len(f.Constants))
	for _, c := range f.Constants {
		if c.Pos == token.NoPos || c.End == token.NoPos {
			continue
		}

		// Convert constant positions to Position for accurate comparison
		constStartPos := f.FileSet.Position(c.Pos)
		constEndPos := f.FileSet.Position(c.End)

		// Check if the position is within the constant's range
		if constStartPos.Filename == posInfo.Filename &&
			constStartPos.Line <= posInfo.Line && posInfo.Line <= constEndPos.Line {
			fmt.Printf("DEBUG: FindElementAtPosition: Found constant %s\n", c.Name)
			return c
		}
	}

	// Check imports
	fmt.Printf("DEBUG: FindElementAtPosition: Checking %d imports\n", len(f.Imports))
	for _, i := range f.Imports {
		if i.Pos == token.NoPos || i.End == token.NoPos {
			continue
		}

		// Convert import positions to Position for accurate comparison
		importStartPos := f.FileSet.Position(i.Pos)
		importEndPos := f.FileSet.Position(i.End)

		// Check if the position is within the import's range
		if importStartPos.Filename == posInfo.Filename &&
			importStartPos.Line <= posInfo.Line && posInfo.Line <= importEndPos.Line {
			fmt.Printf("DEBUG: FindElementAtPosition: Found import %s\n", i.Path)
			return i
		}
	}

	fmt.Printf("DEBUG: FindElementAtPosition: No element found at position %v (line %d)\n",
		pos, posInfo.Line)
	return nil
}

// PositionString returns a string representation of a position in the format "file:line:col"
func (p *Position) String() string {
	if p == nil || p.File == nil {
		return "<unknown position>"
	}

	if p.LineStart == p.LineEnd && p.ColStart == p.ColEnd {
		return fmt.Sprintf("%s:%d:%d", p.File.Path, p.LineStart, p.ColStart)
	}

	return fmt.Sprintf("%s:%d:%d-%d:%d", p.File.Path, p.LineStart, p.ColStart, p.LineEnd, p.ColEnd)
}
