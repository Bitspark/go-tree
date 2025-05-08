package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkgold/core/loader"
)

type analyzeOptions struct {
	// Analysis options
	Format         string
	IncludePrivate bool
	IncludeTests   bool
	SortByName     bool
	SortBySize     bool
	MaxDepth       int
	ShowInterfaces bool
	ShowTypes      bool
	ShowFunctions  bool
	ShowDeps       bool
}

var analyzeOpts analyzeOptions

// newAnalyzeCmd creates the analyze command
func newAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze Go module structure",
		Long:  `Analyzes the structure and content of a Go module.`,
	}

	// Common analyze flags
	cmd.PersistentFlags().StringVar(&analyzeOpts.Format, "format", "text", "Output format (text, json)")
	cmd.PersistentFlags().BoolVar(&analyzeOpts.IncludePrivate, "include-private", false, "Include private (unexported) elements")
	cmd.PersistentFlags().BoolVar(&analyzeOpts.IncludeTests, "include-tests", false, "Include test files")
	cmd.PersistentFlags().BoolVar(&analyzeOpts.SortByName, "sort-by-name", true, "Sort by name")
	cmd.PersistentFlags().BoolVar(&analyzeOpts.SortBySize, "sort-by-size", false, "Sort by size/count (overrides sort-by-name)")
	cmd.PersistentFlags().IntVar(&analyzeOpts.MaxDepth, "max-depth", 0, "Maximum depth to traverse (0 means unlimited)")

	// Add subcommands
	cmd.AddCommand(newStructureCmd())
	cmd.AddCommand(newInterfacesCmd())

	return cmd
}

// newStructureCmd creates the structure analysis command
func newStructureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "structure",
		Short: "Analyze module structure",
		Long:  `Analyzes the packages, types, and functions in a module.`,
		RunE:  runStructureCmd,
	}

	// Additional flags for structure analysis
	cmd.Flags().BoolVar(&analyzeOpts.ShowTypes, "show-types", true, "Show type definitions")
	cmd.Flags().BoolVar(&analyzeOpts.ShowFunctions, "show-functions", true, "Show functions")
	cmd.Flags().BoolVar(&analyzeOpts.ShowDeps, "show-deps", false, "Show dependencies")

	return cmd
}

// newInterfacesCmd creates the interfaces analysis command
func newInterfacesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interfaces",
		Short: "Analyze interfaces and implementations",
		Long:  `Analyzes interfaces and their implementations in the module.`,
		RunE:  runInterfacesCmd,
	}

	// Additional flags for interface analysis
	cmd.Flags().BoolVar(&analyzeOpts.ShowInterfaces, "show-interfaces", true, "Show interface definitions")

	return cmd
}

