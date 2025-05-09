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

// TestCompareStructs tests the enhanced struct comparison functionality
func TestCompareStructs(t *testing.T) {
	// Create package and variable objects for field creation
	pkg := types.NewPackage("example.com/pkg", "pkg")

	// Create base struct with fields
	baseFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Int], false),
		types.NewField(0, pkg, "Private", types.Typ[types.Bool], false),
	}
	baseTags := []string{`json:"name"`, `json:"age"`, `json:"-"`}
	baseStruct := types.NewStruct(baseFields, baseTags)

	// Create struct with added field
	addedFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Int], false),
		types.NewField(0, pkg, "Private", types.Typ[types.Bool], false),
		types.NewField(0, pkg, "Email", types.Typ[types.String], false), // Added field
	}
	addedTags := []string{`json:"name"`, `json:"age"`, `json:"-"`, `json:"email"`}
	addedFieldStruct := types.NewStruct(addedFields, addedTags)

	// Create struct with removed field
	removedFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		// Age field removed
		types.NewField(0, pkg, "Private", types.Typ[types.Bool], false),
	}
	removedTags := []string{`json:"name"`, `json:"-"`}
	removedFieldStruct := types.NewStruct(removedFields, removedTags)

	// Create struct with changed field type
	changedTypeFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Float64], false), // Changed from int to float64
		types.NewField(0, pkg, "Private", types.Typ[types.Bool], false),
	}
	changedTypeTags := []string{`json:"name"`, `json:"age"`, `json:"-"`}
	changedTypeStruct := types.NewStruct(changedTypeFields, changedTypeTags)

	// Create struct with changed tag
	changedTagFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Int], false),
		types.NewField(0, pkg, "Private", types.Typ[types.Bool], false),
	}
	changedTagTags := []string{`json:"name"`, `json:"age,omitempty"`, `json:"-"`} // Changed tag
	changedTagStruct := types.NewStruct(changedTagFields, changedTagTags)

	// Create struct with changed visibility
	changedVisibilityFields := []*types.Var{
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Int], false),
		types.NewField(0, pkg, "private", types.Typ[types.Bool], false), // Changed from Private to private
	}
	changedVisibilityTags := []string{`json:"name"`, `json:"age"`, `json:"-"`}
	changedVisibilityStruct := types.NewStruct(changedVisibilityFields, changedVisibilityTags)

	// Create embedded struct for testing
	embeddedFields := []*types.Var{
		types.NewField(0, pkg, "ID", types.Typ[types.Int], false),
		types.NewField(0, pkg, "CreatedAt", types.Typ[types.String], false),
	}
	embeddedTags := []string{`json:"id"`, `json:"created_at"`}
	embeddedStruct := types.NewStruct(embeddedFields, embeddedTags)
	namedEmbedded := types.NewNamed(
		types.NewTypeName(0, pkg, "BaseEntity", nil),
		embeddedStruct,
		nil,
	)

	// Create struct with embedded field
	withEmbeddedFields := []*types.Var{
		types.NewField(0, pkg, "BaseEntity", namedEmbedded, true), // Embedded
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
		types.NewField(0, pkg, "Age", types.Typ[types.Int], false),
	}
	withEmbeddedTags := []string{``, `json:"name"`, `json:"age"`}
	withEmbeddedStruct := types.NewStruct(withEmbeddedFields, withEmbeddedTags)

	// Create struct with embedded field that has a field overridden
	withOverrideFields := []*types.Var{
		types.NewField(0, pkg, "BaseEntity", namedEmbedded, true),    // Embedded
		types.NewField(0, pkg, "ID", types.Typ[types.String], false), // Overrides BaseEntity.ID
		types.NewField(0, pkg, "Name", types.Typ[types.String], false),
	}
	withOverrideTags := []string{``, `json:"id,string"`, `json:"name"`}
	withOverrideStruct := types.NewStruct(withOverrideFields, withOverrideTags)

	// Create symbols for testing
	baseSymbol := &typesys.Symbol{
		ID:       "base",
		Name:     "Base",
		Kind:     typesys.KindStruct,
		TypeInfo: baseStruct,
	}

	tests := []struct {
		name           string
		otherStruct    *types.Struct
		expectedDiffs  int
		expectedKinds  []DifferenceKind
		expectedFields []string
	}{
		{
			name:           "Added field",
			otherStruct:    addedFieldStruct,
			expectedDiffs:  1,
			expectedKinds:  []DifferenceKind{FieldAdded},
			expectedFields: []string{"Email"},
		},
		{
			name:           "Removed field",
			otherStruct:    removedFieldStruct,
			expectedDiffs:  1,
			expectedKinds:  []DifferenceKind{FieldRemoved},
			expectedFields: []string{"Age"},
		},
		{
			name:           "Changed field type",
			otherStruct:    changedTypeStruct,
			expectedDiffs:  1,
			expectedKinds:  []DifferenceKind{FieldTypeChanged},
			expectedFields: []string{"Age"},
		},
		{
			name:           "Changed field tag",
			otherStruct:    changedTagStruct,
			expectedDiffs:  1,
			expectedKinds:  []DifferenceKind{FieldTypeChanged},
			expectedFields: []string{"Age (tag)"},
		},
		{
			name:           "Changed field visibility",
			otherStruct:    changedVisibilityStruct,
			expectedDiffs:  2, // Removal of Private + Addition of private
			expectedKinds:  []DifferenceKind{FieldRemoved, FieldAdded},
			expectedFields: []string{"Private", "private"},
		},
		{
			name:           "With embedded fields",
			otherStruct:    withEmbeddedStruct,
			expectedDiffs:  4, // Private removed + Added BaseEntity + ID + CreatedAt
			expectedKinds:  []DifferenceKind{FieldRemoved, FieldAdded, FieldAdded, FieldAdded},
			expectedFields: []string{"Private", "ID", "CreatedAt", "BaseEntity"},
		},
		{
			name:           "With overridden embedded field",
			otherStruct:    withOverrideStruct,
			expectedDiffs:  5, // Age removed + Private removed + Added ID + CreatedAt + BaseEntity
			expectedKinds:  []DifferenceKind{FieldRemoved, FieldRemoved, FieldAdded, FieldAdded, FieldAdded},
			expectedFields: []string{"Age", "Private", "ID", "CreatedAt", "BaseEntity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otherSymbol := &typesys.Symbol{
				ID:       "other",
				Name:     "Other",
				Kind:     typesys.KindStruct,
				TypeInfo: tt.otherStruct,
			}

			diffs := compareStructs(baseSymbol, otherSymbol)

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
				for i, diff := range diffs {
					t.Logf("Diff %d: %+v", i, diff)
				}
			}

			// Check that we have the expected kinds of differences
			if len(tt.expectedKinds) > 0 && len(diffs) > 0 {
				// This is a simplified check, assumes order matters
				for i := 0; i < len(tt.expectedKinds) && i < len(diffs); i++ {
					if diffs[i].Kind != tt.expectedKinds[i] {
						t.Errorf("Expected diff kind %s at index %d, got %s",
							tt.expectedKinds[i], i, diffs[i].Kind)
					}

					// Check the field name if provided
					if i < len(tt.expectedFields) && diffs[i].FieldName != tt.expectedFields[i] {
						t.Errorf("Expected field name %s at index %d, got %s",
							tt.expectedFields[i], i, diffs[i].FieldName)
					}
				}
			}
		})
	}
}

