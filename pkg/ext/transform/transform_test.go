package transform

import (
	"bitspark.dev/go-tree/pkg/core/index"
	"fmt"
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
)

// MockTransformer implements the Transformer interface for testing
type MockTransformer struct {
	name        string
	description string
	result      *TransformResult
	err         error
	validateErr error
}

func NewMockTransformer(name string, result *TransformResult, err error) *MockTransformer {
	return &MockTransformer{
		name:        name,
		description: "Mock transformer for testing",
		result:      result,
		err:         err,
	}
}

func (m *MockTransformer) Transform(ctx *Context) (*TransformResult, error) {
	return m.result, m.err
}

func (m *MockTransformer) Validate(ctx *Context) error {
	return m.validateErr
}

func (m *MockTransformer) Name() string {
	return m.name
}

func (m *MockTransformer) Description() string {
	return m.description
}

// Test helper to create a basic test module
func createTestModule() *typesys.Module {
	// Create a simple module with a package and some files
	module := &typesys.Module{
		Path:     "test/module",
		Dir:      "/test/module",
		Packages: make(map[string]*typesys.Package),
	}

	// Add a package
	pkg := &typesys.Package{
		Name:       "testpkg",
		ImportPath: "test/module/testpkg",
		Dir:        "/test/module/testpkg",
		Module:     module,
		Files:      make(map[string]*typesys.File),
		Symbols:    make(map[string]*typesys.Symbol),
	}
	module.Packages[pkg.ImportPath] = pkg

	// Add a file
	file := &typesys.File{
		Path:    "/test/module/testpkg/file.go",
		Name:    "file.go",
		Package: pkg,
	}
	pkg.Files[file.Path] = file

	return module
}

// TestNewContext tests the creation of a transformation context
func TestNewContext(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := index.NewIndex(module)

	// Create context
	ctx := NewContext(module, idx, false)

	// Verify context properties
	assert.Equal(t, module, ctx.Module)
	assert.Equal(t, idx, ctx.Index)
	assert.False(t, ctx.DryRun)
	assert.NotNil(t, ctx.Options)
	assert.NotNil(t, ctx.state)
}

// TestSetOption tests setting options in the context
func TestSetOption(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := index.NewIndex(module)

	// Create context
	ctx := NewContext(module, idx, false)

	// Set options
	ctx.SetOption("testOption", "testValue")
	ctx.SetOption("numOption", 42)

	// Verify options
	assert.Equal(t, "testValue", ctx.Options["testOption"])
	assert.Equal(t, 42, ctx.Options["numOption"])
}

// TestChainedTransformer tests the chained transformer implementation
func TestChainedTransformer(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := index.NewIndex(module)

	// Create context
	ctx := NewContext(module, idx, false)

	// Create mock transformers
	mock1 := NewMockTransformer("Mock1", &TransformResult{
		Summary:       "Mock1 result",
		Success:       true,
		AffectedFiles: []string{"file1.go"},
		Changes: []Change{
			{FilePath: "file1.go", Original: "old1", New: "new1"},
		},
	}, nil)

	mock2 := NewMockTransformer("Mock2", &TransformResult{
		Summary:       "Mock2 result",
		Success:       true,
		AffectedFiles: []string{"file2.go"},
		Changes: []Change{
			{FilePath: "file2.go", Original: "old2", New: "new2"},
		},
	}, nil)

	// Create chained transformer
	chain := NewChainedTransformer("TestChain", "Test chain transformer", mock1, mock2)

	// Verify chain properties
	assert.Equal(t, "TestChain", chain.Name())
	assert.Equal(t, "Test chain transformer", chain.Description())

	// Test transform
	result, err := chain.Transform(ctx)

	// Verify result
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.AffectedFiles, 2)
	assert.Contains(t, result.AffectedFiles, "file1.go")
	assert.Contains(t, result.AffectedFiles, "file2.go")
	assert.Len(t, result.Changes, 2)
}

// TestChainedTransformerError tests handling of errors in chained transformers
func TestChainedTransformerError(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := index.NewIndex(module)

	// Create context
	ctx := NewContext(module, idx, false)

	// Create mock transformers
	mock1 := NewMockTransformer("Mock1", &TransformResult{
		Summary: "Mock1 result",
		Success: true,
	}, nil)

	// Mock2 will return an error
	mock2 := NewMockTransformer("Mock2", &TransformResult{
		Summary: "Mock2 result",
		Success: false,
		Error:   fmt.Errorf("mock error"),
	}, nil)

	// Create chained transformer
	chain := NewChainedTransformer("TestChain", "Test chain transformer", mock1, mock2)

	// Test transform
	result, err := chain.Transform(ctx)

	// Verify result
	assert.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "mock error", result.Error.Error())
}

// TestChainedTransformerValidate tests the validate method of chained transformers
func TestChainedTransformerValidate(t *testing.T) {
	// Create test module and index
	module := createTestModule()
	idx := index.NewIndex(module)

	// Create context
	ctx := NewContext(module, idx, false)

	// Create mock transformers
	mock1 := NewMockTransformer("Mock1", nil, nil)
	mock2 := NewMockTransformer("Mock2", nil, nil)

	// Set validation error on mock2
	mock2.validateErr = fmt.Errorf("validation error")

	// Create chained transformer
	chain := NewChainedTransformer("TestChain", "Test chain transformer", mock1, mock2)

	// Test validate
	err := chain.Validate(ctx)

	// Verify result
	assert.Error(t, err)
	assert.Equal(t, "validation error", err.Error())
}

// TestTransformResult tests the TransformResult struct
func TestTransformResult(t *testing.T) {
	// Create a transform result
	result := &TransformResult{
		Summary:       "Test result",
		Details:       "Test details",
		FilesAffected: 2,
		Success:       true,
		Error:         nil,
		IsDryRun:      false,
		AffectedFiles: []string{"file1.go", "file2.go"},
		Changes: []Change{
			{
				FilePath:  "file1.go",
				StartLine: 10,
				EndLine:   15,
				Original:  "old code",
				New:       "new code",
			},
		},
	}

	// Verify result properties
	assert.Equal(t, "Test result", result.Summary)
	assert.Equal(t, "Test details", result.Details)
	assert.Equal(t, 2, result.FilesAffected)
	assert.True(t, result.Success)
	assert.Nil(t, result.Error)
	assert.False(t, result.IsDryRun)
	assert.Len(t, result.AffectedFiles, 2)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "file1.go", result.Changes[0].FilePath)
	assert.Equal(t, "old code", result.Changes[0].Original)
	assert.Equal(t, "new code", result.Changes[0].New)
}
