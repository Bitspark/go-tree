package typesys

import (
	"go/token"
)

// Reference represents a usage of a symbol within code.
type Reference struct {
	// Target symbol information
	Symbol *Symbol // Symbol being referenced

	// Reference location
	File    *File   // File containing the reference
	Context *Symbol // Context in which reference appears (e.g. function)
	IsWrite bool    // Whether this is a write to the symbol

	// Position
	Pos token.Pos // Start position
	End token.Pos // End position
}

// NewReference creates a new reference to a symbol.
func NewReference(symbol *Symbol, file *File, pos, end token.Pos) *Reference {
	ref := &Reference{
		Symbol: symbol,
		File:   file,
		Pos:    pos,
		End:    end,
	}

	// Add the reference to the symbol
	if symbol != nil {
		symbol.AddReference(ref)
	}

	return ref
}

// GetPosition returns position information for this reference.
func (r *Reference) GetPosition() *PositionInfo {
	if r.File == nil {
		return nil
	}
	return r.File.GetPositionInfo(r.Pos, r.End)
}

// SetContext sets the context symbol for this reference.
func (r *Reference) SetContext(context *Symbol) {
	r.Context = context
}

// SetIsWrite marks this reference as a write operation.
func (r *Reference) SetIsWrite(isWrite bool) {
	r.IsWrite = isWrite
}

// ReferencesFinder defines the interface for finding references to a symbol.
type ReferencesFinder interface {
	// FindReferences finds all references to the given symbol.
	FindReferences(symbol *Symbol) ([]*Reference, error)

	// FindReferencesByName finds references to symbols with the given name.
	FindReferencesByName(name string) ([]*Reference, error)
}

// TypeAwareReferencesFinder implements the ReferencesFinder interface with type information.
type TypeAwareReferencesFinder struct {
	Module *Module
}

// FindReferences finds all references to the given symbol.
func (f *TypeAwareReferencesFinder) FindReferences(symbol *Symbol) ([]*Reference, error) {
	// This is a placeholder that will be implemented later
	// when we have the full type checking integration
	return symbol.References, nil
}

// FindReferencesByName finds references to symbols with the given name.
func (f *TypeAwareReferencesFinder) FindReferencesByName(name string) ([]*Reference, error) {
	// This is a placeholder that will be implemented later
	// when we have the full type checking integration
	var refs []*Reference

	// Find all symbols with this name
	for _, pkg := range f.Module.Packages {
		for _, sym := range pkg.SymbolByName(name) {
			refs = append(refs, sym.References...)
		}
	}

	return refs, nil
}
