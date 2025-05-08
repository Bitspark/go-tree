// Package rename provides transformers for renaming elements in a Go module.
package rename

import (
	"fmt"

	"bitspark.dev/go-tree/pkgold/core/module"
	"bitspark.dev/go-tree/pkgold/transform"
)

// TypeRenamer renames types in a module
type TypeRenamer struct {
	PackagePath string // Package containing the type
	OldName     string // Original type name
	NewName     string // New type name
	DryRun      bool   // Whether to perform a dry run
}

// NewTypeRenamer creates a new type renamer
func NewTypeRenamer(packagePath, oldName, newName string, dryRun bool) *TypeRenamer {
	return &TypeRenamer{
		PackagePath: packagePath,
		OldName:     oldName,
		NewName:     newName,
		DryRun:      dryRun,
	}
}

// Transform implements the ModuleTransformer interface
func (r *TypeRenamer) Transform(mod *module.Module) *transform.TransformationResult {
	result := &transform.TransformationResult{
		Summary:       fmt.Sprintf("Rename type '%s' to '%s' in package '%s'", r.OldName, r.NewName, r.PackagePath),
		Success:       false,
		IsDryRun:      r.DryRun,
		AffectedFiles: []string{},
		Changes:       []transform.ChangePreview{},
	}

	// Find the target package
	var pkg *module.Package
	for _, p := range mod.Packages {
		if p.ImportPath == r.PackagePath {
			pkg = p
			break
		}
	}

	if pkg == nil {
		result.Error = fmt.Errorf("package '%s' not found", r.PackagePath)
		result.Details = "No package matched the given import path"
		return result
	}

	// Check if the type exists in this package
	typeObj, ok := pkg.Types[r.OldName]
	if !ok {
		result.Error = fmt.Errorf("type '%s' not found in package '%s'", r.OldName, r.PackagePath)
		result.Details = "No types matched the given name in the specified package"
		return result
	}

	// Track file information for result
	filePath := ""
	if typeObj.File != nil {
		filePath = typeObj.File.Path
		result.AffectedFiles = append(result.AffectedFiles, filePath)
	}

	// Add the change preview
	lineNum := 0 // In a real implementation, we would get the actual line number

	result.Changes = append(result.Changes, transform.ChangePreview{
		FilePath:   filePath,
		LineNumber: lineNum,
		Original:   r.OldName,
		New:        r.NewName,
	})

	// If this is just a dry run, don't actually make changes
	if !r.DryRun {
		// Store original position and properties
		originalPos := typeObj.Pos
		originalEnd := typeObj.End
		originalMethods := typeObj.Methods
		originalDoc := typeObj.Doc
		originalKind := typeObj.Kind
		originalIsExported := typeObj.IsExported
		originalFile := typeObj.File
		originalFields := typeObj.Fields

		// Create a new type with the new name
		newType := module.NewType(r.NewName, originalKind, originalIsExported)
		newType.Pos = originalPos
		newType.End = originalEnd
		newType.Doc = originalDoc
		newType.File = originalFile

		// Copy fields for struct types
		for name, field := range originalFields {
			newType.Fields[name] = field
		}

		// Copy methods
		for name, method := range originalMethods {
			// Create a copy of the method with updated parent reference
			newMethod := &module.Method{
				Name:       method.Name,
				Signature:  method.Signature,
				IsEmbedded: method.IsEmbedded,
				Doc:        method.Doc,
				Parent:     newType, // Update parent reference to the new type
				Pos:        method.Pos,
				End:        method.End,
			}
			newType.Methods[name] = newMethod
		}

		// Update functions that have this type as a receiver
		for _, fn := range pkg.Functions {
			if fn.IsMethod && fn.Receiver != nil && fn.Receiver.Type == r.OldName {
				fn.Receiver.Type = r.NewName
			} else if fn.IsMethod && fn.Receiver != nil && fn.Receiver.Type == "*"+r.OldName {
				fn.Receiver.Type = "*" + r.NewName
			}
		}

		// Delete the old type
		delete(pkg.Types, r.OldName)

		// Add the new type
		pkg.Types[r.NewName] = newType

		// Mark the package as modified
		pkg.IsModified = true

		// Mark the file as modified
		if newType.File != nil {
			newType.File.IsModified = true
		}
	}

	// Update the result
	result.Success = true
	result.FilesAffected = len(result.AffectedFiles)
	result.Details = fmt.Sprintf("Successfully renamed type '%s' to '%s' in package '%s'",
		r.OldName, r.NewName, r.PackagePath)

	return result
}

// Name returns the name of the transformer
func (r *TypeRenamer) Name() string {
	return "TypeRenamer"
}

// Description returns a description of what the transformer does
func (r *TypeRenamer) Description() string {
	return fmt.Sprintf("Renames type '%s' to '%s' in package '%s'", r.OldName, r.NewName, r.PackagePath)
}

// Rename is a convenience method that performs the rename operation directly on a specific type
func (r *TypeRenamer) Rename() error {
	if r.DryRun {
		return nil
	}

	// Note: In a real implementation, this would need to access the module
	// This is a placeholder
	return fmt.Errorf("direct rename not implemented - use Transform instead")
}
