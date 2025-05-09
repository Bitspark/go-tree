package graph

import (
	"testing"
)

// createPathTestGraph creates a weighted directed graph for path testing
//
//	    B
//	  /   \
//	 /     \
//	A --- C  E
//	 \   / \ /
//	  \ /   D
func createPathTestGraph() *DirectedGraph {
	g := NewDirectedGraph()

	// Add nodes
	g.AddNode("A", nil)
	g.AddNode("B", nil)
	g.AddNode("C", nil)
	g.AddNode("D", nil)
	g.AddNode("E", nil)

	// Add edges with weights in data field
	g.AddEdge("A", "B", 4.0) // A->B with weight 4
	g.AddEdge("A", "C", 2.0) // A->C with weight 2
	g.AddEdge("B", "E", 3.0) // B->E with weight 3
	g.AddEdge("C", "B", 1.0) // C->B with weight 1
	g.AddEdge("C", "D", 2.0) // C->D with weight 2
	g.AddEdge("C", "E", 4.0) // C->E with weight 4
	g.AddEdge("D", "E", 1.0) // D->E with weight 1

	return g
}

// Custom weight function that uses the edge data as weight
func weightFunc(edge *Edge) float64 {
	if w, ok := edge.Data.(float64); ok {
		return w
	}
	return 1.0 // Default weight
}

func TestNewPath(t *testing.T) {
	path := NewPath()

	if path == nil {
		t.Fatal("NewPath should return a non-nil path")
	}

	if path.Nodes == nil {
		t.Error("Path.Nodes should be initialized")
	}

	if path.Edges == nil {
		t.Error("Path.Edges should be initialized")
	}

	if path.Cost != 0 {
		t.Errorf("New path should have cost 0, got %f", path.Cost)
	}

	if path.Length() != 0 {
		t.Errorf("New path should have length 0, got %d", path.Length())
	}
}

func TestPathAddNodeAndEdge(t *testing.T) {
	path := NewPath()
	g := createPathTestGraph()

	// Add nodes and edges to the path
	nodeA := g.GetNode("A")
	nodeB := g.GetNode("B")
	edgeAB := g.GetEdge("A", "B")

	path.AddNode(nodeA)
	if path.Length() != 1 {
		t.Errorf("Path with one node should have length 1, got %d", path.Length())
	}

	path.AddNode(nodeB)
	if path.Length() != 2 {
		t.Errorf("Path with two nodes should have length 2, got %d", path.Length())
	}

	path.AddEdge(edgeAB)
	if len(path.Edges) != 1 {
		t.Errorf("Path should have 1 edge, got %d", len(path.Edges))
	}

	// Check cost (using DefaultEdgeWeight which is 1.0)
	if path.Cost != DefaultEdgeWeight(edgeAB) {
		t.Errorf("Path cost should be %f, got %f", DefaultEdgeWeight(edgeAB), path.Cost)
	}
}

func TestPathClone(t *testing.T) {
	g := createPathTestGraph()
	path := NewPath()

	// Add some nodes and edges
	nodeA := g.GetNode("A")
	nodeC := g.GetNode("C")
	nodeB := g.GetNode("B")
	edgeAC := g.GetEdge("A", "C")
	edgeCB := g.GetEdge("C", "B")

	path.AddNode(nodeA)
	path.AddNode(nodeC)
	path.AddNode(nodeB)
	path.AddEdge(edgeAC)
	path.AddEdge(edgeCB)

	// Clone the path
	cloned := path.Clone()

	// Check the cloned path
	if cloned.Length() != path.Length() {
		t.Errorf("Cloned path should have same length, got %d, expected %d", cloned.Length(), path.Length())
	}

	if len(cloned.Edges) != len(path.Edges) {
		t.Errorf("Cloned path should have same number of edges, got %d, expected %d", len(cloned.Edges), len(path.Edges))
	}

	if cloned.Cost != path.Cost {
		t.Errorf("Cloned path should have same cost, got %f, expected %f", cloned.Cost, path.Cost)
	}

	// Modify original, check that clone is unchanged
	nodeD := g.GetNode("D")
	path.AddNode(nodeD)

	if cloned.Length() == path.Length() {
		t.Error("Cloned path should not be affected by changes to original")
	}
}

