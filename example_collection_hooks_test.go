package autoinit_test

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/user/autoinit"
)

// MicroService represents a microservice
type MicroService struct {
	Name   string
	Status string
}

func (s *MicroService) Init(ctx context.Context) error {
	s.Status = "running"
	return nil
}

// ServiceManager manages multiple services
type ServiceManager struct {
	Services   map[string]*MicroService
	Backups    []*MicroService
	InitCounts map[string]int // Track how many times hooks are called
}

func (m *ServiceManager) PreInit(ctx context.Context) error {
	// Initialize the counts map before any fields are processed
	m.InitCounts = make(map[string]int)
	return nil
}

func (m *ServiceManager) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	switch fieldName {
	case "Services":
		// Receive the entire map before any services are initialized
		if services, ok := fieldValue.(*map[string]*MicroService); ok {
			fmt.Printf("About to initialize %d services\n", len(*services))
			// We could add more services here if needed
			(*services)["monitoring"] = &MicroService{Name: "monitoring"}
		}
	case "Backups":
		// Receive the entire slice before any backups are initialized
		if backups, ok := fieldValue.(*[]*MicroService); ok {
			fmt.Printf("About to initialize %d backup services\n", len(*backups))
		}
	}
	
	if m.InitCounts != nil {
		m.InitCounts["pre-"+fieldName]++
	}
	return nil
}

func (m *ServiceManager) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	switch fieldName {
	case "Services":
		// Receive the entire map after all services are initialized
		if services, ok := fieldValue.(*map[string]*MicroService); ok {
			running := 0
			for _, svc := range *services {
				if svc.Status == "running" {
					running++
				}
			}
			fmt.Printf("All services initialized: %d/%d running\n", running, len(*services))
		}
	case "Backups":
		// Receive the entire slice after all backups are initialized
		if backups, ok := fieldValue.(*[]*MicroService); ok {
			running := 0
			for _, svc := range *backups {
				if svc != nil && svc.Status == "running" {
					running++
				}
			}
			fmt.Printf("All backups initialized: %d/%d running\n", running, len(*backups))
		}
	}
	
	if m.InitCounts != nil {
		m.InitCounts["post-"+fieldName]++
	}
	return nil
}

// Example_collectionHooks demonstrates hooks with collection fields
func Example_collectionHooks() {
	manager := &ServiceManager{
		Services: map[string]*MicroService{
			"api":  {Name: "api"},
			"auth": {Name: "auth"},
			"db":   {Name: "db"},
		},
		Backups: []*MicroService{
			{Name: "backup1"},
			{Name: "backup2"},
		},
	}
	
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	
	if err := autoinit.AutoInitWithOptions(ctx, manager, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	// Check that hooks were called once per collection, not per element
	fmt.Printf("\nHook call counts:\n")
	fmt.Printf("  pre-Services: %d (called once for the map)\n", manager.InitCounts["pre-Services"])
	fmt.Printf("  post-Services: %d (called once for the map)\n", manager.InitCounts["post-Services"])
	fmt.Printf("  pre-Backups: %d (called once for the slice)\n", manager.InitCounts["pre-Backups"])
	fmt.Printf("  post-Backups: %d (called once for the slice)\n", manager.InitCounts["post-Backups"])
	
	// The monitoring service we added in PreFieldInit should be initialized
	if monitoring, ok := manager.Services["monitoring"]; ok {
		fmt.Printf("\nDynamically added service: %s is %s\n", monitoring.Name, monitoring.Status)
	}
	
	// Output:
	// About to initialize 3 services
	// All services initialized: 4/4 running
	// About to initialize 2 backup services
	// All backups initialized: 2/2 running
	//
	// Hook call counts:
	//   pre-Services: 1 (called once for the map)
	//   post-Services: 1 (called once for the map)
	//   pre-Backups: 1 (called once for the slice)
	//   post-Backups: 1 (called once for the slice)
	//
	// Dynamically added service: monitoring is running
}