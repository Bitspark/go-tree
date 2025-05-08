package resolve

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/typesys"
)

// ModuleResolver is the standard implementation of the Resolver interface
type ModuleResolver struct {
	// Options for resolution
	Options ResolveOptions

	// Cache of resolved modules
	resolvedModules map[string]*typesys.Module

	// Cache of module locations
	locationCache map[string]string

	// Track modules being processed (for circular dependency detection)
	inProgress map[string]bool

	// Parsed go.mod replacements: map[moduleDir]map[importPath]replacement
	replacements map[string]map[string]string
}

// NewModuleResolver creates a new module resolver with default options
func NewModuleResolver() *ModuleResolver {
	return NewModuleResolverWithOptions(DefaultResolveOptions())
}

// NewModuleResolverWithOptions creates a new module resolver with the specified options
func NewModuleResolverWithOptions(options ResolveOptions) *ModuleResolver {
	return &ModuleResolver{
		Options:         options,
		resolvedModules: make(map[string]*typesys.Module),
		locationCache:   make(map[string]string),
		inProgress:      make(map[string]bool),
		replacements:    make(map[string]map[string]string),
	}
}

// ResolveModule resolves a module by path and version
func (r *ModuleResolver) ResolveModule(path, version string, opts ResolveOptions) (*typesys.Module, error) {
	// Try to find the module location
	moduleDir, err := r.FindModuleLocation(path, version)
	if err != nil {
		if opts.DownloadMissing {
			moduleDir, err = r.EnsureModuleAvailable(path, version)
			if err != nil {
				return nil, &ResolutionError{
					ImportPath: path,
					Version:    version,
					Reason:     "could not locate or download module",
					Err:        err,
				}
			}
		} else {
			return nil, &ResolutionError{
				ImportPath: path,
				Version:    version,
				Reason:     "could not locate module",
				Err:        err,
			}
		}
	}

	// Load the module
	module, err := loader.LoadModule(moduleDir, &typesys.LoadOptions{
		IncludeTests: opts.IncludeTests,
	})
	if err != nil {
		return nil, &ResolutionError{
			ImportPath: path,
			Version:    version,
			Reason:     "could not load module",
			Err:        err,
		}
	}

	// Cache the resolved module
	cacheKey := path
	if version != "" {
		cacheKey += "@" + version
	}
	r.resolvedModules[cacheKey] = module

	// Resolve dependencies if needed
	if opts.DependencyPolicy != NoDependencies {
		depth := opts.DependencyDepth
		if opts.DependencyPolicy == DirectDependenciesOnly && depth > 1 {
			depth = 1
		}

		if err := r.ResolveDependencies(module, depth); err != nil {
			return module, err // Return the module even if dependencies failed
		}
	}

	return module, nil
}

// ResolveDependencies resolves dependencies for a module
func (r *ModuleResolver) ResolveDependencies(module *typesys.Module, depth int) error {
	// Skip if we've reached max depth
	if r.Options.DependencyDepth > 0 && depth >= r.Options.DependencyDepth {
		if r.Options.Verbose {
			fmt.Printf("Skipping deeper dependencies for %s (at depth %d, max %d)\n",
				module.Path, depth, r.Options.DependencyDepth)
		}
		return nil
	}

	// Read the go.mod file
	goModPath := filepath.Join(module.Dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return &ResolutionError{
			Module: module.Path,
			Reason: "failed to read go.mod file",
			Err:    err,
		}
	}

	// Parse the dependencies
	deps, replacements, err := parseGoMod(string(content))
	if err != nil {
		return &ResolutionError{
			Module: module.Path,
			Reason: "failed to parse go.mod",
			Err:    err,
		}
	}

	// Store replacements for this module
	r.replacements[module.Dir] = replacements

	// Load each dependency
	for importPath, version := range deps {
		// Skip if already loaded
		if r.isModuleLoaded(importPath) {
			continue
		}

		// Try to load the dependency
		if err := r.loadDependency(module, importPath, version, depth); err != nil {
			// Log error but continue with other dependencies
			if r.Options.Verbose {
				fmt.Printf("Warning: %v\n", err)
			}
		}
	}

	return nil
}

