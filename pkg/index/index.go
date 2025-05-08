// Package index provides indexing capabilities for Go code analysis.
package index

import (
	"go/token"

	"bitspark.dev/go-tree/pkg/core/module"
)

// SymbolKind represents the kind of a symbol in the index
type SymbolKind int

const (
	KindFunction SymbolKind = iota
	KindMethod
	KindType
	KindVariable
	KindConstant
	KindField
	KindParameter
	KindImport
)

// Symbol represents a single definition of a code element
type Symbol struct {
	// Basic information
	Name          string     // Symbol name
	Kind          SymbolKind // Type of symbol
	Package       string     // Package import path
	QualifiedName string     // Fully qualified name (pkg.Name)

	// Source location
	File      string    // File path where defined
	Pos       token.Pos // Start position
	End       token.Pos // End position
	LineStart int       // Line number start (1-based)
	LineEnd   int       // Line number end (1-based)

	// Additional information based on Kind
	ReceiverType string // For methods, the receiver type
	ParentType   string // For fields/methods, the parent type
	TypeName     string // For vars/consts/params, the type name
}

// Reference represents a usage of a symbol within the code
type Reference struct {
	// Target symbol information
	TargetSymbol *Symbol

	// Reference location
	File      string    // File path where referenced
	Pos       token.Pos // Start position
	End       token.Pos // End position
	LineStart int       // Line number start (1-based)
	LineEnd   int       // Line number end (1-based)

	// Context
	Context string // Optional context (e.g., inside which function)
}

// Index provides fast lookups for symbols and their references across a codebase
type Index struct {
	// Maps for definitions
	SymbolsByName map[string][]*Symbol // Symbol name -> symbols (may be multiple with same name in different pkgs)
	SymbolsByFile map[string][]*Symbol // File path -> symbols defined in that file
	SymbolsByType map[string][]*Symbol // Type name -> symbols related to that type (methods, fields)

	// Maps for references
	ReferencesBySymbol map[*Symbol][]*Reference // Symbol -> all references to it
	ReferencesByFile   map[string][]*Reference  // File path -> all references in that file

	// FileSet for position information
	FileSet *token.FileSet

	// Module being indexed
	Module *module.Module
}

// NewIndex creates a new empty index
func NewIndex(mod *module.Module) *Index {
	return &Index{
		SymbolsByName:      make(map[string][]*Symbol),
		SymbolsByFile:      make(map[string][]*Symbol),
		SymbolsByType:      make(map[string][]*Symbol),
		ReferencesBySymbol: make(map[*Symbol][]*Reference),
		ReferencesByFile:   make(map[string][]*Reference),
		FileSet:            token.NewFileSet(),
		Module:             mod,
	}
}

// AddSymbol adds a symbol to the index
func (idx *Index) AddSymbol(symbol *Symbol) {
	// Add to name index
	idx.SymbolsByName[symbol.Name] = append(idx.SymbolsByName[symbol.Name], symbol)

	// Add to file index
	idx.SymbolsByFile[symbol.File] = append(idx.SymbolsByFile[symbol.File], symbol)

	// Add to type index if it has a parent or receiver type
	if symbol.ParentType != "" {
		idx.SymbolsByType[symbol.ParentType] = append(idx.SymbolsByType[symbol.ParentType], symbol)
	} else if symbol.ReceiverType != "" {
		idx.SymbolsByType[symbol.ReceiverType] = append(idx.SymbolsByType[symbol.ReceiverType], symbol)
	}
}

// AddReference adds a reference to the index
func (idx *Index) AddReference(symbol *Symbol, ref *Reference) {
	// Add to symbol references index
	idx.ReferencesBySymbol[symbol] = append(idx.ReferencesBySymbol[symbol], ref)

	// Add to file references index
	idx.ReferencesByFile[ref.File] = append(idx.ReferencesByFile[ref.File], ref)
}

// FindReferences returns all references to a given symbol
func (idx *Index) FindReferences(symbol *Symbol) []*Reference {
	return idx.ReferencesBySymbol[symbol]
}

// FindSymbolsByName finds all symbols with the given name
func (idx *Index) FindSymbolsByName(name string) []*Symbol {
	return idx.SymbolsByName[name]
}

// FindSymbolsByFile finds all symbols defined in the given file
func (idx *Index) FindSymbolsByFile(filePath string) []*Symbol {
	return idx.SymbolsByFile[filePath]
}

// FindSymbolsForType finds all symbols related to the given type (methods, fields)
func (idx *Index) FindSymbolsForType(typeName string) []*Symbol {
	return idx.SymbolsByType[typeName]
}

// FindSymbolAtPosition finds a symbol at the given file position
func (idx *Index) FindSymbolAtPosition(filePath string, pos token.Pos) *Symbol {
	// Check all symbols defined in this file
	for _, sym := range idx.SymbolsByFile[filePath] {
		if pos >= sym.Pos && pos <= sym.End {
			return sym
		}
	}
	return nil
}

// FindReferenceAtPosition finds a reference at the given file position
func (idx *Index) FindReferenceAtPosition(filePath string, pos token.Pos) *Reference {
	// Check all references in this file
	for _, ref := range idx.ReferencesByFile[filePath] {
		if pos >= ref.Pos && pos <= ref.End {
			return ref
		}
	}
	return nil
}
