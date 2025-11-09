package autoinit_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/telnet2/autoinit"
)

// Test types for TestFinderSearchInAncestors
type FinderGlobalConfig struct {
	DatabaseURL string
	APIKey      string
}

type FinderDeepService struct {
	Name        string
	foundConfig *FinderGlobalConfig
}

func (ds *FinderDeepService) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, ds, parent)
	result := finder.Find(&autoinit.SearchOption{
		ByType: reflect.TypeOf(&FinderGlobalConfig{}),
	})
	if cfg, ok := result.(*FinderGlobalConfig); ok {
		ds.foundConfig = cfg
	}
	return nil
}

type FinderMiddleLayer struct {
	Service *FinderDeepService
}

type FinderRoot struct {
	Config *FinderGlobalConfig
	Middle *FinderMiddleLayer
}

// TestFinderSearchInAncestors tests that Find can locate components in ancestor objects
func TestFinderSearchInAncestors(t *testing.T) {
	root := &FinderRoot{
		Config: &FinderGlobalConfig{
			DatabaseURL: "postgres://localhost",
			APIKey:      "secret-key",
		},
		Middle: &FinderMiddleLayer{
			Service: &FinderDeepService{Name: "deep-service"},
		},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, root); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	if root.Middle.Service.foundConfig == nil {
		t.Fatal("DeepService should have found GlobalConfig in ancestor (Root)")
	}

	if root.Middle.Service.foundConfig.DatabaseURL != "postgres://localhost" {
		t.Errorf("Expected DatabaseURL 'postgres://localhost', got '%s'",
			root.Middle.Service.foundConfig.DatabaseURL)
	}

	if root.Middle.Service.foundConfig != root.Config {
		t.Error("Found config should be the same instance as root.Config")
	}
}

// Test types for TestFinderSearchAcrossMultipleLevels
type MultiLevelSharedResource struct {
	Name string
}

type MultiLevel4Finder struct {
	Name     string
	resource *MultiLevelSharedResource
}

func (l4 *MultiLevel4Finder) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, l4, parent)
	result := finder.Find(&autoinit.SearchOption{
		ByType: reflect.TypeOf(&MultiLevelSharedResource{}),
	})
	if res, ok := result.(*MultiLevelSharedResource); ok {
		l4.resource = res
	}
	return nil
}

type MultiLevel3Finder struct {
	Level4 *MultiLevel4Finder
}

type MultiLevel2Finder struct {
	Level3 *MultiLevel3Finder
}

type MultiLevel1Finder struct {
	Level2 *MultiLevel2Finder
}

type MultiLevelRootFinder struct {
	Resource *MultiLevelSharedResource
	Level1   *MultiLevel1Finder
}

// TestFinderSearchAcrossMultipleLevels tests searching across many ancestor levels
func TestFinderSearchAcrossMultipleLevels(t *testing.T) {
	root := &MultiLevelRootFinder{
		Resource: &MultiLevelSharedResource{Name: "shared-resource"},
		Level1: &MultiLevel1Finder{
			Level2: &MultiLevel2Finder{
				Level3: &MultiLevel3Finder{
					Level4: &MultiLevel4Finder{Name: "deep-component"},
				},
			},
		},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, root); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	if root.Level1.Level2.Level3.Level4.resource == nil {
		t.Fatal("Level4 should have found SharedResource from Root")
	}

	if root.Level1.Level2.Level3.Level4.resource.Name != "shared-resource" {
		t.Errorf("Expected resource name 'shared-resource', got '%s'",
			root.Level1.Level2.Level3.Level4.resource.Name)
	}

	if root.Level1.Level2.Level3.Level4.resource != root.Resource {
		t.Error("Found resource should be the same instance as root.Resource")
	}
}

// Test types for TestFinderByFieldNameInAncestors
type AncestorDatabase struct {
	Name string
}

type AncestorService struct {
	primaryDB *AncestorDatabase
}

func (s *AncestorService) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, s, parent)
	result := finder.Find(&autoinit.SearchOption{
		ByFieldName: "PrimaryDB",
	})
	if db, ok := result.(*AncestorDatabase); ok {
		s.primaryDB = db
	}
	return nil
}

type AncestorLayer struct {
	Service *AncestorService
}

type AncestorApp struct {
	PrimaryDB *AncestorDatabase
	Layer     *AncestorLayer
}

// TestFinderByFieldNameInAncestors tests finding by field name across ancestors
func TestFinderByFieldNameInAncestors(t *testing.T) {
	app := &AncestorApp{
		PrimaryDB: &AncestorDatabase{Name: "primary-database"},
		Layer: &AncestorLayer{
			Service: &AncestorService{},
		},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	if app.Layer.Service.primaryDB == nil {
		t.Fatal("Service should have found PrimaryDB by field name")
	}

	if app.Layer.Service.primaryDB.Name != "primary-database" {
		t.Errorf("Expected database name 'primary-database', got '%s'",
			app.Layer.Service.primaryDB.Name)
	}

	if app.Layer.Service.primaryDB != app.PrimaryDB {
		t.Error("Found database should be the same instance as app.PrimaryDB")
	}
}

// Test types for TestFinderConsistencyWithAs
type ConsistencyConfig struct {
	Value string
}

type ConsistencyComponent struct {
	Name          string
	configViaAs   *ConsistencyConfig
	configViaFind *ConsistencyConfig
}

func (c *ConsistencyComponent) Init(ctx context.Context, parent interface{}) error {
	// Use As
	autoinit.As(ctx, c, parent, &c.configViaAs)

	// Use Find
	finder := autoinit.NewComponentFinder(ctx, c, parent)
	result := finder.Find(&autoinit.SearchOption{
		ByType: reflect.TypeOf(&ConsistencyConfig{}),
	})
	if cfg, ok := result.(*ConsistencyConfig); ok {
		c.configViaFind = cfg
	}

	return nil
}

type ConsistencyMiddle struct {
	Component *ConsistencyComponent
}

type ConsistencyRoot struct {
	Config *ConsistencyConfig
	Middle *ConsistencyMiddle
}

// TestFinderConsistencyWithAs verifies that Find and As behave consistently
func TestFinderConsistencyWithAs(t *testing.T) {
	root := &ConsistencyRoot{
		Config: &ConsistencyConfig{Value: "test-config"},
		Middle: &ConsistencyMiddle{
			Component: &ConsistencyComponent{Name: "test"},
		},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, root); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}

	// Both As and Find should find the same config
	if root.Middle.Component.configViaAs == nil {
		t.Error("As should have found the config")
	}

	if root.Middle.Component.configViaFind == nil {
		t.Error("Find should have found the config")
	}

	// They should find the exact same instance
	if root.Middle.Component.configViaAs != root.Middle.Component.configViaFind {
		t.Error("As and Find should find the same config instance")
	}

	if root.Middle.Component.configViaAs != root.Config {
		t.Error("Both should find root.Config")
	}
}
