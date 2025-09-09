# AutoInit Best Practices Guide

## 🎯 Quick Start Checklist

Before diving into best practices, ensure you have:
- [ ] Go 1.19+ installed
- [ ] `go get github.com/telnet2/autoinit` completed
- [ ] Basic understanding of struct composition

## 📋 Component Design Best Practices

### 1. Component Structure

**✅ DO: Keep components focused and single-purpose**
```go
// Good: Single responsibility
type Database struct {
    Config *DatabaseConfig
    Pool   *sql.DB
}

func (d *Database) Init(ctx context.Context) error {
    // Database-specific initialization only
    return nil
}

// Bad: Multiple responsibilities
type DatabaseWithCache struct {
    Config *DatabaseConfig
    Pool   *sql.DB
    Cache  *Cache
    Logger *Logger
    // ... too many concerns
}
```

**✅ DO: Use interfaces for dependencies**
```go
type Logger interface {
    Info(msg string, fields ...interface{})
    Error(msg string, err error, fields ...interface{})
}

type Service struct {
    Logger Logger // Interface, not concrete type
    DB     *Database
}
```

### 2. Initialization Patterns

**✅ DO: Choose the right initialization pattern**

| Pattern | Use Case | Example |
|---------|----------|---------|
| `Init()` | Simple, no dependencies | Config loading |
| `Init(ctx)` | Context-aware (timeouts) | Network connections |
| `Init(ctx, parent)` | Needs parent access | Service discovery |

```go
// Context-aware initialization
func (d *Database) Init(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return d.connect(ctx)
}

// Parent-aware initialization  
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    if app, ok := parent.(*App); ok {
        s.AppName = app.Name
    }
    return nil
}
```

## 🔍 Dependency Discovery Patterns

### 1. Modern As Pattern (Recommended)

**✅ DO: Use the As pattern for type-safe discovery**
```go
type Service struct {
    db     *Database
    cache  Cache
    logger Logger
}

func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Type-safe discovery
    autoinit.MustAs(ctx, s, parent, &s.db)     // Required
    
    // Optional discovery with filtering
    var primaryDB *Database
    if autoinit.As(ctx, s, parent, &primaryDB, 
        autoinit.WithFieldName("PrimaryDB"),
        autoinit.WithJSONTag("primary")) {
        s.db = primaryDB
    }
    
    // Interface discovery
    autoinit.MustAs(ctx, s, parent, &s.logger)
    
    return nil
}
```

**❌ DON'T: Use reflection directly**
```go
// Bad: Manual reflection
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    v := reflect.ValueOf(parent).Elem()
    dbField := v.FieldByName("Database")
    if !dbField.IsValid() {
        return errors.New("database not found")
    }
    s.db = dbField.Interface().(*Database)
    return nil
}
```

### 2. Classic Finder Pattern (Legacy)

**✅ DO: Use when you need complex search logic**
```go
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    ctx = autoinit.WithComponentSearch(ctx)
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find by type
    s.db = autoinit.FindByType[*Database](ctx, s, parent)
    
    // Find by tag
    if cache := finder.Find(autoinit.SearchOption{
        ByJSONTag: "cache",
    }); cache != nil {
        s.cache = cache.(*RedisCache)
    }
    
    return nil
}
```

## 🏷️ Tag-Based Control Best Practices

### 1. Initialization Control

**✅ DO: Use tags for explicit control**
```go
type App struct {
    // Always initialize
    Database *Database `autoinit:"init"`
    Cache    *Cache    `autoinit:""`
    
    // Never initialize
    Debug    *Debug    `autoinit:"-"`
    
    // Conditional (default behavior)
    Optional *Service  // No tag
}
```

**✅ DO: Use RequireTags for security**
```go
// Opt-in mode: only tagged fields initialize
options := &autoinit.Options{
    RequireTags: true,  // Explicit initialization required
}
err := autoinit.AutoInitWithOptions(ctx, app, options)
```

### 2. JSON Tag Integration

**✅ DO: Use JSON tags for configuration**
```go
type AppConfig struct {
    Environment string         `yaml:"environment"`
    Database    DatabaseConfig `yaml:"database"`
    Features    map[string]bool `yaml:"features"`
}

type App struct {
    Config AppConfig `yaml:",inline"`
    
    // Components discover config via JSON tags
    Database *Database `json:"primary_db"`
    Cache    *Cache    `json:"cache"`
}
```

## 🪝 Lifecycle Hooks Best Practices

### 1. Hook Implementation

