package autoinit

import (
	"context"
	"testing"
)

// SimpleInit for testing
type SimpleInit struct {
	Name       string
	Initialized bool
}

func (s *SimpleInit) Init(ctx context.Context) error {
	s.Initialized = true
	return nil
}

// Parent with collection fields
type ParentWithCollections struct {
	// Collection fields
	MapField   map[string]*SimpleInit
	SliceField []*SimpleInit
	ArrayField [2]*SimpleInit
	
	// Track what hooks were called
	PreFieldCalls  []string
	PostFieldCalls []string
}

func (p *ParentWithCollections) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PreFieldCalls = append(p.PreFieldCalls, fieldName)
	return nil
}

func (p *ParentWithCollections) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	p.PostFieldCalls = append(p.PostFieldCalls, fieldName)
	return nil
}

func TestHooksWithCollections(t *testing.T) {
	parent := &ParentWithCollections{
		MapField: map[string]*SimpleInit{
			"first":  {Name: "first"},
			"second": {Name: "second"},
		},
		SliceField: []*SimpleInit{
			{Name: "slice1"},
			{Name: "slice2"},
		},
		ArrayField: [2]*SimpleInit{
			{Name: "array1"},
			{Name: "array2"},
		},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, parent)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check if map values were initialized
	for key, val := range parent.MapField {
		if !val.Initialized {
			t.Errorf("MapField[%s] was not initialized", key)
		}
	}
	
	// Check if slice elements were initialized
	for i, val := range parent.SliceField {
		if !val.Initialized {
			t.Errorf("SliceField[%d] was not initialized", i)
		}
	}
	
	// Check if array elements were initialized
	for i, val := range parent.ArrayField {
		if !val.Initialized {
			t.Errorf("ArrayField[%d] was not initialized", i)
		}
	}
	
	// Check which hooks were called
	t.Logf("PreFieldInit calls: %v", parent.PreFieldCalls)
	t.Logf("PostFieldInit calls: %v", parent.PostFieldCalls)
	
	// Now hooks SHOULD be called for collections (the collection itself, not individual elements)
	expectedFields := []string{"MapField", "SliceField", "ArrayField"}
	
	if len(parent.PreFieldCalls) != 3 {
		t.Errorf("PreFieldInit was called %d times; want 3", len(parent.PreFieldCalls))
	}
	if len(parent.PostFieldCalls) != 3 {
		t.Errorf("PostFieldInit was called %d times; want 3", len(parent.PostFieldCalls))
	}
	
	// Check that all expected fields had hooks called
	for _, field := range expectedFields {
		found := false
		for _, call := range parent.PreFieldCalls {
			if call == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PreFieldInit was not called for %s", field)
		}
		
		found = false
		for _, call := range parent.PostFieldCalls {
			if call == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PostFieldInit was not called for %s", field)
		}
	}
}

// Test with struct field alongside collections
type MixedFieldsParent struct {
	StructField SimpleInit
	MapField    map[string]*SimpleInit
	
	PreFieldCalls  []string
	PostFieldCalls []string
}

func (m *MixedFieldsParent) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	m.PreFieldCalls = append(m.PreFieldCalls, fieldName)
	return nil
}

func (m *MixedFieldsParent) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	m.PostFieldCalls = append(m.PostFieldCalls, fieldName)
	return nil
}

func TestHooksMixedFields(t *testing.T) {
	parent := &MixedFieldsParent{
		MapField: map[string]*SimpleInit{
			"item": {Name: "item"},
		},
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, parent)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check struct field was initialized
	if !parent.StructField.Initialized {
		t.Error("StructField was not initialized")
	}
	
	// Check map values were initialized
	if !parent.MapField["item"].Initialized {
		t.Error("MapField[item] was not initialized")
	}
	
	// Hooks should be called for both StructField and MapField
	expectedPre := []string{"StructField", "MapField"}
	expectedPost := []string{"StructField", "MapField"}
	
	if len(parent.PreFieldCalls) != 2 {
		t.Errorf("PreFieldInit was called %d times; want 2", len(parent.PreFieldCalls))
	}
	if len(parent.PostFieldCalls) != 2 {
		t.Errorf("PostFieldInit was called %d times; want 2", len(parent.PostFieldCalls))
	}
	
	// Check both fields had hooks called
	for _, field := range expectedPre {
		found := false
		for _, call := range parent.PreFieldCalls {
			if call == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PreFieldInit was not called for %s", field)
		}
	}
	
	for _, field := range expectedPost {
		found := false
		for _, call := range parent.PostFieldCalls {
			if call == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PostFieldInit was not called for %s", field)
		}
	}
}