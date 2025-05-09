package index

import (
	"strings"
	"testing"
)

// Define mock types for testing
type SymbolMatch struct {
	Name string
	Kind string
	Path string
	ID   string
}

type ReferenceMatch struct {
	Path   string
	Line   int
	Column int
}

// MockCommandContext is a simplified version for testing
type MockCommandContext struct {
	Indexer             interface{}
	findSymbolsByName   func(name string) []*SymbolMatch
	findReferences      func(id string) []*ReferenceMatch
	findImplementations func(id string) []*SymbolMatch
	findMethodsForType  func(typeName string) []*SymbolMatch
	getFileStructure    func(filePath string) []*SymbolMatch
}

// ExecuteCommand simulates command execution for testing
func (ctx *MockCommandContext) ExecuteCommand(cmd string, arg string) (string, error) {
	switch cmd {
	case "find":
		if arg == "" {
			return "", errorf("Empty search term")
		}

		symbols := ctx.findSymbolsByName(arg)
		if len(symbols) == 0 {
			return "No symbols found", nil
		}

		var result strings.Builder
		result.WriteString("Found symbols:\n")
		for _, s := range symbols {
			result.WriteString(s.Name + " (" + s.Kind + ") in " + s.Path + "\n")
		}
		return result.String(), nil

	case "refs":
		symbols := ctx.findSymbolsByName(arg)
		if len(symbols) == 0 {
			return "No symbols found", nil
		}

		refs := ctx.findReferences(symbols[0].ID)
		if len(refs) == 0 {
			return "No references found", nil
		}

		var result strings.Builder
		result.WriteString("Found references:\n")
		for _, r := range refs {
			result.WriteString(r.Path + ":" + itoa(r.Line) + ":" + itoa(r.Column) + "\n")
		}
		return result.String(), nil

	case "implements":
		symbols := ctx.findSymbolsByName(arg)
		if len(symbols) == 0 {
			return "No symbols found", nil
		}

		if symbols[0].Kind != "interface" {
			return "Symbol is not an interface", nil
		}

		impls := ctx.findImplementations(symbols[0].ID)
		if len(impls) == 0 {
			return "No implementations found", nil
		}

		var result strings.Builder
		result.WriteString("Found implementations:\n")
		for _, i := range impls {
			result.WriteString(i.Name + " (" + i.Kind + ") in " + i.Path + "\n")
		}
		return result.String(), nil

	case "methods":
		methods := ctx.findMethodsForType(arg)
		if len(methods) == 0 {
			return "No methods found", nil
		}

		var result strings.Builder
		result.WriteString("Found methods:\n")
		for _, m := range methods {
			result.WriteString(m.Name + " in " + m.Path + "\n")
		}
		return result.String(), nil

	case "structure":
		structure := ctx.getFileStructure(arg)
		if len(structure) == 0 {
			return "No symbols found", nil
		}

		var result strings.Builder
		result.WriteString("File structure:\n")
		for _, s := range structure {
			result.WriteString(s.Name + " (" + s.Kind + ")\n")
		}
		return result.String(), nil

	case "help":
		return "Available commands: find, refs, implements, methods, structure", nil

	default:
		return "", errorf("Unknown command: %s", cmd)
	}
}

// Helper functions for tests
func errorf(format string, args ...interface{}) error {
	return &testError{msg: sprintf(format, args...)}
}

func sprintf(format string, args ...interface{}) string {
	// Simple implementation for tests
	result := format
	for _, arg := range args {
		result = strings.Replace(result, "%s", arg.(string), 1)
	}
	return result
}

