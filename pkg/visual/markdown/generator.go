// Package markdown provides functionality for generating Markdown documentation
// from Go-Tree package model data.
package markdown

import (
	"encoding/json"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/model"
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

// GenerateFromJSON converts JSON package data to a Markdown document
func (g *Generator) GenerateFromJSON(jsonData []byte) (string, error) {
	var pkg model.GoPackage
	if err := json.Unmarshal(jsonData, &pkg); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return g.Generate(&pkg)
}

// Generate converts a GoPackage model to a Markdown document
func (g *Generator) Generate(pkg *model.GoPackage) (string, error) {
	visitor := NewMarkdownVisitor(g.options)
	baseFormatter := formatter.NewBaseFormatter(visitor)
	return baseFormatter.Format(pkg)
}
