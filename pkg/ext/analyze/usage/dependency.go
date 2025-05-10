package usage

import (
	"bitspark.dev/go-tree/pkg/core/graph"
	"bitspark.dev/go-tree/pkg/ext/analyze"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// DependencyNode represents a node in the dependency graph.
type DependencyNode struct {
	// Symbol this node represents
	Symbol *typesys.Symbol

	// Dependencies outgoing from this symbol
	Dependencies []*DependencyEdge

	// Dependents incoming to this symbol
	Dependents []*DependencyEdge
}

// DependencyEdge represents a dependency between two symbols.
type DependencyEdge struct {
	// Source symbol that depends on Target
	From *DependencyNode

	// Target symbol that is depended on by Source
	To *DependencyNode

	// Strength of the dependency (number of references)
	Strength int

	// Types of references in this dependency
	ReferenceTypes map[ReferenceKind]int
}

// DependencyGraph represents a symbol dependency graph.
type DependencyGraph struct {
	// Nodes in the graph, indexed by symbol ID
	Nodes map[string]*DependencyNode

	// The underlying directed graph
	graph *graph.DirectedGraph
}

// NewDependencyGraph creates a new empty dependency graph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		graph: graph.NewDirectedGraph(),
	}
}

// AddNode adds a symbol node to the dependency graph.
func (g *DependencyGraph) AddNode(sym *typesys.Symbol) *DependencyNode {
	// Skip if not a valid symbol
	if sym == nil {
		return nil
	}

	// Generate ID
	id := getSymbolID(sym)

	// Check if the node already exists
	if node, exists := g.Nodes[id]; exists {
		return node
	}

	// Create a new node
	node := &DependencyNode{
		Symbol:       sym,
		Dependencies: make([]*DependencyEdge, 0),
		Dependents:   make([]*DependencyEdge, 0),
	}

	// Add to the graph
	g.graph.AddNode(id, node)
	g.Nodes[id] = node

	return node
}

// AddDependency adds a dependency edge between two symbols.
func (g *DependencyGraph) AddDependency(from, to *typesys.Symbol, kind ReferenceKind) *DependencyEdge {
	// Ensure both nodes exist
	fromNode := g.GetOrCreateNode(from)
	toNode := g.GetOrCreateNode(to)

	if fromNode == nil || toNode == nil {
		return nil
	}

	// Check if the edge already exists
	for _, edge := range fromNode.Dependencies {
		if edge.To.Symbol.ID == toNode.Symbol.ID {
			// Update existing edge
			edge.Strength++
			edge.ReferenceTypes[kind]++
			return edge
		}
	}

	// Create the graph edge - we don't need to store it
	g.graph.AddEdge(getSymbolID(from), getSymbolID(to), nil)

	// Create a new dependency edge
	edge := &DependencyEdge{
		From:           fromNode,
		To:             toNode,
		Strength:       1,
		ReferenceTypes: make(map[ReferenceKind]int),
	}

	// Set the reference type
	edge.ReferenceTypes[kind] = 1

	// Update the node references
	fromNode.Dependencies = append(fromNode.Dependencies, edge)
	toNode.Dependents = append(toNode.Dependents, edge)

	return edge
}

// GetNode gets a node by its symbol.
func (g *DependencyGraph) GetNode(sym *typesys.Symbol) *DependencyNode {
	if sym == nil {
		return nil
	}
	return g.Nodes[getSymbolID(sym)]
}

// GetOrCreateNode gets a node or creates it if it doesn't exist.
func (g *DependencyGraph) GetOrCreateNode(sym *typesys.Symbol) *DependencyNode {
	if sym == nil {
		return nil
	}

	node := g.GetNode(sym)
	if node == nil {
		node = g.AddNode(sym)
	}
	return node
}

// FindCycles finds all dependency cycles in the graph.
func (g *DependencyGraph) FindCycles() [][]*DependencyEdge {
	var cycles [][]*DependencyEdge

	// Check each node for cycles starting from it
	for _, node := range g.Nodes {
		visited := make(map[string]bool)
		stack := make(map[string]bool)
		path := make([]*DependencyEdge, 0)

		g.findCyclesDFS(node, visited, stack, path, &cycles)
	}

	return cycles
}

