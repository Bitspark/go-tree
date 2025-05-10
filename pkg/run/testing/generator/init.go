package generator

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/common"
	"bitspark.dev/go-tree/pkg/run/testing"
)

// init registers the generator factory with the testing package
func init() {
	// Register our generator factory
	testing.RegisterGeneratorFactory(createGenerator)
}

// createGenerator creates a generator that implements the testing.TestGenerator interface
func createGenerator(mod *typesys.Module) testing.TestGenerator {
	// Create the real generator
	gen := NewGenerator(mod)

	// Wrap it in an adapter to match the testing.TestGenerator interface
	return &generatorAdapter{gen: gen}
}

// generatorAdapter adapts Generator to the testing.TestGenerator interface
type generatorAdapter struct {
	gen *Generator
}

// GenerateTests implements testing.TestGenerator.GenerateTests
func (a *generatorAdapter) GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error) {
	return a.gen.GenerateTests(sym)
}

// GenerateMock implements testing.TestGenerator.GenerateMock
func (a *generatorAdapter) GenerateMock(iface *typesys.Symbol) (string, error) {
	return a.gen.GenerateMock(iface)
}

// GenerateTestData implements testing.TestGenerator.GenerateTestData
func (a *generatorAdapter) GenerateTestData(typ *typesys.Symbol) (interface{}, error) {
	return a.gen.GenerateTestData(typ)
}
