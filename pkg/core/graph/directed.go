// Package graph provides generic graph data structures and algorithms for code analysis.
package graph

import (
	"fmt"
	"sync"
)

// DirectedGraph represents a simple directed graph with nodes and edges.
type DirectedGraph struct {
	// Nodes in the graph, indexed by their ID
	Nodes map[interface{}]*Node

	// Edges in the graph, indexed by their string ID (typically fromID->toID)
	Edges map[string]*Edge

	// Mutex for concurrent access
	mu sync.RWMutex
}

// Node represents a node in the graph with its edges.
type Node struct {
	// Unique identifier for the node
	ID interface{}

	// Arbitrary data associated with the node
	Data interface{}

	// Outgoing edges from this node (to -> edge)
	OutEdges map[interface{}]*Edge

	// Incoming edges to this node (from -> edge)
	InEdges map[interface{}]*Edge

	// Reference to containing graph
	graph *DirectedGraph
}

// Edge represents a directed edge between two nodes.
type Edge struct {
	// Unique identifier for the edge
	ID string

	// Source node
	From *Node

	// Target node
	To *Node

	// Arbitrary data associated with the edge
	Data interface{}

	// Reference to containing graph
	graph *DirectedGraph
}

// NewDirectedGraph creates a new empty directed graph.
func NewDirectedGraph() *DirectedGraph {
	return &DirectedGraph{
		Nodes: make(map[interface{}]*Node),
		Edges: make(map[string]*Edge),
	}
}

// AddNode adds a node to the graph with the given ID and data.
// If a node with the given ID already exists, its data is updated.
func (g *DirectedGraph) AddNode(id interface{}, data interface{}) *Node {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if the node already exists
	if node, exists := g.Nodes[id]; exists {
		node.Data = data
		return node
	}

	// Create a new node
	node := &Node{
		ID:       id,
		Data:     data,
		OutEdges: make(map[interface{}]*Edge),
		InEdges:  make(map[interface{}]*Edge),
		graph:    g,
	}

	// Add the node to the graph
	g.Nodes[id] = node

	return node
}

// AddEdge adds a directed edge between two nodes.
// If the nodes do not exist, they are created.
// Returns the created or existing edge.
func (g *DirectedGraph) AddEdge(fromID, toID interface{}, data interface{}) *Edge {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Create the nodes if they don't exist
	from := g.getOrCreateNode(fromID, nil)
	to := g.getOrCreateNode(toID, nil)

	// Generate a unique edge ID
	edgeID := fmt.Sprintf("%v->%v", fromID, toID)

	// Check if the edge already exists
	if edge, exists := g.Edges[edgeID]; exists {
		edge.Data = data
		return edge
	}

	// Create a new edge
	edge := &Edge{
		ID:    edgeID,
		From:  from,
		To:    to,
		Data:  data,
		graph: g,
	}

	// Update the node's edge references
	from.OutEdges[toID] = edge
	to.InEdges[fromID] = edge

	// Add the edge to the graph
	g.Edges[edgeID] = edge

	return edge
}

// getOrCreateNode gets a node by ID or creates a new one if it doesn't exist.
// This is an internal helper method used when adding edges.
func (g *DirectedGraph) getOrCreateNode(id interface{}, data interface{}) *Node {
	if node, exists := g.Nodes[id]; exists {
		return node
	}

	node := &Node{
		ID:       id,
		Data:     data,
		OutEdges: make(map[interface{}]*Edge),
		InEdges:  make(map[interface{}]*Edge),
		graph:    g,
	}

	g.Nodes[id] = node
	return node
}

// RemoveNode removes a node and all its edges from the graph.
func (g *DirectedGraph) RemoveNode(id interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.Nodes[id]
	if !exists {
		return
	}

	// Remove all outgoing edges
	for toID, edge := range node.OutEdges {
		// Remove edge from target node's InEdges
		if to := g.Nodes[toID]; to != nil {
			delete(to.InEdges, id)
		}

		// Remove edge from graph
		delete(g.Edges, edge.ID)
	}

	// Remove all incoming edges
	for fromID, edge := range node.InEdges {
		// Remove edge from source node's OutEdges
		if from := g.Nodes[fromID]; from != nil {
			delete(from.OutEdges, id)
		}

		// Remove edge from graph
		delete(g.Edges, edge.ID)
	}

	// Remove the node itself
	delete(g.Nodes, id)
}

