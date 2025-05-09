# Go-Tree Index Package

The `index` package provides type-aware code indexing capabilities for the Go-Tree tool. It builds on the core type system to offer fast lookup of symbols, references, implementations, and more.

## Key Components

### Index

The core `Index` struct maintains all the indexed data and provides efficient lookup operations.

```go
// Create and build an index
module, _ := typesys.LoadModule("path/to/module", opts)
idx := index.NewIndex(module)
idx.Build()

// Find symbols
symbols := idx.FindSymbolsByName("MyType")
symbolsInFile := idx.FindSymbolsInFile("path/to/file.go")
symbolsByKind := idx.FindSymbolsByKind(typesys.KindInterface)

// Find references
refs := idx.FindReferences(symbol)
refsInFile := idx.FindReferencesInFile("path/to/file.go")

// Find special relationships
methods := idx.FindMethods("MyType")
impls := idx.FindImplementations(interfaceSymbol)
```

### Indexer

The `Indexer` struct wraps the `Index` and provides additional high-level operations.

```go
// Create and build an indexer
opts := index.IndexingOptions{
    IncludeTests:       true,
    IncludePrivate:     true,
    IncrementalUpdates: true,
}
indexer := index.NewIndexer(module, opts)
indexer.BuildIndex()

// Update for changed files
indexer.UpdateIndex([]string{"path/to/changed/file.go"})

// Search and find symbols
results := indexer.Search("MyPattern")
functions := indexer.FindAllFunctions("MyFunc")
types := indexer.FindAllTypes("MyType")

// Get file structure
structure := indexer.GetFileStructure("path/to/file.go")
```

### CommandContext

The `CommandContext` provides a command-line friendly interface for index operations.

```go
// Create a command context
ctx, _ := index.NewCommandContext(module, opts)

// Find symbol usages
ctx.FindUsages("MySymbol", "", 0, 0)
ctx.FindUsages("", "path/to/file.go", 10, 5) // By position

// Find implementations
ctx.FindImplementations("MyInterface")

// Search symbols
ctx.SearchSymbols("My", "type,function")

// List file symbols
ctx.ListFileSymbols("path/to/file.go")
```

## Features

1. **Type-Aware Indexing**: Uses Go's type checking system for accurate analysis
2. **Fast Symbol Lookup**: Find symbols by name, kind, file, or position
3. **Accurate Reference Finding**: Track all usages of symbols with context
4. **Interface Implementation Discovery**: Find all types that implement an interface
5. **Method Resolution**: Find all methods for a type
6. **Incremental Updates**: Efficiently update the index when files change
7. **Structured File View**: Get a structured representation of file contents

## CLI Integration

To integrate with the Go-Tree CLI, use the `CommandContext` in command implementations:

```go
func FindUsagesCommand(c *cli.Context) error {
    // Load the module
    module, err := loadModule(c.String("dir"))
    if err != nil {
        return err
    }
    
    // Create options
    opts := index.IndexingOptions{
        IncludeTests:       c.Bool("tests"),
        IncludePrivate:     c.Bool("private"),
        IncrementalUpdates: true,
    }
    
    // Create command context
    ctx, err := index.NewCommandContext(module, opts)
    if err != nil {
        return err
    }
    
    // Set verbosity
    ctx.Verbose = c.Bool("verbose")
    
    // Execute the command
    return ctx.FindUsages(c.String("name"), c.String("file"), c.Int("line"), c.Int("column"))
}
```

## Using with the Type System

The index package works directly with the type system and depends on its structures:

```go
// Load a module with the type system
module, err := typesys.LoadModule("path/to/module", &typesys.LoadOptions{
    IncludeTests:   true,
    IncludePrivate: true,
})
if err != nil {
    log.Fatalf("Failed to load module: %v", err)
}

// Create and build an index
idx := index.NewIndex(module)
err = idx.Build()
if err != nil {
    log.Fatalf("Failed to build index: %v", err)
}

// Find all interfaces
interfaces := idx.FindSymbolsByKind(typesys.KindInterface)
for _, iface := range interfaces {
    fmt.Printf("Interface: %s\n", iface.Name)
    
    // Find implementations
    impls := idx.FindImplementations(iface)
    for _, impl := range impls {
        fmt.Printf("  Implementation: %s\n", impl.Name)
    }
}
```

## Performance Considerations

1. Initial indexing can be resource-intensive for large codebases
2. Use incremental updates when possible
3. The index maintains in-memory maps which can consume memory
4. Consider filtering out test files and private symbols if not needed

## Future Extensions

Planned extensions to the indexing system include:

1. Fuzzy search capabilities
2. Persistent index storage
3. Background indexing
4. Integration with IDEs via the Language Server Protocol
5. Advanced code navigation features 