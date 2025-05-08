package typesys

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/types/typeutil"
)

// TypeBridge provides a bridge between our type system and Go's type system.
type TypeBridge struct {
	// Maps from our symbols to Go's type objects
	SymToObj map[*Symbol]types.Object

	// Maps from Go's type objects to our symbols
	ObjToSym map[types.Object]*Symbol

	// Maps from AST nodes to our symbols
	NodeToSym map[ast.Node]*Symbol

	// Method set cache for quick lookup of methods
	MethodSets *typeutil.MethodSetCache
}

// NewTypeBridge creates a new type bridge.
func NewTypeBridge() *TypeBridge {
	return &TypeBridge{
		SymToObj:   make(map[*Symbol]types.Object),
		ObjToSym:   make(map[types.Object]*Symbol),
		NodeToSym:  make(map[ast.Node]*Symbol),
		MethodSets: &typeutil.MethodSetCache{},
	}
}

// MapSymbolToObject maps a symbol to a Go type object.
func (b *TypeBridge) MapSymbolToObject(sym *Symbol, obj types.Object) {
	b.SymToObj[sym] = obj
	b.ObjToSym[obj] = sym
}

// MapNodeToSymbol maps an AST node to a symbol.
func (b *TypeBridge) MapNodeToSymbol(node ast.Node, sym *Symbol) {
	b.NodeToSym[node] = sym
}

// GetSymbolForObject returns the symbol for a Go type object.
func (b *TypeBridge) GetSymbolForObject(obj types.Object) *Symbol {
	return b.ObjToSym[obj]
}

// GetObjectForSymbol returns the Go type object for a symbol.
func (b *TypeBridge) GetObjectForSymbol(sym *Symbol) types.Object {
	return b.SymToObj[sym]
}

// GetSymbolForNode returns the symbol for an AST node.
func (b *TypeBridge) GetSymbolForNode(node ast.Node) *Symbol {
	return b.NodeToSym[node]
}

// GetImplementations finds all types that implement an interface.
func (b *TypeBridge) GetImplementations(iface *types.Interface, assignable bool) []*Symbol {
	var result []*Symbol

	// For each symbol in our map
	for sym, obj := range b.SymToObj {
		// Skip non-type symbols
		if sym.Kind != KindType && sym.Kind != KindStruct {
			continue
		}

		// Get the named type
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		// Check if it implements the interface
		if implements(named, iface, assignable) {
			result = append(result, sym)
		}
	}

	return result
}

// Helper function to check if a type implements an interface
func implements(named *types.Named, iface *types.Interface, assignable bool) bool {
	if assignable {
		return types.AssignableTo(named, iface)
	}
	return types.Implements(named, iface)
}

// GetMethodsOfType returns all methods of a type.
func (b *TypeBridge) GetMethodsOfType(typ types.Type) []*Symbol {
	var result []*Symbol

	// Get the method set
	mset := b.MethodSets.MethodSet(typ)

	// Find symbols for each method
	for i := 0; i < mset.Len(); i++ {
		method := mset.At(i).Obj()
		if sym := b.GetSymbolForObject(method); sym != nil {
			result = append(result, sym)
		}
	}

	return result
}

// BuildTypeBridge builds the type bridge for a module.
func BuildTypeBridge(mod *Module) *TypeBridge {
	bridge := NewTypeBridge()

	// Process each package
	for _, pkg := range mod.Packages {
		// Skip packages without type info
		if pkg.TypesInfo == nil {
			continue
		}

		// Process objects defined in this package
		for id, obj := range pkg.TypesInfo.Defs {
			if obj == nil {
				continue
			}

			// Find our symbol for this object
			for _, sym := range pkg.Symbols {
				if sym.Name == id.Name {
					// Check if this is the right symbol based on position
					if pkg.Module.FileSet.Position(sym.Pos).Offset == pkg.Module.FileSet.Position(id.Pos()).Offset {
						bridge.MapSymbolToObject(sym, obj)
						bridge.MapNodeToSymbol(id, sym)

						// Also set the TypeObj in the symbol
						sym.TypeObj = obj

						// If it's a typed symbol, set the TypeInfo
						switch obj.(type) {
						case *types.Var, *types.Const:
							sym.TypeInfo = obj.Type()
						}

						break
					}
				}
			}
		}

		// Process type usages
		for expr, typ := range pkg.TypesInfo.Types {
			if typ.Type == nil {
				continue
			}

			// Find symbols in the same file/position
			for _, file := range pkg.Files {
				for _, sym := range file.Symbols {
					// Check if the positions match
					if file.FileSet.Position(sym.Pos).Offset == file.FileSet.Position(expr.Pos()).Offset {
						// Set the TypeInfo in the symbol
						sym.TypeInfo = typ.Type
						break
					}
				}
			}
		}
	}

	return bridge
}
