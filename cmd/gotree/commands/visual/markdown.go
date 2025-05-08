package visual

import (
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/cmd/gotree/commands"
	"bitspark.dev/go-tree/pkg/visual/markdown"
	"github.com/spf13/cobra"
)

// markdownCmd generates Markdown documentation
var markdownCmd = &cobra.Command{
	Use:   "markdown [output-file]",
	Short: "Generate Markdown documentation",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create service
		svc, err := commands.CreateService()
		if err != nil {
			return err
		}

		// Get output file path
		outputPath := "docs.md"
		if len(args) > 0 {
			outputPath = args[0]
		}
		outputPath = filepath.Clean(outputPath)

		// Get options from flags
		includePrivate, _ := cmd.Flags().GetBool("private")
		includeTests, _ := cmd.Flags().GetBool("tests")
		detailLevel, _ := cmd.Flags().GetInt("detail")
		includeTypes, _ := cmd.Flags().GetBool("types")
		title, _ := cmd.Flags().GetString("title")

		// Create visualization options
		options := &markdown.VisualizationOptions{
			IncludePrivate:         includePrivate,
			IncludeTests:           includeTests,
			DetailLevel:            detailLevel,
			IncludeTypeAnnotations: includeTypes,
			Title:                  title,
		}

		// Create visualizer
		visualizer := markdown.NewMarkdownVisualizer()

		// Generate visualization
		if svc.Config.Verbose {
			fmt.Printf("Generating Markdown documentation to %s...\n", outputPath)
		}

		// Ensure the output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate Markdown content
		content, err := visualizer.Visualize(svc.Module, options)
		if err != nil {
			return fmt.Errorf("visualization failed: %w", err)
		}

		// Write to the output file
		if err := os.WriteFile(outputPath, content, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		if svc.Config.Verbose {
			fmt.Printf("Documentation generated in %s\n", outputPath)
		} else {
			fmt.Println(outputPath)
		}

		return nil
	},
}

func init() {
	markdownCmd.Flags().Bool("private", false, "Include private (unexported) symbols")
	markdownCmd.Flags().Bool("tests", true, "Include test files")
	markdownCmd.Flags().Int("detail", 3, "Detail level (1-5)")
	markdownCmd.Flags().Bool("types", true, "Include type annotations")
	markdownCmd.Flags().String("title", "Go-Tree Documentation", "Title for the documentation")
}
