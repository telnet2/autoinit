# AutoInit - Component-Based Initialization Framework for Go

A lightweight Go SDK that enables component-based architecture through automatic initialization. Build applications by composing self-contained components that initialize themselves - just plug them into your structs and call AutoInit.

## Component-Based Architecture

AutoInit treats any struct with an `Init` method as a **component** - a self-contained, pluggable unit that knows how to initialize itself. This enables true plug-and-play architecture:

```go
// Just add components to your app - no wiring needed!
type App struct {
    Database *DatabaseComponent  // Self-initializing component
    Cache    *CacheComponent     // Just plug it in
    Auth     *AuthComponent      // Automatically wired
    Metrics  *MetricsComponent   // No manual initialization
}

// One call initializes all components
app := &App{...}
autoinit.AutoInit(ctx, app)  // All components ready!
```

See [COMPONENTS.md](COMPONENTS.md) for detailed component architecture documentation.

## Features

- 🔄 **Recursive Traversal**: Automatically initializes all nested structs at any depth
- 📝 **Declaration Order**: Processes fields in the order they're declared
- 🎯 **Smart Detection**: Automatically detects and calls the appropriate Init method
- 🔍 **Detailed Errors**: Provides complete path context when initialization fails
- 📦 **Collection Support**: Handles slices, arrays, and maps containing structs
- 🏗️ **Flexible**: Works with both pointer and value receivers
- 💡 **Lightweight**: Minimal dependencies (only zerolog for optional logging)
- 🌐 **Context Support**: Optional context.Context propagation through initialization chain
- 👪 **Parent Reference**: Optional parent struct reference for child initialization
- 🎭 **Multiple Interfaces**: Supports three initialization patterns for different use cases
- 📝 **Trace Logging**: Built-in zerolog support for detailed traversal logging at TRACE level
- 🏷️ **Tag Control**: Use struct tags to control which fields are initialized
- 🔐 **Explicit Mode**: RequireTags option for opt-in initialization behavior
- 🔄 **Cycle Detection**: Prevents infinite loops in circular references
- 🪝 **Hook System**: Pre/Post initialization hooks for custom logic

## Installation

```bash
go get github.com/user/autoinit
go get github.com/rs/zerolog  # Required for logging support
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "github.com/user/autoinit"
)

// Simple initialization without context
type Config struct {
    Loaded bool
}

func (c *Config) Init() error {
    c.Loaded = true
    return nil
}

// Context-aware initialization
type Database struct {
    Connected bool
}

func (d *Database) Init(ctx context.Context) error {
    // Can use context for timeouts, cancellation, values
    d.Connected = true
    return nil
}

// Child component that needs parent reference
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

type App struct {
    Name     string
    Config   Config   // Uses Init()
    Database Database // Uses Init(ctx)
    Logger   Logger   // Uses Init(ctx, parent)
}

func main() {
    app := &App{Name: "MyApp"}
    ctx := context.Background()
    
    // Initialize all components automatically
    if err := autoinit.AutoInit(ctx, app); err != nil {
        log.Fatal(err)
    }
    
    // All components are initialized with their appropriate methods
}
```

## How It Works

1. **Depth-First Traversal**: The SDK traverses the struct tree depth-first, initializing children before parents
2. **Interface Detection**: Checks each struct for any of the three Init method signatures
3. **Priority Selection**: If a struct implements multiple interfaces, uses the highest priority:
   - `Init(context.Context, interface{})` - Highest priority
   - `Init(context.Context)` - Medium priority  
   - `Init()` - Lowest priority
4. **Automatic Pointer Conversion**: Converts value types to pointers when calling Init() to allow modifications
5. **Error Propagation**: Stops on first error and provides complete path context
6. **Context & Parent Propagation**: Passes context and parent references through the initialization chain

### Initialization Order

```go
type Parent struct {
    Child1 Child  // Initialized first
    Child2 Child  // Initialized second
    // Parent.Init() called last
}
```

## Advanced Usage

### Complex Structures

The SDK handles complex nested structures including:

```go
type System struct {
    Services []Service              // Slice of structs
    Plugins  map[string]*Plugin     // Maps with struct values
    Cache    *CacheLayer            // Pointer fields
    Config   EmbeddedConfig         // Embedded structs
}
```

### Error Handling

When initialization fails, you get detailed context:

```go
err := autoinit.AutoInit(app)
if err != nil {
    // Error: failed to initialize field 'Services.[2].Database.ConnectionPool' 
    //        of type *Pool: connection refused
    
    if initErr, ok := err.(*autoinit.InitError); ok {
        path := initErr.GetPath()      // ["Services", "[2]", "Database", "ConnectionPool"]
        fieldType := initErr.GetFieldType() // "*Pool"
        cause := initErr.Unwrap()       // Original error
    }
}
```

