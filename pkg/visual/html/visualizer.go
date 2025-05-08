package html

import (
	"bytes"
	"html/template"

	"bitspark.dev/go-tree/pkg/typesys"
	"bitspark.dev/go-tree/pkg/visual/formatter"
)

// VisualizationOptions provides options for HTML visualization
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

// HTMLVisualizer creates HTML visualizations of a Go module with full type information
type HTMLVisualizer struct {
	template *template.Template
}

// NewHTMLVisualizer creates a new HTML visualizer
func NewHTMLVisualizer() *HTMLVisualizer {
	tmpl, err := template.New("html").Parse(BaseTemplate)
	if err != nil {
		// This should never happen since the template is hard-coded
		panic("failed to parse HTML template: " + err.Error())
	}

	return &HTMLVisualizer{
		template: tmpl,
	}
}

// Visualize creates an HTML visualization of the module
func (v *HTMLVisualizer) Visualize(module *typesys.Module, opts *VisualizationOptions) ([]byte, error) {
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
	visitor := NewHTMLVisitor(formatOpts)

	// Walk the module with the visitor
	if err := typesys.Walk(visitor, module); err != nil {
		return nil, err
	}

	// Get the content from the visitor
	content, err := visitor.Result()
	if err != nil {
		return nil, err
	}

	// Create the template data
	data := map[string]interface{}{
		"Title":        getTitle(opts, module),
		"ModulePath":   module.Path,
		"GoVersion":    module.GoVersion,
		"PackageCount": len(module.Packages),
		"Content":      template.HTML(content),
	}

	// Execute the template
	var buf bytes.Buffer
	if err := v.template.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Format returns the output format name
func (v *HTMLVisualizer) Format() string {
	return "html"
}

// SupportsTypeAnnotations indicates if this visualizer can show type info
func (v *HTMLVisualizer) SupportsTypeAnnotations() bool {
	return true
}

// Helper function to get a title for the visualization
func getTitle(opts *VisualizationOptions, module *typesys.Module) string {
	if opts != nil && opts.Title != "" {
		return opts.Title
	}

	return "Go Module: " + module.Path
}