func TestPathReverse(t *testing.T) {
	g := createPathTestGraph()
	path := NewPath()

	// Create a path A -> C -> D
	nodeA := g.GetNode("A")
	nodeC := g.GetNode("C")
	nodeD := g.GetNode("D")
	edgeAC := g.GetEdge("A", "C")
	edgeCD := g.GetEdge("C", "D")

	path.AddNode(nodeA)
	path.AddNode(nodeC)
	path.AddNode(nodeD)
	path.AddEdge(edgeAC)
	path.AddEdge(edgeCD)

	// Check original path order
	if path.Nodes[0].ID != "A" || path.Nodes[1].ID != "C" || path.Nodes[2].ID != "D" {
		t.Errorf("Path nodes should be in order A,C,D, got %v,%v,%v",
			path.Nodes[0].ID, path.Nodes[1].ID, path.Nodes[2].ID)
	}

	// Reverse the path
	path.Reverse()

	// Check reversed order
	if path.Nodes[0].ID != "D" || path.Nodes[1].ID != "C" || path.Nodes[2].ID != "A" {
		t.Errorf("Reversed path should be D,C,A, got %v,%v,%v",
			path.Nodes[0].ID, path.Nodes[1].ID, path.Nodes[2].ID)
	}

	// Check edges are reversed too
	if path.Edges[0].From.ID != "C" || path.Edges[0].To.ID != "D" {
		t.Errorf("First edge in reversed path should be C->D, got %v->%v",
			path.Edges[0].From.ID, path.Edges[0].To.ID)
	}

	if path.Edges[1].From.ID != "A" || path.Edges[1].To.ID != "C" {
		t.Errorf("Second edge in reversed path should be A->C, got %v->%v",
			path.Edges[1].From.ID, path.Edges[1].To.ID)
	}
}

func TestPathContains(t *testing.T) {
	g := createPathTestGraph()
	path := NewPath()

	// Create a path A -> C -> D
	nodeA := g.GetNode("A")
	nodeC := g.GetNode("C")
	nodeD := g.GetNode("D")

	path.AddNode(nodeA)
	path.AddNode(nodeC)
	path.AddNode(nodeD)

	// Test contains
	if !path.Contains("A") {
		t.Error("Path should contain node A")
	}

	if !path.Contains("C") {
		t.Error("Path should contain node C")
	}

	if !path.Contains("D") {
		t.Error("Path should contain node D")
	}

	if path.Contains("B") {
		t.Error("Path should not contain node B")
	}

	if path.Contains("E") {
		t.Error("Path should not contain node E")
	}
}

func TestPathExists(t *testing.T) {
	g := createPathTestGraph()

	// Test paths that should exist
	if !PathExists(g, "A", "E") {
		t.Error("Path should exist from A to E")
	}

	if !PathExists(g, "A", "A") {
		t.Error("Path should exist from A to A (self)")
	}

	if !PathExists(g, "C", "E") {
		t.Error("Path should exist from C to E")
	}

	// Test paths that should not exist
	// Add a disconnected node
	g.AddNode("F", nil)
	if PathExists(g, "A", "F") {
		t.Error("Path should not exist from A to F")
	}

	if PathExists(g, "E", "A") {
		t.Error("Path should not exist from E to A (no backward path)")
	}
}

func TestFindShortestPath(t *testing.T) {
	g := createPathTestGraph()

	// Test shortest path from A to E
	// A->C->B->E (path length 3) vs A->C->D->E (path length 3) vs A->C->E (path length 2)
	// Shortest is A->C->E
	path := FindShortestPath(g, "A", "E")

	if path == nil {
		t.Fatal("FindShortestPath should return a path from A to E")
	}

	if path.Length() != 3 { // A, C, E (3 nodes)
		t.Errorf("Shortest path should have 3 nodes, got %d", path.Length())
	}

	// Check the path is A->C->E
	if path.Nodes[0].ID != "A" || path.Nodes[1].ID != "C" || path.Nodes[2].ID != "E" {
		t.Errorf("Shortest path should be A->C->E, got %v->%v->%v",
			path.Nodes[0].ID, path.Nodes[1].ID, path.Nodes[2].ID)
	}

	// Test path to self
	selfPath := FindShortestPath(g, "A", "A")
	if selfPath == nil {
		t.Fatal("FindShortestPath should return a path from A to A")
	}

	if selfPath.Length() != 1 {
		t.Errorf("Path to self should have length 1, got %d", selfPath.Length())
	}

	// Test non-existent path
	g.AddNode("F", nil) // Disconnected node
	noPath := FindShortestPath(g, "A", "F")
	if noPath != nil {
		t.Error("FindShortestPath should return nil for non-existent path")
	}
}

