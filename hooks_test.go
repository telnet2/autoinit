package autoinit

import (
	"context"
	"fmt"
	"testing"
)

// Test types for hook functionality
type ChildWithHooks struct {
	Name       string
	PreCalled  bool
	PostCalled bool
	InitCalled bool
}

func (c *ChildWithHooks) PreInit(ctx context.Context) error {
	c.PreCalled = true
	c.Name = "pre-initialized"
	return nil
}

func (c *ChildWithHooks) Init(ctx context.Context) error {
	c.InitCalled = true
	c.Name = "initialized"
	return nil
}

func (c *ChildWithHooks) PostInit(ctx context.Context) error {
	c.PostCalled = true
	c.Name = "post-initialized"
	return nil
}

// State capture for testing - no Init methods so it won't be initialized
type ChildState struct {
	Name       string
	PreCalled  bool
	PostCalled bool
	InitCalled bool
}

// Parent with field hooks
type ParentWithFieldHooks struct {
	Child          ChildWithHooks
	PreFieldCalls  []string
	PostFieldCalls []string
	PreChildState  *ChildState // Capture state at pre-hook time
	PostChildState *ChildState // Capture state at post-hook time
}

func (p *ParentWithFieldHooks) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PreFieldCalls = append(p.PreFieldCalls, fieldName)

	// Capture the state of the child at pre-hook time
	if fieldName == "Child" {
		if child, ok := fieldValue.(*ChildWithHooks); ok {
			p.PreChildState = &ChildState{
				Name:       child.Name,
				PreCalled:  child.PreCalled,
				PostCalled: child.PostCalled,
				InitCalled: child.InitCalled,
			}
		}
	}
	return nil
}

func (p *ParentWithFieldHooks) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PostFieldCalls = append(p.PostFieldCalls, fieldName)

	// Capture the state of the child at post-hook time
	if fieldName == "Child" {
		if child, ok := fieldValue.(*ChildWithHooks); ok {
			p.PostChildState = &ChildState{
				Name:       child.Name,
				PreCalled:  child.PreCalled,
				PostCalled: child.PostCalled,
				InitCalled: child.InitCalled,
			}
		}
	}
	return nil
}

// Test PreInit and PostInit hooks
func TestPrePostInitHooks(t *testing.T) {
	child := &ChildWithHooks{}

	ctx := context.Background()
	err := AutoInit(ctx, child)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !child.PreCalled {
		t.Error("PreInit was not called")
	}

	if !child.InitCalled {
		t.Error("Init was not called")
	}

	if !child.PostCalled {
		t.Error("PostInit was not called")
	}

	// PostInit should run last, so final name should be "post-initialized"
	if child.Name != "post-initialized" {
		t.Errorf("Expected Name to be 'post-initialized', got '%s'", child.Name)
	}
}

// Test parent field hooks
func TestParentFieldHooks(t *testing.T) {
	parent := &ParentWithFieldHooks{}

	ctx := context.Background()
	err := AutoInit(ctx, parent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that PreFieldInit was called for Child field
	found := false
	for _, call := range parent.PreFieldCalls {
		if call == "Child" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("PreFieldInit not called for Child field: %v", parent.PreFieldCalls)
	}

	// Check that PostFieldInit was called for Child field
	found = false
	for _, call := range parent.PostFieldCalls {
		if call == "Child" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("PostFieldInit not called for Child field: %v", parent.PostFieldCalls)
	}

	// Verify field was initialized
	if !parent.Child.InitCalled {
		t.Error("Child's Init was not called")
	}

	// Check the captured states
	if parent.PreChildState == nil {
		t.Fatal("PreChildState was not captured")
	}
	if parent.PostChildState == nil {
		t.Fatal("PostChildState was not captured")
	}

	// At pre-hook time, child should not be initialized yet
	// Note: The child's PreInit might have been called already since we're passing the address
	if parent.PreChildState.InitCalled {
		t.Error("Child's Init() was already called at PreFieldInit time")
	}
	if parent.PreChildState.PreCalled {
		t.Error("Child's PreInit was already called at PreFieldInit time")
	}

	// At post-hook time, child should be fully initialized
	if !parent.PostChildState.InitCalled {
		t.Error("Child was not initialized at PostFieldInit time")
	}
	if !parent.PostChildState.PreCalled {
		t.Error("Child's PreInit was not called at PostFieldInit time")
	}
	if !parent.PostChildState.PostCalled {
		t.Error("Child's PostInit was not called at PostFieldInit time")
	}
}

// Test with pointer fields
type ParentWithPointerField struct {
	Child          *ChildWithHooks
	PreFieldCalls  []string
	PostFieldCalls []string
}

func (p *ParentWithPointerField) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PreFieldCalls = append(p.PreFieldCalls, fieldName)
	return nil
}

func (p *ParentWithPointerField) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PostFieldCalls = append(p.PostFieldCalls, fieldName)
	return nil
}

