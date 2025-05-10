package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
	"bitspark.dev/go-tree/pkg/service"
)

func main() {
	// Parse command-line flags
	modulePath := flag.String("module", "", "Path to the Go module")
	funcName := flag.String("func", "", "Fully qualified function name (e.g., 'package.Function')")
	flag.Parse()

	// Validate input
	if *modulePath == "" || *funcName == "" {
		fmt.Println("Error: Both module path and function name are required")
		fmt.Println("Usage: runfunc -module=/path/to/module -func=package.Function [args...]")
		os.Exit(1)
	}

	// Split function name into package and function parts
	parts := strings.Split(*funcName, ".")
	if len(parts) < 2 {
		fmt.Println("Error: Function name must be fully qualified (e.g., 'package.Function')")
		os.Exit(1)
	}
	pkgName := parts[0]
	funcBaseName := parts[len(parts)-1]

	// Initialize service
	fmt.Printf("Loading module from %s...\n", *modulePath)
	config := &service.Config{
		ModuleDir:    *modulePath,
		IncludeTests: false,
		WithDeps:     true,
	}

	svc, err := service.NewService(config)
	if err != nil {
		fmt.Printf("Error initializing service: %v\n", err)
		os.Exit(1)
	}

	// Get the main module
	module := svc.GetMainModule()
	if module == nil {
		fmt.Println("Error: Failed to load module")
		os.Exit(1)
	}

	fmt.Printf("Module loaded: %s\n", module.Path)

	// Find the function symbol
	fmt.Printf("Looking for function %s in package %s...\n", funcBaseName, pkgName)

	// First try to find the package
	var pkgPath string
	for path := range module.Packages {
		if strings.HasSuffix(path, "/"+pkgName) || path == pkgName {
			pkgPath = path
			break
		}
	}

	if pkgPath == "" {
		fmt.Printf("Error: Package %s not found in module\n", pkgName)
		os.Exit(1)
	}

	pkg := module.Packages[pkgPath]
	if pkg == nil {
		fmt.Printf("Error: Package %s not found in module\n", pkgName)
		os.Exit(1)
	}

	// Find the function in the package
	var funcSymbol *typesys.Symbol
	for _, symbol := range pkg.Symbols {
		if symbol.Kind == typesys.KindFunction && symbol.Name == funcBaseName {
			funcSymbol = symbol
			break
		}
	}

	if funcSymbol == nil {
		fmt.Printf("Error: Function %s not found in package %s\n", funcBaseName, pkgName)
		os.Exit(1)
	}

	fmt.Printf("Found function %s.%s\n", pkgPath, funcBaseName)

	// Create execution environment
	fmt.Println("Setting up execution environment...")
	env, err := svc.CreateEnvironment([]*typesys.Module{module}, &service.Config{
		IncludeTests: false,
	})
	if err != nil {
		fmt.Printf("Error creating execution environment: %v\n", err)
		os.Exit(1)
	}

	// Get the remaining arguments for the function
	args := flag.Args()

	// Execute the function
	fmt.Printf("Executing %s.%s()...\n", pkgPath, funcBaseName)

	// Create an executor
	executor := execute.NewGoExecutor()

	// Set the working directory to the module path in the environment
	if moduleFSPath, ok := env.GetModulePath(module.Path); ok {
		executor.WorkingDir = moduleFSPath
	}

	// Parse the remaining command-line arguments as function arguments
	// This is a simplified version; in a real application, you would need to parse
	// arguments based on the function's parameter types
	var functionArgs []interface{}
	for _, arg := range args {
		functionArgs = append(functionArgs, arg)
	}

	result, err := executor.ExecuteFunc(env, module, funcSymbol, functionArgs...)
	if err != nil {
		fmt.Printf("Error executing function: %v\n", err)
		os.Exit(1)
	}

	// Output the result
	fmt.Println("Execution successful")
	fmt.Printf("Result: %v\n", result)

	// Clean up the environment if it's temporary
	if env.IsTemporary {
		if err := env.Cleanup(); err != nil {
			fmt.Printf("Warning: Failed to clean up temporary environment: %v\n", err)
		}
	}
}
