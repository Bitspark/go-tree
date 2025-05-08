// Command golm provides a CLI for parsing Go packages and formatting them into a single Go file
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/analysis/interfaceanalysis"
	"bitspark.dev/go-tree/pkg/visual/html"
	"bitspark.dev/go-tree/tree"
)

func main() {
	// Define command-line flags
	srcDir := flag.String("src", ".", "Source directory containing Go package")
	batchDirs := flag.String("batch", "", "Comma-separated list of directories to process in batch mode")
	outFile := flag.String("out-file", "", "Output file (defaults to stdout)")
	outDir := flag.String("out-dir", "", "Output directory where files will be created automatically")
	outputJSON := flag.Bool("json", false, "Output as JSON instead of formatted Go code")
	outputHTML := flag.Bool("html", false, "Output as HTML documentation")
	htmlTitle := flag.String("html-title", "", "Custom title for HTML output")
	analyzeReceivers := flag.Bool("analyze-receivers", false, "Analyze method receivers and group by type")
	extractInterfaces := flag.Bool("extract-interfaces", false, "Extract potential interfaces from method receivers")
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

	// Process packages in batch or single mode
	if *batchDirs != "" {
		processBatchDirectories(*batchDirs, opts, *outDir, *outputJSON, *outputHTML, *htmlTitle, *analyzeReceivers, *extractInterfaces)
		return
	}

	// Single package mode
	processPackage(*srcDir, *outFile, *outDir, opts, *outputJSON, *outputHTML, *htmlTitle, *analyzeReceivers, *extractInterfaces)
}

// processPackage handles processing a single package
func processPackage(srcDir, outFile, outDir string, opts *tree.Options, outputJSON, outputHTML bool, htmlTitle string, analyzeReceivers, extractInterfaces bool) {
	// Parse the package
	pkg, err := tree.ParseWithOptions(srcDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing package: %v\n", err)
		os.Exit(1)
	}

	// Special receiver analysis mode
	if analyzeReceivers || extractInterfaces {
		processReceiverAnalysis(pkg, outFile, outDir, extractInterfaces)
		return
	}

	var output string
	var extension string

	// Handle different output formats
	if outputHTML {
		// Generate HTML output
		htmlOpts := html.DefaultOptions()
		if htmlTitle != "" {
			htmlOpts.Title = htmlTitle
		}
		htmlGenerator := html.NewGenerator(htmlOpts)

		htmlOutput, err := htmlGenerator.Generate(pkg.Model)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating HTML: %v\n", err)
			os.Exit(1)
		}
		output = htmlOutput
		extension = ".html"
	} else if outputJSON {
		// Generate JSON output
		jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
		output = string(jsonData)
		extension = ".json"
	} else {
		// Format the package as Go code
		formattedCode, err := pkg.FormatWithOptions(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting package: %v\n", err)
			os.Exit(1)
		}
		output = formattedCode
		extension = ".go"
	}

	writeOutput(output, outFile, outDir, pkg.Name()+extension)
}

// processBatchDirectories handles processing multiple directories in batch mode
func processBatchDirectories(batchDirs string, opts *tree.Options, outDir string, outputJSON, outputHTML bool, htmlTitle string, analyzeReceivers, extractInterfaces bool) {
	directories := strings.Split(batchDirs, ",")
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

		// Special receiver analysis mode
		if analyzeReceivers || extractInterfaces {
			processReceiverAnalysis(pkg, "", outDir, extractInterfaces)
			continue
		}

		var output string
		var extension string

		// Process according to format selection
		if outputHTML {
			// Generate HTML output
			htmlOpts := html.DefaultOptions()
			if htmlTitle != "" {
				htmlOpts.Title = htmlTitle
			}
			htmlGenerator := html.NewGenerator(htmlOpts)

			htmlOutput, err := htmlGenerator.Generate(pkg.Model)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating HTML for %s: %v\n", dir, err)
				continue
			}
			output = htmlOutput
			extension = ".html"
		} else if outputJSON {
			// Generate JSON output
			jsonData, err := json.MarshalIndent(pkg.Model, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating JSON for %s: %v\n", dir, err)
				continue
			}
			output = string(jsonData)
			extension = ".json"
		} else {
			// Format as Go code
			formattedCode, err := pkg.FormatWithOptions(opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting package for %s: %v\n", dir, err)
				continue
			}
			output = formattedCode
			extension = ".go"
		}

		if outDir != "" {
			outputPath := filepath.Join(outDir, pkg.Name()+extension)
			if err := writeToFile(output, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outputPath, err)
				continue
			}
		} else {
			// Without an output directory, print to stdout with separator
			fmt.Printf("--- Package: %s ---\n", pkg.Name())
			fmt.Println(output)
			fmt.Println()
		}
	}
}