// loadDependency loads a single dependency, considering replacements
func (r *ModuleResolver) loadDependency(fromModule *typesys.Module, importPath, version string, depth int) error {
	// Check for circular dependency
	depKey := importPath + "@" + version
	if r.inProgress[depKey] {
		// We're already loading this dependency, circular reference detected
		if r.Options.Verbose {
			fmt.Printf("Circular dependency detected: %s\n", depKey)
		}
		return nil // Don't treat as error, just stop the recursion
	}

	// Mark as in progress
	r.inProgress[depKey] = true
	defer func() {
		// Remove from in-progress when done
		delete(r.inProgress, depKey)
	}()

	// Check for a replacement
	replacements := r.replacements[fromModule.Dir]
	replacement, hasReplacement := replacements[importPath]

	var depDir string
	var err error

	if hasReplacement {
		// Handle the replacement
		if strings.HasPrefix(replacement, ".") || strings.HasPrefix(replacement, "/") {
			// Local filesystem replacement
			if strings.HasPrefix(replacement, ".") {
				replacement = filepath.Join(fromModule.Dir, replacement)
			}
			depDir = replacement
		} else {
			// Remote replacement, find in cache
			depDir, err = r.FindModuleLocation(replacement, version)
			if err != nil {
				if r.Options.DownloadMissing {
					// Try to download the replacement
					depDir, err = r.EnsureModuleAvailable(replacement, version)
					if err != nil {
						return &ResolutionError{
							ImportPath: importPath,
							Version:    version,
							Module:     fromModule.Path,
							Reason:     "could not locate or download replacement",
							Err:        err,
						}
					}
				} else {
					return &ResolutionError{
						ImportPath: importPath,
						Version:    version,
						Module:     fromModule.Path,
						Reason:     "could not locate replacement",
						Err:        err,
					}
				}
			}
		}
	} else {
		// Standard module resolution
		depDir, err = r.FindModuleLocation(importPath, version)
		if err != nil {
			if r.Options.DownloadMissing {
				// Try to download the dependency
				depDir, err = r.EnsureModuleAvailable(importPath, version)
				if err != nil {
					return &ResolutionError{
						ImportPath: importPath,
						Version:    version,
						Module:     fromModule.Path,
						Reason:     "could not locate or download dependency",
						Err:        err,
					}
				}
			} else {
				return &ResolutionError{
					ImportPath: importPath,
					Version:    version,
					Module:     fromModule.Path,
					Reason:     "could not locate dependency",
					Err:        err,
				}
			}
		}
	}

	// Load the module
	depModule, err := loader.LoadModule(depDir, &typesys.LoadOptions{
		IncludeTests: false, // Usually don't need tests from dependencies
	})
	if err != nil {
		return &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Module:     fromModule.Path,
			Reason:     "could not load dependency",
			Err:        err,
		}
	}

	// Store the resolved module
	r.resolvedModules[depKey] = depModule

	// Recursively load this module's dependencies with incremented depth
	newDepth := depth + 1
	if err := r.ResolveDependencies(depModule, newDepth); err != nil {
		// Log but continue
		if r.Options.Verbose {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return nil
}

// FindModuleLocation finds a module's location in the filesystem
func (r *ModuleResolver) FindModuleLocation(importPath, version string) (string, error) {
	// Check cache first
	cacheKey := importPath
	if version != "" {
		cacheKey += "@" + version
	}

	if cachedDir, ok := r.locationCache[cacheKey]; ok {
		return cachedDir, nil
	}

	// Check GOPATH/pkg/mod
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Fall back to default GOPATH if not set
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("GOPATH not set and could not determine home directory: %w", err)
		}
		gopath = filepath.Join(home, "go")
	}

	// Check GOMODCACHE if available (introduced in Go 1.15)
	gomodcache := os.Getenv("GOMODCACHE")
	if gomodcache == "" {
		// Default location is $GOPATH/pkg/mod
		gomodcache = filepath.Join(gopath, "pkg", "mod")
	}

	// If version is specified, try the module cache
	if version != "" {
		// Format the expected path in the module cache
		// Module paths use @ as a separator between the module path and version
		modPath := filepath.Join(gomodcache, importPath+"@"+version)
		if _, err := os.Stat(modPath); err == nil {
			// Cache the result before returning
			r.locationCache[cacheKey] = modPath
			return modPath, nil
		}

		// Check if it's using a different version format (v prefix vs non-prefix)
		if len(version) > 0 && version[0] == 'v' {
			// Try without v prefix
			altVersion := version[1:]
			altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
			if _, err := os.Stat(altModPath); err == nil {
				// Cache the result before returning
				r.locationCache[cacheKey] = altModPath
				return altModPath, nil
			}
		} else {
			// Try with v prefix
			altVersion := "v" + version
			altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
			if _, err := os.Stat(altModPath); err == nil {
				// Cache the result before returning
				r.locationCache[cacheKey] = altModPath
				return altModPath, nil
			}
		}
	}

	// Check in old-style GOPATH mode (pre-modules)
	oldStylePath := filepath.Join(gopath, "src", importPath)
	if _, err := os.Stat(oldStylePath); err == nil {
		// Cache the result before returning
		r.locationCache[cacheKey] = oldStylePath
		return oldStylePath, nil
	}

	// Try to use go list -m to find the module
	if version == "" {
		// If no version is specified, try to find the latest
		path, ver, err := r.resolveModuleInfo(importPath)
		if err == nil && path != "" {
			// Try the official version returned by go list
			modPath := filepath.Join(gomodcache, path+"@"+ver)
			if _, err := os.Stat(modPath); err == nil {
				// Cache the result before returning
				r.locationCache[cacheKey] = modPath
				return modPath, nil
			}
		}
	}

	return "", &ResolutionError{
		ImportPath: importPath,
		Version:    version,
		Reason:     "could not find module in module cache or GOPATH",
	}
}

