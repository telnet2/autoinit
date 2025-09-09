package autoinit_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/telnet2/autoinit"
)

// Test components for finder
type FinderLogger struct {
	Name       string
	Configured bool
}

func (l *FinderLogger) Init(ctx context.Context) error {
	l.Configured = true
	return nil
}

type FinderCache struct {
	Name   string
	Ready  bool
	logger *FinderLogger
}

func (c *FinderCache) Init(ctx context.Context, parent interface{}) error {
	// Find logger sibling by type
	finder := autoinit.NewComponentFinder(ctx, c, parent)
	if logger := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*FinderLogger)(nil)),
	}); logger != nil {
		c.logger = logger.(*FinderLogger)
	}
	c.Ready = true
	return nil
}

type FinderDatabase struct {
	Name      string
	Connected bool
	logger    *FinderLogger
	cache     *FinderCache
}

func (d *FinderDatabase) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, d, parent)

	// Find logger by field name
	if logger := finder.Find(autoinit.SearchOption{
		ByFieldName: "Logger",
	}); logger != nil {
		d.logger = logger.(*FinderLogger)
	}

	// Find cache by JSON tag
	if cache := finder.Find(autoinit.SearchOption{
		ByJSONTag: "cache",
	}); cache != nil {
		d.cache = cache.(*FinderCache)
	}

	d.Connected = true
	return nil
}

// Service that needs to find components at different levels
type FinderService struct {
	Name   string
	Active bool
	logger *FinderLogger
	cache  *FinderCache
	db     *FinderDatabase
}

func (s *FinderService) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, s, parent)

	// Should find these at various levels in the hierarchy
	if logger := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*FinderLogger)(nil)),
	}); logger != nil {
		s.logger = logger.(*FinderLogger)
	}

	if cache := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*FinderCache)(nil)),
	}); cache != nil {
		s.cache = cache.(*FinderCache)
	}

	if db := finder.Find(autoinit.SearchOption{
		ByType: reflect.TypeOf((*FinderDatabase)(nil)),
	}); db != nil {
		s.db = db.(*FinderDatabase)
	}

	s.Active = true
	return nil
}

func TestFinderSiblingSearch(t *testing.T) {
	// Simple app with siblings
	type App struct {
		Logger   *FinderLogger   `json:"logger"`
		Cache    *FinderCache    `json:"cache"`
		Database *FinderDatabase `json:"db"`
	}

	app := &App{
		Logger:   &FinderLogger{Name: "AppLogger"},
		Cache:    &FinderCache{Name: "AppCache"},
		Database: &FinderDatabase{Name: "AppDB"},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Verify components found their siblings
	if app.Cache.logger == nil {
		t.Error("Cache should have found Logger sibling")
	} else if app.Cache.logger != app.Logger {
		t.Error("Cache should have found the correct Logger")
	}

	if app.Database.logger == nil {
		t.Error("Database should have found Logger sibling")
	}

	if app.Database.cache == nil {
		t.Error("Database should have found Cache sibling")
	} else if app.Database.cache != app.Cache {
		t.Error("Database should have found the correct Cache")
	}
}

func TestFinderHierarchicalSearch(t *testing.T) {
	// Nested structure to test hierarchical search
	type Module struct {
		Service *FinderService
	}

	type Subsystem struct {
		LocalCache *FinderCache // Might be nil initially
		Module     *Module
	}

	type System struct {
		GlobalLogger *FinderLogger
		GlobalCache  *FinderCache
		Database     *FinderDatabase
		Subsystem    *Subsystem
	}

	system := &System{
		GlobalLogger: &FinderLogger{Name: "GlobalLogger"},
		GlobalCache:  &FinderCache{Name: "GlobalCache"},
		Database:     &FinderDatabase{Name: "SystemDB"},
		Subsystem: &Subsystem{
			// No local cache
			Module: &Module{
				Service: &FinderService{Name: "ModuleService"},
			},
		},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, system); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	service := system.Subsystem.Module.Service

	// Service should find components at the system level
	if service.logger == nil {
		t.Error("Service should have found Logger at system level")
	} else if service.logger != system.GlobalLogger {
		t.Error("Service should have found the GlobalLogger")
	}

	if service.cache == nil {
		t.Error("Service should have found Cache at system level")
	} else if service.cache != system.GlobalCache {
		t.Error("Service should have found the GlobalCache")
	}

	if service.db == nil {
		t.Error("Service should have found Database at system level")
	} else if service.db != system.Database {
		t.Error("Service should have found the Database")
	}
}

func TestFinderWithLocalOverride(t *testing.T) {
	// Test that local components override global ones
	type Module struct {
		Service *FinderService
	}

	type Subsystem struct {
		LocalCache *FinderCache // Local override
		Module     *Module
	}

	type System struct {
		GlobalCache *FinderCache
		Subsystem   *Subsystem
	}

	system := &System{
		GlobalCache: &FinderCache{Name: "GlobalCache"},
		Subsystem: &Subsystem{
			LocalCache: &FinderCache{Name: "LocalCache"},
			Module: &Module{
				Service: &FinderService{Name: "ModuleService"},
			},
		},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, system); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	service := system.Subsystem.Module.Service

	// Service should find the local cache, not the global one
	if service.cache == nil {
		t.Error("Service should have found Cache")
	} else if service.cache != system.Subsystem.LocalCache {
		t.Error("Service should have found the LocalCache, not GlobalCache")
	}
}

// Custom service that looks for specific field names
type CustomService struct {
	Name    string
	primary *FinderLogger
}

func (c *CustomService) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, c, parent)
	if logger := finder.Find(autoinit.SearchOption{
		ByFieldName: "PrimaryLogger",
	}); logger != nil {
		c.primary = logger.(*FinderLogger)
	}
	return nil
}

