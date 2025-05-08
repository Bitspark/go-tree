package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"bitspark.dev/go-tree/pkg/typesys"
)

// CommandContext represents the context for executing index commands.
type CommandContext struct {
	// The indexer
	Indexer *Indexer

	// Output settings
	Verbose   bool   // Whether to output verbose information
	OutputFmt string // Output format (text, json)

	// Filter settings
	FilterTests   bool // Whether to filter out test files
	FilterPrivate bool // Whether to filter out private symbols
}

// NewCommandContext creates a new command context.
func NewCommandContext(module *typesys.Module, opts IndexingOptions) (*CommandContext, error) {
	// Create the indexer
	indexer := NewIndexer(module, opts)

	// Build the index
	if err := indexer.BuildIndex(); err != nil {
		return nil, fmt.Errorf("failed to build index: %w", err)
	}

	return &CommandContext{
		Indexer:       indexer,
		Verbose:       false,
		OutputFmt:     "text",
		FilterTests:   !opts.IncludeTests,
		FilterPrivate: !opts.IncludePrivate,
	}, nil
}

// FindUsages finds all usages of a symbol with the given name.
func (ctx *CommandContext) FindUsages(name string, file string, line, column int) error {
	var symbol *typesys.Symbol

	// If file and position provided, look for symbol at that position
	if file != "" && line > 0 {
		// Resolve file path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}

		// Find symbol at position
		symbol = ctx.Indexer.FindSymbolAtPosition(absPath, line, column)
		if symbol == nil {
			// Try to find a reference at that position
			ref := ctx.Indexer.FindReferenceAtPosition(absPath, line, column)
			if ref != nil {
				symbol = ref.Symbol
			}
		}
	}

	// If not found by position, try by name
	if symbol == nil && name != "" {
		// Find by name
		symbols := ctx.Indexer.FindSymbolByNameAndType(name)
		if len(symbols) == 0 {
			return fmt.Errorf("no symbol found with name: %s", name)
		}

		// If multiple symbols found, print a list and ask for clarification
		if len(symbols) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple symbols found with name '%s':\n", name)
			for i, sym := range symbols {
				var location string
				if sym.File != nil {
					pos := sym.GetPosition()
					if pos != nil {
						location = fmt.Sprintf("%s:%d", sym.File.Path, pos.LineStart)
					} else {
						location = sym.File.Path
					}
				}

				fmt.Fprintf(os.Stderr, "  %d. %s (%s) at %s\n", i+1, sym.Name, sym.Kind, location)
			}

			// For now, just use the first one
			fmt.Fprintf(os.Stderr, "Using first match: %s (%s)\n", symbols[0].Name, symbols[0].Kind)
			symbol = symbols[0]
		} else {
			symbol = symbols[0]
		}
	}

	if symbol == nil {
		return fmt.Errorf("could not find symbol")
	}

	// Find usages
	references := ctx.Indexer.FindUsages(symbol)

	// Print output
	if ctx.Verbose {
		fmt.Printf("Found %d usages of '%s' (%s)\n", len(references), symbol.Name, symbol.Kind)
	}

	// Create a tab writer for formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing writer: %v\n", err)
		}
	}()

	// Print header
	if _, err := fmt.Fprintln(w, "File\tLine\tColumn\tContext"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Print usages
	for _, ref := range references {
		var context string
		if ref.Context != nil {
			context = ref.Context.Name
		}

		pos := ref.GetPosition()
		if pos != nil {
			if _, err := fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", ref.File.Path, pos.LineStart, pos.ColumnStart, context); err != nil {
				return fmt.Errorf("failed to write reference: %w", err)
			}
		} else {
			if _, err := fmt.Fprintf(w, "%s\t-\t-\t%s\n", ref.File.Path, context); err != nil {
				return fmt.Errorf("failed to write reference: %w", err)
			}
		}
	}

	return nil
}