// findCyclesDFS uses DFS to find cycles in the graph.
func (g *DependencyGraph) findCyclesDFS(node *DependencyNode,
	visited, stack map[string]bool,
	path []*DependencyEdge, cycles *[][]*DependencyEdge) {

	nodeID := getSymbolID(node.Symbol)

	// Skip if already fully explored
	if visited[nodeID] {
		return
	}

	// Check if we've found a cycle
	if stack[nodeID] {
		// We need to extract the cycle from the path
		for i, edge := range path {
			fromID := getSymbolID(edge.From.Symbol)
			if fromID == nodeID {
				// Found the start of the cycle
				cyclePath := make([]*DependencyEdge, len(path)-i)
				copy(cyclePath, path[i:])
				*cycles = append(*cycles, cyclePath)
				break
			}
		}
		return
	}

	// Mark as in-progress
	stack[nodeID] = true

	// Explore outgoing edges
	for _, edge := range node.Dependencies {
		path = append(path, edge)
		g.findCyclesDFS(edge.To, visited, stack, path, cycles)
		path = path[:len(path)-1]
	}

	// Mark as fully explored
	visited[nodeID] = true
	stack[nodeID] = false
}

// MostDepended returns the symbols with the most dependents.
func (g *DependencyGraph) MostDepended(limit int) []*DependencyNode {
	// Create a slice of all nodes
	nodes := make([]*DependencyNode, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		nodes = append(nodes, node)
	}

	// Sort by number of dependents (descending)
	sortNodesByDependentCount(nodes)

	// Limit results
	if limit > 0 && limit < len(nodes) {
		nodes = nodes[:limit]
	}

	return nodes
}

// MostDependent returns the symbols with the most dependencies.
func (g *DependencyGraph) MostDependent(limit int) []*DependencyNode {
	// Create a slice of all nodes
	nodes := make([]*DependencyNode, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		nodes = append(nodes, node)
	}

	// Sort by number of dependencies (descending)
	sortNodesByDependencyCount(nodes)

	// Limit results
	if limit > 0 && limit < len(nodes) {
		nodes = nodes[:limit]
	}

	return nodes
}

// DependencyAnalyzer analyzes symbol dependencies.
type DependencyAnalyzer struct {
	*analyze.BaseAnalyzer
	Module    *typesys.Module
	Collector *UsageCollector
}

// NewDependencyAnalyzer creates a new dependency analyzer.
func NewDependencyAnalyzer(module *typesys.Module) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		BaseAnalyzer: analyze.NewBaseAnalyzer(
			"DependencyAnalyzer",
			"Analyzes symbol dependencies",
		),
		Module:    module,
		Collector: NewUsageCollector(module),
	}
}

// AnalyzeDependencies creates a dependency graph for the given module.
func (a *DependencyAnalyzer) AnalyzeDependencies() (*DependencyGraph, error) {
	if a.Module == nil {
		return nil, fmt.Errorf("module is nil")
	}

	// Collect usage information for all symbols
	usages, err := a.Collector.CollectUsageForAllSymbols()
	if err != nil {
		return nil, err
	}

	// Create a dependency graph
	graph := NewDependencyGraph()

	// Process each package
	for _, pkg := range a.Module.Packages {
		// Process each symbol in the package
		for _, sym := range pkg.Symbols {
			// Get symbol usage
			usage, found := usages[getSymbolID(sym)]
			if !found {
				continue
			}

			// Process each reference to build dependencies
			for kind, refs := range usage.References {
				for _, ref := range refs {
					if ref.Symbol != nil && ref.Symbol != sym {
						graph.AddDependency(sym, ref.Symbol, ReferenceKind(kind))
					}
				}
			}
		}
	}

	return graph, nil
}

// AnalyzePackageDependencies analyzes dependencies between packages.
func (a *DependencyAnalyzer) AnalyzePackageDependencies() (*DependencyGraph, error) {
	// This would build a higher-level graph showing package-level dependencies
	// by aggregating symbol dependencies.
	return nil, fmt.Errorf("not implemented")
}

// Helper functions

// sortNodesByDependentCount sorts nodes by their dependent count (descending).
func sortNodesByDependentCount(nodes []*DependencyNode) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if len(nodes[j].Dependents) > len(nodes[i].Dependents) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// sortNodesByDependencyCount sorts nodes by their dependency count (descending).
func sortNodesByDependencyCount(nodes []*DependencyNode) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if len(nodes[j].Dependencies) > len(nodes[i].Dependencies) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
