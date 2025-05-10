package service

import (
	"go/types"
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestDetermineSemverImpact tests the semver impact determination
func TestDetermineSemverImpact(t *testing.T) {
	tests := []struct {
		name           string
		differences    []TypeDifference
		expectedImpact SemverImpact
	}{
		{
			name:           "No differences",
			differences:    []TypeDifference{},
			expectedImpact: NoImpact,
		},
		{
			name: "Added field (minor impact)",
			differences: []TypeDifference{
				{
					Kind:      FieldAdded,
					FieldName: "NewField",
				},
			},
			expectedImpact: MinorImpact,
		},
		{
			name: "Added method (major impact)",
			differences: []TypeDifference{
				{
					Kind:      FieldAdded,
					FieldName: "NewMethod (method)",
				},
			},
			expectedImpact: MajorImpact,
		},
		{
			name: "Removed field (major impact)",
			differences: []TypeDifference{
				{
					Kind:      FieldRemoved,
					FieldName: "OldField",
				},
			},
			expectedImpact: MajorImpact,
		},
		{
			name: "Type change (major impact)",
			differences: []TypeDifference{
				{
					Kind:      FieldTypeChanged,
					FieldName: "Field",
					OldType:   "string",
					NewType:   "int",
				},
			},
			expectedImpact: MajorImpact,
		},
		{
			name: "Widening type change (minor impact)",
			differences: []TypeDifference{
				{
					Kind:      FieldTypeChanged,
					FieldName: "Field",
					OldType:   "int32",
					NewType:   "int64",
				},
			},
			expectedImpact: MinorImpact,
		},
		{
			name: "Mixed changes (major impact prevails)",
			differences: []TypeDifference{
				{
					Kind:      FieldAdded,
					FieldName: "NewField",
				},
				{
					Kind:      FieldRemoved,
					FieldName: "OldField",
				},
			},
			expectedImpact: MajorImpact,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := determineSemverImpact(tt.differences)
			if impact != tt.expectedImpact {
				t.Errorf("Expected impact %s, got %s", tt.expectedImpact, impact)
			}
		})
	}
}

// TestIsWideningTypeChange tests the widening type change detection
func TestIsWideningTypeChange(t *testing.T) {
	tests := []struct {
		oldType  string
		newType  string
		expected bool
	}{
		{"int8", "int16", true},
		{"int16", "int32", true},
		{"int32", "int64", true},
		{"int", "int64", true},
		{"float32", "float64", true},
		{"uint8", "uint16", true},
		{"uint16", "uint32", true},
		{"uint32", "uint64", true},
		{"uint", "uint64", true},
		{"int8", "float32", true},
		{"int16", "float64", true},

		// Non-widening changes
		{"int64", "int32", false},
		{"int", "int32", false},
		{"float64", "float32", false},
		{"int", "string", false},
		{"string", "int", false},
	}

	for _, tt := range tests {
		t.Run(tt.oldType+"->"+tt.newType, func(t *testing.T) {
			result := isWideningTypeChange(tt.oldType, tt.newType)
			if result != tt.expected {
				t.Errorf("Expected isWideningTypeChange(%s, %s) to be %v, got %v",
					tt.oldType, tt.newType, tt.expected, result)
			}
		})
	}
}

// TestCalculateCompatibilityScore tests the compatibility score calculation
func TestCalculateCompatibilityScore(t *testing.T) {
	tests := []struct {
		name          string
		differences   []TypeDifference
		expectedScore int
	}{
		{
			name:          "No differences",
			differences:   []TypeDifference{},
			expectedScore: 100,
		},
		{
			name: "All major differences",
			differences: []TypeDifference{
				{
					Kind:      FieldRemoved,
					FieldName: "Field1",
				},
				{
					Kind:      FieldTypeChanged,
					FieldName: "Field2",
					OldType:   "string",
					NewType:   "int",
				},
			},
			expectedScore: 0,
		},
		{
			name: "Mixed differences",
			differences: []TypeDifference{
				{
					Kind:      FieldAdded,
					FieldName: "NewField",
				},
				{
					Kind:      FieldRemoved,
					FieldName: "OldField",
				},
			},
			expectedScore: 50,
		},
		{
			name: "Only minor differences",
			differences: []TypeDifference{
				{
					Kind:      FieldAdded,
					FieldName: "NewField1",
				},
				{
					Kind:      FieldAdded,
					FieldName: "NewField2",
				},
				{
					Kind:      FieldTypeChanged,
					FieldName: "Field3",
					OldType:   "int32",
					NewType:   "int64", // Widening
				},
			},
			expectedScore: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateCompatibilityScore(tt.differences)
			if score != tt.expectedScore {
				t.Errorf("Expected score %d, got %d", tt.expectedScore, score)
			}
		})
	}
}

