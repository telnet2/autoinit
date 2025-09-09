package autoinit_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/rs/zerolog"
	"github.com/telnet2/autoinit"
	"gopkg.in/yaml.v3"
)

// Configuration structs that will be populated from YAML
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	MaxConns int    `yaml:"max_connections"`
}

type RedisConfig struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password"`
	Database int           `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout"`
}

type HTTPServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	TLS          bool          `yaml:"tls"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type AppConfig struct {
	Environment string           `yaml:"environment"`
	Database    DatabaseConfig   `yaml:"database"`
	Redis       RedisConfig      `yaml:"redis"`
	HTTPServer  HTTPServerConfig `yaml:"http_server"`
	Logging     LoggingConfig    `yaml:"logging"`
	Features    map[string]bool  `yaml:"features"`
	Services    []string         `yaml:"services"`
}

// Component implementations that use the configuration
type DatabaseComponent struct {
	Config    *DatabaseConfig
	Connected bool
	connPool  interface{} // Simulated connection pool
}

func (d *DatabaseComponent) Init(ctx context.Context, parent interface{}) error {
	// Discover configuration from parent - look for the embedded config
	if app, ok := parent.(*MicroserviceApp); ok {
		d.Config = &app.Config.Database
	} else {
		return fmt.Errorf("database configuration not found in parent")
	}

	// Simulate database connection
	fmt.Printf("🗄️  Connecting to database: %s@%s:%d/%s (max_conns: %d)\n",
		d.Config.Username, d.Config.Host, d.Config.Port, d.Config.Database, d.Config.MaxConns)

	d.Connected = true
	d.connPool = fmt.Sprintf("ConnectionPool[%s:%d]", d.Config.Host, d.Config.Port)

	return nil
}

type RedisCacheComponent struct {
	Config    *RedisConfig
	Connected bool
	client    interface{} // Simulated Redis client
}

func (c *RedisCacheComponent) Init(ctx context.Context, parent interface{}) error {
	// Discover Redis configuration from parent
	if app, ok := parent.(*MicroserviceApp); ok {
		c.Config = &app.Config.Redis
	} else {
		return fmt.Errorf("redis configuration not found in parent")
	}

	// Simulate Redis connection
	fmt.Printf("🔄 Connecting to Redis: %s:%d (db: %d, timeout: %v)\n",
		c.Config.Host, c.Config.Port, c.Config.Database, c.Config.Timeout)

	c.Connected = true
	c.client = fmt.Sprintf("RedisClient[%s:%d/%d]", c.Config.Host, c.Config.Port, c.Config.Database)

	return nil
}

type LoggerComponent struct {
	Config *LoggingConfig
	Logger interface{} // Simulated logger
}

func (l *LoggerComponent) Init(ctx context.Context, parent interface{}) error {
	// Discover logging configuration from parent
	if app, ok := parent.(*MicroserviceApp); ok {
		l.Config = &app.Config.Logging
	} else {
		return fmt.Errorf("logging configuration not found in parent")
	}

	fmt.Printf("📝 Initializing logger: level=%s, format=%s, output=%s\n",
		l.Config.Level, l.Config.Format, l.Config.Output)

	l.Logger = fmt.Sprintf("Logger[%s:%s]", l.Config.Level, l.Config.Format)

	return nil
}

type AuthService struct {
	db     *DatabaseComponent
	cache  *RedisCacheComponent
	logger *LoggerComponent
	Ready  bool
}

func (a *AuthService) Init(ctx context.Context, parent interface{}) error {
	// Discover dependencies using As pattern
	if !autoinit.As(ctx, a, parent, &a.db) {
		return fmt.Errorf("database component not found")
	}
	if !autoinit.As(ctx, a, parent, &a.cache) {
		return fmt.Errorf("cache component not found")
	}
	if !autoinit.As(ctx, a, parent, &a.logger) {
		return fmt.Errorf("logger component not found")
	}

	fmt.Printf("🔐 Initializing Auth Service with dependencies\n")
	a.Ready = true

	return nil
}

type HTTPServerComponent struct {
	Config     *HTTPServerConfig
	auth       *AuthService
	logger     *LoggerComponent
	Running    bool
	ServerAddr string
}

func (h *HTTPServerComponent) Init(ctx context.Context, parent interface{}) error {
	// Discover configuration from parent
	if app, ok := parent.(*MicroserviceApp); ok {
		h.Config = &app.Config.HTTPServer
	} else {
		return fmt.Errorf("HTTP server configuration not found")
	}
	if !autoinit.As(ctx, h, parent, &h.auth) {
		return fmt.Errorf("auth service not found")
	}
	if !autoinit.As(ctx, h, parent, &h.logger) {
		return fmt.Errorf("logger component not found")
	}

	// Initialize HTTP server
	h.ServerAddr = fmt.Sprintf("%s:%d", h.Config.Host, h.Config.Port)
	tlsStatus := "HTTP"
	if h.Config.TLS {
		tlsStatus = "HTTPS"
	}

	fmt.Printf("🌐 Starting %s server on %s (read_timeout: %v, write_timeout: %v)\n",
		tlsStatus, h.ServerAddr, h.Config.ReadTimeout, h.Config.WriteTimeout)

	h.Running = true

	return nil
}

type HealthCheckService struct {
	db     *DatabaseComponent
	cache  *RedisCacheComponent
	server *HTTPServerComponent
	Ready  bool
}

func (h *HealthCheckService) Init(ctx context.Context, parent interface{}) error {
	// Discover all components this service monitors
	autoinit.MustAs(ctx, h, parent, &h.db)     // Required dependency - will panic if not found
	autoinit.MustAs(ctx, h, parent, &h.cache)  // Required dependency - will panic if not found
	autoinit.MustAs(ctx, h, parent, &h.server) // Required dependency - will panic if not found

	fmt.Printf("💊 Health Check Service monitoring %d components\n", 3)
	h.Ready = true

	return nil
}

// Main application struct that composes all components
type MicroserviceApp struct {
	// Configuration - will be populated from YAML
	Config AppConfig `yaml:",inline"`

	// Infrastructure Components (automatically wired by AutoInit)
	DatabaseComponent *DatabaseComponent   `autoinit:"init"`
	CacheComponent    *RedisCacheComponent `autoinit:"init"`
	LoggerComponent   *LoggerComponent     `autoinit:"init"`

	// Business Logic Services (with dependency discovery)
	AuthService *AuthService         `autoinit:"init"`
	HTTPServer  *HTTPServerComponent `autoinit:"init"`

	// Monitoring Services
	HealthCheck *HealthCheckService `autoinit:"init"`

	// Application state
	Started   bool
	StartTime time.Time
}

func (app *MicroserviceApp) PreInit(ctx context.Context) error {
	fmt.Printf("🚀 Starting %s microservice...\n", app.Config.Environment)
	app.StartTime = time.Now()
	return nil
}

func (app *MicroserviceApp) PostInit(ctx context.Context) error {
	app.Started = true

	fmt.Printf("✅ Microservice startup completed in %s\n", "0s")
	fmt.Printf("📊 Components initialized: Database=%v, Cache=%v, Auth=%v, HTTP=%v, HealthCheck=%v\n",
		app.DatabaseComponent.Connected, app.CacheComponent.Connected, app.AuthService.Ready,
		app.HTTPServer.Running, app.HealthCheck.Ready)

	// Print feature flags status
	fmt.Printf("🎛️  Feature flags: ")

	// Sort features for deterministic output
	var features []string
	for feature := range app.Config.Features {
		features = append(features, feature)
	}
	sort.Strings(features)

	for i, feature := range features {
		enabled := app.Config.Features[feature]
		status := "❌"
		if enabled {
			status = "✅"
		}
		if i > 0 {
			fmt.Printf(" ")
		}
		fmt.Printf("%s %s", status, feature)
	}
	fmt.Printf("\n")

	return nil
}

// ExampleAutoInit_yamlConfiguration demonstrates YAML-driven configuration with AutoInit
func ExampleAutoInit_yamlConfiguration() {
	// 1. Define YAML configuration
	yamlConfig := `