func TestFieldHooksWithPointers(t *testing.T) {
	parent := &ParentWithPointerField{
		Child: &ChildWithHooks{},
	}

	ctx := context.Background()
	err := AutoInit(ctx, parent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check hooks were called
	if len(parent.PreFieldCalls) != 1 || parent.PreFieldCalls[0] != "Child" {
		t.Errorf("PreFieldInit not called correctly: %v", parent.PreFieldCalls)
	}

	if len(parent.PostFieldCalls) != 1 || parent.PostFieldCalls[0] != "Child" {
		t.Errorf("PostFieldInit not called correctly: %v", parent.PostFieldCalls)
	}

	// Verify child was initialized
	if !parent.Child.InitCalled {
		t.Error("Child's Init was not called")
	}
}

// Test hook errors
type FailingPreInit struct {
	PreCalled bool
}

func (f *FailingPreInit) PreInit(ctx context.Context) error {
	f.PreCalled = true
	return fmt.Errorf("PreInit failed")
}

func TestPreInitError(t *testing.T) {
	obj := &FailingPreInit{}

	ctx := context.Background()
	err := AutoInit(ctx, obj)

	if err == nil {
		t.Fatal("expected error from PreInit")
	}

	if !obj.PreCalled {
		t.Error("PreInit was not called")
	}

	// The error should be wrapped in InitError
	if initErr, ok := err.(*InitError); ok {
		if initErr.Unwrap().Error() != "PreInit failed" {
			t.Errorf("unexpected error cause: %v", initErr.Unwrap())
		}
	} else {
		t.Errorf("expected InitError, got %T: %v", err, err)
	}
}

// Test only PreFieldHook (not PostFieldHook)
type ParentWithOnlyPreHook struct {
	Child         ChildWithHooks
	PreFieldCalls []string
}

func (p *ParentWithOnlyPreHook) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PreFieldCalls = append(p.PreFieldCalls, fieldName)
	return nil
}

func TestOnlyPreFieldHook(t *testing.T) {
	parent := &ParentWithOnlyPreHook{}

	ctx := context.Background()
	err := AutoInit(ctx, parent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that PreFieldInit was called
	if len(parent.PreFieldCalls) != 1 || parent.PreFieldCalls[0] != "Child" {
		t.Errorf("PreFieldInit not called correctly: %v", parent.PreFieldCalls)
	}

	// Verify field was still initialized
	if !parent.Child.InitCalled {
		t.Error("Child's Init was not called")
	}
}

// Test only PostFieldHook (not PreFieldHook)
type ParentWithOnlyPostHook struct {
	Child          ChildWithHooks
	PostFieldCalls []string
}

func (p *ParentWithOnlyPostHook) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PostFieldCalls = append(p.PostFieldCalls, fieldName)
	return nil
}

func TestOnlyPostFieldHook(t *testing.T) {
	parent := &ParentWithOnlyPostHook{}

	ctx := context.Background()
	err := AutoInit(ctx, parent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that PostFieldInit was called
	if len(parent.PostFieldCalls) != 1 || parent.PostFieldCalls[0] != "Child" {
		t.Errorf("PostFieldInit not called correctly: %v", parent.PostFieldCalls)
	}

	// Verify field was still initialized
	if !parent.Child.InitCalled {
		t.Error("Child's Init was not called")
	}
}

// Test nested structs with hooks
type GrandParent struct {
	Parent         ParentWithFieldHooks
	PreFieldCalls  []string
	PostFieldCalls []string
}

func (g *GrandParent) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	g.PreFieldCalls = append(g.PreFieldCalls, fieldName)
	return nil
}

func (g *GrandParent) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	g.PostFieldCalls = append(g.PostFieldCalls, fieldName)
	return nil
}

func TestNestedHooks(t *testing.T) {
	grandParent := &GrandParent{}

	ctx := context.Background()
	err := AutoInit(ctx, grandParent)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// GrandParent hooks should be called for Parent field
	if len(grandParent.PreFieldCalls) != 1 || grandParent.PreFieldCalls[0] != "Parent" {
		t.Errorf("GrandParent PreFieldInit not called correctly: %v", grandParent.PreFieldCalls)
	}

	if len(grandParent.PostFieldCalls) != 1 || grandParent.PostFieldCalls[0] != "Parent" {
		t.Errorf("GrandParent PostFieldInit not called correctly: %v", grandParent.PostFieldCalls)
	}

	// Parent hooks should be called for Child field
	found := false
	for _, call := range grandParent.Parent.PreFieldCalls {
		if call == "Child" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Parent PreFieldInit not called for Child field: %v", grandParent.Parent.PreFieldCalls)
	}

	found = false
	for _, call := range grandParent.Parent.PostFieldCalls {
		if call == "Child" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Parent PostFieldInit not called for Child field: %v", grandParent.Parent.PostFieldCalls)
	}

	// Child should be initialized
	if !grandParent.Parent.Child.InitCalled {
		t.Error("Child's Init was not called")
	}
}
