# Service Package

The `service` package provides a unified interface to the entire Go-Tree system, integrating all the individual components into a cohesive whole. It serves as the primary entry point for applications using Go-Tree.

## Contents

The service package includes:

- **Multi-module management**: Handling of multiple Go modules with their interrelationships
- **Version-aware symbol resolution**: Finding symbols across modules with version awareness
- **Unified API**: A cohesive API that combines functionality from all other packages
- **Configuration management**: Centralized configuration for all Go-Tree components
- **Service lifecycle**: Initialization, operation, and shutdown of Go-Tree services

## Key Interfaces

### `Service`
The main service interface that provides access to all Go-Tree capabilities:
- Module management
- Symbol resolution
- Type checking
- Code execution
- Analysis and transformation

### `Config`
Configuration for the service with options for:
- Module handling
- Dependency resolution
- Analysis depth
- Execution environments

## API Reference

### Service Initialization
```go
// Creates a new service instance with the specified configuration
NewService(config *Config) (*Service, error)
```

### Module Management
```go
// Retrieves a module by its path
GetModule(modulePath string) *typesys.Module

// Gets the main module that was loaded
GetMainModule() *typesys.Module

// Returns the paths of all available modules
AvailableModules() []string
```

### Symbol Resolution
```go
// Finds symbols by name across all loaded modules
FindSymbolsAcrossModules(name string) ([]*typesys.Symbol, error)

// Finds symbols by name in a specific module
FindSymbolsIn(modulePath string, name string) ([]*typesys.Symbol, error)

// Resolves a symbol by import path, name, and version
ResolveSymbol(importPath string, name string, version string) ([]*typesys.Symbol, error)

// Finds a type by import path and name across all modules
FindTypeAcrossModules(importPath string, typeName string) map[string]*typesys.Symbol
```

### Package Management
```go
// Resolves an import path to a package, checking in the source module first
ResolveImport(importPath string, fromModule string) (*typesys.Package, error)

// Resolves a package by import path and preferred version
ResolvePackage(importPath string, preferredVersion string) (*ModulePackage, error)

// Resolves a type across all available modules
ResolveTypeAcrossModules(name string) (types.Type, *typesys.Module, error)
```

### Dependency Management
```go
// Adds a dependency to a module
AddDependency(module *typesys.Module, importPath, version string) error

// Removes a dependency from a module
RemoveDependency(module *typesys.Module, importPath string) error
```

### Environment Management
```go
// Creates an execution environment for modules
CreateEnvironment(modules []*typesys.Module, opts *Config) (*materialize.Environment, error)
```

## Architecture

The Service package sits at the top of the Go-Tree architecture, depending on all other packages but providing a simplified, unified interface to them. It is characterized by:

1. **Integration**: Bringing together all components into a cohesive whole
2. **Simplification**: Providing simpler interfaces to complex underlying functionality
3. **Configuration**: Centralizing configuration for all components
4. **Lifecycle**: Managing the lifecycle of all dependent components

## Dependency Structure

```
service → (core/*, io/*, run/*, ext/*)
```

The Service package depends on all other packages but is not depended upon by any of them, forming the top of the dependency hierarchy.

## Usage

The Service package is the primary entry point for applications using Go-Tree:

```go
config := &service.Config{
    ModuleDir: "/path/to/module",
    IncludeTests: true,
    WithDeps: true,
}

svc, err := service.NewService(config)
if err != nil {
    // handle error
}

// Use the service to access Go-Tree functionality
module := svc.GetMainModule()
symbols, _ := svc.FindSymbolsAcrossModules("MyType")
```

## Extension

The Service package is designed to be extensible, allowing for domain-specific services to be built on top of it. These extensions can add functionality specific to particular applications or domains while leveraging the underlying Go-Tree infrastructure. 