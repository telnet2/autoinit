package autoinit

import (
	"context"
	"testing"
)

// Test that hooks can modify fields
type ModifiableChild struct {
	Name  string
	Value int
}

func (m *ModifiableChild) Init(ctx context.Context) error {
	// Always set to 100 during Init, regardless of current value
	m.Value = 100
	return nil
}

type ParentThatModifies struct {
	Child1 ModifiableChild
	Child2 *ModifiableChild
}

func (p *ParentThatModifies) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// Modify the field before it's initialized
	switch fieldName {
	case "Child1":
		if child, ok := fieldValue.(*ModifiableChild); ok {
			child.Name = "Modified by parent (pre)"
			child.Value = 50 // This should be overwritten by Init to 100
		}
	case "Child2":
		// For pointer fields, we receive a pointer to the struct, not a pointer to the pointer
		if child, ok := fieldValue.(*ModifiableChild); ok {
			child.Name = "Pointer modified by parent (pre)"
			child.Value = 75 // This should be overwritten by Init to 100
		}
	}
	return nil
}

func (p *ParentThatModifies) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// Modify the field after it's initialized
	switch fieldName {
	case "Child1":
		if child, ok := fieldValue.(*ModifiableChild); ok {
			child.Name = child.Name + " (post)"
			child.Value = child.Value + 1 // Should become 101
		}
	case "Child2":
		// For pointer fields, we receive a pointer to the struct, not a pointer to the pointer
		if child, ok := fieldValue.(*ModifiableChild); ok {
			child.Name = child.Name + " (post)"
			child.Value = child.Value + 2 // Should become 102
		}
	}
	return nil
}

func TestHooksCanModifyFields(t *testing.T) {
	parent := &ParentThatModifies{
		Child2: &ModifiableChild{},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, parent)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check Child1 (value field)
	if parent.Child1.Name != "Modified by parent (pre) (post)" {
		t.Errorf("Child1.Name = %q; want %q", parent.Child1.Name, "Modified by parent (pre) (post)")
	}
	if parent.Child1.Value != 101 {
		t.Errorf("Child1.Value = %d; want %d", parent.Child1.Value, 101)
	}
	
	// Check Child2 (pointer field)
	if parent.Child2.Name != "Pointer modified by parent (pre) (post)" {
		t.Errorf("Child2.Name = %q; want %q", parent.Child2.Name, "Pointer modified by parent (pre) (post)")
	}
	if parent.Child2.Value != 102 {
		t.Errorf("Child2.Value = %d; want %d", parent.Child2.Value, 102)
	}
}

// Test that PreInit can prepare state that Init uses
type StatefulChild struct {
	Prepared bool
	Ready    bool
}

func (s *StatefulChild) PreInit(ctx context.Context) error {
	s.Prepared = true
	return nil
}

func (s *StatefulChild) Init(ctx context.Context) error {
	if !s.Prepared {
		// This shouldn't happen if PreInit was called
		panic("PreInit was not called before Init")
	}
	s.Ready = true
	return nil
}

type ParentWithStatefulChild struct {
	Child StatefulChild
}

func (p *ParentWithStatefulChild) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// We can verify the child hasn't been initialized yet
	if child, ok := fieldValue.(*StatefulChild); ok {
		if child.Prepared || child.Ready {
			// This shouldn't happen - PreInit hasn't been called yet
			panic("Child was already initialized in PreFieldInit")
		}
	}
	return nil
}

func (p *ParentWithStatefulChild) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// We can verify the child has been fully initialized
	if child, ok := fieldValue.(*StatefulChild); ok {
		if !child.Prepared || !child.Ready {
			// This shouldn't happen - child should be fully initialized
			panic("Child was not fully initialized in PostFieldInit")
		}
	}
	return nil
}

func TestHookTiming(t *testing.T) {
	parent := &ParentWithStatefulChild{}
	
	ctx := context.Background()
	err := AutoInit(ctx, parent)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !parent.Child.Prepared {
		t.Error("Child.Prepared should be true")
	}
	if !parent.Child.Ready {
		t.Error("Child.Ready should be true")
	}
}