// Package modulea provides functionality for module A
package modulea

// Version is the current version of the module
const Version = "1.0.0"

// Sum adds two integers and returns the result
func Sum(a, b int) int {
	return a + b
}

// GetMessage returns a greeting message
func GetMessage() string {
	return "Hello from Module A"
}
