package autoinit

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// Test structs for different interface implementations

// SimpleOnly implements only Init()
type SimpleOnly struct {
	Name        string
	Initialized bool
}

func (s *SimpleOnly) Init() error {
	s.Initialized = true
	s.Name = "simple"
	return nil
}

// ContextOnly implements only Init(ctx)
type ContextOnly struct {
	Name        string
	Initialized bool
	CtxValue    string
}

func (c *ContextOnly) Init(ctx context.Context) error {
	c.Initialized = true
	c.Name = "context"
	// Try to get a value from context
	if val := ctx.Value("test-key"); val != nil {
		c.CtxValue = val.(string)
	}
	return nil
}

// ParentOnly implements only Init(ctx, parent)
type ParentOnly struct {
	Name       string
	ParentName string
	Initialized bool
}

func (p *ParentOnly) Init(ctx context.Context, parent interface{}) error {
	p.Initialized = true
	p.Name = "with-parent"
	
	// Extract parent name if possible
	if parent != nil {
		if parentStruct, ok := parent.(*MixedContainer); ok {
			p.ParentName = parentStruct.Name
		}
	}
	return nil
}

// AllInterfaces implements the highest priority interface (ParentInitializer)
type AllInterfaces struct {
	Name        string
	Method      string
	Initialized bool
}

// Only implement the highest priority interface
func (a *AllInterfaces) Init(ctx context.Context, parent interface{}) error {
	a.Initialized = true
	a.Method = "parent"
	return nil
}

// MixedContainer contains fields with different interface implementations
type MixedContainer struct {
	Name        string
	Simple      SimpleOnly
	Context     ContextOnly
	Parent      ParentOnly
	All         AllInterfaces
	Initialized bool
}

func (m *MixedContainer) Init(ctx context.Context) error {
	m.Initialized = true
	return nil
}

// NestedWithParent uses parent reference to access parent struct
type NestedWithParent struct {
	ChildData   string
	ParentType  string
	ParentField string
}

func (n *NestedWithParent) Init(ctx context.Context, parent interface{}) error {
	n.ChildData = "child-initialized"
	
	// Access parent struct (the parent's Init hasn't been called yet)
	if p, ok := parent.(*ParentAware); ok {
		n.ParentType = "ParentAware"
		// Access a field that was set during struct creation, not in Init
		n.ParentField = p.PresetData
	}
	return nil
}

// ParentAware contains a child that needs parent reference
type ParentAware struct {
	PresetData string // This is set before Init
	ParentData string // This is set in Init
	Child      NestedWithParent
}

func (p *ParentAware) Init(ctx context.Context) error {
	p.ParentData = "parent-data"
	return nil
}

// Test simple Init() interface
func TestSimpleInitInterface(t *testing.T) {
	component := &SimpleOnly{}
	ctx := context.Background()
	
	err := AutoInit(ctx, component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if component.Name != "simple" {
		t.Errorf("expected Name to be 'simple', got '%s'", component.Name)
	}
}

// Test context Init(ctx) interface
func TestContextInitInterface(t *testing.T) {
	component := &ContextOnly{}
	ctx := context.WithValue(context.Background(), "test-key", "test-value")
	
	err := AutoInit(ctx, component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if component.Name != "context" {
		t.Errorf("expected Name to be 'context', got '%s'", component.Name)
	}
	
	if component.CtxValue != "test-value" {
		t.Errorf("expected CtxValue to be 'test-value', got '%s'", component.CtxValue)
	}
}

// Test parent Init(ctx, parent) interface
func TestParentInitInterface(t *testing.T) {
	container := &MixedContainer{
		Name: "parent-container",
	}
	ctx := context.Background()
	
	err := AutoInit(ctx, container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !container.Parent.Initialized {
		t.Error("Parent field was not initialized")
	}
	
	if container.Parent.Name != "with-parent" {
		t.Errorf("expected Parent.Name to be 'with-parent', got '%s'", container.Parent.Name)
	}
	
	if container.Parent.ParentName != "parent-container" {
		t.Errorf("expected Parent.ParentName to be 'parent-container', got '%s'", container.Parent.ParentName)
	}
}

// Test mixed interfaces in one struct
func TestMixedInterfaces(t *testing.T) {
	container := &MixedContainer{
		Name: "mixed-container",
	}
	ctx := context.WithValue(context.Background(), "test-key", "ctx-value")
	
	err := AutoInit(ctx, container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check all fields are initialized
	if !container.Initialized {
		t.Error("container was not initialized")
	}
	
	if !container.Simple.Initialized {
		t.Error("Simple field was not initialized")
	}
	
	if !container.Context.Initialized {
		t.Error("Context field was not initialized")
	}
	
	if container.Context.CtxValue != "ctx-value" {
		t.Errorf("Context field didn't receive context value")
	}
	
	if !container.Parent.Initialized {
		t.Error("Parent field was not initialized")
	}
	
	if container.Parent.ParentName != "mixed-container" {
		t.Errorf("Parent field didn't receive parent reference")
	}
	
	// Check that AllInterfaces used the highest priority (parent) interface
	if !container.All.Initialized {
		t.Error("All field was not initialized")
	}
	
	if container.All.Method != "parent" {
		t.Errorf("expected All.Method to be 'parent' (highest priority), got '%s'", container.All.Method)
	}
}

// Test parent reference propagation through nested structs
func TestParentReferencePropagation(t *testing.T) {
	component := &ParentAware{
		PresetData: "preset-value",
	}
	ctx := context.Background()
	
	err := AutoInit(ctx, component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if component.ParentData != "parent-data" {
		t.Errorf("expected ParentData to be 'parent-data', got '%s'", component.ParentData)
	}
	
	if component.Child.ChildData != "child-initialized" {
		t.Errorf("expected Child.ChildData to be 'child-initialized', got '%s'", component.Child.ChildData)
	}
	
	// Child should have received parent type
	if component.Child.ParentType != "ParentAware" {
		t.Errorf("expected Child.ParentType to be 'ParentAware', got '%s'", component.Child.ParentType)
	}
	
	// Child should have access to parent's preset field
	if component.Child.ParentField != "preset-value" {
		t.Errorf("expected Child.ParentField to be 'preset-value', got '%s'", component.Child.ParentField)
	}
}

// FailingSimple for error testing
type FailingSimple struct {
	ShouldFail bool
}

func (f *FailingSimple) Init() error {
	if f.ShouldFail {
		return fmt.Errorf("simple init failed")
	}
	return nil
}

// Test error in simple Init()
func TestSimpleInitError(t *testing.T) {
	component := &struct {
		Field FailingSimple
	}{
		Field: FailingSimple{ShouldFail: true},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, component)
	
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	
	if !strings.Contains(err.Error(), "simple init failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// SlowInit for context cancellation testing
type SlowInit struct {
	Initialized bool
}

func (s *SlowInit) Init(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.Initialized = true
		return nil
	}
}

// Test context cancellation
func TestContextCancellation(t *testing.T) {
	component := &SlowInit{}
	
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	err := AutoInit(ctx, component)
	
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}

// RootStruct for testing nil parent
type RootStruct struct {
	ReceivedNilParent bool
}

func (r *RootStruct) Init(ctx context.Context, parent interface{}) error {
	r.ReceivedNilParent = (parent == nil)
	return nil
}

// Test with nil parent (root struct)
func TestNilParentForRoot(t *testing.T) {
	component := &RootStruct{}
	
	ctx := context.Background()
	err := AutoInit(ctx, component)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.ReceivedNilParent {
		t.Error("root struct should receive nil parent")
	}
}