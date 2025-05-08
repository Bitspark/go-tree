// Package typesys provides the core type system for the Go-Tree analyzer.
// It wraps and extends golang.org/x/tools/go/types to provide a unified
// approach to code analysis with full type information.
package typesys

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

// Module represents a complete Go module with full type information.
// It serves as the root container for packages, files, and symbols.
type Module struct {
	// Basic information
	Path      string              // Module path from go.mod
	Dir       string              // Root directory of the module
	GoVersion string              // Go version used by the module
	Packages  map[string]*Package // Packages by import path

	// Type system internals
	FileSet   *token.FileSet               // FileSet for position information
	pkgCache  map[string]*packages.Package // Cache of loaded packages
	typeInfo  *types.Info                  // Type information
	typesMaps *typeutil.MethodSetCache     // Cache for method sets

	// Dependency tracking
	dependencies map[string][]string // Map from file to files it imports
	dependents   map[string][]string // Map from file to files that import it
}

// LoadOptions provides configuration for module loading.
type LoadOptions struct {
	IncludeTests   bool // Whether to include test files
	IncludePrivate bool // Whether to include private symbols
	Trace          bool // Enable verbose logging
}

// SaveOptions provides options for saving a module to disk.
type SaveOptions struct {
	FormatCode          bool // Whether to format the code
	IncludeTypeComments bool // Whether to include type information in comments
}

// VisualizeOptions provides options for visualizing a module.
type VisualizeOptions struct {
	IncludeTypeAnnotations bool
	IncludePrivate         bool
	IncludeTests           bool
	DetailLevel            int
	HighlightSymbol        *Symbol
}

// TransformResult contains the result of a transformation.
type TransformResult struct {
	ChangedFiles []string
	Errors       []error
}

// Transformation represents a code transformation.
type Transformation interface {
	// Apply applies the transformation to a module
	Apply(mod *Module) (*TransformResult, error)

	// Validate checks if the transformation would maintain type correctness
	Validate(mod *Module) error

	// Description provides information about the transformation
	Description() string
}

// NewModule creates a new empty module.
func NewModule(dir string) *Module {
	return &Module{
		Dir:          dir,
		Path:         "", // Start with empty path, will be set when go.mod is loaded
		Packages:     make(map[string]*Package),
		FileSet:      token.NewFileSet(),
		pkgCache:     make(map[string]*packages.Package),
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}
}

// PackageForFile returns the package that contains the given file.
func (m *Module) PackageForFile(filePath string) *Package {
	for _, pkg := range m.Packages {
		if _, ok := pkg.Files[filePath]; ok {
			return pkg
		}
	}
	return nil
}

// FileByPath returns a file by its path.
func (m *Module) FileByPath(path string) *File {
	if pkg := m.PackageForFile(path); pkg != nil {
		return pkg.Files[path]
	}
	return nil
}

// AllFiles returns all files in the module.
func (m *Module) AllFiles() []*File {
	files := make([]*File, 0)
	for _, pkg := range m.Packages {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}
	return files
}

// AddDependency records that one file depends on another.
func (m *Module) AddDependency(from, to string) {
	m.dependencies[from] = append(m.dependencies[from], to)
	m.dependents[to] = append(m.dependents[to], from)
}

// FindAffectedFiles identifies all files affected by changes to the given files.
func (m *Module) FindAffectedFiles(changedFiles []string) []string {
	affected := make(map[string]bool)
	for _, file := range changedFiles {
		affected[file] = true
		for _, dependent := range m.dependents[file] {
			affected[dependent] = true
		}
	}

	result := make([]string, 0, len(affected))
	for file := range affected {
		result = append(result, file)
	}
	return result
}

// UpdateChangedFiles updates only the changed files and their dependents.
func (m *Module) UpdateChangedFiles(files []string) error {
	// Group files by package
	filesByPackage := make(map[string][]string)
	for _, file := range files {
		if pkg := m.PackageForFile(file); pkg != nil {
			filesByPackage[pkg.ImportPath] = append(filesByPackage[pkg.ImportPath], file)
		}
	}

	// Process each package incrementally
	for pkgPath, pkgFiles := range filesByPackage {
		if err := m.Packages[pkgPath].UpdateFiles(pkgFiles); err != nil {
			return err
		}
	}

	// Update cross-package references
	return m.UpdateReferences(files)
}

// UpdateReferences updates references for the given files.
func (m *Module) UpdateReferences(files []string) error {
	// This is a placeholder that will be implemented later
	// The reference system depends on the Symbol and Reference types
	return nil
}

// FindAllReferences finds all references to a given symbol.
func (m *Module) FindAllReferences(sym *Symbol) ([]*Reference, error) {
	// This is a placeholder that will be implemented later
	// It depends on the Reference type that will be defined in reference.go
	finder := &TypeAwareReferencesFinder{Module: m}
	return finder.FindReferences(sym)
}

// FindImplementations finds all implementations of an interface.
func (m *Module) FindImplementations(iface *Symbol) ([]*Symbol, error) {
	// This is a placeholder that will be implemented later
	return nil, nil
}

// ApplyTransformation applies a code transformation.
func (m *Module) ApplyTransformation(t Transformation) (*TransformResult, error) {
	// Validate the transformation first
	if err := t.Validate(m); err != nil {
		return nil, fmt.Errorf("invalid transformation: %w", err)
	}

	// Apply the transformation
	return t.Apply(m)
}

// Save persists the module to disk with type verification.
func (m *Module) Save(dir string, opts *SaveOptions) error {
	// This is a placeholder that will be implemented later
	return nil
}

// Visualize creates a visualization of the module.
func (m *Module) Visualize(format string, opts *VisualizeOptions) ([]byte, error) {
	// This is a placeholder that will be implemented later
	return nil, nil
}

// CachePackage stores a loaded package in the module's internal cache.
// This is used by the loader package to maintain a record of loaded packages.
func (m *Module) CachePackage(path string, pkg *packages.Package) {
	m.pkgCache[path] = pkg
}

// GetCachedPackage retrieves a package from the module's internal cache.
func (m *Module) GetCachedPackage(path string) *packages.Package {
	return m.pkgCache[path]
}
