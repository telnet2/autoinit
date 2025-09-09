# Component-Based Architecture with AutoInit

## What is a Component?

In the AutoInit framework, a **component** is any struct that implements one of the initialization interfaces (`Init()`, `Init(ctx)`, or `Init(ctx, parent)`). Components are self-contained, pluggable units that can be composed into larger systems without requiring manual wiring.

## The Component Philosophy

### Traditional Go Initialization
```go
// Manual wiring required for each new component
app := &App{}
app.Database = NewDatabase()
app.Database.Init()
app.Cache = NewCache()
app.Cache.Init()
app.NewService = NewService()  // Added new component
app.NewService.Init()           // Must remember to initialize
```

### Component-Based with AutoInit
```go
// Just plug in the component - no code changes needed
type App struct {
    Database   *DatabaseComponent
    Cache      *CacheComponent
    NewService *ServiceComponent  // Just add it - automatically initialized!
}

app := &App{
    Database:   &DatabaseComponent{},
    Cache:      &CacheComponent{},
    NewService: &ServiceComponent{},  // Plug and play!
}
autoinit.AutoInit(ctx, app)  // All components initialized automatically
```

## Key Benefits of Component Architecture

### 1. **Plug-and-Play**
Add new components without modifying initialization code:
```go
type Application struct {
    Core    *CoreComponent
    Auth    *AuthComponent
    // Add new components anytime - they just work
    Metrics *MetricsComponent `autoinit:"init"`
    Tracing *TracingComponent `autoinit:"init"`
}
```

### 2. **Self-Contained**
Each component manages its own initialization:
```go
type DatabaseComponent struct {
    pool       *sql.DB
    config     Config
    migrations []Migration
}

func (d *DatabaseComponent) Init(ctx context.Context) error {
    // Component knows how to initialize itself
    d.pool = createPool(d.config)
    return d.runMigrations()
}
```

### 3. **Composable**
Components can contain other components:
```go
type APIComponent struct {
    Router   *RouterComponent
    Auth     *AuthComponent
    RateLimiter *RateLimiterComponent
}

type RouterComponent struct {
    Middleware []MiddlewareComponent
    Handlers   map[string]*HandlerComponent
}
```

### 4. **Testable**
Swap components with mocks easily:
```go
// Production
app := &App{
    Database: &PostgresComponent{},
    Cache:    &RedisComponent{},
}

// Testing
app := &App{
    Database: &MockDatabaseComponent{},
    Cache:    &InMemoryCacheComponent{},
}
// Same initialization code!
autoinit.AutoInit(ctx, app)
```

## Component Patterns

### Basic Component
```go
// Any struct with an Init method is a component
type LoggerComponent struct {
    level  string
    output io.Writer
}

func (l *LoggerComponent) Init(ctx context.Context) error {
    // Self-contained initialization
    return l.configure()
}
```

### Nested Components
```go
type ServiceComponent struct {
    // These are all components that will be auto-initialized
    Logger   *LoggerComponent
    Database *DatabaseComponent
    Cache    *CacheComponent
    Queue    *QueueComponent
}

func (s *ServiceComponent) Init(ctx context.Context) error {
    // Child components are already initialized when this runs
    s.Logger.Info("Service initialized with all dependencies ready")
    return nil
}
```

### Optional Components
```go
type AppComponent struct {
    // Required components
    Core     *CoreComponent    `autoinit:"init"`
    Database *DatabaseComponent `autoinit:"init"`
    
    // Optional components (with RequireTags=true)
    Analytics *AnalyticsComponent  // Not initialized unless tagged
    Debug     *DebugComponent      // Not initialized unless tagged
}
```

### Component Arrays
```go
type PluginSystem struct {
    // All plugins are components and will be initialized
    Plugins []PluginComponent
    
    // Dynamic component registration
    RegisteredComponents map[string]Component
}
```

## Component Lifecycle

### 1. Discovery Phase
AutoInit recursively discovers all components in the struct tree.

### 2. Initialization Order
Components are initialized depth-first, in declaration order:
```go
type App struct {
    Config   *ConfigComponent   // 1st: Initialized first
    Database *DatabaseComponent // 2nd: After config
    Services []ServiceComponent // 3rd: After database
}
```

### 3. Parent-Child Relationships
```go
type ParentComponent struct {
    Child *ChildComponent
}

// Child can reference its parent during initialization
func (c *ChildComponent) Init(ctx context.Context, parent interface{}) error {
    if p, ok := parent.(*ParentComponent); ok {
        // Access parent component
    }
    return nil
}
```

