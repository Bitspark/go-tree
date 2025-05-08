// Package module defines file-related types for the module data model.
package module

import (
	"go/ast"
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
	SourceCode string    // Original source code
	AST        *ast.File // AST representation (optional, may be nil)

	// Build information
	BuildTags   []string // Build constraints
	IsTest      bool     // Whether this is a test file
	IsGenerated bool     // Whether this file is generated
}

// NewFile creates a new empty file
func NewFile(path, name string, isTest bool) *File {
	return &File{
		Path:      path,
		Name:      name,
		IsTest:    isTest,
		Imports:   make([]*Import, 0),
		Types:     make([]*Type, 0),
		Functions: make([]*Function, 0),
		Variables: make([]*Variable, 0),
		Constants: make([]*Constant, 0),
		BuildTags: make([]string, 0),
	}
}

// AddImport adds an import to the file
func (f *File) AddImport(i *Import) {
	f.Imports = append(f.Imports, i)
}

// AddType adds a type to the file
func (f *File) AddType(t *Type) {
	f.Types = append(f.Types, t)
	t.File = f
}

// AddFunction adds a function to the file
func (f *File) AddFunction(fn *Function) {
	f.Functions = append(f.Functions, fn)
	fn.File = f
}

// AddVariable adds a variable to the file
func (f *File) AddVariable(v *Variable) {
	f.Variables = append(f.Variables, v)
	v.File = f
}

// AddConstant adds a constant to the file
func (f *File) AddConstant(c *Constant) {
	f.Constants = append(f.Constants, c)
	c.File = f
}
