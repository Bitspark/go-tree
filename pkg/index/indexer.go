package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"

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
// If moduleReload is true, it will attempt to reload the module from disk before updating the index.
func (idx *Indexer) UpdateIndex(changedFiles []string) error {
	if len(changedFiles) == 0 {
		return nil
	}

	// If incremental updates are disabled, rebuild the whole index
	if !idx.Options.IncrementalUpdates {
		return idx.Index.Build()
	}

	// Find all affected files (files that depend on the changed files)
	affectedFiles := make([]string, 0, len(changedFiles))
	for _, file := range changedFiles {
		affectedFiles = append(affectedFiles, file)
		// We should also add files that depend on this file, but for now we'll
		// just use the changed files directly
	}

	// Reload the module content from disk for the affected files
	reloadError := idx.reloadFilesFromDisk(affectedFiles)
	if reloadError != nil {
		// If reload fails, continue with the update anyway, as partial updates are better than none
		fmt.Printf("Warning: Failed to reload files from disk: %v\n", reloadError)
	}

	// Update the index
	return idx.Index.Update(affectedFiles)
}

// reloadFilesFromDisk reloads content from disk for changed files
func (idx *Indexer) reloadFilesFromDisk(changedFiles []string) error {
	// First, collect packages that need to be updated
	packagesToUpdate := make(map[string]bool)

	for _, filePath := range changedFiles {
		// Find the package containing this file
		var foundPkg *typesys.Package
		var foundFile *typesys.File

		// Search through all packages and files to find the one matching our path
		for _, pkg := range idx.Module.Packages {
			for path, file := range pkg.Files {
				if path == filePath {
					foundPkg = pkg
					foundFile = file
					break
				}
			}
			if foundPkg != nil {
				break
			}
		}

		if foundFile == nil {
			// File not found in the module, try to use absolute path
			absolutePath, err := filepath.Abs(filePath)
			if err != nil {
				continue
			}

			// Try again with absolute path
			for _, pkg := range idx.Module.Packages {
				for path, file := range pkg.Files {
					if path == absolutePath {
						foundPkg = pkg
						foundFile = file
						break
					}
				}
				if foundPkg != nil {
					break
				}
			}

			if foundFile == nil {
				// Still not found, skip this file
				continue
			}
		}

		// We found the file's package, mark it for updating
		if foundPkg != nil {
			packagesToUpdate[foundPkg.ImportPath] = true

			// Actually reload the file content from disk
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			// Parse the file to get new AST
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, filePath, fileContent, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("failed to parse file %s: %w", filePath, err)
			}

			// Update the file's AST
			foundFile.AST = astFile

			// Clear existing symbols for this file
			symbolsToRemove := make([]*typesys.Symbol, 0)
			for _, sym := range foundPkg.Symbols {
				if sym.File == foundFile {
					symbolsToRemove = append(symbolsToRemove, sym)
				}
			}

			// Remove old symbols
			for _, sym := range symbolsToRemove {
				foundPkg.RemoveSymbol(sym)
				foundFile.RemoveSymbol(sym)
			}

			// Process the updated file to extract new symbols from the new AST
			// This is a simplified version of the processSymbols function from the loader package
			if astFile != nil {
				// Process declarations (functions, types, vars, consts)
				for _, decl := range astFile.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						// Process function declaration
						name := d.Name.Name
						if name == "" {
							continue
						}

						// Create symbol for the function/method
						kind := typesys.KindFunction
						if d.Recv != nil {
							kind = typesys.KindMethod
						}

						sym := typesys.NewSymbol(name, kind)
						sym.Pos = d.Pos()
						sym.End = d.End()
						sym.File = foundFile
						sym.Package = foundPkg

						// If it's a method, try to find the parent type
						if d.Recv != nil && len(d.Recv.List) > 0 {
							recv := d.Recv.List[0]
							recvTypeExpr := recv.Type

							// If it's a pointer, get the underlying type
							if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
								recvTypeExpr = starExpr.X
							}

							// Try to get type name as string
							if ident, ok := recvTypeExpr.(*ast.Ident); ok {
								// Look for parent type by name
								for _, symbol := range foundPkg.Symbols {
									if symbol.Name == ident.Name &&
										(symbol.Kind == typesys.KindType ||
											symbol.Kind == typesys.KindStruct ||
											symbol.Kind == typesys.KindInterface) {
										sym.Parent = symbol
										break
									}
								}
							}
						}

						// Add to file and package
						foundFile.AddSymbol(sym)

					case *ast.GenDecl:
						// Process general declarations (type, var, const)
						for _, spec := range d.Specs {
							switch s := spec.(type) {
							case *ast.TypeSpec:
								// Process type declarations
								if s.Name == nil || s.Name.Name == "" {
									continue
								}

								// Determine kind
								kind := typesys.KindType
								if _, ok := s.Type.(*ast.StructType); ok {
									kind = typesys.KindStruct
								} else if _, ok := s.Type.(*ast.InterfaceType); ok {
									kind = typesys.KindInterface
								}

								// Create and add symbol
								sym := typesys.NewSymbol(s.Name.Name, kind)
								sym.Pos = s.Pos()
								sym.End = s.End()
								sym.File = foundFile
								sym.Package = foundPkg
								foundFile.AddSymbol(sym)

								// Process struct fields
								if structType, ok := s.Type.(*ast.StructType); ok && structType.Fields != nil {
									for _, field := range structType.Fields.List {
										for _, name := range field.Names {
											if name.Name == "" {
												continue
											}

											// Create field symbol
											fieldSym := typesys.NewSymbol(name.Name, typesys.KindField)
											fieldSym.Pos = name.Pos()
											fieldSym.End = name.End()
											fieldSym.File = foundFile
											fieldSym.Package = foundPkg
											fieldSym.Parent = sym
											foundFile.AddSymbol(fieldSym)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// For test files, the simplest approach is to mark a file for re-indexing
	// rather than trying to reload the entire module
	if len(packagesToUpdate) > 0 {
		// Since we can't easily reload just specific packages,
		// we'll rebuild the entire index after marking these files
		// as needing updates. This is a compromise for the test cases.

		// Force a rebuild of the index
		return idx.Index.Build()
	}

	return nil
}

// flagFileForUpdate marks a file as needing update
// This is a helper for the reloadFilesFromDisk method
func (idx *Indexer) flagFileForUpdate(file *typesys.File) {
	// In a real implementation, we'd add metadata to track file updates
	// For now, this is just a placeholder
}

// FindUsages finds all usages (references) of a symbol.
func (idx *Indexer) FindUsages(symbol *typesys.Symbol) []*typesys.Reference {
	return idx.Index.FindReferences(symbol)
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
