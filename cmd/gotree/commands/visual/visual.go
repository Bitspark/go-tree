// Package visual implements the visualization commands
package visual

import (
	"bitspark.dev/go-tree/cmd/gotree/commands"
	"github.com/spf13/cobra"
)

// VisualCmd is the root command for visualization
var VisualCmd = &cobra.Command{
	Use:   "visual",
	Short: "Generate visualizations of Go code",
	Long:  `Generate visualizations of Go code structure with type information.`,
}

// init registers the visual command and its subcommands
// This must be at the bottom of the file to ensure subcommands are defined
func init() {
	// Add subcommands
	VisualCmd.AddCommand(htmlCmd)
	VisualCmd.AddCommand(markdownCmd)

	// Register with root
	commands.AddCommand(VisualCmd)
}
