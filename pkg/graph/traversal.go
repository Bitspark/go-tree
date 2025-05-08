package graph

import (
	"container/list"
	"errors"
)

// TraversalDirection defines the edge direction to follow during traversal.
type TraversalDirection int

const (
	// DirectionOut follows outgoing edges from the start node.
	DirectionOut TraversalDirection = iota
	// DirectionIn follows incoming edges to the start node.
	DirectionIn
	// DirectionBoth follows both incoming and outgoing edges.
	DirectionBoth
)

// TraversalOrder defines the order in which nodes are visited.
type TraversalOrder int

const (
	// OrderDFS uses depth-first search traversal.
	OrderDFS TraversalOrder = iota
	// OrderBFS uses breadth-first search traversal.
	OrderBFS
)

// TraversalOptions provides configuration options for graph traversal.
type TraversalOptions struct {
	// Direction controls which edges to follow (out, in, both).
	Direction TraversalDirection

	// Order controls the traversal order (DFS, BFS).
	Order TraversalOrder

	// MaxDepth limits the traversal depth (0 = unlimited).
	MaxDepth int

	// SkipFunc allows skipping nodes from traversal.
	SkipFunc func(node *Node) bool

	// IncludeStart determines whether to include the start node in traversal.
	IncludeStart bool
}

// DefaultTraversalOptions returns the default traversal options.
func DefaultTraversalOptions() *TraversalOptions {
	return &TraversalOptions{
		Direction:    DirectionOut,
		Order:        OrderDFS,
		MaxDepth:     0, // Unlimited
		SkipFunc:     nil,
		IncludeStart: true,
	}
}

// VisitFunc is called for each node during traversal.
// Return false to stop traversal immediately.
type VisitFunc func(node *Node) bool

// DFS performs depth-first traversal starting from a node.
func DFS(g *DirectedGraph, startID interface{}, visit VisitFunc) {
	opts := DefaultTraversalOptions()
	opts.Order = OrderDFS
	opts.Direction = DirectionOut

	Traverse(g, startID, opts, visit)
}

// BFS performs breadth-first traversal starting from a node.
func BFS(g *DirectedGraph, startID interface{}, visit VisitFunc) {
	opts := DefaultTraversalOptions()
	opts.Order = OrderBFS
	opts.Direction = DirectionOut

	Traverse(g, startID, opts, visit)
}

// Traverse traverses the graph with the specified options.
func Traverse(g *DirectedGraph, startID interface{}, opts *TraversalOptions, visit VisitFunc) {
	if g == nil || visit == nil {
		return
	}

	if opts == nil {
		opts = DefaultTraversalOptions()
	}

	// Get the start node
	start := g.GetNode(startID)
	if start == nil {
		return
	}

	// Initialize visited map
	visited := make(map[interface{}]bool)

	// Choose the appropriate traversal algorithm
	switch opts.Order {
	case OrderDFS:
		dfsWithOptions(g, start, visited, opts, visit, 0)
	case OrderBFS:
		bfsWithOptions(g, start, visited, opts, visit)
	}
}

// dfsWithOptions implements a depth-first search with options.
func dfsWithOptions(g *DirectedGraph, node *Node, visited map[interface{}]bool, opts *TraversalOptions, visit VisitFunc, depth int) bool {
	// Check if we've reached the maximum depth
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return true
	}

	// Mark as visited
	visited[node.ID] = true

	// Visit the current node (if not the start node, or if we want to include the start)
	if (depth > 0 || opts.IncludeStart) && !skipNode(node, opts) {
		if !visit(node) {
			return false // Stop traversal if visit returns false
		}
	}

	// Get neighbor nodes based on direction
	neighbors := getNeighbors(g, node, opts.Direction)

	// Visit each unvisited neighbor recursively
	for _, neighbor := range neighbors {
		if !visited[neighbor.ID] && !skipNode(neighbor, opts) {
			if !dfsWithOptions(g, neighbor, visited, opts, visit, depth+1) {
				return false
			}
		}
	}

	return true
}

