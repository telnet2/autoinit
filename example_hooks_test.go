package autoinit_test

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/user/autoinit"
)

// Component with lifecycle hooks
type Component struct {
	Name   string
	Status string
}

func (c *Component) PreInit(ctx context.Context) error {
	c.Status = "initializing"
	fmt.Printf("Component %s: Starting initialization\n", c.Name)
	return nil
}

func (c *Component) Init(ctx context.Context) error {
	c.Status = "ready"
	fmt.Printf("Component %s: Initialized\n", c.Name)
	return nil
}

func (c *Component) PostInit(ctx context.Context) error {
	c.Status = "operational"
	fmt.Printf("Component %s: Post-initialization complete\n", c.Name)
	return nil
}

// System with field hooks to monitor component initialization
type System struct {
	Database Component
	Cache    Component
	API      Component
	
	// Track initialization order
	InitOrder []string
}

func (s *System) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	fmt.Printf("System: About to initialize %s\n", fieldName)
	s.InitOrder = append(s.InitOrder, "pre-"+fieldName)
	return nil
}

func (s *System) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	if comp, ok := fieldValue.(*Component); ok {
		fmt.Printf("System: %s is now %s\n", fieldName, comp.Status)
	}
	s.InitOrder = append(s.InitOrder, "post-"+fieldName)
	return nil
}

func (s *System) Init(ctx context.Context) error {
	fmt.Println("System: All components initialized")
	return nil
}

// Example_hooks demonstrates the hook system
func Example_hooks() {
	system := &System{
		Database: Component{Name: "Database"},
		Cache:    Component{Name: "Cache"},
		API:      Component{Name: "API"},
	}
	
	// Initialize with silent logger for cleaner example output
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	
	if err := autoinit.AutoInitWithOptions(ctx, system, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("\nInitialization order: %v\n", system.InitOrder)
	
	// Output:
	// System: About to initialize Database
	// Component Database: Starting initialization
	// Component Database: Initialized
	// Component Database: Post-initialization complete
	// System: Database is now operational
	// System: About to initialize Cache
	// Component Cache: Starting initialization
	// Component Cache: Initialized
	// Component Cache: Post-initialization complete
	// System: Cache is now operational
	// System: About to initialize API
	// Component API: Starting initialization
	// Component API: Initialized
	// Component API: Post-initialization complete
	// System: API is now operational
	// System: All components initialized
	//
	// Initialization order: [pre-Database post-Database pre-Cache post-Cache pre-API post-API]
}