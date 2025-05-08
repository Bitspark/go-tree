package visual

import (
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/cmd/gotree/commands"
	"bitspark.dev/go-tree/pkg/visual/html"
	"github.com/spf13/cobra"
)

// htmlCmd generates HTML documentation
var htmlCmd = &cobra.Command{
	Use:   "html [output-dir]",
	Short: "Generate HTML documentation",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create service
		svc, err := commands.CreateService()
		if err != nil {
			return err
		}

		// Get output directory
		outputDir := "docs"
		if len(args) > 0 {
			outputDir = args[0]
		}
		outputDir = filepath.Clean(outputDir)

		// Get options from flags
		includePrivate, _ := cmd.Flags().GetBool("private")
		includeTests, _ := cmd.Flags().GetBool("tests")
		detailLevel, _ := cmd.Flags().GetInt("detail")
		includeTypes, _ := cmd.Flags().GetBool("types")

		// Create visualization options
		options := &html.VisualizationOptions{
			IncludePrivate:         includePrivate,
			IncludeTests:           includeTests,
			DetailLevel:            detailLevel,
			IncludeTypeAnnotations: includeTypes,
			Title:                  "Go-Tree Documentation",
		}

		// Create visualizer
		visualizer := html.NewHTMLVisualizer()

		// Generate visualization
		if svc.Config.Verbose {
			fmt.Printf("Generating HTML documentation in %s...\n", outputDir)
		}

		// Ensure the output directory exists
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate HTML content
		content, err := visualizer.Visualize(svc.Module, options)
		if err != nil {
			return fmt.Errorf("visualization failed: %w", err)
		}

		// Write to index.html in the output directory
		indexPath := filepath.Join(outputDir, "index.html")
		if err := os.WriteFile(indexPath, content, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		if svc.Config.Verbose {
			fmt.Printf("Documentation generated in %s\n", indexPath)
		} else {
			fmt.Println(indexPath)
		}

		return nil
	},
}

func init() {
	htmlCmd.Flags().Bool("private", false, "Include private (unexported) symbols")
	htmlCmd.Flags().Bool("tests", true, "Include test files")
	htmlCmd.Flags().Int("detail", 3, "Detail level (1-5)")
	htmlCmd.Flags().Bool("types", true, "Include type annotations")
}
