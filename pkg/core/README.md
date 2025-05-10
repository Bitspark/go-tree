# Core Package

The `core` package provides the fundamental building blocks and data structures that underpin the entire Go-Tree system. These components serve as the foundation upon which all other functionality is built.

## Contents

### [`graph`](./graph)
A generic graph data structure implementation that supports directed graphs with arbitrary node and edge data. This package provides:
- A flexible `DirectedGraph` type for representing code structures
- Graph traversal algorithms optimized for code analysis
- Path finding and cycle detection capabilities

### [`typesys`](./typesys)
The type system foundation that represents Go code structure and types. This package provides:
- Rich representation of Go types, symbols, packages, and modules
- Strong type checking capabilities
- Type-aware code manipulation primitives
- Support for generics, interfaces, and other advanced Go features

### [`index`](./index)
Fast indexing capabilities for efficient code navigation and lookup. This package provides:
- Symbol indexing for quick lookups by name or location
- Multi-module symbol resolution
- Optimized data structures for querying code elements

## Architecture

The core packages form the base layer of Go-Tree's architecture. They have minimal external dependencies but are heavily used by higher-level packages. The design follows these principles:

1. **Stability**: Core components change rarely and provide stable APIs
2. **Performance**: Core data structures are optimized for performance
3. **Generality**: Components are designed to be broadly applicable
4. **Type Safety**: All components leverage Go's type system for correctness

## Dependency Structure

```
typesys → graph (for dependency graphs)
index → typesys (for indexing type symbols)
```

Other packages in the codebase build upon these core components but core packages do not depend on higher-level packages. 