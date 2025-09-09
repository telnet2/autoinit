# Cross-Component Dependencies

Strategies for handling dependencies between components while maintaining the plug-and-play architecture of autoinit.

## The Challenge

While autoinit makes it easy to plug components into your application and have them automatically initialized, real-world applications often have components that need to interact with each other. This document explores patterns for handling these cross-component dependencies while preserving the benefits of component-based architecture.

## Pattern 1: Component Registry

Store components in a registry accessible through context, allowing components to discover each other.

```go
// Registry stored in context
type ComponentRegistry struct {
    components map[reflect.Type]interface{}
    mu         sync.RWMutex
}

// Component that registers itself
type CacheComponent struct {
    registry *ComponentRegistry
}

func (c *CacheComponent) Init(ctx context.Context) error {
    // Get registry from context
    if reg := ctx.Value(registryKey).(*ComponentRegistry); reg != nil {
        c.registry = reg
        reg.Register(c)
    }
    return nil
}

// Component that uses the cache
type ServiceComponent struct {
    cache *CacheComponent
}

func (s *ServiceComponent) Init(ctx context.Context) error {
    // Find cache in registry
    if reg := ctx.Value(registryKey).(*ComponentRegistry); reg != nil {
        if cache := reg.Get((*CacheComponent)(nil)); cache != nil {
            s.cache = cache.(*CacheComponent)
        }
    }
    return nil
}

// Usage
app := &App{
    Cache:   &CacheComponent{},
    Service: &ServiceComponent{},
}

ctx := context.WithValue(context.Background(), registryKey, NewRegistry())
autoinit.AutoInit(ctx, app)
```

**Pros:**
- Components remain loosely coupled
- Dynamic discovery at runtime
- No compile-time dependencies

**Cons:**
- Requires context setup
- Runtime type assertions
- No compile-time safety

## Pattern 2: Type-Based Ancestor Search

Components can search up the parent chain for specific types they depend on.

```go
// Modified Init with parent reference
func (c *ChildComponent) Init(ctx context.Context, parent interface{}) error {
    // Search up the parent chain for DatabaseComponent
    if db := findAncestor[*DatabaseComponent](parent); db != nil {
        c.database = db
    }
    return nil
}

// Helper function to search ancestors
func findAncestor[T any](parent interface{}) T {
    var zero T
    if parent == nil {
        return zero
    }
    
    // Check if parent is the type we want
    if match, ok := parent.(T); ok {
        return match
    }
    
    // Use reflection to search fields
    v := reflect.ValueOf(parent)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    
    for i := 0; i < v.NumField(); i++ {
        field := v.Field(i)
        if match, ok := field.Interface().(T); ok {
            return match
        }
    }
    
    return zero
}
```

**Pros:**
- Natural parent-child relationships
- No global state needed
- Works with existing Init(ctx, parent) signature

**Cons:**
- Limited to ancestor components
- Requires parent parameter support
- Can be fragile if structure changes

## Pattern 3: Service Locator Chain

Each level in the hierarchy can provide services to its descendants.

```go
type ServiceProvider interface {
    GetService(serviceType reflect.Type) interface{}
}

type App struct {
    Database *DatabaseComponent
    API      *APIComponent
    services map[reflect.Type]interface{}
}

func (a *App) GetService(serviceType reflect.Type) interface{} {
    return a.services[serviceType]
}

func (a *App) PreInit(ctx context.Context) error {
    // Register services this level provides
    a.services = map[reflect.Type]interface{}{
        reflect.TypeOf((*DatabaseComponent)(nil)): a.Database,
    }
    return nil
}

// Child component requests service
func (api *APIComponent) Init(ctx context.Context, parent interface{}) error {
    if provider, ok := parent.(ServiceProvider); ok {
        if db := provider.GetService(reflect.TypeOf((*DatabaseComponent)(nil))); db != nil {
            api.database = db.(*DatabaseComponent)
        }
    }
    return nil
}
```

