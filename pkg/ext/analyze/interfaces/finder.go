// Package interfaces provides functionality for finding and analyzing interface implementations.
package interfaces

import (
	"bitspark.dev/go-tree/pkg/ext/analyze"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// FindOptions provides filtering options for interface implementation search.
type FindOptions struct {
	// Packages limits the search to specific packages.
	Packages []string

	// ExportedOnly indicates whether to only find exported types.
	ExportedOnly bool

	// Direct indicates whether to only find direct implementations (no embedding).
	Direct bool

	// IncludeGenerics indicates whether to include generic implementations.
	IncludeGenerics bool
}

// DefaultFindOptions returns the default find options.
func DefaultFindOptions() *FindOptions {
	return &FindOptions{
		Packages:        nil, // All packages
		ExportedOnly:    false,
		Direct:          false,
		IncludeGenerics: true,
	}
}

// InterfaceFinder finds implementations of interfaces in a module.
type InterfaceFinder struct {
	*analyze.BaseAnalyzer
	Module *typesys.Module
}

// NewInterfaceFinder creates a new interface finder with the given module.
func NewInterfaceFinder(module *typesys.Module) *InterfaceFinder {
	return &InterfaceFinder{
		BaseAnalyzer: analyze.NewBaseAnalyzer(
			"InterfaceFinder",
			"Finds implementations of interfaces in Go code",
		),
		Module: module,
	}
}

// FindImplementations finds all types implementing the given interface.
func (f *InterfaceFinder) FindImplementations(iface *typesys.Symbol) ([]*typesys.Symbol, error) {
	return f.FindImplementationsMatching(iface, DefaultFindOptions())
}

// FindImplementationsMatching finds interface implementations matching the given criteria.
func (f *InterfaceFinder) FindImplementationsMatching(iface *typesys.Symbol, opts *FindOptions) ([]*typesys.Symbol, error) {
	if f.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	if iface == nil {
		return nil, fmt.Errorf("interface symbol is nil")
	}

	if opts == nil {
		opts = DefaultFindOptions()
	}

	// Verify that the provided symbol is an interface
	if !isInterface(iface) {
		return nil, fmt.Errorf("symbol %s is not an interface", iface.Name)
	}

	// Get all eligible types based on options
	eligibleTypes := f.getEligibleTypes(opts)

	// Filter for implementations
	var implementations []*typesys.Symbol
	for _, typ := range eligibleTypes {
		isImpl, err := f.IsImplementedBy(iface, typ)
		if err != nil {
			continue
		}

		if isImpl {
			implementations = append(implementations, typ)
		}
	}

	return implementations, nil
}

// IsImplementedBy checks if an interface is implemented by a type.
func (f *InterfaceFinder) IsImplementedBy(iface, typ *typesys.Symbol) (bool, error) {
	if f.Module == nil || iface == nil || typ == nil {
		return false, fmt.Errorf("invalid parameters")
	}

	// Verify that the provided symbol is an interface
	if !isInterface(iface) {
		return false, fmt.Errorf("symbol %s is not an interface", iface.Name)
	}

	// Get the method set of the interface
	ifaceMethods := getInterfaceMethods(iface)
	if len(ifaceMethods) == 0 {
		// Empty interface is implemented by all types
		return true, nil
	}

	// Get the method set of the type
	typMethods := getTypeMethods(typ)

	// Check if all interface methods are implemented by the type
	for _, ifaceMethod := range ifaceMethods {
		found := false
		for _, typMethod := range typMethods {
			if isMethodCompatible(ifaceMethod, typMethod) {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

// GetImplementationInfo gets detailed information about how a type implements an interface.
func (f *InterfaceFinder) GetImplementationInfo(iface, typ *typesys.Symbol) (*ImplementationInfo, error) {
	if f.Module == nil || iface == nil || typ == nil {
		return nil, fmt.Errorf("invalid parameters")
	}

	// Check if it's an implementation
	isImpl, err := f.IsImplementedBy(iface, typ)
	if err != nil {
		return nil, err
	}

	if !isImpl {
		return nil, fmt.Errorf("type %s does not implement interface %s", typ.Name, iface.Name)
	}

	// Create the implementation info
	info := &ImplementationInfo{
		Type:      typ,
		Interface: iface,
		MethodMap: make(map[string]MethodImplementation),
	}

	// Get the method sets
	ifaceMethods := getInterfaceMethods(iface)
	typMethods := getTypeMethods(typ)

	// Fill the method map
	for _, ifaceMethod := range ifaceMethods {
		for _, typMethod := range typMethods {
			if isMethodCompatible(ifaceMethod, typMethod) {
				info.MethodMap[ifaceMethod.Name] = MethodImplementation{
					InterfaceMethod:    ifaceMethod,
					ImplementingMethod: typMethod,
					IsDirectMatch:      ifaceMethod.Name == typMethod.Name,
				}
				break
			}
		}
	}

	// Determine if the implementation is through embedding
	info.IsEmbedded = isImplementationThroughEmbedding(typ, iface)
	if info.IsEmbedded {
		info.EmbeddedPath = findEmbeddingPath(typ, iface)
	}

	return info, nil
}

// GetAllImplementedInterfaces finds all interfaces implemented by a type.
func (f *InterfaceFinder) GetAllImplementedInterfaces(typ *typesys.Symbol) ([]*typesys.Symbol, error) {
	if f.Module == nil || typ == nil {
		return nil, fmt.Errorf("invalid parameters")
	}

	// Get all interfaces in the module
	interfaces := getAllInterfaces(f.Module)

	// Check each interface
	var implemented []*typesys.Symbol
	for _, iface := range interfaces {
		isImpl, err := f.IsImplementedBy(iface, typ)
		if err != nil {
			continue
		}

		if isImpl {
			implemented = append(implemented, iface)
		}
	}

	return implemented, nil
}

// Helper functions

// isInterface checks if a symbol is an interface.
func isInterface(sym *typesys.Symbol) bool {
	// Check if the symbol is an interface type
	return sym != nil && sym.Kind == typesys.KindInterface
}

// getInterfaceMethods gets all methods defined by an interface.
func getInterfaceMethods(iface *typesys.Symbol) []*typesys.Symbol {
	if iface == nil || iface.Kind != typesys.KindInterface {
		return nil
	}

	// Use a map to avoid duplicate methods in case of diamond inheritance
	methodMap := make(map[string]*typesys.Symbol)

	// Track visited nodes to avoid cycles
	visited := make(map[string]bool)

	// Helper function for recursive traversal
	var collectMethods func(current *typesys.Symbol)
	collectMethods = func(current *typesys.Symbol) {
		if current == nil || visited[current.ID] {
			return
		}
		visited[current.ID] = true

		// Add methods directly defined on this interface
		for _, ref := range current.References {
			if ref.Symbol == nil {
				continue
			}

			// If the reference is a method defined on this interface
			if ref.Context == current && ref.Symbol.Kind == typesys.KindMethod {
				methodMap[ref.Symbol.Name] = ref.Symbol
			}

			// If the reference is an embedded interface
			if ref.Context == current && ref.Symbol.Kind == typesys.KindInterface {
				// Recursively collect methods from the embedded interface
				collectMethods(ref.Symbol)
			}
		}
	}

	// Start collection from the root interface
	collectMethods(iface)

	// Convert the map to a slice
	var methods []*typesys.Symbol
	for _, method := range methodMap {
		methods = append(methods, method)
	}

	return methods
}

// getTypeMethods gets all methods defined by a type.
func getTypeMethods(typ *typesys.Symbol) []*typesys.Symbol {
	if typ == nil {
		return nil
	}

	// Use a map to avoid duplicate methods
	methodMap := make(map[string]*typesys.Symbol)

	// Track visited nodes to avoid cycles
	visited := make(map[string]bool)

	// Helper function for recursive traversal
	var collectMethods func(current *typesys.Symbol)
	collectMethods = func(current *typesys.Symbol) {
		if current == nil || visited[current.ID] {
			return
		}
		visited[current.ID] = true

		// Add methods directly defined on this type
		for _, ref := range current.References {
			if ref.Symbol == nil {
				continue
			}

			// If the reference is a method defined on this type
			if ref.Context == current && ref.Symbol.Kind == typesys.KindMethod && ref.Symbol.Parent == current {
				methodMap[ref.Symbol.Name] = ref.Symbol
			}

			// If the reference is an embedded type (struct or interface)
			if ref.Context == current && (ref.Symbol.Kind == typesys.KindStruct || ref.Symbol.Kind == typesys.KindInterface) {
				// Recursively collect methods from the embedded type
				collectMethods(ref.Symbol)
			}
		}
	}

	// Start collection from the root type
	collectMethods(typ)

	// Convert the map to a slice
	var methods []*typesys.Symbol
	for _, method := range methodMap {
		methods = append(methods, method)
	}

	return methods
}

// isMethodCompatible checks if a type method is compatible with an interface method.
func isMethodCompatible(ifaceMethod, typMethod *typesys.Symbol) bool {
	// This is a simplified check - in a full implementation, we would also check
	// parameter types, return values, etc.
	return ifaceMethod != nil && typMethod != nil && ifaceMethod.Name == typMethod.Name
}

// isImplementationThroughEmbedding checks if a type implements an interface through embedding.
func isImplementationThroughEmbedding(typ, iface *typesys.Symbol) bool {
	if typ == nil || iface == nil {
		return false
	}

	// Check if the type directly embeds the interface
	for _, ref := range typ.References {
		if ref.Symbol == iface && ref.Context == typ {
			return true
		}
	}

	// Check if any embedded type implements the interface
	visited := make(map[string]bool)
	var checkEmbedded func(current *typesys.Symbol) bool
	checkEmbedded = func(current *typesys.Symbol) bool {
		if current == nil || visited[current.ID] {
			return false
		}
		visited[current.ID] = true

		for _, ref := range current.References {
			if ref.Symbol == nil || ref.Context != current {
				continue
			}

			// If we find the interface embedded anywhere in the type hierarchy
			if ref.Symbol == iface {
				return true
			}

			// Check embedded types recursively
			if ref.Symbol.Kind == typesys.KindStruct || ref.Symbol.Kind == typesys.KindInterface {
				if checkEmbedded(ref.Symbol) {
					return true
				}
			}
		}
		return false
	}

	return checkEmbedded(typ)
}

// findEmbeddingPath finds the path of embedded types that implement the interface.
func findEmbeddingPath(typ, iface *typesys.Symbol) []*typesys.Symbol {
	if typ == nil || iface == nil {
		return nil
	}

	// Using a simpler approach here - in a full implementation we'd use a more
	// sophisticated graph traversal algorithm
	var path []*typesys.Symbol

	// Check for direct embedding
	for _, ref := range typ.References {
		if ref.Symbol == iface && ref.Context == typ {
			path = append(path, iface)
			return path
		}
	}

	// Check for indirect embedding (simplified)
	visited := make(map[string]bool)
	var findPath func(current *typesys.Symbol) []*typesys.Symbol
	findPath = func(current *typesys.Symbol) []*typesys.Symbol {
		if current == nil || visited[current.ID] {
			return nil
		}
		visited[current.ID] = true

		for _, ref := range current.References {
			if ref.Symbol == nil || ref.Context != current {
				continue
			}

			// If we found the interface
			if ref.Symbol == iface {
				return []*typesys.Symbol{ref.Symbol}
			}

			// Check embedded types
			if ref.Symbol.Kind == typesys.KindStruct || ref.Symbol.Kind == typesys.KindInterface {
				subPath := findPath(ref.Symbol)
				if len(subPath) > 0 {
					return append([]*typesys.Symbol{ref.Symbol}, subPath...)
				}
			}
		}
		return nil
	}

	return findPath(typ)
}

// getAllInterfaces gets all interfaces defined in the module.
func getAllInterfaces(module *typesys.Module) []*typesys.Symbol {
	if module == nil {
		return nil
	}

	var interfaces []*typesys.Symbol
	for _, pkg := range module.Packages {
		for _, sym := range pkg.Symbols {
			if sym != nil && sym.Kind == typesys.KindInterface {
				interfaces = append(interfaces, sym)
			}
		}
	}

	return interfaces
}

// getEligibleTypes gets all types that are eligible for interface implementation search.
func (f *InterfaceFinder) getEligibleTypes(opts *FindOptions) []*typesys.Symbol {
	if f.Module == nil {
		return nil
	}

	// Create a package filter if specified
	pkgFilter := make(map[string]bool)
	if len(opts.Packages) > 0 {
		for _, pkgPath := range opts.Packages {
			pkgFilter[pkgPath] = true
		}
	}

	var types []*typesys.Symbol
	for pkgPath, pkg := range f.Module.Packages {
		// Skip this package if not in the filter
		if len(pkgFilter) > 0 && !pkgFilter[pkgPath] {
			continue
		}

		// Collect types from this package
		for _, sym := range pkg.Symbols {
			if sym == nil {
				continue
			}

			// Skip non-struct types (interfaces can't implement interfaces in Go)
			if sym.Kind != typesys.KindStruct {
				continue
			}

			// Skip unexported types if requested
			if opts.ExportedOnly && !sym.Exported {
				continue
			}

			// Skip generic types if requested - for now, we don't have this info
			// if !opts.IncludeGenerics && sym.IsGeneric {
			//     continue
			// }

			types = append(types, sym)
		}
	}

	return types
}
