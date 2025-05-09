package json

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"bitspark.dev/go-tree/pkg/typesys"
)

// VisualizationOptions provides options for JSON visualization
type VisualizationOptions struct {
	IncludeTypeAnnotations bool
	IncludePrivate         bool
	IncludeTests           bool
	DetailLevel            int
	PrettyPrint            bool
}

// JSONVisualizer creates JSON visualizations of a Go module with type information
type JSONVisualizer struct{}

// NewJSONVisualizer creates a new JSON visualizer
func NewJSONVisualizer() *JSONVisualizer {
	return &JSONVisualizer{}
}

// Visualize creates a JSON visualization of the module
func (v *JSONVisualizer) Visualize(module *typesys.Module, opts *VisualizationOptions) ([]byte, error) {
	if opts == nil {
		opts = &VisualizationOptions{
			DetailLevel: 3,
			PrettyPrint: true,
		}
	}

	// Create a modular view of the module with desired level of detail
	moduleView := createModuleView(module, opts)

	// Marshal to JSON
	var result []byte
	var err error

	if opts.PrettyPrint {
		result, err = json.MarshalIndent(moduleView, "", "  ")
	} else {
		result, err = json.Marshal(moduleView)
	}

	return result, err
}

// Format returns the output format name
func (v *JSONVisualizer) Format() string {
	return "json"
}

// SupportsTypeAnnotations indicates if this visualizer can show type info
func (v *JSONVisualizer) SupportsTypeAnnotations() bool {
	return true
}

// ModuleView is a simplified view of a module for serialization
type ModuleView struct {
	Path      string                 `json:"path"`
	GoVersion string                 `json:"goVersion"`
	Dir       string                 `json:"dir,omitempty"`
	Packages  map[string]PackageView `json:"packages"`
}

// PackageView is a simplified view of a package for serialization
type PackageView struct {
	Name       string                `json:"name"`
	ImportPath string                `json:"importPath"`
	Dir        string                `json:"dir,omitempty"`
	Files      []string              `json:"files,omitempty"`
	Symbols    map[string]SymbolView `json:"symbols,omitempty"`
}

// SymbolView is a simplified view of a symbol for serialization
type SymbolView struct {
	Name        string       `json:"name"`
	Kind        string       `json:"kind"`
	Exported    bool         `json:"exported"`
	TypeInfo    string       `json:"typeInfo,omitempty"`
	Position    string       `json:"position,omitempty"`
	Fields      []SymbolView `json:"fields,omitempty"`
	Methods     []SymbolView `json:"methods,omitempty"`
	ParentName  string       `json:"parentName,omitempty"`
	PackageName string       `json:"packageName,omitempty"`
}

// createModuleView creates a simplified view of the module for JSON serialization
func createModuleView(module *typesys.Module, opts *VisualizationOptions) ModuleView {
	view := ModuleView{
		Path:      module.Path,
		GoVersion: module.GoVersion,
		Dir:       module.Dir,
		Packages:  make(map[string]PackageView),
	}

	// Add packages
	for importPath, pkg := range module.Packages {
		// Skip test packages if not requested
		if !opts.IncludeTests && isTestPackage(pkg) {
			continue
		}

		// Create package view
		packageView := PackageView{
			Name:       pkg.Name,
			ImportPath: pkg.ImportPath,
			Dir:        pkg.Dir,
			Files:      []string{},
			Symbols:    make(map[string]SymbolView),
		}

		// Add file names
		for _, file := range pkg.Files {
			// Skip test files if not requested
			if !opts.IncludeTests && file.IsTest {
				continue
			}

			packageView.Files = append(packageView.Files, file.Name)

			// Add symbols from this file
			for _, symbol := range file.Symbols {
				// Skip private symbols if not requested
				if !opts.IncludePrivate && !symbol.Exported {
					continue
				}

				// Create symbol view
				symbolView := createSymbolView(symbol, opts)

				// Add to the package's symbols
				packageView.Symbols[symbol.Name] = symbolView
			}
		}

		// Add to the module's packages
		view.Packages[importPath] = packageView
	}

	return view
}

