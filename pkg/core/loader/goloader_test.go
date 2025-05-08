package loader

import (
	"testing"

	"go/token"

	"strings"

	"bitspark.dev/go-tree/pkg/core/module"
)

func TestGoModuleLoader_Load(t *testing.T) {
	// Create a new loader
	loader := NewGoModuleLoader()

	// Load the sample module
	mod, err := loader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify basic module properties
	if mod.Path != "test" {
		t.Errorf("Expected module path to be 'test', got %q", mod.Path)
	}

	// Verify package was loaded
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	// Verify package properties
	if samplePkg.Name != "samplepackage" {
		t.Errorf("Expected package name to be 'samplepackage', got %q", samplePkg.Name)
	}

	// Verify files were loaded
	typesFile, ok := samplePkg.Files["types.go"]
	if !ok {
		t.Fatalf("Expected to find file 'types.go'")
	}

	// Verify functions file exists
	_, ok = samplePkg.Files["functions.go"]
	if !ok {
		t.Fatalf("Expected to find file 'functions.go'")
	}

	// Verify types were loaded
	userType, ok := samplePkg.Types["User"]
	if !ok {
		t.Fatalf("Expected to find type 'User'")
	}

	// DEBUG: Print type information
	t.Logf("User type: Name=%s, Kind=%s, Pos=%v, End=%v",
		userType.Name, userType.Kind, userType.Pos, userType.End)
	t.Logf("User type has %d fields, %d methods",
		len(userType.Fields), len(userType.Methods))

	// Verify type properties
	if userType.Kind != "struct" {
		t.Errorf("Expected User to be a struct, got %q", userType.Kind)
	}

	// Verify functions were loaded
	newUserFunc, ok := samplePkg.Functions["NewUser"]
	if !ok {
		t.Fatalf("Expected to find function 'NewUser'")
	}

	// DEBUG: Print function information
	t.Logf("NewUser function: Pos=%v, End=%v", newUserFunc.Pos, newUserFunc.End)

	// Verify position information
	if userType.Pos == token.NoPos || userType.End == token.NoPos {
		t.Error("Expected User type to have position information")
	}

	if newUserFunc.Pos == token.NoPos || newUserFunc.End == token.NoPos {
		t.Error("Expected NewUser function to have position information")
	}

	// DEBUG: Print TokenFile information
	if typesFile.TokenFile == nil {
		t.Logf("WARNING: TokenFile is nil")
	} else {
		t.Logf("TokenFile: Base=%v, Size=%v", typesFile.TokenFile.Base(), typesFile.TokenFile.Size())
	}

	// DEBUG: List all functions in package
	t.Logf("All functions in package:")
	for name, fn := range samplePkg.Functions {
		receiver := "none"
		if fn.Receiver != nil {
			receiver = fn.Receiver.Type
		}
		t.Logf("  %s: Receiver=%s, IsMethod=%v", name, receiver, fn.IsMethod)
	}

	// DEBUG: List all methods for User type
	t.Logf("Methods for User type:")
	for _, method := range userType.Methods {
		t.Logf("  %s: Pos=%v, End=%v", method.Name, method.Pos, method.End)
	}

	// Test FindElementAtPosition
	if typesFile.TokenFile != nil {
		// Find a position inside the User type
		userPos := userType.Pos + 10 // Position inside User type

		// DEBUG: Print position debugging info
		t.Logf("Looking for element at position %v (User type is at %v-%v)",
			userPos, userType.Pos, userType.End)

		element := typesFile.FindElementAtPosition(userPos)

		if element == nil {
			t.Logf("No element found at position %v", userPos)
		} else {
			t.Logf("Found element of type %T", element)
		}

		// Verify we found the User type
		foundType, ok := element.(*module.Type)
		if !ok {
			t.Errorf("Expected to find a Type at position, got %T", element)
		} else if foundType.Name != "User" {
			t.Errorf("Expected to find User type, got %q", foundType.Name)
		}
	}

	// Verify methods were loaded for User type
	var foundUpdatePassword bool
	for _, method := range userType.Methods {
		if method.Name == "UpdatePassword" {
			foundUpdatePassword = true

			// Verify method has position information
			if method.Pos == token.NoPos || method.End == token.NoPos {
				t.Error("Expected UpdatePassword method to have position information")
			}

			break
		}
	}

	if !foundUpdatePassword {
		t.Error("Expected to find UpdatePassword method on User type")
	}
}

func TestPositionInfo(t *testing.T) {
	// Create a new loader
	loader := NewGoModuleLoader()

	// Load the sample module
	mod, err := loader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get sample package
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	// Get functions file
	functionsFile, ok := samplePkg.Files["functions.go"]
	if !ok {
		t.Fatalf("Expected to find file 'functions.go'")
	}

	// Verify source code was captured
	if functionsFile.SourceCode == "" {
		t.Fatal("Expected source code to be captured")
	}

	// Test GetPositionInfo
	newUserFunc, ok := samplePkg.Functions["NewUser"]
	if !ok {
		t.Fatalf("Expected to find function 'NewUser'")
	}

	// Get position info
	pos := newUserFunc.GetPosition()
	if pos == nil {
		t.Fatal("Expected to get position information for NewUser function")
	}

	// Verify position details
	if pos.LineStart <= 0 || pos.ColStart <= 0 {
		t.Errorf("Expected valid line/column information, got line %d, col %d",
			pos.LineStart, pos.ColStart)
	}

	// Verify position string
	posStr := pos.String()
	if posStr == "<unknown position>" {
		t.Error("Expected a valid position string, got '<unknown position>'")
	}

	// Check that position string contains functions.go
	if !strings.Contains(posStr, "functions.go") {
		t.Errorf("Expected position string to contain 'functions.go', got '%s'", posStr)
	}

	// Check that position string contains line and column numbers
	if !strings.Contains(posStr, ":") {
		t.Errorf("Expected position string to contain line/column information (with ':'), got '%s'", posStr)
	}
}

func TestEncodedStructTags(t *testing.T) {
	// Create a new loader
	loader := NewGoModuleLoader()

	// Load the sample module
	mod, err := loader.Load("../../../testdata")
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Get the User type
	samplePkg, ok := mod.Packages["test/samplepackage"]
	if !ok {
		t.Fatalf("Expected to find package 'test/samplepackage'")
	}

	userType, ok := samplePkg.Types["User"]
	if !ok {
		t.Fatalf("Expected to find type 'User'")
	}

	// Verify struct tags were loaded correctly
	foundIDTag := false
	foundEmailTag := false

	for _, field := range userType.Fields {
		switch field.Name {
		case "ID":
			foundIDTag = true
			expectedTag := "`json:\"id\"`"
			if field.Tag != expectedTag {
				t.Errorf("Expected ID tag to be '%s', got '%s'", expectedTag, field.Tag)
			}
		case "Email":
			foundEmailTag = true
			expectedTag := "`json:\"email,omitempty\"`"
			if field.Tag != expectedTag {
				t.Errorf("Expected Email tag to be '%s', got '%s'", expectedTag, field.Tag)
			}
		}
	}

	if !foundIDTag {
		t.Error("Expected to find ID field with tag")
	}

	if !foundEmailTag {
		t.Error("Expected to find Email field with tag")
	}
}