// EnsureModuleAvailable ensures a module is available, downloading if necessary
func (r *ModuleResolver) EnsureModuleAvailable(importPath, version string) (string, error) {
	// First try to find it locally
	dir, err := r.FindModuleLocation(importPath, version)
	if err == nil {
		return dir, nil // Already exists
	}

	if r.Options.Verbose {
		fmt.Printf("Downloading module: %s@%s\n", importPath, version)
	}

	// Not found, try to download it
	versionSpec := importPath
	if version != "" {
		versionSpec += "@" + version
	}

	cmd := exec.Command("go", "get", "-d", versionSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Reason:     "failed to download module",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Now try to find it again
	return r.FindModuleLocation(importPath, version)
}

// FindModuleVersion finds the latest version of a module
func (r *ModuleResolver) FindModuleVersion(importPath string) (string, error) {
	_, version, err := r.resolveModuleInfo(importPath)
	if err != nil {
		return "", &ResolutionError{
			ImportPath: importPath,
			Reason:     "failed to find module version",
			Err:        err,
		}
	}

	return version, nil
}

// BuildDependencyGraph builds a dependency graph for visualization
func (r *ModuleResolver) BuildDependencyGraph(module *typesys.Module) (map[string][]string, error) {
	graph := make(map[string][]string)

	// Read the go.mod file
	goModPath := filepath.Join(module.Dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, &ResolutionError{
			Module: module.Path,
			Reason: "failed to read go.mod file",
			Err:    err,
		}
	}

	// Parse the dependencies
	deps, _, err := parseGoMod(string(content))
	if err != nil {
		return nil, &ResolutionError{
			Module: module.Path,
			Reason: "failed to parse go.mod",
			Err:    err,
		}
	}

	// Add dependencies to the graph
	depPaths := make([]string, 0, len(deps))
	for depPath := range deps {
		depPaths = append(depPaths, depPath)

		// Recursively build the graph for this dependency
		depModule, ok := r.getResolvedModule(depPath)
		if ok {
			depGraph, err := r.BuildDependencyGraph(depModule)
			if err != nil {
				// Log error but continue
				if r.Options.Verbose {
					fmt.Printf("Warning: %v\n", err)
				}
			} else {
				// Merge the dependency's graph with the main graph
				for k, v := range depGraph {
					graph[k] = v
				}
			}
		}
	}

	graph[module.Path] = depPaths
	return graph, nil
}

// resolveModuleInfo executes 'go list -m' to get information about a module
func (r *ModuleResolver) resolveModuleInfo(importPath string) (string, string, error) {
	cmd := exec.Command("go", "list", "-m", importPath)
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get module information for %s: %w", importPath, err)
	}

	// Parse output (format: "path version")
	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected output format from go list -m: %s", output)
	}

	path := parts[0]
	version := parts[1]

	return path, version, nil
}

