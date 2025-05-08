// Package commands defines the CLI commands for the gotree tool.
package commands

import (
	"github.com/spf13/cobra"
)

// Options holds common command options
type Options struct {
	// Input options
	InputDir string

	// Output options
	OutputFile string
	OutputDir  string

	// Common flags
	Verbose bool
}

// GlobalOptions holds the global options for all commands
var GlobalOptions Options

// NewRootCommand initializes and returns the root command
func NewRootCommand() *cobra.Command {
	// Create a new root command
	cmd := &cobra.Command{
		Use:   "gotree",
		Short: "Go-Tree analyzes, visualizes, and transforms Go modules",
		Long: `Go-Tree is a toolkit for working with Go modules.
It provides capabilities for analyzing code, extracting interfaces,
generating documentation, and executing code.

This tool uses a module-centered architecture where operations
are performed on a Go module as a single entity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand provided, display help
			return cmd.Help()
		},
	}

	// Add persistent flags for common options
	cmd.PersistentFlags().StringVarP(&GlobalOptions.InputDir, "input", "i", ".", "Input directory containing a Go module")
	cmd.PersistentFlags().StringVarP(&GlobalOptions.OutputFile, "output", "o", "", "Output file (defaults to stdout)")
	cmd.PersistentFlags().StringVarP(&GlobalOptions.OutputDir, "out-dir", "d", "", "Output directory where files will be created automatically")
	cmd.PersistentFlags().BoolVarP(&GlobalOptions.Verbose, "verbose", "v", false, "Enable verbose output")

	// Add commands
	cmd.AddCommand(newTransformCmd())
	cmd.AddCommand(newVisualizeCmd())
	cmd.AddCommand(newAnalyzeCmd())
	cmd.AddCommand(newExecuteCmd())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return NewRootCommand().Execute()
}
