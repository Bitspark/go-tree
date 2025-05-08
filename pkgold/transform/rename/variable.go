// Package rename provides transformers for renaming elements in a Go module.
package rename

import (
	"fmt"

	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/transform"
)

// VariableRenamer renames variables in a module
type VariableRenamer struct {
	OldName string // Original variable name
	NewName string // New variable name
	DryRun  bool   // Whether to perform a dry run
}

// NewVariableRenamer creates a new variable renamer
func NewVariableRenamer(oldName, newName string, dryRun bool) *VariableRenamer {
	return &VariableRenamer{
		OldName: oldName,
		NewName: newName,
		DryRun:  dryRun,
	}
}

// Transform implements the ModuleTransformer interface
func (r *VariableRenamer) Transform(mod *module.Module) *transform.TransformationResult {
	result := &transform.TransformationResult{
		Summary:       fmt.Sprintf("Rename variable '%s' to '%s'", r.OldName, r.NewName),
		Success:       false,
		IsDryRun:      r.DryRun,
		AffectedFiles: []string{},
		Changes:       []transform.ChangePreview{},
	}

	// Track if we found the variable to rename
	found := false

	// Iterate through all packages
	for _, pkg := range mod.Packages {
		// Check if the variable exists in this package
		variable, ok := pkg.Variables[r.OldName]
		if ok {
			found = true

			// Track file information for result
			filePath := ""
			if variable.File != nil {
				filePath = variable.File.Path

				// Add to affected files if not already there
				fileAlreadyAdded := false
				for _, f := range result.AffectedFiles {
					if f == filePath {
						fileAlreadyAdded = true
						break
					}
				}

				if !fileAlreadyAdded {
					result.AffectedFiles = append(result.AffectedFiles, filePath)
				}
			}

			// Add the change preview
			lineNum := 0

			result.Changes = append(result.Changes, transform.ChangePreview{
				FilePath:   filePath,
				LineNumber: lineNum,
				Original:   r.OldName,
				New:        r.NewName,
			})

			// If this is just a dry run, don't actually make changes
			if r.DryRun {
				continue
			}

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
		}
	}

	if !found {
		result.Error = fmt.Errorf("variable '%s' not found", r.OldName)
		result.Details = "No variables matched the given name"
		return result
	}

	// Update the result
	result.Success = true
	result.FilesAffected = len(result.AffectedFiles)
	result.Details = fmt.Sprintf("Successfully renamed '%s' to '%s' in %d file(s)",
		r.OldName, r.NewName, result.FilesAffected)

	return result
}

// Name returns the name of the transformer
func (r *VariableRenamer) Name() string {
	return "VariableRenamer"
}

// Description returns a description of what the transformer does
func (r *VariableRenamer) Description() string {
	return fmt.Sprintf("Renames variable '%s' to '%s'", r.OldName, r.NewName)
}
