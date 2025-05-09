package testing

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/testing/common"
)

func TestDefaultTestGenerator(t *testing.T) {
	// Create a mock module
	mod := &typesys.Module{
		Path: "test-module",
	}

	// Test default generator creation
	generator := DefaultTestGenerator(mod)
	if generator == nil {
		t.Error("DefaultTestGenerator returned nil")
	}

	// Test that it's the expected type when no factory is registered
	_, isNullGen := generator.(*nullGenerator)
	if !isNullGen {
		t.Error("Expected nullGenerator when no factory is registered")
	}

	// Test with a registered factory
	oldFactory := generatorFactory
	defer func() { generatorFactory = oldFactory }()

	// Register a mock factory
	mockGenerator := &mockTestGenerator{}
	generatorFactory = func(m *typesys.Module) TestGenerator {
		if m != mod {
			t.Error("Factory called with wrong module")
		}
		return mockGenerator
	}

	// Now test again
	generator = DefaultTestGenerator(mod)
	if generator != mockGenerator {
		t.Error("DefaultTestGenerator did not use registered factory")
	}
}

func TestDefaultTestRunner(t *testing.T) {
	// Test default runner creation
	runner := DefaultTestRunner()
	if runner == nil {
		t.Error("DefaultTestRunner returned nil")
	}

	// Test that it's the expected type when no factory is registered
	_, isNullRunner := runner.(*nullRunner)
	if !isNullRunner {
		t.Error("Expected nullRunner when no factory is registered")
	}

	// Test with a registered factory
	oldFactory := runnerFactory
	defer func() { runnerFactory = oldFactory }()

	// Register a mock factory
	mockRunner := &mockTestRunner{}
	runnerFactory = func() TestRunner {
		return mockRunner
	}

	// Now test again
	runner = DefaultTestRunner()
	if runner != mockRunner {
		t.Error("DefaultTestRunner did not use registered factory")
	}
}

func TestGenerateTestsWithDefaults(t *testing.T) {
	// Create a mock module and symbol
	mod := &typesys.Module{
		Path: "test-module",
	}
	sym := &typesys.Symbol{
		Name: "TestSymbol",
		Package: &typesys.Package{
			Name: "test-package",
		},
	}

	// Test with default generator
	oldFactory := generatorFactory
	defer func() { generatorFactory = oldFactory }()

	// Set up a mock generator that returns a specific test suite
	expectedSuite := &common.TestSuite{
		PackageName: "test-package",
		Tests:       []*common.Test{},
		SourceCode:  "// Test source code",
	}
	mockGenerator := &mockTestGenerator{
		suite: expectedSuite,
	}
	generatorFactory = func(m *typesys.Module) TestGenerator {
		return mockGenerator
	}

	// Call the function
	suite, err := GenerateTestsWithDefaults(mod, sym)
	if err != nil {
		t.Errorf("GenerateTestsWithDefaults returned error: %v", err)
	}
	if suite != expectedSuite {
		t.Error("GenerateTestsWithDefaults didn't return expected suite")
	}
}

func TestNullGenerator(t *testing.T) {
	// Create a null generator
	mod := &typesys.Module{Path: "test-module"}
	generator := &nullGenerator{mod: mod}

	// Test GenerateTests
	sym := &typesys.Symbol{
		Package: &typesys.Package{Name: "test-package"},
	}
	suite, err := generator.GenerateTests(sym)
	if err != nil {
		t.Errorf("nullGenerator.GenerateTests returned error: %v", err)
	}
	if suite.PackageName != sym.Package.Name {
		t.Errorf("Expected package name %s, got %s", sym.Package.Name, suite.PackageName)
	}
	if len(suite.Tests) != 0 {
		t.Errorf("Expected empty test slice, got %d tests", len(suite.Tests))
	}

	// Test GenerateMock
	mock, err := generator.GenerateMock(sym)
	if err != nil {
		t.Errorf("nullGenerator.GenerateMock returned error: %v", err)
	}
	if mock != "// Not implemented" {
		t.Errorf("Expected comment string, got: %s", mock)
	}

	// Test GenerateTestData
	data, err := generator.GenerateTestData(sym)
	if err != nil {
		t.Errorf("nullGenerator.GenerateTestData returned error: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil data, got: %v", data)
	}
}

func TestNullRunner(t *testing.T) {
	// Create a null runner
	runner := &nullRunner{}

	// Test RunTests
	mod := &typesys.Module{Path: "test-module"}
	result, err := runner.RunTests(mod, "test/package", &common.RunOptions{})
	if err != nil {
		t.Errorf("nullRunner.RunTests returned error: %v", err)
	}
	if result.Package != "test/package" {
		t.Errorf("Expected package 'test/package', got %s", result.Package)
	}
	if len(result.Tests) != 0 {
		t.Errorf("Expected empty tests slice, got %d tests", len(result.Tests))
	}

	// Test AnalyzeCoverage
	coverage, err := runner.AnalyzeCoverage(mod, "test/package")
	if err != nil {
		t.Errorf("nullRunner.AnalyzeCoverage returned error: %v", err)
	}
	if coverage.Percentage != 0.0 {
		t.Errorf("Expected 0.0 coverage, got %f", coverage.Percentage)
	}
}

// Mock implementations for testing

type mockTestGenerator struct {
	suite   *common.TestSuite
	mockStr string
	data    interface{}
}

func (g *mockTestGenerator) GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error) {
	return g.suite, nil
}

func (g *mockTestGenerator) GenerateMock(iface *typesys.Symbol) (string, error) {
	return g.mockStr, nil
}

func (g *mockTestGenerator) GenerateTestData(typ *typesys.Symbol) (interface{}, error) {
	return g.data, nil
}

type mockTestRunner struct {
	result   *common.TestResult
	coverage *common.CoverageResult
}

func (r *mockTestRunner) RunTests(mod *typesys.Module, pkgPath string, opts *common.RunOptions) (*common.TestResult, error) {
	return r.result, nil
}

func (r *mockTestRunner) AnalyzeCoverage(mod *typesys.Module, pkgPath string) (*common.CoverageResult, error) {
	return r.coverage, nil
}
