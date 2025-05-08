// Package module defines function-related structures for the module data model.
package module

import (
	"go/ast"
)

// Function represents a Go function or method
type Function struct {
	// Function identity
	Name    string   // Function name
	File    *File    // File where this function is defined
	Package *Package // Package this function belongs to

	// Function information
	Signature  string       // Function signature
	Receiver   *Receiver    // Receiver if this is a method (nil for functions)
	Parameters []*Parameter // Function parameters
	Results    []*Parameter // Function results
	IsExported bool         // Whether the function is exported
	IsMethod   bool         // Whether this is a method
	IsTest     bool         // Whether this is a test function

	// Function body
	Body string        // Function body as source code
	AST  *ast.FuncDecl // AST node (optional, may be nil)

	// Documentation
	Doc string // Documentation comment
}

// Receiver represents a method receiver
type Receiver struct {
	Name      string // Receiver name (may be empty)
	Type      string // Receiver type (e.g. "*T" or "T")
	IsPointer bool   // Whether the receiver is a pointer
}

// Parameter represents a function parameter or result
type Parameter struct {
	Name       string // Parameter name (may be empty for unnamed results)
	Type       string // Parameter type
	IsVariadic bool   // Whether this is a variadic parameter
}

// NewFunction creates a new function
func NewFunction(name string, isExported bool, isTest bool) *Function {
	return &Function{
		Name:       name,
		IsExported: isExported,
		IsTest:     isTest,
		Parameters: make([]*Parameter, 0),
		Results:    make([]*Parameter, 0),
	}
}

// SetReceiver sets the receiver for a method
func (f *Function) SetReceiver(name, typeName string, isPointer bool) {
	f.Receiver = &Receiver{
		Name:      name,
		Type:      typeName,
		IsPointer: isPointer,
	}
	f.IsMethod = true
}

// AddParameter adds a parameter to the function
func (f *Function) AddParameter(name, typeName string, isVariadic bool) *Parameter {
	param := &Parameter{
		Name:       name,
		Type:       typeName,
		IsVariadic: isVariadic,
	}
	f.Parameters = append(f.Parameters, param)
	return param
}

// AddResult adds a result to the function
func (f *Function) AddResult(name, typeName string) *Parameter {
	result := &Parameter{
		Name: name,
		Type: typeName,
	}
	f.Results = append(f.Results, result)
	return result
}
