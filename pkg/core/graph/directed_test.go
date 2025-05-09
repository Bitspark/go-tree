package graph

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewDirectedGraph(t *testing.T) {
	g := NewDirectedGraph()
	if g == nil {
		t.Fatal("Expected non-nil graph")
	}
	if g.Nodes == nil {
		t.Error("Nodes map should be initialized")
	}
	if g.Edges == nil {
		t.Error("Edges map should be initialized")
	}

	nodes, edges := g.Size()
	if nodes != 0 || edges != 0 {
		t.Errorf("New graph should be empty, got %d nodes, %d edges", nodes, edges)
	}
}

func TestAddNode(t *testing.T) {
	g := NewDirectedGraph()

	// Basic node addition
	node := g.AddNode("node1", "data1")
	if node == nil {
		t.Fatal("AddNode should return the added node")
	}

	if node.ID != "node1" {
		t.Errorf("Node ID should be 'node1', got %v", node.ID)
	}

	if node.Data != "data1" {
		t.Errorf("Node data should be 'data1', got %v", node.Data)
	}

	if node.graph != g {
		t.Error("Node's graph reference is incorrect")
	}

	// Check node is in the graph
	nodes, _ := g.Size()
	if nodes != 1 {
		t.Errorf("Graph should have 1 node, got %d", nodes)
	}

	// Test updating existing node data
	updatedNode := g.AddNode("node1", "updated_data")
	if updatedNode != node {
		t.Error("Updating a node should return the same node instance")
	}

	if updatedNode.Data != "updated_data" {
		t.Errorf("Updated node data should be 'updated_data', got %v", updatedNode.Data)
	}

	// Still just one node
	nodes, _ = g.Size()
	if nodes != 1 {
		t.Errorf("Graph should still have 1 node, got %d", nodes)
	}
}

func TestAddEdge(t *testing.T) {
	g := NewDirectedGraph()

	// Add nodes first
	node1 := g.AddNode("node1", "data1")
	node2 := g.AddNode("node2", "data2")

	// Add edge
	edge := g.AddEdge("node1", "node2", "edge_data")
	if edge == nil {
		t.Fatal("AddEdge should return the added edge")
	}

	if edge.From != node1 {
		t.Error("Edge's From node is incorrect")
	}

	if edge.To != node2 {
		t.Error("Edge's To node is incorrect")
	}

	if edge.Data != "edge_data" {
		t.Errorf("Edge data should be 'edge_data', got %v", edge.Data)
	}

	if edge.graph != g {
		t.Error("Edge's graph reference is incorrect")
	}

	// Check edge is in the graph
	_, edges := g.Size()
	if edges != 1 {
		t.Errorf("Graph should have 1 edge, got %d", edges)
	}

	// Check edge is in node's edge maps
	if len(node1.OutEdges) != 1 {
		t.Errorf("From node should have 1 outgoing edge, got %d", len(node1.OutEdges))
	}

	if len(node2.InEdges) != 1 {
		t.Errorf("To node should have 1 incoming edge, got %d", len(node2.InEdges))
	}

	// Test adding edge with non-existent nodes (should create them)
	edge2 := g.AddEdge("node3", "node4", "new_edge_data")
	if edge2 == nil {
		t.Fatal("Edge between new nodes should be created")
	}

	nodes, _ := g.Size()
	if nodes != 4 {
		t.Errorf("Graph should have 4 nodes, got %d", nodes)
	}

	// Test updating existing edge data
	updatedEdge := g.AddEdge("node1", "node2", "updated_edge_data")
	if updatedEdge != edge {
		t.Error("Updating an edge should return the same edge instance")
	}

	if updatedEdge.Data != "updated_edge_data" {
		t.Errorf("Updated edge data should be 'updated_edge_data', got %v", updatedEdge.Data)
	}
}

