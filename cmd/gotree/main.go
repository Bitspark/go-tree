// Command gotree provides a CLI for working with Go modules using the module-centered architecture
package main

import (
	"fmt"
	"os"

	"bitspark.dev/go-tree/cmd/gotree/commands"
	// Import command packages to register them
	_ "bitspark.dev/go-tree/cmd/gotree/commands/visual"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