// processReceiverAnalysis handles method receiver analysis
func processReceiverAnalysis(pkg *tree.Package, outFile, outDir string, extractInterfaces bool) {
	// Create analyzer
	analyzer := interfaceanalysis.NewAnalyzer()

	// Analyze receivers
	analysis := analyzer.AnalyzeReceivers(pkg.Model)

	// Create summary
	summary := analyzer.CreateSummary(analysis)

	// Generate the output
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Method Receiver Analysis for %s\n", pkg.Name()))
	builder.WriteString("======================================\n\n")

	// Add summary statistics
	builder.WriteString("Summary:\n")
	builder.WriteString(fmt.Sprintf("- Total methods: %d\n", summary.TotalMethods))
	builder.WriteString(fmt.Sprintf("- Unique receiver types: %d\n", summary.TotalReceiverTypes))
	builder.WriteString(fmt.Sprintf("- Pointer receivers: %d\n", summary.PointerReceivers))
	builder.WriteString(fmt.Sprintf("- Value receivers: %d\n", summary.ValueReceivers))
	builder.WriteString("\n")

	// Add receiver groups
	builder.WriteString("Receiver Groups:\n")
	for receiverType, group := range analysis.Groups {
		builder.WriteString(fmt.Sprintf("\n[%s] (%d methods)\n", receiverType, len(group.Methods)))
		for _, method := range group.Methods {
			signature := method.Signature
			if signature == "" {
				signature = "()"
			}
			builder.WriteString(fmt.Sprintf("- %s%s\n", method.Name, signature))
		}
	}

	// Extract potential interfaces if requested
	if extractInterfaces {
		interfaces := analyzer.ExtractInterfaces(analysis)

		if len(interfaces) > 0 {
			builder.WriteString("\nPotential Interfaces:\n")
			builder.WriteString("====================\n\n")

			for _, intf := range interfaces {
				builder.WriteString(analyzer.GenerateInterfaceCode(intf))
				builder.WriteString("\n\n")
			}
		} else {
			builder.WriteString("\nNo potential interfaces found.\n")
		}
	}

	// Write the output
	outputPath := ""
	if outDir != "" {
		if extractInterfaces {
			outputPath = filepath.Join(outDir, pkg.Name()+"_interfaces.go")
		} else {
			outputPath = filepath.Join(outDir, pkg.Name()+"_receivers.txt")
		}
	}

	writeOutput(builder.String(), outFile, outDir, outputPath)
}

// writeOutput handles writing output to a file or stdout
func writeOutput(output, outFile, outDir, defaultPath string) {
	// Determine where to write the output
	if outFile != "" {
		if err := writeToFile(output, outFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outFile, err)
			os.Exit(1)
		}
	} else if outDir != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outDir, 0750); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", outDir, err)
			os.Exit(1)
		}

		// Use default path
		outputPath := filepath.Join(outDir, filepath.Base(defaultPath))

		if err := writeToFile(output, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outputPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Successfully wrote output to %s\n", outputPath)
	} else {
		// Write to stdout
		fmt.Print(output)
	}
}

// writeToFile writes content to a file, creating parent directories if needed
func writeToFile(content, path string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
