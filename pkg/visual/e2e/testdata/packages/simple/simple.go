// Package simple contains a minimal set of Go declarations for testing.
package simple

// SimpleStruct is a basic struct with a few fields.
type SimpleStruct struct {
	Name  string // A string field
	Count int    // An integer field
}

// SimpleInterface defines a simple interface with one method.
type SimpleInterface interface {
	DoSomething() error
}

// SimpleFunction is a basic function with no special features.
func SimpleFunction(input string) string {
	return "Hello, " + input
}

// SimpleMethod is a method defined on SimpleStruct.
func (s *SimpleStruct) SimpleMethod() string {
	return s.Name
}

// SimpleConstant is a basic constant.
const SimpleConstant = "constant value"

// SimpleVariable is a basic variable.
var SimpleVariable = 42
