package autoinit

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Test structs for various scenarios

// Simple struct with Init method
type SimpleComponent struct {
	Name       string
	Initialized bool
}

func (s *SimpleComponent) Init(ctx context.Context) error {
	s.Initialized = true
	s.Name = "initialized"
	return nil
}

// Struct with nested components
type NestedComponent struct {
	ID         int
	SubComponent SimpleComponent
	Initialized bool
}

func (n *NestedComponent) Init(ctx context.Context) error {
	n.Initialized = true
	n.ID = 42
	return nil
}

// Struct with pointer fields
type PointerComponent struct {
	Sub        *SimpleComponent
	Nested     *NestedComponent
	Initialized bool
}

func (p *PointerComponent) Init(ctx context.Context) error {
	p.Initialized = true
	return nil
}

// Struct that returns error from Init
type FailingComponent struct {
	ShouldFail bool
}

func (f *FailingComponent) Init(ctx context.Context) error {
	if f.ShouldFail {
		return errors.New("initialization failed as expected")
	}
	return nil
}

// Complex struct with multiple levels
type ComplexApp struct {
	Config     ConfigManager
	Services   []Service
	Cache      *CacheLayer
	Plugins    map[string]Plugin
	Initialized bool
}

func (c *ComplexApp) Init(ctx context.Context) error {
	c.Initialized = true
	return nil
}

type ConfigManager struct {
	Settings   map[string]string
	Initialized bool
}

func (c *ConfigManager) Init(ctx context.Context) error {
	c.Initialized = true
	if c.Settings == nil {
		c.Settings = make(map[string]string)
	}
	c.Settings["initialized"] = "true"
	return nil
}

type Service struct {
	Name       string
	Database   *Database
	Initialized bool
}

func (s *Service) Init(ctx context.Context) error {
	s.Initialized = true
	return nil
}

type Database struct {
	Connected  bool
	ShouldFail bool
}

func (d *Database) Init(ctx context.Context) error {
	if d.ShouldFail {
		return errors.New("database connection failed")
	}
	d.Connected = true
	return nil
}

type CacheLayer struct {
	Entries    map[string]interface{}
	Initialized bool
}

func (c *CacheLayer) Init(ctx context.Context) error {
	c.Initialized = true
	if c.Entries == nil {
		c.Entries = make(map[string]interface{})
	}
	return nil
}

type Plugin struct {
	Active     bool
	Initialized bool
}

func (p *Plugin) Init(ctx context.Context) error {
	p.Initialized = true
	p.Active = true
	return nil
}

// Struct without Init method
type NoInitStruct struct {
	Value string
}

// Struct with embedded field
type EmbeddedStruct struct {
	SimpleComponent // embedded
	Extra          string
	Initialized    bool
}

func (e *EmbeddedStruct) Init(ctx context.Context) error {
	e.Initialized = true
	e.Extra = "embedded"
	return nil
}

