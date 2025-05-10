// Package moduleb provides functionality for module B
package moduleb

import (
	"fmt"

	modulea "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-a"
)

// Version is the current version of the module
const Version = "1.0.0"

// GetMessage returns a greeting message that includes module A's message
func GetMessage() string {
	return fmt.Sprintf("Message from Module B (using Module A): %s", modulea.GetMessage())
}

// Calculate performs a calculation using module A's sum function
func Calculate(x, y int) int {
	return modulea.Sum(x, y) * 2
}
