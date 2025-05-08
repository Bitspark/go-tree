package typesys

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Package represents a Go package with full type information.
type Package struct {
	// Basic information
	Module     *Module            // Parent module
	Name       string             // Package name (not import path)
	ImportPath string             // Import path
	Dir        string             // Package directory
	Files      map[string]*File   // Files by path
	Symbols    map[string]*Symbol // Symbols by ID

	// Cross-references
	Imports  map[string]*Import // Imports by import path
	Exported map[string]*Symbol // Exported symbols by name

	// Type information
	TypesPackage *types.Package // Go's type representation
	TypesInfo    *types.Info    // Type information
	astPackage   *ast.Package   // AST package
}

// Import represents an import in a Go file
type Import struct {
	Path string    // Import path
	Name string    // Local name (may be "")
	File *File     // Containing file
	Pos  token.Pos // Import position
	End  token.Pos // End position
}

// NewPackage creates a new package with the given name and import path.
func NewPackage(mod *Module, name, importPath string) *Package {
	return &Package{
		Module:     mod,
		Name:       name,
		ImportPath: importPath,
		Files:      make(map[string]*File),
		Symbols:    make(map[string]*Symbol),
		Imports:    make(map[string]*Import),
		Exported:   make(map[string]*Symbol),
	}
}

// SymbolByName finds symbols by name, optionally filtering by kind.
// If name is a prefix (not an exact match), it returns all symbols that start with that prefix.
func (p *Package) SymbolByName(name string, kinds ...SymbolKind) []*Symbol {
	var result []*Symbol
	for _, sym := range p.Symbols {
		// Check if the symbol name starts with the given name (prefix matching)
		if sym.Name == name || (len(name) < len(sym.Name) && sym.Name[:len(name)] == name) {
			if len(kinds) == 0 || containsKind(kinds, sym.Kind) {
				result = append(result, sym)
			}
		}
	}
	return result
}

// SymbolByID returns a symbol by its ID.
func (p *Package) SymbolByID(id string) *Symbol {
	return p.Symbols[id]
}

// UpdateFiles processes only changed files in the package.
func (p *Package) UpdateFiles(files []string) error {
	// This is a placeholder that will be implemented when we have file.go and loader.go
	return nil
}

// AddSymbol adds a symbol to the package.
func (p *Package) AddSymbol(sym *Symbol) {
	// Set the package reference on the symbol itself
	sym.Package = p

	// Add symbol to the package's maps
	p.Symbols[sym.ID] = sym
	if sym.Exported {
		p.Exported[sym.Name] = sym
	}
}

// RemoveSymbol removes a symbol from the package.
func (p *Package) RemoveSymbol(sym *Symbol) {
	delete(p.Symbols, sym.ID)
	if sym.Exported {
		delete(p.Exported, sym.Name)
	}
}

// AddFile adds a file to the package.
func (p *Package) AddFile(file *File) {
	p.Files[file.Path] = file
	file.Package = p
}

// RemoveFile removes a file from the package.
func (p *Package) RemoveFile(path string) {
	delete(p.Files, path)
}

// Helper function to check if a slice contains a kind
func containsKind(kinds []SymbolKind, kind SymbolKind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
