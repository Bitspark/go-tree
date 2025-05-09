# Run Package

The `run` package contains components for executing Go code, running tests, and providing toolkit utilities. These packages enable the dynamic evaluation and interaction with Go code at runtime.

## Contents

### [`execute`](./execute)
Provides execution capabilities for Go code. This package includes:
- Module execution with type awareness
- Dynamic code evaluation
- Support for executing Go functions with proper type checking
- Sandboxed execution environments

### [`testing`](./testing)
Extends the standard Go testing framework with enhanced capabilities. This package provides:
- Advanced test discovery and execution
- Test result analysis
- Type-aware testing utilities
- Coverage analysis tools

### [`toolkit`](./toolkit)
General-purpose utilities and tools for working with Go code. This package includes:
- File system abstractions
- Common middleware components
- Standard utilities for code manipulation
- Developer-friendly helpers

## Architecture

The Run packages provide runtime capabilities that build on the core and IO layers. They are characterized by:

1. **Dynamic Operation**: Components operate at runtime rather than static analysis
2. **Execution Context**: They establish and maintain execution contexts
3. **Type Safety**: All execution maintains type safety via the type system
4. **Test Support**: First-class support for testing and validation

## Dependency Structure

```
execute → (io/materialize, core/typesys)
testing → (execute, core/typesys)
toolkit → (minimal dependencies)
```

The Run packages form the execution layer of the Go-Tree system, enabling dynamic interaction with the code that has been loaded and analyzed by the lower layers.

## Usage Patterns

Run components are typically used after code has been loaded via the IO packages and potentially analyzed or transformed by the Ext packages. They represent the final stage in many workflows, where code is actually executed rather than just analyzed. 