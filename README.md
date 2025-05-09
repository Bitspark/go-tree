[![CI](https://github.com/Bitspark/go-tree/actions/workflows/main-pipeline.yml/badge.svg)](https://github.com/Bitspark/go-tree/actions/workflows/main-pipeline.yml)
[![codecov](https://codecov.io/gh/Bitspark/go-tree/branch/main/graph/badge.svg?token=CRRt8eRJBz)](https://app.codecov.io/gh/Bitspark/go-tree/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bitspark/go-tree)](https://goreportcard.com/report/github.com/Bitspark/go-tree)
![Go Version](https://img.shields.io/github/go-mod/go-version/Bitspark/go-tree)
[![Go Reference](https://pkg.go.dev/badge/github.com/Bitspark/go-tree.svg)](https://pkg.go.dev/bitspark.dev/go-tree)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/Bitspark/go-tree)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# [![](assets/go-tree@h80px.png)](https://bitspark.dev/go-tree) Go-Tree

> Library and command-line tool for analyzing and modifying Go code.

### **Documentation**: [bitspark.dev/go-tree](https://bitspark.dev/go-tree)

## Features

- Parse and format Go from source files
- Extract structured package data (functions, types, constants, variables)
- Generate JSON representations fully capturing the codebase
- Configurable and available as CLI and Go library

## Installation

```bash
# Install the library
go get bitspark.dev/go-tree

# Install the CLI tool
go install bitspark.dev/go-tree/cmd/gotree@latest
```

## Library Usage

```go
package main

import (
	"fmt"
	"bitspark.dev/go-tree/tree"
)

func main() {
	// Parse a Go package
	pkg, err := tree.Parse("./path/to/package")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	// Get package info
	fmt.Printf("Package: %s\n", pkg.Name())
	fmt.Printf("Functions: %v\n", pkg.FunctionNames())
	fmt.Printf("Types: %v\n", pkg.TypeNames())
	
	// Format package to a single file
	output, err := pkg.Format()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Println(output)
}
```

## CLI Usage

```bash
# Parse a package and output to stdout
gotree -src ./path/to/package

# Parse and save to file with options
gotree -src ./path/to/package -out-file output.go -include-tests -preserve-formatting

# Generate JSON documentation
gotree -src ./path/to/package -json -out-dir ./docs/json

# Process multiple packages in batch mode
gotree -batch "/path/to/pkg1,/path/to/pkg2" -json -out-dir ./docs/json
```

## License

MIT