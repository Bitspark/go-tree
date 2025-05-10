# Ext Package

The `ext` package contains extension components that provide advanced analysis, transformation, and visualization capabilities for Go code. These packages extend the core functionality with higher-level features.

## Contents

### [`analyze`](./analyze)
Provides static code analysis capabilities. This package includes:
- Call graph generation and analysis
- Interface implementation detection
- Usage analysis for symbols
- Type hierarchy analysis
- Code complexity metrics

### [`transform`](./transform)
Enables code transformation with type safety. This package provides:
- Refactoring tools (rename, extract, inline)
- Code generation capabilities
- AST transformations
- Type-preserving code changes

### [`visual`](./visual)
Visualization components for code structures. This package includes:
- Graph visualization for dependencies
- Type hierarchy visualization
- Call graph visualization
- Interactive code structure diagrams

## Architecture

The Ext packages build on the core and IO layers to provide higher-level functionality. They are characterized by:

1. **Analysis**: Deep code analysis capabilities
2. **Transformation**: Type-safe code modifications
3. **Visualization**: Representing code structures visually
4. **Extension**: Providing extension points for domain-specific features

## Dependency Structure

```
analyze → (core/typesys, core/graph, core/index)
transform → (core/typesys, analyze)
visual → (core/graph, analyze)
```

The Ext packages represent the analytical and transformational layer of Go-Tree, sitting between the foundational layers (Core, IO) and the application layer (Service).

## Usage Patterns

Ext components are typically used after code has been loaded via the IO packages but before it is executed by the Run packages. They enable understanding, modifying, and visualizing code before execution or deployment.

## Extension Points

Each package provides extension points for custom analyzers, transformers, and visualizers:

- Analyze: Custom analyzer interfaces
- Transform: Transformation framework
- Visual: Pluggable visualization formats 