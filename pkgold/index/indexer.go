// Package index provides indexing capabilities for Go code analysis.
package index

import (
	"fmt"
	"go/ast"
	"strings"

	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/visitor"
)

// Indexer builds and maintains an index for a Go module
type Indexer struct {
	// The resulting index
	Index *Index

	// Maps to keep track of symbols during indexing
	symbolsByNode map[ast.Node]*Symbol

	// Options
	includeTests   bool
	includePrivate bool
}

// NewIndexer creates a new indexer for the given module
func NewIndexer(mod *module.Module) *Indexer {
	return &Indexer{
		Index:          NewIndex(mod),
		symbolsByNode:  make(map[ast.Node]*Symbol),
		includeTests:   false,
		includePrivate: false,
	}
}

// WithTests configures whether test files should be indexed
func (i *Indexer) WithTests(include bool) *Indexer {
	i.includeTests = include
	return i
}

// WithPrivate configures whether unexported elements should be indexed
func (i *Indexer) WithPrivate(include bool) *Indexer {
	i.includePrivate = include
	return i
}

// BuildIndex builds a complete index for the module
func (i *Indexer) BuildIndex() (*Index, error) {
	// Create a visitor to collect symbols
	v := &indexingVisitor{indexer: i}

	// Create a walker to traverse the module
	walker := visitor.NewModuleWalker(v)
	walker.IncludePrivate = i.includePrivate
	walker.IncludeTests = i.includeTests

	// Walk the module to collect symbols
	if err := walker.Walk(i.Index.Module); err != nil {
		return nil, fmt.Errorf("failed to collect symbols: %w", err)
	}

	// Process references after collecting all symbols
	if err := i.processReferences(); err != nil {
		return nil, fmt.Errorf("failed to process references: %w", err)
	}

	return i.Index, nil
}

// indexingVisitor implements the ModuleVisitor interface to collect symbols during module traversal
type indexingVisitor struct {
	indexer *Indexer
}

// VisitModule is called when visiting a module
func (v *indexingVisitor) VisitModule(mod *module.Module) error {
	// Nothing to do at module level
	return nil
}

// VisitPackage is called when visiting a package
func (v *indexingVisitor) VisitPackage(pkg *module.Package) error {
	// Nothing to do at package level
	return nil
}

// VisitFile is called when visiting a file
func (v *indexingVisitor) VisitFile(file *module.File) error {
	// Nothing to do at file level, individual elements will be visited
	return nil
}