// bfsWithOptions implements a breadth-first search with options.
func bfsWithOptions(g *DirectedGraph, start *Node, visited map[interface{}]bool, opts *TraversalOptions, visit VisitFunc) {
	// Create a queue and add the start node
	queue := list.New()

	// Track node depths
	depths := make(map[interface{}]int)

	// Add start node to queue
	queue.PushBack(start)
	depths[start.ID] = 0
	visited[start.ID] = true

	// Process the queue
	for queue.Len() > 0 {
		// Get the next node
		element := queue.Front()
		queue.Remove(element)

		node := element.Value.(*Node)
		depth := depths[node.ID]

		// Check if we've reached the maximum depth
		if opts.MaxDepth > 0 && depth > opts.MaxDepth {
			continue
		}

		// Visit the current node
		if (depth > 0 || opts.IncludeStart) && !skipNode(node, opts) {
			if !visit(node) {
				return // Stop traversal if visit returns false
			}
		}

		// Add unvisited neighbors to the queue
		neighbors := getNeighbors(g, node, opts.Direction)
		for _, neighbor := range neighbors {
			if !visited[neighbor.ID] && !skipNode(neighbor, opts) {
				visited[neighbor.ID] = true
				queue.PushBack(neighbor)
				depths[neighbor.ID] = depth + 1
			}
		}
	}
}

// getNeighbors returns the neighbors of a node based on the traversal direction.
func getNeighbors(g *DirectedGraph, node *Node, direction TraversalDirection) []*Node {
	var neighbors []*Node

	switch direction {
	case DirectionOut:
		// Get nodes connected by outgoing edges
		neighbors = g.GetOutNodes(node.ID)
	case DirectionIn:
		// Get nodes connected by incoming edges
		neighbors = g.GetInNodes(node.ID)
	case DirectionBoth:
		// Get both outgoing and incoming nodes
		outNodes := g.GetOutNodes(node.ID)
		inNodes := g.GetInNodes(node.ID)

		// Combine both sets, avoiding duplicates
		neighbors = outNodes
		nodeMap := make(map[interface{}]bool)

		for _, n := range outNodes {
			nodeMap[n.ID] = true
		}

		for _, n := range inNodes {
			if !nodeMap[n.ID] {
				neighbors = append(neighbors, n)
			}
		}
	}

	return neighbors
}

// skipNode checks if a node should be skipped based on options.
func skipNode(node *Node, opts *TraversalOptions) bool {
	if opts.SkipFunc != nil {
		return opts.SkipFunc(node)
	}
	return false
}

// CollectNodes traverses the graph and collects all visited nodes.
func CollectNodes(g *DirectedGraph, startID interface{}, opts *TraversalOptions) []*Node {
	if g == nil {
		return nil
	}

	var result []*Node

	// Define a visitor that collects nodes
	visitor := func(node *Node) bool {
		result = append(result, node)
		return true // Continue traversal
	}

	// Traverse the graph
	Traverse(g, startID, opts, visitor)

	return result
}

// FindAllReachable finds all nodes reachable from the start node.
func FindAllReachable(g *DirectedGraph, startID interface{}) []*Node {
	return CollectNodes(g, startID, &TraversalOptions{
		Direction:    DirectionOut,
		Order:        OrderBFS,
		IncludeStart: true,
	})
}

// TopologicalSort performs a topological sort of the graph.
// Returns an error if the graph contains a cycle.
func TopologicalSort(g *DirectedGraph) ([]*Node, error) {
	if g == nil {
		return nil, nil
	}

	// Create a copy of the graph's node list to avoid locking issues
	nodes := g.NodeList()

	// Track visited and temp-marked nodes (for cycle detection)
	visited := make(map[interface{}]bool)
	tempMarked := make(map[interface{}]bool)

	// Result list (in reverse order)
	var result []*Node

	// Helper function for DFS
	var visit func(node *Node) error
	visit = func(node *Node) error {
		// Check for cycle
		if tempMarked[node.ID] {
			return errors.New("graph contains a cycle")
		}

		// Skip if already visited
		if visited[node.ID] {
			return nil
		}

		// Mark temporarily
		tempMarked[node.ID] = true

		// Visit outgoing edges
		for _, neighbor := range g.GetOutNodes(node.ID) {
			if err := visit(neighbor); err != nil {
				return err
			}
		}

		// Mark as visited
		visited[node.ID] = true
		tempMarked[node.ID] = false

		// Add to result
		result = append(result, node)

		return nil
	}

	// Visit each unvisited node
	for _, node := range nodes {
		if !visited[node.ID] {
			if err := visit(node); err != nil {
				return nil, err
			}
		}
	}

	// Reverse the result to get topological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}
