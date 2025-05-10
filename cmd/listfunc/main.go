package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/service"
)

func main() {
	// Parse command line arguments
	flag.Parse()

	// Get module path from argument, default to current directory
	modulePath := "."
	if flag.NArg() > 0 {
		modulePath = flag.Arg(0)
	}

	// Convert to absolute path for better error messages
	absPath, err := filepath.Abs(modulePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to absolute path: %v\n", err)
		os.Exit(1)
	}

	// Check if the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Directory does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// Check if go.mod exists
	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No go.mod file found in: %s\n", absPath)
		os.Exit(1)
	}

	// Create service configuration
	config := &service.Config{
		ModuleDir:       absPath,
		IncludeTests:    true,
		WithDeps:        false,
		DependencyDepth: 0,
		DownloadMissing: false,
		Verbose:         false,
	}

	// Create service to load the module
	svc, err := service.NewService(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading module: %v\n", err)
		os.Exit(1)
	}

	// Get the main module
	mainModule := svc.GetMainModule()
	if mainModule == nil {
		fmt.Fprintf(os.Stderr, "Failed to load main module\n")
		os.Exit(1)
	}

	// Print module information
	fmt.Printf("Module: %s\n", mainModule.Path)
	fmt.Printf("Directory: %s\n", mainModule.Dir)
	fmt.Printf("\nFunctions:\n")

	// Collect all functions
	var functions []*typesys.Symbol

	// Iterate through all packages in the module
	for _, pkg := range mainModule.Packages {
		// Iterate through all symbols in the package
		for _, symbol := range pkg.Symbols {
			// Check if the symbol is a function
			if symbol.Kind == typesys.KindFunction {
				functions = append(functions, symbol)
			}
		}
	}

	// Sort functions by package and name for nicer output
	sort.Slice(functions, func(i, j int) bool {
		if functions[i].Package.ImportPath != functions[j].Package.ImportPath {
			return functions[i].Package.ImportPath < functions[j].Package.ImportPath
		}
		return functions[i].Name < functions[j].Name
	})

	// Print functions grouped by package
	lastPackage := ""
	for _, fn := range functions {
		// Print package header when changing packages
		if fn.Package.ImportPath != lastPackage {
			fmt.Printf("\n[%s]\n", fn.Package.ImportPath)
			lastPackage = fn.Package.ImportPath
		}

		// Print function name and signature if available
		if fn.TypeInfo != nil {
			fmt.Printf("  %s%s\n", fn.Name, fn.TypeInfo)
		} else {
			fmt.Printf("  %s\n", fn.Name)
		}
	}

	// Print summary
	fmt.Printf("\nTotal functions: %d\n", len(functions))
}
