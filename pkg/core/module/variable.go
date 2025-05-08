// Package module defines variable and constant related structures for the module data model.
package module

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
	}
}

// NewConstant creates a new constant
func NewConstant(name, typeName, value string, isExported bool) *Constant {
	return &Constant{
		Name:       name,
		Type:       typeName,
		Value:      value,
		IsExported: isExported,
	}
}
