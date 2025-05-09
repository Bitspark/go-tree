package typesys

import (
	"go/token"
	"testing"
)

func TestSymbolCreation(t *testing.T) {
	testCases := []struct {
		name       string
		symbolName string
		kind       SymbolKind
		expectedID string
		isExported bool
	}{
		{
			name:       "exported function",
			symbolName: "ExportedFunc",
			kind:       KindFunction,
			expectedID: "ExportedFunc:2",
			isExported: true,
		},
		{
			name:       "unexported variable",
			symbolName: "unexportedVar",
			kind:       KindVariable,
			expectedID: "unexportedVar:5",
			isExported: false,
		},
		{
			name:       "exported struct",
			symbolName: "ExportedStruct",
			kind:       KindStruct,
			expectedID: "ExportedStruct:10",
			isExported: true,
		},
		{
			name:       "embedded field",
			symbolName: "EmbeddedType",
			kind:       KindEmbeddedField,
			expectedID: "EmbeddedType:13",
			isExported: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sym := NewSymbol(tc.symbolName, tc.kind)

			if sym.Name != tc.symbolName {
				t.Errorf("Symbol.Name = %q, want %q", sym.Name, tc.symbolName)
			}

			if sym.Kind != tc.kind {
				t.Errorf("Symbol.Kind = %v, want %v", sym.Kind, tc.kind)
			}

			if sym.ID != tc.expectedID {
				t.Errorf("Symbol.ID = %q, want %q", sym.ID, tc.expectedID)
			}

			if sym.Exported != tc.isExported {
				t.Errorf("Symbol.Exported = %v, want %v", sym.Exported, tc.isExported)
			}

			if len(sym.Definitions) != 0 {
				t.Errorf("New symbol should have no definitions, got %d", len(sym.Definitions))
			}

			if len(sym.References) != 0 {
				t.Errorf("New symbol should have no references, got %d", len(sym.References))
			}
		})
	}
}

func TestSymbolKindString(t *testing.T) {
	testCases := []struct {
		kind     SymbolKind
		expected string
	}{
		{KindUnknown, "unknown"},
		{KindPackage, "package"},
		{KindFunction, "function"},
		{KindMethod, "method"},
		{KindType, "type"},
		{KindVariable, "variable"},
		{KindConstant, "constant"},
		{KindField, "field"},
		{KindParameter, "parameter"},
		{KindInterface, "interface"},
		{KindStruct, "struct"},
		{KindImport, "import"},
		{KindLabel, "label"},
		{KindEmbeddedField, "embedded_field"},
		{KindEmbeddedInterface, "embedded_interface"},
		{SymbolKind(999), "unknown"}, // Unknown value should return "unknown"
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			got := tc.kind.String()
			if got != tc.expected {
				t.Errorf("(%d).String() = %q, want %q", tc.kind, got, tc.expected)
			}
		})
	}
}

func TestAddReference(t *testing.T) {
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)

	// Create a symbol
	sym := NewSymbol("TestSymbol", KindFunction)

	// Create a reference
	ref := &Reference{
		Symbol:  sym,
		File:    file,
		Pos:     token.Pos(10),
		End:     token.Pos(20),
		IsWrite: false,
	}

	// Add the reference
	sym.AddReference(ref)

	// Check that it was added
	if len(sym.References) != 1 {
		t.Errorf("Symbol should have 1 reference, got %d", len(sym.References))
	}

	if sym.References[0] != ref {
		t.Errorf("Symbol reference not correctly added")
	}
}

func TestAddDefinition(t *testing.T) {
	// Create a symbol
	sym := NewSymbol("TestSymbol", KindFunction)

	// Add a definition position
	sym.AddDefinition("/test/module/file.go", token.Pos(10), 5, 3)

	// Check that it was added correctly
	if len(sym.Definitions) != 1 {
		t.Errorf("Symbol should have 1 definition, got %d", len(sym.Definitions))
	}

	def := sym.Definitions[0]

	if def.File != "/test/module/file.go" {
		t.Errorf("Definition file = %q, want %q", def.File, "/test/module/file.go")
	}

	if def.Pos != token.Pos(10) {
		t.Errorf("Definition pos = %d, want %d", def.Pos, 10)
	}

	if def.Line != 5 {
		t.Errorf("Definition line = %d, want %d", def.Line, 5)
	}

	if def.Column != 3 {
		t.Errorf("Definition column = %d, want %d", def.Column, 3)
	}
}

func TestGetPosition(t *testing.T) {
	// Create a file set
	fset := token.NewFileSet()

	// Create module, package, and file
	module := NewModule("/test/module")
	pkg := NewPackage(module, "testpkg", "github.com/example/testpkg")
	file := NewFile("/test/module/file.go", pkg)
	file.FileSet = fset

	// Create a symbol
	sym := NewSymbol("TestSymbol", KindFunction)
	sym.File = file
	sym.Pos = token.Pos(10)
	sym.End = token.Pos(20)

	// Test GetPosition
	posInfo := sym.GetPosition()

	// Since we're not using a real FileSet with a real file, this should return nil
	if posInfo != nil {
		t.Errorf("GetPosition should return nil with mock FileSet")
	}

	// Test with nil file
	sym.File = nil
	posInfo = sym.GetPosition()

	if posInfo != nil {
		t.Errorf("GetPosition should return nil with nil file")
	}
}
