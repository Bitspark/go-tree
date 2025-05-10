// Package visual provides interfaces and implementations for visualizing Go modules with type information.
package visual

import (
	"bitspark.dev/go-tree/pkg/core/typesys"
)

// TypeAwareVisualizer creates visual representations of a module with full type information
type TypeAwareVisualizer interface {
	// Visualize creates a visualization with type information
	Visualize(module *typesys.Module, opts *VisualizationOptions) ([]byte, error)

	// Format returns the output format (e.g., "html", "markdown")
	Format() string

	// SupportsTypeAnnotations indicates if this visualizer can show type info
	SupportsTypeAnnotations() bool
}

// VisualizationOptions controls visualization behavior
type VisualizationOptions struct {
	// Whether to include type annotations in the output
	IncludeTypeAnnotations bool

	// Whether to include private (unexported) elements
	IncludePrivate bool

	// Whether to include test files in the visualization
	IncludeTests bool

	// Level of detail to include (1=minimal, 5=complete)
	DetailLevel int

	// Symbol to highlight in the visualization (if any)
	HighlightSymbol *typesys.Symbol

	// Custom title for the visualization
	Title string

	// Whether to include generated files
	IncludeGenerated bool

	// Whether to show relationships between symbols
	ShowRelationships bool

	// Style customization options (implementation-specific)
	StyleOptions map[string]interface{}
}

// VisualizerRegistry maintains a collection of available visualizers
type VisualizerRegistry struct {
	visualizers map[string]TypeAwareVisualizer
}

// NewVisualizerRegistry creates a new registry
func NewVisualizerRegistry() *VisualizerRegistry {
	return &VisualizerRegistry{
		visualizers: make(map[string]TypeAwareVisualizer),
	}
}

// Register adds a visualizer to the registry
func (r *VisualizerRegistry) Register(v TypeAwareVisualizer) {
	r.visualizers[v.Format()] = v
}

// Get returns a visualizer by format name
func (r *VisualizerRegistry) Get(format string) TypeAwareVisualizer {
	return r.visualizers[format]
}

// Available returns a list of available visualizer format names
func (r *VisualizerRegistry) Available() []string {
	formats := make([]string, 0, len(r.visualizers))
	for format := range r.visualizers {
		formats = append(formats, format)
	}
	return formats
}
