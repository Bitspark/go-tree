package materialize

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/saver"
	"bitspark.dev/go-tree/pkg/toolkit"
	"bitspark.dev/go-tree/pkg/typesys"
)

// ModuleMaterializer is the standard implementation of the Materializer interface
type ModuleMaterializer struct {
	Options MaterializeOptions
	Saver   saver.ModuleSaver

	// Toolchain for Go operations
	toolchain toolkit.GoToolchain

	// Filesystem for module operations
	fs toolkit.ModuleFS
}

// NewModuleMaterializer creates a new materializer with default options
func NewModuleMaterializer() *ModuleMaterializer {
	return NewModuleMaterializerWithOptions(DefaultMaterializeOptions())
}

// NewModuleMaterializerWithOptions creates a new materializer with the specified options
func NewModuleMaterializerWithOptions(options MaterializeOptions) *ModuleMaterializer {
	return &ModuleMaterializer{
		Options:   options,
		Saver:     saver.NewGoModuleSaver(),
		toolchain: toolkit.NewStandardGoToolchain(),
		fs:        toolkit.NewStandardModuleFS(),
	}
}

// WithToolchain sets a custom toolchain
func (m *ModuleMaterializer) WithToolchain(toolchain toolkit.GoToolchain) *ModuleMaterializer {
	m.toolchain = toolchain
	return m
}

// WithFS sets a custom filesystem
func (m *ModuleMaterializer) WithFS(fs toolkit.ModuleFS) *ModuleMaterializer {
	m.fs = fs
	return m
}

// Materialize writes a module to disk with dependencies
func (m *ModuleMaterializer) Materialize(module *typesys.Module, opts MaterializeOptions) (*Environment, error) {
	return m.materializeModules([]*typesys.Module{module}, opts)
}

// MaterializeForExecution prepares a module for running
func (m *ModuleMaterializer) MaterializeForExecution(module *typesys.Module, opts MaterializeOptions) (*Environment, error) {
	env, err := m.Materialize(module, opts)
	if err != nil {
		return nil, err
	}

	// Run additional setup for execution
	if opts.RunGoModTidy {
		modulePath, ok := env.ModulePaths[module.Path]
		if ok {
			// Create context for toolchain operations
			ctx := context.Background()

			// Run go mod tidy using toolchain abstraction
			if opts.Verbose {
				fmt.Printf("Running go mod tidy in %s\n", modulePath)
			}

			// Set working directory for the command
			customToolchain := *m.toolchain.(*toolkit.StandardGoToolchain)
			customToolchain.WorkDir = modulePath

			output, err := customToolchain.RunCommand(ctx, "mod", "tidy")
			if err != nil {
				return env, &MaterializationError{
					ModulePath: module.Path,
					Message:    "failed to run go mod tidy",
					Err:        fmt.Errorf("%w: %s", err, string(output)),
				}
			}
		}
	}

	return env, nil
}

// MaterializeMultipleModules materializes multiple modules together
func (m *ModuleMaterializer) MaterializeMultipleModules(modules []*typesys.Module, opts MaterializeOptions) (*Environment, error) {
	return m.materializeModules(modules, opts)
}

// materializeModules is the core materialization implementation
func (m *ModuleMaterializer) materializeModules(modules []*typesys.Module, opts MaterializeOptions) (*Environment, error) {
	// Use provided options or fall back to defaults
	if opts.TargetDir == "" && len(opts.EnvironmentVars) == 0 && !opts.RunGoModTidy &&
		!opts.IncludeTests && !opts.Verbose && !opts.Preserve {
		opts = m.Options
	}

	// Create root directory if needed
	rootDir := opts.TargetDir
	isTemporary := false

	if rootDir == "" {
		// Create a temporary directory
		var err error
		rootDir, err = m.fs.TempDir("", "go-tree-materialized-*")
		if err != nil {
			return nil, &MaterializationError{
				Message: "failed to create temporary directory",
				Err:     err,
			}
		}
		isTemporary = true
	} else {
		// Ensure the directory exists
		if err := m.fs.MkdirAll(rootDir, 0755); err != nil {
			return nil, &MaterializationError{
				Message: "failed to create target directory",
				Err:     err,
			}
		}
	}

	// Create environment
	env := &Environment{
		RootDir:     rootDir,
		ModulePaths: make(map[string]string),
		IsTemporary: isTemporary && !opts.Preserve,
		EnvVars:     make(map[string]string),
	}

	// Process each module
	for _, module := range modules {
		if err := m.materializeModule(module, rootDir, env, opts); err != nil {
			// Clean up on error unless Preserve is set
			if env.IsTemporary && !opts.Preserve {
				env.Cleanup()
			}
			return nil, err
		}
	}

	return env, nil
}

