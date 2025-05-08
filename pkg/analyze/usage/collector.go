// Package usage provides functionality for analyzing symbol usage throughout the codebase.
package usage

import (
	"fmt"

	"bitspark.dev/go-tree/pkg/analyze"
	"bitspark.dev/go-tree/pkg/typesys"
)

// ReferenceKind represents the kind of reference to a symbol.
type ReferenceKind int

const (
	// ReferenceUnknown is an unknown reference kind.
	ReferenceUnknown ReferenceKind = iota
	// ReferenceRead is a read of a symbol.
	ReferenceRead
	// ReferenceWrite is a write to a symbol.
	ReferenceWrite
	// ReferenceCall is a call to a function or method.
	ReferenceCall
	// ReferenceImport is an import of a package.
	ReferenceImport
	// ReferenceType is a use of a type.
	ReferenceType
	// ReferenceEmbed is an embedding of a type.
	ReferenceEmbed
)

// String returns a string representation of the reference kind.
func (k ReferenceKind) String() string {
	switch k {
	case ReferenceRead:
		return "read"
	case ReferenceWrite:
		return "write"
	case ReferenceCall:
		return "call"
	case ReferenceImport:
		return "import"
	case ReferenceType:
		return "type"
	case ReferenceEmbed:
		return "embed"
	default:
		return "unknown"
	}
}

// SymbolUsage represents usage information for a symbol.
type SymbolUsage struct {
	// The symbol being analyzed
	Symbol *typesys.Symbol

	// References to the symbol, categorized by kind
	References map[ReferenceKind][]*typesys.Reference

	// Files where the symbol is used
	Files map[string]bool

	// Packages where the symbol is used
	Packages map[string]bool

	// Contexts (functions, methods) where the symbol is used
	Contexts map[string]*typesys.Symbol
}

// NewSymbolUsage creates a new symbol usage for the given symbol.
func NewSymbolUsage(sym *typesys.Symbol) *SymbolUsage {
	return &SymbolUsage{
		Symbol:     sym,
		References: make(map[ReferenceKind][]*typesys.Reference),
		Files:      make(map[string]bool),
		Packages:   make(map[string]bool),
		Contexts:   make(map[string]*typesys.Symbol),
	}
}

// AddReference adds a reference to the symbol usage.
func (u *SymbolUsage) AddReference(ref *typesys.Reference, kind ReferenceKind) {
	// Add to references by kind
	u.References[kind] = append(u.References[kind], ref)

	// Track file usage
	if ref.File != nil {
		u.Files[ref.File.Path] = true
		if ref.File.Package != nil {
			u.Packages[ref.File.Package.ImportPath] = true
		}
	}

	// Track context usage
	if ref.Context != nil {
		u.Contexts[getSymbolID(ref.Context)] = ref.Context
	}
}

// GetReferenceCount returns the total number of references to the symbol.
func (u *SymbolUsage) GetReferenceCount() int {
	count := 0
	for _, refs := range u.References {
		count += len(refs)
	}
	return count
}

// GetReferenceCountByKind returns the number of references of the given kind.
func (u *SymbolUsage) GetReferenceCountByKind(kind ReferenceKind) int {
	return len(u.References[kind])
}

// GetFileCount returns the number of files where the symbol is used.
func (u *SymbolUsage) GetFileCount() int {
	return len(u.Files)
}

// GetPackageCount returns the number of packages where the symbol is used.
func (u *SymbolUsage) GetPackageCount() int {
	return len(u.Packages)
}

// GetContextCount returns the number of contexts where the symbol is used.
func (u *SymbolUsage) GetContextCount() int {
	return len(u.Contexts)
}

// UsageCollector collects usage information for symbols.
type UsageCollector struct {
	*analyze.BaseAnalyzer
	Module *typesys.Module
}

// NewUsageCollector creates a new usage collector.
func NewUsageCollector(module *typesys.Module) *UsageCollector {
	return &UsageCollector{
		BaseAnalyzer: analyze.NewBaseAnalyzer(
			"UsageCollector",
			"Collects usage information for symbols",
		),
		Module: module,
	}
}

// CollectUsage collects usage information for a specific symbol.
func (c *UsageCollector) CollectUsage(sym *typesys.Symbol) (*SymbolUsage, error) {
	if c.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	if sym == nil {
		return nil, fmt.Errorf("symbol is nil")
	}

	// Create a new symbol usage
	usage := NewSymbolUsage(sym)

	// Process references to the symbol
	for _, ref := range sym.References {
		kind := determineReferenceKind(ref)
		usage.AddReference(ref, kind)
	}

	return usage, nil
}

// CollectUsageForAllSymbols collects usage information for all symbols in the module.
func (c *UsageCollector) CollectUsageForAllSymbols() (map[string]*SymbolUsage, error) {
	if c.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	usages := make(map[string]*SymbolUsage)

	// Process each package
	for _, pkg := range c.Module.Packages {
		// Process each symbol in the package
		for _, sym := range pkg.Symbols {
			usage, err := c.CollectUsage(sym)
			if err != nil {
				continue
			}
			usages[getSymbolID(sym)] = usage
		}
	}

	return usages, nil
}

// CollectionResult represents the result of a usage collection operation.
type CollectionResult struct {
	*analyze.BaseResult
	Usages map[string]*SymbolUsage
}

// GetUsages returns the collected usage information.
func (r *CollectionResult) GetUsages() map[string]*SymbolUsage {
	return r.Usages
}

// NewCollectionResult creates a new collection result.
func NewCollectionResult(collector *UsageCollector, usages map[string]*SymbolUsage, err error) *CollectionResult {
	return &CollectionResult{
		BaseResult: analyze.NewBaseResult(collector, err),
		Usages:     usages,
	}
}

// CollectAsync collects usage information asynchronously and returns a result channel.
func (c *UsageCollector) CollectAsync() <-chan *CollectionResult {
	resultCh := make(chan *CollectionResult, 1)

	go func() {
		usages, err := c.CollectUsageForAllSymbols()
		resultCh <- NewCollectionResult(c, usages, err)
		close(resultCh)
	}()

	return resultCh
}

// Helper functions

// determineReferenceKind determines the kind of reference.
func determineReferenceKind(ref *typesys.Reference) ReferenceKind {
	if ref == nil || ref.Symbol == nil {
		return ReferenceUnknown
	}

	// Check for write reference
	if ref.IsWrite {
		return ReferenceWrite
	}

	// Check symbol kind to determine reference kind
	switch ref.Symbol.Kind {
	case typesys.KindFunction, typesys.KindMethod:
		return ReferenceCall
	case typesys.KindPackage:
		return ReferenceImport
	case typesys.KindType, typesys.KindStruct, typesys.KindInterface:
		return ReferenceType
	default:
		return ReferenceRead
	}
}

// getSymbolID gets a unique ID for a symbol.
func getSymbolID(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	// For functions, include the package path for uniqueness
	// For methods, include the receiver type as well
	if sym.Package != nil {
		pkg := sym.Package.ImportPath
		if sym.Kind == typesys.KindMethod && sym.Parent != nil {
			return fmt.Sprintf("%s.%s.%s", pkg, sym.Parent.Name, sym.Name)
		}
		return fmt.Sprintf("%s.%s", pkg, sym.Name)
	}

	return sym.Name
}
