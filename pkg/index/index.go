package index

import (
	"fmt"
	"go/types"
	"sync"

	"bitspark.dev/go-tree/pkg/typesys"
)

// Index provides fast lookup capabilities for symbols and references in a module.
// It builds on the typesys package to provide type-aware indexing.
type Index struct {
	// The module being indexed
	Module *typesys.Module

	// Maps for fast lookup
	symbolsByID       map[string]*typesys.Symbol               // ID -> Symbol
	symbolsByName     map[string][]*typesys.Symbol             // Name -> Symbols
	symbolsByFile     map[string][]*typesys.Symbol             // File path -> Symbols
	symbolsByKind     map[typesys.SymbolKind][]*typesys.Symbol // Kind -> Symbols
	referencesByID    map[string][]*typesys.Reference          // Symbol ID -> References
	referencesByFile  map[string][]*typesys.Reference          // File path -> References
	methodsByReceiver map[string][]*typesys.Symbol             // Receiver type -> Methods

	// Type-specific lookup maps
	interfaceImpls map[string][]*typesys.Symbol // Interface ID -> Implementors

	// Cache of type bridge for type-based operations
	typeBridge *typesys.TypeBridge

	// Mutex for concurrent access
	mu sync.RWMutex
}

// NewIndex creates a new empty index for the given module.
func NewIndex(mod *typesys.Module) *Index {
	return &Index{
		Module:            mod,
		symbolsByID:       make(map[string]*typesys.Symbol),
		symbolsByName:     make(map[string][]*typesys.Symbol),
		symbolsByFile:     make(map[string][]*typesys.Symbol),
		symbolsByKind:     make(map[typesys.SymbolKind][]*typesys.Symbol),
		referencesByID:    make(map[string][]*typesys.Reference),
		referencesByFile:  make(map[string][]*typesys.Reference),
		methodsByReceiver: make(map[string][]*typesys.Symbol),
		interfaceImpls:    make(map[string][]*typesys.Symbol),
	}
}

// Build rebuilds the entire index from the module.
func (idx *Index) Build() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear existing maps
	idx.clear()

	// Debug info
	fmt.Printf("Building index for module with %d packages\n", len(idx.Module.Packages))

	// Print packages
	for pkgPath, pkg := range idx.Module.Packages {
		fmt.Printf("Package: %s with %d symbols and %d files\n", pkgPath, len(pkg.Symbols), len(pkg.Files))
		// Print first few symbols
		count := 0
		for _, sym := range pkg.Symbols {
			if count >= 5 {
				fmt.Printf("  ... and %d more symbols\n", len(pkg.Symbols)-5)
				break
			}
			fmt.Printf("  Symbol: %s (%s)\n", sym.Name, sym.Kind)
			count++
		}
	}

	// Build type bridge for the module
	idx.typeBridge = typesys.BuildTypeBridge(idx.Module)

	// Process all symbols
	symbolCount := 0
	for _, pkg := range idx.Module.Packages {
		for _, sym := range pkg.Symbols {
			idx.indexSymbol(sym)
			symbolCount++
		}
	}
	fmt.Printf("Indexed %d symbols in total\n", symbolCount)

	// Process references after all symbols are indexed
	refCount := 0
	for _, pkg := range idx.Module.Packages {
		for _, sym := range pkg.Symbols {
			for _, ref := range sym.References {
				idx.indexReference(ref)
				refCount++
			}
		}
	}
	fmt.Printf("Indexed %d references in total\n", refCount)

	// Build additional lookup maps
	idx.buildMethodIndex()
	idx.buildInterfaceImplIndex()

	// Check result
	fmt.Printf("Index stats:\n")
	fmt.Printf("  Symbols by ID: %d\n", len(idx.symbolsByID))
	fmt.Printf("  Symbols by Name: %d\n", len(idx.symbolsByName))
	fmt.Printf("  Symbols by File: %d\n", len(idx.symbolsByFile))
	fmt.Printf("  Symbols by Kind: %d\n", len(idx.symbolsByKind))

	return nil
}

// Update updates the index for the given files.
func (idx *Index) Update(files []string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove existing entries for these files
	for _, file := range files {
		idx.removeFileEntries(file)
	}

	// Add new entries
	for _, file := range files {
		fileObj := idx.Module.FileByPath(file)
		if fileObj == nil {
			continue
		}

		// Index symbols in this file
		for _, sym := range fileObj.Symbols {
			idx.indexSymbol(sym)
		}
	}

	// Update references
	for _, file := range files {
		fileObj := idx.Module.FileByPath(file)
		if fileObj == nil {
			continue
		}

		// Index references in this file
		for _, sym := range fileObj.Symbols {
			for _, ref := range sym.References {
				idx.indexReference(ref)
			}
		}
	}

	// Rebuild method and interface indices
	idx.buildMethodIndex()
	idx.buildInterfaceImplIndex()

	return nil
}

// GetSymbolByID returns a symbol by its ID.
func (idx *Index) GetSymbolByID(id string) *typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.symbolsByID[id]
}

// FindSymbolsByName returns all symbols with the given name.
func (idx *Index) FindSymbolsByName(name string) []*typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.symbolsByName[name]
}

