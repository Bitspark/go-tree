# IO Package

The `io` package contains components that handle the input, output, and resolution of Go code. These packages provide the bridge between the filesystem and the in-memory representations used by the Go-Tree system.

## Contents

### [`loader`](./loader)
Handles loading Go code from the filesystem into memory. This package provides:
- Module loading capabilities
- Fast and correct parsing of Go source files
- Type-aware code loading
- Integration with the Go module system

### [`saver`](./saver)
Manages saving in-memory code representations back to the filesystem. This package provides:
- Code generation and persistence
- Formatting and serialization of Go source
- Preserving comments and formatting during saves

### [`resolve`](./resolve)
Resolves module dependencies and handles version management. This package provides:
- Module resolution based on import paths
- Dependency resolution with version constraints
- Integration with the Go module system
- Handling of module replacement directives

### [`materialize`](./materialize)
Materializes resolved modules onto the filesystem for execution. This package provides:
- Creation of temporary module structures for execution
- Preparation of dependencies for compilation
- Environment setup for running Go code

## Architecture

The IO packages form the interface layer between Go-Tree and the filesystem. They have these key characteristics:

1. **Bidirectional**: They handle both reading from and writing to the filesystem
2. **Module-Aware**: All components understand Go modules and their semantics
3. **Version-Aware**: Components handle module versioning
4. **Resolution**: They resolve references and dependencies across modules

## Dependency Structure

```
loader → (core/typesys)
saver → (core/typesys)
resolve → (core/typesys, core/graph)
materialize → (resolve, core/typesys)
```

The IO packages primarily depend on core packages and provide services to higher-level packages like `run` and `ext`. 