**Pros:**
- Explicit service contracts
- Hierarchical service resolution
- Components declare what they provide

**Cons:**
- Requires ServiceProvider implementation
- Manual service registration
- Type safety through reflection

## Pattern 4: Dependency Injection Container

Use a DI container pattern where components declare dependencies through tags or interfaces.

```go
type App struct {
    Container *DIContainer
    Database  *DatabaseComponent `provide:"database"`
    Cache     *CacheComponent    `provide:"cache"`
    Service   *ServiceComponent  `inject:"database,cache"`
}

// DIContainer manages dependencies
type DIContainer struct {
    providers map[string]interface{}
    mu        sync.RWMutex
}

// Component with dependencies
type ServiceComponent struct {
    Database *DatabaseComponent `inject:"database"`
    Cache    *CacheComponent    `inject:"cache"`
}

func (s *ServiceComponent) Init(ctx context.Context) error {
    // Dependencies are injected before Init is called
    // by the autoinit framework with DI support
    return nil
}
```

**Pros:**
- Declarative dependencies
- Automatic wiring
- Clear dependency graph

**Cons:**
- Requires framework extension
- More complex than basic autoinit
- Compile-time checking limited

## Pattern 5: Path-Based Component Access

Components can be accessed using a path notation, similar to filesystem paths.

```go
// Component registry with path-based access
type ComponentTree struct {
    root interface{}
    paths map[string]interface{}
}

// Build paths during initialization
func buildComponentPaths(obj interface{}, path string, tree *ComponentTree) {
    tree.paths[path] = obj
    // Recursively build paths for nested components
}

// Access components by path
func (c *Component) Init(ctx context.Context) error {
    if tree := ctx.Value(treeKey).(*ComponentTree); tree != nil {
        // Access parent's database
        if db := tree.Get("../Database"); db != nil {
            c.database = db.(*DatabaseComponent)
        }
        // Access root-level cache
        if cache := tree.Get("/Cache"); cache != nil {
            c.cache = cache.(*CacheComponent)
        }
    }
    return nil
}
```

**Pros:**
- Intuitive path-based access
- Can access any component in tree
- Supports relative and absolute paths

**Cons:**
- Requires path construction
- String-based (no compile-time checking)
- Can break if structure changes

## Pattern 6: Event-Based Communication

Components communicate through events rather than direct dependencies.

```go
type EventBus struct {
    subscribers map[string][]func(interface{})
    mu          sync.RWMutex
}

// Component publishes events
type DatabaseComponent struct {
    eventBus *EventBus
}

func (d *DatabaseComponent) Init(ctx context.Context) error {
    d.eventBus = ctx.Value(eventBusKey).(*EventBus)
    // Publish availability
    d.eventBus.Publish("database.ready", d)
    return nil
}

// Component subscribes to events
type ServiceComponent struct {
    database *DatabaseComponent
}

func (s *ServiceComponent) Init(ctx context.Context) error {
    bus := ctx.Value(eventBusKey).(*EventBus)
    // Subscribe to database ready event
    bus.Subscribe("database.ready", func(data interface{}) {
        s.database = data.(*DatabaseComponent)
    })
    return nil
}
```

**Pros:**
- Completely decoupled
- Dynamic relationships
- Supports many-to-many relationships

**Cons:**
- Asynchronous complexity
- Harder to trace dependencies
- Potential race conditions

## Recommended Approach

For most applications, we recommend a **hybrid approach**:

### 1. Simple Dependencies: Use Context Values

For widely-used services (logger, database, cache), pass them through context:

```go
ctx := context.WithValue(context.Background(), "database", db)
ctx = context.WithValue(ctx, "cache", cache)
autoinit.AutoInit(ctx, app)
```

### 2. Parent-Child Dependencies: Use Parent Parameter

For components that naturally depend on their parent:

