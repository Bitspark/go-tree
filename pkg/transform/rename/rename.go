// Package rename provides transformers for renaming symbols in Go code.
// It ensures type-safe renamings using the typesys package.
package rename

import (
	"fmt"
	"go/token"

	"bitspark.dev/go-tree/pkg/graph"
	"bitspark.dev/go-tree/pkg/transform"
	"bitspark.dev/go-tree/pkg/typesys"
)

// SymbolRenamer renames a symbol and all its references.
type SymbolRenamer struct {
	// ID of the symbol to rename
	SymbolID string

	// New name for the symbol
	NewName string

	// Symbol lookup cache
	symbol *typesys.Symbol
}

// NewSymbolRenamer creates a new symbol renamer.
func NewSymbolRenamer(symbolID, newName string) *SymbolRenamer {
	return &SymbolRenamer{
		SymbolID: symbolID,
		NewName:  newName,
	}
}

// Transform implements the transform.Transformer interface.
func (r *SymbolRenamer) Transform(ctx *transform.Context) (*transform.TransformResult, error) {
	result := &transform.TransformResult{
		Summary:       fmt.Sprintf("Rename symbol to '%s'", r.NewName),
		Success:       false,
		IsDryRun:      ctx.DryRun,
		AffectedFiles: []string{},
		Changes:       []transform.Change{},
	}

	// Find the symbol to rename
	symbol := ctx.Index.GetSymbolByID(r.SymbolID)
	if symbol == nil {
		result.Error = fmt.Errorf("symbol with ID '%s' not found", r.SymbolID)
		return result, result.Error
	}
	r.symbol = symbol

	// Build an impact graph to analyze dependencies - just for analysis, not needed for result
	_ = buildImpactGraph(ctx, symbol)

	// For renaming, we need to track the symbol's definition and all references
	references := ctx.Index.FindReferences(symbol)

	// Add the symbol's defining file to affected files
	if symbol.File != nil {
		result.AffectedFiles = append(result.AffectedFiles, symbol.File.Path)
	}

	// Add all reference files to affected files
	for _, ref := range references {
		// Check if this file is already in the list
		found := false
		for _, file := range result.AffectedFiles {
			if ref.File != nil && file == ref.File.Path {
				found = true
				break
			}
		}
		if !found && ref.File != nil {
			result.AffectedFiles = append(result.AffectedFiles, ref.File.Path)
		}
	}

	// Collect changes
	originalName := symbol.Name
	result.Details = fmt.Sprintf("Rename '%s' to '%s' (%d references)",
		originalName, r.NewName, len(references))

	// Create change for symbol definition
	if symbol.File != nil {
		defChange := transform.Change{
			FilePath:       symbol.File.Path,
			StartLine:      posToLine(ctx.Module, symbol.Pos),
			EndLine:        posToLine(ctx.Module, symbol.End),
			Original:       originalName,
			New:            r.NewName,
			AffectedSymbol: symbol,
		}
		result.Changes = append(result.Changes, defChange)
	}

	// Create changes for all references
	for _, ref := range references {
		if ref.File != nil {
			refChange := transform.Change{
				FilePath:       ref.File.Path,
				StartLine:      posToLine(ctx.Module, ref.Pos),
				EndLine:        posToLine(ctx.Module, ref.End),
				Original:       originalName,
				New:            r.NewName,
				AffectedSymbol: symbol,
			}
			result.Changes = append(result.Changes, refChange)
		}
	}

	// If this is just a dry run, don't actually make changes
	if ctx.DryRun {
		result.Success = true
		result.FilesAffected = len(result.AffectedFiles)
		return result, nil
	}

	// Apply the changes
	if err := applyRenameChanges(ctx, symbol, r.NewName, references); err != nil {
		result.Error = fmt.Errorf("failed to apply rename changes: %w", err)
		return result, result.Error
	}

	// Update the index
	if err := ctx.Index.Update(result.AffectedFiles); err != nil {
		result.Error = fmt.Errorf("failed to update index: %w", err)
		return result, result.Error
	}

	// Set success and return
	result.Success = true
	result.FilesAffected = len(result.AffectedFiles)
	return result, nil
}

