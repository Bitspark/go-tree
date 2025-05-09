// Package simplemath provides simple math operations for testing
package simplemath

// Add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// Subtract returns the difference of two integers
func Subtract(a, b int) int {
	return a - b
}

// Multiply returns the product of two integers
func Multiply(a, b int) int {
	return a * b
}

// Divide returns the quotient of two integers
// Returns 0 if b is 0
func Divide(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// GetPerson returns a person struct for testing complex return types
func GetPerson(name string) Person {
	return Person{
		Name: name,
		Age:  30,
	}
}

// Person is a simple struct for testing complex return types
type Person struct {
	Name string
	Age  int
}