```go
func (c *ChildComponent) Init(ctx context.Context, parent interface{}) error {
    if p, ok := parent.(*ParentComponent); ok {
        c.parentConfig = p.Config
    }
    return nil
}
```

### 3. Complex Dependencies: Component Registry

For complex inter-component dependencies, use a lightweight registry:

```go
type App struct {
    Registry *ComponentRegistry
    // Components register themselves and discover others
}
```

## Best Practices

### 1. Minimize Dependencies
Before adding a dependency, consider if the component can be self-contained. Often, passing configuration through context is sufficient.

### 2. Use Interfaces
Define interfaces for dependencies rather than concrete types:

```go
type CacheProvider interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}

type ServiceComponent struct {
    cache CacheProvider // Not *CacheComponent
}
```

### 3. Fail Gracefully
Components should handle missing dependencies gracefully:

```go
func (s *ServiceComponent) Init(ctx context.Context) error {
    // Try to get cache, but don't fail if not available
    if cache := ctx.Value("cache"); cache != nil {
        s.cache = cache.(CacheProvider)
        s.cacheEnabled = true
    }
    return nil
}
```

### 4. Document Dependencies
Clearly document what dependencies a component requires:

```go
// ServiceComponent requires:
// - DatabaseComponent (via context key "database")
// - CacheComponent (optional, via context key "cache")
type ServiceComponent struct {
    // ...
}
```

## Example: Complete Application

Here's a complete example combining several patterns:

```go
package main

import (
    "context"
    "github.com/telnet2/autoinit"
)

// Registry for component discovery
type Registry struct {
    components map[string]interface{}
}

func (r *Registry) Register(name string, component interface{}) {
    r.components[name] = component
}

func (r *Registry) Get(name string) interface{} {
    return r.components[name]
}

// Core components
type DatabaseComponent struct {
    Connected bool
}

func (d *DatabaseComponent) Init(ctx context.Context) error {
    // Register self
    if reg := ctx.Value("registry").(*Registry); reg != nil {
        reg.Register("database", d)
    }
    d.Connected = true
    return nil
}

type CacheComponent struct {
    Ready bool
}

func (c *CacheComponent) Init(ctx context.Context) error {
    // Register self
    if reg := ctx.Value("registry").(*Registry); reg != nil {
        reg.Register("cache", c)
    }
    c.Ready = true
    return nil
}

// Service component with dependencies
type APIComponent struct {
    database *DatabaseComponent
    cache    *CacheComponent
}

func (a *APIComponent) Init(ctx context.Context) error {
    // Discover dependencies
    if reg := ctx.Value("registry").(*Registry); reg != nil {
        if db := reg.Get("database"); db != nil {
            a.database = db.(*DatabaseComponent)
        }
        if cache := reg.Get("cache"); cache != nil {
            a.cache = cache.(*CacheComponent)
        }
    }
    return nil
}

// Application
type App struct {
    Database *DatabaseComponent
    Cache    *CacheComponent
    API      *APIComponent
}

func main() {
    app := &App{
        Database: &DatabaseComponent{},
        Cache:    &CacheComponent{},
        API:      &APIComponent{},
    }
    
    // Create context with registry
    registry := &Registry{
        components: make(map[string]interface{}),
    }
    ctx := context.WithValue(context.Background(), "registry", registry)
    
    // Initialize all components
    if err := autoinit.AutoInit(ctx, app); err != nil {
        panic(err)
    }
    
    // Components are now wired together
    // API component has references to Database and Cache
}
```

## Conclusion

While autoinit's strength is in its simplicity and plug-and-play nature, real applications need ways to handle cross-component dependencies. The patterns presented here provide various approaches, from simple context passing to sophisticated service discovery.

Choose the pattern that best fits your application's complexity:
- **Simple apps**: Context values and parent parameters
- **Medium complexity**: Component registry or service locator
- **Complex systems**: DI container or event-based architecture

Remember that the goal is to maintain the plug-and-play benefits of autoinit while enabling the component interactions your application requires.