**✅ DO: Implement all relevant hooks for complex components**
```go
type Database struct {
    Config *DatabaseConfig
    Pool   *sql.DB
    Status string
}

// Called before child components initialize
func (d *Database) PreInit(ctx context.Context) error {
    d.Status = "initializing"
    log.Printf("Starting database initialization")
    return nil
}

// Main initialization
func (d *Database) Init(ctx context.Context) error {
    pool, err := sql.Open("postgres", d.Config.DSN)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    
    d.Pool = pool
    d.Status = "ready"
    return nil
}

// Called after all children initialized
func (d *Database) PostInit(ctx context.Context) error {
    d.Status = "operational"
    log.Printf("Database operational")
    return nil
}
```

### 2. Parent Hooks

**✅ DO: Use parent hooks for coordination**
```go
type App struct {
    Database *Database
    Cache    *Cache
    Services []*Service
}

func (a *App) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
    log.Printf("Initializing: %s (%T)", fieldName, fieldValue)
    
    // Custom initialization order
    if fieldName == "Database" {
        // Ensure database is ready before other components
        return nil
    }
    
    return nil
}
```

## 🧪 Testing Best Practices

### 1. Component Testing

**✅ DO: Test components in isolation**
```go
func TestDatabase(t *testing.T) {
    db := &Database{
        Config: &DatabaseConfig{DSN: "test-db"},
    }
    
    ctx := context.Background()
    err := db.Init(ctx)
    
    require.NoError(t, err)
    assert.True(t, db.Connected)
}
```

**✅ DO: Use table-driven tests for complex scenarios**
```go
func TestServiceInitialization(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() *App
        wantErr bool
    }{
        {
            name: "valid configuration",
            setup: func() *App {
                return &App{
                    Database: &Database{Config: validConfig},
                    Cache:    &Cache{},
                }
            },
            wantErr: false,
        },
        {
            name: "invalid database config",
            setup: func() *App {
                return &App{
                    Database: &Database{Config: invalidConfig},
                }
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            app := tt.setup()
            err := autoinit.AutoInit(context.Background(), app)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Mock Testing

**✅ DO: Use interfaces for easy mocking**
```go
type Service struct {
    DB     DatabaseInterface
    Logger LoggerInterface
}

// In tests
type mockDB struct{}
func (m *mockDB) Query(query string) error { return nil }

func TestServiceWithMock(t *testing.T) {
    service := &Service{
        DB:     &mockDB{},
        Logger: &mockLogger{},
    }
    
    err := autoinit.AutoInit(context.Background(), service)
    assert.NoError(t, err)
}
```

## 🏗️ Architecture Patterns

### 1. Layered Architecture

**✅ DO: Organize components in layers**
```go
// Infrastructure layer
type Database struct{ /* ... */ }
type Cache struct{ /* ... */ }
type Logger struct{ /* ... */ }

// Service layer
type UserService struct {
    DB     *Database
    Cache  *Cache
    Logger Logger
}

// Application layer
type App struct {
    Database    *Database
    Cache       *Cache
    Logger      *Logger
    UserService *UserService
}
```

### 2. Plugin Architecture

**✅ DO: Use conditional initialization for plugins**
```go
type App struct {
    Core *CoreEngine
    
    // Optional plugins
    AuthPlugin      *AuthPlugin      `autoinit:"-"`
    MetricsPlugin   *MetricsPlugin  `autoinit:"-"`
    CachePlugin     *CachePlugin    `autoinit:"-"`
}

func main() {
    app := &App{Core: &CoreEngine{}}
    
    // Conditionally add plugins
    if os.Getenv("ENABLE_AUTH") == "true" {
        app.AuthPlugin = &AuthPlugin{}
    }
    
    if os.Getenv("ENABLE_METRICS") == "true" {
        app.MetricsPlugin = &MetricsPlugin{}
    }
    
    autoinit.AutoInit(context.Background(), app)
}
```

## 🔄 Configuration Management

### 1. YAML-Driven Configuration

**✅ DO: Use YAML for external configuration**
```go
type AppConfig struct {
    Environment string         `yaml:"environment"`
    Database    DatabaseConfig `yaml:"database"`
    Features    map[string]bool `yaml:"features"`
}

type App struct {
    Config AppConfig `yaml:",inline"`
    
    Database *Database
    Cache    *Cache
    Services []*Service
}

func main() {
    app := &App{}
    
    // Load configuration
    yamlData, err := os.ReadFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    if err := yaml.Unmarshal(yamlData, app); err != nil {
        log.Fatal(err)
    }
    
    // Initialize with configuration
    if err := autoinit.AutoInit(context.Background(), app); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Environment-Specific Configuration

**✅ DO: Use environment variables for secrets**
```go
type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"-"` // Skip YAML
}