### Skip Fields

Fields without any Init method are automatically skipped:

```go
type Mixed struct {
    NeedsInit    Component  // Has Init() - will be initialized
    PlainStruct  Config     // No Init() - skipped
    SimpleField  string     // Not a struct - skipped
}
```

### Tag-Based Control

Use struct tags to control which fields are initialized:

```go
type App struct {
    // Always initialize these
    Database *DB     `autoinit:"init"`
    Cache    *Cache  `autoinit:""`      // Empty tag also means init
    
    // Never initialize this
    Logger   *Logger `autoinit:"-"`      // Explicitly skip
    
    // Conditional based on RequireTags option
    Service  *Service                    // No tag - depends on RequireTags
}

// Opt-in mode: only initialize tagged fields
options := &autoinit.Options{
    RequireTags: true,  // Only fields with autoinit tags will be initialized
}
err := autoinit.AutoInitWithOptions(ctx, &app, options)
```

See [TAGS.md](TAGS.md) for detailed tag documentation.

## Examples

Run the examples with:
```bash
go test -run Example
```

Check the example test files:
- `example_simple_test.go` - Basic usage
- `example_nested_test.go` - Deeply nested structures
- `example_complex_test.go` - Slices, maps, and error handling

## Logging

The SDK includes comprehensive trace logging using zerolog:

### Default Logging

```go
// Uses default logger to stdout with TRACE level
ctx := context.Background()
err := autoinit.AutoInit(ctx, &app)
```

Default logger output:
```json
{"level":"trace","target_type":"*main.App","time":"2024-01-01T10:00:00Z","message":"Starting AutoInit"}
{"level":"trace","path":"Database","type":"main.Database","message":"Processing struct"}
{"level":"trace","path":"Database","method":"Init(ctx)","message":"Calling initializer"}
{"level":"trace","message":"AutoInit completed successfully"}
```

### Custom Logger

```go
import "github.com/rs/zerolog"

// Create custom logger
logger := zerolog.New(os.Stderr).With().
    Timestamp().
    Str("service", "my-app").
    Logger().
    Level(zerolog.TraceLevel)

options := &autoinit.Options{
    Logger: &logger,
}

err := autoinit.AutoInitWithOptions(ctx, &app, options)
```

### Log Levels

- **TRACE**: Full traversal details (field visits, method calls, completions)
- **ERROR**: Only initialization failures
- Higher levels suppress trace output

### Log Information

Trace logs include:
- Struct types and field names
- Full path to each field (e.g., `App.Services.[0].Database`)
- Which Init method variant was called
- Success/failure status
- Skipped fields (nil pointers, unexported, non-structs)

## Use Cases

Perfect for:

- **Dependency Injection**: Initialize service dependencies automatically
- **Application Bootstrap**: Reduce boilerplate in complex applications
- **Plugin Systems**: Initialize plugins without manual wiring
- **Microservices**: Initialize service components in correct order
- **Testing**: Auto-initialize test fixtures

## Performance

The SDK uses reflection which has a performance cost. However, since initialization typically happens once at startup, this overhead is negligible for most applications.

## API Reference

### Main Functions

```go
func AutoInit(ctx context.Context, target interface{}) error
```

Recursively initializes all fields with default logging to stdout.

```go
func AutoInitWithOptions(ctx context.Context, target interface{}, options *Options) error
```

Recursively initializes all fields with custom options.

**Parameters:**
- `ctx`: Context for cancellation, timeout, and value propagation
- `target`: Pointer to the struct to initialize
- `options`: Optional configuration including custom logger

**Returns:**
- `error`: nil on success, or detailed error with path context on failure

### Options

```go
type Options struct {
    // Logger for trace logging during traversal
    // If nil, uses default stdout logger with trace level
    Logger *zerolog.Logger
}
```

### Initializer Interfaces

The SDK supports three initialization interfaces:

```go
// SimpleInitializer - Basic initialization without context
type SimpleInitializer interface {
    Init() error
}

// ContextInitializer - Context-aware initialization
type ContextInitializer interface {
    Init(ctx context.Context) error
}

// ParentInitializer - Initialization with context and parent reference
type ParentInitializer interface {
    Init(ctx context.Context, parent interface{}) error
}
```

**Interface Selection**:
- Structs can implement any one of these interfaces
- If multiple are implemented, the highest priority interface is used
- Parent reference is `nil` for the root struct, otherwise references the containing struct

**Use Cases**:
- **SimpleInitializer**: Basic setup, no external dependencies
- **ContextInitializer**: Timeout control, cancellation, request-scoped values
- **ParentInitializer**: Access parent configuration, dependency injection, hierarchical setup

### Error Types

```go
type InitError struct {
    Path      []string  // Path to the failing field
    FieldType string    // Type of the field
    Cause     error     // Original error from Init()
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details