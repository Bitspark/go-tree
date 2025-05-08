package service

import (
	"go/types"
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
)

// TestAnalyzeTypeCompatibility tests compatibility analysis between type versions
func TestAnalyzeTypeCompatibility(t *testing.T) {
	// Create a service with mock modules containing similar types
	service := &Service{
		Modules: map[string]*typesys.Module{
			"mod1": {
				Path: "mod1",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {
						ImportPath: "pkg/foo",
						Symbols: map[string]*typesys.Symbol{
							"sym1": {
								ID:       "sym1",
								Name:     "MyType",
								Kind:     typesys.KindStruct,
								TypeInfo: types.NewStruct([]*types.Var{}, []string{}),
							},
						},
					},
				},
			},
			"mod2": {
				Path: "mod2",
				Packages: map[string]*typesys.Package{
					"pkg/foo": {
						ImportPath: "pkg/foo",
						Symbols: map[string]*typesys.Symbol{
							"sym2": {
								ID:       "sym2",
								Name:     "MyType",
								Kind:     typesys.KindStruct,
								TypeInfo: types.NewStruct([]*types.Var{}, []string{}),
							},
						},
					},
				},
			},
		},
	}

	// Test analyzing compatibility of identical types
	report := service.AnalyzeTypeCompatibility("pkg/foo", "MyType")
	if !report.Compatible {
		t.Errorf("Expected identical types to be compatible")
	}
	if len(report.Differences) != 0 {
		t.Errorf("Expected no differences between identical types, got %d", len(report.Differences))
	}
	if len(report.Versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(report.Versions))
	}
}

// TestCompareTypes tests comparing different types for compatibility
func TestCompareTypes(t *testing.T) {
	// Test comparing different types
	baseType := &typesys.Symbol{
		ID:       "base",
		Name:     "BaseType",
		Kind:     typesys.KindStruct,
		TypeInfo: types.NewStruct([]*types.Var{}, []string{}),
	}

	// Same type (should be compatible)
	sameType := &typesys.Symbol{
		ID:       "same",
		Name:     "SameType",
		Kind:     typesys.KindStruct,
		TypeInfo: types.NewStruct([]*types.Var{}, []string{}),
	}

	// Different type (should be incompatible)
	differentType := &typesys.Symbol{
		ID:   "diff",
		Name: "DiffType",
		Kind: typesys.KindInterface,
		TypeInfo: types.NewInterface(
			[]*types.Func{},
			[]*types.Named{},
		),
	}

	// Test comparing same type
	diffs := compareTypes(baseType, sameType)
	if len(diffs) != 0 {
		t.Errorf("Expected no differences between same types, got %d", len(diffs))
	}

	// Test comparing different types
	diffs = compareTypes(baseType, differentType)
	if len(diffs) == 0 {
		t.Errorf("Expected differences between different types")
	}
}

// TestCompareInterfaces tests comparing interface types
func TestCompareInterfaces(t *testing.T) {
	// Create two interface types with different method counts
	baseIface := types.NewInterface(
		[]*types.Func{},
		[]*types.Named{},
	)

	otherIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, nil, "Method1", types.NewSignature(nil, nil, nil, false)),
		},
		[]*types.Named{},
	)

	baseType := &typesys.Symbol{
		ID:       "base",
		Name:     "BaseIface",
		Kind:     typesys.KindInterface,
		TypeInfo: baseIface,
	}

	otherType := &typesys.Symbol{
		ID:       "other",
		Name:     "OtherIface",
		Kind:     typesys.KindInterface,
		TypeInfo: otherIface,
	}

	// Test comparing interfaces with different method counts
	diffs := compareInterfaces(baseType, otherType)
	if len(diffs) == 0 {
		t.Errorf("Expected differences between interfaces with different method counts")
	}

	// Check that difference kind is correct
	if len(diffs) > 0 && diffs[0].Kind != InterfaceRequirementsChanged {
		t.Errorf("Expected InterfaceRequirementsChanged, got %s", diffs[0].Kind)
	}
}

// TestVersionPolicies tests version policy constants
func TestVersionPolicies(t *testing.T) {
	// Test that version policies are defined correctly
	policies := []VersionPolicy{
		FromCallingModule,
		PreferLatest,
		VersionSpecific,
		Reconcile,
	}

	// Check they have different values
	seen := make(map[VersionPolicy]bool)
	for _, policy := range policies {
		if seen[policy] {
			t.Errorf("Duplicate version policy value: %v", policy)
		}
		seen[policy] = true
	}

	// Just a simple test to ensure they're all different values
	if len(seen) != 4 {
		t.Errorf("Expected 4 distinct version policies")
	}
}
