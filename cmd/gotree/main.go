package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gotree",
		Short: "Go-Tree CLI tools for Go code analysis",
		Long:  `Go-Tree provides tools for analyzing, visualizing, and understanding Go codebases.`,
	}

	// Add commands
	rootCmd.AddCommand(
		newVisualCmd(),
		// Add other commands here as they are implemented
	)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
