package resolve

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/toolkit"
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

	// Toolchain for Go operations
	toolchain toolkit.GoToolchain

	// Filesystem for module operations
	fs toolkit.ModuleFS

	// Middleware chain for resolution
	middlewareChain *toolkit.MiddlewareChain
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
		toolchain:       toolkit.NewStandardGoToolchain(),
		fs:              toolkit.NewStandardModuleFS(),
		middlewareChain: toolkit.NewMiddlewareChain(),
	}
}

// WithToolchain sets a custom toolchain
func (r *ModuleResolver) WithToolchain(toolchain toolkit.GoToolchain) *ModuleResolver {
	r.toolchain = toolchain
	return r
}

// WithFS sets a custom filesystem
func (r *ModuleResolver) WithFS(fs toolkit.ModuleFS) *ModuleResolver {
	r.fs = fs
	return r
}

// Use adds middleware to the chain
func (r *ModuleResolver) Use(middleware ...toolkit.ResolutionMiddleware) *ModuleResolver {
	r.middlewareChain.Add(middleware...)
	return r
}

// ResolveModule resolves a module by path and version
func (r *ModuleResolver) ResolveModule(path, version string, opts ResolveOptions) (*typesys.Module, error) {
	// Create context for toolchain operations
	ctx := context.Background()

	// Apply any options from the middleware chain
	if opts.UseResolutionCache && r.middlewareChain != nil {
		// Add caching middleware if enabled
		r.middlewareChain.Add(toolkit.NewCachingMiddleware())
	}

	// Try to find the module location
	moduleDir, err := r.FindModuleLocation(path, version)
	if err != nil {
		if opts.DownloadMissing {
			if opts.Verbose {
				fmt.Printf("Module %s@%s not found, attempting to download...\n", path, version)
			}

			moduleDir, err = r.EnsureModuleAvailable(path, version)
			if err != nil {
				return nil, &ResolutionError{
					ImportPath: path,
					Version:    version,
					Reason:     "could not locate or download module",
					Err:        err,
				}
			}

			if opts.Verbose {
				fmt.Printf("Successfully downloaded module %s@%s to %s\n", path, version, moduleDir)
			}
		} else {
			return nil, &ResolutionError{
				ImportPath: path,
				Version:    version,
				Reason:     "could not locate module and auto-download is disabled",
				Err:        err,
			}
		}
	}

	// Execute middleware chain or directly load the module
	var module *typesys.Module

	// Create resolution function for middleware chain or direct execution
	resolveFunc := func() (*typesys.Module, error) {
		mod, err := loader.LoadModule(moduleDir, &typesys.LoadOptions{
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
		return mod, nil
	}

	// If middleware chain is empty, execute directly
	if r.middlewareChain != nil && len(r.middlewareChain.Middlewares()) > 0 {
		module, err = r.middlewareChain.Execute(ctx, path, version, resolveFunc)
	} else {
		module, err = resolveFunc()
	}

	if err != nil {
		return nil, err
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

// CircularDependencyError represents a circular dependency detection error
type CircularDependencyError struct {
	ImportPath string
	Version    string
	Module     string
	Path       []string
}

// Error returns a string representation of the error
func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s@%s in path: %s",
		e.ImportPath, e.Version, strings.Join(e.Path, " -> "))
}

// ResolveDependencies resolves dependencies for a module
func (r *ModuleResolver) ResolveDependencies(module *typesys.Module, depth int) error {
	// Create initial resolution path
	path := []string{module.Path}

	// Call helper function with path tracking
	return r.resolveDependenciesWithPath(module, depth, path)
}

// resolveDependenciesWithPath resolves dependencies with path tracking for circular dependency detection
func (r *ModuleResolver) resolveDependenciesWithPath(module *typesys.Module, depth int, path []string) error {
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
	content, err := r.fs.ReadFile(goModPath)
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

		// Check for circular dependency
		depKey := importPath + "@" + version
		if r.inProgress[depKey] {
			// Check if we should treat this as an error
			if r.Options.StrictCircularDeps {
				return &CircularDependencyError{
					ImportPath: importPath,
					Version:    version,
					Module:     module.Path,
					Path:       append(path, importPath),
				}
			}

			// Just log and continue
			if r.Options.Verbose {
				fmt.Printf("Circular dependency detected: %s -> %s\n",
					strings.Join(path, " -> "), importPath)
			}
			continue
		}

		// Mark as in progress
		r.inProgress[depKey] = true
		defer func(key string) {
			// Remove from in-progress when done
			delete(r.inProgress, key)
		}(depKey)

		// Build new path for this dependency
		newPath := append([]string{}, path...)
		newPath = append(newPath, importPath)

		// Try to load the dependency with path tracking
		if err := r.loadDependencyWithPath(module, importPath, version, depth, newPath); err != nil {
			// Log error but continue with other dependencies
			if r.Options.Verbose {
				fmt.Printf("Warning: %v\n", err)
			}
		}
	}

	return nil
}

// loadDependencyWithPath loads a single dependency with path tracking for circular dependency detection
func (r *ModuleResolver) loadDependencyWithPath(fromModule *typesys.Module, importPath, version string, depth int, path []string) error {
	// Handle replacement first
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
	r.resolvedModules[importPath+"@"+version] = depModule

	// Recursively load this module's dependencies with incremented depth and path
	newDepth := depth + 1
	if err := r.resolveDependenciesWithPath(depModule, newDepth, path); err != nil {
		// Log but continue
		if r.Options.Verbose {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	return nil
}

// FindModuleLocation finds a module's location in the filesystem
func (r *ModuleResolver) FindModuleLocation(importPath, version string) (string, error) {
	// Create context for toolchain operations
	ctx := context.Background()

	// Check cache first
	cacheKey := importPath
	if version != "" {
		cacheKey += "@" + version
	}

	if cachedDir, ok := r.locationCache[cacheKey]; ok {
		return cachedDir, nil
	}

	// Use toolchain to find the module
	modPath, err := r.toolchain.FindModule(ctx, importPath, version)
	if err == nil {
		// Cache the result before returning
		r.locationCache[cacheKey] = modPath
		return modPath, nil
	}

	// Try to use go list -m to find the module if no version is specified
	if version == "" {
		// If no version is specified, try to find the latest
		path, ver, err := r.resolveModuleInfo(importPath)
		if err == nil && path != "" {
			// Try the official version returned by go list
			modPath, err := r.toolchain.FindModule(ctx, path, ver)
			if err == nil {
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
	// Create context for toolchain operations
	ctx := context.Background()

	// First try to find it locally
	dir, err := r.FindModuleLocation(importPath, version)
	if err == nil {
		return dir, nil // Already exists
	}

	if r.Options.Verbose {
		fmt.Printf("Downloading module: %s@%s\n", importPath, version)
	}

	// Not found, try to download it with retries
	const maxRetries = 3
	var downloadErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if r.Options.Verbose && attempt > 1 {
			fmt.Printf("Retry %d/%d downloading module: %s@%s\n", attempt, maxRetries, importPath, version)
		}

		downloadErr = r.toolchain.DownloadModule(ctx, importPath, version)
		if downloadErr == nil {
			break
		}

		// If this is not the last attempt, wait a bit before retrying
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	if downloadErr != nil {
		return "", &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Reason:     fmt.Sprintf("failed to download module after %d attempts", maxRetries),
			Err:        downloadErr,
		}
	}

	// Verify the download by checking if we can now find the module
	dir, err = r.FindModuleLocation(importPath, version)
	if err != nil {
		return "", &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Reason:     "module was downloaded but cannot be found in module cache",
			Err:        err,
		}
	}

	if r.Options.Verbose {
		fmt.Printf("Successfully downloaded module to: %s\n", dir)
	}

	return dir, nil
}

// FindModuleVersion finds the latest version of a module
func (r *ModuleResolver) FindModuleVersion(importPath string) (string, error) {
	// Create context for toolchain operations
	ctx := context.Background()

	// Use toolchain to get module info
	_, version, err := r.toolchain.GetModuleInfo(ctx, importPath)
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
	content, err := r.fs.ReadFile(goModPath)
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

// Implement AddDependency and RemoveDependency to use the toolchain abstraction
func (r *ModuleResolver) AddDependency(module *typesys.Module, importPath, version string) error {
	if module == nil {
		return &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Reason:     "module cannot be nil",
		}
	}

	// Create context for toolchain operations
	ctx := context.Background()

	// Run go get to add the dependency
	versionSpec := importPath
	if version != "" {
		versionSpec += "@" + version
	}

	_, err := r.toolchain.RunCommand(ctx, "get", "-d", versionSpec)
	if err != nil {
		return &ResolutionError{
			ImportPath: importPath,
			Version:    version,
			Module:     module.Path,
			Reason:     "failed to add dependency",
			Err:        err,
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

	// Create context for toolchain operations
	ctx := context.Background()

	// Run go get with @none flag to remove the dependency
	_, err := r.toolchain.RunCommand(ctx, "get", importPath+"@none")
	if err != nil {
		return &ResolutionError{
			ImportPath: importPath,
			Module:     module.Path,
			Reason:     "failed to remove dependency",
			Err:        err,
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

// resolveModuleInfo executes 'go list -m' to get information about a module
func (r *ModuleResolver) resolveModuleInfo(importPath string) (string, string, error) {
	// Create context for toolchain operations
	ctx := context.Background()

	// Use toolchain to get module info
	return r.toolchain.GetModuleInfo(ctx, importPath)
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