func TestRemoveNode(t *testing.T) {
	g := NewDirectedGraph()

	// Setup test graph
	g.AddNode("node1", "data1")
	g.AddNode("node2", "data2")
	g.AddNode("node3", "data3")

	g.AddEdge("node1", "node2", "edge12")
	g.AddEdge("node2", "node3", "edge23")
	g.AddEdge("node3", "node1", "edge31")

	// Remove a node
	g.RemoveNode("node2")

	// Check node was removed
	if g.HasNode("node2") {
		t.Error("Node2 should be removed")
	}

	// Check edges were removed
	if g.HasEdge("node1", "node2") {
		t.Error("Edge node1->node2 should be removed")
	}

	if g.HasEdge("node2", "node3") {
		t.Error("Edge node2->node3 should be removed")
	}

	// Check remaining graph structure
	nodes, edges := g.Size()
	if nodes != 2 {
		t.Errorf("Graph should have 2 nodes after removal, got %d", nodes)
	}

	if edges != 1 {
		t.Errorf("Graph should have 1 edge after removal, got %d", edges)
	}

	// Check removing non-existent node (should not crash)
	g.RemoveNode("nonexistent")
	nodesAfter, edgesAfter := g.Size()
	if nodesAfter != nodes || edgesAfter != edges {
		t.Error("Removing non-existent node should not change graph")
	}
}

func TestRemoveEdge(t *testing.T) {
	g := NewDirectedGraph()

	// Setup test graph
	g.AddNode("node1", "data1")
	g.AddNode("node2", "data2")
	g.AddEdge("node1", "node2", "edge_data")

	// Remove edge
	g.RemoveEdge("node1", "node2")

	// Check edge was removed
	if g.HasEdge("node1", "node2") {
		t.Error("Edge should be removed")
	}

	// Check nodes are still there
	if !g.HasNode("node1") || !g.HasNode("node2") {
		t.Error("Nodes should still exist after edge removal")
	}

	// Check node edge maps
	node1 := g.GetNode("node1")
	node2 := g.GetNode("node2")

	if len(node1.OutEdges) != 0 {
		t.Errorf("Node1 should have 0 outgoing edges, got %d", len(node1.OutEdges))
	}

	if len(node2.InEdges) != 0 {
		t.Errorf("Node2 should have 0 incoming edges, got %d", len(node2.InEdges))
	}

	// Check removing non-existent edge (should not crash)
	g.RemoveEdge("node1", "nonexistent")
	g.RemoveEdge("nonexistent", "node2")
}

func TestGraphQueryMethods(t *testing.T) {
	g := NewDirectedGraph()

	// Setup test graph
	g.AddNode("node1", "data1")
	g.AddNode("node2", "data2")
	g.AddNode("node3", "data3")

	g.AddEdge("node1", "node2", "edge12")
	g.AddEdge("node1", "node3", "edge13")
	g.AddEdge("node2", "node3", "edge23")

	// Test GetNode
	node := g.GetNode("node1")
	if node == nil {
		t.Fatal("GetNode should return the node")
	}

	if node.ID != "node1" {
		t.Errorf("GetNode returned wrong node, got ID %v", node.ID)
	}

	// Test GetEdge
	edge := g.GetEdge("node1", "node2")
	if edge == nil {
		t.Fatal("GetEdge should return the edge")
	}

	if edge.From.ID != "node1" || edge.To.ID != "node2" {
		t.Errorf("GetEdge returned wrong edge, got %s->%s", edge.From.ID, edge.To.ID)
	}

	// Test GetOutNodes
	outNodes := g.GetOutNodes("node1")
	if len(outNodes) != 2 {
		t.Errorf("GetOutNodes should return 2 nodes, got %d", len(outNodes))
	}

	outIDs := map[interface{}]bool{}
	for _, n := range outNodes {
		outIDs[n.ID] = true
	}

	if !outIDs["node2"] || !outIDs["node3"] {
		t.Error("GetOutNodes didn't return expected nodes")
	}

	// Test GetInNodes
	inNodes := g.GetInNodes("node3")
	if len(inNodes) != 2 {
		t.Errorf("GetInNodes should return 2 nodes, got %d", len(inNodes))
	}

	inIDs := map[interface{}]bool{}
	for _, n := range inNodes {
		inIDs[n.ID] = true
	}

	if !inIDs["node1"] || !inIDs["node2"] {
		t.Error("GetInNodes didn't return expected nodes")
	}

	// Test non-existent nodes
	if g.GetNode("nonexistent") != nil {
		t.Error("GetNode should return nil for non-existent nodes")
	}

	if g.GetEdge("node1", "nonexistent") != nil {
		t.Error("GetEdge should return nil for non-existent edges")
	}

	if len(g.GetOutNodes("nonexistent")) != 0 {
		t.Error("GetOutNodes should return empty slice for non-existent nodes")
	}

	if len(g.GetInNodes("nonexistent")) != 0 {
		t.Error("GetInNodes should return empty slice for non-existent nodes")
	}
}

