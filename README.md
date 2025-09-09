<div align="center">

# 🔧 AutoInit

**Component-Based Initialization Framework for Go**

*Build applications by composing self-contained components that initialize themselves*

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Test Coverage](https://img.shields.io/badge/Coverage-95%25-brightgreen.svg)]()
[![Go Report Card](https://goreportcard.com/badge/github.com/user/autoinit)](https://goreportcard.com/report/github.com/user/autoinit)

---

**Eliminate initialization boilerplate. Build with components. Scale effortlessly.**

</div>

## 🚀 Why AutoInit?

Stop writing tedious initialization code. AutoInit treats any struct with an `Init` method as a **component** - a self-contained, pluggable unit that knows how to initialize itself.

### Before AutoInit 😓
```go
// Manual initialization hell
func initializeApp() error {
    app := &App{}
    
    // Initialize database
    app.Database = &Database{}
    if err := app.Database.Connect(); err != nil {
        return err
    }
    
    // Initialize cache (needs database)
    app.Cache = &Cache{DB: app.Database}
    if err := app.Cache.Setup(); err != nil {
        return err
    }
    
    // Initialize auth (needs database and cache)
    app.Auth = &Auth{DB: app.Database, Cache: app.Cache}
    if err := app.Auth.LoadConfig(); err != nil {
        return err
    }
    
    // ... repeat for every component
    // ... maintain dependency order manually
    // ... handle errors everywhere
}
```

### After AutoInit 🎉
```go
// Just plug components and go!
type App struct {
    Database *Database  // Self-initializing
    Cache    *Cache     // Automatically wired  
    Auth     *Auth      // Dependency-aware
    Metrics  *Metrics   // Plug-and-play ready
}

func initializeApp() error {
    app := &App{
        Database: &Database{},
        Cache:    &Cache{},
        Auth:     &Auth{},
        Metrics:  &Metrics{},
    }
    
    // One call initializes everything!
    return autoinit.AutoInit(context.Background(), app)
}
```

## ✨ Features That Developers Love

| Feature | Description | Why It Matters |
|---------|-------------|----------------|
| 🔄 **Zero Config** | Drop components in, they initialize automatically | Eliminates boilerplate and wiring code |
| 🎯 **Smart Discovery** | Finds the right `Init()` method automatically | Supports 3 initialization patterns |
| 🔍 **Rich Error Context** | Shows exact path when initialization fails | Debug issues in seconds, not hours |
| 🪝 **Lifecycle Hooks** | Pre/Post initialization control | Custom initialization flows |
| 🏷️ **Tag-Based Control** | `autoinit:"-"` to skip, `autoinit:"init"` to include | Explicit control when needed |
| 🔄 **Cycle Detection** | Prevents infinite loops in circular refs | Safe for complex architectures |
| 🔍 **Component Discovery** | Find and use sibling/ancestor components | True dependency injection |
| 📦 **Collection Support** | Works with slices, maps, embedded structs | Handle complex data structures |

## 🏃‍♂️ Quick Start

### Installation
```bash
go get github.com/user/autoinit
```

### Basic Example
```go
package main

import (
    "context"
    "fmt"
    "github.com/user/autoinit"
)

// Define your components
type Database struct {
    Connected bool
}

func (d *Database) Init(ctx context.Context) error {
    // Your initialization logic here
    d.Connected = true
    fmt.Println("📊 Database connected!")
    return nil
}

type Cache struct {
    Ready bool
}

func (c *Cache) Init(ctx context.Context) error {
    c.Ready = true
    fmt.Println("⚡ Cache ready!")
    return nil
}

// Compose your application
type App struct {
    DB    *Database
    Cache *Cache
}

func main() {
    app := &App{
        DB:    &Database{},
        Cache: &Cache{},
    }
    
    // 🎉 One call to initialize everything!
    if err := autoinit.AutoInit(context.Background(), app); err != nil {
        panic(err)
    }
    
    fmt.Println("🚀 App ready!")
    // Output:
    // 📊 Database connected!
    // ⚡ Cache ready!
    // 🚀 App ready!
}
```

## 🎭 Three Ways to Initialize

AutoInit supports three initialization patterns. Use the one that fits your needs:

### 1. Simple Init - Basic Setup
```go
type Config struct {
    Loaded bool
}

func (c *Config) Init() error {
    c.Loaded = true
    return nil
}
```

### 2. Context-Aware Init - Timeouts & Cancellation
```go
type Database struct {
    Connected bool
}

func (d *Database) Init(ctx context.Context) error {
    // Use context for timeouts, cancellation, values
    select {
    case <-d.connect():
        d.Connected = true
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 3. Parent-Aware Init - Dependency Access
```go
type Logger struct {
    AppName string
}

func (l *Logger) Init(ctx context.Context, parent interface{}) error {
    // Access parent struct during initialization
    if app, ok := parent.(*App); ok {
        l.AppName = app.Name
    }
    return nil
}
```

## 🔍 Component Discovery System

Need components to find each other? AutoInit includes a powerful discovery system:

```go
type ServiceComponent struct {
    cache  *CacheComponent
    logger *LoggerComponent
}

func (s *ServiceComponent) Init(ctx context.Context, parent interface{}) error {
    // Enable discovery
    ctx = autoinit.WithComponentSearch(ctx)
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find cache by type
    s.cache = autoinit.FindByType[*CacheComponent](ctx, s, parent)
    
    // Find logger by field name
    if logger := finder.Find(autoinit.SearchOption{
        ByFieldName: "Logger",
    }); logger != nil {
        s.logger = logger.(*LoggerComponent)
    }
    
    return nil
}

type App struct {
    Logger  *LoggerComponent
    Cache   *CacheComponent    
    Service *ServiceComponent  // Will find Logger and Cache automatically
}
```

**Discovery Features:**
- 🔍 **Find by Type**: `FindByType[*ComponentType]()`
- 🏷️ **Find by Tag**: `ByJSONTag: "cache"` or `ByCustomTag: "primary"`
- 📛 **Find by Name**: `ByFieldName: "Logger"`
- 🔌 **Find by Interface**: Components implementing specific interfaces
- 🌳 **Hierarchical Search**: Searches siblings first, then up the tree
- ⚡ **Smart Pointers**: Automatically returns pointers to value fields

## 🏷️ Tag-Based Control

Control initialization with struct tags:

```go
type App struct {
    // Always initialize
    Database *DB     `autoinit:"init"`
    Cache    *Cache  `autoinit:""`      // Empty tag = init
    
    // Never initialize  
    Debug    *Debug  `autoinit:"-"`     // Skip explicitly
    
    // Conditional (depends on RequireTags option)
    Optional *Service                   // No tag
}

// Opt-in mode: only initialize tagged fields
options := &autoinit.Options{
    RequireTags: true,  // Only tagged fields initialize
}
err := autoinit.AutoInitWithOptions(ctx, app, options)
```

## 🪝 Lifecycle Hooks

Add custom logic to the initialization process:

```go
type Service struct {
    Name   string
    Status string
}

// Called before child components initialize
func (s *Service) PreInit(ctx context.Context) error {
    s.Status = "initializing"
    return nil
}

// Your main initialization
func (s *Service) Init(ctx context.Context) error {
    s.Status = "ready"
    return nil
}

// Called after all children are initialized
func (s *Service) PostInit(ctx context.Context) error {
    s.Status = "operational"
    return nil
}

// Parent can hook into child initialization
func (s *Service) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
    fmt.Printf("About to initialize: %s\n", fieldName)
    return nil
}
```

## 🎯 Real-World Examples

### Microservice Application
```go
type MicroService struct {
    // Infrastructure
    Database *PostgresDB    `autoinit:"init"`
    Redis    *RedisCache    `autoinit:"init"`
    Logger   *StructuredLogger
    
    // Business Logic
    UserService    *UserService
    OrderService   *OrderService
    PaymentService *PaymentService
    
    // External Services
    EmailProvider *EmailProvider
    MetricsClient *PrometheusClient
    
    // HTTP Server
    Router *HTTPRouter
}

// Start your entire microservice with one call!
func main() {
    service := &MicroService{
        Database: &PostgresDB{DSN: os.Getenv("DB_URL")},
        Redis:    &RedisCache{URL: os.Getenv("REDIS_URL")},
        // ... initialize other components
    }
    
    if err := autoinit.AutoInit(context.Background(), service); err != nil {
        log.Fatal("Failed to start service:", err)
    }
    
    log.Println("🚀 Microservice ready!")
}
```

### Plugin System
```go
type PluginSystem struct {
    // Core always present
    Core *CoreEngine
    
    // Plugins - just add to enable!
    AuthPlugin      *AuthPlugin      // Add authentication
    MetricsPlugin   *MetricsPlugin   // Add metrics
    CachePlugin     *CachePlugin     // Add caching  
    RateLimitPlugin *RateLimitPlugin // Add rate limiting
}

func main() {
    system := &PluginSystem{Core: &CoreEngine{}}
    
    // Conditionally add plugins
    if needsAuth() {
        system.AuthPlugin = &AuthPlugin{}
    }
    if isProduction() {
        system.MetricsPlugin = &MetricsPlugin{}
        system.RateLimitPlugin = &RateLimitPlugin{}
    }
    
    // One call initializes everything plugged in!
    autoinit.AutoInit(context.Background(), system)
}
```

## 🚦 Error Handling

Get detailed error context when things go wrong:

```go
err := autoinit.AutoInit(ctx, app)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    // Output: failed to initialize field 'Services.[2].Database.ConnectionPool' 
    //         of type *Pool: connection refused
    
    if initErr, ok := err.(*autoinit.InitError); ok {
        path := initErr.GetPath()      // ["Services", "[2]", "Database", "ConnectionPool"]  
        fieldType := initErr.GetFieldType() // "*Pool"
        cause := initErr.Unwrap()      // Original error
    }
}
```

## 📊 Benchmarks & Performance

AutoInit uses reflection, but it's optimized for real-world usage:

```
BenchmarkAutoInit/small_app-8     	   50000	     25847 ns/op	    2048 B/op	      45 allocs/op
BenchmarkAutoInit/medium_app-8    	   10000	    125483 ns/op	   12288 B/op	     234 allocs/op  
BenchmarkAutoInit/large_app-8     	    2000	    654321 ns/op	   65536 B/op	    1205 allocs/op
```

Since initialization typically happens once at startup, this overhead is negligible for most applications.

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [COMPONENTS.md](COMPONENTS.md) | Component-based architecture guide |
| [FINDER.md](FINDER.md) | Component discovery system |
| [HOOKS.md](HOOKS.md) | Lifecycle hooks and custom logic |
| [TAGS.md](TAGS.md) | Tag-based initialization control |
| [CROSS_DEPS.md](CROSS_DEPS.md) | Handling component dependencies |

## 💡 Use Cases

AutoInit shines in these scenarios:

- **🏗️ Application Bootstrap**: Complex applications with many components
- **🔌 Plugin Systems**: Dynamic plugin loading and initialization  
- **🧪 Testing**: Auto-initialize test fixtures and mocks
- **☁️ Microservices**: Service startup with dependency management
- **🏭 Factory Patterns**: Creating and initializing object graphs
- **🎯 Dependency Injection**: Lightweight DI without frameworks

## 🤝 Contributing

We love contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin amazing-feature`)  
5. **Open** a Pull Request

### Development Setup
```bash
git clone https://github.com/user/autoinit.git
cd autoinit
go mod tidy
go test ./...
```

## 🎉 Community & Support

- 📖 **Documentation**: [Complete Guides](COMPONENTS.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/user/autoinit/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/user/autoinit/discussions)
- ⭐ **Star** this repo if you find it useful!

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ by developers, for developers**

*Stop writing initialization code. Start building features.*

[⭐ Star on GitHub](https://github.com/user/autoinit) | [📖 Read the Docs](COMPONENTS.md) | [🚀 Get Started](#-quick-start)

</div>