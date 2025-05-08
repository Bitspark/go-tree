package index

import (
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/typesys"
)

// IndexingOptions provides configuration options for the indexer.
type IndexingOptions struct {
	IncludeTests       bool // Whether to include test files in the index
	IncludePrivate     bool // Whether to include private (unexported) symbols
	IncrementalUpdates bool // Whether to use incremental updates when possible
}

// Indexer provides high-level indexing functionality for Go code.
// It wraps an Index and provides additional methods for searching and navigating.
type Indexer struct {
	Index   *Index
	Module  *typesys.Module
	Options IndexingOptions
}

// NewIndexer creates a new indexer for the given module.
func NewIndexer(mod *typesys.Module, options IndexingOptions) *Indexer {
	return &Indexer{
		Index:   NewIndex(mod),
		Module:  mod,
		Options: options,
	}
}

// BuildIndex builds the initial index for the module.
func (idx *Indexer) BuildIndex() error {
	// Build the index
	return idx.Index.Build()
}

// UpdateIndex updates the index for the changed files.
func (idx *Indexer) UpdateIndex(changedFiles []string) error {
	if len(changedFiles) == 0 {
		return nil
	}

	// If incremental updates are disabled, rebuild the whole index
	if !idx.Options.IncrementalUpdates {
		return idx.Index.Build()
	}

	// Find all affected files (files that depend on the changed files)
	affectedFiles := idx.Module.FindAffectedFiles(changedFiles)

	// Update the module first
	if err := idx.Module.UpdateChangedFiles(affectedFiles); err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}

	// Update the index
	return idx.Index.Update(affectedFiles)
}

// FindUsages finds all usages (references) of a symbol.
func (idx *Indexer) FindUsages(symbol *typesys.Symbol) []*typesys.Reference {
	return idx.Index.FindReferences(symbol)
}

// FindImplementations finds all implementations of an interface.
func (idx *Indexer) FindImplementations(interfaceSymbol *typesys.Symbol) []*typesys.Symbol {
	return idx.Index.FindImplementations(interfaceSymbol)
}

// FindSymbolByNameAndType searches for symbols matching a name and optional type kind.
func (idx *Indexer) FindSymbolByNameAndType(name string, kinds ...typesys.SymbolKind) []*typesys.Symbol {
	if len(kinds) == 0 {
		return idx.Index.FindSymbolsByName(name)
	}

	var results []*typesys.Symbol
	for _, sym := range idx.Index.FindSymbolsByName(name) {
		for _, kind := range kinds {
			if sym.Kind == kind {
				results = append(results, sym)
				break
			}
		}
	}
	return results
}

// FindMethodsOfType finds all methods for a given type symbol.
func (idx *Indexer) FindMethodsOfType(typeSymbol *typesys.Symbol) []*typesys.Symbol {
	return idx.Index.FindMethods(typeSymbol.Name)
}

// FindSymbolAtPosition finds the symbol at the given position in a file.
func (idx *Indexer) FindSymbolAtPosition(filePath string, line, column int) *typesys.Symbol {
	file := idx.Module.FileByPath(filePath)
	if file == nil {
		return nil
	}

	// Check all symbols in the file
	for _, sym := range idx.Index.FindSymbolsInFile(filePath) {
		pos := sym.GetPosition()
		if pos == nil {
			continue
		}

		// Check if position is within symbol bounds
		if (pos.LineStart < line || (pos.LineStart == line && pos.ColumnStart <= column)) &&
			(pos.LineEnd > line || (pos.LineEnd == line && pos.ColumnEnd >= column)) {
			return sym
		}
	}

	return nil
}

// FindReferenceAtPosition finds the reference at the given position in a file.
func (idx *Indexer) FindReferenceAtPosition(filePath string, line, column int) *typesys.Reference {
	file := idx.Module.FileByPath(filePath)
	if file == nil {
		return nil
	}

	// Check all references in the file
	for _, ref := range idx.Index.FindReferencesInFile(filePath) {
		pos := ref.GetPosition()
		if pos == nil {
			continue
		}

		// Check if position is within reference bounds
		if (pos.LineStart < line || (pos.LineStart == line && pos.ColumnStart <= column)) &&
			(pos.LineEnd > line || (pos.LineEnd == line && pos.ColumnEnd >= column)) {
			return ref
		}
	}

	return nil
}

// Search performs a general search across the index.
func (idx *Indexer) Search(query string) []*typesys.Symbol {
	var results []*typesys.Symbol

	// Try exact name match first
	exactMatches := idx.Index.FindSymbolsByName(query)
	if len(exactMatches) > 0 {
		results = append(results, exactMatches...)
	}

	// Try fuzzy matching if no exact matches or requested
	if len(results) == 0 {
		// Search for partial name matches
		for name, symbols := range idx.Index.symbolsByName {
			if strings.Contains(name, query) {
				results = append(results, symbols...)
			}
		}
	}

	return results
}

// FindAllFunctions finds all functions matching the given name pattern.
func (idx *Indexer) FindAllFunctions(namePattern string) []*typesys.Symbol {
	var results []*typesys.Symbol

	functions := idx.Index.FindSymbolsByKind(typesys.KindFunction)
	for _, fn := range functions {
		if strings.Contains(fn.Name, namePattern) {
			results = append(results, fn)
		}
	}

	return results
}

// FindAllTypes finds all types matching the given name pattern.
func (idx *Indexer) FindAllTypes(namePattern string) []*typesys.Symbol {
	var results []*typesys.Symbol

	// Include all type-like kinds
	typeKinds := []typesys.SymbolKind{
		typesys.KindType,
		typesys.KindStruct,
		typesys.KindInterface,
	}

	for _, kind := range typeKinds {
		types := idx.Index.FindSymbolsByKind(kind)
		for _, t := range types {
			if strings.Contains(t.Name, namePattern) {
				results = append(results, t)
			}
		}
	}

	return results
}

// GetFileSymbols returns all symbols in a file, organized by type.
func (idx *Indexer) GetFileSymbols(filePath string) map[typesys.SymbolKind][]*typesys.Symbol {
	result := make(map[typesys.SymbolKind][]*typesys.Symbol)

	symbols := idx.Index.FindSymbolsInFile(filePath)
	for _, sym := range symbols {
		result[sym.Kind] = append(result[sym.Kind], sym)
	}

	return result
}

// GetFileStructure returns a structured representation of the file contents.
// This can be used for displaying a file outline in an IDE.
func (idx *Indexer) GetFileStructure(filePath string) []*SymbolNode {
	symbols := idx.Index.FindSymbolsInFile(filePath)
	return buildSymbolTree(symbols)
}

// SymbolNode represents a node in the symbol tree for file structure.
type SymbolNode struct {
	Symbol   *typesys.Symbol
	Children []*SymbolNode
}

// buildSymbolTree organizes symbols into a tree structure.
func buildSymbolTree(symbols []*typesys.Symbol) []*SymbolNode {
	// First pass: create nodes for all symbols
	nodesBySymbol := make(map[*typesys.Symbol]*SymbolNode)
	for _, sym := range symbols {
		nodesBySymbol[sym] = &SymbolNode{
			Symbol:   sym,
			Children: make([]*SymbolNode, 0),
		}
	}

	// Second pass: build the tree
	var roots []*SymbolNode
	for sym, node := range nodesBySymbol {
		if sym.Parent == nil {
			// This is a root node
			roots = append(roots, node)
		} else {
			// This is a child node
			if parentNode, ok := nodesBySymbol[sym.Parent]; ok {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// Parent isn't in our map, treat as root
				roots = append(roots, node)
			}
		}
	}

	return roots
}
