// Package markdown provides functionality for generating Markdown documentation
// from Go-Tree module data.
package markdown

import (
	"encoding/json"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/module"
	"bitspark.dev/go-tree/pkg/visual/formatter"
)

// Options configures Markdown generation
type Options struct {
	// IncludeCodeBlocks determines whether to include Go code blocks in the output
	IncludeCodeBlocks bool

	// IncludeLinks determines whether to include internal links in the document
	IncludeLinks bool

	// IncludeTOC determines whether to include a table of contents
	IncludeTOC bool
}

// DefaultOptions returns default Markdown options
func DefaultOptions() Options {
	return Options{
		IncludeCodeBlocks: true,
		IncludeLinks:      true,
		IncludeTOC:        true,
	}
}

// Generator handles Markdown generation
type Generator struct {
	options Options
}

// NewGenerator creates a new Markdown generator
func NewGenerator(options Options) *Generator {
	return &Generator{
		options: options,
	}
}

// GenerateFromJSON converts JSON module data to a Markdown document
func (g *Generator) GenerateFromJSON(jsonData []byte) (string, error) {
	var mod module.Module
	if err := json.Unmarshal(jsonData, &mod); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return g.Generate(&mod)
}

// Generate converts a Module to a Markdown document
func (g *Generator) Generate(mod *module.Module) (string, error) {
	visitor := NewMarkdownVisitor(g.options)
	baseFormatter := formatter.NewBaseFormatter(visitor)
	return baseFormatter.Format(mod)
}
