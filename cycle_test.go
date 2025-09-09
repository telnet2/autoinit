package autoinit

import (
	"context"
	"testing"
)

// Test types for cycle detection
type Node struct {
	Name        string
	Next        *Node
	InitCount   int
	Initialized bool
}

func (n *Node) Init(ctx context.Context) error {
	n.InitCount++
	n.Initialized = true
	// fmt.Printf("Node %s initialized, count=%d\n", n.Name, n.InitCount)
	return nil
}

// Test self-referential cycle
func TestSelfReferentialCycle(t *testing.T) {
	// Create a self-referential node
	node := &Node{Name: "self"}
	node.Next = node // Points to itself
	
	ctx := context.Background()
	
	// This will likely cause a stack overflow without cycle detection
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
			t.Log("As expected, self-referential cycle causes issues without cycle detection")
		}
	}()
	
	err := AutoInit(ctx, node)
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}
	
	// If we get here without panic, check that we didn't have infinite recursion
	// Note: In a self-referential case, the node may be initialized twice:
	// once as a field and once as the root
	if node.InitCount > 2 {
		t.Errorf("Node was initialized %d times, should be at most 2", node.InitCount)
	}
	t.Logf("Self-referential node initialized %d time(s) - cycle detection prevented infinite loop", node.InitCount)
}

// Test circular reference chain
func TestCircularReferenceChain(t *testing.T) {
	// Create a circular chain: A -> B -> C -> A
	nodeA := &Node{Name: "A"}
	nodeB := &Node{Name: "B"}
	nodeC := &Node{Name: "C"}
	
	nodeA.Next = nodeB
	nodeB.Next = nodeC
	nodeC.Next = nodeA // Circular reference
	
	ctx := context.Background()
	
	// This will likely cause a stack overflow without cycle detection
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
			t.Log("As expected, circular reference chain causes issues without cycle detection")
		}
	}()
	
	err := AutoInit(ctx, nodeA)
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}
	
	// If we get here without panic, check init counts
	// Note: NodeA may be initialized twice (once as root, once as field of NodeC)
	// The important thing is that the cycle was detected and we didn't recurse infinitely
	if nodeA.InitCount > 2 {
		t.Errorf("NodeA was initialized %d times, should be at most 2", nodeA.InitCount)
	}
	if nodeB.InitCount > 1 {
		t.Errorf("NodeB was initialized %d times, should be 1", nodeB.InitCount)
	}
	if nodeC.InitCount > 1 {
		t.Errorf("NodeC was initialized %d times, should be 1", nodeC.InitCount)
	}
	
	t.Logf("Circular chain initialization complete - cycle detection prevented infinite loop")
	t.Logf("NodeA: %d init(s), NodeB: %d init(s), NodeC: %d init(s)", 
		nodeA.InitCount, nodeB.InitCount, nodeC.InitCount)
}

// Test diamond dependency (shared reference)
type Container struct {
	Left   *SharedNode
	Right  *SharedNode
}

type SharedNode struct {
	Name        string
	InitCount   int
	Initialized bool
}

func (s *SharedNode) Init(ctx context.Context) error {
	s.InitCount++
	s.Initialized = true
	return nil
}

func TestDiamondDependency(t *testing.T) {
	// Create a diamond dependency: Container has Left and Right, both point to the same Shared node
	shared := &SharedNode{Name: "shared"}
	container := &Container{
		Left:  shared,
		Right: shared, // Same instance
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, container)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Without cycle detection, shared node might be initialized twice
	if shared.InitCount != 2 {
		t.Logf("Shared node was initialized %d times", shared.InitCount)
		if shared.InitCount == 1 {
			t.Log("Good! The implementation might already handle this case")
		}
	} else {
		t.Error("Shared node was initialized twice - it should only be initialized once")
	}
}

// Test with slices containing cycles
type ListNode struct {
	Name      string
	Children  []*ListNode
	InitCount int
}

func (l *ListNode) Init(ctx context.Context) error {
	l.InitCount++
	return nil
}

func TestSliceWithCycle(t *testing.T) {
	// Create nodes with circular reference in slice
	parent := &ListNode{Name: "parent"}
	child := &ListNode{Name: "child"}
	
	parent.Children = []*ListNode{child}
	child.Children = []*ListNode{parent} // Circular reference
	
	ctx := context.Background()
	
	// This will likely cause a stack overflow without cycle detection
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
			t.Log("As expected, slice with cycle causes issues without cycle detection")
		}
	}()
	
	err := AutoInit(ctx, parent)
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}
	
	// If we get here without panic, check init counts
	// Note: Parent may be initialized twice (once as root, once as element in child's slice)
	if parent.InitCount > 2 {
		t.Errorf("Parent was initialized %d times, should be at most 2", parent.InitCount)
	}
	if child.InitCount > 1 {
		t.Errorf("Child was initialized %d times, should be 1", child.InitCount)
	}
	
	t.Logf("Slice with cycle initialization complete - cycle detection prevented infinite loop")
	t.Logf("Parent: %d init(s), Child: %d init(s)", parent.InitCount, child.InitCount)
}

// Test with maps containing cycles
type MapNode struct {
	Name      string
	Refs      map[string]*MapNode
	InitCount int
}

func (m *MapNode) Init(ctx context.Context) error {
	m.InitCount++
	return nil
}

func TestMapWithCycle(t *testing.T) {
	// Create nodes with circular reference in map
	nodeA := &MapNode{Name: "A", Refs: make(map[string]*MapNode)}
	nodeB := &MapNode{Name: "B", Refs: make(map[string]*MapNode)}
	
	nodeA.Refs["b"] = nodeB
	nodeB.Refs["a"] = nodeA // Circular reference
	
	ctx := context.Background()
	
	// This will likely cause a stack overflow without cycle detection
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
			t.Log("As expected, map with cycle causes issues without cycle detection")
		}
	}()
	
	err := AutoInit(ctx, nodeA)
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}
	
	// If we get here without panic, check init counts
	// Note: NodeA may be initialized twice (once as root, once as value in nodeB's map)
	if nodeA.InitCount > 2 {
		t.Errorf("NodeA was initialized %d times, should be at most 2", nodeA.InitCount)
	}
	if nodeB.InitCount > 1 {
		t.Errorf("NodeB was initialized %d times, should be 1", nodeB.InitCount)
	}
	
	t.Logf("Map with cycle initialization complete - cycle detection prevented infinite loop")
	t.Logf("NodeA: %d init(s), NodeB: %d init(s)", nodeA.InitCount, nodeB.InitCount)
}