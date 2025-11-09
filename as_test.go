package autoinit_test

import (
	"context"
	"testing"

	"github.com/telnet2/autoinit"
)

const (
	primaryDBName = "primary"
	primaryTag    = "primary"
)

// Test components for As pattern testing
type TestDatabase struct {
	Name      string
	Connected bool
}

type TestCache struct {
	Name  string
	Ready bool
}

type TestLogger interface {
	Log(message string)
}

type TestStructLogger struct {
	Name string
	Logs []string
}

func (l *TestStructLogger) Log(message string) {
	l.Logs = append(l.Logs, message)
}

// TestAsBasicTypeMatching tests basic type matching without filters
func TestAsBasicTypeMatching(t *testing.T) {
	type App struct {
		DB    *TestDatabase
		Cache *TestCache
	}

	app := &App{
		DB:    &TestDatabase{Name: "mainDB", Connected: true},
		Cache: &TestCache{Name: "mainCache", Ready: true},
	}

	ctx := context.Background()

	// Test finding a Database
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db) {
		t.Error("Failed to find TestDatabase")
	}
	if db.Name != "mainDB" {
		t.Errorf("Expected DB name 'mainDB', got '%s'", db.Name)
	}

	// Test finding a Cache
	var cache *TestCache
	if !autoinit.As(ctx, nil, app, &cache) {
		t.Error("Failed to find TestCache")
	}
	if cache.Name != "mainCache" {
		t.Errorf("Expected Cache name 'mainCache', got '%s'", cache.Name)
	}

	// Test not finding a non-existent type
	type NonExistent struct{}
	var ne *NonExistent
	if autoinit.As(ctx, nil, app, &ne) {
		t.Error("Should not find NonExistent type")
	}
}

// TestAsWithFieldNameFilter tests filtering by field name
func TestAsWithFieldNameFilter(t *testing.T) {
	type App struct {
		PrimaryDB   *TestDatabase
		SecondaryDB *TestDatabase
		MainCache   *TestCache
	}

	app := &App{
		PrimaryDB:   &TestDatabase{Name: primaryDBName, Connected: true},
		SecondaryDB: &TestDatabase{Name: "secondary", Connected: false},
		MainCache:   &TestCache{Name: "main", Ready: true},
	}

	ctx := context.Background()

	// Find PrimaryDB specifically
	var primaryDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &primaryDB, autoinit.WithFieldName("PrimaryDB")) {
		t.Error("Failed to find PrimaryDB")
	}
	if primaryDB.Name != primaryDBName {
		t.Errorf("Expected DB name 'primary', got '%s'", primaryDB.Name)
	}

	// Find SecondaryDB specifically
	var secondaryDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &secondaryDB, autoinit.WithFieldName("SecondaryDB")) {
		t.Error("Failed to find SecondaryDB")
	}
	if secondaryDB.Name != "secondary" {
		t.Errorf("Expected DB name 'secondary', got '%s'", secondaryDB.Name)
	}

	// Should not find with wrong field name
	var notFound *TestDatabase
	if autoinit.As(ctx, nil, app, &notFound, autoinit.WithFieldName("NonExistentDB")) {
		t.Error("Should not find with non-existent field name")
	}
}

// TestAsWithJSONTagFilter tests filtering by JSON tag
func TestAsWithJSONTagFilter(t *testing.T) {
	type App struct {
		MainDB    *TestDatabase `json:"primary"`
		BackupDB  *TestDatabase `json:"backup"`
		TempCache *TestCache    `json:"temp"`
	}

	app := &App{
		MainDB:    &TestDatabase{Name: "main", Connected: true},
		BackupDB:  &TestDatabase{Name: "backup", Connected: false},
		TempCache: &TestCache{Name: "temp", Ready: true},
	}

	ctx := context.Background()

	// Find database with json:"primary" tag
	var primaryDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &primaryDB, autoinit.WithJSONTag(primaryTag)) {
		t.Error("Failed to find database with json:primary tag")
	}
	if primaryDB.Name != "main" {
		t.Errorf("Expected DB name 'main', got '%s'", primaryDB.Name)
	}

	// Find database with json:"backup" tag
	var backupDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &backupDB, autoinit.WithJSONTag("backup")) {
		t.Error("Failed to find database with json:backup tag")
	}
	if backupDB.Name != "backup" {
		t.Errorf("Expected DB name 'backup', got '%s'", backupDB.Name)
	}
}