// VisitType is called when visiting a type
func (v *indexingVisitor) VisitType(typ *module.Type) error {
	if !v.indexer.includePrivate && !typ.IsExported {
		return nil
	}

	// Create a symbol for this type
	symbol := &Symbol{
		Name:          typ.Name,
		Kind:          KindType,
		Package:       typ.Package.ImportPath,
		QualifiedName: typ.Package.ImportPath + "." + typ.Name,
		File:          typ.File.Path,
		Pos:           typ.Pos,
		End:           typ.End,
	}

	// Add position information if available
	if pos := typ.File.GetPositionInfo(typ.Pos, typ.End); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// VisitFunction is called when visiting a function
func (v *indexingVisitor) VisitFunction(fn *module.Function) error {
	if !v.indexer.includePrivate && !fn.IsExported {
		return nil
	}

	// Skip test functions if not including tests
	if fn.IsTest && !v.indexer.includeTests {
		return nil
	}

	// Create a symbol for this function
	symbol := &Symbol{
		Name:          fn.Name,
		Kind:          KindFunction,
		Package:       fn.Package.ImportPath,
		QualifiedName: fn.Package.ImportPath + "." + fn.Name,
		File:          fn.File.Path,
		Pos:           fn.Pos,
		End:           fn.End,
	}

	// For methods, update the kind and add receiver information
	if fn.IsMethod && fn.Receiver != nil {
		symbol.Kind = KindMethod
		symbol.ReceiverType = fn.Receiver.Type
		// Remove pointer if present for the parent type
		symbol.ParentType = strings.TrimPrefix(fn.Receiver.Type, "*")
		// Update qualified name to include the receiver type
		symbol.QualifiedName = fn.Package.ImportPath + "." + symbol.ParentType + "." + fn.Name
	}

	// Add position information if available
	if pos := fn.File.GetPositionInfo(fn.Pos, fn.End); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	// Store mapping from AST node to symbol if available
	if fn.AST != nil {
		v.indexer.symbolsByNode[fn.AST] = symbol
	}

	return nil
}

// VisitMethod is called when visiting a method from a type definition
func (v *indexingVisitor) VisitMethod(method *module.Method) error {
	// Method on type (different from a function with a receiver)
	// These are typically collected with types, but we index them separately as well

	// Skip if parent type is not exported and we're not including private elements
	if method.Parent != nil && !v.indexer.includePrivate && !method.Parent.IsExported {
		return nil
	}

	// Create a symbol for this method
	symbol := &Symbol{
		Name: method.Name,
		Kind: KindMethod,
		File: method.Parent.File.Path,
		Pos:  method.Pos,
		End:  method.End,
	}

	// Add type context if available
	if method.Parent != nil {
		symbol.Package = method.Parent.Package.ImportPath
		symbol.QualifiedName = method.Parent.Package.ImportPath + "." + method.Parent.Name + "." + method.Name
		symbol.ParentType = method.Parent.Name
	}

	// Add position information if available
	if pos := method.GetPosition(); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// VisitField is called when visiting a struct field
func (v *indexingVisitor) VisitField(field *module.Field) error {
	// Create a symbol for this field
	symbol := &Symbol{
		Name:          field.Name,
		Kind:          KindField,
		Package:       field.Parent.Package.ImportPath,
		QualifiedName: field.Parent.Package.ImportPath + "." + field.Parent.Name + "." + field.Name,
		File:          field.Parent.File.Path,
		Pos:           field.Pos,
		End:           field.End,
		ParentType:    field.Parent.Name,
		TypeName:      field.Type,
	}

	// Add position information if available
	if pos := field.GetPosition(); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// VisitVariable is called when visiting a variable
func (v *indexingVisitor) VisitVariable(variable *module.Variable) error {
	if !v.indexer.includePrivate && !variable.IsExported {
		return nil
	}

	// Create a symbol for this variable
	symbol := &Symbol{
		Name:          variable.Name,
		Kind:          KindVariable,
		Package:       variable.Package.ImportPath,
		QualifiedName: variable.Package.ImportPath + "." + variable.Name,
		File:          variable.File.Path,
		Pos:           variable.Pos,
		End:           variable.End,
		TypeName:      variable.Type,
	}

	// Add position information if available
	if pos := variable.File.GetPositionInfo(variable.Pos, variable.End); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// VisitConstant is called when visiting a constant
func (v *indexingVisitor) VisitConstant(constant *module.Constant) error {
	if !v.indexer.includePrivate && !constant.IsExported {
		return nil
	}

	// Create a symbol for this constant
	symbol := &Symbol{
		Name:          constant.Name,
		Kind:          KindConstant,
		Package:       constant.Package.ImportPath,
		QualifiedName: constant.Package.ImportPath + "." + constant.Name,
		File:          constant.File.Path,
		Pos:           constant.Pos,
		End:           constant.End,
		TypeName:      constant.Type,
	}

	// Add position information if available
	if pos := constant.File.GetPositionInfo(constant.Pos, constant.End); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// VisitImport is called when visiting an import
func (v *indexingVisitor) VisitImport(imp *module.Import) error {
	// Create a symbol for this import
	symbol := &Symbol{
		Name:          imp.Name,
		Kind:          KindImport,
		Package:       imp.File.Package.ImportPath,
		QualifiedName: imp.Path,
		File:          imp.File.Path,
		Pos:           imp.Pos,
		End:           imp.End,
	}

	// Add position information if available
	if pos := imp.File.GetPositionInfo(imp.Pos, imp.End); pos != nil {
		symbol.LineStart = pos.LineStart
		symbol.LineEnd = pos.LineEnd
	}

	// Add to index
	v.indexer.Index.AddSymbol(symbol)

	return nil
}

// countSymbols counts the total number of symbols in the map
func countSymbols(symbolsByName map[string][]*Symbol) int {
	count := 0
	for _, symbols := range symbolsByName {
		count += len(symbols)
	}
	return count
}

// processReferences analyzes the AST of each file to find references to symbols
func (i *Indexer) processReferences() error {
	// Enable debug output for finding references
	debug := false

	if debug {
		fmt.Printf("DEBUG: Looking for references to %d symbols\n", countSymbols(i.Index.SymbolsByName))
	}

	// Iterate through all packages in the module
	for _, pkg := range i.Index.Module.Packages {
		// Skip test packages if not including tests
		if pkg.IsTest && !i.includeTests {
			continue
		}

		if debug {
			fmt.Printf("DEBUG: Processing package %s for references\n", pkg.Name)
		}

		// Process each file in the package
		for _, file := range pkg.Files {
			// Skip test files if not including tests
			if file.IsTest && !i.includeTests {
				continue
			}

			// Skip files without AST
			if file.AST == nil {
				if debug {
					fmt.Printf("DEBUG: Skipping file %s - no AST\n", file.Path)
				}
				continue
			}

			if debug {
				fmt.Printf("DEBUG: Processing file %s for references\n", file.Path)
				fmt.Printf("DEBUG: AST: %T %+v\n", file.AST, file.AST.Name)
			}

			// Process the file to find references
			if err := i.processFileReferences(file, debug); err != nil {
				return fmt.Errorf("failed to process references in file %s: %w", file.Path, err)
			}
		}
	}

	return nil
}

// processFileReferences finds references to symbols in a file's AST
func (i *Indexer) processFileReferences(file *module.File, debug bool) error {
	// Create an AST visitor to find references
	astVisitor := &referenceVisitor{
		indexer: i,
		file:    file,
		debug:   debug,
	}

	// Visit the entire AST
	ast.Walk(astVisitor, file.AST)

	return nil
}

// referenceVisitor implements the ast.Visitor interface to find references to symbols
type referenceVisitor struct {
	indexer *Indexer
	file    *module.File
	debug   bool

	// Current context (e.g., function we're inside)
	currentFunc *ast.FuncDecl
}

// Visit processes AST nodes to find references
func (v *referenceVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	// Track context
	switch n := node.(type) {
	case *ast.FuncDecl:
		v.currentFunc = n
		defer func() { v.currentFunc = nil }()

	case *ast.Ident:
		// Skip blank identifiers
		if n.Name == "_" {
			return v
		}

		if v.debug {
			fmt.Printf("DEBUG: Found identifier %s at pos %v\n", n.Name, n.Pos())
		}

		// Look for this identifier in the symbols by name
		symbols := v.indexer.Index.FindSymbolsByName(n.Name)
		if len(symbols) > 0 {
			if v.debug {
				fmt.Printf("DEBUG: Found symbol match for %s: %d matches\n", n.Name, len(symbols))
			}

			// Create a reference to this symbol
			// For simplicity, we're just using the first matching symbol
			// A more sophisticated implementation would resolve which symbol this actually refers to
			symbol := symbols[0]

			// Skip self-references (where the identifier is the definition itself)
			// This prevents counting definition as a reference
			if symbol.File == v.file.Path {
				filePos := v.file.FileSet.Position(n.Pos())
				symbolPos := v.file.FileSet.Position(symbol.Pos)

				// If positions are very close, this might be the definition itself
				// We need to ignore variable declarations but keep references
				if filePos.Line == symbolPos.Line && filePos.Column >= symbolPos.Column && filePos.Column <= symbolPos.Column+len(symbol.Name) {
					if v.debug {
						fmt.Printf("DEBUG: Skipping self-reference at line %d, col %d\n", filePos.Line, filePos.Column)
					}
					return v
				}
			}

			// Get position info
			var lineStart, lineEnd int
			pos := n.Pos()
			end := n.End()

			if v.file.FileSet != nil {
				posInfo := v.file.FileSet.Position(pos)
				endInfo := v.file.FileSet.Position(end)
				lineStart = posInfo.Line
				lineEnd = endInfo.Line

				if v.debug {
					fmt.Printf("DEBUG: Adding reference to %s at line %d\n", n.Name, lineStart)
				}
			}

			// Create the reference
			ref := &Reference{
				TargetSymbol: symbol,
				File:         v.file.Path,
				Pos:          pos,
				End:          end,
				LineStart:    lineStart,
				LineEnd:      lineEnd,
			}

			// Add context information if available
			if v.currentFunc != nil {
				if v.currentFunc.Name != nil {
					ref.Context = v.currentFunc.Name.Name
				}
			}

			// Add to index
			v.indexer.Index.AddReference(symbol, ref)
		}

	case *ast.SelectorExpr:
		// Handle qualified references like pkg.Name
		if ident, ok := n.X.(*ast.Ident); ok {
			// Check if the selector (X.Sel) might be a reference to a symbol
			// This is a simplified implementation; a proper one would resolve package aliases
			// and check more carefully if this is a real reference
			if ident.Name != "" && n.Sel != nil && n.Sel.Name != "" {
				qualifiedName := ident.Name + "." + n.Sel.Name

				if v.debug {
					fmt.Printf("DEBUG: Found selector expr %s at pos %v\n", qualifiedName, n.Pos())
				}

				// First try to match by fully qualified name
				// This helps with package imports
				found := false
				for _, symbols := range v.indexer.Index.SymbolsByName {
					for _, symbol := range symbols {
						// Check if this is a direct reference to the symbol
						// e.g., somepackage.Something or Type.Method
						if strings.HasSuffix(symbol.QualifiedName, qualifiedName) ||
							(symbol.Name == n.Sel.Name && (symbol.ParentType == ident.Name || symbol.Package == ident.Name)) {

							if v.debug {
								fmt.Printf("DEBUG: Found qualified reference to %s.%s (%s)\n",
									ident.Name, n.Sel.Name, symbol.QualifiedName)
							}

							// Get position info
							var lineStart, lineEnd int
							pos := n.Pos()
							end := n.End()

							if v.file.FileSet != nil {
								posInfo := v.file.FileSet.Position(pos)
								endInfo := v.file.FileSet.Position(end)
								lineStart = posInfo.Line
								lineEnd = endInfo.Line
							}

							// Create the reference
							ref := &Reference{
								TargetSymbol: symbol,
								File:         v.file.Path,
								Pos:          pos,
								End:          end,
								LineStart:    lineStart,
								LineEnd:      lineEnd,
							}

							// Add context information if available
							if v.currentFunc != nil {
								if v.currentFunc.Name != nil {
									ref.Context = v.currentFunc.Name.Name
								}
							}

							// Add to index
							v.indexer.Index.AddReference(symbol, ref)
							found = true
							break
						}
					}
					if found {
						break
					}
				}

				// If we haven't found a match, try looking just for the selector part
				// This helps with methods on variables
				if !found {
					symbols := v.indexer.Index.FindSymbolsByName(n.Sel.Name)
					for _, symbol := range symbols {
						// For methods, make sure this is a method on a type
						if symbol.Kind == KindMethod && symbol.ParentType != "" {
							if v.debug {
								fmt.Printf("DEBUG: Found potential method reference: %s on %s\n",
									symbol.Name, symbol.ParentType)
							}

							// Get position info
							var lineStart, lineEnd int
							pos := n.Sel.Pos()
							end := n.Sel.End()

							if v.file.FileSet != nil {
								posInfo := v.file.FileSet.Position(pos)
								endInfo := v.file.FileSet.Position(end)
								lineStart = posInfo.Line
								lineEnd = endInfo.Line
							}

							// Create the reference
							ref := &Reference{
								TargetSymbol: symbol,
								File:         v.file.Path,
								Pos:          pos,
								End:          end,
								LineStart:    lineStart,
								LineEnd:      lineEnd,
							}

							// Add context information if available
							if v.currentFunc != nil {
								if v.currentFunc.Name != nil {
									ref.Context = v.currentFunc.Name.Name
								}
							}

							// Add to index
							v.indexer.Index.AddReference(symbol, ref)
						}
					}
				}
			}
		}
	}

	return v
}