environment: production

database:
  host: postgres.internal
  port: 5432
  database: myapp
  username: app_user
  password: secure_password_123
  max_connections: 50

redis:
  host: redis.internal
  port: 6379
  password: redis_password_456
  database: 0
  timeout: 5s

http_server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  tls: true

logging:
  level: info
  format: json
  output: stdout

features:
  user_registration: true
  advanced_analytics: false
  beta_features: true
  rate_limiting: true

services:
  - authentication
  - user_management
  - analytics
`

	// 2. Create application instance with components
	app := &MicroserviceApp{
		// Components will be auto-initialized by AutoInit
		DatabaseComponent: &DatabaseComponent{},
		CacheComponent:    &RedisCacheComponent{},
		LoggerComponent:   &LoggerComponent{},
		AuthService:       &AuthService{},
		HTTPServer:        &HTTPServerComponent{},
		HealthCheck:       &HealthCheckService{},
	}

	// 3. Parse YAML configuration into the app struct
	if err := yaml.Unmarshal([]byte(yamlConfig), app); err != nil {
		fmt.Printf("Failed to parse YAML config: %v\n", err)
		return
	}

	fmt.Printf("📄 Configuration loaded from YAML\n")
	fmt.Printf("📊 Environment: %s\n", app.Config.Environment)
	fmt.Printf("📊 Services configured: %v\n", app.Config.Services)

	// 4. One-shot initialization and dependency discovery with AutoInit
	ctx := context.Background()

	// Use silent logger for clean example output
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}

	fmt.Println("\n🔧 Starting AutoInit dependency discovery and initialization...")

	// This single call:
	// - Discovers all component dependencies using the As pattern
	// - Initializes components in the correct order
	// - Handles all lifecycle hooks (PreInit, Init, PostInit)
	// - Provides detailed error context if anything fails
	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return
	}

	// 5. Application is now fully initialized and ready to use
	fmt.Printf("\n🎉 Application ready! Running on %s\n", app.HTTPServer.ServerAddr)

	// Output:
	// 📄 Configuration loaded from YAML
	// 📊 Environment: production
	// 📊 Services configured: [authentication user_management analytics]
	//
	// 🔧 Starting AutoInit dependency discovery and initialization...
	// 🚀 Starting production microservice...
	// 🗄️  Connecting to database: app_user@postgres.internal:5432/myapp (max_conns: 50)
	// 🔄 Connecting to Redis: redis.internal:6379 (db: 0, timeout: 5s)
	// 📝 Initializing logger: level=info, format=json, output=stdout
	// 🔐 Initializing Auth Service with dependencies
	// 🌐 Starting HTTPS server on 0.0.0.0:8080 (read_timeout: 10s, write_timeout: 10s)
	// 💊 Health Check Service monitoring 3 components
	// ✅ Microservice startup completed in 0s
	// 📊 Components initialized: Database=true, Cache=true, Auth=true, HTTP=true, HealthCheck=true
	// 🎛️  Feature flags: ❌ advanced_analytics ✅ beta_features ✅ rate_limiting ✅ user_registration
	//
	// 🎉 Application ready! Running on 0.0.0.0:8080
}

// ExampleAutoInit_yamlFromFile demonstrates loading configuration from an actual YAML file
func ExampleAutoInit_yamlFromFile() {
	// Create a temporary YAML file
	yamlContent := `
