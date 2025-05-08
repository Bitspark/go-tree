// Package module defines package-related types for the module data model.
package module

import (
	"go/token"
)

// Package represents a Go package within a module
type Package struct {
	// Package identity
	Name       string  // Package name (final component of import path)
	ImportPath string  // Full import path
	Dir        string  // Directory containing the package
	Module     *Module // Reference to parent module
	IsTest     bool    // Whether this is a test package

	// Package content
	Files         map[string]*File     // Map of filenames to files
	Types         map[string]*Type     // Types defined in this package
	Functions     map[string]*Function // Functions defined in this package
	Variables     map[string]*Variable // Variables defined in this package
	Constants     map[string]*Constant // Constants defined in this package
	Imports       map[string]*Import   // Packages imported by this package
	Documentation string               // Package documentation

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source

	// Tracking
	IsModified bool // Whether this package has been modified since loading
}

// Import represents a package import
type Import struct {
	Path    string // Import path
	Name    string // Local name (if renamed, otherwise "")
	IsBlank bool   // Whether it's a blank import (_)
	Doc     string // Documentation comment
	File    *File  // File that contains this import

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source
}

// NewPackage creates a new empty package
func NewPackage(name, importPath, dir string) *Package {
	return &Package{
		Name:       name,
		ImportPath: importPath,
		Dir:        dir,
		Files:      make(map[string]*File),
		Types:      make(map[string]*Type),
		Functions:  make(map[string]*Function),
		Variables:  make(map[string]*Variable),
		Constants:  make(map[string]*Constant),
		Imports:    make(map[string]*Import),
		Pos:        token.NoPos,
		End:        token.NoPos,
		IsModified: false,
	}
}

// AddFile adds a file to the package
func (p *Package) AddFile(file *File) {
	p.Files[file.Name] = file
	file.Package = p
	p.IsModified = true
}

// AddType adds a type to the package
func (p *Package) AddType(typ *Type) {
	p.Types[typ.Name] = typ
	typ.Package = p
	p.IsModified = true
}

// AddFunction adds a function to the package
func (p *Package) AddFunction(fn *Function) {
	p.Functions[fn.Name] = fn
	fn.Package = p
	p.IsModified = true
}

// AddVariable adds a variable to the package
func (p *Package) AddVariable(v *Variable) {
	p.Variables[v.Name] = v
	v.Package = p
	p.IsModified = true
}

// AddConstant adds a constant to the package
func (p *Package) AddConstant(c *Constant) {
	p.Constants[c.Name] = c
	c.Package = p
	p.IsModified = true
}

// AddImport adds an import to the package
func (p *Package) AddImport(i *Import) {
	p.Imports[i.Path] = i
	p.IsModified = true
}

// GetFunction gets a function by name
func (p *Package) GetFunction(name string) *Function {
	return p.Functions[name]
}

// GetType gets a type by name
func (p *Package) GetType(name string) *Type {
	return p.Types[name]
}

// GetVariable gets a variable by name
func (p *Package) GetVariable(name string) *Variable {
	return p.Variables[name]
}

// GetConstant gets a constant by name
func (p *Package) GetConstant(name string) *Constant {
	return p.Constants[name]
}

// SetPosition sets the position information for this package
func (p *Package) SetPosition(pos, end token.Pos) {
	p.Pos = pos
	p.End = end
}

// NewImport creates a new import
func NewImport(path, name string, isBlank bool) *Import {
	return &Import{
		Path:    path,
		Name:    name,
		IsBlank: isBlank,
		Pos:     token.NoPos,
		End:     token.NoPos,
	}
}

// SetPosition sets the position information for this import
func (i *Import) SetPosition(pos, end token.Pos) {
	i.Pos = pos
	i.End = end
}

// GetPosition returns the position of this import
func (i *Import) GetPosition() *Position {
	if i.File == nil {
		return nil
	}
	return i.File.GetPositionInfo(i.Pos, i.End)
}
