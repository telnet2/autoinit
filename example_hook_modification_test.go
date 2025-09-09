package autoinit_test

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/user/autoinit"
)

// Config represents a service configuration
type Config struct {
	Host string
	Port int
}

func (c *Config) Init(ctx context.Context) error {
	// Set defaults if not already set
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	return nil
}

// ServiceContainer manages services with environment-specific configuration
type ServiceContainer struct {
	Environment string
	APIConfig   Config
	DBConfig    Config
}

func (s *ServiceContainer) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// Modify configurations based on environment before they're initialized
	config, ok := fieldValue.(*Config)
	if !ok {
		return nil
	}

	switch s.Environment {
	case "production":
		switch fieldName {
		case "APIConfig":
			config.Host = "api.prod.example.com"
			config.Port = 443
		case "DBConfig":
			config.Host = "db.prod.example.com"
			config.Port = 5432
		}
	case "staging":
		switch fieldName {
		case "APIConfig":
			config.Host = "api.staging.example.com"
			config.Port = 8443
		case "DBConfig":
			config.Host = "db.staging.example.com"
			config.Port = 5432
		}
		// For development, let the defaults from Init() be used
	}

	fmt.Printf("PreFieldInit: Setting %s config for %s environment\n", fieldName, s.Environment)
	return nil
}

func (s *ServiceContainer) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
	// Log the final configuration after initialization
	config, ok := fieldValue.(*Config)
	if !ok {
		return nil
	}

	fmt.Printf("PostFieldInit: %s configured at %s:%d\n", fieldName, config.Host, config.Port)
	return nil
}

// Example_hookModification demonstrates how hooks can modify fields
func Example_hookModification() {
	// Production environment
	prodContainer := &ServiceContainer{
		Environment: "production",
	}

	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}

	fmt.Println("=== Production Environment ===")
	if err := autoinit.AutoInitWithOptions(ctx, prodContainer, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Development environment
	devContainer := &ServiceContainer{
		Environment: "development",
	}

	fmt.Println("\n=== Development Environment ===")
	if err := autoinit.AutoInitWithOptions(ctx, devContainer, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Output:
	// === Production Environment ===
	// PreFieldInit: Setting APIConfig config for production environment
	// PostFieldInit: APIConfig configured at api.prod.example.com:443
	// PreFieldInit: Setting DBConfig config for production environment
	// PostFieldInit: DBConfig configured at db.prod.example.com:5432
	//
	// === Development Environment ===
	// PreFieldInit: Setting APIConfig config for development environment
	// PostFieldInit: APIConfig configured at localhost:8080
	// PreFieldInit: Setting DBConfig config for development environment
	// PostFieldInit: DBConfig configured at localhost:8080
}