environment: development

database:
  host: localhost
  port: 5432
  database: devdb
  username: dev_user
  password: dev_pass
  max_connections: 10

redis:
  host: localhost
  port: 6379
  password: ""
  database: 1
  timeout: 2s

http_server:
  host: localhost
  port: 3000
  read_timeout: 5s
  write_timeout: 5s
  tls: false

logging:
  level: debug
  format: text
  output: stdout

features:
  user_registration: true
  advanced_analytics: true
  beta_features: false
  rate_limiting: false
`

	// Write to temporary file
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		fmt.Printf("Failed to write config file: %v\n", err)
		return
	}
	tmpFile.Close()

	// Load configuration from file
	configData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		fmt.Printf("Failed to read config file: %v\n", err)
		return
	}

	// Create and configure application
	app := &MicroserviceApp{
		DatabaseComponent: &DatabaseComponent{},
		CacheComponent:    &RedisCacheComponent{},
		LoggerComponent:   &LoggerComponent{},
		AuthService:       &AuthService{},
		HTTPServer:        &HTTPServerComponent{},
		HealthCheck:       &HealthCheckService{},
	}

	// Parse YAML from file
	if err := yaml.Unmarshal(configData, app); err != nil {
		fmt.Printf("Failed to parse YAML: %v\n", err)
		return
	}

	fmt.Printf("📁 Configuration loaded from file: %s\n", "/tmp/config123.yaml")

	// Initialize with AutoInit
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}

	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return
	}

	fmt.Printf("🎉 Development environment ready on %s\n", app.HTTPServer.ServerAddr)

	// Output:
	// 📁 Configuration loaded from file: /tmp/config123.yaml
	// 🚀 Starting development microservice...
	// 🗄️  Connecting to database: dev_user@localhost:5432/devdb (max_conns: 10)
	// 🔄 Connecting to Redis: localhost:6379 (db: 1, timeout: 2s)
	// 📝 Initializing logger: level=debug, format=text, output=stdout
	// 🔐 Initializing Auth Service with dependencies
	// 🌐 Starting HTTP server on localhost:3000 (read_timeout: 5s, write_timeout: 5s)
	// 💊 Health Check Service monitoring 3 components
	// ✅ Microservice startup completed in 0s
	// 📊 Components initialized: Database=true, Cache=true, Auth=true, HTTP=true, HealthCheck=true
	// 🎛️  Feature flags: ✅ advanced_analytics ❌ beta_features ❌ rate_limiting ✅ user_registration
	// 🎉 Development environment ready on localhost:3000
}

// ExampleAutoInit_conditionalComponents demonstrates how to conditionally include components based on configuration
func ExampleAutoInit_conditionalComponents() {
	yamlConfig := `
