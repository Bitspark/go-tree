package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bitspark.dev/go-tree/pkg/index"
	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/typesys"
)

// DependencyError represents a specific dependency-related error with context
type DependencyError struct {
	ImportPath string
	Version    string
	Module     string
	Reason     string
	Err        error
}

func (e *DependencyError) Error() string {
	msg := fmt.Sprintf("dependency error for %s@%s", e.ImportPath, e.Version)
	if e.Module != "" {
		msg += fmt.Sprintf(" in module %s", e.Module)
	}
	msg += fmt.Sprintf(": %s", e.Reason)
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

// DependencyManager handles dependency operations for the service
type DependencyManager struct {
	service      *Service
	replacements map[string]map[string]string // map[moduleDir]map[importPath]replacement
	inProgress   map[string]bool              // Track modules currently being loaded to detect circular deps
	dirCache     map[string]string            // Cache of already resolved dependency directories
	maxDepth     int                          // Maximum dependency loading depth
}

// NewDependencyManager creates a new DependencyManager
func NewDependencyManager(service *Service) *DependencyManager {
	var maxDepth int = 1 // Default value

	// Check if Config is initialized
	if service.Config != nil {
		maxDepth = service.Config.DependencyDepth
		if maxDepth <= 0 {
			maxDepth = 1 // Default to direct dependencies only
		}
	}

	return &DependencyManager{
		service:      service,
		replacements: make(map[string]map[string]string),
		inProgress:   make(map[string]bool),
		dirCache:     make(map[string]string),
		maxDepth:     maxDepth,
	}
}

// LoadDependencies loads all dependencies for all modules
func (dm *DependencyManager) LoadDependencies() error {
	// Process each module's dependencies
	for modPath, mod := range dm.service.Modules {
		if err := dm.LoadModuleDependencies(mod, 0); err != nil {
			return fmt.Errorf("error loading dependencies for module %s: %w", modPath, err)
		}
	}

	return nil
}

// LoadModuleDependencies loads dependencies for a specific module
func (dm *DependencyManager) LoadModuleDependencies(module *typesys.Module, depth int) error {
	// Skip if we've reached max depth
	if dm.maxDepth > 0 && depth >= dm.maxDepth {
		if dm.service.Config != nil && dm.service.Config.Verbose {
			fmt.Printf("Skipping deeper dependencies for %s (at depth %d, max %d)\n",
				module.Path, depth, dm.maxDepth)
		}
		return nil
	}

	// Read the go.mod file
	goModPath := filepath.Join(module.Dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return &DependencyError{
			Module: module.Path,
			Reason: "failed to read go.mod file",
			Err:    err,
		}
	}

	// Parse the dependencies
	deps, replacements, err := parseGoMod(string(content))
	if err != nil {
		return &DependencyError{
			Module: module.Path,
			Reason: "failed to parse go.mod",
			Err:    err,
		}
	}

	// Store replacements for this module
	dm.replacements[module.Dir] = replacements

	// Load each dependency
	for importPath, version := range deps {
		// Skip if already loaded
		if dm.service.isPackageLoaded(importPath) {
			continue
		}

		// Try to load the dependency
		if err := dm.loadDependency(module, importPath, version, depth); err != nil {
			// Log error but continue with other dependencies
			if dm.service.Config != nil && dm.service.Config.Verbose {
				fmt.Printf("Warning: %v\n", err)
			}
		}
	}

	return nil
}

// loadDependency loads a single dependency, considering replacements
func (dm *DependencyManager) loadDependency(fromModule *typesys.Module, importPath, version string, depth int) error {
	// Check for circular dependency
	depKey := importPath + "@" + version
	if dm.inProgress[depKey] {
		// We're already loading this dependency, circular reference detected
		if dm.service.Config != nil && dm.service.Config.Verbose {
			fmt.Printf("Circular dependency detected: %s\n", depKey)
		}
		return nil // Don't treat as error, just stop the recursion
	}

	// Mark as in progress
	dm.inProgress[depKey] = true
	defer func() {
		// Remove from in-progress when done
		delete(dm.inProgress, depKey)
	}()

	// Check for a replacement
	replacements := dm.replacements[fromModule.Dir]
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
			depDir, err = dm.findDependencyDir(replacement, version)
			if err != nil {
				if dm.service.Config != nil && dm.service.Config.DownloadMissing {
					// Try to download the replacement
					depDir, err = dm.EnsureDependencyDownloaded(replacement, version)
					if err != nil {
						return &DependencyError{
							ImportPath: importPath,
							Version:    version,
							Module:     fromModule.Path,
							Reason:     "could not locate or download replacement",
							Err:        err,
						}
					}
				} else {
					return &DependencyError{
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
		depDir, err = dm.findDependencyDir(importPath, version)
		if err != nil {
			if dm.service.Config != nil && dm.service.Config.DownloadMissing {
				// Try to download the dependency
				depDir, err = dm.EnsureDependencyDownloaded(importPath, version)
				if err != nil {
					return &DependencyError{
						ImportPath: importPath,
						Version:    version,
						Module:     fromModule.Path,
						Reason:     "could not locate or download dependency",
						Err:        err,
					}
				}
			} else {
				return &DependencyError{
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
		return &DependencyError{
			ImportPath: importPath,
			Version:    version,
			Module:     fromModule.Path,
			Reason:     "could not load dependency",
			Err:        err,
		}
	}

	// Store the module
	dm.service.Modules[depModule.Path] = depModule

	// Create an index for the module
	dm.service.Indices[depModule.Path] = index.NewIndex(depModule)

	// Store version information
	dm.service.recordPackageVersions(depModule, version)

	// Recursively load this module's dependencies with incremented depth
	if dm.service.Config != nil && dm.service.Config.WithDeps {
		if err := dm.LoadModuleDependencies(depModule, depth+1); err != nil {
			// Log but continue
			if dm.service.Config != nil && dm.service.Config.Verbose {
				fmt.Printf("Warning: %v\n", err)
			}
		}
	}

	return nil
}

// EnsureDependencyDownloaded attempts to download a dependency if it doesn't exist
func (dm *DependencyManager) EnsureDependencyDownloaded(importPath, version string) (string, error) {
	// First try to find it locally
	dir, err := dm.findDependencyDir(importPath, version)
	if err == nil {
		return dir, nil // Already exists
	}

	if dm.service.Config != nil && dm.service.Config.Verbose {
		fmt.Printf("Downloading dependency: %s@%s\n", importPath, version)
	}

	// Not found, try to download it
	cmd := exec.Command("go", "get", "-d", importPath+"@"+version)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &DependencyError{
			ImportPath: importPath,
			Version:    version,
			Reason:     "failed to download dependency",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Now try to find it again
	return dm.findDependencyDir(importPath, version)
}

// FindDependencyInformation executes 'go list -m' to get information about a module
func (dm *DependencyManager) FindDependencyInformation(importPath string) (string, string, error) {
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

// AddDependency adds a dependency to a module and loads it
func (dm *DependencyManager) AddDependency(moduleDir, importPath, version string) error {
	// First, check if module exists
	mod, ok := dm.FindModuleByDir(moduleDir)
	if !ok {
		return fmt.Errorf("module not found at directory: %s", moduleDir)
	}

	// Run go get to add the dependency
	cmd := exec.Command("go", "get", importPath+"@"+version)
	cmd.Dir = moduleDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &DependencyError{
			ImportPath: importPath,
			Version:    version,
			Module:     mod.Path,
			Reason:     "failed to add dependency",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Reload the module's dependencies
	return dm.LoadModuleDependencies(mod, 0)
}

// RemoveDependency removes a dependency from a module
func (dm *DependencyManager) RemoveDependency(moduleDir, importPath string) error {
	// First, check if module exists
	mod, ok := dm.FindModuleByDir(moduleDir)
	if !ok {
		return fmt.Errorf("module not found at directory: %s", moduleDir)
	}

	// Run go get with @none flag to remove the dependency
	cmd := exec.Command("go", "get", importPath+"@none")
	cmd.Dir = moduleDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &DependencyError{
			ImportPath: importPath,
			Module:     mod.Path,
			Reason:     "failed to remove dependency",
			Err:        fmt.Errorf("%w: %s", err, string(output)),
		}
	}

	// Reload the module's dependencies
	return dm.LoadModuleDependencies(mod, 0)
}

// FindModuleByDir finds a module by its directory
func (dm *DependencyManager) FindModuleByDir(dir string) (*typesys.Module, bool) {
	for _, mod := range dm.service.Modules {
		if mod.Dir == dir {
			return mod, true
		}
	}
	return nil, false
}

// BuildDependencyGraph builds a dependency graph for visualization
func (dm *DependencyManager) BuildDependencyGraph() map[string][]string {
	graph := make(map[string][]string)

	// Process each module
	for modPath, mod := range dm.service.Modules {
		// Read the go.mod file
		goModPath := filepath.Join(mod.Dir, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err != nil {
			continue // Skip modules without go.mod
		}

		// Parse the dependencies
		deps, _, err := parseGoMod(string(content))
		if err != nil {
			continue // Skip modules with unparseable go.mod
		}

		// Add dependencies to the graph
		depPaths := make([]string, 0, len(deps))
		for depPath := range deps {
			depPaths = append(depPaths, depPath)
		}
		graph[modPath] = depPaths
	}

	return graph
}

// findDependencyDir locates a dependency in the GOPATH or module cache
func (dm *DependencyManager) findDependencyDir(importPath, version string) (string, error) {
	// Check cache first
	cacheKey := importPath + "@" + version
	if cachedDir, ok := dm.dirCache[cacheKey]; ok {
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

	// Format the expected path in the module cache
	// Module paths use @ as a separator between the module path and version
	modPath := filepath.Join(gomodcache, importPath+"@"+version)
	if _, err := os.Stat(modPath); err == nil {
		// Cache the result before returning
		dm.dirCache[cacheKey] = modPath
		return modPath, nil
	}

	// Check if it's using a different version format (v prefix vs non-prefix)
	if len(version) > 0 && version[0] == 'v' {
		// Try without v prefix
		altVersion := version[1:]
		altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
		if _, err := os.Stat(altModPath); err == nil {
			// Cache the result before returning
			dm.dirCache[cacheKey] = altModPath
			return altModPath, nil
		}
	} else {
		// Try with v prefix
		altVersion := "v" + version
		altModPath := filepath.Join(gomodcache, importPath+"@"+altVersion)
		if _, err := os.Stat(altModPath); err == nil {
			// Cache the result before returning
			dm.dirCache[cacheKey] = altModPath
			return altModPath, nil
		}
	}

	// Check in old-style GOPATH mode (pre-modules)
	oldStylePath := filepath.Join(gopath, "src", importPath)
	if _, err := os.Stat(oldStylePath); err == nil {
		// Cache the result before returning
		dm.dirCache[cacheKey] = oldStylePath
		return oldStylePath, nil
	}

	// Try to use go list -m to find the module
	path, ver, err := dm.FindDependencyInformation(importPath)
	if err == nil {
		// Try the official version returned by go list
		modPath = filepath.Join(gomodcache, path+"@"+ver)
		if _, err := os.Stat(modPath); err == nil {
			// Cache the result before returning
			dm.dirCache[cacheKey] = modPath
			return modPath, nil
		}
	}

	return "", &DependencyError{
		ImportPath: importPath,
		Version:    version,
		Reason:     "could not find dependency in module cache or GOPATH",
	}
}
