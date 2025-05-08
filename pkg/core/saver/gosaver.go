// Package saver provides implementations for saving Go modules.
package saver

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"

	"bitspark.dev/go-tree/pkg/core/module"
)

// GoModuleSaver implements ModuleSaver for Go modules
type GoModuleSaver struct {
	// Embedded fields
}

// NewGoModuleSaver creates a new module saver for Go modules
func NewGoModuleSaver() *GoModuleSaver {
	return &GoModuleSaver{}
}

// Save writes a module back to its original location
func (s *GoModuleSaver) Save(module *module.Module) error {
	return s.SaveWithOptions(module, DefaultSaveOptions())
}

// SaveTo writes a module to a new location
func (s *GoModuleSaver) SaveTo(module *module.Module, dir string) error {
	return s.SaveToWithOptions(module, dir, DefaultSaveOptions())
}

// SaveWithOptions writes a module with custom options
func (s *GoModuleSaver) SaveWithOptions(module *module.Module, options SaveOptions) error {
	return s.SaveToWithOptions(module, module.Dir, options)
}

// SaveToWithOptions writes a module to a new location with custom options
func (s *GoModuleSaver) SaveToWithOptions(module *module.Module, dir string, options SaveOptions) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Save go.mod file
	if err := s.saveGoMod(module, dir); err != nil {
		return fmt.Errorf("failed to save go.mod: %w", err)
	}

	// Save packages
	for _, pkg := range module.Packages {
		if err := s.savePackage(pkg, dir, options); err != nil {
			return fmt.Errorf("failed to save package %s: %w", pkg.ImportPath, err)
		}
	}

	return nil
}

// saveGoMod saves the go.mod file
func (s *GoModuleSaver) saveGoMod(module *module.Module, dir string) error {
	// In a real implementation, would generate proper go.mod content
	// This is a simplified example
	goModPath := filepath.Join(dir, "go.mod")

	content := fmt.Sprintf("module %s\n\ngo %s\n", module.Path, module.GoVersion)

	// Add dependencies
	if len(module.Dependencies) > 0 {
		content += "\nrequire (\n"
		for _, dep := range module.Dependencies {
			indirect := ""
			if dep.Indirect {
				indirect = " // indirect"
			}
			content += fmt.Sprintf("\t%s %s%s\n", dep.Path, dep.Version, indirect)
		}
		content += ")\n"
	}

	// Add replacements
	if len(module.Replace) > 0 {
		content += "\nreplace (\n"
		for _, rep := range module.Replace {
			content += fmt.Sprintf("\t%s => %s %s\n",
				rep.Old.Path, rep.New.Path, rep.New.Version)
		}
		content += ")\n"
	}

	return os.WriteFile(goModPath, []byte(content), 0600)
}

