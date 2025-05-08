// Package service provides a unified interface to Go-Tree functionality
package service

import (
	"fmt"
	"go/types"

	"bitspark.dev/go-tree/pkg/index"
	"bitspark.dev/go-tree/pkg/loader"
	"bitspark.dev/go-tree/pkg/typesys"
)

// Config holds service configuration with multi-module support
type Config struct {
	// Core parameters
	ModuleDir    string // Main module directory
	IncludeTests bool   // Whether to include test files

	// Multi-module parameters
	WithDeps        bool                     // Whether to load dependencies
	DependencyDepth int                      // Maximum depth for dependency loading (0 means only direct dependencies)
	DownloadMissing bool                     // Whether to download missing dependencies
	ExtraModules    []string                 // Additional module directories to load
	ModuleConfig    map[string]*ModuleConfig // Per-module configuration
	Verbose         bool                     // Enable verbose logging
}

// ModuleConfig holds configuration for a specific module
type ModuleConfig struct {
	IncludeTests  bool
	AnalysisDepth int // How deep to analyze this module
}

// ModulePackage associates a package with its containing module and version
type ModulePackage struct {
	Module     *typesys.Module
	Package    *typesys.Package
	ImportPath string
	Version    string // Semver version from go.mod
}

// Service provides a unified interface to Go-Tree functionality
type Service struct {
	// Multiple modules support
	Modules map[string]*typesys.Module // Modules indexed by module path
	Indices map[string]*index.Index    // Indices for each module

	// Main module (the one specified in ModuleDir)
	MainModulePath string

	// Version tracking
	PackageVersions map[string]map[string]*ModulePackage // map[importPath]map[version]*ModulePackage

	// Dependency management
	DependencyManager *DependencyManager

	// Configuration
	Config *Config
}

// NewService creates a new multi-module service instance
func NewService(config *Config) (*Service, error) {
	service := &Service{
		Modules:         make(map[string]*typesys.Module),
		Indices:         make(map[string]*index.Index),
		PackageVersions: make(map[string]map[string]*ModulePackage),
		Config:          config,
	}

	// Load main module first
	mainModule, err := loader.LoadModule(config.ModuleDir, &typesys.LoadOptions{
		IncludeTests: config.IncludeTests,
	})
	if err != nil {
		return nil, err
	}

	service.MainModulePath = mainModule.Path
	service.Modules[mainModule.Path] = mainModule
	service.Indices[mainModule.Path] = index.NewIndex(mainModule)

	// Load extra modules if specified
	for _, moduleDir := range config.ExtraModules {
		moduleConfig := config.ModuleConfig[moduleDir]
		includeTests := config.IncludeTests
		if moduleConfig != nil {
			includeTests = moduleConfig.IncludeTests
		}

		module, err := loader.LoadModule(moduleDir, &typesys.LoadOptions{
			IncludeTests: includeTests,
		})
		if err != nil {
			return nil, err
		}

		service.Modules[module.Path] = module
		service.Indices[module.Path] = index.NewIndex(module)
	}

	// Initialize dependency manager
	service.DependencyManager = NewDependencyManager(service)

	// Load dependencies if requested
	if config.WithDeps {
		if err := service.loadDependencies(); err != nil {
			return nil, err
		}
	}

	return service, nil
}

// GetModule returns a module by its path
func (s *Service) GetModule(modulePath string) *typesys.Module {
	return s.Modules[modulePath]
}

// GetMainModule returns the main module
func (s *Service) GetMainModule() *typesys.Module {
	return s.Modules[s.MainModulePath]
}

// FindSymbolsAcrossModules finds symbols by name across all loaded modules
func (s *Service) FindSymbolsAcrossModules(name string) ([]*typesys.Symbol, error) {
	var results []*typesys.Symbol

	for _, idx := range s.Indices {
		symbols := idx.FindSymbolsByName(name)
		results = append(results, symbols...)
	}

	return results, nil
}

// FindSymbolsIn finds symbols by name in a specific module
func (s *Service) FindSymbolsIn(modulePath string, name string) ([]*typesys.Symbol, error) {
	idx, ok := s.Indices[modulePath]
	if !ok {
		return nil, fmt.Errorf("module %s not found", modulePath)
	}
	return idx.FindSymbolsByName(name), nil
}

// ResolveImport resolves an import path to a package, checking in the source module first
func (s *Service) ResolveImport(importPath string, fromModule string) (*typesys.Package, error) {
	// Try to resolve in the source module first
	if mod := s.Modules[fromModule]; mod != nil {
		if pkg := mod.Packages[importPath]; pkg != nil {
			return pkg, nil
		}
	}

	// Try to resolve in other loaded modules
	for _, mod := range s.Modules {
		if pkg := mod.Packages[importPath]; pkg != nil {
			return pkg, nil
		}
	}

	// Not found in any loaded module
	return nil, fmt.Errorf("package %s not found in any loaded module", importPath)
}

