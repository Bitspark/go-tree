// Package module defines variable and constant related structures for the module data model.
package module

import (
	"go/token"
)

// Variable represents a Go variable declaration
type Variable struct {
	// Variable identity
	Name    string   // Variable name
	File    *File    // File where this variable is defined
	Package *Package // Package this variable belongs to

	// Variable information
	Type       string // Type of the variable
	Value      string // Initial value expression (if any)
	IsExported bool   // Whether the variable is exported

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source

	// Documentation
	Doc string // Documentation comment
}

// Constant represents a Go constant declaration
type Constant struct {
	// Constant identity
	Name    string   // Constant name
	File    *File    // File where this constant is defined
	Package *Package // Package this constant belongs to

	// Constant information
	Type       string // Type of the constant (may be inferred)
	Value      string // Value of the constant
	IsExported bool   // Whether the constant is exported

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source

	// Documentation
	Doc string // Documentation comment
}

// NewVariable creates a new variable
func NewVariable(name, typeName, value string, isExported bool) *Variable {
	return &Variable{
		Name:       name,
		Type:       typeName,
		Value:      value,
		IsExported: isExported,
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
}

// NewConstant creates a new constant
func NewConstant(name, typeName, value string, isExported bool) *Constant {
	return &Constant{
		Name:       name,
		Type:       typeName,
		Value:      value,
		IsExported: isExported,
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
}

// SetPosition sets the position information for this variable
func (v *Variable) SetPosition(pos, end token.Pos) {
	v.Pos = pos
	v.End = end
}

// GetPosition returns the position of this variable
func (v *Variable) GetPosition() *Position {
	if v.File == nil {
		return nil
	}
	return v.File.GetPositionInfo(v.Pos, v.End)
}

// SetPosition sets the position information for this constant
func (c *Constant) SetPosition(pos, end token.Pos) {
	c.Pos = pos
	c.End = end
}

// GetPosition returns the position of this constant
func (c *Constant) GetPosition() *Position {
	if c.File == nil {
		return nil
	}
	return c.File.GetPositionInfo(c.Pos, c.End)
}
