package extract

import (
	"testing"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"github.com/stretchr/testify/assert"
)

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