// createSymbolView creates a simplified view of a symbol for JSON serialization
func createSymbolView(symbol *typesys.Symbol, opts *VisualizationOptions) SymbolView {
	view := SymbolView{
		Name:     symbol.Name,
		Kind:     symbol.Kind.String(),
		Exported: symbol.Exported,
	}

	// Add type information if requested
	if opts.IncludeTypeAnnotations && symbol.TypeInfo != nil {
		view.TypeInfo = symbol.TypeInfo.String()
	}

	// Add position information if available
	if symbol.File != nil {
		filename := symbol.File.Name
		// In a real implementation, we would get the line number from the symbol's position
		view.Position = fmt.Sprintf("%s", filename)
	}

	// Add parent and package info
	if symbol.Parent != nil {
		view.ParentName = symbol.Parent.Name
	}

	if symbol.Package != nil {
		view.PackageName = symbol.Package.Name
	}

	// For detailed views, add struct fields if available
	if opts.DetailLevel >= 3 && symbol.Kind == typesys.KindStruct {
		// Find child symbols that are fields of this struct
		fieldSymbols := getStructFields(symbol)
		if len(fieldSymbols) > 0 {
			view.Fields = make([]SymbolView, 0, len(fieldSymbols))
			for _, field := range fieldSymbols {
				// Skip private fields if not requested
				if !opts.IncludePrivate && !field.Exported {
					continue
				}

				fieldView := createSymbolView(field, opts)
				view.Fields = append(view.Fields, fieldView)
			}
		}

		// Find methods associated with this type
		methodSymbols := getTypeMethods(symbol)
		if len(methodSymbols) > 0 {
			view.Methods = make([]SymbolView, 0, len(methodSymbols))
			for _, method := range methodSymbols {
				// Skip private methods if not requested
				if !opts.IncludePrivate && !method.Exported {
					continue
				}

				methodView := createSymbolView(method, opts)
				view.Methods = append(view.Methods, methodView)
			}
		}
	}

	return view
}

// isTestPackage determines if a package is a test package
func isTestPackage(pkg *typesys.Package) bool {
	// Check if package name ends with _test
	if pkg.Name == "main_test" || pkg.Name == "test" {
		return true
	}

	// Check if package is in a test directory
	if filepath.Base(pkg.Dir) == "testdata" {
		return true
	}

	// Check if the package only contains test files
	testFilesOnly := true
	for _, file := range pkg.Files {
		if !file.IsTest {
			testFilesOnly = false
			break
		}
	}

	return testFilesOnly
}

// getStructFields returns the field symbols for a struct type
func getStructFields(symbol *typesys.Symbol) []*typesys.Symbol {
	if symbol == nil || symbol.Kind != typesys.KindStruct {
		return nil
	}

	// In a real implementation, we would use proper struct information
	// from the type system. For now, we'll use a simple approach to find
	// child symbols that are fields of this struct based on parent reference.
	var fields []*typesys.Symbol

	if symbol.File != nil {
		for _, s := range symbol.File.Symbols {
			if s.Parent == symbol && s.Kind == typesys.KindField {
				fields = append(fields, s)
			}
		}
	}

	return fields
}

// getTypeMethods returns the method symbols for a type
func getTypeMethods(symbol *typesys.Symbol) []*typesys.Symbol {
	if symbol == nil {
		return nil
	}

	// In a real implementation, we would use proper method information
	// from the type system. For now, we'll use a simple approach to find
	// symbols that are methods of this type.
	var methods []*typesys.Symbol

	if symbol.Package != nil {
		for _, file := range symbol.Package.Files {
			for _, s := range file.Symbols {
				if s.Kind == typesys.KindMethod && s.Parent == symbol {
					methods = append(methods, s)
				}
			}
		}
	}

	return methods
}
