package service

import (
	"fmt"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// SemverImpact represents the impact level of a change according to semver rules
type SemverImpact string

const (
	// NoImpact means the change doesn't affect compatibility
	NoImpact SemverImpact = "none"

	// PatchImpact is a non-breaking change that only fixes bugs (0.0.X)
	PatchImpact SemverImpact = "patch"

	// MinorImpact is a non-breaking change that adds functionality (0.X.0)
	MinorImpact SemverImpact = "minor"

	// MajorImpact is a breaking change that requires client code modification (X.0.0)
	MajorImpact SemverImpact = "major"
)

// CompatibilityReport contains information about compatibility between two versions
type SemverCompatibilityReport struct {
	// The name of the type being compared
	TypeName string

	// Versions that were compared
	OldVersion string
	NewVersion string

	// Overall impact level
	Impact SemverImpact

	// Detailed differences
	Differences []TypeDifference

	// Is backward compatible (no major impact)
	IsBackwardCompatible bool

	// Score from 0-100 (percentage of compatible APIs)
	CompatibilityScore int

	// Suggestions for fixing incompatibilities
	Suggestions []string
}

// AnalyzeSemverCompatibility performs a semver-based compatibility analysis between two types
func (s *Service) AnalyzeSemverCompatibility(importPath, typeName, oldVersion, newVersion string) (*SemverCompatibilityReport, error) {
	// Find the symbols in the respective versions
	oldType, err := s.FindSymbolInModuleVersion(importPath, typeName, oldVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find old version: %w", err)
	}

	newType, err := s.FindSymbolInModuleVersion(importPath, typeName, newVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find new version: %w", err)
	}

	// Build the report
	report := &SemverCompatibilityReport{
		TypeName:             typeName,
		OldVersion:           oldVersion,
		NewVersion:           newVersion,
		Impact:               NoImpact,
		Differences:          compareTypes(oldType, newType),
		IsBackwardCompatible: true,
		CompatibilityScore:   100,
		Suggestions:          []string{},
	}

	// Analyze the differences to determine impact
	report.Impact = determineSemverImpact(report.Differences)
	report.IsBackwardCompatible = (report.Impact != MajorImpact)
	report.CompatibilityScore = calculateCompatibilityScore(report.Differences)
	report.Suggestions = generateSuggestions(report.Differences)

	return report, nil
}

// FindSymbolInModuleVersion finds a specific symbol in a specific module version
func (s *Service) FindSymbolInModuleVersion(importPath, typeName, version string) (*typesys.Symbol, error) {
	// Find all modules with matching import paths
	var moduleMatches []*typesys.Module
	for _, mod := range s.Modules {
		if strings.Contains(mod.Path, version) {
			moduleMatches = append(moduleMatches, mod)
		}
	}

	if len(moduleMatches) == 0 {
		return nil, fmt.Errorf("no modules found with version %s", version)
	}

	// Look for the symbol in all matching modules
	for _, mod := range moduleMatches {
		for _, pkg := range mod.Packages {
			if pkg.ImportPath == importPath || strings.HasSuffix(pkg.ImportPath, importPath) {
				for _, sym := range pkg.Symbols {
					if sym.Name == typeName {
						return sym, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("symbol %s not found in module version %s", typeName, version)
}

// determineSemverImpact calculates the overall semver impact of the changes
func determineSemverImpact(diffs []TypeDifference) SemverImpact {
	impact := NoImpact

	for _, diff := range diffs {
		diffImpact := determineDifferenceImpact(diff)
		impact = maxImpact(impact, diffImpact)
	}

	return impact
}

// determineDifferenceImpact classifies a single difference based on semver rules
func determineDifferenceImpact(diff TypeDifference) SemverImpact {
	switch diff.Kind {
	case FieldAdded:
		// Adding exported fields to struct is typically a minor change
		// unless it's an interface method, which is a major change
		if strings.Contains(diff.FieldName, "(method)") {
			return MajorImpact
		}
		return MinorImpact

	case FieldRemoved:
		// Removing anything is a major change
		return MajorImpact

	case FieldTypeChanged:
		// Type changes are generally major changes
		// But might be minor in special cases (widening conversion)
		if isWideningTypeChange(diff.OldType, diff.NewType) {
			return MinorImpact
		}
		return MajorImpact

	case MethodSignatureChanged:
		// Method signature changes are always major changes
		return MajorImpact

	case InterfaceRequirementsChanged:
		// Any change to interface requirements is a major change
		return MajorImpact

	default:
		// Unknown changes are considered as patch changes
		return PatchImpact
	}
}

// isWideningTypeChange checks if a type change is widening (non-breaking)
// For example, int32 to int64 is a widening change
func isWideningTypeChange(oldType, newType string) bool {
	// Define pairs of types where changing from old to new is widening
	wideningPairs := map[string][]string{
		"int8":    {"int16", "int32", "int64", "int", "float32", "float64"},
		"int16":   {"int32", "int64", "int", "float32", "float64"},
		"int32":   {"int64", "float64"},
		"int":     {"int64", "float64"},
		"float32": {"float64"},
		"uint8":   {"uint16", "uint32", "uint64", "uint", "int16", "int32", "int64", "int", "float32", "float64"},
		"uint16":  {"uint32", "uint64", "uint", "int32", "int64", "int", "float32", "float64"},
		"uint32":  {"uint64", "int64", "float64"},
		"uint":    {"uint64", "float64"},
	}

	if wideningTypes, ok := wideningPairs[oldType]; ok {
		for _, wideType := range wideningTypes {
			if newType == wideType {
				return true
			}
		}
	}

	return false
}

// calculateCompatibilityScore calculates a score from 0-100 indicating compatibility
func calculateCompatibilityScore(diffs []TypeDifference) int {
	if len(diffs) == 0 {
		return 100
	}

	// Count major breaking changes
	majorChanges := 0
	for _, diff := range diffs {
		if determineDifferenceImpact(diff) == MajorImpact {
			majorChanges++
		}
	}

	// Simple formula: 100 - (% of major changes)
	score := 100 - (majorChanges * 100 / len(diffs))

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateSuggestions creates hints to fix incompatibilities
func generateSuggestions(diffs []TypeDifference) []string {
	var suggestions []string

	for _, diff := range diffs {
		switch diff.Kind {
		case FieldRemoved:
			suggestions = append(suggestions,
				fmt.Sprintf("Consider adding back field '%s' for backward compatibility", diff.FieldName))

		case FieldTypeChanged:
			suggestions = append(suggestions,
				fmt.Sprintf("Type change in '%s' from '%s' to '%s' is breaking. Consider keeping the original type or providing a conversion",
					diff.FieldName, diff.OldType, diff.NewType))

		case MethodSignatureChanged:
			suggestions = append(suggestions,
				fmt.Sprintf("Method signature change in '%s' is breaking. Consider adding an adapter or keeping the original method",
					diff.FieldName))

		case InterfaceRequirementsChanged:
			suggestions = append(suggestions,
				"Interface changes are breaking. Consider creating a new interface that extends the old one")
		}
	}

	return suggestions
}

// maxImpact returns the maximum of two impact levels
// Major > Minor > Patch > None
func maxImpact(a, b SemverImpact) SemverImpact {
	if a == MajorImpact || b == MajorImpact {
		return MajorImpact
	}
	if a == MinorImpact || b == MinorImpact {
		return MinorImpact
	}
	if a == PatchImpact || b == PatchImpact {
		return PatchImpact
	}
	return NoImpact
}
