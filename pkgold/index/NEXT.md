# Implementation Plan: Improving Symbol Reference Detection

## Current Limitations

The current implementation of symbol reference detection in the indexing system has several limitations:

1. **Limited AST Traversal**: Our current approach only looks at identifiers and selector expressions, missing references in complex expressions, type assertions, and other contexts.

2. **No Type Resolution**: We don't properly resolve which symbol a name refers to when multiple symbols have the same name in different packages or scopes.

3. **No Scope Awareness**: The system cannot differentiate between new declarations and references to existing symbols.

4. **No Import Resolution**: The system doesn't properly resolve imported packages and their aliases.

5. **No Pointer/Value Distinction**: We don't reliably track whether a method is invoked on a pointer or value receiver.

## Proposed Solution: Integration with Go's Type Checking System

To address these limitations, we need to integrate our indexing system with Go's type checking package (`golang.org/x/tools/go/types`). This will provide:

- Precise symbol resolution across packages
- Correct scope handling
- Proper import resolution
- Exact type information

## Implementation Plan

### Phase 1: Setup Type Checking Integration

1. **Add new dependencies**:
   - `golang.org/x/tools/go/packages` for loading Go packages with type information
   - `golang.org/x/tools/go/types/typeutil` for utilities to work with types

2. **Create a new indexer implementation** that uses the type checking system:
   - Create `pkg/index/typeindexer.go` to hold the type-aware indexer
   - Implement a `TypeAwareIndexer` struct that extends the current `Indexer`

3. **Implement package loading with type information**:
   - Use the `packages.Load` function instead of our custom loader
   - Configure type checking options to analyze dependencies as well

### Phase 2: Symbol Collection with Type Information

1. **Collect definitions with full type information**:
   - Extract symbols from the type-checked AST
   - Store type information along with symbols
   - Map Go's type objects to our symbols for later reference

2. **Improve symbol representation**:
   - Add type information to the `Symbol` struct
   - Add scope information to track where symbols are valid
   - Add fields to store the Go type system's object references

3. **Handle type-specific cases**:
   - Methods on interfaces
   - Type embedding
   - Type aliases and named types
   - Generic types and instantiations

### Phase 3: Reference Detection

1. **Implement a type-aware visitor**:
   - Create a new AST visitor that uses type information
   - Track the current scope during traversal

2. **Resolve references using the type system**:
   - For each identifier, use `types.Info.Uses` to find what it refers to
   - For selector expressions, use `types.Info.Selections` to analyze field/method references
   - For type assertions and conversions, extract the referenced types

3. **Handle special cases**:
   - References to embedded fields and methods
   - References through type aliases
   - References through interfaces
   - References through imports with aliases

### Phase 4: Test and Optimize

1. **Create comprehensive test suite**:
   - Test edge cases like shadowing, package aliases, generics
   - Test with large, real-world codebases
   - Update TestFindReferences to verify accuracy

2. **Performance optimization**:
   - Add caching for parsed and type-checked packages
   - Add incremental update capability
   - Optimize memory usage for large codebases

3. **Integrate with CLI**:
   - Update the find commands to use the new type-aware indexer
   - Add new flags for controlling type checking behavior

## Detailed Implementation Guide

### Type-Aware Indexer Structure

```go
// TypeAwareIndexer builds an index using Go's type checking system
type TypeAwareIndexer struct {
    Index        *Index
    PackageCache map[string]*packages.Package
    TypesInfo    map[*ast.File]*types.Info
    ObjectToSym  map[types.Object]*Symbol
}
```

### Loading Packages with Type Information

```go
func loadPackagesWithTypes(dir string) ([]*packages.Package, error) {
    cfg := &packages.Config{
        Mode: packages.NeedName | 
              packages.NeedFiles | 
              packages.NeedCompiledGoFiles |
              packages.NeedImports |
              packages.NeedTypes | 
              packages.NeedTypesSizes |
              packages.NeedSyntax | 
              packages.NeedTypesInfo |
              packages.NeedDeps,
        Dir:  dir,
        Tests: true,
    }
    
    pkgs, err := packages.Load(cfg, "./...")
    if err != nil {
        return nil, fmt.Errorf("failed to load packages: %w", err)
    }
    
    return pkgs, nil
}
```

### Reference Resolution with Type Checking

```go
func (i *TypeAwareIndexer) findReferences() error {
    // For each file in each package
    for _, pkg := range i.PackageCache {
        for _, file := range pkg.Syntax {
            info := pkg.TypesInfo
            
            // Find all identifier uses
            ast.Inspect(file, func(n ast.Node) bool {
                switch node := n.(type) {
                case *ast.Ident:
                    // Skip identifiers that are part of declarations
                    if obj := info.Defs[node]; obj != nil {
                        return true
                    }
                    
                    // Find what this identifier refers to
                    if obj := info.Uses[node]; obj != nil {
                        // Get our symbol for this object
                        if sym, ok := i.ObjectToSym[obj]; ok {
                            // Create a reference
                            ref := &Reference{
                                TargetSymbol: sym,
                                File:         pkg.GoFiles[0], // Simplified
                                Pos:          node.Pos(),
                                End:          node.End(),
                            }
                            
                            // Add to index
                            i.Index.AddReference(sym, ref)
                        }
                    }
                }
                return true
            })
        }
    }
    return nil
}
```

## Timeline and Milestones

1. **Week 1**: Setup type checking integration and test with simple cases
   - Complete Phase 1
   - Begin Phase 2 implementation

2. **Week 2**: Complete symbol collection with type information
   - Finish Phase 2
   - Test symbol collection on sample codebases

3. **Week 3**: Implement reference detection
   - Complete Phase 3
   - Basic test cases for reference detection

4. **Week 4**: Comprehensive testing and optimization
   - Complete Phase 4
   - Full test suite
   - Performance optimization
   - CLI integration

## Potential Challenges and Solutions

1. **Performance**: Type checking can be resource-intensive for large codebases.
   - Solution: Implement caching and incremental updates
   - Consider parsing but not type-checking certain files (like tests) when not needed

2. **Handling vendored dependencies**: Type checking may require access to dependencies.
   - Solution: Add support for vendor directories and module proxies

3. **Generics complexity**: Go 1.18+ generics add complexity to type resolution.
   - Solution: Add specific handling for generic types and their instantiations

4. **Import cycles**: These can cause issues with the type checker.
   - Solution: Add special handling for import cycles with fallback to AST-only analysis

## Conclusion

By integrating Go's type checking system, we will significantly improve the accuracy and completeness of reference detection in Go-Tree. This will turn it into a powerful tool for code analysis, refactoring, and navigation.

The implementation will require careful attention to Go's type system details, but the result will be a robust indexing system that can reliably find all usages of any symbol in a Go codebase. 