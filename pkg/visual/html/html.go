// Package html provides functionality for generating HTML documentation
// from Go-Tree JSON output.
package html

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"bitspark.dev/go-tree/pkg/core/model"
)

// Generator handles HTML generation from package data
type Generator struct {
	// Options for HTML generation
	options Options
}

// Options configures the behavior of the HTML generator
type Options struct {
	// Title is the HTML document title
	Title string

	// SyntaxHighlighting determines whether to include syntax highlighting
	SyntaxHighlighting bool

	// IncludeCSS determines whether to embed CSS in the HTML or link to external files
	IncludeCSS bool

	// CustomCSS allows for custom CSS to be included
	CustomCSS string
}

// DefaultOptions returns the default HTML generator options
func DefaultOptions() Options {
	return Options{
		Title:              "Go Package Documentation",
		SyntaxHighlighting: true,
		IncludeCSS:         true,
		CustomCSS:          "",
	}
}

// NewGenerator creates a new HTML generator with the given options
func NewGenerator(options Options) *Generator {
	return &Generator{
		options: options,
	}
}

// GenerateFromJSON converts JSON package data to an HTML document
func (g *Generator) GenerateFromJSON(jsonData []byte) (string, error) {
	var pkg model.GoPackage
	if err := json.Unmarshal(jsonData, &pkg); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return g.Generate(&pkg)
}

// Generate converts a GoPackage model to an HTML document
func (g *Generator) Generate(pkg *model.GoPackage) (string, error) {
	// Prepare template data
	data := struct {
		Package            *model.GoPackage
		Title              string
		IncludeCSS         bool
		CustomCSS          template.CSS
		EnableHighlighting bool
	}{
		Package:            pkg,
		Title:              g.options.Title,
		IncludeCSS:         g.options.IncludeCSS,
		CustomCSS:          template.CSS(g.options.CustomCSS),
		EnableHighlighting: g.options.SyntaxHighlighting,
	}

	// Initialize template
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"formatCode":    formatCode,
		"typeKindClass": typeKindClass,
		"formatDoc":     formatDocComment,
	}).Parse(htmlTemplate)

	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// formatCode formats Go code with syntax highlighting classes
func formatCode(code string) template.HTML {
	// Simple formatting for now - in a real implementation,
	// this would use a syntax highlighter library

	// Escape HTML characters
	code = template.HTMLEscapeString(code)

	// Add basic syntax highlighting classes
	code = strings.ReplaceAll(code, "func ", "<span class=\"keyword\">func</span> ")
	code = strings.ReplaceAll(code, "type ", "<span class=\"keyword\">type</span> ")
	code = strings.ReplaceAll(code, "struct ", "<span class=\"keyword\">struct</span> ")
	code = strings.ReplaceAll(code, "interface ", "<span class=\"keyword\">interface</span> ")
	code = strings.ReplaceAll(code, "package ", "<span class=\"keyword\">package</span> ")
	code = strings.ReplaceAll(code, "import ", "<span class=\"keyword\">import</span> ")
	code = strings.ReplaceAll(code, "const ", "<span class=\"keyword\">const</span> ")
	code = strings.ReplaceAll(code, "var ", "<span class=\"keyword\">var</span> ")
	code = strings.ReplaceAll(code, "return ", "<span class=\"keyword\">return</span> ")

	// Add string literals highlighting
	parts := strings.Split(code, "\"")
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			parts[i] = "<span class=\"string\">\"" + parts[i] + "\"</span>"
		}
	}
	code = strings.Join(parts, "")

	// Add line numbers and wrap in code element
	lines := strings.Split(code, "\n")
	result := "<pre class=\"code\">"
	for i, line := range lines {
		result += fmt.Sprintf("<span class=\"line-number\">%d</span>%s\n", i+1, line)
	}
	result += "</pre>"

	return template.HTML(result)
}

// typeKindClass returns a CSS class based on the type kind
func typeKindClass(kind string) string {
	switch kind {
	case "struct":
		return "type-struct"
	case "interface":
		return "type-interface"
	case "alias":
		return "type-alias"
	default:
		return "type-other"
	}
}

// formatDocComment formats a documentation comment for HTML display
func formatDocComment(doc string) template.HTML {
	if doc == "" {
		return ""
	}

	// Replace newlines with <br> for HTML display
	doc = strings.ReplaceAll(doc, "\n", "<br>")

	return template.HTML("<div class=\"doc-comment\">" + doc + "</div>")
}