type Database struct {
    Config *DatabaseConfig
}

func (d *Database) Init(ctx context.Context) error {
    // Load sensitive data from environment
    d.Config.Password = os.Getenv("DB_PASSWORD")
    
    return d.connect()
}
```

## 🚨 Common Pitfalls and How to Avoid Them

### 1. Circular Dependencies

**❌ DON'T: Create circular references**
```go
// Bad: Circular dependency
// A depends on B, B depends on A
type A struct { B *B }
type B struct { A *A }
```

**✅ DO: Use interfaces to break cycles**
```go
type AInterface interface{ DoA() }
type BInterface interface{ DoB() }

type A struct { B BInterface }
type B struct { A AInterface }
```

### 2. Nil Pointer Issues

**❌ DON'T: Assume fields are initialized**
```go
// Bad: Will panic if Database is nil
func (s *Service) Init(ctx context.Context) error {
    return s.Database.Ping() // Panic if nil
}
```

**✅ DO: Use discovery patterns**
```go
// Good: Safe discovery
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    autoinit.MustAs(ctx, s, parent, &s.Database) // Panics with clear message
    return s.Database.Ping()
}
```

### 3. Initialization Order

**❌ DON'T: Rely on field order**
```go
// Bad: Assumes Database is initialized first
type App struct {
    Database *Database
    Service  *Service // Depends on Database
}
```

**✅ DO: Use dependency discovery**
```go
// Good: Service discovers Database
type Service struct {
    DB *Database
}

func (s *Service) Init(ctx context.Context, parent interface{}) error {
    autoinit.MustAs(ctx, s, parent, &s.DB)
    return nil
}
```

## 📊 Performance Optimization

### 1. Minimize Reflection Usage

**✅ DO: Cache reflection results**
```go
type Service struct {
    db *Database
}

// Cache discovered components
var dbType = reflect.TypeOf(&Database{})

func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Use cached type instead of reflection each time
    return nil
}
```

### 2. Use Value Types When Possible

**✅ DO: Prefer value types for simple components**
```go
// Good: Value type for simple config
type Config struct {
    Port int
    Host string
}

func (c Config) Init(ctx context.Context) error {
    // Value receiver for simple types
    return nil
}
```

## 🏗️ Container Pattern for Dependency Organization

The container pattern is a powerful way to organize related dependencies into logical groups, providing better organization and maintainability for complex applications.

### Why Use Containers?

**Benefits of Container Pattern**:
- **Logical Organization**: Related components grouped together (DAOs, handlers, services)
- **Simplified Injection**: Inject entire containers instead of individual components
- **Better Encapsulation**: Container boundaries define clear architectural layers
- **Easier Testing**: Mock entire containers for unit testing
- **Reduced Coupling**: Components depend on containers, not individual implementations

### Container Types and Examples

#### 1. DAOContainer - Data Access Layer

Groups all database access objects:

```go
// dao_container.go
type DAOContainer struct {
    // Entity management
    EntityDAO    *EntityDAO    `yaml:"entity_dao"`
    SchemaDAO    *SchemaDAO    `yaml:"schema_dao"`
    
    // User management
    UserDAO      *UserDAO      `yaml:"user_dao"`
    UserGroupDAO *UserGroupDAO `yaml:"user_group_dao"`
    
    // Content management
    GroupDAO     *GroupDAO     `yaml:"group_dao"`
    FieldDAO     *FieldDAO     `yaml:"field_dao"`
    
    // System
    ConfigDAO    *ConfigDAO    `yaml:"config_dao"`
}

// Usage in handlers
func (h *EntityHandler) Init(ctx context.Context, parent interface{}) error {
    // Discover the entire DAO container
    autoinit.MustAs(ctx, h, parent, &h.daos)
    return nil
}

func (h *EntityHandler) CreateEntity(ctx context.Context, req *CreateEntityRequest) error {
    // Access DAOs through container
    entity, err := h.daos.EntityDAO.Create(ctx, req.Entity)
    if err != nil {
        return err
    }
    
    // Use related DAOs as needed
    schema, err := h.daos.SchemaDAO.GetByID(ctx, req.SchemaID)
    // ...
}
```

#### 2. HandlerContainer - Business Logic Layer

Groups all HTTP request handlers:

```go
// handler_container.go
type HandlerContainer struct {
    // Core entity handlers
    EntityHandler *EntityHandler `yaml:"entity_handler"`
    SchemaHandler *SchemaHandler `yaml:"schema_handler"`
    
    // User management handlers
    UserHandler   *UserHandler   `yaml:"user_handler"`
    GroupHandler  *GroupHandler  `yaml:"group_handler"`
    
    // System handlers
    ConfigHandler *ConfigHandler `yaml:"config_handler"`
    HealthHandler *HealthHandler `yaml:"health_handler"`
}