// RemoveEdge removes an edge between two nodes.
func (g *DirectedGraph) RemoveEdge(fromID, toID interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Generate the edge ID
	edgeID := fmt.Sprintf("%v->%v", fromID, toID)

	// Check if the edge exists
	if _, exists := g.Edges[edgeID]; !exists {
		return
	}

	// Remove references from nodes
	if from := g.Nodes[fromID]; from != nil {
		delete(from.OutEdges, toID)
	}

	if to := g.Nodes[toID]; to != nil {
		delete(to.InEdges, fromID)
	}

	// Remove the edge from the graph
	delete(g.Edges, edgeID)
}

// GetNode gets a node by ID.
// Returns nil if the node does not exist.
func (g *DirectedGraph) GetNode(id interface{}) *Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.Nodes[id]
}

// GetEdge gets an edge by source and target node IDs.
// Returns nil if the edge does not exist.
func (g *DirectedGraph) GetEdge(fromID, toID interface{}) *Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edgeID := fmt.Sprintf("%v->%v", fromID, toID)
	return g.Edges[edgeID]
}

// GetOutNodes returns all nodes connected by outgoing edges from the given node.
func (g *DirectedGraph) GetOutNodes(id interface{}) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node := g.Nodes[id]
	if node == nil {
		return nil
	}

	result := make([]*Node, 0, len(node.OutEdges))
	for toID := range node.OutEdges {
		if to := g.Nodes[toID]; to != nil {
			result = append(result, to)
		}
	}

	return result
}

// GetInNodes returns all nodes connected by incoming edges to the given node.
func (g *DirectedGraph) GetInNodes(id interface{}) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node := g.Nodes[id]
	if node == nil {
		return nil
	}

	result := make([]*Node, 0, len(node.InEdges))
	for fromID := range node.InEdges {
		if from := g.Nodes[fromID]; from != nil {
			result = append(result, from)
		}
	}

	return result
}

// Size returns the number of nodes and edges in the graph.
func (g *DirectedGraph) Size() (nodes, edges int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.Nodes), len(g.Edges)
}

// Clear removes all nodes and edges from the graph.
func (g *DirectedGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Nodes = make(map[interface{}]*Node)
	g.Edges = make(map[string]*Edge)
}

// NodeIDs returns a slice of all node IDs in the graph.
func (g *DirectedGraph) NodeIDs() []interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	ids := make([]interface{}, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}

	return ids
}

// NodeList returns a slice of all nodes in the graph.
func (g *DirectedGraph) NodeList() []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*Node, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// EdgeList returns a slice of all edges in the graph.
func (g *DirectedGraph) EdgeList() []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := make([]*Edge, 0, len(g.Edges))
	for _, edge := range g.Edges {
		edges = append(edges, edge)
	}

	return edges
}

// HasNode checks if a node with the given ID exists in the graph.
func (g *DirectedGraph) HasNode(id interface{}) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	_, exists := g.Nodes[id]
	return exists
}

// HasEdge checks if an edge between the given nodes exists in the graph.
func (g *DirectedGraph) HasEdge(fromID, toID interface{}) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edgeID := fmt.Sprintf("%v->%v", fromID, toID)
	_, exists := g.Edges[edgeID]
	return exists
}

// OutDegree returns the number of outgoing edges from a node.
func (g *DirectedGraph) OutDegree(id interface{}) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if node := g.Nodes[id]; node != nil {
		return len(node.OutEdges)
	}

	return 0
}

// InDegree returns the number of incoming edges to a node.
func (g *DirectedGraph) InDegree(id interface{}) int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if node := g.Nodes[id]; node != nil {
		return len(node.InEdges)
	}

	return 0
}
