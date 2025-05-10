package execute

import (
	"strings"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// Mock typesys.Module and typesys.Symbol for testing
func createMockModule() *typesys.Module {
	// Create a new module
	module := typesys.NewModule("")
	module.Path = "github.com/test/simplemath"

	// Create a package
	pkg := typesys.NewPackage(module, "simplemath", "github.com/test/simplemath")
	module.Packages["github.com/test/simplemath"] = pkg

	// Create a function symbol
	addFunc := typesys.NewSymbol("Add", typesys.KindFunction)
	addFunc.Package = pkg

	// Add the symbol to the package
	pkg.Symbols[addFunc.ID] = addFunc

	return module
}

func TestGenerateFunctionWrapper(t *testing.T) {
	module := createMockModule()
	// Find the Add function in the package
	var addFunc *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			addFunc = sym
			break
		}
	}

	if addFunc == nil {
		t.Fatal("Failed to find Add function in mock module")
	}

	generator := NewTypeAwareGenerator()
	code, err := generator.GenerateFunctionWrapper(module, addFunc, 5, 3)

	if err != nil {
		t.Fatalf("Failed to generate wrapper code: %v", err)
	}

	// Check that the generated code contains the expected elements
	expectedElements := []string{
		"package main",
		"import",
		"pkg \"github.com/test/simplemath\"",
		"result := pkg.Add",
		"5, 3",
		"json.Marshal",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(code, expected) {
			t.Errorf("Generated code missing expected element: %s", expected)
		}
	}
}

func TestGenerateFunctionWrapper_WithDifferentTypes(t *testing.T) {
	module := createMockModule()
	// Find the Add function in the package
	var addFunc *typesys.Symbol
	for _, sym := range module.Packages["github.com/test/simplemath"].Symbols {
		if sym.Name == "Add" && sym.Kind == typesys.KindFunction {
			addFunc = sym
			break
		}
	}

	if addFunc == nil {
		t.Fatal("Failed to find Add function in mock module")
	}

	testCases := []struct {
		name     string
		args     []interface{}
		expected []string
	}{
		{
			name: "string arguments",
			args: []interface{}{"hello", "world"},
			expected: []string{
				"\"hello\", \"world\"",
			},
		},
		{
			name: "bool arguments",
			args: []interface{}{true, false},
			expected: []string{
				"true, false",
			},
		},
		{
			name: "float arguments",
			args: []interface{}{1.5, 2.5},
			expected: []string{
				"1.500000, 2.500000",
			},
		},
		{
			name: "mixed arguments",
			args: []interface{}{42, "test", true},
			expected: []string{
				"42, \"test\", true",
			},
		},
	}

	generator := NewTypeAwareGenerator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := generator.GenerateFunctionWrapper(module, addFunc, tc.args...)

			if err != nil {
				t.Fatalf("Failed to generate wrapper code: %v", err)
			}

			for _, expected := range tc.expected {
				if !strings.Contains(code, expected) {
					t.Errorf("Generated code missing expected element: %s", expected)
				}
			}
		})
	}
}

func TestGenerateFunctionWrapper_InvalidInputs(t *testing.T) {
	generator := NewTypeAwareGenerator()

	// Test with nil module
	_, err := generator.GenerateFunctionWrapper(nil, &typesys.Symbol{}, 1, 2)
	if err == nil {
		t.Error("Expected error for nil module, got nil")
	}

	// Test with nil function symbol
	module := createMockModule()
	_, err = generator.GenerateFunctionWrapper(module, nil, 1, 2)
	if err == nil {
		t.Error("Expected error for nil function symbol, got nil")
	}

	// Test with non-function symbol
	nonFuncSymbol := typesys.NewSymbol("NotAFunction", typesys.KindVariable)
	_, err = generator.GenerateFunctionWrapper(module, nonFuncSymbol, 1, 2)
	if err == nil {
		t.Error("Expected error for non-function symbol, got nil")
	}
}
