package parse

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
)

func TestParsePackage(t *testing.T) {
	// Test parsing the sample package we created
	pkg, err := ParsePackage("../../../testdata/samplepackage")
	if err != nil {
		t.Fatalf("Failed to parse package: %v", err)
	}

	// Check package name
	if pkg.Name != "samplepackage" {
		t.Errorf("Expected package name 'samplepackage', got '%s'", pkg.Name)
	}

	// Check imports
	expectedImports := []string{
		"errors",
		"fmt",
		"time",
	}
	if len(pkg.Imports) != len(expectedImports) {
		t.Errorf("Expected %d imports, got %d", len(expectedImports), len(pkg.Imports))
	} else {
		importPaths := make([]string, len(pkg.Imports))
		for i, imp := range pkg.Imports {
			importPaths[i] = imp.Path
		}
		for _, expected := range expectedImports {
			found := false
			for _, actual := range importPaths {
				if expected == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected import '%s' not found", expected)
			}
		}
	}

	// Check if types were correctly parsed
	expectedTypes := []string{
		"User",
		"Authentication",
		"timestamps",
		"Role",
		"Authenticator",
		"Validator",
		"UserMap",
		"AuthHandler",
	}

	if len(pkg.Types) != len(expectedTypes) {
		t.Errorf("Expected %d types, got %d", len(expectedTypes), len(pkg.Types))
	} else {
		typeNames := make([]string, len(pkg.Types))
		for i, typ := range pkg.Types {
			typeNames[i] = typ.Name
		}
		for _, expected := range expectedTypes {
			found := false
			for _, actual := range typeNames {
				if expected == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected type '%s' not found", expected)
			}
		}
	}

	// Check if constants were correctly parsed
	expectedConstants := []string{
		"RoleAdmin",
		"RoleUser",
		"RoleGuest",
	}

	if len(pkg.Constants) != len(expectedConstants) {
		t.Errorf("Expected %d constants, got %d", len(expectedConstants), len(pkg.Constants))
	} else {
		constNames := make([]string, len(pkg.Constants))
		for i, c := range pkg.Constants {
			constNames[i] = c.Name
		}
		for _, expected := range expectedConstants {
			found := false
			for _, actual := range constNames {
				if expected == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected constant '%s' not found", expected)
			}
		}
	}

	// Check if variables were correctly parsed
	expectedVars := []string{
		"DefaultTimeout",
		"ErrInvalidCredentials",
		"ErrPermissionDenied",
	}

	if len(pkg.Variables) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(pkg.Variables))
	} else {
		varNames := make([]string, len(pkg.Variables))
		for i, v := range pkg.Variables {
			varNames[i] = v.Name
		}
		for _, expected := range expectedVars {
			found := false
			for _, actual := range varNames {
				if expected == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected variable '%s' not found", expected)
			}
		}
	}

	// Check if functions were correctly parsed
	expectedFuncs := []string{
		"NewUser",
		"UpdatePassword",
		"Login",
		"Logout",
		"Validate",
		"FormatUser",
	}

	if len(pkg.Functions) != len(expectedFuncs) {
		t.Errorf("Expected %d functions, got %d", len(expectedFuncs), len(pkg.Functions))
	} else {
		funcNames := make([]string, len(pkg.Functions))
		for i, f := range pkg.Functions {
			funcNames[i] = f.Name
		}
		for _, expected := range expectedFuncs {
			found := false
			for _, actual := range funcNames {
				if expected == actual {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected function '%s' not found", expected)
			}
		}
	}
}

// Test parsing a specific type declaration
func TestParseType(t *testing.T) {
	pkg, err := ParsePackage("../../../testdata/samplepackage")
	if err != nil {
		t.Fatalf("Failed to parse package: %v", err)
	}

	// Find the User struct
	var userType *model.GoType
	for _, typ := range pkg.Types {
		if typ.Name == "User" {
			userType = &typ
			break
		}
	}

	if userType == nil {
		t.Fatalf("User type not found")
	}

	// Check type kind
	if userType.Kind != "struct" {
		t.Errorf("Expected User type kind to be 'struct', got '%s'", userType.Kind)
	}

	// Check fields
	expectedFields := []string{
		"ID",
		"Name",
		"Email",
		"Phone",
		"", // Embedded Authentication
		"", // Embedded timestamps
	}

	if len(userType.Fields) != len(expectedFields) {
		t.Errorf("Expected %d fields, got %d", len(expectedFields), len(userType.Fields))
	} else {
		for i, expected := range expectedFields {
			if userType.Fields[i].Name != expected {
				if expected == "" {
					// For embedded fields, Name is empty
					if userType.Fields[i].Name != "" {
						t.Errorf("Expected embedded field (empty name) at index %d, got '%s'", i, userType.Fields[i].Name)
					}
				} else {
					t.Errorf("Expected field '%s' at index %d, got '%s'", expected, i, userType.Fields[i].Name)
				}
			}
		}
	}
}
