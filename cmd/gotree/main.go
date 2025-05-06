// Command golm provides a CLI for parsing Go packages and formatting them into a single Go file
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"bitspark.dev/go-tree/tree"
)

func main() {
	// Define command-line flags
	srcDir := flag.String("src", ".", "Source directory containing Go package")
	outFile := flag.String("out", "", "Output file (defaults to stdout)")
	includeTests := flag.Bool("include-tests", false, "Include test files in parsing")
	preserveFormatting := flag.Bool("preserve-formatting", true, "Preserve original formatting style")
	skipComments := flag.Bool("skip-comments", false, "Skip comments during parsing")
	customPkg := flag.String("package", "", "Custom package name for output (defaults to original)")

	// Parse flags
	flag.Parse()

	// Build options
	opts := tree.DefaultOptions()
	opts.IncludeTestFiles = *includeTests
	opts.PreserveFormattingStyle = *preserveFormatting
	opts.SkipComments = *skipComments
	opts.CustomPackageName = *customPkg

	// Parse the package
	pkg, err := tree.ParseWithOptions(*srcDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	// Format the package
	output, err := pkg.FormatWithOptions(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting package: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if *outFile == "" {
		// Write to stdout
		fmt.Print(output)
	} else {
		// Create output directory if it doesn't exist
		outDir := filepath.Dir(*outFile)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", outDir, err)
			os.Exit(1)
		}

		// Write to file
		if err := os.WriteFile(*outFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", *outFile, err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Successfully wrote package to %s\n", *outFile)
	}
}
