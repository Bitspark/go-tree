package graph

import (
	"testing"
)

// createTestGraph creates a test directed graph with the following structure:
//
//	  A
//	 / \
//	B   C
//	|   | \
//	D   E  F
//	|      |
//	G------+
func createTestGraph() *DirectedGraph {
	g := NewDirectedGraph()

	// Add nodes
	g.AddNode("A", "Node A")
	g.AddNode("B", "Node B")
	g.AddNode("C", "Node C")
	g.AddNode("D", "Node D")
	g.AddNode("E", "Node E")
	g.AddNode("F", "Node F")
	g.AddNode("G", "Node G")

	// Add edges
	g.AddEdge("A", "B", nil)
	g.AddEdge("A", "C", nil)
	g.AddEdge("B", "D", nil)
	g.AddEdge("C", "E", nil)
	g.AddEdge("C", "F", nil)
	g.AddEdge("D", "G", nil)
	g.AddEdge("G", "F", nil)

	return g
}

// createCyclicGraph creates a test directed graph with cycles
//
//	A -> B -> C
//	^         |
//	|         v
//	+---- D <-+
func createCyclicGraph() *DirectedGraph {
	g := NewDirectedGraph()

	// Add nodes
	g.AddNode("A", "Node A")
	g.AddNode("B", "Node B")
	g.AddNode("C", "Node C")
	g.AddNode("D", "Node D")

	// Add edges to form a cycle
	g.AddEdge("A", "B", nil)
	g.AddEdge("B", "C", nil)
	g.AddEdge("C", "D", nil)
	g.AddEdge("D", "A", nil)

	return g
}

func TestDFSBasic(t *testing.T) {
	g := createTestGraph()

	// Expected traversal order for DFS from A
	expectedOrder := []string{"A", "B", "D", "G", "C", "E", "F"}

	// Perform DFS traversal
	var visitedOrder []string
	visitor := func(node *Node) bool {
		visitedOrder = append(visitedOrder, node.ID.(string))
		return true
	}

	DFS(g, "A", visitor)

	// Check if the order matches expected (note: DFS order can vary depending on implementation)
	if len(visitedOrder) != len(expectedOrder) {
		t.Errorf("DFS visited %d nodes, expected %d", len(visitedOrder), len(expectedOrder))
	}

	// All nodes should be visited
	if len(visitedOrder) != 7 {
		t.Errorf("DFS should visit all 7 nodes, got %d: %v", len(visitedOrder), visitedOrder)
	}

	// First node should be A
	if visitedOrder[0] != "A" {
		t.Errorf("DFS should start with A, got %s", visitedOrder[0])
	}
}

func TestBFSBasic(t *testing.T) {
	g := createTestGraph()

	// Expected traversal order for BFS from A (specific to our implementation)
	// This is the expected order with our BFS algorithm
	// The exact ordering can vary by implementation, so we're asserting on the final content

	// Perform BFS traversal
	var visitedOrder []string
	visitor := func(node *Node) bool {
		visitedOrder = append(visitedOrder, node.ID.(string))
		return true
	}

	BFS(g, "A", visitor)

	// Check visited count
	if len(visitedOrder) != 7 {
		t.Errorf("BFS should visit all 7 nodes, got %d: %v", len(visitedOrder), visitedOrder)
	}

	// First node should be A
	if visitedOrder[0] != "A" {
		t.Errorf("BFS should start with A, got %s", visitedOrder[0])
	}

	// Level 1 should be B and C (order may vary)
	level1 := map[string]bool{visitedOrder[1]: true, visitedOrder[2]: true}
	if !level1["B"] || !level1["C"] {
		t.Errorf("BFS level 1 should contain B and C, got %v", level1)
	}

	// Check that G is visited
	foundG := false
	for _, id := range visitedOrder {
		if id == "G" {
			foundG = true
			break
		}
	}

	if !foundG {
		t.Errorf("BFS should visit G, but didn't find it in %v", visitedOrder)
	}
}

