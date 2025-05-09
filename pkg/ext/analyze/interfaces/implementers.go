package interfaces

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
)

// ImplementationInfo contains details about how a type implements an interface.
type ImplementationInfo struct {
	// Type is the implementing type
	Type *typesys.Symbol

	// Interface is the implemented interface
	Interface *typesys.Symbol

	// MethodMap maps interface method names to their implementations
	MethodMap map[string]MethodImplementation

	// IsEmbedded indicates whether the implementation is through type embedding
	IsEmbedded bool

	// EmbeddedPath contains the path of embedded types if not direct
	EmbeddedPath []*typesys.Symbol
}

// MethodImplementation represents how an interface method is implemented.
type MethodImplementation struct {
	// InterfaceMethod is the method from the interface
	InterfaceMethod *typesys.Symbol

	// ImplementingMethod is the method from the implementing type
	ImplementingMethod *typesys.Symbol

	// IsDirectMatch indicates whether the method names match directly
	IsDirectMatch bool
}

// ImplementerMap stores interface implementers for efficient lookup.
type ImplementerMap struct {
	// Maps interface ID to a map of implementing type IDs
	interfaces map[string]map[string]*ImplementationInfo
}

// NewImplementerMap creates a new empty implementer map.
func NewImplementerMap() *ImplementerMap {
	return &ImplementerMap{
		interfaces: make(map[string]map[string]*ImplementationInfo),
	}
}

// Add adds an implementation to the map.
func (m *ImplementerMap) Add(info *ImplementationInfo) {
	ifaceID := getSymbolID(info.Interface)
	typID := getSymbolID(info.Type)

	// Create maps if they don't exist
	if _, exists := m.interfaces[ifaceID]; !exists {
		m.interfaces[ifaceID] = make(map[string]*ImplementationInfo)
	}

	// Store the implementation info
	m.interfaces[ifaceID][typID] = info
}

// GetImplementers gets all implementers of an interface.
func (m *ImplementerMap) GetImplementers(iface *typesys.Symbol) []*ImplementationInfo {
	ifaceID := getSymbolID(iface)
	impls, exists := m.interfaces[ifaceID]
	if !exists {
		return nil
	}

	// Convert map to slice
	result := make([]*ImplementationInfo, 0, len(impls))
	for _, info := range impls {
		result = append(result, info)
	}

	return result
}

// GetImplementation gets the implementation info for a specific type-interface pair.
func (m *ImplementerMap) GetImplementation(iface, typ *typesys.Symbol) *ImplementationInfo {
	ifaceID := getSymbolID(iface)
	typID := getSymbolID(typ)

	impls, exists := m.interfaces[ifaceID]
	if !exists {
		return nil
	}

	return impls[typID]
}

// Clear removes all entries from the map.
func (m *ImplementerMap) Clear() {
	m.interfaces = make(map[string]map[string]*ImplementationInfo)
}

// getSymbolID gets a unique ID for a symbol.
func getSymbolID(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	// In a real implementation, this would create a unique ID based on
	// package path, name, and other distinguishing characteristics
	return sym.Name
}
