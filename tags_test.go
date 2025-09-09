package autoinit

import (
	"context"
	"testing"
)

// Test types for tag-based initialization
type TaggedStruct struct {
	// Fields with explicit tags
	InitMe    *SimpleComponent `autoinit:"init"`    // Should initialize
	AlsoInit  *SimpleComponent `autoinit:""`        // Empty tag also means init
	SkipMe    *SimpleComponent `autoinit:"-"`       // Explicitly skip
	
	// Field without tag
	NoTag     *SimpleComponent                      // Behavior depends on RequireTags
	
	// Test if nested structs respect tags
	Nested    NestedWithTags  `autoinit:"init"`
	
	// Collections
	Services  []*SimpleComponent `autoinit:"init"`
	SkipList  []*SimpleComponent `autoinit:"-"`
}

type NestedWithTags struct {
	// This field should be initialized when parent is initialized
	Inner *SimpleComponent
	Initialized bool
}

func (n *NestedWithTags) Init(ctx context.Context) error {
	n.Initialized = true
	return nil
}

// Test default behavior (RequireTags = false)
func TestTagsDefaultBehavior(t *testing.T) {
	s := &TaggedStruct{
		InitMe:   &SimpleComponent{Name: "init-me"},
		AlsoInit: &SimpleComponent{Name: "also-init"},
		SkipMe:   &SimpleComponent{Name: "skip-me"},
		NoTag:    &SimpleComponent{Name: "no-tag"},
		Nested: NestedWithTags{
			Inner: &SimpleComponent{Name: "inner"},
		},
		Services: []*SimpleComponent{
			{Name: "service1"},
			{Name: "service2"},
		},
		SkipList: []*SimpleComponent{
			{Name: "skip1"},
			{Name: "skip2"},
		},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, s)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check fields with init tags were initialized
	if !s.InitMe.Initialized {
		t.Error("InitMe should be initialized")
	}
	if !s.AlsoInit.Initialized {
		t.Error("AlsoInit should be initialized")
	}
	
	// Check field with skip tag was NOT initialized
	if s.SkipMe.Initialized {
		t.Error("SkipMe should NOT be initialized (has autoinit:\"-\" tag)")
	}
	
	// Check field without tag WAS initialized (default behavior)
	if !s.NoTag.Initialized {
		t.Error("NoTag should be initialized (default behavior when RequireTags=false)")
	}
	
	// Check nested struct was initialized
	if !s.Nested.Initialized {
		t.Error("Nested should be initialized")
	}
	if !s.Nested.Inner.Initialized {
		t.Error("Nested.Inner should be initialized")
	}
	
	// Check collections
	for i, svc := range s.Services {
		if !svc.Initialized {
			t.Errorf("Services[%d] should be initialized", i)
		}
	}
	
	for i, skip := range s.SkipList {
		if skip.Initialized {
			t.Errorf("SkipList[%d] should NOT be initialized (has autoinit:\"-\" tag)", i)
		}
	}
}