// TestAsWithCustomTagFilter tests filtering by custom tags
func TestAsWithCustomTagFilter(t *testing.T) {
	type App struct {
		PrimaryDB   *TestDatabase `component:"primary"`
		SecondaryDB *TestDatabase `component:"secondary"`
		Cache       *TestCache    `component:"cache"`
	}

	app := &App{
		PrimaryDB:   &TestDatabase{Name: primaryDBName, Connected: true},
		SecondaryDB: &TestDatabase{Name: "secondary", Connected: false},
		Cache:       &TestCache{Name: "cache", Ready: true},
	}

	ctx := context.Background()

	// Find component with custom tag
	var primaryDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &primaryDB, autoinit.WithTag("component", primaryTag)) {
		t.Error("Failed to find database with component:primary tag")
	}
	if primaryDB.Name != primaryDBName {
		t.Errorf("Expected DB name 'primary', got '%s'", primaryDB.Name)
	}

	// Find secondary component
	var secondaryDB *TestDatabase
	if !autoinit.As(ctx, nil, app, &secondaryDB, autoinit.WithTag("component", "secondary")) {
		t.Error("Failed to find database with component:secondary tag")
	}
	if secondaryDB.Name != "secondary" {
		t.Errorf("Expected DB name 'secondary', got '%s'", secondaryDB.Name)
	}
}

// TestAsConjunctiveFilters tests multiple filters applied together (AND logic)
func TestAsConjunctiveFilters(t *testing.T) {
	type App struct {
		PrimaryDB   *TestDatabase `json:"main" component:"primary"`
		SecondaryDB *TestDatabase `json:"backup" component:"secondary"`
		TertiaryDB  *TestDatabase `json:"temp" component:"tertiary"`
	}

	app := &App{
		PrimaryDB:   &TestDatabase{Name: primaryDBName, Connected: true},
		SecondaryDB: &TestDatabase{Name: "secondary", Connected: false},
		TertiaryDB:  &TestDatabase{Name: "tertiary", Connected: false},
	}

	ctx := context.Background()

	// Find with multiple filters - all must match
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db,
		autoinit.WithFieldName("PrimaryDB"),
		autoinit.WithJSONTag("main"),
		autoinit.WithTag("component", primaryTag)) {
		t.Error("Failed to find database with all matching filters")
	}
	if db.Name != primaryDBName {
		t.Errorf("Expected DB name 'primary', got '%s'", db.Name)
	}

	// Should not find when one filter doesn't match
	var notFound *TestDatabase
	if autoinit.As(ctx, nil, app, &notFound,
		autoinit.WithFieldName("PrimaryDB"),         // matches
		autoinit.WithJSONTag("wrong"),               // doesn't match
		autoinit.WithTag("component", primaryTag)) { // matches
		t.Error("Should not find when one filter doesn't match")
	}

	// Another combination that should not match
	if autoinit.As(ctx, nil, app, &notFound,
		autoinit.WithFieldName("SecondaryDB"), // matches SecondaryDB
		autoinit.WithJSONTag("main")) {        // but this doesn't match SecondaryDB
		t.Error("Should not find when filters don't all match the same field")
	}
}

// TestAsInterfaceMatching tests matching interface types
func TestAsInterfaceMatching(t *testing.T) {
	type App struct {
		Logger *TestStructLogger
		DB     *TestDatabase
	}

	app := &App{
		Logger: &TestStructLogger{Name: "appLogger"},
		DB:     &TestDatabase{Name: "mainDB"},
	}

	ctx := context.Background()

	// Find by interface type
	var logger TestLogger
	if !autoinit.As(ctx, nil, app, &logger) {
		t.Error("Failed to find TestLogger interface")
	}
	if _, ok := logger.(*TestStructLogger); !ok {
		t.Error("Logger should be *TestStructLogger")
	}

	// Test that it actually works
	logger.Log("test message")
	if len(app.Logger.Logs) != 1 || app.Logger.Logs[0] != "test message" {
		t.Error("Logger interface method call failed")
	}
}

// TestAsValueTypes tests finding value types (non-pointer fields)
func TestAsValueTypes(t *testing.T) {
	type App struct {
		DB    TestDatabase // Value type, not pointer
		Cache *TestCache   // Pointer type
	}

	app := &App{
		DB:    TestDatabase{Name: "valueDB", Connected: true},
		Cache: &TestCache{Name: "ptrCache", Ready: true},
	}

	ctx := context.Background()

	// Should find value type and return pointer to it
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db) {
		t.Error("Failed to find value type TestDatabase")
	}
	if db.Name != "valueDB" {
		t.Errorf("Expected DB name 'valueDB', got '%s'", db.Name)
	}

	// Modifying through the pointer should modify the original
	db.Connected = false
	if app.DB.Connected != false {
		t.Error("Modification through pointer should affect original value")
	}
}

