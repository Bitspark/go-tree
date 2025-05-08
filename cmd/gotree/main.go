// Command gotree provides a CLI for working with Go modules using the module-centered architecture
package main

import (
	"fmt"
	"os"

	"bitspark.dev/go-tree/cmd/gotree/commands"
)

func main() {
	// Execute the root command
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