// Usage in server
func (s *Server) Init(ctx context.Context, parent interface{}) error {
    // Discover the entire handler container
    autoinit.MustAs(ctx, s, parent, &s.handlers)
    return nil
}

func (s *Server) mountRoutes() {
    // Access handlers through container
    s.engine.GET("/entities", s.handlers.EntityHandler.List)
    s.engine.POST("/entities", s.handlers.EntityHandler.Create)
    s.engine.GET("/users", s.handlers.UserHandler.List)
    s.engine.POST("/groups", s.handlers.GroupHandler.Create)
}
```

#### 3. ServiceContainer - Infrastructure Layer

Groups external service integrations:

```go
// service_container.go
type ServiceContainer struct {
    // External services
    EmailService    *EmailService    `yaml:"email_service"`
    StorageService  *StorageService  `yaml:"storage_service"`
    CacheService    *CacheService    `yaml:"cache_service"`
    
    // Monitoring
    MetricsService  *MetricsService  `yaml:"metrics_service"`
    LoggingService  *LoggingService  `yaml:"logging_service"`
}
```

### Container Grouping Strategies

#### 1. By Architectural Layer

**Data Access Layer (DAOs)**:
- All database interactions
- Repository pattern implementations
- Data validation and transformation

**Business Logic Layer (Handlers/Services)**:
- HTTP request processing
- Business rule enforcement
- API endpoint implementations

**Infrastructure Layer (External Services)**:
- External service integrations
- Cross-cutting concerns
- System utilities

#### 2. By Domain/Feature

**User Management Domain**:
```go
type UserDomainContainer struct {
    // DAOs
    UserDAO      *UserDAO      `yaml:"user_dao"`
    UserGroupDAO *UserGroupDAO `yaml:"user_group_dao"`
    
    // Handlers
    UserHandler  *UserHandler  `yaml:"user_handler"`
    GroupHandler *GroupHandler `yaml:"group_handler"`
    
    // Services
    UserService  *UserService  `yaml:"user_service"`
}
```

**Content Management Domain**:
```go
type ContentDomainContainer struct {
    // DAOs
    EntityDAO *EntityDAO `yaml:"entity_dao"`
    SchemaDAO *SchemaDAO `yaml:"schema_dao"`
    FieldDAO *FieldDAO `yaml:"field_dao"`
    
    // Handlers
    EntityHandler *EntityHandler `yaml:"entity_handler"`
    SchemaHandler *SchemaHandler `yaml:"schema_handler"`
    
    // Services
    ContentService *ContentService `yaml:"content_service"`
}
```

#### 3. By Responsibility

**Read Operations**:
```go
type ReadContainer struct {
    EntityReader *EntityReader `yaml:"entity_reader"`
    SchemaReader *SchemaReader `yaml:"schema_reader"`
    UserReader   *UserReader   `yaml:"user_reader"`
}
```

**Write Operations**:
```go
type WriteContainer struct {
    EntityWriter *EntityWriter `yaml:"entity_writer"`
    SchemaWriter *SchemaWriter `yaml:"schema_writer"`
    UserWriter   *UserWriter   `yaml:"user_writer"`
}
```

### Container Implementation Best Practices

#### 1. Container Structure Guidelines

```go
type MyContainer struct {
    // Group related components with clear naming
    ComponentA *ComponentA `yaml:"component_a"`
    ComponentB *ComponentB `yaml:"component_b"`
    ComponentC *ComponentC `yaml:"component_c"`
}

// Optional: Implement initialization if container needs setup
func (c *MyContainer) Init(ctx context.Context) error {
    // Container-level initialization (rarely needed)
    return nil
}
```

#### 2. Main Application Structure

```go
type App struct {
    // Infrastructure (initialized first)
    Database *Database `component:"true" yaml:"database"`
    Cache    *Cache    `component:"true" yaml:"cache"`
    
    // Containers (initialized after infrastructure)
    DAOs      *DAOContainer      `component:"true" yaml:"dao"`
    Handlers  *HandlerContainer  `component:"true" yaml:"handler"`
    Services  *ServiceContainer  `component:"true" yaml:"service"`
    
    // Server (initialized last, uses all containers)
    Server *Server `component:"true" yaml:"server"`
}
```

#### 3. Conditional Container Loading

```go
// Enable/disable containers based on configuration
type App struct {
    // Core always present
    Database *Database `component:"true" yaml:"database"`
    
    // Optional containers
    AuthContainer    *AuthContainer    `component:"true" yaml:"auth"`
    MetricsContainer *MetricsContainer `component:"true" yaml:"metrics"`
    
    // Feature flags
    Config *AppConfig `yaml:",inline"`
}

