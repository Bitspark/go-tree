package commands

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"bitspark.dev/go-tree/pkg/core/loader"
	"bitspark.dev/go-tree/pkg/index"
)

type findOptions struct {
	// Find options
	Symbol         string
	Type           string
	IncludeTests   bool
	IncludePrivate bool
	Format         string
}

var findOpts findOptions

// NewFindCmd creates the find command
func NewFindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find elements and their usages in Go code",
		Long:  `Finds elements like types, functions, and variables in Go code and analyzes their usages.`,
	}

	// Add subcommands
	cmd.AddCommand(newFindUsagesCmd())
	cmd.AddCommand(newFindTypesCmd())

	return cmd
}

// newFindUsagesCmd creates the find usages command
func newFindUsagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usages",
		Short: "Find all usages of a symbol",
		Long:  `Finds all references to a given symbol (function, variable, type, etc.) in the codebase.`,
		RunE:  runFindUsagesCmd,
	}

	// Add flags
	cmd.Flags().StringVar(&findOpts.Symbol, "symbol", "", "The symbol to find usages of")
	cmd.Flags().StringVar(&findOpts.Type, "type", "", "Optional type name to scope the search (for methods/fields)")
	cmd.Flags().BoolVar(&findOpts.IncludeTests, "include-tests", false, "Include test files in search")
	cmd.Flags().BoolVar(&findOpts.IncludePrivate, "include-private", false, "Include private (unexported) elements")
	cmd.Flags().StringVar(&findOpts.Format, "format", "text", "Output format (text, json)")

	// Make the symbol flag required
	if err := cmd.MarkFlagRequired("symbol"); err != nil {
		panic(err)
	}

	return cmd
}

// newFindTypesCmd creates the find types command
func newFindTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Find all types in the codebase",
		Long:  `Lists all types defined in the codebase and their attributes.`,
		RunE:  runFindTypesCmd,
	}

	// Add flags
	cmd.Flags().BoolVar(&findOpts.IncludeTests, "include-tests", false, "Include test files in search")
	cmd.Flags().BoolVar(&findOpts.IncludePrivate, "include-private", false, "Include private (unexported) elements")
	cmd.Flags().StringVar(&findOpts.Format, "format", "text", "Output format (text, json)")

	return cmd
}

