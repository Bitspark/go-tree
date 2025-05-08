// Command visualize generates visualizations of Go modules.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/visual/cmd"
)

func main() {
	// Parse command line flags
	moduleDir := flag.String("dir", ".", "Directory of the Go module to visualize")
	outputFile := flag.String("output", "", "Output file path (defaults to stdout)")
	format := flag.String("format", "html", "Output format (html, markdown)")
	includeTypes := flag.Bool("types", true, "Include type annotations")
	includePrivate := flag.Bool("private", false, "Include private elements")
	includeTests := flag.Bool("tests", false, "Include test files")
	title := flag.String("title", "", "Custom title for the visualization")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Ensure module directory exists
	if _, err := os.Stat(*moduleDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory %s does not exist\n", *moduleDir)
		os.Exit(1)
	}

	// If output file is specified, ensure it has the correct extension
	if *outputFile != "" {
		switch *format {
		case "html":
			if !hasExtension(*outputFile, ".html") {
				*outputFile = *outputFile + ".html"
			}
		case "markdown", "md":
			if !hasExtension(*outputFile, ".md") {
				*outputFile = *outputFile + ".md"
			}
			*format = "markdown" // Normalize format name
		}
	}

	// Create visualization options
	opts := &cmd.VisualizeOptions{
		ModuleDir:      *moduleDir,
		OutputFile:     *outputFile,
		Format:         *format,
		IncludeTypes:   *includeTypes,
		IncludePrivate: *includePrivate,
		IncludeTests:   *includeTests,
		Title:          *title,
	}

	// Generate visualization
	if err := cmd.Visualize(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Helper function to print usage information
func printHelp() {
	fmt.Println("Visualize: Generate visualizations of Go modules")
	fmt.Println("\nUsage:")
	fmt.Println("  visualize [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  visualize -dir ./myproject -format html -output docs/module.html")
	fmt.Println("  visualize -dir . -format markdown -output README.md -types=false")
}

// Helper function to check if a file has a specific extension
func hasExtension(path, ext string) bool {
	fileExt := filepath.Ext(path)
	return fileExt == ext
}
