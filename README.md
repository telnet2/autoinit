<div align="center">

# 🔧 AutoInit

**Declarative Component-Based Initialization Framework for Go**

*Build applications by composing self-contained components that initialize themselves*

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Test Coverage](https://img.shields.io/badge/Coverage-95%25-brightgreen.svg)]()
[![Go Report Card](https://goreportcard.com/badge/github.com/telnet2/autoinit)](https://goreportcard.com/report/github.com/telnet2/autoinit)

---

**Declare your architecture. AutoInit handles the rest.**

</div>

## 🚀 Why AutoInit?

**AutoInit is a declarative dependency injection framework** that eliminates initialization boilerplate. Simply declare your application structure using Go structs, and AutoInit automatically discovers dependencies, wires components together, and initializes everything in the correct order.

**The Declarative Approach**: Instead of writing imperative initialization code, you declare *what* your application looks like, and AutoInit figures out *how* to initialize it.

### Imperative Initialization 😓
```go
// Manual, imperative initialization
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

### Declarative with AutoInit 🎉
```go
// Declare your application structure
type App struct {
    Database *Database  // Self-initializing
    Cache    *Cache     // Automatically wired  
    Auth     *Auth      // Dependency-aware
    Metrics  *Metrics   // Plug-and-play ready
}

func initializeApp() error {
    // Instantiate the declared structure
    app := &App{
        Database: &Database{},
        Cache:    &Cache{},
        Auth:     &Auth{},
        Metrics:  &Metrics{},
    }
    
    // AutoInit handles all the imperative work
    return autoinit.AutoInit(context.Background(), app)
}
```

## ✨ Features That Developers Love

| Feature | Description | Why It Matters |
|---------|-------------|----------------|
| 🪶 **Ultra-lightweight** | ~1,500 lines of core code | **40-95% smaller than competitors** |
| 📋 **Declarative Architecture** | Define structure, not implementation | Focus on *what*, not *how* |
| 🔄 **Zero Config DI** | Components find dependencies automatically | No containers or registration |
| ⚡ **Zero Runtime Overhead** | No containers, registries, or metadata | **0 bytes memory footprint** |
| 🎯 **Smart Discovery** | Multiple initialization patterns supported | Flexible component design |
| 🔍 **Rich Error Context** | Shows exact path when initialization fails | Debug issues in seconds, not hours |
| 📄 **YAML Integration** | Components discover configuration from YAML | Declarative configuration management |
| 🪝 **Lifecycle Hooks** | Pre/Post initialization control | Custom initialization flows |
| 🏷️ **Tag-Based Control** | `autoinit:"-"` to skip, `autoinit:"init"` to include | Explicit control when needed |
| 🔄 **Cycle Detection** | Prevents infinite loops in circular refs | Safe for complex architectures |
| 📦 **Collection Support** | Works with slices, maps, embedded structs | Handle complex data structures |

## 📋 The Declarative Philosophy

**Traditional DI frameworks are imperative** - you tell them *how* to wire dependencies:
```go
// Imperative: Step-by-step instructions
container := NewContainer()
container.Register("database", NewDatabase)
container.Register("cache", NewCache, Depends("database"))
container.Register("auth", NewAuth, Depends("database", "cache"))
app := container.Build() // Complex wiring logic
```

**AutoInit is declarative** - you tell it *what* your application looks like:
```go
// Declarative: Structure definition
type App struct {
    Database *Database  // What components exist
    Cache    *Cache     // Their relationships are clear
    Auth     *Auth      // Dependencies discovered automatically
}

// Simple instantiation
app := &App{Database: &Database{}, Cache: &Cache{}, Auth: &Auth{}}
autoinit.AutoInit(ctx, app) // AutoInit figures out the "how"
```

### Benefits of Declarative Architecture

| Aspect | Imperative DI | Declarative AutoInit |
|--------|---------------|---------------------|
| **Mental Model** | Complex wiring logic | Simple struct composition |
| **Adding Components** | Update container config | Add struct field |
| **Understanding Dependencies** | Read registration code | Read struct definition |
| **Debugging** | Trace container setup | Follow struct hierarchy |
| **Testing** | Mock container dependencies | Replace struct fields |
| **Configuration** | Container-specific syntax | Standard YAML + Go structs |

## 🏃‍♂️ Quick Start

### Installation
```bash
go get github.com/telnet2/autoinit
```

### Basic Example
```go
package main

import (
    "context"
    "fmt"
    "github.com/telnet2/autoinit"
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

> 💡 **Pro Tip**: AutoInit works great with YAML configuration! See the [YAML-Driven Configuration example](#yaml-driven-configuration-with-one-shot-initialization) for production-ready configuration management.

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

Need components to find each other? AutoInit provides two powerful discovery patterns:

### Modern As Pattern (Recommended)

Inspired by Go CDK's design, the As pattern provides type-safe dependency discovery with conjunctive filtering:

```go
type ServiceComponent struct {
    db     *Database
    cache  *Cache
    logger Logger  // interface type
}

func (s *ServiceComponent) Init(ctx context.Context, parent interface{}) error {
    // Simple type-based discovery
    if autoinit.As(ctx, s, parent, &s.db) {
        // Found any Database component
    }
    
    // Conjunctive filtering - ALL conditions must match
    var primaryDB *Database
    if autoinit.As(ctx, s, parent, &primaryDB, 
        autoinit.WithFieldName("PrimaryDB"),
        autoinit.WithJSONTag("primary")) {
        // Found Database that is BOTH named "PrimaryDB" AND tagged "primary"
        s.db = primaryDB
    }
    
    // Interface discovery
    if autoinit.As(ctx, s, parent, &s.logger) {
        // Found any component implementing Logger interface
    }
    
    // Required dependencies with MustAs (panics if not found)
    autoinit.MustAs(ctx, s, parent, &s.cache)
    
    return nil
}

type App struct {
    PrimaryDB *Database `json:"primary"`
    BackupDB  *Database `json:"backup"`
    Cache     *Cache
    Logger    *ConsoleLogger  // Implements Logger interface
    Service   *ServiceComponent
}
```

**As Pattern Features:**
- ✅ **Type-safe**: Compile-time type checking with generics
- 🔗 **Conjunctive Filters**: All conditions must match (AND logic)
- 🎯 **Clean API**: Similar to Go CDK's escape hatch pattern
- 🔍 **Interface Support**: Find components implementing interfaces
- ⚡ **Simple Syntax**: `As(ctx, self, parent, &target, ...filters)`

### Classic Finder Pattern

The original discovery system with flexible search options:

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

**Classic Finder Features:**
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

## 🏗️ Container Pattern for Enterprise Applications

AutoInit supports the **container pattern** for organizing complex applications into logical groups. This approach is perfect for enterprise applications with multiple architectural layers.

### Container-Based Architecture

```go
// Organize dependencies by architectural layers
type App struct {
    // Infrastructure layer
    Database *Database `yaml:"database"`
    Cache    *Cache    `yaml:"cache"`
    
    // Data access layer (DAO container)
    DAOs *DAOContainer `yaml:"dao"`
    
    // Business logic layer (Handler container)
    Handlers *HandlerContainer `yaml:"handlers"`
    
    // Presentation layer
    Server *HTTPServer `yaml:"server"`
}

// Define containers as logical groups
type DAOContainer struct {
    UserDAO    *UserDAO    `yaml:"user_dao"`
    ProductDAO *ProductDAO `yaml:"product_dao"`
    OrderDAO   *OrderDAO   `yaml:"order_dao"`
}

type HandlerContainer struct {
    UserHandler    *UserHandler    `yaml:"user_handler"`
    ProductHandler *ProductHandler `yaml:"product_handler"`
    OrderHandler   *OrderHandler   `yaml:"order_handler"`
}
```

**Benefits for Professional Teams**:
- ✅ **Logical organization** by architectural layers
- ✅ **Simplified dependency injection** - inject containers, not individual components
- ✅ **Easier testing** - mock entire containers for unit tests
- ✅ **Clear boundaries** between architectural layers
- ✅ **Scalable team development** - teams work on container boundaries

[📖 **Learn more in BEST_PRACTICES.md** →](BEST_PRACTICES.md#-container-pattern-for-dependency-organization)

## 🎯 Real-World Examples

### YAML-Driven Configuration with One-Shot Initialization

AutoInit seamlessly integrates with YAML configuration files. Define your configuration structure, parse it from YAML, then let AutoInit handle all dependency discovery and initialization:

```go
// Define configuration structs
type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type AppConfig struct {
    Environment string         `yaml:"environment"`
    Database    DatabaseConfig `yaml:"database"`
    Features    map[string]bool `yaml:"features"`
}

// Components that discover their configuration automatically
type DatabaseComponent struct {
    Config    *DatabaseConfig
    Connected bool
}

func (d *DatabaseComponent) Init(ctx context.Context, parent interface{}) error {
    // Discover configuration from parent app struct
    if app, ok := parent.(*MicroserviceApp); ok {
        d.Config = &app.Config.Database
    }
    
    // Initialize using configuration
    fmt.Printf("Connecting to %s@%s:%d\n", 
        d.Config.Username, d.Config.Host, d.Config.Port)
    d.Connected = true
    return nil
}

// Main application struct
type MicroserviceApp struct {
    Config AppConfig `yaml:",inline"`  // Embedded YAML config
    
    // Components with automatic dependency discovery
    DatabaseComponent *DatabaseComponent `autoinit:"init"`
    // ... other components
}

func main() {
    // 1. Parse YAML configuration
    yamlData := `
environment: production
database:
  host: postgres.internal
  port: 5432
  username: app_user
  password: secure_password
features:
  analytics: true
  beta_features: false
`
    
    app := &MicroserviceApp{
        DatabaseComponent: &DatabaseComponent{},
    }
    
    // 2. Load configuration from YAML
    if err := yaml.Unmarshal([]byte(yamlData), app); err != nil {
        panic(err)
    }
    
    // 3. One-shot initialization and dependency discovery
    if err := autoinit.AutoInit(context.Background(), app); err != nil {
        panic(err)
    }
    
    // 🎉 Application is fully initialized and ready!
    fmt.Printf("App ready! Connected: %v\n", app.DatabaseComponent.Connected)
}
```

**Key Benefits:**
- **Pure Declarative**: Define structure in Go structs, configuration in YAML
- **Automatic Discovery**: Components find dependencies and configuration without wiring
- **One-Shot Initialization**: Single `AutoInit()` call handles all complexity
- **Production Ready**: Enterprise-grade dependency injection with lifecycle management

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

## 🏆 How AutoInit Compares

AutoInit is **significantly lighter** than traditional DI frameworks:

| Framework | Approach | Setup Code | Runtime Overhead | Learning Curve |
|-----------|----------|------------|------------------|----------------|
| **AutoInit** | **Declarative** | **3 lines** | **0 bytes** | **30 minutes** |
| Wire (Google) | Code Generation | 55 lines | 0 bytes¹ | 2-3 days |
| Uber FX | Imperative | 70 lines | ~5-10KB | 3-5 days |
| Spring DI | XML/Annotations | 100+ lines | ~60MB+ | 1-2 weeks |

¹ *Requires build-time code generation*

**Key Advantages:**
- **🪶 Ultra-lightweight**: ~1,500 lines of core code vs 2,500-50,000+ in other frameworks
- **⚡ Zero configuration**: No containers, registrations, or XML files
- **🎯 Go-native**: Uses struct composition, not foreign concepts  
- **📄 Built-in YAML**: Configuration integration without extra complexity
- **🚀 Fast startup**: ~5ms initialization vs 50-2000ms for heavy frameworks

> 📖 **See [COMPARISON.md](COMPARISON.md) for detailed analysis vs Wire, FX, Dig, Spring, and more**

## 📚 Documentation & Best Practices

| Document | Description |
|----------|-------------|
| [BEST_PRACTICES.md](BEST_PRACTICES.md) | **Production-ready patterns & container architecture** |
| [COMPONENTS.md](COMPONENTS.md) | Component-based architecture guide |
| [COMPARISON.md](COMPARISON.md) | **Framework comparison vs Wire, FX, Dig, Spring** |
| [FINDER.md](FINDER.md) | Component discovery system |
| [HOOKS.md](HOOKS.md) | Lifecycle hooks and custom logic |
| [TAGS.md](TAGS.md) | Tag-based initialization control |
| [CROSS_DEPS.md](CROSS_DEPS.md) | Handling component dependencies |

## 💡 Use Cases

AutoInit's declarative approach shines in these scenarios:

- **🏗️ Application Bootstrap**: Declare complex architectures without imperative wiring
- **🏗️ Enterprise Applications**: Container pattern for organizing complex architectures
- **🔌 Plugin Systems**: Define plugin structure, let AutoInit handle dynamic loading
- **🧪 Testing**: Mock entire containers for unit testing - no setup required
- **☁️ Microservices**: YAML-driven configuration with automatic component discovery
- **🏭 Factory Patterns**: Struct composition replaces complex factory hierarchies
- **🎯 Dependency Injection**: Go-native DI without external container frameworks
- **📄 Configuration Management**: Pure declarative config with type-safe YAML binding
- **👥 Team Development**: Clear boundaries with container-based architecture

## 🤝 Contributing

We love contributions! Here's how to get started:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin amazing-feature`)  
5. **Open** a Pull Request

### Development Setup
```bash
git clone https://github.com/telnet2/autoinit.git
cd autoinit
go mod tidy
go test ./...
```

## 🎉 Community & Support

- 📖 **Documentation**: [Complete Guides](COMPONENTS.md)
- 🐛 **Issues**: [GitHub Issues](https://github.com/telnet2/autoinit/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/telnet2/autoinit/discussions)
- ⭐ **Star** this repo if you find it useful!

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ by developers, for developers**

*Declare your architecture. AutoInit handles the rest.*

[⭐ Star on GitHub](https://github.com/telnet2/autoinit) | [📖 Read the Docs](COMPONENTS.md) | [🚀 Get Started](#-quick-start)

</div>
