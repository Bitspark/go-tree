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
	outFile := flag.String("out", "", "Output file (defaults to stdout)")
	outputJSON := flag.Bool("json", false, "Output as JSON instead of formatted Go code")
	docsDir := flag.String("docs-dir", "", "Output directory for documentation JSON files")
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

			// Only JSON output makes sense in batch mode
			jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating JSON for %s: %v\n", dir, err)
				continue
			}

			if *docsDir != "" {
				// Create docs directory if it doesn't exist
				if err := os.MkdirAll(*docsDir, 0755); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating docs directory %s: %v\n", *docsDir, err)
					continue
				}

				// Use package name as filename
				outPath := filepath.Join(*docsDir, pkg.Name()+".json")
				if err := os.WriteFile(outPath, jsonData, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing JSON for %s to %s: %v\n", dir, outPath, err)
					continue
				}
				fmt.Fprintf(os.Stderr, "Successfully wrote package JSON for %s to %s\n", dir, outPath)
			} else {
				// Without a docs directory, print to stdout with separator
				fmt.Printf("--- Package: %s ---\n", pkg.Name())
				fmt.Println(string(jsonData))
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

	// Handle different output modes
	if *outputJSON {
		// Generate JSON output
		jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}

		if *docsDir != "" {
			// Create docs directory if it doesn't exist
			if err := os.MkdirAll(*docsDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating docs directory %s: %v\n", *docsDir, err)
				os.Exit(1)
			}

			// Write to file in docs directory
			outPath := filepath.Join(*docsDir, pkg.Name()+".json")
			if err := os.WriteFile(outPath, jsonData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", outPath, err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Successfully wrote package JSON to %s\n", outPath)
		} else if *outFile != "" {
			// Create output directory if it doesn't exist
			outDir := filepath.Dir(*outFile)
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", outDir, err)
				os.Exit(1)
			}

			// Write to specified output file
			if err := os.WriteFile(*outFile, jsonData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing JSON to %s: %v\n", *outFile, err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Successfully wrote package JSON to %s\n", *outFile)
		} else {
			// Write to stdout
			fmt.Println(string(jsonData))
		}
	} else {
		// Format the package as Go code (original behavior)
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
}