// TestAsInSlices tests finding components in slices
func TestAsInSlices(t *testing.T) {
	type App struct {
		Databases []*TestDatabase
		Caches    []TestCache // Value type slice
	}

	app := &App{
		Databases: []*TestDatabase{
			{Name: "db1", Connected: true},
			{Name: "db2", Connected: false},
		},
		Caches: []TestCache{
			{Name: "cache1", Ready: true},
			{Name: "cache2", Ready: false},
		},
	}

	ctx := context.Background()

	// Find in pointer slice
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db) {
		t.Error("Failed to find TestDatabase in slice")
	}
	if db.Name != "db1" {
		t.Errorf("Expected first DB 'db1', got '%s'", db.Name)
	}

	// Find in value slice
	var cache *TestCache
	if !autoinit.As(ctx, nil, app, &cache) {
		t.Error("Failed to find TestCache in value slice")
	}
	if cache.Name != "cache1" {
		t.Errorf("Expected first cache 'cache1', got '%s'", cache.Name)
	}
}

// TestAsInMaps tests finding components in maps
func TestAsInMaps(t *testing.T) {
	type App struct {
		Services map[string]*TestDatabase
	}

	app := &App{
		Services: map[string]*TestDatabase{
			"primary":   {Name: "primaryDB", Connected: true},
			"secondary": {Name: "secondaryDB", Connected: false},
		},
	}

	ctx := context.Background()

	// Find in map
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db) {
		t.Error("Failed to find TestDatabase in map")
	}
	// Note: map iteration order is not guaranteed, so we just check that we found one
	if db.Name != "primaryDB" && db.Name != "secondaryDB" {
		t.Errorf("Expected to find one of the databases, got '%s'", db.Name)
	}
}

// TestAsEmbeddedStructs tests finding components in embedded structs
func TestAsEmbeddedStructs(t *testing.T) {
	type BaseServices struct {
		DB    *TestDatabase
		Cache *TestCache
	}

	type App struct {
		BaseServices // Embedded
		Logger       *TestStructLogger
	}

	app := &App{
		BaseServices: BaseServices{
			DB:    &TestDatabase{Name: "embeddedDB", Connected: true},
			Cache: &TestCache{Name: "embeddedCache", Ready: true},
		},
		Logger: &TestStructLogger{Name: "logger"},
	}

	ctx := context.Background()

	// Should find embedded fields
	var db *TestDatabase
	if !autoinit.As(ctx, nil, app, &db) {
		t.Error("Failed to find TestDatabase in embedded struct")
	}
	if db.Name != "embeddedDB" {
		t.Errorf("Expected embedded DB, got '%s'", db.Name)
	}

	var cache *TestCache
	if !autoinit.As(ctx, nil, app, &cache) {
		t.Error("Failed to find TestCache in embedded struct")
	}
	if cache.Name != "embeddedCache" {
		t.Errorf("Expected embedded cache, got '%s'", cache.Name)
	}
}

// TestMustAs tests the panic behavior of MustAs
func TestMustAs(t *testing.T) {
	type App struct {
		DB *TestDatabase
	}

	app := &App{
		DB: &TestDatabase{Name: "mainDB"},
	}

	ctx := context.Background()

	// Should not panic when found
	var db *TestDatabase
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Error("MustAs should not panic when dependency is found")
			}
		}()
		autoinit.MustAs(ctx, nil, app, &db)
	}()

	// Should panic when not found
	var cache *TestCache
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustAs should panic when dependency is not found")
			}
		}()
		autoinit.MustAs(ctx, nil, app, &cache)
	}()
}

// TestAsType tests the convenience AsType function
func TestAsType(t *testing.T) {
	type App struct {
		DB    *TestDatabase
		Cache *TestCache
	}

	app := &App{
		DB:    &TestDatabase{Name: "mainDB", Connected: true},
		Cache: &TestCache{Name: "mainCache", Ready: true},
	}

	ctx := context.Background()

	// Test finding with AsType
	db, ok := autoinit.AsType[*TestDatabase](ctx, nil, app)
	if !ok {
		t.Error("AsType failed to find TestDatabase")
	}
	if db.Name != "mainDB" {
		t.Errorf("Expected DB name 'mainDB', got '%s'", db.Name)
	}

	// Test not finding with AsType
	type NonExistent struct{}
	_, ok = autoinit.AsType[*NonExistent](ctx, nil, app)
	if ok {
		t.Error("AsType should return false for non-existent type")
	}
}