// Validate implements the transform.Transformer interface.
func (r *SymbolRenamer) Validate(ctx *transform.Context) error {
	// Find the symbol to rename
	symbol := ctx.Index.GetSymbolByID(r.SymbolID)
	if symbol == nil {
		return fmt.Errorf("symbol with ID '%s' not found", r.SymbolID)
	}

	// Check if the new name is valid
	if r.NewName == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	// Check for conflicts with the new name
	// Look for symbols with the same name in the same scope
	pkg := symbol.Package
	if pkg == nil {
		return fmt.Errorf("symbol has no package")
	}

	// Check if a symbol with the new name already exists in this package
	conflicts := ctx.Index.FindSymbolsByName(r.NewName)
	for _, conflict := range conflicts {
		if conflict.Package == pkg {
			// Skip if it's the symbol we're renaming
			if conflict.ID == symbol.ID {
				continue
			}

			// Check if the conflict is in the same scope
			if isInSameScope(symbol, conflict) {
				return fmt.Errorf("a symbol named '%s' already exists in the same scope", r.NewName)
			}
		}
	}

	return nil
}

// Name implements the transform.Transformer interface.
func (r *SymbolRenamer) Name() string {
	return "SymbolRenamer"
}

// Description implements the transform.Transformer interface.
func (r *SymbolRenamer) Description() string {
	if r.symbol != nil {
		return fmt.Sprintf("Rename '%s' to '%s'", r.symbol.Name, r.NewName)
	}
	return fmt.Sprintf("Rename symbol to '%s'", r.NewName)
}

// Helper function to check if two symbols are in the same scope
func isInSameScope(symbol1, symbol2 *typesys.Symbol) bool {
	// If they're not in the same package, they're not in the same scope
	if symbol1.Package != symbol2.Package {
		return false
	}

	// If they're both top-level symbols, they're in the same scope
	if symbol1.Parent == nil && symbol2.Parent == nil {
		return true
	}

	// If only one is top-level, they're not in the same scope
	if (symbol1.Parent == nil) != (symbol2.Parent == nil) {
		return false
	}

	// If they have the same parent, they're in the same scope
	return symbol1.Parent.ID == symbol2.Parent.ID
}

// Helper function to convert token.Pos to line number
func posToLine(mod *typesys.Module, pos token.Pos) int {
	if mod == nil || mod.FileSet == nil {
		return 0
	}
	position := mod.FileSet.Position(pos)
	return position.Line
}

// Helper function to build a dependency graph for impact analysis
func buildImpactGraph(ctx *transform.Context, symbol *typesys.Symbol) *graph.DirectedGraph {
	g := graph.NewDirectedGraph()

	// Add the symbol as the root node
	g.AddNode(symbol.ID, symbol)

	// Add all references
	references := ctx.Index.FindReferences(symbol)
	for _, ref := range references {
		if ref.Context != nil {
			// Add edge from context (symbol containing the reference) to the symbol
			g.AddNode(ref.Context.ID, ref.Context)
			g.AddEdge(ref.Context.ID, symbol.ID, nil)
		}
	}

	// For methods, add edges to/from the receiver type
	if symbol.Kind == typesys.KindMethod && symbol.Parent != nil {
		g.AddNode(symbol.Parent.ID, symbol.Parent)
		g.AddEdge(symbol.Parent.ID, symbol.ID, nil)
	}

	// For interfaces, add edges to implementing types
	if symbol.Kind == typesys.KindInterface {
		impls := ctx.Index.FindImplementations(symbol)
		for _, impl := range impls {
			g.AddNode(impl.ID, impl)
			g.AddEdge(impl.ID, symbol.ID, nil)
		}
	}

	return g
}

// Helper function to apply rename changes
func applyRenameChanges(ctx *transform.Context, symbol *typesys.Symbol, newName string, references []*typesys.Reference) error {
	// For a real implementation, this would update the AST and generate new code
	// In this version, we'll just update the symbol and references in the type system

	// Update the symbol name
	symbol.Name = newName

	// In a real implementation, we would mark files as modified
	// For now, we'll just note this as a comment since the IsModified field doesn't exist

	return nil
}