func TestTraversalOptions(t *testing.T) {
	g := createTestGraph()

	// Test with custom options
	opts := &TraversalOptions{
		Direction:    DirectionOut,
		Order:        OrderDFS,
		MaxDepth:     2, // Limit depth to 2 (A -> B/C -> D/E/F)
		IncludeStart: true,
	}

	var visitedNodes []string
	visitor := func(node *Node) bool {
		visitedNodes = append(visitedNodes, node.ID.(string))
		return true
	}

	Traverse(g, "A", opts, visitor)

	// Should visit A, B, C, D, E, F but not G (G is at depth 3)
	if len(visitedNodes) != 6 {
		t.Errorf("Depth-limited traversal should visit 6 nodes, got %d: %v", len(visitedNodes), visitedNodes)
	}

	// Check if G is not visited
	for _, id := range visitedNodes {
		if id == "G" {
			t.Errorf("Node G should not be visited with depth limit 2, but was found in %v", visitedNodes)
		}
	}

	// Now test with a skip function
	opts = &TraversalOptions{
		Direction: DirectionOut,
		Order:     OrderDFS,
		SkipFunc: func(node *Node) bool {
			// Skip node C and its subtree
			return node.ID == "C"
		},
		IncludeStart: true,
	}

	// Clear previous results
	visitedNodes = nil
	Traverse(g, "A", opts, visitor)

	// Should visit A, B, D, G, and possibly F (since F can be reached from G)
	// The important part is that C and E should NEVER be visited
	requiredNodes := map[string]bool{
		"A": true, "B": true, "D": true, "G": true,
	}
	forbiddenNodes := map[string]bool{
		"C": true, "E": true,
	}

	// Check for required nodes
	for _, id := range visitedNodes {
		delete(requiredNodes, id)
		// Check if any forbidden nodes were visited
		if forbiddenNodes[id] {
			t.Errorf("Node %s should not be visited with skip function", id)
		}
	}

	// Check if all required nodes were visited
	if len(requiredNodes) > 0 {
		t.Errorf("Some required nodes were not visited: %v", requiredNodes)
	}
}

