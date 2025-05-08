// Package module defines type-related structures for the module data model.
package module

import (
	"go/token"
)

// Type represents a Go type definition
type Type struct {
	// Type identity
	Name    string   // Type name
	File    *File    // File where this type is defined
	Package *Package // Package this type belongs to

	// Type information
	Kind       string // "struct", "interface", "alias", etc.
	Underlying string // Underlying type for type aliases
	IsExported bool   // Whether the type is exported

	// Type details (dependent on Kind)
	Fields     []*Field  // Fields for structs
	Methods    []*Method // Methods for this type
	Interfaces []*Method // Methods for interfaces

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source

	// Documentation
	Doc string // Documentation comment
}

// Field represents a field in a struct type
type Field struct {
	Name       string // Field name (empty for embedded fields)
	Type       string // Field type
	Tag        string // Struct tag string, if any
	IsEmbedded bool   // Whether this is an embedded field
	Doc        string // Documentation comment
	Parent     *Type  // Parent type

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source
}

// Method represents a method in an interface or a struct type
type Method struct {
	Name       string // Method name
	Signature  string // Method signature
	IsEmbedded bool   // Whether this is an embedded interface
	Doc        string // Documentation comment
	Parent     *Type  // Parent type

	// Position information
	Pos token.Pos // Start position in source
	End token.Pos // End position in source
}

// NewType creates a new type
func NewType(name, kind string, isExported bool) *Type {
	return &Type{
		Name:       name,
		Kind:       kind,
		IsExported: isExported,
		Fields:     make([]*Field, 0),
		Methods:    make([]*Method, 0),
		Interfaces: make([]*Method, 0),
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
}

// AddField adds a field to a struct type
func (t *Type) AddField(name, fieldType, tag string, isEmbedded bool, doc string) *Field {
	field := &Field{
		Name:       name,
		Type:       fieldType,
		Tag:        tag,
		IsEmbedded: isEmbedded,
		Doc:        doc,
		Parent:     t,
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
	t.Fields = append(t.Fields, field)
	return field
}

// AddMethod adds a method to a type
func (t *Type) AddMethod(name, signature string, isEmbedded bool, doc string) *Method {
	method := &Method{
		Name:       name,
		Signature:  signature,
		IsEmbedded: isEmbedded,
		Doc:        doc,
		Parent:     t,
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
	t.Methods = append(t.Methods, method)
	return method
}

// AddInterfaceMethod adds a method to an interface type
func (t *Type) AddInterfaceMethod(name, signature string, isEmbedded bool, doc string) *Method {
	method := &Method{
		Name:       name,
		Signature:  signature,
		IsEmbedded: isEmbedded,
		Doc:        doc,
		Parent:     t,
		Pos:        token.NoPos,
		End:        token.NoPos,
	}
	t.Interfaces = append(t.Interfaces, method)
	return method
}

// SetPosition sets the position information for this type
func (t *Type) SetPosition(pos, end token.Pos) {
	t.Pos = pos
	t.End = end
}

// GetPosition returns the position of this type
func (t *Type) GetPosition() *Position {
	if t.File == nil {
		return nil
	}
	return t.File.GetPositionInfo(t.Pos, t.End)
}

// SetFieldPosition sets the position information for a field
func (f *Field) SetPosition(pos, end token.Pos) {
	f.Pos = pos
	f.End = end
}

// GetPosition returns the position of this field
func (f *Field) GetPosition() *Position {
	if f.Parent == nil || f.Parent.File == nil {
		return nil
	}
	return f.Parent.File.GetPositionInfo(f.Pos, f.End)
}

// SetMethodPosition sets the position information for a method
func (m *Method) SetPosition(pos, end token.Pos) {
	m.Pos = pos
	m.End = end
}

// GetPosition returns the position of this method
func (m *Method) GetPosition() *Position {
	if m.Parent == nil || m.Parent.File == nil {
		return nil
	}
	return m.Parent.File.GetPositionInfo(m.Pos, m.End)
}
