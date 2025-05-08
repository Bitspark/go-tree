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

	// For proper struct comparison, we'd need to access the struct fields
	// This is a simplified version that assumes the symbols have field information

	// In a real implementation, this would be much more comprehensive
	// using type reflection to compare struct fields in detail

	// Just check if their string representations are different for now
	if baseType.TypeInfo.String() != otherType.TypeInfo.String() {
		differences = append(differences, TypeDifference{
			Kind:    FieldTypeChanged,
			OldType: baseType.TypeInfo.String(),
			NewType: otherType.TypeInfo.String(),
		})
	}

	return differences
}

// compareInterfaces compares two interface types for compatibility
func compareInterfaces(baseType, otherType *typesys.Symbol) []TypeDifference {
	var differences []TypeDifference

	// For proper interface comparison, we'd need to compare method sets
	// This is a simplified version that assumes the symbols have method information

	// In a real implementation, this would be much more comprehensive
	// using type reflection to compare interface method sets in detail

	baseIface, ok1 := baseType.TypeInfo.(*types.Interface)
	otherIface, ok2 := otherType.TypeInfo.(*types.Interface)

	if !ok1 || !ok2 {
		return []TypeDifference{{
			Kind:    InterfaceRequirementsChanged,
			OldType: fmt.Sprintf("%T", baseType.TypeInfo),
			NewType: fmt.Sprintf("%T", otherType.TypeInfo),
		}}
	}

	// Compare method counts as a simple heuristic
	if baseIface.NumMethods() != otherIface.NumMethods() {
		differences = append(differences, TypeDifference{
			Kind:    InterfaceRequirementsChanged,
			OldType: fmt.Sprintf("methods: %d", baseIface.NumMethods()),
			NewType: fmt.Sprintf("methods: %d", otherIface.NumMethods()),
		})
	}

	return differences
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
