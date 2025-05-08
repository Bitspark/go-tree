// Package example demonstrates how to use the Go-Tree index package.
package main

import (
	"fmt"
	"log"
	"os"

	"bitspark.dev/go-tree/pkg/index"
	"bitspark.dev/go-tree/pkg/typesys"
)

func main() {
	// Get the module directory (default to current directory)
	moduleDir := "."
	if len(os.Args) > 1 {
		moduleDir = os.Args[1]
	}

	// Load the module with type system
	module, err := typesys.LoadModule(moduleDir, &typesys.LoadOptions{
		IncludeTests:   true,
		IncludePrivate: true,
		Trace:          true, // Enable verbose output
	})
	if err != nil {
		log.Fatalf("Failed to load module: %v", err)
	}

	fmt.Printf("Loaded module: %s with %d packages\n", module.Path, len(module.Packages))

	// Create indexer
	indexer := index.NewIndexer(module, index.IndexingOptions{
		IncludeTests:       true,
		IncludePrivate:     true,
		IncrementalUpdates: true,
	})

	// Build the index
	fmt.Println("Building index...")
	if err := indexer.BuildIndex(); err != nil {
		log.Fatalf("Failed to build index: %v", err)
	}

	// Example: Find all interfaces in the module
	fmt.Println("\nInterfaces in the module:")
	interfaces := indexer.Index.FindSymbolsByKind(typesys.KindInterface)
	for _, iface := range interfaces {
		fmt.Printf("- %s (in %s)\n", iface.Name, iface.Package.Name)

		// Find implementations of this interface
		impls := indexer.FindImplementations(iface)
		if len(impls) > 0 {
			fmt.Printf("  Implementations:\n")
			for _, impl := range impls {
				fmt.Printf("  - %s (in %s)\n", impl.Name, impl.Package.Name)
			}
		}
	}

	// Example: Find all functions with "Find" in their name
	fmt.Println("\nFunctions containing 'Find':")
	findFuncs := indexer.FindAllFunctions("Find")
	for _, fn := range findFuncs {
		pos := fn.GetPosition()
		var location string
		if pos != nil {
			location = fmt.Sprintf("%s:%d", fn.File.Path, pos.LineStart)
		} else {
			location = fn.File.Path
		}
		fmt.Printf("- %s at %s\n", fn.Name, location)
	}

	// Example: Find usages of a symbol
	if len(findFuncs) > 0 {
		fmt.Printf("\nUsages of '%s':\n", findFuncs[0].Name)
		refs := indexer.FindUsages(findFuncs[0])
		for _, ref := range refs {
			pos := ref.GetPosition()
			if pos != nil {
				fmt.Printf("- %s:%d:%d\n", ref.File.Path, pos.LineStart, pos.ColumnStart)
			}
		}
	}

	// Example: Get file structure
	if len(os.Args) > 2 {
		filePath := os.Args[2]
		fmt.Printf("\nStructure of %s:\n", filePath)

		structure := indexer.GetFileStructure(filePath)
		printStructure(structure, "")
	}
}

// Helper function to print the symbol tree
func printStructure(nodes []*index.SymbolNode, indent string) {
	for _, node := range nodes {
		sym := node.Symbol
		var typeInfo string
		if sym.TypeInfo != nil {
			typeInfo = fmt.Sprintf(" : %s", sym.TypeInfo)
		}
		fmt.Printf("%s- %s (%s)%s\n", indent, sym.Name, sym.Kind, typeInfo)

		// Recursively print children with increased indent
		if len(node.Children) > 0 {
			printStructure(node.Children, indent+"  ")
		}
	}
}
