package markdown

import (
	"bitspark.dev/go-tree/pkg/typesys"
	"bitspark.dev/go-tree/pkg/visual/formatter"
)

// VisualizationOptions provides options for Markdown visualization
type VisualizationOptions struct {
	IncludeTypeAnnotations bool
	IncludePrivate         bool
	IncludeTests           bool
	DetailLevel            int
	HighlightSymbol        *typesys.Symbol
	Title                  string
	IncludeGenerated       bool
	ShowRelationships      bool
	StyleOptions           map[string]interface{}
}

// MarkdownVisualizer creates Markdown visualizations of a Go module with full type information
type MarkdownVisualizer struct{}

// NewMarkdownVisualizer creates a new Markdown visualizer
func NewMarkdownVisualizer() *MarkdownVisualizer {
	return &MarkdownVisualizer{}
}

// Visualize creates a Markdown visualization of the module
func (v *MarkdownVisualizer) Visualize(module *typesys.Module, opts *VisualizationOptions) ([]byte, error) {
	if opts == nil {
		opts = &VisualizationOptions{
			DetailLevel: 3,
		}
	}

	// Create formatter options from visualization options
	formatOpts := &formatter.FormatOptions{
		IncludeTypeAnnotations: opts.IncludeTypeAnnotations,
		IncludePrivate:         opts.IncludePrivate,
		IncludeTests:           opts.IncludeTests,
		DetailLevel:            opts.DetailLevel,
		HighlightSymbol:        opts.HighlightSymbol,
		IncludeGenerated:       opts.IncludeGenerated,
	}

	// Create a visitor to traverse the module
	visitor := NewMarkdownVisitor(formatOpts)

	// Walk the module with the visitor
	if err := typesys.Walk(visitor, module); err != nil {
		return nil, err
	}

	// Get the content from the visitor
	content, err := visitor.Result()
	if err != nil {
		return nil, err
	}

	// Add a title if one was provided
	if opts.Title != "" {
		header := "# " + opts.Title + "\n\n"
		content = header + content
	}

	return []byte(content), nil
}

// Format returns the output format name
func (v *MarkdownVisualizer) Format() string {
	return "markdown"
}

// SupportsTypeAnnotations indicates if this visualizer can show type info
func (v *MarkdownVisualizer) SupportsTypeAnnotations() bool {
	return true
}
