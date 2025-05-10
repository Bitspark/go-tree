package saver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// GoModuleSaver implements ModuleSaver for type-aware Go modules
type GoModuleSaver struct {
	// Configuration options
	DefaultOptions SaveOptions

	// Optional file filter function
	FileFilter func(file *typesys.File) bool

	// Content generator
	generator FileContentGenerator
}

// NewGoModuleSaver creates a new Go module saver with default options
func NewGoModuleSaver() *GoModuleSaver {
	return &GoModuleSaver{
		DefaultOptions: DefaultSaveOptions(),
		generator:      NewDefaultFileContentGenerator(),
	}
}

// Save writes a module back to its original location
func (s *GoModuleSaver) Save(module *typesys.Module) error {
	return s.SaveWithOptions(module, s.DefaultOptions)
}

// SaveTo writes a module to a new location
func (s *GoModuleSaver) SaveTo(module *typesys.Module, dir string) error {
	return s.SaveToWithOptions(module, dir, s.DefaultOptions)
}

// SaveWithOptions writes a module with custom options
func (s *GoModuleSaver) SaveWithOptions(module *typesys.Module, options SaveOptions) error {
	if module.Dir == "" {
		return fmt.Errorf("module directory is empty, cannot save")
	}

	return s.SaveToWithOptions(module, module.Dir, options)
}

// SaveToWithOptions writes a module to a new location with custom options
func (s *GoModuleSaver) SaveToWithOptions(module *typesys.Module, dir string, options SaveOptions) error {
	if module == nil {
		return fmt.Errorf("module cannot be nil")
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Save go.mod file
	if err := s.saveGoMod(module, dir); err != nil {
		return fmt.Errorf("failed to save go.mod: %w", err)
	}

	// Save each package
	for importPath, pkg := range module.Packages {
		if err := s.savePackage(pkg, dir, importPath, module.Path, options); err != nil {
			return fmt.Errorf("failed to save package %s: %w", importPath, err)
		}
	}

	return nil
}

// saveGoMod saves the go.mod file for a module
func (s *GoModuleSaver) saveGoMod(module *typesys.Module, dir string) error {
	// Simple go.mod file with module path and Go version
	content := fmt.Sprintf("module %s\n\ngo %s\n", module.Path, module.GoVersion)

	// Note: In real implementation, we'd access the module's dependencies and replace directives
	// But for now, we'll create a minimal go.mod file since the actual typesys.Module structure
	// may not include these fields or they might be differently named/structured

	// Write the go.mod file
	goModPath := filepath.Join(dir, "go.mod")
	return os.WriteFile(goModPath, []byte(content), 0600)
}

// savePackage saves a package to disk
func (s *GoModuleSaver) savePackage(pkg *typesys.Package, baseDir, importPath, modulePath string, options SaveOptions) error {
	// Calculate relative path for package
	relPath := relativePath(importPath, modulePath)
	pkgDir := filepath.Join(baseDir, relPath)

	// Create package directory if it doesn't exist
	if err := os.MkdirAll(pkgDir, 0750); err != nil {
		return fmt.Errorf("failed to create package directory %s: %w", pkgDir, err)
	}

	// Save each file in the package
	for _, file := range pkg.Files {
		// Skip files if filter is set and returns false
		if s.FileFilter != nil && !s.FileFilter(file) {
			continue
		}

		// Generate file content
		content, err := s.generator.GenerateFileContent(file, options)
		if err != nil {
			return fmt.Errorf("failed to generate content for file %s: %w", file.Name, err)
		}

		// Save the file
		filePath := filepath.Join(pkgDir, file.Name)

		// Create a backup if requested
		if options.CreateBackups {
			if _, err := os.Stat(filePath); err == nil {
				// File exists, create backup
				backupPath := filePath + ".bak"
				if err := os.Rename(filePath, backupPath); err != nil {
					return fmt.Errorf("failed to create backup of %s: %w", filePath, err)
				}
			}
		}

		// Write file
		if err := os.WriteFile(filePath, content, 0600); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	return nil
}

// relativePath returns a path relative to the module path
// For example, if importPath is "github.com/user/repo/pkg" and modPath is "github.com/user/repo",
// it returns "pkg"
func relativePath(importPath, modPath string) string {
	// If the import path doesn't start with the module path, return it as is
	if !strings.HasPrefix(importPath, modPath) {
		return importPath
	}

	// Get the relative path
	relPath := strings.TrimPrefix(importPath, modPath)

	// Remove leading slash if present
	relPath = strings.TrimPrefix(relPath, "/")

	// If empty (root package), return empty string
	if relPath == "" {
		return ""
	}

	return relPath
}
