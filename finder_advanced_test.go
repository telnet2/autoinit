package autoinit_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/telnet2/autoinit"
)

// Interface for testing interface-based search
type DataProvider interface {
	GetData() string
}

// Value component that implements DataProvider
type ValueComponent struct {
	Name        string
	Initialized bool
}

func (v *ValueComponent) Init(ctx context.Context) error {
	v.Initialized = true
	return nil
}

func (v *ValueComponent) GetData() string {
	return v.Name
}

// Pointer component that implements DataProvider
type PointerComponent struct {
	Name        string
	Initialized bool
}

func (p *PointerComponent) Init(ctx context.Context) error {
	p.Initialized = true
	return nil
}

func (p *PointerComponent) GetData() string {
	return p.Name
}

// Component that searches for others
type SearcherComponent struct {
	foundValue     *ValueComponent
	foundPointer   *PointerComponent
	foundInterface DataProvider
	foundNilField  interface{}
}

func (s *SearcherComponent) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, s, parent)

	// Try to find value component (should get pointer to it)
	if val := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*ValueComponent)(nil)),
	}); val != nil {
		s.foundValue = val.(*ValueComponent)
	}

	// Try to find pointer component
	if ptr := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*PointerComponent)(nil)),
	}); ptr != nil {
		s.foundPointer = ptr.(*PointerComponent)
	}

	// Try to find by interface
	if provider := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*DataProvider)(nil)).Elem(),
	}); provider != nil {
		s.foundInterface = provider.(DataProvider)
	}

	// Try to find uninitialized (nil) field
	if nilField := finder.Find(autoinit.SearchOption{
		ByFieldName: "UninitializedPtr",
	}); nilField != nil {
		s.foundNilField = nilField
	}

	return nil
}

func TestFinderValueFieldPointer(t *testing.T) {
	type App struct {
		ValueComp   ValueComponent    // Value field
		PointerComp *PointerComponent // Pointer field
		Searcher    *SearcherComponent
	}

	app := &App{
		ValueComp:   ValueComponent{Name: "ValueData"},
		PointerComp: &PointerComponent{Name: "PointerData"},
		Searcher:    &SearcherComponent{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Check if finder found the value component (as a pointer)
	if app.Searcher.foundValue == nil {
		t.Error("Should have found ValueComponent")
	} else {
		// Verify it's actually a pointer to the value field
		if app.Searcher.foundValue != &app.ValueComp {
			t.Error("Should have found pointer to the actual ValueComp field")
		}
		if !app.Searcher.foundValue.Initialized {
			t.Error("ValueComponent should be initialized")
		}
	}

	// Check if finder found the pointer component
	if app.Searcher.foundPointer == nil {
		t.Error("Should have found PointerComponent")
	} else {
		if app.Searcher.foundPointer != app.PointerComp {
			t.Error("Should have found the actual PointerComp")
		}
		if !app.Searcher.foundPointer.Initialized {
			t.Error("PointerComponent should be initialized")
		}
	}
}

func TestFinderInterfaceSearch(t *testing.T) {
	type App struct {
		Provider1 *PointerComponent // Implements DataProvider
		Provider2 ValueComponent    // Also implements DataProvider
		Searcher  *SearcherComponent
	}

	app := &App{
		Provider1: &PointerComponent{Name: "Provider1"},
		Provider2: ValueComponent{Name: "Provider2"},
		Searcher:  &SearcherComponent{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Check if finder found something implementing the interface
	if app.Searcher.foundInterface == nil {
		t.Error("Should have found a DataProvider")
	} else {
		// It should find the first one (Provider1)
		data := app.Searcher.foundInterface.GetData()
		if data != "Provider1" && data != "Provider2" {
			t.Errorf("Should have found one of the providers, got: %s", data)
		}
	}
}

func TestFinderSkipsNilFields(t *testing.T) {
	type App struct {
		InitializedComp  *PointerComponent
		UninitializedPtr *PointerComponent // nil - should be skipped
		Searcher         *SearcherComponent
	}

	app := &App{
		InitializedComp: &PointerComponent{Name: "Initialized"},
		// UninitializedPtr is nil
		Searcher: &SearcherComponent{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Should not find the nil field
	if app.Searcher.foundNilField != nil {
		t.Error("Should not find nil/uninitialized fields")
	}

	// Should find the initialized field
	if app.Searcher.foundPointer == nil {
		t.Error("Should have found the initialized PointerComponent")
	} else if app.Searcher.foundPointer != app.InitializedComp {
		t.Error("Should have found the correct initialized component")
	}
}

// ModifierComponent modifies found components
type ModifierComponent struct {
	target *ValueComponent
}

func (m *ModifierComponent) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, m, parent)
	if val := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*ValueComponent)(nil)),
	}); val != nil {
		m.target = val.(*ValueComponent)
		// Modify the found component
		m.target.Name = "Modified"
	}
	return nil
}

// Test that finder gets pointer to value fields that can be modified
func TestFinderValueFieldModification(t *testing.T) {

	type App struct {
		Value    ValueComponent
		Modifier *ModifierComponent
	}

	app := &App{
		Value:    ValueComponent{Name: "Original"},
		Modifier: &ModifierComponent{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Check if the modification affected the original
	if app.Value.Name != "Modified" {
		t.Errorf("Expected value to be modified, got: %s", app.Value.Name)
	}
}

// Interface with different method signatures
type ConfigProvider interface {
	GetConfig() string
}

// Value receiver implementation
type ValueConfig struct {
	Config string
}

func (v ValueConfig) GetConfig() string {
	return v.Config
}

func (v *ValueConfig) Init(ctx context.Context) error {
	v.Config = "ValueConfig"
	return nil
}

// Pointer receiver implementation
type PointerConfig struct {
	Config string
}

func (p *PointerConfig) GetConfig() string {
	return p.Config
}

func (p *PointerConfig) Init(ctx context.Context) error {
	p.Config = "PointerConfig"
	return nil
}

// Searcher that looks for ConfigProvider
type ConfigSearcher struct {
	foundConfigs []ConfigProvider
}

func (c *ConfigSearcher) Init(ctx context.Context, parent interface{}) error {
	// In real implementation, we'd need to search all fields
	// For now, let's test if we can find by interface type
	finder := autoinit.NewComponentFinder(ctx, c, parent)

	// Try to find first ConfigProvider
	if provider := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*ConfigProvider)(nil)).Elem(),
	}); provider != nil {
		c.foundConfigs = append(c.foundConfigs, provider.(ConfigProvider))
	}

	return nil
}

// Test finding components by interface with both pointer and value receivers
func TestFinderInterfaceWithMixedReceivers(t *testing.T) {

	type App struct {
		ValConfig ValueConfig
		PtrConfig *PointerConfig
		Searcher  *ConfigSearcher
	}

	app := &App{
		ValConfig: ValueConfig{},
		PtrConfig: &PointerConfig{},
		Searcher:  &ConfigSearcher{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Should find at least one ConfigProvider
	if len(app.Searcher.foundConfigs) == 0 {
		t.Error("Should have found at least one ConfigProvider")
	}
}