func TestTraversalDirections(t *testing.T) {
	g := createTestGraph()

	// Test outgoing direction (already tested in other tests)
	outOpts := &TraversalOptions{
		Direction:    DirectionOut,
		Order:        OrderBFS,
		IncludeStart: true,
	}

	var outNodes []string
	outVisitor := func(node *Node) bool {
		outNodes = append(outNodes, node.ID.(string))
		return true
	}

	Traverse(g, "C", outOpts, outVisitor)

	// Verify C -> E, F edges
	if !contains(outNodes, "C") || !contains(outNodes, "E") || !contains(outNodes, "F") {
		t.Errorf("Outgoing traversal from C should visit C, E, F, got %v", outNodes)
	}

	// Test incoming direction
	inOpts := &TraversalOptions{
		Direction:    DirectionIn,
		Order:        OrderBFS,
		IncludeStart: true,
	}

	var inNodes []string
	inVisitor := func(node *Node) bool {
		inNodes = append(inNodes, node.ID.(string))
		return true
	}

	Traverse(g, "G", inOpts, inVisitor)

	// Verify G has incoming edges from D
	foundG := contains(inNodes, "G")
	foundD := contains(inNodes, "D")

	if !foundG {
		t.Errorf("Incoming traversal from G should visit G, got %v", inNodes)
	}

	if !foundD {
		t.Errorf("Incoming traversal from G should visit D, got %v", inNodes)
	}

	// Test both directions
	bothOpts := &TraversalOptions{
		Direction:    DirectionBoth,
		Order:        OrderBFS,
		IncludeStart: true,
	}

	var bothNodes []string
	bothVisitor := func(node *Node) bool {
		bothNodes = append(bothNodes, node.ID.(string))
		return true
	}

	Traverse(g, "G", bothOpts, bothVisitor)

	// Check that D and F are found (G has incoming edge from D, outgoing to F)
	foundG = contains(bothNodes, "G")
	foundD = contains(bothNodes, "D")
	var foundF = contains(bothNodes, "F")

	if !foundG || !foundD || !foundF {
		t.Errorf("Bidirectional traversal from G should find G, D, and F, got %v", bothNodes)
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestCollectNodes(t *testing.T) {
	g := createTestGraph()

	// Collect all nodes reachable from A
	nodes := CollectNodes(g, "A", DefaultTraversalOptions())

	// All 7 nodes should be reachable
	if len(nodes) != 7 {
		t.Errorf("CollectNodes should find all 7 nodes, got %d", len(nodes))
	}

	// Collect with depth limit of 1
	opts := &TraversalOptions{
		Direction:    DirectionOut,
		Order:        OrderBFS,
		MaxDepth:     1,
		IncludeStart: true,
	}

	nodes = CollectNodes(g, "A", opts)

	// Should find A, B, C
	if len(nodes) != 3 {
		t.Errorf("Depth-limited CollectNodes should find 3 nodes, got %d", len(nodes))
	}

	// Check node IDs
	idMap := make(map[interface{}]bool)
	for _, node := range nodes {
		idMap[node.ID] = true
	}

	if !idMap["A"] || !idMap["B"] || !idMap["C"] {
		t.Errorf("Depth-limited CollectNodes should find A, B, C, got %v", idMap)
	}
}

func TestFindAllReachable(t *testing.T) {
	g := createTestGraph()

	// Find all nodes reachable from D
	reachable := FindAllReachable(g, "D")

	// D can reach G and F
	if len(reachable) != 3 {
		t.Errorf("FindAllReachable from D should find 3 nodes, got %d", len(reachable))
	}

	// Check node IDs
	idMap := make(map[interface{}]bool)
	for _, node := range reachable {
		idMap[node.ID] = true
	}

	if !idMap["D"] || !idMap["G"] || !idMap["F"] {
		t.Errorf("FindAllReachable from D should find D, G, F, got %v", idMap)
	}
}

func TestTopologicalSort(t *testing.T) {
	// Create a simple DAG for topo sort
	g := NewDirectedGraph()

	// 1 -> 2 -> 3
	//  \-> 4 -/
	g.AddNode("1", nil)
	g.AddNode("2", nil)
	g.AddNode("3", nil)
	g.AddNode("4", nil)

	g.AddEdge("1", "2", nil)
	g.AddEdge("1", "4", nil)
	g.AddEdge("2", "3", nil)
	g.AddEdge("4", "3", nil)

	// Get topological order
	sorted, err := TopologicalSort(g)

	// Should be no error
	if err != nil {
		t.Errorf("TopologicalSort returned error: %v", err)
	}

	// Should have all 4 nodes
	if len(sorted) != 4 {
		t.Errorf("TopologicalSort should return 4 nodes, got %d", len(sorted))
	}

	// Check the order
	if sorted[0].ID != "1" {
		t.Errorf("First node should be 1, got %v", sorted[0].ID)
	}

	if sorted[3].ID != "3" {
		t.Errorf("Last node should be 3, got %v", sorted[3].ID)
	}

	// Test with a cyclic graph
	cyclic := createCyclicGraph()
	_, err = TopologicalSort(cyclic)

	// Should return an error for cyclic graph
	if err == nil {
		t.Error("TopologicalSort should return error for cyclic graph")
	}
}

func TestStopTraversalEarly(t *testing.T) {
	g := createTestGraph()

	// Visitor that stops after 3 nodes
	counter := 0
	visitor := func(node *Node) bool {
		counter++
		return counter < 3 // Stop after visiting 3 nodes
	}

	DFS(g, "A", visitor)

	// Counter should be 3
	if counter != 3 {
		t.Errorf("Visitor should visit exactly 3 nodes, got %d", counter)
	}

	// Test with BFS as well
	counter = 0
	BFS(g, "A", visitor)

	// Counter should be 3
	if counter != 3 {
		t.Errorf("BFS visitor should visit exactly 3 nodes, got %d", counter)
	}
}