// AvailableModules returns the paths of all available modules.
// This implements the typesys.ModuleResolver interface.
func (s *Service) AvailableModules() []string {
	modules := make([]string, 0, len(s.Modules))
	for path := range s.Modules {
		modules = append(modules, path)
	}
	return modules
}

// ResolveTypeAcrossModules resolves a type across all available modules.
// This implements the typesys.ModuleResolver interface.
func (s *Service) ResolveTypeAcrossModules(name string) (types.Type, *typesys.Module, error) {
	// First try to resolve in the main module
	mainModule := s.GetMainModule()
	if mainModule != nil {
		if typ, err := mainModule.ResolveType(name); err == nil {
			return typ, mainModule, nil
		}
	}

	// If not found, try other modules
	for modPath, mod := range s.Modules {
		if modPath == s.MainModulePath {
			continue // Skip main module, we already checked it
		}

		if typ, err := mod.ResolveType(name); err == nil {
			return typ, mod, nil
		}
	}

	return nil, nil, fmt.Errorf("type %s not found in any module", name)
}

// ResolvePackage resolves a package by import path and preferred version
func (s *Service) ResolvePackage(importPath string, preferredVersion string) (*ModulePackage, error) {
	// Check if we have versioned packages for this import path
	versionMap, ok := s.PackageVersions[importPath]
	if !ok {
		// Not found in version map, try to resolve in any module
		for _, mod := range s.Modules {
			if pkg := mod.Packages[importPath]; pkg != nil {
				// Create a ModulePackage entry
				modPkg := &ModulePackage{
					Module:     mod,
					Package:    pkg,
					ImportPath: importPath,
					// We don't know the version, leave it empty for now
					// This will be filled in when we implement dependency loading
				}

				// We found it but without version information
				return modPkg, nil
			}
		}

		return nil, fmt.Errorf("package %s not found in any module", importPath)
	}

	// If we have a preferred version, try that first
	if preferredVersion != "" {
		if modPkg, ok := versionMap[preferredVersion]; ok {
			return modPkg, nil
		}
	}

	// Otherwise just return the first available version
	// In future, we could implement more sophisticated selection logic
	for _, modPkg := range versionMap {
		return modPkg, nil
	}

	return nil, fmt.Errorf("package %s not found with any version", importPath)
}

// ResolveSymbol resolves a symbol by import path, name, and version
func (s *Service) ResolveSymbol(importPath string, name string, version string) ([]*typesys.Symbol, error) {
	// First resolve the package
	modPkg, err := s.ResolvePackage(importPath, version)
	if err != nil {
		return nil, err
	}

	// Now find symbols in that package
	pkg := modPkg.Package
	symbols := pkg.SymbolByName(name)

	return symbols, nil
}

// FindTypeAcrossModules finds a type by import path and name across all modules
func (s *Service) FindTypeAcrossModules(importPath string, typeName string) map[string]*typesys.Symbol {
	result := make(map[string]*typesys.Symbol)

	// Check for the type in each module
	for modPath, mod := range s.Modules {
		if pkg := mod.Packages[importPath]; pkg != nil {
			// Find symbols by name matching the type name
			symbols := pkg.SymbolByName(typeName, typesys.KindType, typesys.KindStruct, typesys.KindInterface)

			// If found, add it to the result map with the module path as key
			if len(symbols) > 0 {
				result[modPath] = symbols[0]
			}
		}
	}

	return result
}

// loadDependencies loads dependencies for all modules using the DependencyManager
func (s *Service) loadDependencies() error {
	return s.DependencyManager.LoadDependencies()
}

// isPackageLoaded checks if a package is already loaded
func (s *Service) isPackageLoaded(importPath string) bool {
	for _, mod := range s.Modules {
		if _, ok := mod.Packages[importPath]; ok {
			return true
		}
	}
	return false
}

// recordPackageVersions records version information for packages in a module
func (s *Service) recordPackageVersions(module *typesys.Module, version string) {
	for importPath, pkg := range module.Packages {
		// Initialize map if needed
		if _, ok := s.PackageVersions[importPath]; !ok {
			s.PackageVersions[importPath] = make(map[string]*ModulePackage)
		}

		// Create ModulePackage entry
		modPkg := &ModulePackage{
			Module:     module,
			Package:    pkg,
			ImportPath: importPath,
			Version:    version,
		}

		// Record the version
		s.PackageVersions[importPath][version] = modPkg
	}
}
