package extract

import (
	"testing"

	"bitspark.dev/go-tree/pkg/typesys"
	"github.com/stretchr/testify/assert"
)

// createTestModule creates a test module with types that have common method patterns
func createTestModule() *typesys.Module {
	module := &typesys.Module{
		Path:     "test/module",
		Dir:      "/test/module",
		Packages: make(map[string]*typesys.Package),
		FileSet:  nil, // In a real test, would initialize this
	}

	// Create a package
	pkg := &typesys.Package{
		Name:       "testpkg",
		ImportPath: "test/module/testpkg",
		Dir:        "/test/module/testpkg",
		Module:     module,
		Files:      make(map[string]*typesys.File),
		Symbols:    make(map[string]*typesys.Symbol),
	}
	module.Packages[pkg.ImportPath] = pkg

	// Create a file
	file := &typesys.File{
		Path:    "/test/module/testpkg/file.go",
		Name:    "file.go",
		Package: pkg,
		Symbols: []*typesys.Symbol{}, // Will add symbols here
	}
	pkg.Files[file.Path] = file

	// Create first struct type (FileReader)
	type1 := &typesys.Symbol{
		ID:      "type_FileReader",
		Name:    "FileReader",
		Kind:    typesys.KindStruct,
		File:    file,
		Package: pkg,
	}

	// Create second struct type (BufferReader)
	type2 := &typesys.Symbol{
		ID:      "type_BufferReader",
		Name:    "BufferReader",
		Kind:    typesys.KindStruct,
		File:    file,
		Package: pkg,
	}

	// Create third struct type (HttpHandler)
	type3 := &typesys.Symbol{
		ID:      "type_HttpHandler",
		Name:    "HttpHandler",
		Kind:    typesys.KindStruct,
		File:    file,
		Package: pkg,
	}

	// Create fourth struct type (WebSocketHandler)
	type4 := &typesys.Symbol{
		ID:      "type_WebSocketHandler",
		Name:    "WebSocketHandler",
		Kind:    typesys.KindStruct,
		File:    file,
		Package: pkg,
	}

	// Create methods for FileReader
	readMethod1 := &typesys.Symbol{
		ID:      "method_FileReader_Read",
		Name:    "Read",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type1, // Indicates this is a method of FileReader
	}

	closeMethod1 := &typesys.Symbol{
		ID:      "method_FileReader_Close",
		Name:    "Close",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type1, // Indicates this is a method of FileReader
	}

	// Create methods for BufferReader
	readMethod2 := &typesys.Symbol{
		ID:      "method_BufferReader_Read",
		Name:    "Read",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type2, // Indicates this is a method of BufferReader
	}

	closeMethod2 := &typesys.Symbol{
		ID:      "method_BufferReader_Close",
		Name:    "Close",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type2, // Indicates this is a method of BufferReader
	}

	// Create methods for HttpHandler
	handleMethod1 := &typesys.Symbol{
		ID:      "method_HttpHandler_Handle",
		Name:    "Handle",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type3, // Indicates this is a method of HttpHandler
	}

	// Create methods for WebSocketHandler
	handleMethod2 := &typesys.Symbol{
		ID:      "method_WebSocketHandler_Handle",
		Name:    "Handle",
		Kind:    typesys.KindMethod,
		File:    file,
		Package: pkg,
		Parent:  type4, // Indicates this is a method of WebSocketHandler
	}

	// Add symbols to file
	file.Symbols = append(file.Symbols,
		type1, type2, type3, type4,
		readMethod1, closeMethod1, readMethod2, closeMethod2,
		handleMethod1, handleMethod2)

	// Add symbols to package
	pkg.Symbols[type1.ID] = type1
	pkg.Symbols[type2.ID] = type2
	pkg.Symbols[type3.ID] = type3
	pkg.Symbols[type4.ID] = type4
	pkg.Symbols[readMethod1.ID] = readMethod1
	pkg.Symbols[closeMethod1.ID] = closeMethod1
	pkg.Symbols[readMethod2.ID] = readMethod2
	pkg.Symbols[closeMethod2.ID] = closeMethod2
	pkg.Symbols[handleMethod1.ID] = handleMethod1
	pkg.Symbols[handleMethod2.ID] = handleMethod2

	return module
}

// Note on testing approach:
//
// In a real production environment, we would use one of the following approaches:
// 1. Create a proper mocking framework for the index
// 2. Use interface abstraction in the transform package instead of concrete types
// 3. Test with real files and a built index
//
// For this implementation, we're using smoke tests and option tests only.
// Full integration tests would require a more sophisticated setup.

// TestExtractor runs all tests for the interface extractor
func TestExtractor(t *testing.T) {
	t.Run("SmokeTest", func(t *testing.T) {
		// Create interface extractor with default options
		extractor := NewInterfaceExtractor(DefaultOptions())

		// Just test that the transformer can be created without errors
		assert.NotNil(t, extractor)
		assert.Equal(t, "InterfaceExtractor", extractor.Name())
		assert.Contains(t, extractor.Description(), "interface")
	})

	t.Run("OptionsTest", func(t *testing.T) {
		// Create options with different settings
		options := Options{
			MinimumTypes:    3, // Higher threshold
			MinimumMethods:  2, // Only interfaces with at least 2 methods
			MethodThreshold: 0.9,
			NamingStrategy: func(types []*typesys.Symbol, methodNames []string) string {
				return "Custom" // Always return Custom as name
			},
			ExcludeMethods: []string{"Close"}, // Exclude Close method
		}

		// Verify option values
		assert.Equal(t, 3, options.MinimumTypes)
		assert.Equal(t, 2, options.MinimumMethods)
		assert.Equal(t, 0.9, options.MethodThreshold)
		assert.NotNil(t, options.NamingStrategy)

		// Test exclude methods functionality
		assert.True(t, options.IsExcludedMethod("Close"))
		assert.False(t, options.IsExcludedMethod("Read"))
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to count symbols in a module
func countSymbols(module *typesys.Module) int {
	count := 0
	for _, pkg := range module.Packages {
		count += len(pkg.Symbols)
	}
	return count
}
