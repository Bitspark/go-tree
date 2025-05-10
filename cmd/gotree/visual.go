package main

import (
	visualcmd "bitspark.dev/go-tree/pkg/ext/visual/cmd"
	"bitspark.dev/go-tree/pkg/ext/visual/json"
	"bitspark.dev/go-tree/pkg/io/loader"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

func newVisualCmd() *cobra.Command {
	visualCmd := &cobra.Command{
		Use:   "visual",
		Short: "Generate structured representations and visualizations of Go modules",
		Long:  "Create structured representations of Go modules in various formats (HTML, Markdown, JSON, etc.)",
	}

	visualCmd.AddCommand(
		newHTMLCmd(),
		newMarkdownCmd(),
		newJSONCmd(),
		newDiagramCmd(),
	)

	return visualCmd
}

func newHTMLCmd() *cobra.Command {
	var moduleDir string
	var outputFile string
	var includeTypes bool
	var includePrivate bool
	var includeTests bool
	var detailLevel int
	var title string

	cmd := &cobra.Command{
		Use:   "html [flags]",
		Short: "Generate HTML visualization of a Go module",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate module directory
			if moduleDir == "" {
				moduleDir = "."
			}

			// Create visualization options
			opts := &visualcmd.VisualizeOptions{
				ModuleDir:      moduleDir,
				OutputFile:     outputFile,
				Format:         "html",
				IncludeTypes:   includeTypes,
				IncludePrivate: includePrivate,
				IncludeTests:   includeTests,
				Title:          title,
			}

			// Run the visualization
			if err := visualcmd.Visualize(opts); err != nil {
				return fmt.Errorf("failed to generate HTML visualization: %w", err)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&moduleDir, "module", "m", ".", "Directory of the Go module to visualize")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if empty, output to stdout)")
	cmd.Flags().BoolVarP(&includeTypes, "types", "t", true, "Include type annotations")
	cmd.Flags().BoolVarP(&includePrivate, "private", "p", false, "Include private elements")
	cmd.Flags().BoolVarP(&includeTests, "tests", "", false, "Include test files")
	cmd.Flags().IntVarP(&detailLevel, "detail", "d", 3, "Detail level (1=minimal, 5=complete)")
	cmd.Flags().StringVar(&title, "title", "", "Title for the visualization")

	return cmd
}

func newMarkdownCmd() *cobra.Command {
	var moduleDir string
	var outputFile string
	var includeTypes bool
	var includePrivate bool
	var includeTests bool
	var detailLevel int
	var title string

	cmd := &cobra.Command{
		Use:     "markdown [flags]",
		Short:   "Generate Markdown visualization of a Go module",
		Aliases: []string{"md"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate module directory
			if moduleDir == "" {
				moduleDir = "."
			}

			// Create visualization options
			opts := &visualcmd.VisualizeOptions{
				ModuleDir:      moduleDir,
				OutputFile:     outputFile,
				Format:         "markdown",
				IncludeTypes:   includeTypes,
				IncludePrivate: includePrivate,
				IncludeTests:   includeTests,
				Title:          title,
			}

			// Run the visualization
			if err := visualcmd.Visualize(opts); err != nil {
				return fmt.Errorf("failed to generate Markdown visualization: %w", err)
			}

			return nil
		},
	}

	// Add flags (same as HTML command)
	cmd.Flags().StringVarP(&moduleDir, "module", "m", ".", "Directory of the Go module to visualize")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if empty, output to stdout)")
	cmd.Flags().BoolVarP(&includeTypes, "types", "t", true, "Include type annotations")
	cmd.Flags().BoolVarP(&includePrivate, "private", "p", false, "Include private elements")
	cmd.Flags().BoolVarP(&includeTests, "tests", "", false, "Include test files")
	cmd.Flags().IntVarP(&detailLevel, "detail", "d", 3, "Detail level (1=minimal, 5=complete)")
	cmd.Flags().StringVar(&title, "title", "", "Title for the visualization")

	return cmd
}

