package autoinit_test

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/telnet2/autoinit"
)

// Connection pool for database
type ConnectionPool struct {
	MaxConnections int
	Active         int
}

func (c *ConnectionPool) Init(ctx context.Context) error {
	c.MaxConnections = 10
	c.Active = 0
	return nil
}

// Cache configuration
type CacheConfig struct {
	TTL     int
	MaxSize int
}

func (c *CacheConfig) Init(ctx context.Context) error {
	c.TTL = 300
	c.MaxSize = 1000
	return nil
}

// Cache layer
type Cache struct {
	Config CacheConfig
	Ready  bool
}

func (c *Cache) Init(ctx context.Context) error {
	c.Ready = true
	return nil
}

// Database engine with nested components
type DatabaseEngine struct {
	Pool  ConnectionPool
	Cache Cache
	Ready bool
}

func (d *DatabaseEngine) Init(ctx context.Context) error {
	d.Ready = true
	return nil
}

// API server
type APIServer struct {
	Port    int
	Running bool
}

func (a *APIServer) Init(ctx context.Context) error {
	a.Port = 8080
	a.Running = true
	return nil
}

// Microservice with nested structures
type Microservice struct {
	Name     string
	Database *DatabaseEngine
	API      APIServer
	Ready    bool
}

func (m *Microservice) Init(ctx context.Context) error {
	m.Ready = true
	return nil
}

// ExampleAutoInit_nestedStructures demonstrates initialization of deeply nested structures
func ExampleAutoInit_nestedStructures() {
	service := &Microservice{
		Name:     "UserService",
		Database: &DatabaseEngine{},
	}

	// Initialize the entire tree with silent logger for examples
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	if err := autoinit.AutoInitWithOptions(ctx, service, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Service ready: %v\n", service.Ready)
	fmt.Printf("Database ready: %v\n", service.Database.Ready)
	fmt.Printf("API running on port: %d\n", service.API.Port)
	fmt.Printf("Cache TTL: %d seconds\n", service.Database.Cache.Config.TTL)
	fmt.Printf("Max connections: %d\n", service.Database.Pool.MaxConnections)

	// Output:
	// Service ready: true
	// Database ready: true
	// API running on port: 8080
	// Cache TTL: 300 seconds
	// Max connections: 10
}