// TestAnalyzeSemverCompatibility tests the full semver compatibility analysis
func TestAnalyzeSemverCompatibility(t *testing.T) {
	// Create a service with different module versions for testing
	service := &Service{
		Modules: map[string]*typesys.Module{
			"example.com/mod@v1.0.0": {
				Path: "example.com/mod@v1.0.0",
				Packages: map[string]*typesys.Package{
					"example.com/mod/types": {
						ImportPath: "example.com/mod/types",
						Symbols: map[string]*typesys.Symbol{
							"oldStruct": {
								ID:       "oldStruct",
								Name:     "User",
								Kind:     typesys.KindStruct,
								TypeInfo: createTestStruct([]string{"ID", "Name", "Age"}, []types.Type{types.Typ[types.Int], types.Typ[types.String], types.Typ[types.Int]}),
							},
							"oldInterface": {
								ID:       "oldInterface",
								Name:     "UserManager",
								Kind:     typesys.KindInterface,
								TypeInfo: createTestInterface([]string{"GetUser", "SaveUser"}, 2),
							},
						},
					},
				},
			},
			"example.com/mod@v2.0.0": {
				Path: "example.com/mod@v2.0.0",
				Packages: map[string]*typesys.Package{
					"example.com/mod/types": {
						ImportPath: "example.com/mod/types",
						Symbols: map[string]*typesys.Symbol{
							"newStruct": {
								ID:       "newStruct",
								Name:     "User",
								Kind:     typesys.KindStruct,
								TypeInfo: createTestStruct([]string{"ID", "Name", "Email"}, []types.Type{types.Typ[types.Int], types.Typ[types.String], types.Typ[types.String]}),
							},
							"newInterface": {
								ID:       "newInterface",
								Name:     "UserManager",
								Kind:     typesys.KindInterface,
								TypeInfo: createTestInterface([]string{"GetUser", "SaveUser", "DeleteUser"}, 3),
							},
						},
					},
				},
			},
		},
	}

	// Test struct with field changes
	t.Run("struct with changes", func(t *testing.T) {
		report, err := service.AnalyzeSemverCompatibility(
			"example.com/mod/types", "User", "v1.0.0", "v2.0.0")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check report contents
		if report.TypeName != "User" {
			t.Errorf("Expected TypeName 'User', got '%s'", report.TypeName)
		}

		if report.OldVersion != "v1.0.0" || report.NewVersion != "v2.0.0" {
			t.Errorf("Incorrect versions in report: %s -> %s", report.OldVersion, report.NewVersion)
		}

		// Expect major impact (field removal)
		if report.Impact != MajorImpact {
			t.Errorf("Expected MajorImpact, got %s", report.Impact)
		}

		// Check if we found the right differences
		hasMissingAge := false
		hasAddedEmail := false

		for _, diff := range report.Differences {
			if diff.Kind == FieldRemoved && diff.FieldName == "Age" {
				hasMissingAge = true
			}
			if diff.Kind == FieldAdded && diff.FieldName == "Email" {
				hasAddedEmail = true
			}
		}

		if !hasMissingAge {
			t.Error("Expected to detect removal of 'Age' field")
		}

		if !hasAddedEmail {
			t.Error("Expected to detect addition of 'Email' field")
		}

		// Check suggestions
		if len(report.Suggestions) == 0 {
			t.Error("Expected suggestions for fixing compatibility issues")
		}
	})

	// Test interface with method changes
	t.Run("interface with changes", func(t *testing.T) {
		report, err := service.AnalyzeSemverCompatibility(
			"example.com/mod/types", "UserManager", "v1.0.0", "v2.0.0")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Expect major impact (added interface method)
		if report.Impact != MajorImpact {
			t.Errorf("Expected MajorImpact, got %s", report.Impact)
		}

		// Check if we found the right differences
		hasAddedMethod := false

		for _, diff := range report.Differences {
			if diff.Kind == MethodSignatureChanged && strings.Contains(diff.FieldName, "DeleteUser") {
				hasAddedMethod = true
			}
		}

		if !hasAddedMethod {
			t.Error("Expected to detect addition of 'DeleteUser' method")
		}
	})
}

// createTestStruct creates a struct type with the specified fields
func createTestStruct(fieldNames []string, fieldTypes []types.Type) *types.Struct {
	pkg := types.NewPackage("example.com/test", "test")
	fields := make([]*types.Var, len(fieldNames))
	tags := make([]string, len(fieldNames))

	for i, name := range fieldNames {
		fields[i] = types.NewField(0, pkg, name, fieldTypes[i], false)
		tags[i] = ""
	}

	return types.NewStruct(fields, tags)
}

// createTestInterface creates an interface type with the specified methods
func createTestInterface(methodNames []string, numMethods int) *types.Interface {
	pkg := types.NewPackage("example.com/test", "test")
	var methods []*types.Func

	for i := 0; i < numMethods && i < len(methodNames); i++ {
		// Create a method signature (func(int) string)
		sig := types.NewSignatureType(
			nil, // receiver
			nil, // type params
			nil, // instance
			types.NewTuple(types.NewVar(0, pkg, "arg", types.Typ[types.Int])),
			types.NewTuple(types.NewVar(0, pkg, "", types.Typ[types.String])),
			false, // variadic
		)

		// Create the method
		methods = append(methods, types.NewFunc(0, pkg, methodNames[i], sig))
	}

	return types.NewInterfaceType(methods, nil)
}
