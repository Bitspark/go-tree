// Package modulec provides functionality for module C
package modulec

import (
	"fmt"

	moduleb "bitspark.dev/go-tree/pkg/io/resolve/testdata/module-b"
)

// Version is the current version of the module
const Version = "1.0.0"

// GetMessage returns a greeting message that includes module B's message
func GetMessage() string {
	return fmt.Sprintf("Message from Module C (using Module B): %s", moduleb.GetMessage())
}

// Calculate performs a calculation using module B's calculate function
func Calculate(x, y int) int {
	return moduleb.Calculate(x, y) + 10
}
