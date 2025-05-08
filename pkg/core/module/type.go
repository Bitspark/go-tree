// Package module defines type-related structures for the module data model.
package module

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
}

// Method represents a method in an interface or a struct type
type Method struct {
	Name       string // Method name
	Signature  string // Method signature
	IsEmbedded bool   // Whether this is an embedded interface
	Doc        string // Documentation comment
	Parent     *Type  // Parent type
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
	}
	t.Interfaces = append(t.Interfaces, method)
	return method
}
