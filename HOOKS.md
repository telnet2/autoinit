# Component Hook System

The autoinit package provides hooks that allow components to customize their initialization process and coordinate with parent and child components.

## Available Hook Interfaces

### 1. PreInitializer and PostInitializer

These hooks are called on a component before and after its child components are initialized.

```go
type PreInitializer interface {
    PreInit(ctx context.Context) error
}

type PostInitializer interface {
    PostInit(ctx context.Context) error
}
```

**Execution Order for a component:**
1. `PreInit()` - Called first
2. All child components are recursively initialized
3. `Init()` (or `Init(ctx)` or `Init(ctx, parent)`) - Called after child components
4. `PostInit()` - Called last

### 2. PreFieldHook and PostFieldHook

These hooks allow a parent component to intercept the initialization of its child components.

```go
type PreFieldHook interface {
    PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error
}

type PostFieldHook interface {
    PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error
}
```

**When these are called:**
- `PreFieldInit` - Called on the parent component BEFORE a child component is initialized
- `PostFieldInit` - Called on the parent component AFTER a child component is fully initialized

## Example Usage

### Using PreInit and PostInit

```go
type Service struct {
    Name   string
    Status string
}

func (s *Service) PreInit(ctx context.Context) error {
    s.Status = "initializing"
    // Prepare resources before initialization
    return nil
}

func (s *Service) Init(ctx context.Context) error {
    s.Status = "ready"
    // Main initialization logic
    return nil
}

func (s *Service) PostInit(ctx context.Context) error {
    s.Status = "operational"
    // Final setup after everything is initialized
    return nil
}
```

### Using Field Hooks

```go
type System struct {
    Database Service
    Cache    Service
    
    InitOrder []string
}

func (s *System) PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
    // Called before each field is initialized
    fmt.Printf("About to initialize: %s\n", fieldName)
    s.InitOrder = append(s.InitOrder, "pre-"+fieldName)
    return nil
}

func (s *System) PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error {
    // Called after each field is initialized
    if svc, ok := fieldValue.(*Service); ok {
        fmt.Printf("%s is now %s\n", fieldName, svc.Status)
    }
    s.InitOrder = append(s.InitOrder, "post-"+fieldName)
    return nil
}
```

## Complete Initialization Flow

For a parent component with child components, the complete flow is:

1. Parent component's `PreInit()` (if implemented)
2. For each child component:
   - Parent's `PreFieldInit()` (if implemented)
   - Child component's complete initialization (recursive, including its PreInit, children, Init, PostInit)
   - Parent's `PostFieldInit()` (if implemented)
3. Parent component's `Init()` (if implemented)
4. Parent component's `PostInit()` (if implemented)

## Use Cases

### PreInit and PostInit
- **Resource preparation**: Open connections, allocate resources before component initialization
- **Cleanup**: Release temporary resources after component initialization
- **State tracking**: Track component initialization progress
- **Validation**: Perform final validation after all child components are initialized

### Field Hooks
- **Component injection**: Modify child component configuration before initialization
- **Monitoring**: Track which components are being plugged in and initialized
- **Coordination**: Ensure proper initialization order between components
- **Validation**: Verify child component state after initialization
- **Logging**: Add detailed logging around component initialization

## Notes

- All hooks are optional - implement only what you need
- Hooks can return errors to stop the initialization process
- **Field hooks are called for**:
  - Struct fields (always)
  - Maps containing struct types or pointers to structs (the entire map is passed)
  - Slices/arrays containing struct types or pointers to structs (the entire collection is passed)
  - NOT called for primitive types or collections of primitives
- **The `fieldValue` parameter in field hooks**:
  - For struct fields: Always a pointer (`*StructType`) to allow modification
  - For maps: Pointer to the map (`*map[K]V`) - you receive the entire map
  - For slices: Pointer to the slice (`*[]T`) - you receive the entire slice
  - For arrays: Pointer to the array (`*[N]T`) - you receive the entire array
- **Hook timing for collections**:
  - PreFieldInit is called before any elements are initialized
  - Elements are initialized individually
  - PostFieldInit is called after all elements are initialized
- Hooks are called even for nil pointer fields (you can check for nil in your hook implementation)