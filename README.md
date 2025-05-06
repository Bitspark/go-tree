# Go-Tree

> Library and command-line tool for parsing and formatting Go packages.

[![](assets/go-tree.png)](https://bitspark.dev/go-tree)

**Website**: [bitspark.dev/go-tree](https://bitspark.dev/go-tree)

## Features

- Parse Go packages from directories
- Extract package metadata (functions, types, constants, variables)
- Format Go packages into a single source file
- Generate JSON representation for use with static site generators
- Configurable parsing and formatting options

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
gotree -src ./path/to/package -out output.go -include-tests -preserve-formatting

# Generate JSON documentation
gotree -src ./path/to/package -json -docs-dir ./docs/json

# Process multiple packages in batch mode
gotree -batch "/path/to/pkg1,/path/to/pkg2" -json -docs-dir ./docs/json
```

### CLI Options

- `-src`: Source directory containing Go package (default: current directory)
- `-out`: Output file (default: stdout)
- `-json`: Output as JSON instead of formatted Go code
- `-docs-dir`: Output directory for documentation JSON files
- `-batch`: Comma-separated list of directories to process in batch mode
- `-include-tests`: Include test files in parsing
- `-preserve-formatting`: Preserve original formatting style
- `-skip-comments`: Skip comments during parsing
- `-package`: Custom package name for output

## Documentation Generation

This repository includes scripts to help with documentation generation:

```bash
# Using the provided script (Unix/Linux/macOS with Bash)
./scripts/generate.sh -src ./path/to/package -docs-dir ./docs/json

# Windows users can either use WSL, Git Bash, or call gotree directly
gotree -src ./path/to/package -json -docs-dir ./docs/json
```

The generated JSON contains structured documentation of your Go packages and can be used with, e.g., a static site generator.

## License

MIT
