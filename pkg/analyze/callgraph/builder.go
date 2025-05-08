package callgraph

import (
	"fmt"

	"bitspark.dev/go-tree/pkg/analyze"
	"bitspark.dev/go-tree/pkg/typesys"
)

// BuildOptions provides options for building the call graph.
type BuildOptions struct {
	// IncludeStdLib determines whether to include standard library calls.
	IncludeStdLib bool

	// IncludeDynamic determines whether to include interface method calls.
	IncludeDynamic bool

	// IncludeImplicit determines whether to include implicit calls (like defer).
	IncludeImplicit bool

	// ExcludePackages is a list of package import paths to exclude from the graph.
	ExcludePackages []string
}

// DefaultBuildOptions returns the default build options.
func DefaultBuildOptions() *BuildOptions {
	return &BuildOptions{
		IncludeStdLib:   false,
		IncludeDynamic:  true,
		IncludeImplicit: true,
		ExcludePackages: nil,
	}
}

// CallGraphBuilder builds a call graph for a module.
type CallGraphBuilder struct {
	*analyze.BaseAnalyzer
	Module *typesys.Module
}

// NewCallGraphBuilder creates a new call graph builder.
func NewCallGraphBuilder(module *typesys.Module) *CallGraphBuilder {
	return &CallGraphBuilder{
		BaseAnalyzer: analyze.NewBaseAnalyzer(
			"CallGraphBuilder",
			"Builds a call graph from a module",
		),
		Module: module,
	}
}

// Build builds a call graph for the module.
func (b *CallGraphBuilder) Build(opts *BuildOptions) (*CallGraph, error) {
	if b.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	if opts == nil {
		opts = DefaultBuildOptions()
	}

	// Create a new call graph
	graph := NewCallGraph(b.Module)

	// Find all callable symbols (functions and methods)
	callables := b.findCallableSymbols(opts)

	// Add all callable symbols to the graph
	for _, sym := range callables {
		graph.AddNode(sym)
	}

	// Process each callable to find its calls
	for _, caller := range callables {
		b.processCallable(graph, caller, opts)
	}

	return graph, nil
}

// BuildResult represents the result of a call graph build operation.
type BuildResult struct {
	*analyze.BaseResult
	Graph *CallGraph
}

// GetGraph returns the call graph from the result.
func (r *BuildResult) GetGraph() *CallGraph {
	return r.Graph
}

// NewBuildResult creates a new build result.
func NewBuildResult(builder *CallGraphBuilder, graph *CallGraph, err error) *BuildResult {
	return &BuildResult{
		BaseResult: analyze.NewBaseResult(builder, err),
		Graph:      graph,
	}
}

// BuildAsync builds a call graph asynchronously and returns a result channel.
func (b *CallGraphBuilder) BuildAsync(opts *BuildOptions) <-chan *BuildResult {
	resultCh := make(chan *BuildResult, 1)

	go func() {
		graph, err := b.Build(opts)
		resultCh <- NewBuildResult(b, graph, err)
		close(resultCh)
	}()

	return resultCh
}

// findCallableSymbols finds all functions and methods in the module.
func (b *CallGraphBuilder) findCallableSymbols(opts *BuildOptions) []*typesys.Symbol {
	var callables []*typesys.Symbol

	// Filter function to determine if a symbol should be included
	shouldInclude := func(sym *typesys.Symbol) bool {
		// Check if it's a function or method
		if !isCallable(sym) {
			return false
		}

		// Check package exclusions
		if b.isExcludedPackage(sym, opts.ExcludePackages) {
			return false
		}

		// Check standard library exclusion
		if !opts.IncludeStdLib && b.isStdLibPackage(sym) {
			return false
		}

		return true
	}

	// Traverse all packages in the module
	for _, pkg := range b.Module.Packages {
		// Add all functions and methods from this package
		for _, sym := range pkg.Symbols {
			if shouldInclude(sym) {
				callables = append(callables, sym)
			}
		}
	}

	return callables
}

// processCallable processes a callable symbol to find its calls.
func (b *CallGraphBuilder) processCallable(graph *CallGraph, caller *typesys.Symbol, opts *BuildOptions) {
	// Get references made by this function
	for _, ref := range caller.References {
		// Process each reference based on its kind
		b.processReference(graph, caller, ref, opts)
	}
}

// processReference processes a reference to determine if it's a call.
func (b *CallGraphBuilder) processReference(graph *CallGraph, caller *typesys.Symbol, ref *typesys.Reference, opts *BuildOptions) {
	// Skip if not a function call reference
	if !isCallReference(ref) {
		return
	}

	// Get the target function symbol
	target := ref.Symbol
	if target == nil {
		return
	}

	// Skip if the target is excluded
	if b.isExcludedPackage(target, opts.ExcludePackages) {
		return
	}

	// Skip standard library calls if not included
	if !opts.IncludeStdLib && b.isStdLibPackage(target) {
		return
	}

	// Skip dynamic calls if not included
	isDynamic := isInterfaceMethodCall(ref)
	if !opts.IncludeDynamic && isDynamic {
		return
	}

	// Get position information
	pos := ref.GetPosition()
	line := 0
	column := 0
	if pos != nil {
		line = pos.LineStart
		column = pos.ColumnStart
	}

	// Create a call site
	site := &CallSite{
		File:    ref.File,
		Line:    line,
		Column:  column,
		Context: caller,
	}

	// Add the call to the graph
	graph.AddCall(caller, target, site, isDynamic)
}

// Helper functions

// isExcludedPackage checks if a symbol is from an excluded package.
func (b *CallGraphBuilder) isExcludedPackage(sym *typesys.Symbol, excludePackages []string) bool {
	if sym == nil || sym.Package == nil || len(excludePackages) == 0 {
		return false
	}

	pkg := sym.Package.ImportPath
	for _, excluded := range excludePackages {
		if pkg == excluded {
			return true
		}
	}

	return false
}

// isStdLibPackage checks if a symbol is from the standard library.
func (b *CallGraphBuilder) isStdLibPackage(sym *typesys.Symbol) bool {
	if sym == nil || sym.Package == nil {
		return false
	}

	// In Go, standard library packages don't have a dot in their import path
	// This is a simple heuristic - a more sophisticated implementation would use build.IsStandardPackage
	pkg := sym.Package.ImportPath
	for i := 0; i < len(pkg); i++ {
		if pkg[i] == '.' {
			return false
		}
	}

	return true
}

// isCallReference checks if a reference is a function call.
func isCallReference(ref *typesys.Reference) bool {
	// Check if the reference is to a callable symbol and it's not a write operation
	return ref != nil && ref.Symbol != nil &&
		isCallable(ref.Symbol) && !ref.IsWrite
}

// isInterfaceMethodCall checks if a reference is a call to an interface method.
func isInterfaceMethodCall(ref *typesys.Reference) bool {
	// Check if the target symbol is a method on an interface
	// In a real implementation, we would check more thoroughly
	return ref != nil && ref.Symbol != nil &&
		ref.Symbol.Kind == typesys.KindMethod &&
		ref.Symbol.Parent != nil &&
		ref.Symbol.Parent.Kind == typesys.KindInterface
}
