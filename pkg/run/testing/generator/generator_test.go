package generator

import (
	"bitspark.dev/go-tree/pkg/run/common"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TestMockGenerator implements the TestGenerator interface for testing
type TestMockGenerator struct {
	GenerateTestsResult    *common.TestSuite
	GenerateTestsError     error
	GenerateMockResult     string
	GenerateMockError      error
	GenerateTestDataResult interface{}
	GenerateTestDataError  error
	GenerateTestsCalled    bool
	GenerateMockCalled     bool
	GenerateTestDataCalled bool
	SymbolTested           *typesys.Symbol
}

func (m *TestMockGenerator) GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error) {
	m.GenerateTestsCalled = true
	m.SymbolTested = sym
	return m.GenerateTestsResult, m.GenerateTestsError
}

func (m *TestMockGenerator) GenerateMock(iface *typesys.Symbol) (string, error) {
	m.GenerateMockCalled = true
	m.SymbolTested = iface
	return m.GenerateMockResult, m.GenerateMockError
}

func (m *TestMockGenerator) GenerateTestData(typ *typesys.Symbol) (interface{}, error) {
	m.GenerateTestDataCalled = true
	m.SymbolTested = typ
	return m.GenerateTestDataResult, m.GenerateTestDataError
}

// createSimpleSymbol creates a simple symbol for testing functions that don't require type information
func createSimpleSymbol(name string, kind typesys.SymbolKind, pkgName string) *typesys.Symbol {
	pkg := &typesys.Package{
		Name:       pkgName,
		ImportPath: "github.com/example/" + pkgName,
	}

	return &typesys.Symbol{
		Name:    name,
		Kind:    kind,
		Package: pkg,
	}
}

// TestGeneratorInterfaceConformance verifies our mock objects conform to interfaces
func TestGeneratorInterfaceConformance(t *testing.T) {
	// Test that TestMockGenerator implements TestGenerator
	var _ TestGenerator = &TestMockGenerator{}
}

// TestNewGenerator tests creating a new generator
func TestNewGenerator(t *testing.T) {
	// Create a module
	mod := &typesys.Module{
		Path: "test-module",
	}

	// Create a generator
	gen := NewGenerator(mod)

	// Verify the generator was created properly
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if gen.Module != mod {
		t.Error("Generator has incorrect module reference")
	}

	if gen.Analyzer == nil {
		t.Error("Generator analyzer is nil")
	}

	// Check that templates were initialized
	requiredTemplates := []string{"basic", "table", "parallel", "mock"}
	for _, tmplName := range requiredTemplates {
		if _, exists := gen.templates[tmplName]; !exists {
			t.Errorf("Template %s not initialized", tmplName)
		}
	}
}

// TestFactory tests the factory pattern for creating generators
func TestFactory(t *testing.T) {
	// Create a factory function
	mockGenerator := &TestMockGenerator{}
	factory := func(mod *typesys.Module) TestGenerator {
		return mockGenerator
	}

	// Create a module
	mod := &typesys.Module{
		Path: "test-module",
	}

	// Call the factory
	generator := factory(mod)
	if generator != mockGenerator {
		t.Error("Factory did not return the expected generator")
	}
}

// TestAnalyzerFunctionNeedsTests tests the FunctionNeedsTests method which doesn't require complex type info
func TestAnalyzerFunctionNeedsTests(t *testing.T) {
	// Create a module
	mod := &typesys.Module{
		Path: "test-module",
	}

	// Create an analyzer
	analyzer := NewAnalyzer(mod)

	// Test with nil symbol
	if analyzer.FunctionNeedsTests(nil) {
		t.Error("Expected FunctionNeedsTests to return false for nil symbol")
	}

	// Test with different symbol kinds
	testCases := []struct {
		name     string
		symbol   *typesys.Symbol
		expected bool
	}{
		{
			name:     "regular function",
			symbol:   createSimpleSymbol("ProcessData", typesys.KindFunction, "testpkg"),
			expected: true,
		},
		{
			name:     "test function",
			symbol:   createSimpleSymbol("TestProcessData", typesys.KindFunction, "testpkg"),
			expected: false,
		},
		{
			name:     "benchmark function",
			symbol:   createSimpleSymbol("BenchmarkProcessData", typesys.KindFunction, "testpkg"),
			expected: false,
		},
		{
			name:     "struct type",
			symbol:   createSimpleSymbol("User", typesys.KindStruct, "testpkg"),
			expected: false,
		},
		{
			name:     "interface type",
			symbol:   createSimpleSymbol("Handler", typesys.KindInterface, "testpkg"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.FunctionNeedsTests(tc.symbol)
			if result != tc.expected {
				t.Errorf("Expected FunctionNeedsTests to return %v for %s, got %v",
					tc.expected, tc.name, result)
			}
		})
	}
}

// TestMockGeneratorCalls tests that our mock generator correctly records calls
func TestMockGeneratorCalls(t *testing.T) {
	mockGen := &TestMockGenerator{
		GenerateTestsResult: &common.TestSuite{
			PackageName: "testpkg",
		},
		GenerateMockResult:     "mock implementation",
		GenerateTestDataResult: "test data",
	}

	// Test GenerateTests
	sym := createSimpleSymbol("TestFunc", typesys.KindFunction, "testpkg")
	result, err := mockGen.GenerateTests(sym)

	if !mockGen.GenerateTestsCalled {
		t.Error("GenerateTests call not recorded")
	}

	if mockGen.SymbolTested != sym {
		t.Error("Symbol not correctly passed to GenerateTests")
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.PackageName != "testpkg" {
		t.Errorf("Expected PackageName 'testpkg', got '%s'", result.PackageName)
	}

	// Test GenerateMock
	iface := createSimpleSymbol("Handler", typesys.KindInterface, "testpkg")
	mockResult, _ := mockGen.GenerateMock(iface)

	if !mockGen.GenerateMockCalled {
		t.Error("GenerateMock call not recorded")
	}

	if mockGen.SymbolTested != iface {
		t.Error("Symbol not correctly passed to GenerateMock")
	}

	if mockResult != "mock implementation" {
		t.Errorf("Expected 'mock implementation', got '%s'", mockResult)
	}

	// Test GenerateTestData
	typ := createSimpleSymbol("User", typesys.KindStruct, "testpkg")
	dataResult, _ := mockGen.GenerateTestData(typ)

	if !mockGen.GenerateTestDataCalled {
		t.Error("GenerateTestData call not recorded")
	}

	if mockGen.SymbolTested != typ {
		t.Error("Symbol not correctly passed to GenerateTestData")
	}

	if dataResult != "test data" {
		t.Errorf("Expected 'test data', got '%v'", dataResult)
	}
}

// TestRegisterFactoryFunction tests the factory registration
func TestRegisterFactoryFunction(t *testing.T) {
	// Create a mock factory
	mockFactory := func(mod *typesys.Module) TestGenerator {
		return &TestMockGenerator{
			GenerateTestsResult: &common.TestSuite{
				PackageName: "customsuite",
			},
		}
	}

	// Call the factory
	mod := &typesys.Module{Path: "test-module"}
	generator := mockFactory(mod)

	if generator == nil {
		t.Error("Factory returned nil generator")
	}

	mockGen, ok := generator.(*TestMockGenerator)
	if !ok {
		t.Error("Factory returned wrong type")
	} else {
		if mockGen.GenerateTestsResult.PackageName != "customsuite" {
			t.Errorf("Factory returned wrong generator, expected PackageName 'customsuite', got '%s'",
				mockGen.GenerateTestsResult.PackageName)
		}
	}
}
