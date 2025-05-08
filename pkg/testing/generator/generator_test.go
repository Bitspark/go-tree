package generator

import (
	"testing"

	"bitspark.dev/go-tree/pkg/testing/common"
	"bitspark.dev/go-tree/pkg/typesys"
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

// Create mock structures for testing
type MockSymbol struct {
	*typesys.Symbol
}

func createMockSymbol(name string, kind typesys.SymbolKind) *typesys.Symbol {
	return &typesys.Symbol{
		Name: name,
		Kind: kind,
		Package: &typesys.Package{
			Name:       "mockpkg",
			ImportPath: "github.com/example/mockpkg",
		},
	}
}

// TestGeneratorInterfaceConformance verifies our mock objects conform to interfaces
func TestGeneratorInterfaceConformance(t *testing.T) {
	// Test that TestMockGenerator implements TestGenerator
	var _ TestGenerator = &TestMockGenerator{}
}

// At this point, we would test specific generator implementations
// Since the actual generator code is complex, we'll add more specialized
// tests for each generator type in separate test files.

// For example, here's a simple test for a generator factory registration
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
