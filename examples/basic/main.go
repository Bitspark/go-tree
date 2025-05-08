// Example usage of the go-tree module-centered architecture
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/core/loader"
	"bitspark.dev/go-tree/pkg/core/saver"
)

func main() {
	// Check if a directory was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	// Create a module loader
	goLoader := loader.NewGoModuleLoader()

	// Load the module
	mod, err := goLoader.Load(os.Args[1])
	if err != nil {
		fmt.Printf("Error loading module: %v\n", err)
		os.Exit(1)
	}

	// Print module information
	fmt.Printf("Module: %s\n", mod.Path)

	// Print package information
	fmt.Println("\nPackages:")
	for pkgPath, pkg := range mod.Packages {
		fmt.Printf("  - %s (%s)\n", pkg.Name, pkgPath)

		fmt.Println("\n  Imports:")
		for _, imp := range pkg.Imports {
			fmt.Printf("    - %s\n", imp.Path)
		}

		fmt.Println("\n  Functions:")
		for name, fn := range pkg.Functions {
			if fn.IsMethod {
				fmt.Printf("    - method %s on %s\n", name, fn.Receiver.Type)
			} else {
				fmt.Printf("    - func %s\n", name)
			}
		}

		fmt.Println("\n  Types:")
		for name, t := range pkg.Types {
			fmt.Printf("    - %s %s\n", t.Kind, name)
		}

		fmt.Println("\n  Constants:")
		for name := range pkg.Constants {
			fmt.Printf("    - %s\n", name)
		}

		fmt.Println("\n  Variables:")
		for name := range pkg.Variables {
			fmt.Printf("    - %s\n", name)
		}
	}

	// Format and save the module to a single directory
	outDir := "formatted"
	goSaver := saver.NewGoModuleSaver()

	// Ensure the output directory exists
	if err := os.MkdirAll(outDir, 0750); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", outDir, err)
		os.Exit(1)
	}

	// Save the module
	if err := goSaver.SaveTo(mod, outDir); err != nil {
		fmt.Printf("Error saving module: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nModule formatted and saved to %s/\n", outDir)
	fmt.Println("Files created:")
	files, _ := os.ReadDir(outDir)
	for _, file := range files {
		if !file.IsDir() {
			fmt.Printf("  - %s\n", filepath.Join(outDir, file.Name()))
		}
	}
}
