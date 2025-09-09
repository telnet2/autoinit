package autoinit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/user/autoinit"
)

// Service component for example
type ExampleService struct {
	db    *ExampleDatabase
	cache *ExampleCache
}

func (s *ExampleService) Init(ctx context.Context, parent interface{}) error {
	// Find any Database
	if autoinit.As(ctx, s, parent, &s.db) {
		fmt.Printf("Found database: %s\n", s.db.Name)
	}

	// Find Cache - required dependency
	autoinit.MustAs(ctx, s, parent, &s.cache)
	fmt.Printf("Found cache: %s\n", s.cache.Name)

	return nil
}

type ExampleDatabase struct {
	Name      string
	Connected bool
}

type ExampleCache struct {
	Name  string
	Ready bool
}

// Example of using As pattern for dependency discovery
func ExampleAs() {
	// Application structure
	type App struct {
		MainDB  *ExampleDatabase `json:"primary"`
		Cache   *ExampleCache
		Service *ExampleService
	}

	// Initialize application
	app := &App{
		MainDB:  &ExampleDatabase{Name: "PostgreSQL", Connected: true},
		Cache:   &ExampleCache{Name: "Redis", Ready: true},
		Service: &ExampleService{},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, app); err != nil {
		panic(err)
	}

}

// Service2 for filter example
type ExampleService2 struct {
	primaryDB *ExampleDatabase
}

func (s *ExampleService2) Init(ctx context.Context, parent interface{}) error {
	// Find database that matches ALL criteria
	if autoinit.As(ctx, s, parent, &s.primaryDB,
		autoinit.WithFieldName("PrimaryDB"),
		autoinit.WithJSONTag("main")) {
		fmt.Printf("Found primary database: %s\n", s.primaryDB.Name)
	}
	return nil
}

// Example of using As pattern with conjunctive filters
func ExampleAs_withFilters() {
	type App struct {
		PrimaryDB   *ExampleDatabase `json:"main" component:"primary"`
		SecondaryDB *ExampleDatabase `json:"backup" component:"secondary"`
		Service     *ExampleService2
	}

	app := &App{
		PrimaryDB:   &ExampleDatabase{Name: "Primary-PostgreSQL"},
		SecondaryDB: &ExampleDatabase{Name: "Secondary-PostgreSQL"},
		Service:     &ExampleService2{},
	}

	ctx := context.Background()
	if err := autoinit.AutoInit(ctx, app); err != nil {
		panic(err)
	}

}

// Test the examples work correctly
func TestExamples(t *testing.T) {
	// Run the examples to ensure they work
	t.Run("BasicAs", func(t *testing.T) {
		ExampleAs()
	})

	t.Run("AsWithFilters", func(t *testing.T) {
		ExampleAs_withFilters()
	})
}
