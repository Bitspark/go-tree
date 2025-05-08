package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkgold/core/loader"
	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/saver"
	"bitspark.dev/go-tree/pkgold/transform/extract"
)

type transformOptions struct {
	// Interface extraction options
	MinTypes        int
	MinMethods      int
	MethodThreshold float64
	NamingStrategy  string
	ExcludePackages string
	ExcludeTypes    string
	ExcludeMethods  string
	CreateNewFiles  bool
	TargetPackage   string
}

var transformOpts transformOptions

// newTransformCmd creates the transform command
func newTransformCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transform",
		Short: "Transform Go module code",
		Long:  `Transforms Go module code using various transformers.`,
	}

	// Add subcommands
	cmd.AddCommand(newExtractCmd())

	return cmd
}

// newExtractCmd creates the extract command
func newExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract interfaces from implementation types",
		Long: `Extract potential interfaces from implementation types.
This analyzes method signatures across different types and creates 
interface definitions for methods that are common across multiple types.`,
		RunE: runExtractCmd,
	}

	// Add flags specific to interface extraction
	cmd.Flags().IntVar(&transformOpts.MinTypes, "min-types", 2, "Minimum number of types for an interface")
	cmd.Flags().IntVar(&transformOpts.MinMethods, "min-methods", 1, "Minimum number of methods for an interface")
	cmd.Flags().Float64Var(&transformOpts.MethodThreshold, "method-threshold", 0.8, "Threshold for method similarity (0.0-1.0)")
	cmd.Flags().StringVar(&transformOpts.NamingStrategy, "naming", "default", "Interface naming strategy (default, prefix, suffix)")
	cmd.Flags().StringVar(&transformOpts.ExcludePackages, "exclude-packages", "", "Comma-separated list of packages to exclude")
	cmd.Flags().StringVar(&transformOpts.ExcludeTypes, "exclude-types", "", "Comma-separated list of types to exclude")
	cmd.Flags().StringVar(&transformOpts.ExcludeMethods, "exclude-methods", "", "Comma-separated list of methods to exclude")
	cmd.Flags().BoolVar(&transformOpts.CreateNewFiles, "create-files", false, "Create new files for interfaces")
	cmd.Flags().StringVar(&transformOpts.TargetPackage, "target-package", "", "Package where interfaces should be created")

	return cmd
}

// runExtractCmd executes the interface extraction
func runExtractCmd(cmd *cobra.Command, args []string) error {
	// Create a loader to load the module
	modLoader := loader.NewGoModuleLoader()

	// Configure load options
	loadOpts := loader.DefaultLoadOptions()
	loadOpts.LoadDocs = true

	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	mod, err := modLoader.LoadWithOptions(GlobalOptions.InputDir, loadOpts)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Configure the extractor
	extractOpts := extract.Options{
		MinimumTypes:    transformOpts.MinTypes,
		MinimumMethods:  transformOpts.MinMethods,
		MethodThreshold: transformOpts.MethodThreshold,
		NamingStrategy:  getNamingStrategy(transformOpts.NamingStrategy),
		TargetPackage:   transformOpts.TargetPackage,
		CreateNewFiles:  transformOpts.CreateNewFiles,
		ExcludePackages: splitCSV(transformOpts.ExcludePackages),
		ExcludeTypes:    splitCSV(transformOpts.ExcludeTypes),
		ExcludeMethods:  splitCSV(transformOpts.ExcludeMethods),
	}

	// Create and run the extractor
	extractor := extract.NewInterfaceExtractor(extractOpts)
	fmt.Fprintln(os.Stderr, "Extracting interfaces...")
	if err := extractor.Transform(mod); err != nil {
		return fmt.Errorf("failed to extract interfaces: %w", err)
	}

	// Save the result
	if GlobalOptions.OutputFile != "" || GlobalOptions.OutputDir != "" {
		// Save to file or directory
		saver := saver.NewGoModuleSaver()

		if GlobalOptions.OutputDir != "" {
			// Save to output directory
			fmt.Fprintf(os.Stderr, "Saving extracted interfaces to %s\n", GlobalOptions.OutputDir)
			if err := saver.SaveTo(mod, GlobalOptions.OutputDir); err != nil {
				return fmt.Errorf("failed to save module: %w", err)
			}
		} else {
			// Save to a specific file
			// This is a simplification - in reality, we'd need to extract just the interfaces
			fmt.Fprintf(os.Stderr, "Saving extracted interfaces to %s\n", GlobalOptions.OutputFile)
			if err := saver.SaveTo(mod, os.TempDir()); err != nil {
				return fmt.Errorf("failed to save module: %w", err)
			}
			// TODO: Copy just the interface file to the output location
		}
	} else {
		// Print to stdout
		fmt.Fprintln(os.Stderr, "Interfaces extracted successfully. Use --output or --out-dir to save to a file.")
	}

	return nil
}

// getNamingStrategy returns the appropriate naming strategy function
func getNamingStrategy(strategy string) extract.NamingStrategy {
	switch strategy {
	case "prefix":
		return prefixNamingStrategy
	case "suffix":
		return suffixNamingStrategy
	default:
		return defaultNamingStrategy
	}
}

// defaultNamingStrategy provides a default naming strategy
func defaultNamingStrategy(types []*module.Type, signatures []string) string {
	// Example: If we have Reader and Writer types with Read() and Write() methods,
	// we might call the interface "ReadWriter"
	if len(types) == 0 {
		return "Interface"
	}

	// Use a simple approach: take first type's name as a base
	baseName := types[0].Name
	// Remove common type suffixes
	baseName = strings.TrimSuffix(baseName, "Impl")
	baseName = strings.TrimSuffix(baseName, "Implementation")

	return baseName + "Interface"
}

// prefixNamingStrategy names interfaces based on common method name prefixes
func prefixNamingStrategy(types []*module.Type, signatures []string) string {
	if len(signatures) == 0 {
		return "Interface"
	}

	// Look for common method name prefixes
	// For simplicity, just use the first method's first word
	methodName := signatures[0]
	parts := strings.FieldsFunc(methodName, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9')
	})

	if len(parts) > 0 {
		// Capitalize the first letter
		prefix := parts[0]
		if len(prefix) > 0 {
			prefix = strings.ToUpper(prefix[:1]) + prefix[1:]
		}
		return prefix + "er"
	}

	return "Interface"
}

// suffixNamingStrategy names interfaces based on common type name suffixes
func suffixNamingStrategy(types []*module.Type, signatures []string) string {
	if len(types) == 0 {
		return "Interface"
	}

	// Collect all type names
	typeNames := make([]string, 0, len(types))
	for _, t := range types {
		typeNames = append(typeNames, t.Name)
	}

	// For simplicity, use a common suffix if found
	for _, t := range typeNames {
		if strings.HasSuffix(t, "Handler") {
			return "Handler"
		}
		if strings.HasSuffix(t, "Service") {
			return "Service"
		}
		if strings.HasSuffix(t, "Repository") {
			return "Repository"
		}
	}

	return "Interface"
}

// splitCSV splits a comma-separated string into a slice
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
