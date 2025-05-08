// Package module defines the core data model for representing Go modules.
package module

import (
	"path/filepath"
)

// Module represents a complete Go module
type Module struct {
	// Core module identity
	Path      string // Module path (e.g., "github.com/user/repo")
	Version   string // Semantic version if applicable
	GoVersion string // Go version requirement

	// Content
	Packages    map[string]*Package // Map of package import paths to packages
	MainPackage *Package            // Main package if this is an executable module

	// Module relationships
	Dependencies []*ModuleDependency // Other modules this module depends on
	Replace      []*ModuleReplace    // Module replacements

	// Build information
	BuildFlags map[string]string // Build flags
	BuildTags  []string          // Build constraints

	// Module metadata
	Dir   string // Root directory path
	GoMod string // Path to go.mod file
}

// ModuleDependency represents a dependency on another module
type ModuleDependency struct {
	Path     string // Module path
	Version  string // Required version
	Indirect bool   // Whether it's an indirect dependency
}

// ModuleReplace represents a module replacement directive
type ModuleReplace struct {
	Old *ModuleDependency // Module to replace
	New *ModuleDependency // Replacement module
}

// NewModule creates a new empty module with the given path
func NewModule(path, dir string) *Module {
	return &Module{
		Path:         path,
		Dir:          dir,
		GoMod:        filepath.Join(dir, "go.mod"),
		Packages:     make(map[string]*Package),
		BuildFlags:   make(map[string]string),
		BuildTags:    make([]string, 0),
		Dependencies: make([]*ModuleDependency, 0),
		Replace:      make([]*ModuleReplace, 0),
	}
}

// AddPackage adds a package to the module
func (m *Module) AddPackage(pkg *Package) {
	m.Packages[pkg.ImportPath] = pkg
	pkg.Module = m
}

// FindType finds a type by its fully qualified name (package/type)
func (m *Module) FindType(fullName string) *Type {
	for _, pkg := range m.Packages {
		for _, typ := range pkg.Types {
			if pkg.ImportPath+"."+typ.Name == fullName {
				return typ
			}
		}
	}
	return nil
}

// FindFunction finds a function by its fully qualified name (package/function)
func (m *Module) FindFunction(fullName string) *Function {
	for _, pkg := range m.Packages {
		for _, fn := range pkg.Functions {
			if pkg.ImportPath+"."+fn.Name == fullName {
				return fn
			}
		}
	}
	return nil
}

// AddDependency adds a module dependency
func (m *Module) AddDependency(path, version string, indirect bool) {
	m.Dependencies = append(m.Dependencies, &ModuleDependency{
		Path:     path,
		Version:  version,
		Indirect: indirect,
	})
}

// AddReplace adds a module replacement
func (m *Module) AddReplace(oldPath, oldVersion, newPath, newVersion string) {
	m.Replace = append(m.Replace, &ModuleReplace{
		Old: &ModuleDependency{
			Path:    oldPath,
			Version: oldVersion,
		},
		New: &ModuleDependency{
			Path:    newPath,
			Version: newVersion,
		},
	})
}
