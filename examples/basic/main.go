// Example usage of the go-tree package
package main

import (
	"fmt"
	"os"

	"bitspark.dev/go-tree/tree"
)

func main() {
	// Check if a directory was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	// Parse the directory
	pkg, err := tree.Parse(os.Args[1])
	if err != nil {
		fmt.Printf("Error parsing package: %v\n", err)
		os.Exit(1)
	}

	// Print package information
	fmt.Printf("Package: %s\n", pkg.Name())

	fmt.Println("\nImports:")
	for _, imp := range pkg.Imports() {
		fmt.Printf("  - %s\n", imp)
	}

	fmt.Println("\nFunctions:")
	for _, fn := range pkg.FunctionNames() {
		f := pkg.GetFunction(fn)
		if f.IsMethod() {
			fmt.Printf("  - method %s on %s\n", fn, f.ReceiverType())
		} else {
			fmt.Printf("  - func %s\n", fn)
		}
	}

	fmt.Println("\nTypes:")
	for _, tn := range pkg.TypeNames() {
		t := pkg.GetType(tn)
		if t.IsStruct() {
			fmt.Printf("  - struct %s\n", tn)
		} else if t.IsInterface() {
			fmt.Printf("  - interface %s\n", tn)
		} else {
			fmt.Printf("  - %s %s\n", t.Kind(), tn)
		}
	}

	fmt.Println("\nConstants:")
	for _, c := range pkg.ConstantNames() {
		fmt.Printf("  - %s\n", c)
	}

	fmt.Println("\nVariables:")
	for _, v := range pkg.VariableNames() {
		fmt.Printf("  - %s\n", v)
	}

	// Format the package to a single file
	formatted, err := pkg.Format()
	if err != nil {
		fmt.Printf("Error formatting package: %v\n", err)
		os.Exit(1)
	}

	// Write result to file
	outFile := "formatted.go"
	if err := os.WriteFile(outFile, []byte(formatted), 0644); err != nil {
		fmt.Printf("Error writing to %s: %v\n", outFile, err)
		os.Exit(1)
	}

	fmt.Printf("\nPackage formatted and written to %s\n", outFile)
}
