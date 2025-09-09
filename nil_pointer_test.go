package autoinit

import (
	"context"
	"testing"
)

// NilPointerStruct for testing nil pointer handling
type NilPointerStruct struct {
	Name        string
	InitCalled  bool
}

func (n *NilPointerStruct) Init(ctx context.Context) error {
	n.InitCalled = true
	n.Name = "initialized"
	return nil
}

// ContainerWithNilPointer contains nil pointer fields
type ContainerWithNilPointer struct {
	ValidPtr   *NilPointerStruct
	NilPtr     *NilPointerStruct  // This will be nil
	Initialized bool
}

func (c *ContainerWithNilPointer) Init(ctx context.Context) error {
	c.Initialized = true
	return nil
}

// Test that nil pointers are skipped and don't cause panics
func TestNilPointerSkipping(t *testing.T) {
	container := &ContainerWithNilPointer{
		ValidPtr: &NilPointerStruct{Name: "valid"},
		NilPtr:   nil, // Explicitly nil
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, container)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Container should be initialized
	if !container.Initialized {
		t.Error("container was not initialized")
	}
	
	// Valid pointer should be initialized
	if !container.ValidPtr.InitCalled {
		t.Error("ValidPtr was not initialized")
	}
	
	if container.ValidPtr.Name != "initialized" {
		t.Errorf("ValidPtr.Name = %s; want 'initialized'", container.ValidPtr.Name)
	}
	
	// Nil pointer should remain nil
	if container.NilPtr != nil {
		t.Error("NilPtr should remain nil")
	}
}

// AllNil struct for testing
type AllNil struct {
	Ptr1 *NilPointerStruct
	Ptr2 *NilPointerStruct
	Ptr3 *NilPointerStruct
	Initialized bool
}

func (a *AllNil) Init(ctx context.Context) error {
	a.Initialized = true
	return nil
}

// Test struct with all nil pointers
func TestAllNilPointers(t *testing.T) {
	container := &AllNil{
		// All pointers are nil by default
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, container)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Even with all nil pointers, the container itself should initialize
	if !container.Initialized {
		t.Error("container should be initialized even with all nil pointer fields")
	}
	
	// All pointers should remain nil
	if container.Ptr1 != nil || container.Ptr2 != nil || container.Ptr3 != nil {
		t.Error("nil pointers should remain nil")
	}
}

// Test nested nil pointers
func TestNestedNilPointers(t *testing.T) {
	type Nested struct {
		DeepPtr *NilPointerStruct
	}
	
	type Root struct {
		Level1 *Nested
		Initialized bool
	}
	
	// Case 1: Level1 is nil
	root1 := &Root{
		Level1: nil,
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, root1)
	
	if err != nil {
		t.Fatalf("unexpected error with nil Level1: %v", err)
	}
	
	if root1.Level1 != nil {
		t.Error("nil Level1 should remain nil")
	}
	
	// Case 2: Level1 exists but DeepPtr is nil
	root2 := &Root{
		Level1: &Nested{
			DeepPtr: nil,
		},
	}
	
	err = AutoInit(ctx, root2)
	
	if err != nil {
		t.Fatalf("unexpected error with nil DeepPtr: %v", err)
	}
	
	if root2.Level1.DeepPtr != nil {
		t.Error("nil DeepPtr should remain nil")
	}
}

// Test that nil pointers in slices are handled
func TestNilPointersInSlice(t *testing.T) {
	type Container struct {
		Items []*NilPointerStruct
		Initialized bool
	}
	
	container := &Container{
		Items: []*NilPointerStruct{
			{Name: "first"},
			nil, // nil in the middle
			{Name: "third"},
			nil, // nil at the end
		},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, container)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check non-nil items are initialized
	if !container.Items[0].InitCalled {
		t.Error("Items[0] should be initialized")
	}
	
	if container.Items[1] != nil {
		t.Error("Items[1] should remain nil")
	}
	
	if !container.Items[2].InitCalled {
		t.Error("Items[2] should be initialized")
	}
	
	if container.Items[3] != nil {
		t.Error("Items[3] should remain nil")
	}
}

// Test that nil pointers in maps are handled
func TestNilPointersInMap(t *testing.T) {
	type Container struct {
		Items map[string]*NilPointerStruct
		Initialized bool
	}
	
	container := &Container{
		Items: map[string]*NilPointerStruct{
			"valid": {Name: "valid"},
			"nil":   nil,
			"another": {Name: "another"},
		},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, container)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check non-nil items are initialized
	if !container.Items["valid"].InitCalled {
		t.Error("Items['valid'] should be initialized")
	}
	
	if container.Items["nil"] != nil {
		t.Error("Items['nil'] should remain nil")
	}
	
	if !container.Items["another"].InitCalled {
		t.Error("Items['another'] should be initialized")
	}
}