## Best Practices

### 1. **Keep Components Focused**
Each component should have a single responsibility:
```go
// Good: Focused components
type AuthComponent struct{}
type DatabaseComponent struct{}
type CacheComponent struct{}

// Bad: Monolithic component
type EverythingComponent struct {
    // Too many responsibilities
}
```

### 2. **Use Interfaces for Flexibility**
```go
type Component interface {
    Init(context.Context) error
}

type StorageComponent interface {
    Component
    Store(key string, value interface{}) error
    Retrieve(key string) (interface{}, error)
}
```

### 3. **Document Component Dependencies**
```go
type ServiceComponent struct {
    // Required components - will panic if nil
    Database *DatabaseComponent `autoinit:"init" required:"true"`
    
    // Optional components - can be nil
    Cache    *CacheComponent   `autoinit:"init" optional:"true"`
}
```

### 4. **Make Components Idempotent**
```go
func (c *Component) Init(ctx context.Context) error {
    if c.initialized {
        return nil  // Safe to call multiple times
    }
    // ... initialization logic
    c.initialized = true
    return nil
}
```

## Component Testing

### Unit Testing Individual Components
```go
func TestDatabaseComponent(t *testing.T) {
    component := &DatabaseComponent{
        Config: testConfig,
    }
    
    err := component.Init(context.Background())
    assert.NoError(t, err)
    assert.True(t, component.IsConnected())
}
```

### Integration Testing Component Trees
```go
func TestApplicationComponents(t *testing.T) {
    app := &Application{
        Database: &MockDatabaseComponent{},
        Cache:    &MockCacheComponent{},
        Service:  &ServiceComponent{},
    }
    
    err := autoinit.AutoInit(context.Background(), app)
    assert.NoError(t, err)
    
    // All components initialized and wired together
    assert.True(t, app.Service.IsReady())
}
```

## Component Registry Pattern

For dynamic component systems:

```go
type ComponentRegistry struct {
    components map[string]Component
    mu         sync.RWMutex
}

func (r *ComponentRegistry) Register(name string, component Component) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.components[name] = component
}

func (r *ComponentRegistry) Init(ctx context.Context) error {
    // AutoInit doesn't know about future components
    // but the registry can initialize them as they're added
    for name, component := range r.components {
        if err := autoinit.AutoInit(ctx, component); err != nil {
            return fmt.Errorf("failed to init component %s: %w", name, err)
        }
    }
    return nil
}
```

## Real-World Example: Microservice Components

```go
// Each microservice is built from pluggable components
type Microservice struct {
    // Core components
    Config      *ConfigComponent      `autoinit:"init"`
    Logger      *LoggerComponent      `autoinit:"init"`
    Metrics     *MetricsComponent     `autoinit:"init"`
    Health      *HealthComponent      `autoinit:"init"`
    
    // Service-specific components
    Database    *DatabaseComponent    `autoinit:"init"`
    Cache       *CacheComponent       `autoinit:"init"`
    MessageBus  *MessageBusComponent  `autoinit:"init"`
    
    // API components
    HTTPServer  *HTTPServerComponent  `autoinit:"init"`
    GRPCServer  *GRPCServerComponent  `autoinit:"init"`
    
    // Optional components (controlled by config)
    Profiler    *ProfilerComponent    // Only in debug mode
    Tracing     *TracingComponent     // Only in production
}

// Starting the service is simple
func main() {
    service := &Microservice{
        Config:     &ConfigComponent{Path: "config.yaml"},
        Logger:     &LoggerComponent{Level: "info"},
        Metrics:    &PrometheusComponent{Port: 9090},
        Health:     &HealthComponent{Port: 8080},
        Database:   &PostgresComponent{},
        Cache:      &RedisComponent{},
        MessageBus: &KafkaComponent{},
        HTTPServer: &HTTPServerComponent{Port: 3000},
        GRPCServer: &GRPCServerComponent{Port: 3001},
    }
    
    // All components initialized with one call
    if err := autoinit.AutoInit(context.Background(), service); err != nil {
        log.Fatal(err)
    }
    
    // Service is ready with all components initialized
    service.HTTPServer.Start()
}
```

## Summary

The component model with AutoInit enables:
- **Modularity**: Build systems from independent, reusable components
- **Flexibility**: Add/remove components without changing initialization code
- **Testability**: Easy to mock and test components in isolation
- **Maintainability**: Each component is self-contained and focused
- **Scalability**: Compose simple components into complex systems

Think of components as LEGO blocks - each piece knows how to connect itself, and you can build complex structures by simply putting them together.