func TestFindShortestWeightedPath(t *testing.T) {
	g := createPathTestGraph()

	// Test weighted path from A to E
	// Paths with edge weights:
	// A->B->E (4+3=7)
	// A->C->B->E (2+1+3=6)
	// A->C->E (2+4=6)
	// A->C->D->E (2+2+1=5) <- shortest

	path := FindShortestWeightedPath(g, "A", "E", weightFunc)

	if path == nil {
		t.Fatal("FindShortestWeightedPath should return a path from A to E")
	}

	// Check the path is A->C->D->E
	if path.Length() != 4 { // A, C, D, E (4 nodes)
		t.Errorf("Shortest weighted path should have 4 nodes, got %d", path.Length())
	}

	expectedNodes := []string{"A", "C", "D", "E"}
	for i, expectedID := range expectedNodes {
		if i < len(path.Nodes) && path.Nodes[i].ID != expectedID {
			t.Errorf("Shortest weighted path node %d should be %s, got %v", i, expectedID, path.Nodes[i].ID)
		}
	}

	// Check the path cost
	expectedCost := 5.0 // A->C (2) + C->D (2) + D->E (1) = 5
	if path.Cost != expectedCost {
		t.Errorf("Path cost should be %f, got %f", expectedCost, path.Cost)
	}

	// Test with default weight function (all weights = 1.0)
	// Should be the same result as FindShortestPath (fewest edges)
	defaultPath := FindShortestWeightedPath(g, "A", "E", nil)
	if defaultPath.Length() != 3 { // A, C, E (3 nodes, 2 edges)
		t.Errorf("Path with default weights should have 3 nodes, got %d", defaultPath.Length())
	}

	// Test path to self
	selfPath := FindShortestWeightedPath(g, "A", "A", weightFunc)
	if selfPath == nil {
		t.Fatal("FindShortestWeightedPath should return a path from A to A")
	}

	if selfPath.Length() != 1 {
		t.Errorf("Path to self should have length 1, got %d", selfPath.Length())
	}

	if selfPath.Cost != 0 {
		t.Errorf("Path to self should have cost 0, got %f", selfPath.Cost)
	}

	// Test with non-existent node
	nonExistentPath := FindShortestWeightedPath(g, "A", "Z", weightFunc)
	if nonExistentPath != nil {
		t.Error("FindShortestWeightedPath should return nil for non-existent end node")
	}

	nonExistentPath = FindShortestWeightedPath(g, "Z", "A", weightFunc)
	if nonExistentPath != nil {
		t.Error("FindShortestWeightedPath should return nil for non-existent start node")
	}
}

func TestFindAllPaths(t *testing.T) {
	g := createPathTestGraph()

	// Find all paths from A to E with no limit
	paths := FindAllPaths(g, "A", "E", 0)

	// Should find at least 4 paths: A->B->E, A->C->B->E, A->C->E, A->C->D->E
	if len(paths) < 4 {
		t.Errorf("Should find at least 4 paths from A to E, got %d", len(paths))
	}

	// Check all paths start with A and end with E
	for i, p := range paths {
		if p.Nodes[0].ID != "A" {
			t.Errorf("Path %d should start with A, got %v", i, p.Nodes[0].ID)
		}

		if p.Nodes[len(p.Nodes)-1].ID != "E" {
			t.Errorf("Path %d should end with E, got %v", i, p.Nodes[len(p.Nodes)-1].ID)
		}
	}

	// Test with max length 3 (allowing 3 nodes, so 2 edges)
	limitedPaths := FindAllPaths(g, "A", "E", 3)

	// Should find 2 paths: A->B->E, A->C->E
	if len(limitedPaths) != 2 {
		t.Errorf("Should find 2 paths with max length 3, got %d", len(limitedPaths))
	}

	// Verify max length
	for i, p := range limitedPaths {
		if p.Length() > 3 {
			t.Errorf("Path %d exceeds max length 3, got %d", i, p.Length())
		}
	}
}

func TestCustomWeightFunctions(t *testing.T) {
	g := createPathTestGraph()

	// Define a custom weight function that prefers certain nodes
	// Make paths through B very expensive
	avoidBWeight := func(edge *Edge) float64 {
		if edge.To.ID == "B" || edge.From.ID == "B" {
			return 100.0 // Very high weight for edges involving B
		}
		return weightFunc(edge) // Normal weight for other edges
	}

	path := FindShortestWeightedPath(g, "A", "E", avoidBWeight)

	// Check that path avoids B
	for _, node := range path.Nodes {
		if node.ID == "B" {
			t.Error("Path should avoid node B when using custom weight function")
		}
	}

	// Should choose A->C->D->E
	expectedPath := []string{"A", "C", "D", "E"}
	for i, expected := range expectedPath {
		if i < len(path.Nodes) && path.Nodes[i].ID != expected {
			t.Errorf("Expected path node %d to be %s, got %v", i, expected, path.Nodes[i].ID)
		}
	}
}