func TestFinderByFieldName(t *testing.T) {
	type App struct {
		PrimaryLogger   *FinderLogger
		SecondaryLogger *FinderLogger
		Service         *FinderService
	}

	app := &App{
		PrimaryLogger:   &FinderLogger{Name: "Primary"},
		SecondaryLogger: &FinderLogger{Name: "Secondary"},
		Service:         &FinderService{Name: "Service"},
	}

	// Add custom service to test
	type TestApp struct {
		App
		Custom *CustomService
	}

	testApp := &TestApp{
		App:    *app,
		Custom: &CustomService{Name: "Custom"},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, testApp); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Custom service should have found the primary logger
	if testApp.Custom.primary == nil {
		t.Error("CustomService should have found PrimaryLogger")
	} else if testApp.Custom.primary != testApp.PrimaryLogger {
		t.Error("CustomService should have found the correct PrimaryLogger")
	}
}

// Service that looks for specific component tags
type TaggedService struct {
	main     *FinderCache
	fallback *FinderCache
}

func (ts *TaggedService) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, ts, parent)

	// Find main cache
	if cache := finder.Find(autoinit.SearchOption{
		ByCustomTag: "main",
		TagKey:      "component",
	}); cache != nil {
		ts.main = cache.(*FinderCache)
	}

	// Find fallback cache
	if cache := finder.Find(autoinit.SearchOption{
		ByCustomTag: "fallback",
		TagKey:      "component",
	}); cache != nil {
		ts.fallback = cache.(*FinderCache)
	}

	return nil
}

func TestFinderWithCustomTags(t *testing.T) {
	type App struct {
		MainCache     *FinderCache `component:"main"`
		FallbackCache *FinderCache `component:"fallback"`
		TempCache     *FinderCache `component:"temp"`
	}

	type TestApp struct {
		App
		Service *TaggedService
	}

	app := &TestApp{
		App: App{
			MainCache:     &FinderCache{Name: "Main"},
			FallbackCache: &FinderCache{Name: "Fallback"},
			TempCache:     &FinderCache{Name: "Temp"},
		},
		Service: &TaggedService{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Service should have found tagged caches
	if app.Service.main != app.MainCache {
		t.Error("Service should have found main cache")
	}

	if app.Service.fallback != app.FallbackCache {
		t.Error("Service should have found fallback cache")
	}
}

// Service using helper functions
type HelperService struct {
	logger *FinderLogger
	cache  *FinderCache
}

func (h *HelperService) Init(ctx context.Context, parent interface{}) error {
	// Use generic helper
	h.logger = autoinit.FindByType[*FinderLogger](ctx, h, parent)

	// Use name helper
	if cache := autoinit.FindByName(ctx, h, parent, "Cache"); cache != nil {
		h.cache = cache.(*FinderCache)
	}

	return nil
}

// Test helper functions
func TestFinderHelperFunctions(t *testing.T) {
	type App struct {
		Logger  *FinderLogger
		Cache   *FinderCache
		Service *FinderService
	}

	type TestApp struct {
		App
		Helper *HelperService
	}

	app := &TestApp{
		App: App{
			Logger:  &FinderLogger{Name: "Logger"},
			Cache:   &FinderCache{Name: "Cache"},
			Service: &FinderService{Name: "Service"},
		},
		Helper: &HelperService{},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Helper service should have found components
	if app.Helper.logger != app.Logger {
		t.Error("Helper should have found logger using FindByType")
	}

	if app.Helper.cache != app.Cache {
		t.Error("Helper should have found cache using FindByName")
	}
}

// Test searching in collections (slices and maps)
func TestFinderWithCollections(t *testing.T) {
	type App struct {
		Loggers  []*FinderLogger
		CacheMap map[string]*FinderCache
		Service  *FinderService
	}

	app := &App{
		Loggers: []*FinderLogger{
			{Name: "Logger1"},
			{Name: "Logger2"},
		},
		CacheMap: map[string]*FinderCache{
			"primary": {Name: "PrimaryCache"},
			"backup":  {Name: "BackupCache"},
		},
		Service: &FinderService{Name: "Service"},
	}

	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Service should find components in collections
	if app.Service.logger == nil {
		t.Error("Service should have found a logger from the slice")
	}

	if app.Service.cache == nil {
		t.Error("Service should have found a cache from the map")
	}
}
