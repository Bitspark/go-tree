package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/core/saver"
	"bitspark.dev/go-tree/pkgold/transform/rename"
)

type renameOptions struct {
	// Common options
	DryRun bool

	// Variable renaming options
	OldName string
	NewName string

	// Type of element to rename
	ElementType string
}

var renameOpts renameOptions

// newRenameCmd creates the rename command
func newRenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename elements in Go code",
		Long:  `Renames variables, constants, functions, and types in Go code.`,
	}

	// Add subcommands
	cmd.AddCommand(newRenameVariableCmd())

	return cmd
}

// newRenameVariableCmd creates the variable rename command
func newRenameVariableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Rename variables in Go code",
		Long:  `Renames a variable and updates all references to it.`,
		RunE:  runRenameVariableCmd,
	}

	// Add flags specific to variable renaming
	cmd.Flags().StringVar(&renameOpts.OldName, "old", "", "Original name of the variable")
	cmd.Flags().StringVar(&renameOpts.NewName, "new", "", "New name for the variable")
	cmd.Flags().BoolVar(&renameOpts.DryRun, "dry-run", false, "Show changes without applying them")

	// Make the oldName and newName required
	if err := cmd.MarkFlagRequired("old"); err != nil {
		// Handle error or just panic - we'll panic as this is during initialization
		panic(err)
	}
	if err := cmd.MarkFlagRequired("new"); err != nil {
		panic(err)
	}

	return cmd
}

// runRenameVariableCmd executes the variable renaming
func runRenameVariableCmd(cmd *cobra.Command, args []string) error {
	// Validate inputs
	if renameOpts.OldName == "" || renameOpts.NewName == "" {
		return fmt.Errorf("both --old and --new must be provided")
	}

	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Create the variable renamer
	variableRenamer := rename.NewVariableRenamer(renameOpts.OldName, renameOpts.NewName, renameOpts.DryRun)

	// Run the transformation
	dryRunText := ""
	if renameOpts.DryRun {
		dryRunText = " (dry run)"
	}
	fmt.Fprintf(os.Stderr, "Renaming variable '%s' to '%s'%s...\n",
		renameOpts.OldName,
		renameOpts.NewName,
		dryRunText)

	result := variableRenamer.Transform(mod)

	// Handle the result
	if !result.Success {
		return fmt.Errorf("failed to rename variable: %v", result.Error)
	}

	// If this was a dry run, display a preview of the changes
	if renameOpts.DryRun {
		fmt.Println("\nDRY RUN - No changes applied")
		fmt.Printf("Summary: %s\n", result.Summary)

		if len(result.AffectedFiles) > 0 {
			fmt.Printf("\nFiles that would be affected (%d):\n", len(result.AffectedFiles))
			for _, file := range result.AffectedFiles {
				fmt.Printf("  - %s\n", file)
			}
		}

		if len(result.Changes) > 0 {
			fmt.Printf("\nChanges that would be made (%d):\n", len(result.Changes))
			for i, change := range result.Changes {
				fmt.Printf("  %d. In %s", i+1, change.FilePath)
				if change.LineNumber > 0 {
					fmt.Printf(" (line %d)", change.LineNumber)
				}
				fmt.Printf(":\n")
				fmt.Printf("     - Before: %s\n", change.Original)
				fmt.Printf("     - After:  %s\n", change.New)
			}
		}

		fmt.Println("\nRun without --dry-run to apply these changes.")
		return nil
	}

	// Save the result
	if GlobalOptions.OutputDir != "" {
		// Save to output directory
		saver := saver.NewGoModuleSaver()
		fmt.Fprintf(os.Stderr, "Saving renamed module to %s\n", GlobalOptions.OutputDir)
		if err := saver.SaveTo(mod, GlobalOptions.OutputDir); err != nil {
			return fmt.Errorf("failed to save module: %w", err)
		}
	} else {
		// If no output dir provided, save in-place
		saver := saver.NewGoModuleSaver()
		fmt.Fprintf(os.Stderr, "Saving renamed module in-place\n")
		if err := saver.SaveTo(mod, mod.Dir); err != nil {
			return fmt.Errorf("failed to save module: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Successfully renamed '%s' to '%s' in %d file(s)\n",
		renameOpts.OldName, renameOpts.NewName, result.FilesAffected)

	return nil
}