// materializeModule materializes a single module
func (m *ModuleMaterializer) materializeModule(module *typesys.Module, rootDir string, env *Environment, opts MaterializeOptions) error {
	// Determine module directory based on layout strategy
	var moduleDir string

	switch opts.LayoutStrategy {
	case FlatLayout:
		// Use module name as directory name
		safeName := strings.ReplaceAll(module.Path, "/", "_")
		moduleDir = filepath.Join(rootDir, safeName)

	case HierarchicalLayout:
		// Use full path hierarchy
		moduleDir = filepath.Join(rootDir, module.Path)

	case GoPathLayout:
		// Use GOPATH-like layout with src directory
		moduleDir = filepath.Join(rootDir, "src", module.Path)

	default:
		// Default to flat layout
		safeName := strings.ReplaceAll(module.Path, "/", "_")
		moduleDir = filepath.Join(rootDir, safeName)
	}

	// Create module directory
	if err := m.fs.MkdirAll(moduleDir, 0755); err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to create module directory",
			Err:        err,
		}
	}

	// Save the module using the saver
	if err := m.Saver.SaveTo(module, moduleDir); err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to save module",
			Err:        err,
		}
	}

	// Store the module path in the environment
	env.ModulePaths[module.Path] = moduleDir

	// Handle dependencies based on policy
	if opts.DependencyPolicy != NoDependencies {
		if err := m.materializeDependencies(module, rootDir, env, opts); err != nil {
			return err
		}
	}

	// Generate/update go.mod file with proper dependencies and replacements
	if err := m.generateGoMod(module, moduleDir, env, opts); err != nil {
		return err
	}

	return nil
}

// materializeDependencies materializes dependencies of a module
func (m *ModuleMaterializer) materializeDependencies(module *typesys.Module, rootDir string, env *Environment, opts MaterializeOptions) error {
	// Parse the go.mod file to get dependencies
	goModPath := filepath.Join(module.Dir, "go.mod")
	content, err := m.fs.ReadFile(goModPath)
	if err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to read go.mod file",
			Err:        err,
		}
	}

	deps, replacements, err := parseGoMod(string(content))
	if err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to parse go.mod",
			Err:        err,
		}
	}

	// Skip if we have no dependencies
	if len(deps) == 0 {
		return nil
	}

	// Process each dependency based on the selected policy
	for depPath, version := range deps {
		// Skip if already materialized
		if _, ok := env.ModulePaths[depPath]; ok {
			continue
		}

		// If we have a replacement, handle it
		if replacement, ok := replacements[depPath]; ok {
			if strings.HasPrefix(replacement, ".") || strings.HasPrefix(replacement, "/") {
				// Local replacement - directory path
				var resolvedPath string
				if strings.HasPrefix(replacement, ".") {
					// Relative path - resolve relative to the original module
					resolvedPath = filepath.Join(module.Dir, replacement)
				} else {
					// Absolute path
					resolvedPath = replacement
				}

				// Copy the directory to the materialization location
				moduleDir, err := m.materializeLocalModule(resolvedPath, depPath, rootDir, env, opts)
				if err != nil {
					if opts.Verbose {
						fmt.Printf("Warning: failed to materialize local replacement %s: %v\n", depPath, err)
					}
					continue
				}

				// Store the module path
				env.ModulePaths[depPath] = moduleDir

				// If recursive and all dependencies policy, process its dependencies too
				if opts.DependencyPolicy == AllDependencies {
					// Load information about the module to use in recursive call
					depModule := &typesys.Module{
						Path: depPath,
						Dir:  resolvedPath,
					}
					if err := m.materializeDependencies(depModule, rootDir, env, opts); err != nil {
						if opts.Verbose {
							fmt.Printf("Warning: failed to materialize dependencies of %s: %v\n", depPath, err)
						}
					}
				}
			} else {
				// Remote replacement - module path
				// We can just let the go.mod file handle this via replace directive
				continue
			}
		} else {
			// No replacement - regular dependency
			// Try to find the module in the module cache
			ctx := context.Background()
			depDir, err := m.toolchain.FindModule(ctx, depPath, version)
			if err != nil {
				if opts.Verbose {
					fmt.Printf("Warning: could not find module %s@%s in cache: %v\n", depPath, version, err)
				}
				continue
			}

			// Copy the module to the materialization location
			moduleDir, err := m.materializeLocalModule(depDir, depPath, rootDir, env, opts)
			if err != nil {
				if opts.Verbose {
					fmt.Printf("Warning: failed to materialize dependency %s: %v\n", depPath, err)
				}
				continue
			}

			// Store the module path
			env.ModulePaths[depPath] = moduleDir

			// If recursive and all dependencies policy, process its dependencies too
			if opts.DependencyPolicy == AllDependencies {
				// Load information about the module to use in recursive call
				depModule := &typesys.Module{
					Path: depPath,
					Dir:  depDir,
				}
				if err := m.materializeDependencies(depModule, rootDir, env, opts); err != nil {
					if opts.Verbose {
						fmt.Printf("Warning: failed to materialize dependencies of %s: %v\n", depPath, err)
					}
				}
			}
		}
	}

	return nil
}