// Test basic initialization
func TestSimpleInit(t *testing.T) {
	component := &SimpleComponent{}
	ctx := context.Background()
	
	err := AutoInit(ctx, component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if component.Name != "initialized" {
		t.Errorf("expected Name to be 'initialized', got '%s'", component.Name)
	}
}

// Test nested struct initialization
func TestNestedInit(t *testing.T) {
	component := &NestedComponent{}
	
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if !component.SubComponent.Initialized {
		t.Error("SubComponent was not initialized")
	}
	
	if component.ID != 42 {
		t.Errorf("expected ID to be 42, got %d", component.ID)
	}
}

// Test pointer field initialization
func TestPointerFieldInit(t *testing.T) {
	component := &PointerComponent{
		Sub:    &SimpleComponent{},
		Nested: &NestedComponent{},
	}
	
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if !component.Sub.Initialized {
		t.Error("Sub was not initialized")
	}
	
	if !component.Nested.Initialized {
		t.Error("Nested was not initialized")
	}
	
	if !component.Nested.SubComponent.Initialized {
		t.Error("Nested.SubComponent was not initialized")
	}
}

// Test nil pointer handling
func TestNilPointerHandling(t *testing.T) {
	component := &PointerComponent{
		Sub: nil, // nil pointer should be skipped
		Nested: &NestedComponent{},
	}
	
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if component.Sub != nil {
		t.Error("nil pointer should remain nil")
	}
	
	if !component.Nested.Initialized {
		t.Error("Nested was not initialized")
	}
}

// Test error propagation
func TestErrorPropagation(t *testing.T) {
	component := &struct {
		Working SimpleComponent
		Failing FailingComponent
	}{
		Failing: FailingComponent{ShouldFail: true},
	}
	
	err := AutoInit(context.Background(), component)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	
	if !strings.Contains(err.Error(), "initialization failed as expected") {
		t.Errorf("unexpected error message: %v", err)
	}
	
	if !strings.Contains(err.Error(), "Failing") {
		t.Errorf("error should contain field name 'Failing': %v", err)
	}
	
	// Working component should have been initialized before failure
	if !component.Working.Initialized {
		t.Error("Working component should have been initialized before failure")
	}
}

// Test complex struct with slices and maps
func TestComplexStructInit(t *testing.T) {
	app := &ComplexApp{
		Services: []Service{
			{Name: "service1", Database: &Database{}},
			{Name: "service2", Database: &Database{}},
		},
		Cache: &CacheLayer{},
		Plugins: map[string]Plugin{
			"plugin1": {},
			"plugin2": {},
		},
	}
	
	err := AutoInit(context.Background(), app)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check app initialization
	if !app.Initialized {
		t.Error("app was not initialized")
	}
	
	// Check config initialization
	if !app.Config.Initialized {
		t.Error("Config was not initialized")
	}
	
	if app.Config.Settings["initialized"] != "true" {
		t.Error("Config settings not properly initialized")
	}
	
	// Check services initialization
	for i, service := range app.Services {
		if !service.Initialized {
			t.Errorf("Service[%d] was not initialized", i)
		}
		if !service.Database.Connected {
			t.Errorf("Service[%d].Database was not connected", i)
		}
	}
	
	// Check cache initialization
	if !app.Cache.Initialized {
		t.Error("Cache was not initialized")
	}
	
	// Check plugins initialization
	for name, plugin := range app.Plugins {
		if !plugin.Initialized {
			t.Errorf("Plugin[%s] was not initialized", name)
		}
		if !plugin.Active {
			t.Errorf("Plugin[%s] was not activated", name)
		}
	}
}

// Test error in nested slice element
func TestErrorInSliceElement(t *testing.T) {
	app := &ComplexApp{
		Services: []Service{
			{Name: "service1", Database: &Database{}},
			{Name: "service2", Database: &Database{ShouldFail: true}}, // This will fail
			{Name: "service3", Database: &Database{}},
		},
	}
	
	err := AutoInit(context.Background(), app)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	
	initErr, ok := err.(*InitError)
	if !ok {
		t.Fatalf("expected *InitError, got %T", err)
	}
	
	// Check error path contains Services[1].Database
	path := strings.Join(initErr.GetPath(), ".")
	if !strings.Contains(path, "Services.[1].Database") {
		t.Errorf("error path should contain 'Services.[1].Database', got: %s", path)
	}
	
	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("error should contain original message: %v", err)
	}
}

// Test struct without Init method
func TestStructWithoutInit(t *testing.T) {
	component := &struct {
		NoInit NoInitStruct
		WithInit SimpleComponent
	}{}
	
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// NoInit should be skipped without error
	if !component.WithInit.Initialized {
		t.Error("WithInit component was not initialized")
	}
}

// Test embedded struct
func TestEmbeddedStruct(t *testing.T) {
	component := &EmbeddedStruct{}
	
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if !component.SimpleComponent.Initialized {
		t.Error("embedded SimpleComponent was not initialized")
	}
	
	if component.Extra != "embedded" {
		t.Errorf("expected Extra to be 'embedded', got '%s'", component.Extra)
	}
}

// Test value struct (not pointer)
func TestValueStruct(t *testing.T) {
	component := SimpleComponent{}
	
	err := AutoInit(context.Background(), &component) // Pass pointer to allow modification
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
}

// Test nil target
func TestNilTarget(t *testing.T) {
	err := AutoInit(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
	
	if !strings.Contains(err.Error(), "nil target") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// Test nil pointer target
func TestNilPointerTarget(t *testing.T) {
	var component *SimpleComponent
	
	err := AutoInit(context.Background(), component)
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
	
	if !strings.Contains(err.Error(), "nil pointer") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// Test non-struct target
func TestNonStructTarget(t *testing.T) {
	var intValue int = 42
	
	err := AutoInit(context.Background(), &intValue)
	if err == nil {
		t.Fatal("expected error for non-struct target")
	}
	
	if !strings.Contains(err.Error(), "must be a struct") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// Test initialization order
func TestInitializationOrder(t *testing.T) {
	type OrderedComponent struct {
		Name string
	}
	
	component := &struct {
		First  OrderedComponent
		Second OrderedComponent
		Third  OrderedComponent
	}{
		First:  OrderedComponent{Name: "first"},
		Second: OrderedComponent{Name: "second"},
		Third:  OrderedComponent{Name: "third"},
	}
	
	// We can't easily test order with current implementation,
	// but we ensure all fields are processed
	err := AutoInit(context.Background(), component)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// All fields should maintain their values
	if component.First.Name != "first" {
		t.Error("First field value changed unexpectedly")
	}
	if component.Second.Name != "second" {
		t.Error("Second field value changed unexpectedly")
	}
	if component.Third.Name != "third" {
		t.Error("Third field value changed unexpectedly")
	}
}