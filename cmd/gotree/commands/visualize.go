package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkg/core/loader"
	"bitspark.dev/go-tree/pkg/visual/html"
)

type visualizeOptions struct {
	// Common visualization options
	IncludePrivate   bool
	IncludeTests     bool
	IncludeGenerated bool
	Title            string

	// HTML-specific options
	SyntaxHighlight bool
	CustomCSS       string
}

var visualizeOpts visualizeOptions

// newVisualizeCmd creates the visualize command
func newVisualizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visualize",
		Short: "Visualize Go module",
		Long:  `Generates visual representations of a Go module.`,
	}

	// Add subcommands
	cmd.AddCommand(newHtmlCmd())

	return cmd
}

// newHtmlCmd creates the HTML visualization command
func newHtmlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "html",
		Short: "Generate HTML documentation",
		Long:  `Generates HTML documentation for a Go module.`,
		RunE:  runHtmlCmd,
	}

	// Add flags for HTML visualization
	cmd.Flags().BoolVar(&visualizeOpts.IncludePrivate, "include-private", false, "Include private (unexported) elements")
	cmd.Flags().BoolVar(&visualizeOpts.IncludeTests, "include-tests", false, "Include test files")
	cmd.Flags().BoolVar(&visualizeOpts.IncludeGenerated, "include-generated", false, "Include generated files")
	cmd.Flags().StringVar(&visualizeOpts.Title, "title", "", "Custom title for documentation")
	cmd.Flags().BoolVar(&visualizeOpts.SyntaxHighlight, "syntax-highlight", true, "Include CSS for syntax highlighting")
	cmd.Flags().StringVar(&visualizeOpts.CustomCSS, "custom-css", "", "Custom CSS to include in HTML")

	return cmd
}

// runHtmlCmd executes the HTML visualization
func runHtmlCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()
	loadOpts.IncludeTests = visualizeOpts.IncludeTests
	loadOpts.IncludeGenerated = visualizeOpts.IncludeGenerated
	loadOpts.LoadDocs = true

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Configure the HTML visualizer
	htmlOpts := html.DefaultOptions()
	htmlOpts.IncludePrivate = visualizeOpts.IncludePrivate
	htmlOpts.IncludeTests = visualizeOpts.IncludeTests
	htmlOpts.IncludeGenerated = visualizeOpts.IncludeGenerated

	if visualizeOpts.Title != "" {
		htmlOpts.Title = visualizeOpts.Title
	}

	htmlOpts.IncludeCSS = visualizeOpts.SyntaxHighlight
	if visualizeOpts.CustomCSS != "" {
		htmlOpts.CustomCSS = visualizeOpts.CustomCSS
	}

	// Create and run the visualizer
	visualizer := html.NewHTMLVisualizer(htmlOpts)
	fmt.Fprintln(os.Stderr, "Generating HTML documentation...")

	htmlBytes, err := visualizer.Visualize(mod)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Determine output destination
	if GlobalOptions.OutputFile != "" {
		// Write to specified file
		fmt.Fprintf(os.Stderr, "Writing HTML to %s\n", GlobalOptions.OutputFile)
		if err := os.WriteFile(GlobalOptions.OutputFile, htmlBytes, 0600); err != nil {
			return fmt.Errorf("failed to write HTML to file: %w", err)
		}
	} else if GlobalOptions.OutputDir != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(GlobalOptions.OutputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Save to index.html in the output directory
		outputPath := filepath.Join(GlobalOptions.OutputDir, "index.html")
		fmt.Fprintf(os.Stderr, "Writing HTML to %s\n", outputPath)
		if err := os.WriteFile(outputPath, htmlBytes, 0600); err != nil {
			return fmt.Errorf("failed to write HTML to file: %w", err)
		}
	} else {
		// Write to stdout
		if _, err := fmt.Fprintln(os.Stdout, string(htmlBytes)); err != nil {
			return fmt.Errorf("failed to write HTML to stdout: %w", err)
		}
	}

	return nil
}
