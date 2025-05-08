// Package generator provides functionality for generating tests
// based on the type system.
package generator

import (
	"bitspark.dev/go-tree/pkg/testing/common"
	"bitspark.dev/go-tree/pkg/typesys"
)

// TestGenerator generates tests for Go code
type TestGenerator interface {
	// GenerateTests generates tests for a symbol
	GenerateTests(sym *typesys.Symbol) (*common.TestSuite, error)

	// GenerateMock generates a mock implementation of an interface
	GenerateMock(iface *typesys.Symbol) (string, error)

	// GenerateTestData generates test data with correct types
	GenerateTestData(typ *typesys.Symbol) (interface{}, error)
}

// Factory is a factory method type for creating test generators
type Factory func(mod *typesys.Module) TestGenerator