// Test with RequireTags = true
func TestRequireTagsBehavior(t *testing.T) {
	s := &TaggedStruct{
		InitMe:   &SimpleComponent{Name: "init-me"},
		AlsoInit: &SimpleComponent{Name: "also-init"},
		SkipMe:   &SimpleComponent{Name: "skip-me"},
		NoTag:    &SimpleComponent{Name: "no-tag"},
		Nested: NestedWithTags{
			Inner: &SimpleComponent{Name: "inner"},
		},
		Services: []*SimpleComponent{
			{Name: "service1"},
			{Name: "service2"},
		},
		SkipList: []*SimpleComponent{
			{Name: "skip1"},
			{Name: "skip2"},
		},
	}
	
	ctx := context.Background()
	options := &Options{
		RequireTags: true,
	}
	err := AutoInitWithOptions(ctx, s, options)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check fields with init tags were initialized
	if !s.InitMe.Initialized {
		t.Error("InitMe should be initialized (has autoinit tag)")
	}
	if !s.AlsoInit.Initialized {
		t.Error("AlsoInit should be initialized (has empty autoinit tag)")
	}
	
	// Check field with skip tag was NOT initialized
	if s.SkipMe.Initialized {
		t.Error("SkipMe should NOT be initialized (has autoinit:\"-\" tag)")
	}
	
	// Check field without tag was NOT initialized (RequireTags=true)
	if s.NoTag.Initialized {
		t.Error("NoTag should NOT be initialized (no tag with RequireTags=true)")
	}
	
	// Check nested struct was initialized (parent has tag)
	if !s.Nested.Initialized {
		t.Error("Nested should be initialized (has autoinit tag)")
	}
	// Inner field doesn't have a tag, so it won't be initialized when RequireTags=true
	if s.Nested.Inner.Initialized {
		t.Error("Nested.Inner should NOT be initialized (no tag with RequireTags=true)")
	}
	
	// Check collections
	for i, svc := range s.Services {
		if !svc.Initialized {
			t.Errorf("Services[%d] should be initialized (field has tag)", i)
		}
	}
	
	for i, skip := range s.SkipList {
		if skip.Initialized {
			t.Errorf("SkipList[%d] should NOT be initialized (has autoinit:\"-\" tag)", i)
		}
	}
}

// Test mixed structs with and without tags
type MixedTagParent struct {
	// When RequireTags=true, only these should be processed
	Tagged   TaggedChild   `autoinit:"init"`
	Untagged UntaggedChild // No tag, should be skipped when RequireTags=true
}

type TaggedChild struct {
	Value       string
	Initialized bool
}

func (t *TaggedChild) Init(ctx context.Context) error {
	t.Initialized = true
	return nil
}

type UntaggedChild struct {
	Value       string
	Initialized bool
}

func (u *UntaggedChild) Init(ctx context.Context) error {
	u.Initialized = true
	return nil
}

func TestMixedTagsWithRequireTags(t *testing.T) {
	parent := &MixedTagParent{
		Tagged:   TaggedChild{Value: "tagged"},
		Untagged: UntaggedChild{Value: "untagged"},
	}
	
	ctx := context.Background()
	options := &Options{
		RequireTags: true,
	}
	err := AutoInitWithOptions(ctx, parent, options)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Tagged child should be initialized
	if !parent.Tagged.Initialized {
		t.Error("Tagged child should be initialized")
	}
	
	// Untagged child should NOT be initialized when RequireTags=true
	if parent.Untagged.Initialized {
		t.Error("Untagged child should NOT be initialized with RequireTags=true")
	}
}

// Test that nested struct fields are still processed even when parent has RequireTags
type ParentWithRequireTags struct {
	Child ChildWithMixedFields `autoinit:"init"`
}

type ChildWithMixedFields struct {
	// These fields don't have tags, but should still be processed
	// because the parent field has a tag
	Service1 *SimpleComponent
	Service2 *SimpleComponent
	
	Initialized bool
}

func (c *ChildWithMixedFields) Init(ctx context.Context) error {
	c.Initialized = true
	return nil
}

func TestNestedFieldsWithParentTag(t *testing.T) {
	parent := &ParentWithRequireTags{
		Child: ChildWithMixedFields{
			Service1: &SimpleComponent{Name: "svc1"},
			Service2: &SimpleComponent{Name: "svc2"},
		},
	}
	
	ctx := context.Background()
	options := &Options{
		RequireTags: true,
	}
	err := AutoInitWithOptions(ctx, parent, options)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Child should be initialized (has tag)
	if !parent.Child.Initialized {
		t.Error("Child should be initialized")
	}
	
	// Child's fields won't be initialized because they don't have tags
	// and RequireTags applies throughout the tree
	if parent.Child.Service1.Initialized {
		t.Error("Child.Service1 should NOT be initialized (no tag with RequireTags=true)")
	}
	if parent.Child.Service2.Initialized {
		t.Error("Child.Service2 should NOT be initialized (no tag with RequireTags=true)")
	}
}