// runStructureCmd executes the structure analysis
func runStructureCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()
	loadOpts.IncludeTests = analyzeOpts.IncludeTests
	loadOpts.LoadDocs = true

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Generate structure analysis
	if analyzeOpts.Format == "json" {
		// Output module structure as JSON
		jsonData, err := json.MarshalIndent(mod, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize module to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Text output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if _, err := fmt.Fprintf(w, "Module: %s\n", mod.Path); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
		if mod.Version != "" {
			if _, err := fmt.Fprintf(w, "Version: %s\n", mod.Version); err != nil {
				return fmt.Errorf("failed to write to output: %w", err)
			}
		}
		if _, err := fmt.Fprintf(w, "Go Version: %s\n", mod.GoVersion); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "Directory: %s\n\n", mod.Dir); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}

		// Packages
		if _, err := fmt.Fprintf(w, "Packages (%d):\n", len(mod.Packages)); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}

		// Create sorted list of packages
		pkgs := make([]string, 0, len(mod.Packages))
		for pkgPath := range mod.Packages {
			pkgs = append(pkgs, pkgPath)
		}
		sort.Strings(pkgs)

		// Display package information
		for _, pkgPath := range pkgs {
			pkg := mod.Packages[pkgPath]
			if _, err := fmt.Fprintf(w, "  %s\n", pkgPath); err != nil {
				return fmt.Errorf("failed to write to output: %w", err)
			}

			// Types
			if analyzeOpts.ShowTypes && len(pkg.Types) > 0 {
				if _, err := fmt.Fprintf(w, "    Types (%d):\n", len(pkg.Types)); err != nil {
					return fmt.Errorf("failed to write to output: %w", err)
				}
				types := make([]string, 0, len(pkg.Types))
				for typeName, typeObj := range pkg.Types {
					if !analyzeOpts.IncludePrivate && !typeObj.IsExported {
						continue
					}
					types = append(types, typeName)
				}
				sort.Strings(types)

				for _, typeName := range types {
					typeObj := pkg.Types[typeName]
					exported := ""
					if !typeObj.IsExported {
						exported = " (unexported)"
					}
					if _, err := fmt.Fprintf(w, "      %s %s%s\n", typeName, typeObj.Kind, exported); err != nil {
						return fmt.Errorf("failed to write to output: %w", err)
					}
				}
			}

			// Functions
			if analyzeOpts.ShowFunctions && len(pkg.Functions) > 0 {
				if _, err := fmt.Fprintf(w, "    Functions (%d):\n", len(pkg.Functions)); err != nil {
					return fmt.Errorf("failed to write to output: %w", err)
				}
				funcs := make([]string, 0, len(pkg.Functions))
				for funcName, funcObj := range pkg.Functions {
					if !analyzeOpts.IncludePrivate && !funcObj.IsExported {
						continue
					}
					funcs = append(funcs, funcName)
				}
				sort.Strings(funcs)

				for _, funcName := range funcs {
					funcObj := pkg.Functions[funcName]
					exported := ""
					if !funcObj.IsExported {
						exported = " (unexported)"
					}
					if _, err := fmt.Fprintf(w, "      %s%s\n", funcName, exported); err != nil {
						return fmt.Errorf("failed to write to output: %w", err)
					}
				}
			}

			if _, err := fmt.Fprintln(w); err != nil {
				return fmt.Errorf("failed to write to output: %w", err)
			}
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return nil
}

// runInterfacesCmd executes the interfaces analysis
func runInterfacesCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()
	loadOpts.IncludeTests = analyzeOpts.IncludeTests
	loadOpts.LoadDocs = true

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Find interfaces and their implementations
	interfaces := make(map[string]map[string]bool) // interface name -> implementors

	// First pass: collect interfaces
	for _, pkg := range mod.Packages {
		for typeName, typeObj := range pkg.Types {
			if typeObj.Kind == "interface" {
				if !analyzeOpts.IncludePrivate && !typeObj.IsExported {
					continue
				}
				fullName := pkg.ImportPath + "." + typeName
				interfaces[fullName] = make(map[string]bool)
			}
		}
	}

	// Unfortunately a proper implementation would require deeper analysis
	// to match interface methods with implementations, which is beyond
	// the scope of this implementation. For now, just show interfaces.

	// Output results
	if analyzeOpts.Format == "json" {
		// JSON output
		jsonData, err := json.MarshalIndent(interfaces, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize interfaces to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Text output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if _, err := fmt.Fprintf(w, "Interfaces in %s:\n\n", mod.Path); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}

		// Create sorted list of interfaces
		ifaceNames := make([]string, 0, len(interfaces))
		for name := range interfaces {
			ifaceNames = append(ifaceNames, name)
		}
		sort.Strings(ifaceNames)

		// Display interface information
		for _, name := range ifaceNames {
			if _, err := fmt.Fprintf(w, "%s\n", name); err != nil {
				return fmt.Errorf("failed to write to output: %w", err)
			}

			// Split name into package and type
			// ... code to find and display interface methods would go here
			// in a full implementation
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return nil
}