// TestCompareInterfaces tests comparing interface types
func TestCompareInterfaces(t *testing.T) {
	// Create a package for our tests
	pkg := types.NewPackage("example.com/pkg", "pkg")

	// Create base interface with no methods
	baseIface := types.NewInterface(
		nil, // methods
		nil, // embedded interfaces
	)

	// Create interface with one method
	oneMethodIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),   // results
				false, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create interface with different method signature
	differentSignatureIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.String])), // params (different type)
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),      // results
				false, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create interface with different return type
	differentReturnIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.String])), // results (different type)
				false, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create interface with variadic method
	variadicIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "args", types.NewSlice(types.Typ[types.Int]))), // variadic params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),                    // results
				true, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create interface with multiple methods
	multiMethodIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),   // results
				false, // variadic
			)),
			types.NewFunc(0, pkg, "Method2", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.String])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Int])),       // results
				false, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create an interface to embed
	embeddedIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "EmbeddedMethod", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),   // results
				false, // variadic
			)),
		},
		nil, // embedded interfaces
	)

	// Create a named version of the embedded interface
	namedEmbedded := types.NewNamed(
		types.NewTypeName(0, pkg, "Embedded", nil),
		embeddedIface,
		nil,
	)

	// Create interface that embeds another interface
	withEmbeddedIface := types.NewInterface(
		[]*types.Func{
			types.NewFunc(0, pkg, "Method1", types.NewSignature(
				nil, // receiver
				types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])), // params
				types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.Bool])),   // results
				false, // variadic
			)),
		},
		[]*types.Named{namedEmbedded}, // embedded interfaces
	)

	// Create symbols for the interfaces
	baseSymbol := &typesys.Symbol{
		ID:       "base",
		Name:     "BaseIface",
		Kind:     typesys.KindInterface,
		TypeInfo: baseIface,
	}

	tests := []struct {
		name            string
		otherIface      *types.Interface
		expectedDiffs   int
		expectedMethods []string
		expectedKinds   []DifferenceKind
	}{
		{
			name:            "Method added",
			otherIface:      oneMethodIface,
			expectedDiffs:   1,
			expectedMethods: []string{"Method1 (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged},
		},
		{
			name:            "Different method signature",
			otherIface:      differentSignatureIface,
			expectedDiffs:   1,
			expectedMethods: []string{"Method1 (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged},
		},
		{
			name:            "Different return type",
			otherIface:      differentReturnIface,
			expectedDiffs:   1,
			expectedMethods: []string{"Method1 (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged},
		},
		{
			name:            "Variadic method",
			otherIface:      variadicIface,
			expectedDiffs:   1,
			expectedMethods: []string{"Method1 (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged},
		},
		{
			name:            "Multiple methods",
			otherIface:      multiMethodIface,
			expectedDiffs:   2,
			expectedMethods: []string{"Method1 (method)", "Method2 (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged, MethodSignatureChanged},
		},
		{
			name:            "Embedded interface",
			otherIface:      withEmbeddedIface,
			expectedDiffs:   2,
			expectedMethods: []string{"Method1 (method)", "EmbeddedMethod (method)"},
			expectedKinds:   []DifferenceKind{MethodSignatureChanged, MethodSignatureChanged},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otherSymbol := &typesys.Symbol{
				ID:       "other",
				Name:     "OtherIface",
				Kind:     typesys.KindInterface,
				TypeInfo: tt.otherIface,
			}

			diffs := compareInterfaces(baseSymbol, otherSymbol)

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
				for i, diff := range diffs {
					t.Logf("Diff %d: %+v", i, diff)
				}
			}

			// Check that we have the expected kinds of differences
			if len(tt.expectedKinds) > 0 && len(diffs) > 0 {
				// Check each expected method is found in the diffs
				// (we don't enforce order here)
				methodFound := make([]bool, len(tt.expectedMethods))

				for _, diff := range diffs {
					for i, method := range tt.expectedMethods {
						if diff.FieldName == method && !methodFound[i] {
							methodFound[i] = true
							break
						}
					}

					// Check that the difference kind is one of the expected kinds
					kindFound := false
					for _, kind := range tt.expectedKinds {
						if diff.Kind == kind {
							kindFound = true
							break
						}
					}

					if !kindFound {
						t.Errorf("Unexpected difference kind: %s", diff.Kind)
					}
				}

				// Check all expected methods were found
				for i, found := range methodFound {
					if !found {
						t.Errorf("Expected method difference for %s not found", tt.expectedMethods[i])
					}
				}
			}
		})
	}

	// Test comparing interfaces with embedded methods
	oneMethodSymbol := &typesys.Symbol{
		ID:       "oneMethod",
		Name:     "OneMethodIface",
		Kind:     typesys.KindInterface,
		TypeInfo: oneMethodIface,
	}

	// Compare one method interface with embedded interface that has the same method and more
	t.Run("Compare with embedded containing same method", func(t *testing.T) {
		embeddedSameMethodIface := types.NewInterface(
			nil, // no direct methods
			[]*types.Named{
				types.NewNamed(
					types.NewTypeName(0, pkg, "Embedded", nil),
					oneMethodIface, // this has Method1
					nil,
				),
			},
		)

		embedsOneMethodSymbol := &typesys.Symbol{
			ID:       "embedsOneMethod",
			Name:     "EmbedsOneMethod",
			Kind:     typesys.KindInterface,
			TypeInfo: embeddedSameMethodIface,
		}

		diffs := compareInterfaces(oneMethodSymbol, embedsOneMethodSymbol)

		// No differences expected because the method sets are the same
		if len(diffs) != 0 {
			t.Errorf("Expected no differences, got %d", len(diffs))
			for i, diff := range diffs {
				t.Logf("Diff %d: %+v", i, diff)
			}
		}
	})
}

// TestTypesAreEqual tests the typesAreEqual function
func TestTypesAreEqual(t *testing.T) {
	tests := []struct {
		name     string
		type1    types.Type
		type2    types.Type
		expected bool
	}{
		{
			name:     "Identical basic types",
			type1:    types.Typ[types.Int],
			type2:    types.Typ[types.Int],
			expected: true,
		},
		{
			name:     "Different basic types",
			type1:    types.Typ[types.Int],
			type2:    types.Typ[types.String],
			expected: false,
		},
		{
			name:     "Identical slice types",
			type1:    types.NewSlice(types.Typ[types.Int]),
			type2:    types.NewSlice(types.Typ[types.Int]),
			expected: true,
		},
		{
			name:     "Different slice types",
			type1:    types.NewSlice(types.Typ[types.Int]),
			type2:    types.NewSlice(types.Typ[types.String]),
			expected: false,
		},
		{
			name:     "Identical array types",
			type1:    types.NewArray(types.Typ[types.Int], 5),
			type2:    types.NewArray(types.Typ[types.Int], 5),
			expected: true,
		},
		{
			name:     "Arrays with different lengths",
			type1:    types.NewArray(types.Typ[types.Int], 5),
			type2:    types.NewArray(types.Typ[types.Int], 10),
			expected: false,
		},
		{
			name:     "Identical map types",
			type1:    types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			type2:    types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			expected: true,
		},
		{
			name:     "Maps with different key types",
			type1:    types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			type2:    types.NewMap(types.Typ[types.Int], types.Typ[types.Int]),
			expected: false,
		},
		{
			name:     "Identical pointer types",
			type1:    types.NewPointer(types.Typ[types.Int]),
			type2:    types.NewPointer(types.Typ[types.Int]),
			expected: true,
		},
		{
			name:     "Different pointer types",
			type1:    types.NewPointer(types.Typ[types.Int]),
			type2:    types.NewPointer(types.Typ[types.String]),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typesAreEqual(tt.type1, tt.type2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
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