// runFindUsagesCmd executes the find usages command
func runFindUsagesCmd(cmd *cobra.Command, args []string) error {
	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load(GlobalOptions.InputDir)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Build an index of the module
	fmt.Fprintf(os.Stderr, "Building index...\n")
	indexer := index.NewIndexer(mod).
		WithTests(findOpts.IncludeTests).
		WithPrivate(findOpts.IncludePrivate)

	idx, err := indexer.BuildIndex()
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Find symbols matching the name
	symbols := idx.FindSymbolsByName(findOpts.Symbol)
	if len(symbols) == 0 {
		return fmt.Errorf("no symbols found with name '%s'", findOpts.Symbol)
	}

	// Filter by type if specified
	if findOpts.Type != "" {
		var filtered []*index.Symbol
		for _, sym := range symbols {
			if sym.ParentType == findOpts.Type || sym.ReceiverType == findOpts.Type {
				filtered = append(filtered, sym)
			}
		}
		symbols = filtered

		if len(symbols) == 0 {
			return fmt.Errorf("no symbols found with name '%s' on type '%s'", findOpts.Symbol, findOpts.Type)
		}
	}

	// Output findings
	if findOpts.Format == "json" {
		// TODO: Implement JSON output if needed
		return fmt.Errorf("JSON output format not yet implemented")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		for i, symbol := range symbols {
			// Print header for each found symbol
			if i > 0 {
				fmt.Fprintln(w, "---")
			}

			// Output symbol info
			fmt.Fprintf(w, "Symbol: %s\n", symbol.Name)
			fmt.Fprintf(w, "Kind: %v\n", symbol.Kind)
			fmt.Fprintf(w, "Package: %s\n", symbol.Package)
			fmt.Fprintf(w, "Defined at: %s:%d\n", symbol.File, symbol.LineStart)

			if symbol.ParentType != "" {
				fmt.Fprintf(w, "Type: %s\n", symbol.ParentType)
			}
			if symbol.ReceiverType != "" {
				fmt.Fprintf(w, "Receiver: %s\n", symbol.ReceiverType)
			}

			// Find references to this symbol
			references := idx.FindReferences(symbol)
			fmt.Fprintf(w, "\nFound %d references:\n", len(references))

			// Sort references by file and line number
			sort.Slice(references, func(i, j int) bool {
				if references[i].File != references[j].File {
					return references[i].File < references[j].File
				}
				return references[i].LineStart < references[j].LineStart
			})

			// Group references by file
			refsByFile := make(map[string][]*index.Reference)
			for _, ref := range references {
				refsByFile[ref.File] = append(refsByFile[ref.File], ref)
			}

			// Output references by file
			fileKeys := make([]string, 0, len(refsByFile))
			for file := range refsByFile {
				fileKeys = append(fileKeys, file)
			}
			sort.Strings(fileKeys)

			for _, file := range fileKeys {
				refs := refsByFile[file]
				fmt.Fprintf(w, "  File: %s\n", file)

				for _, ref := range refs {
					context := ""
					if ref.Context != "" {
						context = fmt.Sprintf(" (in %s)", ref.Context)
					}
					fmt.Fprintf(w, "    Line %d%s\n", ref.LineStart, context)
				}
			}

			fmt.Fprintln(w)
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return nil
}

// runFindTypesCmd executes the find types command
func runFindTypesCmd(cmd *cobra.Command, args []string) error {
	// Load the module
	fmt.Fprintf(os.Stderr, "Loading module from %s\n", GlobalOptions.InputDir)
	moduleLoader := loader.NewGoModuleLoader()
	mod, err := moduleLoader.Load(GlobalOptions.InputDir)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	// Build an index of the module
	fmt.Fprintf(os.Stderr, "Building index...\n")
	indexer := index.NewIndexer(mod).
		WithTests(findOpts.IncludeTests).
		WithPrivate(findOpts.IncludePrivate)

	idx, err := indexer.BuildIndex()
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Collect all type symbols
	var types []*index.Symbol
	for _, symbols := range idx.SymbolsByName {
		for _, symbol := range symbols {
			if symbol.Kind == index.KindType {
				types = append(types, symbol)
			}
		}
	}

	// Sort types by package and name
	sort.Slice(types, func(i, j int) bool {
		if types[i].Package != types[j].Package {
			return types[i].Package < types[j].Package
		}
		return types[i].Name < types[j].Name
	})

	// Output findings
	if findOpts.Format == "json" {
		// TODO: Implement JSON output if needed
		return fmt.Errorf("JSON output format not yet implemented")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintf(w, "Found %d types:\n\n", len(types))

		// Group types by package
		typesByPkg := make(map[string][]*index.Symbol)
		for _, t := range types {
			typesByPkg[t.Package] = append(typesByPkg[t.Package], t)
		}

		pkgKeys := make([]string, 0, len(typesByPkg))
		for pkg := range typesByPkg {
			pkgKeys = append(pkgKeys, pkg)
		}
		sort.Strings(pkgKeys)

		for _, pkg := range pkgKeys {
			pkgTypes := typesByPkg[pkg]
			fmt.Fprintf(w, "Package: %s\n", pkg)

			for _, t := range pkgTypes {
				// Find fields and methods for this type
				typeName := t.Name
				symbols := idx.FindSymbolsForType(typeName)

				var fields, methods int
				for _, sym := range symbols {
					if sym.Kind == index.KindField {
						fields++
					} else if sym.Kind == index.KindMethod {
						methods++
					}
				}

				fmt.Fprintf(w, "  %s (fields: %d, methods: %d)\n", typeName, fields, methods)
			}

			fmt.Fprintln(w)
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return nil
}
