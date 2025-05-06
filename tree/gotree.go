// Package tree provides utilities for parsing and formatting Go packages.
package tree

import (
	"bitspark.dev/go-tree/internal/formatter"
	"bitspark.dev/go-tree/internal/model"
	"bitspark.dev/go-tree/internal/parser"
)

// Package represents a parsed Go package
type Package struct {
	Model *model.GoPackage
}

// Parse parses a Go package from the given directory and returns a Package.
func Parse(dir string) (*Package, error) {
	pkg, err := parser.ParsePackage(dir)
	if err != nil {
		return nil, err
	}
	return &Package{Model: pkg}, nil
}

// Format formats a Package into a single Go source file.
func (p *Package) Format() (string, error) {
	return formatter.FormatPackage(p.Model)
}

// Get package model getters

// Name returns the package name
func (p *Package) Name() string {
	return p.Model.Name
}

// Imports returns the package imports
func (p *Package) Imports() []string {
	imports := make([]string, len(p.Model.Imports))
	for i, imp := range p.Model.Imports {
		imports[i] = imp.Path
	}
	return imports
}

// FunctionNames returns the names of all functions in the package
func (p *Package) FunctionNames() []string {
	names := make([]string, len(p.Model.Functions))
	for i, fn := range p.Model.Functions {
		names[i] = fn.Name
	}
	return names
}

// TypeNames returns the names of all types in the package
func (p *Package) TypeNames() []string {
	names := make([]string, len(p.Model.Types))
	for i, t := range p.Model.Types {
		names[i] = t.Name
	}
	return names
}

// ConstantNames returns the names of all constants in the package
func (p *Package) ConstantNames() []string {
	names := make([]string, len(p.Model.Constants))
	for i, c := range p.Model.Constants {
		names[i] = c.Name
	}
	return names
}

// VariableNames returns the names of all variables in the package
func (p *Package) VariableNames() []string {
	names := make([]string, len(p.Model.Variables))
	for i, v := range p.Model.Variables {
		names[i] = v.Name
	}
	return names
}

// GetFunction returns a function by name, or nil if not found
func (p *Package) GetFunction(name string) *Function {
	for _, fn := range p.Model.Functions {
		if fn.Name == name {
			return &Function{&fn}
		}
	}
	return nil
}

// GetType returns a type by name, or nil if not found
func (p *Package) GetType(name string) *Type {
	for _, t := range p.Model.Types {
		if t.Name == name {
			return &Type{&t}
		}
	}
	return nil
}

// Function represents a function in a parsed Go package
type Function struct {
	model *model.GoFunction
}

// Name returns the function name
func (f *Function) Name() string {
	return f.model.Name
}

// IsMethod returns true if the function is a method (has a receiver)
func (f *Function) IsMethod() bool {
	return f.model.Receiver != nil
}

// ReceiverType returns the receiver type for methods, or empty string for functions
func (f *Function) ReceiverType() string {
	if f.model.Receiver != nil {
		return f.model.Receiver.Type
	}
	return ""
}

// Signature returns the function signature
func (f *Function) Signature() string {
	return f.model.Signature
}

// Code returns the full function code
func (f *Function) Code() string {
	return f.model.Code
}

// Type represents a type in a parsed Go package
type Type struct {
	model *model.GoType
}

// Name returns the type name
func (t *Type) Name() string {
	return t.model.Name
}

// Kind returns the kind of type (struct, interface, alias, etc.)
func (t *Type) Kind() string {
	return t.model.Kind
}

// IsStruct returns true if the type is a struct
func (t *Type) IsStruct() bool {
	return t.model.Kind == "struct"
}

// IsInterface returns true if the type is an interface
func (t *Type) IsInterface() bool {
	return t.model.Kind == "interface"
}

// Code returns the full type definition code
func (t *Type) Code() string {
	return t.model.Code
}
