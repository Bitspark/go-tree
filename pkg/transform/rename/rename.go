// Package rename provides transformers for renaming elements in a Go module.
package rename

import (
	"fmt"

	"bitspark.dev/go-tree/pkg/core/module"
)

// VariableRenamer renames variables in a module
type VariableRenamer struct {
	OldName string // Original variable name
	NewName string // New variable name
}

// NewVariableRenamer creates a new variable renamer
func NewVariableRenamer(oldName, newName string) *VariableRenamer {
	return &VariableRenamer{
		OldName: oldName,
		NewName: newName,
	}
}

// Transform implements the ModuleTransformer interface
func (r *VariableRenamer) Transform(mod *module.Module) error {
	// Track if we found the variable to rename
	found := false

	// Iterate through all packages
	for _, pkg := range mod.Packages {
		// Check if the variable exists in this package
		variable, ok := pkg.Variables[r.OldName]
		if ok {
			// Store original position
			originalPos := variable.Pos
			originalEnd := variable.End

			// Create a new variable with the new name
			newVar := &module.Variable{
				Name:       r.NewName,
				File:       variable.File,
				Package:    variable.Package,
				Type:       variable.Type,
				Value:      variable.Value,
				IsExported: variable.IsExported,
				Doc:        variable.Doc,
				// Use the same position
				Pos: originalPos,
				End: originalEnd,
			}

			// Delete the old variable
			delete(pkg.Variables, r.OldName)

			// Add the new variable
			pkg.Variables[r.NewName] = newVar

			// Mark the package as modified
			pkg.IsModified = true

			// Mark the file as modified
			if newVar.File != nil {
				newVar.File.IsModified = true
			}

			found = true
		}
	}

	if !found {
		return fmt.Errorf("variable '%s' not found", r.OldName)
	}

	return nil
}

// Name returns the name of the transformer
func (r *VariableRenamer) Name() string {
	return "VariableRenamer"
}

// Description returns a description of what the transformer does
func (r *VariableRenamer) Description() string {
	return fmt.Sprintf("Renames variable '%s' to '%s'", r.OldName, r.NewName)
}