func newJSONCmd() *cobra.Command {
	var moduleDir string
	var outputFile string
	var includeTypes bool
	var includePrivate bool
	var includeTests bool
	var prettyPrint bool

	cmd := &cobra.Command{
		Use:   "json [flags]",
		Short: "Generate JSON representation of a Go module",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate module directory
			if moduleDir == "" {
				moduleDir = "."
			}

			// Load the module
			module, err := loader.LoadModule(moduleDir, &typesys.LoadOptions{
				IncludeTests:   includeTests,
				IncludePrivate: includePrivate,
			})
			if err != nil {
				return fmt.Errorf("failed to load module: %w", err)
			}

			// Create JSON visualizer
			visualizer := json.NewJSONVisualizer()

			// Create visualization options
			opts := &json.VisualizationOptions{
				IncludeTypeAnnotations: includeTypes,
				IncludePrivate:         includePrivate,
				IncludeTests:           includeTests,
				DetailLevel:            3, // Medium detail by default
				PrettyPrint:            prettyPrint,
			}

			// Generate JSON
			output, err := visualizer.Visualize(module, opts)
			if err != nil {
				return fmt.Errorf("failed to generate JSON visualization: %w", err)
			}

			// Output the result
			if outputFile == "" {
				// Output to stdout
				fmt.Println(string(output))
			} else {
				// Ensure output directory exists
				outputDir := filepath.Dir(outputFile)
				if err := os.MkdirAll(outputDir, 0750); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}

				// Write to the output file
				if err := os.WriteFile(outputFile, output, 0600); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}

				fmt.Printf("JSON visualization saved to %s\n", outputFile)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&moduleDir, "module", "m", ".", "Directory of the Go module to visualize")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if empty, output to stdout)")
	cmd.Flags().BoolVarP(&includeTypes, "types", "t", true, "Include type annotations")
	cmd.Flags().BoolVarP(&includePrivate, "private", "p", false, "Include private elements")
	cmd.Flags().BoolVarP(&includeTests, "tests", "", false, "Include test files")
	cmd.Flags().BoolVarP(&prettyPrint, "pretty", "", true, "Pretty-print the JSON output")

	return cmd
}

func newDiagramCmd() *cobra.Command {
	var moduleDir string
	var outputFile string
	var diagramType string
	var includePrivate bool
	var includeTests bool

	cmd := &cobra.Command{
		Use:   "diagram [flags]",
		Short: "Generate diagrams of a Go module",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate module directory
			if moduleDir == "" {
				moduleDir = "."
			}

			// Validate diagram type
			validTypes := []string{"package", "type", "dependency", "symbols", "imports"}
			valid := false
			for _, t := range validTypes {
				if diagramType == t {
					valid = true
					break
				}
			}

			if !valid {
				return fmt.Errorf("invalid diagram type: %s. Valid types: %v", diagramType, validTypes)
			}

			// Load the module
			module, err := loader.LoadModule(moduleDir, &typesys.LoadOptions{
				IncludeTests:   includeTests,
				IncludePrivate: includePrivate,
			})
			if err != nil {
				return fmt.Errorf("failed to load module: %w", err)
			}

			fmt.Printf("Module: %s (Go %s)\n", module.Path, module.GoVersion)
			fmt.Printf("Packages: %d\n", len(module.Packages))

			// TODO: Implement diagram visualization
			fmt.Printf("Diagram generation not yet implemented for type: %s\n", diagramType)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&moduleDir, "module", "m", ".", "Directory of the Go module to visualize")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (if empty, output to stdout)")
	cmd.Flags().StringVarP(&diagramType, "type", "t", "package", "Type of diagram (package, type, dependency, symbols, imports)")
	cmd.Flags().BoolVarP(&includePrivate, "private", "p", false, "Include private elements")
	cmd.Flags().BoolVarP(&includeTests, "tests", "", false, "Include test files")

	return cmd
}
