// Command golm provides a CLI for parsing Go packages and formatting them into a single Go file
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/tree"
)

func main() {
	// Define command-line flags
	srcDir := flag.String("src", ".", "Source directory containing Go package")
	batchDirs := flag.String("batch", "", "Comma-separated list of directories to process in batch mode")
	outFile := flag.String("out-file", "", "Output file (defaults to stdout)")
	outDir := flag.String("out-dir", "", "Output directory where files will be created automatically")
	outputJSON := flag.Bool("json", false, "Output as JSON instead of formatted Go code")
	includeTests := flag.Bool("include-tests", false, "Include test files in parsing")
	preserveFormatting := flag.Bool("preserve-formatting", true, "Preserve original formatting style")
	skipComments := flag.Bool("skip-comments", false, "Skip comments during parsing")
	customPkg := flag.String("package", "", "Custom package name for output (defaults to original)")

	// Legacy support (deprecated, will be removed in future)
	legacyOut := flag.String("out", "", "Deprecated: Use --out-file instead")
	legacyDocsDir := flag.String("docs-dir", "", "Deprecated: Use --out-dir instead")

	// Parse flags
	flag.Parse()

	// Handle legacy flags
	if *legacyOut != "" && *outFile == "" {
		fmt.Fprintf(os.Stderr, "Warning: --out is deprecated, use --out-file instead\n")
		*outFile = *legacyOut
	}

	if *legacyDocsDir != "" && *outDir == "" {
		fmt.Fprintf(os.Stderr, "Warning: --docs-dir is deprecated, use --out-dir instead\n")
		*outDir = *legacyDocsDir
	}

	// Build options
	opts := tree.DefaultOptions()
	opts.IncludeTestFiles = *includeTests
	opts.PreserveFormattingStyle = *preserveFormatting
	opts.SkipComments = *skipComments
	opts.CustomPackageName = *customPkg

	// Check if we're in batch mode
	if *batchDirs != "" {
		directories := strings.Split(*batchDirs, ",")
		fmt.Fprintf(os.Stderr, "Processing %d packages in batch mode\n", len(directories))

		for _, dir := range directories {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}

			fmt.Fprintf(os.Stderr, "Processing package in %s\n", dir)

			// Parse the package
			pkg, err := tree.ParseWithOptions(dir, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing package in %s: %v\n", dir, err)
				continue // Skip this package but continue with others
			}

			var output string
			var outputPath string

			// Process according to format selection
			if *outputJSON {
				// Generate JSON output
				jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error generating JSON for %s: %v\n", dir, err)
					continue
				}
				output = string(jsonData)
			} else {
				// Format as Go code
				formattedCode, err := pkg.FormatWithOptions(opts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error formatting package for %s: %v\n", dir, err)
					continue
				}
				output = formattedCode
			}

			if *outDir != "" {
				// Create output directory if it doesn't exist
				if err := os.MkdirAll(*outDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", *outDir, err)
					continue
				}

				// Use package name as filename with appropriate extension
				ext := ".go"
				if *outputJSON {
					ext = ".json"
				}
				outputPath = filepath.Join(*outDir, pkg.Name()+ext)

				// Write to file
				if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outputPath, err)
					continue
				}
				fmt.Fprintf(os.Stderr, "Successfully wrote package to %s\n", outputPath)
			} else {
				// Without an output directory, print to stdout with separator
				fmt.Printf("--- Package: %s ---\n", pkg.Name())
				fmt.Println(output)
				fmt.Println()
			}
		}
		return
	}

	// Single package mode (original behavior)
	// Parse the package
	pkg, err := tree.ParseWithOptions(*srcDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	var output string

	// Handle different output formats
	if *outputJSON {
		// Generate JSON output
		jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
		output = string(jsonData)
	} else {
		// Format the package as Go code
		formattedCode, err := pkg.FormatWithOptions(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting package: %v\n", err)
			os.Exit(1)
		}
		output = formattedCode
	}

	// Determine where to write the output
	if *outFile != "" {
		// Create output directory if it doesn't exist
		outDir := filepath.Dir(*outFile)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", outDir, err)
			os.Exit(1)
		}

		// Write to specified output file
		if err := os.WriteFile(*outFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", *outFile, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Successfully wrote output to %s\n", *outFile)
	} else if *outDir != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(*outDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", *outDir, err)
			os.Exit(1)
		}

		// Use package name as filename with appropriate extension
		ext := ".go"
		if *outputJSON {
			ext = ".json"
		}
		outputPath := filepath.Join(*outDir, pkg.Name()+ext)

		// Write to file in output directory
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outputPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Successfully wrote output to %s\n", outputPath)
	} else {
		// Write to stdout
		fmt.Print(output)
	}
}