// isModuleLoaded checks if a module is already loaded
func (r *ModuleResolver) isModuleLoaded(importPath string) bool {
	for _, mod := range r.resolvedModules {
		if mod.Path == importPath {
			return true
		}

		// Check if any package in this module matches the import path
		for pkgPath := range mod.Packages {
			if pkgPath == importPath {
				return true
			}
		}
	}
	return false
}

// getResolvedModule tries to find a resolved module by import path
func (r *ModuleResolver) getResolvedModule(importPath string) (*typesys.Module, bool) {
	// First try exact match by module path
	for _, mod := range r.resolvedModules {
		if mod.Path == importPath {
			return mod, true
		}
	}

	// Then try by package path
	for _, mod := range r.resolvedModules {
		if _, ok := mod.Packages[importPath]; ok {
			return mod, true
		}
	}

	return nil, false
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
				// Ensure version has v prefix if numeric
				version := parts[1]
				if len(version) > 0 && version[0] >= '0' && version[0] <= '9' {
					version = "v" + version
				}
				deps[parts[0]] = version
			}
			continue
		}

		// Handle require within block
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Ensure version has v prefix if numeric
				version := parts[1]
				if len(version) > 0 && version[0] >= '0' && version[0] <= '9' {
					version = "v" + version
				}
				deps[parts[0]] = version
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

// AddDependency adds a dependency to a module and loads it
func (r *ModuleResolver) AddDependency(module *typesys.Module, importPath, version string) error {
	if module == nil {
		return &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Reason:     "module cannot be nil",
		}
	}

	// Run go get to add the dependency
	cmd := exec.Command("go", "get", importPath+"@"+version)
	cmd.Dir = module.Dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Module:     module.Path,
			Reason:     "failed to add dependency",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Reload the module's dependencies
	return r.ResolveDependencies(module, 0)
}

// RemoveDependency removes a dependency from a module
func (r *ModuleResolver) RemoveDependency(module *typesys.Module, importPath string) error {
	if module == nil {
		return &ResolutionError{
			ImportPath: importPath,
			Reason:     "module cannot be nil",
		}
	}

	// Run go get with @none flag to remove the dependency
	cmd := exec.Command("go", "get", importPath+"@none")
	cmd.Dir = module.Dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ResolutionError{
			ImportPath: importPath,
			Module:     module.Path,
			Reason:     "failed to remove dependency",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Reload the module's dependencies
	return r.ResolveDependencies(module, 0)
}

// FindModuleByDir finds a module by its directory
func (r *ModuleResolver) FindModuleByDir(dir string) (*typesys.Module, bool) {
	// Check all resolved modules
	for _, mod := range r.resolvedModules {
		if mod.Dir == dir {
			return mod, true
		}
	}
	return nil, false
}