// materializeLocalModule copies a module from a local directory to the materialization location
func (m *ModuleMaterializer) materializeLocalModule(srcDir, modulePath, rootDir string, env *Environment, opts MaterializeOptions) (string, error) {
	// Determine module directory based on layout strategy
	var moduleDir string

	switch opts.LayoutStrategy {
	case FlatLayout:
		// Use module name as directory name
		safeName := strings.ReplaceAll(modulePath, "/", "_")
		moduleDir = filepath.Join(rootDir, safeName)

	case HierarchicalLayout:
		// Use full path hierarchy
		moduleDir = filepath.Join(rootDir, modulePath)

	case GoPathLayout:
		// Use GOPATH-like layout with src directory
		moduleDir = filepath.Join(rootDir, "src", modulePath)

	default:
		// Default to flat layout
		safeName := strings.ReplaceAll(modulePath, "/", "_")
		moduleDir = filepath.Join(rootDir, safeName)
	}

	// Create module directory
	if err := m.fs.MkdirAll(moduleDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create module directory: %w", err)
	}

	// Create a temporary module representation for the saver
	tempModule := &typesys.Module{
		Path: modulePath,
		Dir:  srcDir,
	}

	// Save the module using the saver
	if err := m.Saver.SaveTo(tempModule, moduleDir); err != nil {
		return "", fmt.Errorf("failed to save module: %w", err)
	}

	return moduleDir, nil
}

