package resolve

import (
	"errors"
	"path/filepath"
	"sync"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// ResolvedModule contains all resolution information for a module
type ResolvedModule struct {
	// Import path (e.g., "github.com/user/repo")
	ImportPath string

	// Filesystem path (e.g., "/path/to/module")
	FilesystemPath string

	// Loaded module (may be nil if not loaded yet)
	Module *typesys.Module

	// Version (may be empty for local modules)
	Version string

	// Whether this is a local filesystem module
	IsLocal bool
}

// GetModule returns the module
func (r *ResolvedModule) GetModule() *typesys.Module {
	return r.Module
}

// GetFilesystemPath returns the filesystem path
func (r *ResolvedModule) GetFilesystemPath() string {
	return r.FilesystemPath
}

// GetImportPath returns the import path
func (r *ResolvedModule) GetImportPath() string {
	return r.ImportPath
}

// ModuleRegistry defines a registry that maps import paths to filesystem paths
type ModuleRegistry interface {
	// RegisterModule registers a module by its import path and filesystem location
	RegisterModule(importPath, fsPath string, isLocal bool) error

	// FindModule finds a module by import path
	FindModule(importPath string) (*ResolvedModule, bool)

	// FindByPath finds a module by filesystem path
	FindByPath(fsPath string) (*ResolvedModule, bool)

	// ListModules returns all registered modules
	ListModules() []*ResolvedModule

	// CreateResolver creates a resolver configured with this registry
	CreateResolver() Resolver
}

// ErrModuleAlreadyRegistered is returned when attempting to register a module that's already registered
var ErrModuleAlreadyRegistered = errors.New("module already registered with different path")

// StandardModuleRegistry provides a basic implementation of ModuleRegistry
type StandardModuleRegistry struct {
	modules     map[string]*ResolvedModule // Key: import path
	pathModules map[string]*ResolvedModule // Key: filesystem path
	mu          sync.RWMutex
}

// NewStandardModuleRegistry creates a new standard module registry
func NewStandardModuleRegistry() *StandardModuleRegistry {
	return &StandardModuleRegistry{
		modules:     make(map[string]*ResolvedModule),
		pathModules: make(map[string]*ResolvedModule),
	}
}

// RegisterModule registers a module by its import path and filesystem location
func (r *StandardModuleRegistry) RegisterModule(importPath, fsPath string, isLocal bool) error {
	if importPath == "" || fsPath == "" {
		return errors.New("import path and filesystem path cannot be empty")
	}

	// Normalize paths
	fsPath = filepath.Clean(fsPath)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered with different path
	if existing, ok := r.modules[importPath]; ok {
		if existing.FilesystemPath != fsPath {
			return ErrModuleAlreadyRegistered
		}
		// Already registered with same path, just update
		existing.IsLocal = isLocal
		return nil
	}

	// Create new resolved module
	module := &ResolvedModule{
		ImportPath:     importPath,
		FilesystemPath: fsPath,
		IsLocal:        isLocal,
	}

	// Register by import path and filesystem path
	r.modules[importPath] = module
	r.pathModules[fsPath] = module

	return nil
}

// FindModule finds a module by import path
func (r *StandardModuleRegistry) FindModule(importPath string) (*ResolvedModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.modules[importPath]
	return module, ok
}

// FindByPath finds a module by filesystem path
func (r *StandardModuleRegistry) FindByPath(fsPath string) (*ResolvedModule, bool) {
	// Normalize path
	fsPath = filepath.Clean(fsPath)

	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.pathModules[fsPath]
	return module, ok
}

// ListModules returns all registered modules
func (r *StandardModuleRegistry) ListModules() []*ResolvedModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]*ResolvedModule, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}

	return modules
}

// CreateResolver creates a resolver configured with this registry
func (r *StandardModuleRegistry) CreateResolver() Resolver {
	// Create a new resolver and set this registry
	return NewModuleResolver().WithRegistry(r)
}
