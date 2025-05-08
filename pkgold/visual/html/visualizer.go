package html

import (
	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/core/visitor"
	"bitspark.dev/go-tree/pkgold/visual"
)

// Options defines configuration options for the HTML visualizer
type Options struct {
	// Embed the common base options
	visual.BaseVisualizerOptions

	// Additional HTML-specific options could be added here
	IncludeCSS bool   // Whether to include CSS in the HTML output
	CustomCSS  string // Custom CSS to include
}

// HTMLVisualizer implements the ModuleVisualizer interface for generating
// HTML documentation from Go modules
type HTMLVisualizer struct {
	options Options
}

// NewHTMLVisualizer creates a new HTML visualizer with the given options
func NewHTMLVisualizer(options Options) *HTMLVisualizer {
	return &HTMLVisualizer{
		options: options,
	}
}

// DefaultOptions returns the default options for the HTML visualizer
func DefaultOptions() Options {
	return Options{
		BaseVisualizerOptions: visual.BaseVisualizerOptions{
			IncludePrivate:   false,
			IncludeTests:     false,
			IncludeGenerated: false,
			Title:            "Go Module Documentation",
		},
		IncludeCSS: true,
		CustomCSS:  "",
	}
}

// Name returns the name of this visualizer
func (v *HTMLVisualizer) Name() string {
	return "HTML Visualizer"
}

// Description returns a description of what this visualizer produces
func (v *HTMLVisualizer) Description() string {
	return "Generates HTML documentation for Go modules"
}

// Visualize creates HTML documentation for a module
func (v *HTMLVisualizer) Visualize(mod *module.Module) ([]byte, error) {
	// Create an HTML visitor
	htmlVisitor := NewHTMLVisitor()

	// Apply options
	htmlVisitor.IncludePrivate = v.options.IncludePrivate
	htmlVisitor.IncludeTests = v.options.IncludeTests
	htmlVisitor.IncludeGenerated = v.options.IncludeGenerated
	htmlVisitor.Title = v.options.Title

	// Create a module walker with the HTML visitor
	walker := visitor.NewModuleWalker(htmlVisitor)

	// Configure the walker with the same options
	walker.IncludePrivate = v.options.IncludePrivate
	walker.IncludeTests = v.options.IncludeTests
	walker.IncludeGenerated = v.options.IncludeGenerated

	// Walk the module to generate HTML
	if err := walker.Walk(mod); err != nil {
		return nil, err
	}

	// Close HTML document
	htmlVisitor.dedent()
	htmlVisitor.writeString("</body>\n")
	htmlVisitor.writeString("</html>\n")

	return htmlVisitor.buffer.Bytes(), nil
}
