package typesys

import (
	"go/types"
)

// ModuleResolver provides cross-module resolution capabilities
type ModuleResolver interface {
	// GetModule returns a module by path
	GetModule(path string) *Module

	// AvailableModules returns the paths of all available modules
	AvailableModules() []string

	// ResolveTypeAcrossModules resolves a type across all available modules
	ResolveTypeAcrossModules(name string) (types.Type, *Module, error)
}

// ModuleResolverFunc is a helper that allows normal functions to implement ModuleResolver
type ModuleResolverFunc func(path string) *Module

// GetModule implements ModuleResolver for ModuleResolverFunc
func (f ModuleResolverFunc) GetModule(path string) *Module {
	return f(path)
}

// AvailableModules returns an empty slice for ModuleResolverFunc
// This should be implemented properly by actual implementations
func (f ModuleResolverFunc) AvailableModules() []string {
	return nil
}

// ResolveTypeAcrossModules returns nil for ModuleResolverFunc
// This should be implemented properly by actual implementations
func (f ModuleResolverFunc) ResolveTypeAcrossModules(name string) (types.Type, *Module, error) {
	return nil, nil, nil
}