func main() {
    app := &App{}
    
    // Load configuration
    if err := yaml.Unmarshal(configData, app); err != nil {
        log.Fatal(err)
    }
    
    // Conditionally enable containers
    if app.Config.Features.Auth {
        app.AuthContainer = &AuthContainer{}
    }
    
    if app.Config.Features.Metrics {
        app.MetricsContainer = &MetricsContainer{}
    }
    
    autoinit.AutoInit(context.Background(), app)
}
```

### Testing with Containers

#### 1. Mocking Entire Containers

```go
// Mock container for testing
type MockDAOContainer struct {
    EntityDAO *MockEntityDAO
    UserDAO   *MockUserDAO
}

func TestHandlerWithMockContainer(t *testing.T) {
    handler := &EntityHandler{
        daos: &MockDAOContainer{
            EntityDAO: &MockEntityDAO{},
            UserDAO:   &MockUserDAO{},
        },
    }
    
    // Test with mocked dependencies
    result := handler.CreateEntity(context.Background(), testRequest)
    assert.NoError(t, result)
}
```

#### 2. Partial Container Mocking

```go
func TestServiceWithPartialMock(t *testing.T) {
    // Real DAO container with one mocked component
    daos := &DAOContainer{
        EntityDAO: &RealEntityDAO{},
        UserDAO:   &MockUserDAO{}, // Mock only UserDAO
    }
    
    service := &UserService{daos: daos}
    // Test service logic
}
```

### Container Anti-Patterns to Avoid

#### ❌ Don't: Over-Granular Containers
```go
// Bad: Too many small containers
type EntityDAOContainer struct { EntityDAO *EntityDAO }
type UserDAOContainer struct { UserDAO *UserDAO }
type SchemaDAOContainer struct { SchemaDAO *SchemaDAO }
```

#### ✅ Do: Logical Grouping
```go
// Good: Logical grouping by layer/domain
type DAOContainer struct {
    EntityDAO *EntityDAO
    UserDAO   *UserDAO
    SchemaDAO *SchemaDAO
}
```

#### ❌ Don't: Circular Container Dependencies
```go
// Bad: Circular dependency
// ContainerA depends on ContainerB, ContainerB depends on ContainerA
type ContainerA struct { B *ContainerB }
type ContainerB struct { A *ContainerA }
```

#### ✅ Do: Clear Dependency Direction
```go
// Good: Clear dependency direction
type InfrastructureContainer struct {
    Database *Database
    Cache    *Cache
}

type ServiceContainer struct {
    Infra *InfrastructureContainer
    // Services depend on infrastructure, not vice versa
}
```

## 🔍 Debugging and Troubleshooting

**✅ DO: Add logging for complex initialization**
```go
func (c *ComplexComponent) Init(ctx context.Context) error {
    log.Printf("Initializing %T with config: %+v", c, c.Config)
    
    // Initialization logic
    
    log.Printf("%T initialized successfully", c)
    return nil
}
```

### 2. Use Error Context

**✅ DO: Provide detailed error messages**
```go
func (d *Database) Init(ctx context.Context) error {
    if d.Config.DSN == "" {
        return fmt.Errorf("database DSN is required for %T", d)
    }
    
    pool, err := sql.Open("postgres", d.Config.DSN)
    if err != nil {
        return fmt.Errorf("failed to open database connection: %w", err)
    }
    
    d.Pool = pool
    return nil
}
```

## 🚀 Production Deployment Checklist

### Pre-deployment
- [ ] All components have proper error handling
- [ ] Configuration validated (YAML parsing)
- [ ] Environment variables set
- [ ] Health checks implemented
- [ ] Graceful shutdown implemented

### Monitoring
- [ ] Initialization metrics collected
- [ ] Error tracking configured
- [ ] Performance monitoring enabled
- [ ] Dependency health checks

### Security
- [ ] No hardcoded secrets
- [ ] Environment variables for sensitive data
- [ ] Configuration validation
- [ ] Input sanitization

## 📚 Additional Resources

- [Component Architecture Guide](COMPONENTS.md)
- [Framework Comparison](COMPARISON.md)
- [Component Discovery](FINDER.md)
- [Lifecycle Hooks](HOOKS.md)
- [Tag-based Control](TAGS.md)

---

**Remember**: AutoInit is designed to be simple and declarative. When in doubt, prefer struct composition over complex initialization logic.