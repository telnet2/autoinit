package autoinit_test

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/telnet2/autoinit"
)

// Plugin represents a system plugin
type Plugin struct {
	Name    string
	Version string
	Active  bool
}

func (p *Plugin) Init(ctx context.Context) error {
	if p.Name == "" {
		return errors.New("plugin name is required")
	}
	p.Active = true
	return nil
}

// Service represents a microservice
type Service struct {
	ID       int
	Name     string
	Endpoint string
	Ready    bool
}

func (s *Service) Init(ctx context.Context) error {
	if s.Name == "" {
		return errors.New("service name is required")
	}
	s.Endpoint = fmt.Sprintf("http://localhost:8080/%s", s.Name)
	s.Ready = true
	return nil
}

// SystemManager manages services and plugins
type SystemManager struct {
	Name     string
	Services []Service          // Slice of structs
	Plugins  map[string]*Plugin // Map with pointer values
	Ready    bool
}

func (s *SystemManager) Init(ctx context.Context) error {
	s.Ready = true
	return nil
}

// ExampleAutoInit_slicesAndMaps demonstrates initialization of collections
func ExampleAutoInit_slicesAndMaps() {
	system := &SystemManager{
		Name: "Production",
		Services: []Service{
			{ID: 1, Name: "auth"},
			{ID: 2, Name: "api"},
			{ID: 3, Name: "webhook"},
		},
		Plugins: map[string]*Plugin{
			"cache": {Name: "cache", Version: "1.0"},
			"log":   {Name: "log", Version: "2.1"},
		},
	}

	// Initialize everything with silent logger for examples
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	if err := autoinit.AutoInitWithOptions(ctx, system, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("System ready: %v\n", system.Ready)
	fmt.Printf("Service count: %d\n", len(system.Services))
	fmt.Printf("First service endpoint: %s\n", system.Services[0].Endpoint)
	fmt.Printf("Cache plugin active: %v\n", system.Plugins["cache"].Active)

	// Output:
	// System ready: true
	// Service count: 3
	// First service endpoint: http://localhost:8080/auth
	// Cache plugin active: true
}

// FailingDatabase for error example
type FailingDatabase struct {
	ShouldFail bool
}

func (d *FailingDatabase) Init(ctx context.Context) error {
	if d.ShouldFail {
		return errors.New("connection refused")
	}
	return nil
}

// FailingService for error example
type FailingService struct {
	Name     string
	Database FailingDatabase
}

// ExampleAutoInit_errorHandling demonstrates error propagation
func ExampleAutoInit_errorHandling() {
	service := &FailingService{
		Name: "TestService",
		Database: FailingDatabase{
			ShouldFail: true,
		},
	}

	// Try to initialize - this will fail
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	err := autoinit.AutoInitWithOptions(ctx, service, options)
	if err != nil {
		// Extract detailed error information
		if initErr, ok := err.(*autoinit.InitError); ok {
			path := initErr.GetPath()
			fmt.Printf("Failed at: %s.%s\n", path[0], path[len(path)-1])
			fmt.Printf("Error: %v\n", initErr.Unwrap())
		}
	}

	// Output:
	// Failed at: Database.Database
	// Error: connection refused
}