// Test types for TestAsSearchInAncestors
type AncestorConfig struct {
	MaxConnections int
	Timeout        int
}

type AncestorGrandChild struct {
	Name          string
	foundConfig   *AncestorConfig
	configFound   bool
	initCalled    bool
	parentWasNil  bool
}

func (gc *AncestorGrandChild) Init(ctx context.Context, parent interface{}) error {
	gc.initCalled = true
	if parent == nil {
		gc.parentWasNil = true
	}
	// Try to find Config - it's in Root (grandparent), not in Child (immediate parent)
	gc.configFound = autoinit.As(ctx, gc, parent, &gc.foundConfig)
	return nil
}

type AncestorChild struct {
	GrandChild *AncestorGrandChild
	// No Config here - GrandChild will need to search up to Root
}

type AncestorRoot struct {
	Config *AncestorConfig // Config is at the root level
	Child  *AncestorChild
}

// TestAsSearchInAncestors tests that As can find components in ancestor objects,
// not just in the immediate parent (siblings). This test uses AutoInit to properly
// set up the parent chain.
func TestAsSearchInAncestors(t *testing.T) {
	// Setup the structure
	root := &AncestorRoot{
		Config: &AncestorConfig{MaxConnections: 100, Timeout: 30},
		Child: &AncestorChild{
			GrandChild: &AncestorGrandChild{Name: "deep"},
		},
	}

	// Run AutoInit which will properly set up the parent chain
	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, root); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Verify that GrandChild's Init was called
	if !root.Child.GrandChild.initCalled {
		t.Fatal("GrandChild.Init was not called")
	}

	// Verify that the config was found via ancestor search
	if !root.Child.GrandChild.configFound {
		t.Error("Failed to find Config in ancestor (Root)")
	}

	if root.Child.GrandChild.foundConfig == nil {
		t.Fatal("foundConfig should not be nil")
	}

	if root.Child.GrandChild.foundConfig.MaxConnections != 100 {
		t.Errorf("Expected MaxConnections=100, got %d", root.Child.GrandChild.foundConfig.MaxConnections)
	}

	if root.Child.GrandChild.foundConfig.Timeout != 30 {
		t.Errorf("Expected Timeout=30, got %d", root.Child.GrandChild.foundConfig.Timeout)
	}

	// Verify that the found config is the exact same instance
	if root.Child.GrandChild.foundConfig != root.Config {
		t.Error("Found config should be the same instance as root.Config")
	}
}

// Test types for TestAsSearchAcrossLevels
type MultiLevelDatabase struct {
	Name string
}

type MultiLevel3 struct {
	Name    string
	foundDB *MultiLevelDatabase
	dbFound bool
}

func (l3 *MultiLevel3) Init(ctx context.Context, parent interface{}) error {
	// Search for Database - it's in Root, not in Level2 or Level1
	l3.dbFound = autoinit.As(ctx, l3, parent, &l3.foundDB)
	return nil
}

type MultiLevel2 struct {
	Level3 *MultiLevel3
	// No Database here
}

type MultiLevel1 struct {
	Level2 *MultiLevel2
	// No Database here either
}

type MultiLevelRoot struct {
	Database *MultiLevelDatabase // Database is only at Root
	Level1   *MultiLevel1
}

// TestAsSearchAcrossLevels tests searching across multiple ancestor levels
// Database is at Root level, and Level3 needs to find it by searching through
// multiple ancestor levels (Level2 -> Level1 -> Root)
func TestAsSearchAcrossLevels(t *testing.T) {
	root := &MultiLevelRoot{
		Database: &MultiLevelDatabase{Name: "MainDB"},
		Level1: &MultiLevel1{
			Level2: &MultiLevel2{
				Level3: &MultiLevel3{Name: "deep"},
			},
		},
	}

	// Run AutoInit to properly set up the parent chain
	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, root); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Verify that Level3 found the Database from Root
	if !root.Level1.Level2.Level3.dbFound {
		t.Error("Failed to find Database across multiple ancestor levels")
	}

	if root.Level1.Level2.Level3.foundDB == nil {
		t.Fatal("foundDB should not be nil")
	}

	if root.Level1.Level2.Level3.foundDB.Name != "MainDB" {
		t.Errorf("Expected Database name 'MainDB', got '%s'", root.Level1.Level2.Level3.foundDB.Name)
	}

	if root.Level1.Level2.Level3.foundDB != root.Database {
		t.Error("Found database should be the same instance as root.Database")
	}
}
