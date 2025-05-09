package service

import (
	"fmt"
	"go/types"
	"sort"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TypeDifference represents a difference between two type versions
type TypeDifference struct {
	FieldName string
	OldType   string
	NewType   string
	Kind      DifferenceKind
}

// DifferenceKind represents the kind of difference between types
type DifferenceKind string

const (
	// Field added in newer version
	FieldAdded DifferenceKind = "added"

	// Field removed in newer version
	FieldRemoved DifferenceKind = "removed"

	// Field type changed
	FieldTypeChanged DifferenceKind = "type_changed"

	// Method signature changed
	MethodSignatureChanged DifferenceKind = "method_signature_changed"

	// Interface requirements changed
	InterfaceRequirementsChanged DifferenceKind = "interface_requirements_changed"
)

// CompatibilityReport contains the result of a compatibility analysis
type CompatibilityReport struct {
	TypeName    string
	Versions    []string
	Compatible  bool
	Differences []TypeDifference
}

// VersionPolicy represents different strategies for resolving version conflicts
type VersionPolicy int

const (
	// Use version from the module where the operation started
	FromCallingModule VersionPolicy = iota

	// Use the latest version available
	PreferLatest

	// Treat different versions as distinct types (most accurate)
	VersionSpecific

	// Try to reconcile across versions when possible
	Reconcile
)

// AnalyzeTypeCompatibility determines if types across versions are compatible
func (s *Service) AnalyzeTypeCompatibility(importPath string, typeName string) *CompatibilityReport {
	// Find all versions of this type
	typeVersions := s.FindTypeAcrossModules(importPath, typeName)

	// Create a report
	report := &CompatibilityReport{
		TypeName: typeName,
		Versions: make([]string, 0, len(typeVersions)),
	}

	// No versions found
	if len(typeVersions) == 0 {
		return report
	}

	// Only one version found - always compatible with itself
	if len(typeVersions) == 1 {
		for modPath := range typeVersions {
			report.Versions = append(report.Versions, modPath)
		}
		report.Compatible = true
		return report
	}

	// Multiple versions - we need to compare them
	var baseType *typesys.Symbol
	var baseModPath string

	// Get base type (first one alphabetically for stable comparison)
	paths := make([]string, 0, len(typeVersions))
	for path := range typeVersions {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	baseModPath = paths[0]
	baseType = typeVersions[baseModPath]

	// Add versions to report
	report.Versions = paths

	// Compare base type with all other versions
	for _, modPath := range paths[1:] {
		otherType := typeVersions[modPath]

		// Compare the two types
		diffs := compareTypes(baseType, otherType)

		// Add differences to report
		report.Differences = append(report.Differences, diffs...)
	}

	// If there are no differences, types are compatible
	report.Compatible = len(report.Differences) == 0

	return report
}

// compareTypes compares two types and returns their differences
func compareTypes(baseType, otherType *typesys.Symbol) []TypeDifference {
	var differences []TypeDifference

	// Get the actual Go types
	baseTypeObj := baseType.TypeInfo
	otherTypeObj := otherType.TypeInfo

	// If either type is nil, we can't compare
	if baseTypeObj == nil || otherTypeObj == nil {
		return []TypeDifference{
			{
				Kind:    FieldTypeChanged,
				OldType: fmt.Sprintf("%T", baseTypeObj),
				NewType: fmt.Sprintf("%T", otherTypeObj),
			},
		}
	}

	// Based on the kind of type, do different comparisons
	switch baseType.Kind {
	case typesys.KindStruct:
		return compareStructs(baseType, otherType)
	case typesys.KindInterface:
		return compareInterfaces(baseType, otherType)
	default:
		// For other types, just compare their string representation
		if baseTypeObj.String() != otherTypeObj.String() {
			differences = append(differences, TypeDifference{
				Kind:    FieldTypeChanged,
				OldType: baseTypeObj.String(),
				NewType: otherTypeObj.String(),
			})
		}
	}

	return differences
}

// compareStructs compares two struct types for compatibility
func compareStructs(baseType, otherType *typesys.Symbol) []TypeDifference {
	var differences []TypeDifference

	// Get the underlying struct types
	baseStruct, ok1 := baseType.TypeInfo.Underlying().(*types.Struct)
	otherStruct, ok2 := otherType.TypeInfo.Underlying().(*types.Struct)

	if !ok1 || !ok2 {
		return []TypeDifference{{
			Kind:    FieldTypeChanged,
			OldType: fmt.Sprintf("%T", baseType.TypeInfo),
			NewType: fmt.Sprintf("%T", otherType.TypeInfo),
		}}
	}

	// Create maps of fields by name for easier comparison
	baseFields := makeFieldMap(baseStruct)
	otherFields := makeFieldMap(otherStruct)

	// Check for fields in base that don't exist in other (removed fields)
	for name, field := range baseFields {
		if _, exists := otherFields[name]; !exists {
			differences = append(differences, TypeDifference{
				FieldName: name,
				OldType:   field.Type.String(),
				NewType:   "",
				Kind:      FieldRemoved,
			})
		}
	}

	// Check for fields in other that don't exist in base (added fields)
	for name, field := range otherFields {
		if _, exists := baseFields[name]; !exists {
			differences = append(differences, TypeDifference{
				FieldName: name,
				OldType:   "",
				NewType:   field.Type.String(),
				Kind:      FieldAdded,
			})
		}
	}

	// Compare fields that exist in both
	for name, baseField := range baseFields {
		if otherField, exists := otherFields[name]; exists {
			// Compare field types
			if !typesAreEqual(baseField.Type, otherField.Type) {
				differences = append(differences, TypeDifference{
					FieldName: name,
					OldType:   baseField.Type.String(),
					NewType:   otherField.Type.String(),
					Kind:      FieldTypeChanged,
				})
			}

			// Compare field tags
			if baseField.Tag != otherField.Tag {
				differences = append(differences, TypeDifference{
					FieldName: name + " (tag)",
					OldType:   baseField.Tag,
					NewType:   otherField.Tag,
					Kind:      FieldTypeChanged,
				})
			}

			// Check if field visibility changed (exported vs unexported)
			if baseField.Exported != otherField.Exported {
				var oldVisibility, newVisibility string
				if baseField.Exported {
					oldVisibility = "exported"
					newVisibility = "unexported"
				} else {
					oldVisibility = "unexported"
					newVisibility = "exported"
				}

				differences = append(differences, TypeDifference{
					FieldName: name + " (visibility)",
					OldType:   oldVisibility,
					NewType:   newVisibility,
					Kind:      FieldTypeChanged,
				})
			}
		}
	}

	return differences
}

// makeFieldMap creates a map of field information from a struct type
func makeFieldMap(structType *types.Struct) map[string]struct {
	Type     types.Type
	Tag      string
	Exported bool
} {
	fields := make(map[string]struct {
		Type     types.Type
		Tag      string
		Exported bool
	})

	// Process all fields, including those from embedded structs
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tag := structType.Tag(i)

		// Skip methods, only process fields
		if !field.IsField() {
			continue
		}

		fieldName := field.Name()

		// If field is embedded and is a struct, need special handling
		if field.Embedded() {
			// For embedded fields, we need to check if it's a struct type
			// and process its fields recursively
			if embeddedStruct, ok := field.Type().Underlying().(*types.Struct); ok {
				embeddedFields := makeFieldMap(embeddedStruct)

				// Add embedded fields to our map with proper qualification
				for embName, embField := range embeddedFields {
					// Skip if there's already a field with this name (field from embedding struct takes precedence)
					if _, exists := fields[embName]; !exists {
						fields[embName] = embField
					}
				}
			}
		}

		// Add this field to our map
		fields[fieldName] = struct {
			Type     types.Type
			Tag      string
			Exported bool
		}{
			Type:     field.Type(),
			Tag:      tag,
			Exported: field.Exported(),
		}
	}

	return fields
}

// typesAreEqual performs a deeper comparison of types beyond just their string representation
func typesAreEqual(t1, t2 types.Type) bool {
	// For basic types, comparing the string representation is sufficient
	if types.Identical(t1, t2) {
		return true
	}

	// For more complex types, additional checks may be needed
	switch t1u := t1.Underlying().(type) {
	case *types.Struct:
		// Compare structs field by field
		if t2u, ok := t2.Underlying().(*types.Struct); ok {
			if t1u.NumFields() != t2u.NumFields() {
				return false
			}

			// This is a simplified check, a more robust solution would
			// recursively compare each field
			for i := 0; i < t1u.NumFields(); i++ {
				f1 := t1u.Field(i)
				f2 := t2u.Field(i)

				if f1.Name() != f2.Name() ||
					!typesAreEqual(f1.Type(), f2.Type()) ||
					t1u.Tag(i) != t2u.Tag(i) {
					return false
				}
			}
			return true
		}
		return false

	case *types.Interface:
		// Compare interfaces method by method
		if t2u, ok := t2.Underlying().(*types.Interface); ok {
			if t1u.NumMethods() != t2u.NumMethods() {
				return false
			}

			// This is a simplified check, a more robust solution would
			// compare method signatures in detail
			for i := 0; i < t1u.NumMethods(); i++ {
				m1 := t1u.Method(i)
				m2 := t2u.Method(i)

				if m1.Name() != m2.Name() ||
					!typesAreEqual(m1.Type(), m2.Type()) {
					return false
				}
			}
			return true
		}
		return false

	case *types.Slice:
		// Compare slice element types
		if t2u, ok := t2.Underlying().(*types.Slice); ok {
			return typesAreEqual(t1u.Elem(), t2u.Elem())
		}
		return false

	case *types.Array:
		// Compare array length and element types
		if t2u, ok := t2.Underlying().(*types.Array); ok {
			return t1u.Len() == t2u.Len() && typesAreEqual(t1u.Elem(), t2u.Elem())
		}
		return false

	case *types.Map:
		// Compare map key and value types
		if t2u, ok := t2.Underlying().(*types.Map); ok {
			return typesAreEqual(t1u.Key(), t2u.Key()) && typesAreEqual(t1u.Elem(), t2u.Elem())
		}
		return false

	case *types.Chan:
		// Compare channel direction and element type
		if t2u, ok := t2.Underlying().(*types.Chan); ok {
			return t1u.Dir() == t2u.Dir() && typesAreEqual(t1u.Elem(), t2u.Elem())
		}
		return false

	case *types.Pointer:
		// Compare pointer element types
		if t2u, ok := t2.Underlying().(*types.Pointer); ok {
			return typesAreEqual(t1u.Elem(), t2u.Elem())
		}
		return false

	default:
		// For other types, fallback to string comparison
		return t1.String() == t2.String()
	}
}

// compareInterfaces compares two interface types for compatibility
func compareInterfaces(baseType, otherType *typesys.Symbol) []TypeDifference {
	var differences []TypeDifference

	// Get the underlying interface types
	baseIface, ok1 := baseType.TypeInfo.Underlying().(*types.Interface)
	otherIface, ok2 := otherType.TypeInfo.Underlying().(*types.Interface)

	if !ok1 || !ok2 {
		return []TypeDifference{{
			Kind:    InterfaceRequirementsChanged,
			OldType: fmt.Sprintf("%T", baseType.TypeInfo),
			NewType: fmt.Sprintf("%T", otherType.TypeInfo),
		}}
	}

	// Create maps of methods by name for easier comparison
	baseMethods := makeMethodMap(baseIface)
	otherMethods := makeMethodMap(otherIface)

	// Check for methods in base that don't exist in other (removed methods)
	for name, method := range baseMethods {
		if _, exists := otherMethods[name]; !exists {
			differences = append(differences, TypeDifference{
				FieldName: name + " (method)",
				OldType:   method.Type.String(),
				NewType:   "",
				Kind:      MethodSignatureChanged,
			})
		}
	}

	// Check for methods in other that don't exist in base (added methods)
	for name, method := range otherMethods {
		if _, exists := baseMethods[name]; !exists {
			differences = append(differences, TypeDifference{
				FieldName: name + " (method)",
				OldType:   "",
				NewType:   method.Type.String(),
				Kind:      MethodSignatureChanged,
			})
		}
	}

	// Compare methods that exist in both
	for name, baseMethod := range baseMethods {
		if otherMethod, exists := otherMethods[name]; exists {
			// Check if method signatures are compatible
			if !methodSignaturesCompatible(baseMethod.Type, otherMethod.Type) {
				differences = append(differences, TypeDifference{
					FieldName: name + " (signature)",
					OldType:   baseMethod.Type.String(),
					NewType:   otherMethod.Type.String(),
					Kind:      MethodSignatureChanged,
				})
			}
		}
	}

	return differences
}

// makeMethodMap creates a map of method information from an interface type
func makeMethodMap(ifaceType *types.Interface) map[string]struct {
	Type     *types.Signature
	Position int
} {
	methods := make(map[string]struct {
		Type     *types.Signature
		Position int
	})

	// Process all methods
	for i := 0; i < ifaceType.NumMethods(); i++ {
		method := ifaceType.Method(i)
		methodName := method.Name()

		// Get the method signature
		signature, ok := method.Type().(*types.Signature)
		if !ok {
			// This shouldn't happen for interface methods, but let's be safe
			continue
		}

		// Add to the method map
		methods[methodName] = struct {
			Type     *types.Signature
			Position int
		}{
			Type:     signature,
			Position: i,
		}
	}

	// Handle embedded interfaces
	for i := 0; i < ifaceType.NumEmbeddeds(); i++ {
		embedded := ifaceType.EmbeddedType(i)

		// If it's an interface, get its methods
		if embeddedIface, ok := embedded.Underlying().(*types.Interface); ok {
			embeddedMethods := makeMethodMap(embeddedIface)

			// Add embedded methods to our map
			for name, method := range embeddedMethods {
				// Only add if the method doesn't already exist (method from embedding interface takes precedence)
				if _, exists := methods[name]; !exists {
					methods[name] = method
				}
			}
		}
	}

	return methods
}

// methodSignaturesCompatible checks if two method signatures are compatible
// This follows Go's rules for method set checking and interface satisfaction
func methodSignaturesCompatible(sig1, sig2 *types.Signature) bool {
	// Check if identical
	if types.Identical(sig1, sig2) {
		return true
	}

	// Compare receiver parameters (for methods) - not strictly needed for interfaces
	if (sig1.Recv() == nil) != (sig2.Recv() == nil) {
		return false
	}

	// Compare parameters
	params1 := sig1.Params()
	params2 := sig2.Params()
	if params1.Len() != params2.Len() {
		return false
	}

	// Check each parameter
	for i := 0; i < params1.Len(); i++ {
		param1 := params1.At(i)
		param2 := params2.At(i)

		// Parameter types must be identical in interfaces
		if !types.Identical(param1.Type(), param2.Type()) {
			return false
		}
	}

	// Compare return values
	results1 := sig1.Results()
	results2 := sig2.Results()
	if results1.Len() != results2.Len() {
		return false
	}

	// Check each result
	for i := 0; i < results1.Len(); i++ {
		result1 := results1.At(i)
		result2 := results2.At(i)

		// Result types must be identical in interfaces
		if !types.Identical(result1.Type(), result2.Type()) {
			return false
		}
	}

	// Check variadic status
	if sig1.Variadic() != sig2.Variadic() {
		return false
	}

	return true
}

// FindReferences finds all references to a symbol using a specific version policy
func (s *Service) FindReferences(symbol *typesys.Symbol, policy VersionPolicy) ([]*typesys.Reference, error) {
	var allReferences []*typesys.Reference

	// Get the module containing this symbol
	var containingModule *typesys.Module
	for _, mod := range s.Modules {
		for _, pkg := range mod.Packages {
			if pkg.Symbols[symbol.ID] == symbol {
				containingModule = mod
				break
			}
		}
		if containingModule != nil {
			break
		}
	}

	if containingModule == nil {
		return nil, fmt.Errorf("symbol %s not found in any module", symbol.Name)
	}

	// Different behavior based on policy
	switch policy {
	case FromCallingModule:
		// Only find references in the containing module
		idx := s.Indices[containingModule.Path]
		if idx != nil {
			allReferences = idx.FindReferences(symbol)
		}

	case PreferLatest:
		// Find references in all modules, but prioritize latest
		for _, idx := range s.Indices {
			refs := idx.FindReferences(symbol)
			allReferences = append(allReferences, refs...)
		}

	case VersionSpecific:
		// Only find references in the containing module
		idx := s.Indices[containingModule.Path]
		if idx != nil {
			allReferences = idx.FindReferences(symbol)
		}

	case Reconcile:
		// Find references to similarly named symbols in all modules
		for _, idx := range s.Indices {
			// Find similar symbols first
			similarSymbols := idx.FindSymbolsByName(symbol.Name)
			for _, sym := range similarSymbols {
				// Only consider symbols of the same kind
				if sym.Kind == symbol.Kind {
					refs := idx.FindReferences(sym)
					allReferences = append(allReferences, refs...)
				}
			}
		}
	}

	return allReferences, nil
}
