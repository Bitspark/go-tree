package materialize

import (
	"os"
	"path/filepath"
	"strings"
)

// DependencyPolicy determines which dependencies get materialized
type DependencyPolicy int

const (
	// AllDependencies materializes all dependencies recursively
	AllDependencies DependencyPolicy = iota

	// DirectDependenciesOnly materializes only direct dependencies
	DirectDependenciesOnly

	// NoDependencies only materializes the specified modules
	NoDependencies
)

// ReplaceStrategy determines how replace directives are generated
type ReplaceStrategy int

const (
	// RelativeReplace uses relative paths for local replacements
	RelativeReplace ReplaceStrategy = iota

	// AbsoluteReplace uses absolute paths for local replacements
	AbsoluteReplace

	// NoReplace doesn't add replace directives
	NoReplace
)

// LayoutStrategy determines how modules are laid out on disk
type LayoutStrategy int

const (
	// FlatLayout puts all modules in separate directories under the root
	FlatLayout LayoutStrategy = iota

	// HierarchicalLayout maintains module hierarchy in directories
	HierarchicalLayout

	// GoPathLayout mimics traditional GOPATH structure
	GoPathLayout
)

// MaterializeOptions configures materialization behavior
type MaterializeOptions struct {
	// Target directory for materialization, if empty a temporary directory is used
	TargetDir string

	// Policy for which dependencies to include
	DependencyPolicy DependencyPolicy

	// Strategy for generating replace directives
	ReplaceStrategy ReplaceStrategy

	// Strategy for module layout on disk
	LayoutStrategy LayoutStrategy

	// Whether to run go mod tidy after materialization
	RunGoModTidy bool

	// Whether to include test files
	IncludeTests bool

	// Environment variables to set during execution
	EnvironmentVars map[string]string

	// Enable verbose logging
	Verbose bool

	// Whether to preserve the environment after cleanup
	Preserve bool
}

// DefaultMaterializeOptions returns a MaterializeOptions with default values
func DefaultMaterializeOptions() MaterializeOptions {
	return MaterializeOptions{
		DependencyPolicy: DirectDependenciesOnly,
		ReplaceStrategy:  RelativeReplace,
		LayoutStrategy:   FlatLayout,
		RunGoModTidy:     true,
		IncludeTests:     false,
		EnvironmentVars:  make(map[string]string),
		Verbose:          false,
		Preserve:         false,
	}
}

// NewTemporaryMaterializeOptions creates options for a temporary environment
func NewTemporaryMaterializeOptions() MaterializeOptions {
	opts := DefaultMaterializeOptions()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "go-tree-materialized-*")
	if err == nil {
		opts.TargetDir = tmpDir
	}

	return opts
}

// IsTemporary returns true if the options specify a temporary environment
func (o MaterializeOptions) IsTemporary() bool {
	// If TargetDir is empty, we'll create a temporary directory
	if o.TargetDir == "" {
		return true
	}

	// If TargetDir is in the system temp directory, it's probably temporary
	tempDir := os.TempDir()
	// Use a safer path comparison than HasPrefix
	targetAbs, err := filepath.Abs(o.TargetDir)
	if err != nil {
		return false
	}
	tempAbs, err := filepath.Abs(tempDir)
	if err != nil {
		return false
	}

	targetAbs = filepath.Clean(targetAbs)
	tempAbs = filepath.Clean(tempAbs)

	// Check if targetAbs starts with tempAbs + separator
	return targetAbs == tempAbs ||
		strings.HasPrefix(targetAbs, tempAbs+string(filepath.Separator))
}
