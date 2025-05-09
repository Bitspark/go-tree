// Package callgraph provides functionality for analyzing function call relationships.
package callgraph

import (
	"bitspark.dev/go-tree/pkg/core/graph"
	"fmt"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// CallGraph represents a call graph for a module.
type CallGraph struct {
	// The module this call graph represents
	Module *typesys.Module

	// Nodes in the graph, indexed by symbol ID
	Nodes map[string]*CallNode

	// The underlying directed graph
	graph *graph.DirectedGraph
}

// CallNode represents a function or method in the call graph.
type CallNode struct {
	// The function or method symbol
	Symbol *typesys.Symbol

	// ID of the node (same as symbol ID)
	ID string

	// Outgoing calls from this function
	Calls []*CallEdge

	// Incoming calls to this function
	CalledBy []*CallEdge
}

// CallEdge represents a call from one function to another.
type CallEdge struct {
	// Source function node
	From *CallNode

	// Target function node
	To *CallNode

	// Locations where the call occurs
	Sites []*CallSite

	// Whether this is a dynamic call (interface method call)
	Dynamic bool

	// The underlying graph edge
	edge *graph.Edge
}

// CallSite represents a specific location where a call occurs.
type CallSite struct {
	// The file containing the call
	File *typesys.File

	// Line number (1-based)
	Line int

	// Column number (1-based)
	Column int

	// The function containing this call site
	Context *typesys.Symbol
}

// NewCallGraph creates a new empty call graph for the given module.
func NewCallGraph(module *typesys.Module) *CallGraph {
	return &CallGraph{
		Module: module,
		Nodes:  make(map[string]*CallNode),
		graph:  graph.NewDirectedGraph(),
	}
}

// AddNode adds a function or method node to the call graph.
func (g *CallGraph) AddNode(sym *typesys.Symbol) *CallNode {
	// Skip if not a function or method
	if !isCallable(sym) {
		return nil
	}

	// Generate ID
	id := getSymbolID(sym)

	// Check if the node already exists
	if node, exists := g.Nodes[id]; exists {
		return node
	}

	// Create a new node
	node := &CallNode{
		Symbol:   sym,
		ID:       id,
		Calls:    make([]*CallEdge, 0),
		CalledBy: make([]*CallEdge, 0),
	}

	// Add to the graph
	g.graph.AddNode(id, node)
	g.Nodes[id] = node

	return node
}

// AddCall adds a call edge between two functions.
func (g *CallGraph) AddCall(from, to *typesys.Symbol, site *CallSite, dynamic bool) *CallEdge {
	// Ensure both nodes exist
	fromNode := g.GetOrCreateNode(from)
	toNode := g.GetOrCreateNode(to)

	if fromNode == nil || toNode == nil {
		return nil
	}

	// Check if the edge already exists
	for _, edge := range fromNode.Calls {
		if edge.To.ID == toNode.ID {
			// Add the call site to the existing edge
			if site != nil {
				edge.Sites = append(edge.Sites, site)
			}
			return edge
		}
	}

	// Create the graph edge
	graphEdge := g.graph.AddEdge(fromNode.ID, toNode.ID, nil)

	// Create a new call edge
	edge := &CallEdge{
		From:    fromNode,
		To:      toNode,
		Sites:   make([]*CallSite, 0),
		Dynamic: dynamic,
		edge:    graphEdge,
	}

	// Add the call site if provided
	if site != nil {
		edge.Sites = append(edge.Sites, site)
	}

	// Update the node references
	fromNode.Calls = append(fromNode.Calls, edge)
	toNode.CalledBy = append(toNode.CalledBy, edge)

	return edge
}

// GetNode gets a node by its symbol.
func (g *CallGraph) GetNode(sym *typesys.Symbol) *CallNode {
	if sym == nil {
		return nil
	}
	return g.Nodes[getSymbolID(sym)]
}

// GetOrCreateNode gets a node or creates it if it doesn't exist.
func (g *CallGraph) GetOrCreateNode(sym *typesys.Symbol) *CallNode {
	if sym == nil || !isCallable(sym) {
		return nil
	}

	node := g.GetNode(sym)
	if node == nil {
		node = g.AddNode(sym)
	}
	return node
}

// FindPaths finds all paths between two functions, up to maxLength.
// If maxLength is 0 or negative, there is no length limit.
func (g *CallGraph) FindPaths(from, to *CallNode, maxLength int) [][]*CallEdge {
	if from == nil || to == nil {
		return nil
	}

	var paths [][]*CallEdge
	visited := make(map[string]bool)
	currentPath := make([]*CallEdge, 0)

	// DFS to find all paths
	g.findPathsDFS(from, to, currentPath, visited, &paths, maxLength)

	return paths
}

// findPathsDFS uses depth-first search to find all paths between two nodes.
func (g *CallGraph) findPathsDFS(current, target *CallNode,
	path []*CallEdge, visited map[string]bool,
	allPaths *[][]*CallEdge, maxLength int) {

	// Mark the current node as visited
	visited[current.ID] = true
	defer func() { delete(visited, current.ID) }() // Unmark when backtracking

	// Check if we've reached the target
	if current.ID == target.ID {
		// Clone the path and add it to the result
		pathCopy := make([]*CallEdge, len(path))
		copy(pathCopy, path)
		*allPaths = append(*allPaths, pathCopy)
		return
	}

	// Check if we've exceeded the maximum path length
	if maxLength > 0 && len(path) >= maxLength {
		return
	}

	// Visit all unvisited neighbors
	for _, edge := range current.Calls {
		nextNode := edge.To
		if !visited[nextNode.ID] {
			// Add this edge to the path
			path = append(path, edge)

			// Recurse to the next node
			g.findPathsDFS(nextNode, target, path, visited, allPaths, maxLength)

			// Backtrack
			path = path[:len(path)-1]
		}
	}
}

// Size returns the number of nodes and edges in the graph.
func (g *CallGraph) Size() (nodes, edges int) {
	return len(g.Nodes), len(g.graph.Edges)
}

// FindCycles finds all cycles in the call graph.
func (g *CallGraph) FindCycles() [][]*CallEdge {
	var cycles [][]*CallEdge

	// Check each node for cycles starting from it
	for _, node := range g.Nodes {
		visited := make(map[string]bool)
		stack := make(map[string]bool)
		path := make([]*CallEdge, 0)

		g.findCyclesDFS(node, visited, stack, path, &cycles)
	}

	return cycles
}

// findCyclesDFS uses DFS to find cycles in the graph.
func (g *CallGraph) findCyclesDFS(node *CallNode,
	visited, stack map[string]bool,
	path []*CallEdge, cycles *[][]*CallEdge) {

	// Skip if already fully explored
	if visited[node.ID] {
		return
	}

	// Check if we've found a cycle
	if stack[node.ID] {
		// We need to extract the cycle from the path
		for i, edge := range path {
			if edge.From.ID == node.ID {
				// Found the start of the cycle
				cyclePath := make([]*CallEdge, len(path)-i)
				copy(cyclePath, path[i:])
				*cycles = append(*cycles, cyclePath)
				break
			}
		}
		return
	}

	// Mark as in-progress
	stack[node.ID] = true

	// Explore outgoing edges
	for _, edge := range node.Calls {
		path = append(path, edge)
		g.findCyclesDFS(edge.To, visited, stack, path, cycles)
		path = path[:len(path)-1]
	}

	// Mark as fully explored
	visited[node.ID] = true
	stack[node.ID] = false
}

// DeadFunctions finds functions that are never called.
// Excludes main functions and exported functions if specified.
func (g *CallGraph) DeadFunctions(excludeExported, excludeMain bool) []*CallNode {
	var deadFuncs []*CallNode

	for _, node := range g.Nodes {
		// Skip main functions if requested
		if excludeMain && isMainFunction(node.Symbol) {
			continue
		}

		// Skip exported functions if requested
		if excludeExported && node.Symbol.Exported {
			continue
		}

		// A function is dead if it has no incoming calls
		if len(node.CalledBy) == 0 {
			deadFuncs = append(deadFuncs, node)
		}
	}

	return deadFuncs
}

// Helper functions

// isCallable checks if a symbol is a function or method.
func isCallable(sym *typesys.Symbol) bool {
	if sym == nil {
		return false
	}
	return sym.Kind == typesys.KindFunction || sym.Kind == typesys.KindMethod
}

// getSymbolID gets a unique ID for a symbol.
func getSymbolID(sym *typesys.Symbol) string {
	if sym == nil {
		return ""
	}

	// For functions, include the package path for uniqueness
	// For methods, include the receiver type as well
	if sym.Package != nil {
		pkg := sym.Package.ImportPath
		if sym.Kind == typesys.KindMethod && sym.Parent != nil {
			return fmt.Sprintf("%s.%s.%s", pkg, sym.Parent.Name, sym.Name)
		}
		return fmt.Sprintf("%s.%s", pkg, sym.Name)
	}

	return sym.Name
}

// isMainFunction checks if a function is a main function.
func isMainFunction(sym *typesys.Symbol) bool {
	// Check if it's the main function in the main package
	return sym != nil && sym.Name == "main" &&
		sym.Package != nil && sym.Package.Name == "main"
}