// SearchSymbols searches for symbols matching the given pattern.
func (ctx *CommandContext) SearchSymbols(pattern string, kindFilter string) error {
	var symbols []*typesys.Symbol

	// Apply kind filter if provided
	if kindFilter != "" {
		// Parse kind filter
		kinds := parseKindFilter(kindFilter)

		// Search for each kind
		for _, kind := range kinds {
			// Find symbols of this kind
			for _, sym := range ctx.Indexer.Index.FindSymbolsByKind(kind) {
				if strings.Contains(sym.Name, pattern) {
					symbols = append(symbols, sym)
				}
			}
		}
	} else {
		// General search
		symbols = ctx.Indexer.Search(pattern)
	}

	// Print output
	if ctx.Verbose {
		fmt.Printf("Found %d symbols matching '%s'\n", len(symbols), pattern)
	}

	// Create a tab writer for formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing writer: %v\n", err)
		}
	}()

	// Print header
	if _, err := fmt.Fprintln(w, "Name\tKind\tPackage\tFile\tLine"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Print symbols
	for _, sym := range symbols {
		var location string
		pos := sym.GetPosition()
		if pos != nil {
			location = fmt.Sprintf("%d", pos.LineStart)
		} else {
			location = "-"
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", sym.Name, sym.Kind, sym.Package.Name, sym.File.Path, location); err != nil {
			return fmt.Errorf("failed to write symbol: %w", err)
		}
	}

	return nil
}

// ListFileSymbols lists all symbols in a file.
func (ctx *CommandContext) ListFileSymbols(filePath string) error {
	// Resolve file path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Get symbols in file
	symbolsByKind := ctx.Indexer.GetFileSymbols(absPath)

	// Calculate total count
	var total int
	for _, symbols := range symbolsByKind {
		total += len(symbols)
	}

	// Print output
	if ctx.Verbose {
		fmt.Printf("Found %d symbols in %s\n", total, filePath)
	}

	// Create a tab writer for formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "Error flushing writer: %v\n", err)
		}
	}()

	// Print header
	if _, err := fmt.Fprintln(w, "Name\tKind\tLine\tColumn"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Process kinds in a specific order
	kindOrder := []typesys.SymbolKind{
		typesys.KindType,
		typesys.KindStruct,
		typesys.KindInterface,
		typesys.KindFunction,
		typesys.KindMethod,
		typesys.KindVariable,
		typesys.KindConstant,
		typesys.KindField,
	}

	// Print symbols by kind
	for _, kind := range kindOrder {
		symbols := symbolsByKind[kind]
		if len(symbols) == 0 {
			continue
		}

		// Print symbols of this kind
		for _, sym := range symbols {
			var line, column string
			pos := sym.GetPosition()
			if pos != nil {
				line = fmt.Sprintf("%d", pos.LineStart)
				column = fmt.Sprintf("%d", pos.ColumnStart)
			} else {
				line = "-"
				column = "-"
			}

			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", sym.Name, sym.Kind, line, column); err != nil {
				return fmt.Errorf("failed to write symbol: %w", err)
			}
		}
	}

	return nil
}

// Helper function to parse a kind filter string into a list of SymbolKinds
func parseKindFilter(kindFilter string) []typesys.SymbolKind {
	var kinds []typesys.SymbolKind

	// Split by comma
	tokens := strings.Split(kindFilter, ",")
	for _, token := range tokens {
		token = strings.TrimSpace(token)

		// Map to kinds
		switch strings.ToLower(token) {
		case "type", "types":
			kinds = append(kinds, typesys.KindType)
		case "struct", "structs":
			kinds = append(kinds, typesys.KindStruct)
		case "interface", "interfaces":
			kinds = append(kinds, typesys.KindInterface)
		case "function", "func", "functions", "funcs":
			kinds = append(kinds, typesys.KindFunction)
		case "method", "methods":
			kinds = append(kinds, typesys.KindMethod)
		case "variable", "var", "variables", "vars":
			kinds = append(kinds, typesys.KindVariable)
		case "constant", "const", "constants", "consts":
			kinds = append(kinds, typesys.KindConstant)
		case "field", "fields":
			kinds = append(kinds, typesys.KindField)
		case "import", "imports":
			kinds = append(kinds, typesys.KindImport)
		case "package", "packages":
			kinds = append(kinds, typesys.KindPackage)
		}
	}

	return kinds
}
