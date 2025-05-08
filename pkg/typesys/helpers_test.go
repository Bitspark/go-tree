package typesys

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"
)

func TestPathHelpers(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string // This will be compared after platform-specific normalization
		isAbs    bool
	}{
		{
			name:     "clean relative path with slash",
			path:     "pkg/typesys/",
			expected: "pkg/typesys",
			isAbs:    false,
		},
		{
			name:     "path with dot segments",
			path:     "pkg/./typesys/../typesys",
			expected: "pkg/typesys",
			isAbs:    false,
		},
		{
			name:     "duplicate slashes",
			path:     "pkg//typesys",
			expected: "pkg/typesys",
			isAbs:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePath(tc.path)

			// Convert both the result and expected value to use platform-specific separators
			expectedWithOSSep := filepath.FromSlash(tc.expected)

			if result != expectedWithOSSep {
				t.Errorf("normalizePath(%q) = %q, want %q", tc.path, result, expectedWithOSSep)
			}

			absPath := ensureAbsolutePath(tc.path)
			isAbs := filepath.IsAbs(absPath)
			if isAbs != true {
				t.Errorf("ensureAbsolutePath(%q) should return absolute path, got %q", tc.path, absPath)
			}
		})
	}
}

func TestSymbolHelpers(t *testing.T) {
	// Create a test file set
	fset := token.NewFileSet()

	// Create a test package
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")

	// Add types info to package
	pkg.TypesInfo = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	// Create a test file
	file := NewFile("/test/module/file.go", pkg)
	file.FileSet = fset

	// Test createSymbol
	sym := createSymbol(pkg, file, "TestSymbol", KindFunction, token.Pos(10), token.Pos(20), nil)

	if sym.Name != "TestSymbol" {
		t.Errorf("Symbol name = %q, want %q", sym.Name, "TestSymbol")
	}

	if sym.Kind != KindFunction {
		t.Errorf("Symbol kind = %v, want %v", sym.Kind, KindFunction)
	}

	if sym.Package != pkg {
		t.Errorf("Symbol package not set correctly")
	}

	if sym.File != file {
		t.Errorf("Symbol file not set correctly")
	}
}

func TestSymbolFiltering(t *testing.T) {
	tests := []struct {
		name       string
		opts       LoadOptions
		symbolName string
		expected   bool
	}{
		{
			name:       "Include private with ExportedSymbol",
			opts:       LoadOptions{IncludePrivate: true},
			symbolName: "ExportedSymbol",
			expected:   true,
		},
		{
			name:       "Include private with unexportedSymbol",
			opts:       LoadOptions{IncludePrivate: true},
			symbolName: "unexportedSymbol",
			expected:   true,
		},
		{
			name:       "Exclude private with ExportedSymbol",
			opts:       LoadOptions{IncludePrivate: false},
			symbolName: "ExportedSymbol",
			expected:   true,
		},
		{
			name:       "Exclude private with unexportedSymbol",
			opts:       LoadOptions{IncludePrivate: false},
			symbolName: "unexportedSymbol",
			expected:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldIncludeSymbol(tc.symbolName, &tc.opts)
			if result != tc.expected {
				t.Errorf("shouldIncludeSymbol(%q, %v) = %v, want %v",
					tc.symbolName, tc.opts.IncludePrivate, result, tc.expected)
			}
		})
	}
}

func TestLoggingHelpers(t *testing.T) {
	// These functions just print if the trace flag is set, so we're just
	// testing that they don't panic

	// With nil options
	tracef(nil, "This is a trace message")
	warnf(nil, "This is a warning message")
	errorf(nil, "This is an error message")

	// With trace disabled
	opts := &LoadOptions{Trace: false}
	tracef(opts, "This is a trace message")
	warnf(opts, "This is a warning message")
	errorf(opts, "This is an error message")

	// With trace enabled (will print to stdout but we're just checking no panic)
	opts = &LoadOptions{Trace: true}
	tracef(opts, "This is a trace message with %s", "formatting")
	warnf(opts, "This is a warning message with %s", "formatting")
	errorf(opts, "This is an error message with %s", "formatting")
}

func TestProcessSafely(t *testing.T) {
	// Create a test file
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)

	// Test successful function
	err := processSafely(file, func() error {
		return nil
	}, nil)

	if err != nil {
		t.Errorf("processSafely with successful function returned error: %v", err)
	}

	// Test function that returns error
	expectedErr := fmt.Errorf("test error")
	err = processSafely(file, func() error {
		return expectedErr
	}, nil)

	if err != expectedErr {
		t.Errorf("processSafely with error function returned %v, want %v", err, expectedErr)
	}

	// Test function that panics
	err = processSafely(file, func() error {
		panic("test panic")
		return nil
	}, nil)

	if err == nil {
		t.Errorf("processSafely with panicking function should return error, got nil")
	}
}
