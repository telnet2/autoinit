# Tag-Based Component Control

The autoinit package supports struct tags to control which components are plugged into your system and initialized.

## Basic Tag Usage

### Always Skip a Component

Use `autoinit:"-"` to explicitly exclude a component from being plugged into your system:

```go
type MyStruct struct {
    Database *DB                  // Will be initialized (default)
    Logger   *Logger `autoinit:"-"` // Will NOT be initialized
}
```

### Explicitly Include a Component

Use any non-dash tag value to explicitly mark a component for inclusion in your system:

```go
type MyStruct struct {
    Database *DB     `autoinit:"init"` // Explicitly marked for init
    Cache    *Cache  `autoinit:""`     // Empty tag also means init
    Service  *Service                   // Behavior depends on RequireTags option
}
```

## RequireTags Option

The `RequireTags` option changes the default behavior for components without tags.

### Default Behavior (RequireTags = false)

When `RequireTags` is false (the default), all components are plugged in and initialized unless explicitly excluded:

```go
type MyStruct struct {
    Database *DB                  // ✅ Will be initialized
    Cache    *Cache `autoinit:""`  // ✅ Will be initialized
    Logger   *Logger `autoinit:"-"` // ❌ Will NOT be initialized
}

// Usage
err := autoinit.AutoInit(ctx, &myStruct)
```

### Opt-In Mode (RequireTags = true)

When `RequireTags` is true, only components with explicit autoinit tags are plugged into the system:

```go
type MyStruct struct {
    Database *DB     `autoinit:"init"` // ✅ Will be initialized
    Cache    *Cache  `autoinit:""`     // ✅ Will be initialized  
    Service  *Service                   // ❌ Will NOT be initialized (no tag)
    Logger   *Logger `autoinit:"-"`     // ❌ Will NOT be initialized
}

// Usage
options := &autoinit.Options{
    RequireTags: true,
}
err := autoinit.AutoInitWithOptions(ctx, &myStruct, options)
```

## Use Cases

### 1. Selective Component Loading

When you have many available components but only want to plug in specific ones:

```go
type Application struct {
    // Essential components to plug in
    Database    *DatabaseComponent    `autoinit:"init"`
    Cache       *CacheComponent      `autoinit:"init"`
    MessageBus  *MessageBusComponent `autoinit:"init"`
    
    // Optional components - not plugged in with RequireTags=true
    Metrics     *MetricsComponent    // No tag - not plugged in
    Analytics   *AnalyticsComponent
    Monitoring  *MonitorComponent
    
    // Explicitly excluded component
    TestHelper  *TestComponent  `autoinit:"-"`
}

options := &autoinit.Options{
    RequireTags: true,  // Only plug in tagged components
}
err := autoinit.AutoInitWithOptions(ctx, &app, options)
```

### 2. Selective Collection Initialization

Control which slices and maps get processed:

```go
type ServiceManager struct {
    // Initialize these services
    CoreServices    []*Service `autoinit:"init"`
    CriticalBackups []*Service `autoinit:"init"`
    
    // Skip these - managed separately
    OptionalPlugins []*Plugin `autoinit:"-"`
    DebugTools     []*Tool   `autoinit:"-"`
}
```

### 3. Development vs Production Components

Use tags to control which components are plugged in based on environment:

```go
type Application struct {
    // Core components - always plugged in
    Database *DatabaseComponent `autoinit:"init"`
    Cache    *CacheComponent   `autoinit:"init"`
    
    // Development components - only in dev mode
    DevTools   *DevToolsComponent
    Profiler   *ProfilerComponent
    DebugPanel *DebugComponent
}

// Production: only plug in tagged components
prodOptions := &autoinit.Options{
    RequireTags: true,  // Dev components not plugged in
}

// Development: plug in all components
devOptions := &autoinit.Options{
    RequireTags: false,  // All components plugged in (except "-" tagged)
}
```

## Important Notes

### Tag Scope

When `RequireTags` is enabled, it applies throughout the entire tree:

```go
type Parent struct {
    Child ChildStruct `autoinit:"init"`  // Child will be initialized
}

type ChildStruct struct {
    Service *Service  // This will NOT be initialized if RequireTags=true
                     // (no tag on this field)
}
```

### Collections

Tags apply to the entire collection, not individual elements:

```go
type MyStruct struct {
    Services []*Service `autoinit:"init"`  // All services in slice will be initialized
    Configs  map[string]*Config `autoinit:"-"`  // No configs will be initialized
}
```

### Interaction with Hooks

Tags are checked before hooks are called. If a field is skipped due to tags, its hooks won't be invoked:

```go
type Parent struct {
    Child *Child `autoinit:"-"`  // PreFieldInit/PostFieldInit won't be called
}
```

## Migration Guide

If you have existing code and want to adopt `RequireTags`:

1. **Audit your structs**: Identify which fields actually need initialization
2. **Add tags progressively**: Start by adding `autoinit:"-"` to fields you want to skip
3. **Test with RequireTags**: Enable `RequireTags` and add `autoinit:"init"` tags where needed
4. **Clean up**: Remove unnecessary dependencies from your initialization chain

## Best Practices

1. **Be explicit**: When using `RequireTags`, always use clear tag values like `"init"` rather than empty strings
2. **Document intent**: Comment why certain fields are included or excluded
3. **Group related fields**: Keep tagged and untagged fields grouped for clarity
4. **Test both modes**: Ensure your structs work correctly with and without `RequireTags`