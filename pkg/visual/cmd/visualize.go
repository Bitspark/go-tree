// Package cmd provides command-line utilities for the visual package.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/typesys"
	"bitspark.dev/go-tree/pkg/visual/html"
	"bitspark.dev/go-tree/pkg/visual/markdown"
)

// VisualizeOptions contains options for the Visualize command
type VisualizeOptions struct {
	// Directory of the Go module to visualize
	ModuleDir string

	// Output file path (if empty, output to stdout)
	OutputFile string

	// Format to use (html, markdown)
	Format string

	// Whether to include type annotations
	IncludeTypes bool

	// Whether to include private elements
	IncludePrivate bool

	// Whether to include test files
	IncludeTests bool

	// Title for the visualization
	Title string
}

// Visualize generates a visualization of a Go module
func Visualize(opts *VisualizeOptions) error {
	if opts == nil {
		return fmt.Errorf("visualization options cannot be nil")
	}

	// Default to HTML if no format specified
	if opts.Format == "" {
		opts.Format = "html"
	}

	// Load the module with type information
	module, err := typesys.LoadModule(opts.ModuleDir, &typesys.LoadOptions{
		IncludeTests:   opts.IncludeTests,
		IncludePrivate: opts.IncludePrivate,
		Trace:          false,
	})
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Create visualization options based on format
	var output []byte

	switch opts.Format {
	case "html":
		htmlOpts := &html.VisualizationOptions{
			IncludeTypeAnnotations: opts.IncludeTypes,
			IncludePrivate:         opts.IncludePrivate,
			IncludeTests:           opts.IncludeTests,
			DetailLevel:            3, // Medium detail by default
			Title:                  opts.Title,
		}

		visualizer := html.NewHTMLVisualizer()
		output, err = visualizer.Visualize(module, htmlOpts)

	case "markdown", "md":
		mdOpts := &markdown.VisualizationOptions{
			IncludeTypeAnnotations: opts.IncludeTypes,
			IncludePrivate:         opts.IncludePrivate,
			IncludeTests:           opts.IncludeTests,
			DetailLevel:            3, // Medium detail by default
			Title:                  opts.Title,
		}

		visualizer := markdown.NewMarkdownVisualizer()
		output, err = visualizer.Visualize(module, mdOpts)

	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate visualization: %w", err)
	}

	// Output the result
	if opts.OutputFile == "" {
		// Output to stdout
		fmt.Println(string(output))
	} else {
		// Ensure output directory exists
		outputDir := filepath.Dir(opts.OutputFile)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write to the output file
		if err := os.WriteFile(opts.OutputFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("Visualization saved to %s\n", opts.OutputFile)
	}

	return nil
}