// FindSymbolsByKind returns all symbols of the given kind.
func (idx *Index) FindSymbolsByKind(kind typesys.SymbolKind) []*typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.symbolsByKind[kind]
}

// FindSymbolsInFile returns all symbols defined in the given file.
func (idx *Index) FindSymbolsInFile(filePath string) []*typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.symbolsByFile[filePath]
}

// FindReferences returns all references to the given symbol.
func (idx *Index) FindReferences(symbol *typesys.Symbol) []*typesys.Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.referencesByID[symbol.ID]
}

// FindReferencesInFile returns all references in the given file.
func (idx *Index) FindReferencesInFile(filePath string) []*typesys.Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.referencesByFile[filePath]
}

// FindMethods returns all methods for the given type.
func (idx *Index) FindMethods(typeName string) []*typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.methodsByReceiver[typeName]
}

// FindImplementations returns all implementations of the given interface.
func (idx *Index) FindImplementations(interfaceSym *typesys.Symbol) []*typesys.Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.interfaceImpls[interfaceSym.ID]
}

// clear clears all maps in the index.
func (idx *Index) clear() {
	idx.symbolsByID = make(map[string]*typesys.Symbol)
	idx.symbolsByName = make(map[string][]*typesys.Symbol)
	idx.symbolsByFile = make(map[string][]*typesys.Symbol)
	idx.symbolsByKind = make(map[typesys.SymbolKind][]*typesys.Symbol)
	idx.referencesByID = make(map[string][]*typesys.Reference)
	idx.referencesByFile = make(map[string][]*typesys.Reference)
	idx.methodsByReceiver = make(map[string][]*typesys.Symbol)
	idx.interfaceImpls = make(map[string][]*typesys.Symbol)
}

// indexSymbol adds a symbol to the index.
func (idx *Index) indexSymbol(sym *typesys.Symbol) {
	// Add to ID index
	idx.symbolsByID[sym.ID] = sym

	// Add to name index
	idx.symbolsByName[sym.Name] = append(idx.symbolsByName[sym.Name], sym)

	// Add to file index
	if sym.File != nil {
		idx.symbolsByFile[sym.File.Path] = append(idx.symbolsByFile[sym.File.Path], sym)
	}

	// Add to kind index
	idx.symbolsByKind[sym.Kind] = append(idx.symbolsByKind[sym.Kind], sym)
}

// indexReference adds a reference to the index.
func (idx *Index) indexReference(ref *typesys.Reference) {
	// Add to symbol references
	if ref.Symbol != nil {
		idx.referencesByID[ref.Symbol.ID] = append(idx.referencesByID[ref.Symbol.ID], ref)
	}

	// Add to file references
	if ref.File != nil {
		idx.referencesByFile[ref.File.Path] = append(idx.referencesByFile[ref.File.Path], ref)
	}
}

// removeFileEntries removes all index entries for the given file.
func (idx *Index) removeFileEntries(filePath string) {
	// Remove symbols
	for _, sym := range idx.symbolsByFile[filePath] {
		delete(idx.symbolsByID, sym.ID)

		// Remove from name index
		idx.symbolsByName[sym.Name] = removeSymbol(idx.symbolsByName[sym.Name], sym)

		// Remove from kind index
		idx.symbolsByKind[sym.Kind] = removeSymbol(idx.symbolsByKind[sym.Kind], sym)
	}

	// Clear file entry
	delete(idx.symbolsByFile, filePath)

	// Remove references
	delete(idx.referencesByFile, filePath)

	// Note: We'll rebuild the references for all symbols later
}

// buildMethodIndex builds the method lookup index.
func (idx *Index) buildMethodIndex() {
	idx.methodsByReceiver = make(map[string][]*typesys.Symbol)

	// Find all methods
	methods := idx.symbolsByKind[typesys.KindMethod]
	for _, method := range methods {
		// Skip methods without a parent
		if method.Parent == nil {
			continue
		}

		// Add to receiver index
		receiverName := method.Parent.Name
		idx.methodsByReceiver[receiverName] = append(idx.methodsByReceiver[receiverName], method)
	}
}

// buildInterfaceImplIndex builds the interface implementation lookup index.
func (idx *Index) buildInterfaceImplIndex() {
	idx.interfaceImpls = make(map[string][]*typesys.Symbol)

	// Find all interfaces
	interfaces := idx.symbolsByKind[typesys.KindInterface]
	for _, iface := range interfaces {
		// Skip interfaces without type object
		ifaceObj := idx.typeBridge.GetObjectForSymbol(iface)
		if ifaceObj == nil {
			continue
		}

		// Get the interface type
		ifaceType, ok := ifaceObj.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		// Find implementations
		impls := idx.typeBridge.GetImplementations(ifaceType, true)
		idx.interfaceImpls[iface.ID] = impls
	}
}

// Helper function to remove a symbol from a slice
func removeSymbol(syms []*typesys.Symbol, sym *typesys.Symbol) []*typesys.Symbol {
	for i, s := range syms {
		if s == sym {
			// Remove the element at index i
			return append(syms[:i], syms[i+1:]...)
		}
	}
	return syms
}