func TestDirectedGraphConcurrency(t *testing.T) {
	g := NewDirectedGraph()

	// Add a few initial nodes
	g.AddNode("main", "main node")

	// Run concurrent operations
	var wg sync.WaitGroup
	concurrency := 50
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()

			// Add node
			nodeID := fmt.Sprintf("node%d", id)
			g.AddNode(nodeID, id)

			// Add edge to main
			g.AddEdge(nodeID, "main", id)

			// Read some data
			g.GetNode(nodeID)
			g.GetEdge(nodeID, "main")
		}(i)
	}

	wg.Wait()

	// Verify results
	nodes, edges := g.Size()
	if nodes != concurrency+1 { // +1 for the main node
		t.Errorf("Expected %d nodes, got %d", concurrency+1, nodes)
	}

	if edges != concurrency {
		t.Errorf("Expected %d edges, got %d", concurrency, edges)
	}
}

func TestGraphUtilityMethods(t *testing.T) {
	g := NewDirectedGraph()

	// Setup test graph
	g.AddNode("node1", "data1")
	g.AddNode("node2", "data2")
	g.AddNode("node3", "data3")

	g.AddEdge("node1", "node2", "edge12")
	g.AddEdge("node2", "node3", "edge23")

	// Test NodeIDs
	ids := g.NodeIDs()
	if len(ids) != 3 {
		t.Errorf("Expected 3 node IDs, got %d", len(ids))
	}

	idSet := make(map[interface{}]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	if !idSet["node1"] || !idSet["node2"] || !idSet["node3"] {
		t.Error("NodeIDs didn't return all expected IDs")
	}

	// Test NodeList
	nodes := g.NodeList()
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}

	// Test EdgeList
	edges := g.EdgeList()
	if len(edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(edges))
	}

	// Test HasNode and HasEdge
	if !g.HasNode("node1") {
		t.Error("HasNode should return true for existing node")
	}

	if g.HasNode("nonexistent") {
		t.Error("HasNode should return false for non-existent node")
	}

	if !g.HasEdge("node1", "node2") {
		t.Error("HasEdge should return true for existing edge")
	}

	if g.HasEdge("node1", "node3") {
		t.Error("HasEdge should return false for non-existent edge")
	}

	// Test InDegree and OutDegree
	if g.OutDegree("node1") != 1 {
		t.Errorf("Expected OutDegree of 1 for node1, got %d", g.OutDegree("node1"))
	}

	if g.InDegree("node2") != 1 {
		t.Errorf("Expected InDegree of 1 for node2, got %d", g.InDegree("node2"))
	}

	if g.InDegree("node3") != 1 {
		t.Errorf("Expected InDegree of 1 for node3, got %d", g.InDegree("node3"))
	}

	if g.OutDegree("node3") != 0 {
		t.Errorf("Expected OutDegree of 0 for node3, got %d", g.OutDegree("node3"))
	}

	// Test Clear
	g.Clear()
	nodesCount, edgesCount := g.Size()
	if nodesCount != 0 || edgesCount != 0 {
		t.Errorf("Graph should be empty after Clear, got %d nodes, %d edges", nodesCount, edgesCount)
	}
}
