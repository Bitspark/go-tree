package format

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/model"
	"bitspark.dev/go-tree/pkg/core/parse"
)

func TestFormatPackage(t *testing.T) {
	// First parse a package
	pkg, err := parse.ParsePackage("../../test/samplepackage")
	if err != nil {
		t.Fatalf("Failed to parse package: %v", err)
	}

	// Format the package
	formatted, err := FormatPackage(pkg)
	if err != nil {
		t.Fatalf("Failed to format package: %v", err)
	}

	// Check that the formatted code contains all necessary elements
	// Package name
	if !strings.Contains(formatted, "package samplepackage") {
		t.Error("Formatted code doesn't contain the package name")
	}

	// Imports
	for _, imp := range []string{"time", "errors", "fmt"} {
		if !strings.Contains(formatted, "\""+imp+"\"") {
			t.Errorf("Formatted code doesn't contain import %q", imp)
		}
	}

	// Type definitions
	for _, typeName := range []string{"User", "Authentication", "Role", "Authenticator"} {
		if !strings.Contains(formatted, "type "+typeName) {
			t.Errorf("Formatted code doesn't contain type %s", typeName)
		}
	}

	// Constants
	for _, constant := range []string{"RoleAdmin", "RoleUser", "RoleGuest"} {
		if !strings.Contains(formatted, constant) {
			t.Errorf("Formatted code doesn't contain constant %s", constant)
		}
	}

	// Variables
	for _, variable := range []string{"DefaultTimeout", "ErrInvalidCredentials"} {
		if !strings.Contains(formatted, variable) {
			t.Errorf("Formatted code doesn't contain variable %s", variable)
		}
	}

	// Functions
	for _, function := range []string{"NewUser", "Login", "Validate", "FormatUser"} {
		if !strings.Contains(formatted, "func") || !strings.Contains(formatted, function) {
			t.Errorf("Formatted code doesn't contain function %s", function)
		}
	}
}

func TestFormatConstants(t *testing.T) {
	// Create some test constants
	constants := []model.GoConstant{
		{
			Name:  "MaxRetries",
			Type:  "int",
			Value: "3",
			Doc:   "Maximum number of retry attempts",
		},
		{
			Name:  "Timeout",
			Type:  "time.Duration",
			Value: "30 * time.Second",
		},
		{
			Name:    "Version",
			Type:    "string",
			Value:   "\"1.0.0\"",
			Comment: "Current version",
		},
	}

	// Format constants
	var out strings.Builder
	formatConstants(&out, constants)
	result := out.String()

	// Check the output
	t.Log(result)

	// Verify constants are included
	for _, c := range constants {
		if !strings.Contains(result, c.Name) {
			t.Errorf("Formatted constants doesn't contain %s", c.Name)
		}
		if !strings.Contains(result, c.Type) {
			t.Errorf("Formatted constants doesn't contain type %s", c.Type)
		}
		if !strings.Contains(result, c.Value) {
			t.Errorf("Formatted constants doesn't contain value %s", c.Value)
		}
	}
}

func TestFormatType(t *testing.T) {
	// Create a test struct type
	structType := model.GoType{
		Name: "Person",
		Kind: "struct",
		Fields: []model.GoField{
			{
				Name:    "Name",
				Type:    "string",
				Tag:     "`json:\"name\"`",
				Comment: "Person's name",
			},
			{
				Name: "Age",
				Type: "int",
				Tag:  "`json:\"age\"`",
			},
			{
				Name: "",
				Type: "Address",
			},
		},
		Doc: "Person represents a person in the system",
	}

	// Format the type
	var out strings.Builder
	formatType(&out, structType)
	result := out.String()

	// Check the output
	t.Log(result)

	// Verify type definition is included
	if !strings.Contains(result, "type Person struct") {
		t.Error("Missing struct type definition")
	}

	// Verify fields are included
	if !strings.Contains(result, "Name string") {
		t.Error("Missing Name field")
	}
	if !strings.Contains(result, "Age int") {
		t.Error("Missing Age field")
	}
	if !strings.Contains(result, "Address") {
		t.Error("Missing embedded Address field")
	}

	// Verify docs and comments
	if !strings.Contains(result, "Person represents a person") {
		t.Error("Missing doc comment")
	}
	if !strings.Contains(result, "Person's name") {
		t.Error("Missing field comment")
	}
}