environment: production

database:
  host: prod-db.internal
  port: 5432
  database: prodapp
  username: prod_user
  password: very_secure_password
  max_connections: 100

redis:
  host: prod-redis.internal
  port: 6379
  password: redis_prod_password
  database: 0
  timeout: 3s

http_server:
  host: 0.0.0.0
  port: 443
  read_timeout: 15s
  write_timeout: 15s
  tls: true

logging:
  level: warn
  format: json
  output: stdout

features:
  user_registration: true
  advanced_analytics: true
  beta_features: false
  rate_limiting: true
`

	// Parse configuration first
	var config AppConfig
	if err := yaml.Unmarshal([]byte(yamlConfig), &config); err != nil {
		fmt.Printf("Failed to parse config: %v\n", err)
		return
	}

	// Create app with conditional components based on configuration
	app := &MicroserviceApp{
		Config:            config,
		DatabaseComponent: &DatabaseComponent{},
		CacheComponent:    &RedisCacheComponent{},
		LoggerComponent:   &LoggerComponent{},
		AuthService:       &AuthService{},
		HTTPServer:        &HTTPServerComponent{},
	}

	// Conditionally add health check only in production
	if config.Environment == "production" {
		app.HealthCheck = &HealthCheckService{}
		fmt.Println("🏥 Health check enabled for production environment")
	}

	// Initialize everything
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}

	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return
	}

	fmt.Printf("🎉 Production application ready!\n")

	// Output:
	// 🏥 Health check enabled for production environment
	// 🚀 Starting production microservice...
	// 🗄️  Connecting to database: prod_user@prod-db.internal:5432/prodapp (max_conns: 100)
	// 🔄 Connecting to Redis: prod-redis.internal:6379 (db: 0, timeout: 3s)
	// 📝 Initializing logger: level=warn, format=json, output=stdout
	// 🔐 Initializing Auth Service with dependencies
	// 🌐 Starting HTTPS server on 0.0.0.0:443 (read_timeout: 15s, write_timeout: 15s)
	// 💊 Health Check Service monitoring 3 components
	// ✅ Microservice startup completed in 0s
	// 📊 Components initialized: Database=true, Cache=true, Auth=true, HTTP=true, HealthCheck=true
	// 🎛️  Feature flags: ✅ advanced_analytics ❌ beta_features ✅ rate_limiting ✅ user_registration
	// 🎉 Production application ready!
}