// savePackage saves a package to disk
func (s *GoModuleSaver) savePackage(pkg *module.Package, baseDir string, options SaveOptions) error {
	// Calculate package directory relative to module root
	relDir := strings.TrimPrefix(pkg.ImportPath, pkg.Module.Path)
	relDir = strings.TrimPrefix(relDir, "/")

	// Create full package directory path
	pkgDir := filepath.Join(baseDir, relDir)
	if relDir == "" {
		pkgDir = baseDir // Root package
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(pkgDir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", pkgDir, err)
	}

	// Save each file in the package
	for _, file := range pkg.Files {
		if err := s.saveFile(file, pkgDir, options); err != nil {
			return fmt.Errorf("failed to save file %s: %w", file.Name, err)
		}
	}

	return nil
}

// saveFile saves a single file to disk
func (s *GoModuleSaver) saveFile(file *module.File, dir string, options SaveOptions) error {
	// Generate the Go source code for the file
	source, err := s.generateFileSource(file, options)
	if err != nil {
		return fmt.Errorf("failed to generate source code: %w", err)
	}

	// Format the source code if requested
	if options.Format {
		if options.OrganizeImports {
			// Use goimports to format and organize imports
			formatted, err := imports.Process(file.Name, source, nil)
			if err != nil {
				return fmt.Errorf("failed to format source code with imports: %w", err)
			}
			source = formatted
		} else {
			// Use standard go formatter
			formatted, err := format.Source(source)
			if err != nil {
				return fmt.Errorf("failed to format source code: %w", err)
			}
			source = formatted
		}
	}

	// Create the file path
	filePath := filepath.Join(dir, file.Name)

	// Check if the file exists and we need to create a backup
	if options.CreateBackups {
		if _, err := os.Stat(filePath); err == nil {
			backupPath := filePath + ".bak"
			if err := os.Rename(filePath, backupPath); err != nil {
				return fmt.Errorf("failed to create backup of %s: %w", filePath, err)
			}
		}
	}

	// Write the file
	return os.WriteFile(filePath, source, 0600)
}

// generateFileSource generates the Go source code for a file
func (s *GoModuleSaver) generateFileSource(file *module.File, options SaveOptions) ([]byte, error) {
	// In a real implementation, this would be much more sophisticated
	// For this example, we're just doing a basic reconstruction

	// If we have the original source and AST, we would use that for reconstruction
	if file.SourceCode != "" && !hasModifications(file) {
		return []byte(file.SourceCode), nil
	}

	// Otherwise, generate from scratch (simplified)
	var builder strings.Builder

	// Package declaration
	builder.WriteString(fmt.Sprintf("package %s\n\n", file.Package.Name))

	// Imports
	if len(file.Imports) > 0 {
		builder.WriteString("import (\n")
		for _, imp := range file.Imports {
			if imp.IsBlank {
				builder.WriteString(fmt.Sprintf("\t_ \"%s\"\n", imp.Path))
			} else if imp.Name != "" {
				builder.WriteString(fmt.Sprintf("\t%s \"%s\"\n", imp.Name, imp.Path))
			} else {
				builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp.Path))
			}
		}
		builder.WriteString(")\n\n")
	}

	// Constants
	for _, c := range file.Constants {
		if c.Doc != "" {
			builder.WriteString(fmt.Sprintf("// %s\n", c.Doc))
		}

		if c.Type != "" {
			builder.WriteString(fmt.Sprintf("const %s %s = %s\n\n", c.Name, c.Type, c.Value))
		} else {
			builder.WriteString(fmt.Sprintf("const %s = %s\n\n", c.Name, c.Value))
		}
	}

	// Variables
	for _, v := range file.Variables {
		if v.Doc != "" {
			builder.WriteString(fmt.Sprintf("// %s\n", v.Doc))
		}

		if v.Type != "" && v.Value != "" {
			builder.WriteString(fmt.Sprintf("var %s %s = %s\n\n", v.Name, v.Type, v.Value))
		} else if v.Type != "" {
			builder.WriteString(fmt.Sprintf("var %s %s\n\n", v.Name, v.Type))
		} else {
			builder.WriteString(fmt.Sprintf("var %s = %s\n\n", v.Name, v.Value))
		}
	}

	// Types
	for _, t := range file.Types {
		if t.Doc != "" {
			builder.WriteString(fmt.Sprintf("// %s\n", t.Doc))
		}

		switch t.Kind {
		case "struct":
			builder.WriteString(fmt.Sprintf("type %s struct {\n", t.Name))
			for _, f := range t.Fields {
				if f.IsEmbedded {
					if f.Tag != "" {
						builder.WriteString(fmt.Sprintf("\t%s %s\n", f.Type, f.Tag))
					} else {
						builder.WriteString(fmt.Sprintf("\t%s\n", f.Type))
					}
				} else {
					if f.Tag != "" {
						builder.WriteString(fmt.Sprintf("\t%s %s %s\n", f.Name, f.Type, f.Tag))
					} else {
						builder.WriteString(fmt.Sprintf("\t%s %s\n", f.Name, f.Type))
					}
				}
			}
			builder.WriteString("}\n\n")

		case "interface":
			builder.WriteString(fmt.Sprintf("type %s interface {\n", t.Name))
			for _, m := range t.Interfaces {
				if m.IsEmbedded {
					builder.WriteString(fmt.Sprintf("\t%s\n", m.Name))
				} else {
					builder.WriteString(fmt.Sprintf("\t%s%s\n", m.Name, m.Signature))
				}
			}
			builder.WriteString("}\n\n")

		case "alias":
			builder.WriteString(fmt.Sprintf("type %s = %s\n\n", t.Name, t.Underlying))

		default:
			builder.WriteString(fmt.Sprintf("type %s %s\n\n", t.Name, t.Underlying))
		}
	}

	// Functions and methods
	for _, fn := range file.Functions {
		if fn.Doc != "" {
			builder.WriteString(fmt.Sprintf("// %s\n", fn.Doc))
		}

		if fn.IsMethod {
			builder.WriteString(fmt.Sprintf("func (%s) %s%s {\n",
				formatReceiver(fn.Receiver), fn.Name, fn.Signature))
		} else {
			builder.WriteString(fmt.Sprintf("func %s%s {\n", fn.Name, fn.Signature))
		}

		if fn.Body != "" {
			builder.WriteString(fn.Body)
		} else {
			builder.WriteString("\t// Implementation\n")
		}

		builder.WriteString("}\n\n")
	}

	return []byte(builder.String()), nil
}

// formatReceiver formats a method receiver
func formatReceiver(r *module.Receiver) string {
	if r == nil {
		return ""
	}

	if r.Name == "" {
		if r.IsPointer {
			return fmt.Sprintf("*%s", r.Type)
		}
		return r.Type
	}

	if r.IsPointer {
		return fmt.Sprintf("%s *%s", r.Name, r.Type)
	}
	return fmt.Sprintf("%s %s", r.Name, r.Type)
}

// hasModifications checks if a file has been modified since loading
// This is a placeholder - a real implementation would track modifications
func hasModifications(file *module.File) bool {
	// For this example, we always generate new code
	return true
}
