// Package commands implements the CLI commands for go-tree
package commands

import (
	"bitspark.dev/go-tree/pkg/service"
	"github.com/spf13/cobra"
)

var config = &service.Config{
	ModuleDir:    ".",
	IncludeTests: true,
	WithDeps:     false,
	Verbose:      false,
}

var rootCmd = &cobra.Command{
	Use:   "gotree",
	Short: "Go-Tree is a tool for analyzing and manipulating Go code",
	Long: `Go-Tree provides a comprehensive set of tools for working with Go code.
It leverages Go's type system to provide accurate code analysis, visualization,
and transformation.`,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&config.ModuleDir, "dir", "d", ".", "Directory of the Go module")
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&config.IncludeTests, "with-tests", true, "Include test files")
	rootCmd.PersistentFlags().BoolVar(&config.WithDeps, "with-deps", false, "Include dependencies")
}

// CreateService creates a service instance from configuration
func CreateService() (*service.Service, error) {
	return service.NewService(config)
}

// AddCommand adds a subcommand to the root command
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
