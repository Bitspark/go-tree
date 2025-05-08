package typesys

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
)

// createSymbol centralizes the common logic for creating and initializing symbols
func createSymbol(pkg *Package, file *File, name string, kind SymbolKind, pos, end token.Pos, parent *Symbol) *Symbol {
	sym := NewSymbol(name, kind)
	sym.Pos = pos
	sym.End = end
	sym.File = file
	sym.Package = pkg
	sym.Parent = parent

	// Get position info
	if posInfo := file.GetPositionInfo(pos, end); posInfo != nil {
		sym.AddDefinition(file.Path, pos, posInfo.LineStart, posInfo.ColumnStart)
	}

	return sym
}

// extractTypeInfo centralizes getting type information from the type checker
func extractTypeInfo(pkg *Package, name *ast.Ident, expr ast.Expr) (types.Object, types.Type) {
	if name != nil && pkg.TypesInfo != nil {
		if obj := pkg.TypesInfo.ObjectOf(name); obj != nil {
			return obj, obj.Type()
		}
	}

	if expr != nil && pkg.TypesInfo != nil {
		return nil, pkg.TypesInfo.TypeOf(expr)
	}

	return nil, nil
}

// shouldIncludeSymbol determines if a symbol should be included based on options
func shouldIncludeSymbol(name string, opts *LoadOptions) bool {
	return opts.IncludePrivate || ast.IsExported(name)
}

// processSafely executes a function with panic recovery
func processSafely(file *File, fn func() error, opts *LoadOptions) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("Panic when processing file %s: %v", file.Path, r)
				err = fmt.Errorf(errMsg)
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
func tracef(opts *LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf(format, args...)
	}
}

// warnf logs a warning message if tracing is enabled
func warnf(opts *LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf("WARNING: "+format, args...)
	}
}

// errorf logs an error message if tracing is enabled
func errorf(opts *LoadOptions, format string, args ...interface{}) {
	if opts != nil && opts.Trace {
		fmt.Printf("ERROR: "+format, args...)
	}
}
