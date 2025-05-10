package loader

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// createSymbol centralizes the common logic for creating and initializing symbols
func createSymbol(pkg *typesys.Package, file *typesys.File, name string, kind typesys.SymbolKind, pos, end token.Pos, parent *typesys.Symbol) *typesys.Symbol {
	sym := typesys.NewSymbol(name, kind)
	sym.Pos = pos
	sym.End = end

	// Verify we're using the correct file for this symbol based on its position
	if pkg != nil && pkg.Module != nil && pkg.Module.FileSet != nil && pos.IsValid() {
		posInfo := pkg.Module.FileSet.Position(pos)
		if posInfo.IsValid() && posInfo.Filename != "" {
			posFilename := filepath.Clean(posInfo.Filename)
			fileFilename := filepath.Clean(file.Path)

			// If position's filename differs from provided file, try to find correct file
			if posFilename != fileFilename {
				// Check if it's a test file that was mistakenly added to a non-test file
				for _, pkgFile := range pkg.Files {
					cleanPath := filepath.Clean(pkgFile.Path)
					if cleanPath == posFilename {
						// Found the correct file based on position - use it instead
						file = pkgFile
						break
					}
				}
			}
		}
	}

	sym.File = file
	sym.Package = pkg
	sym.Parent = parent

	// Get position info
	if posInfo := file.GetPositionInfo(pos, end); posInfo != nil {
		sym.AddDefinition(file.Path, pos, posInfo.LineStart, posInfo.ColumnStart)
	}

	return sym
}

// shouldIncludeSymbol determines if a symbol should be included based on options
func shouldIncludeSymbol(name string, opts *typesys.LoadOptions) bool {
	return opts.IncludePrivate || ast.IsExported(name)
}

// processSafely executes a function with panic recovery
func processSafely(file *typesys.File, fn func() error, opts *typesys.LoadOptions) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("Panic when processing file %s: %v", file.Path, r)
				err = errors.New(errMsg)
				if opts != nil && opts.Trace {
					fmt.Printf("ERROR: %s\n", errMsg)
				}
			}
		}()
		err = fn()
	}()
	return err
}

// Path normalization helpers

// normalizePath ensures consistent path formatting
func normalizePath(path string) string {
	return filepath.Clean(path)
}

// ensureAbsolutePath makes a path absolute if it isn't already
func ensureAbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

// Logging helpers

// tracef logs a message if tracing is enabled
func tracef(opts *typesys.LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf(format, args...)
	}
}

// warnf logs a warning message if tracing is enabled
func warnf(opts *typesys.LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf("WARNING: "+format, args...)
	}
}

// errorf logs an error message if tracing is enabled
func errorf(opts *typesys.LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf("ERROR: "+format, args...)
	}
}

// Helper function to convert an expression to a string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	}
	return ""
}