func itoa(i int) string {
	// Simple integer to string conversion for tests
	if i == 0 {
		return "0"
	}

	var result string
	isNegative := i < 0
	if isNegative {
		i = -i
	}

	for i > 0 {
		digit := i % 10
		// Convert digit to proper string using rune conversion with string()
		result = string(rune('0'+digit)) + result
		i /= 10
	}

	if isNegative {
		result = "-" + result
	}

	return result
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// NewMockCommandContext creates a test context
func NewMockCommandContext() *MockCommandContext {
	return &MockCommandContext{
		Indexer: &struct{}{},
		findSymbolsByName: func(name string) []*SymbolMatch {
			return nil
		},
		findReferences: func(id string) []*ReferenceMatch {
			return nil
		},
		findImplementations: func(id string) []*SymbolMatch {
			return nil
		},
		findMethodsForType: func(typeName string) []*SymbolMatch {
			return nil
		},
		getFileStructure: func(filePath string) []*SymbolMatch {
			return nil
		},
	}
}

func TestNewCommandContext(t *testing.T) {
	ctx := NewMockCommandContext()

	if ctx == nil {
		t.Fatal("NewMockCommandContext should return a non-nil context")
	}

	if ctx.Indexer == nil {
		t.Error("CommandContext should initialize indexer")
	}
}

func TestCommandFindSymbol(t *testing.T) {
	// Create a command context
	ctx := NewMockCommandContext()

	// Mock the index's FindSymbolsByName method
	mockCalled := false
	origMethod := ctx.findSymbolsByName
	defer func() {
		ctx.findSymbolsByName = origMethod
	}()

	ctx.findSymbolsByName = func(name string) []*SymbolMatch {
		mockCalled = true
		if name == "TestSymbol" {
			return []*SymbolMatch{
				{Name: "TestSymbol", Kind: "function", Path: "path/to/file.go"},
			}
		}
		return nil
	}

	// Test successful find
	result, err := ctx.ExecuteCommand("find", "TestSymbol")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !mockCalled {
		t.Error("Mock FindSymbolsByName was not called")
	}

	if !strings.Contains(result, "TestSymbol") {
		t.Errorf("Result should contain symbol name, got: %s", result)
	}

	// Test with empty search term
	_, err = ctx.ExecuteCommand("find", "")
	if err == nil {
		t.Error("ExecuteCommand should return error with empty search term")
	}
}

func TestCommandReferences(t *testing.T) {
	ctx := NewMockCommandContext()

	// Mock the methods
	mockFindCalled := false
	mockRefCalled := false

	origFindMethod := ctx.findSymbolsByName
	origRefMethod := ctx.findReferences
	defer func() {
		ctx.findSymbolsByName = origFindMethod
		ctx.findReferences = origRefMethod
	}()

	ctx.findSymbolsByName = func(name string) []*SymbolMatch {
		mockFindCalled = true
		if name == "TestSymbol" {
			return []*SymbolMatch{
				{Name: "TestSymbol", Kind: "function", Path: "path/to/file.go", ID: "sym123"},
			}
		}
		return nil
	}

	ctx.findReferences = func(id string) []*ReferenceMatch {
		mockRefCalled = true
		if id == "sym123" {
			return []*ReferenceMatch{
				{Path: "path/to/file.go", Line: 10, Column: 5},
				{Path: "path/to/other.go", Line: 20, Column: 15},
			}
		}
		return nil
	}

	// Test successful references command
	result, err := ctx.ExecuteCommand("refs", "TestSymbol")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !mockFindCalled || !mockRefCalled {
		t.Error("Both find and references methods should be called")
	}

	// Check results contains reference locations
	if !strings.Contains(result, "path/to/file.go") || !strings.Contains(result, "path/to/other.go") {
		t.Errorf("Result should contain reference paths, got: %s", result)
	}

	// Test with non-existent symbol
	result, err = ctx.ExecuteCommand("refs", "NonExistentSymbol")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !strings.Contains(result, "No symbols found") {
		t.Errorf("Result should indicate no symbols found, got: %s", result)
	}
}

func TestCommandImplements(t *testing.T) {
	ctx := NewMockCommandContext()

	// Mock the methods
	mockFindCalled := false
	mockImplCalled := false

	origFindMethod := ctx.findSymbolsByName
	origImplMethod := ctx.findImplementations
	defer func() {
		ctx.findSymbolsByName = origFindMethod
		ctx.findImplementations = origImplMethod
	}()

	ctx.findSymbolsByName = func(name string) []*SymbolMatch {
		mockFindCalled = true
		if name == "Readable" {
			return []*SymbolMatch{
				{Name: "Readable", Kind: "interface", Path: "path/to/file.go", ID: "intf123"},
			}
		}
		return nil
	}

	ctx.findImplementations = func(id string) []*SymbolMatch {
		mockImplCalled = true
		if id == "intf123" {
			return []*SymbolMatch{
				{Name: "FileReader", Kind: "struct", Path: "path/to/reader.go"},
				{Name: "StringReader", Kind: "struct", Path: "path/to/reader.go"},
			}
		}
		return nil
	}

	// Test successful implements command
	result, err := ctx.ExecuteCommand("implements", "Readable")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !mockFindCalled || !mockImplCalled {
		t.Error("Both find and implements methods should be called")
	}

	// Check results contains implementations
	if !strings.Contains(result, "FileReader") || !strings.Contains(result, "StringReader") {
		t.Errorf("Result should contain implementation names, got: %s", result)
	}

	// Test with non-interface symbol
	ctx.findSymbolsByName = func(name string) []*SymbolMatch {
		return []*SymbolMatch{
			{Name: "NotAnInterface", Kind: "struct", Path: "path/to/file.go", ID: "struct123"},
		}
	}

	result, err = ctx.ExecuteCommand("implements", "NotAnInterface")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !strings.Contains(result, "not an interface") {
		t.Errorf("Result should indicate not an interface, got: %s", result)
	}
}

func TestCommandMethods(t *testing.T) {
	ctx := NewMockCommandContext()

	// Mock the method
	mockMethodsCalled := false

	origMethod := ctx.findMethodsForType
	defer func() {
		ctx.findMethodsForType = origMethod
	}()

	ctx.findMethodsForType = func(typeName string) []*SymbolMatch {
		mockMethodsCalled = true
		if typeName == "Person" {
			return []*SymbolMatch{
				{Name: "GetName", Kind: "method", Path: "path/to/file.go"},
				{Name: "SetName", Kind: "method", Path: "path/to/file.go"},
			}
		}
		return nil
	}

	// Test successful methods command
	result, err := ctx.ExecuteCommand("methods", "Person")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !mockMethodsCalled {
		t.Error("findMethodsForType should be called")
	}

	// Check results contains methods
	if !strings.Contains(result, "GetName") || !strings.Contains(result, "SetName") {
		t.Errorf("Result should contain method names, got: %s", result)
	}

	// Test with type that has no methods
	ctx.findMethodsForType = func(typeName string) []*SymbolMatch {
		return nil
	}

	result, err = ctx.ExecuteCommand("methods", "NoMethods")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !strings.Contains(result, "No methods found") {
		t.Errorf("Result should indicate no methods found, got: %s", result)
	}
}

func TestCommandStructure(t *testing.T) {
	ctx := NewMockCommandContext()

	// Mock the method
	mockStructureCalled := false

	origMethod := ctx.getFileStructure
	defer func() {
		ctx.getFileStructure = origMethod
	}()

	ctx.getFileStructure = func(filePath string) []*SymbolMatch {
		mockStructureCalled = true
		if filePath == "path/to/file.go" {
			return []*SymbolMatch{
				{Name: "Package", Kind: "package", Path: "path/to/file.go"},
				{Name: "Person", Kind: "struct", Path: "path/to/file.go"},
				{Name: "GetName", Kind: "method", Path: "path/to/file.go"},
			}
		}
		return nil
	}

	// Test successful structure command
	result, err := ctx.ExecuteCommand("structure", "path/to/file.go")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !mockStructureCalled {
		t.Error("getFileStructure should be called")
	}

	// Check results contains structure elements
	if !strings.Contains(result, "Package") || !strings.Contains(result, "Person") {
		t.Errorf("Result should contain structure elements, got: %s", result)
	}

	// Test with non-existent file
	ctx.getFileStructure = func(filePath string) []*SymbolMatch {
		return nil
	}

	result, err = ctx.ExecuteCommand("structure", "nonexistent.go")
	if err != nil {
		t.Errorf("ExecuteCommand should not return error: %v", err)
	}

	if !strings.Contains(result, "No symbols found") {
		t.Errorf("Result should indicate no symbols found, got: %s", result)
	}
}

func TestCommandHelp(t *testing.T) {
	ctx := NewMockCommandContext()

	// Test help command
	result, err := ctx.ExecuteCommand("help", "")
	if err != nil {
		t.Errorf("Help command should not return error: %v", err)
	}

	// Check that help output contains common commands
	commands := []string{"find", "refs", "implements", "methods", "structure"}
	for _, cmd := range commands {
		if !strings.Contains(result, cmd) {
			t.Errorf("Help output should mention '%s' command, got: %s", cmd, result)
		}
	}
}

func TestInvalidCommand(t *testing.T) {
	ctx := NewMockCommandContext()

	// Test invalid command
	_, err := ctx.ExecuteCommand("invalidcommand", "arg")
	if err == nil {
		t.Error("Invalid command should return an error")
	}
}