// generateGoMod generates or updates the go.mod file for a materialized module
func (m *ModuleMaterializer) generateGoMod(module *typesys.Module, moduleDir string, env *Environment, opts MaterializeOptions) error {
	// Read the original go.mod
	originalGoModPath := filepath.Join(module.Dir, "go.mod")
	content, err := m.fs.ReadFile(originalGoModPath)
	if err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to read original go.mod file",
			Err:        err,
		}
	}

	// Parse the go.mod file
	deps, replacements, err := parseGoMod(string(content))
	if err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to parse go.mod",
			Err:        err,
		}
	}

	// Create the new go.mod content
	var buf bytes.Buffer

	// Write module declaration
	buf.WriteString(fmt.Sprintf("module %s\n\n", module.Path))

	// Write go version
	goVersion := extractGoVersion(string(content))
	if goVersion != "" {
		buf.WriteString(fmt.Sprintf("go %s\n\n", goVersion))
	} else {
		buf.WriteString("go 1.16\n\n") // Default to Go 1.16 if not specified
	}

	// Write requires
	if len(deps) > 0 {
		if len(deps) == 1 {
			// Single dependency, write as a standalone require
			for path, version := range deps {
				buf.WriteString(fmt.Sprintf("require %s %s\n\n", path, version))
			}
		} else {
			// Multiple dependencies, write as a block
			buf.WriteString("require (\n")
			for path, version := range deps {
				buf.WriteString(fmt.Sprintf("\t%s %s\n", path, version))
			}
			buf.WriteString(")\n\n")
		}
	}

	// Generate replace directives if needed
	if opts.ReplaceStrategy != NoReplace {
		// Get materialized dependencies
		replacePaths := make(map[string]string)

		// Also consider existing replacements
		for origPath, replPath := range replacements {
			replacePaths[origPath] = replPath
		}

		for depPath := range deps {
			if depDir, ok := env.ModulePaths[depPath]; ok {
				// We have this dependency materialized, add a replace directive
				var replacePath string

				if opts.ReplaceStrategy == RelativeReplace {
					// Use relative path
					relPath, err := filepath.Rel(moduleDir, depDir)
					if err == nil {
						replacePath = relPath
					} else {
						// Fall back to absolute path if relative path fails
						replacePath = depDir
					}
				} else {
					// Use absolute path
					replacePath = depDir
				}

				replacePaths[depPath] = replacePath
			}
		}

		// Write replace directives
		if len(replacePaths) > 0 {
			if len(replacePaths) == 1 {
				// Single replacement, write as a standalone replace
				for path, replacement := range replacePaths {
					buf.WriteString(fmt.Sprintf("replace %s => %s\n", path, replacement))
				}
			} else {
				// Multiple replacements, write as a block
				buf.WriteString("replace (\n")
				for path, replacement := range replacePaths {
					buf.WriteString(fmt.Sprintf("\t%s => %s\n", path, replacement))
				}
				buf.WriteString(")\n")
			}
		}
	}

	// Write the new go.mod file
	targetGoModPath := filepath.Join(moduleDir, "go.mod")
	if err := m.fs.WriteFile(targetGoModPath, buf.Bytes(), 0644); err != nil {
		return &MaterializationError{
			ModulePath: module.Path,
			Message:    "failed to write go.mod file",
			Err:        err,
		}
	}

	return nil
}

// extractGoVersion extracts the Go version from a go.mod file
func extractGoVersion(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return strings.TrimSpace(line[3:])
		}
	}
	return ""
}

// parseGoMod parses a go.mod file and extracts dependencies and replacements
func parseGoMod(content string) (map[string]string, map[string]string, error) {
	deps := make(map[string]string)
	replacements := make(map[string]string)

	// Simple line-by-line parsing (a more robust implementation would use a proper parser)
	lines := strings.Split(content, "\n")
	inRequire := false
	inReplace := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle require blocks
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		// Handle replace blocks
		if line == "replace (" {
			inReplace = true
			continue
		}
		if inReplace && line == ")" {
			inReplace = false
			continue
		}

		// Handle standalone require
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line[len("require "):])
			if len(parts) >= 2 {
				deps[parts[0]] = parts[1]
			}
			continue
		}

		// Handle require within block
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps[parts[0]] = parts[1]
			}
			continue
		}

		// Handle standalone replace
		if strings.HasPrefix(line, "replace ") {
			handleReplace(line[len("replace "):], replacements)
			continue
		}

		// Handle replace within block
		if inReplace {
			handleReplace(line, replacements)
			continue
		}
	}

	return deps, replacements, nil
}

// handleReplace parses a replacement line from go.mod
func handleReplace(line string, replacements map[string]string) {
	// Format: original => replacement
	parts := strings.Split(line, "=>")
	if len(parts) != 2 {
		return
	}

	original := strings.TrimSpace(parts[0])
	replacement := strings.TrimSpace(parts[1])

	// Handle version in replacement
	repParts := strings.Fields(replacement)
	if len(repParts) >= 1 {
		replacement = repParts[0]
	}

	// Handle version in original
	origParts := strings.Fields(original)
	if len(origParts) >= 1 {
		original = origParts[0]
	}

	replacements[original] = replacement
}
