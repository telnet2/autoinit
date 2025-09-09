# Component Finder - Sibling and Ancestor Search

The autoinit package includes a powerful component discovery system that allows components to find their dependencies among siblings and ancestors without explicit wiring.

## Overview

The Component Finder enables components to discover other components by searching:
1. **Siblings** - Components at the same level in the struct hierarchy
2. **Parent's siblings** - Components at the parent's level (aunts/uncles)
3. **Ancestors** - Components up the parent chain

## Search Order

When using `Find()`, the search follows this order:
1. Search among siblings at the current level
2. If not found, move up to parent and search its siblings
3. Continue up the hierarchy until found or root reached

This ensures that local components override global ones naturally.

## Usage

### Enable Component Search

```go
import "github.com/user/autoinit"

// Enable component search in context
ctx := autoinit.WithComponentSearch(context.Background())

// Initialize with search-enabled context
err := autoinit.AutoInit(ctx, app)
```

### Basic Component Discovery

```go
type ServiceComponent struct {
    cache  *CacheComponent
    logger *LoggerComponent
}

func (s *ServiceComponent) Init(ctx context.Context, parent interface{}) error {
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find by type
    if cache := finder.Find(autoinit.SearchOption{
        ByType: reflect.TypeOf((*CacheComponent)(nil)),
    }); cache != nil {
        s.cache = cache.(*CacheComponent)
    }
    
    return nil
}
```

## Search Options

### 1. Search by Type

Find a component that matches a specific type:

```go
finder.Find(autoinit.SearchOption{
    ByType: reflect.TypeOf((*CacheComponent)(nil)),
})
```

### 2. Search by Field Name

Find a component by its field name in the parent struct:

```go
finder.Find(autoinit.SearchOption{
    ByFieldName: "PrimaryCache",
})
```

### 3. Search by JSON Tag

Find a component by its JSON tag value:

```go
type App struct {
    Cache *CacheComponent `json:"cache"`
}

// Find by JSON tag
finder.Find(autoinit.SearchOption{
    ByJSONTag: "cache",
})
```

### 4. Search by Custom Tag

Find a component by any custom struct tag:

```go
type App struct {
    MainCache     *CacheComponent `component:"main"`
    FallbackCache *CacheComponent `component:"fallback"`
}

// Find by custom tag
finder.Find(autoinit.SearchOption{
    ByCustomTag: "main",
    TagKey:      "component",
})
```

## Search Methods

### Find() - Full Hierarchical Search

Searches siblings first, then moves up the hierarchy:

```go
cache := finder.Find(searchOption)
```

### FindSibling() - Siblings Only

Searches only among siblings at the same level:

```go
cache := finder.FindSibling(searchOption)
```

### FindAncestor() - Ancestors Only

Searches only up the parent chain:

```go
parentCache := finder.FindAncestor(searchOption)
```

## Helper Functions

### Type-Safe Generic Helper

```go
// Find component by type with type safety
logger := autoinit.FindByType[*LoggerComponent](ctx, self, parent)
```

### Find by Interface

```go
// Find component that implements an interface
cache := autoinit.FindByInterface[CacheInterface](ctx, self, parent)
```

### Find by Name

```go
// Find component by field name
component := autoinit.FindByName(ctx, self, parent, "Cache")
```

### Find by Tag

```go
// Find component by JSON tag
component := autoinit.FindByTag(ctx, self, parent, "cache")
```

## Examples

### Example 1: Simple Sibling Discovery

```go
type App struct {
    Logger   *LoggerComponent
    Cache    *CacheComponent
    Service  *ServiceComponent  // Needs logger and cache
}

func (s *ServiceComponent) Init(ctx context.Context, parent interface{}) error {
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find logger sibling
    s.logger = autoinit.FindByType[*LoggerComponent](ctx, s, parent)
    
    // Find cache sibling
    s.cache = autoinit.FindByType[*CacheComponent](ctx, s, parent)
    
    return nil
}
```

### Example 2: Hierarchical Search with Local Override

```go
type System struct {
    GlobalCache *CacheComponent
    Subsystem   *Subsystem
}

type Subsystem struct {
    LocalCache *CacheComponent  // Optional local override
    Service    *ServiceComponent
}

func (s *ServiceComponent) Init(ctx context.Context, parent interface{}) error {
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Will find LocalCache if present, otherwise GlobalCache
    s.cache = autoinit.FindByType[*CacheComponent](ctx, s, parent)
    
    return nil
}
```

### Example 3: Using Multiple Search Criteria

```go
type App struct {
    PrimaryDB   *Database `component:"primary" json:"primaryDb"`
    SecondaryDB *Database `component:"secondary" json:"secondaryDb"`
    Service     *Service
}

func (s *Service) Init(ctx context.Context, parent interface{}) error {
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find primary database by custom tag
    if db := finder.Find(autoinit.SearchOption{
        ByCustomTag: "primary",
        TagKey:      "component",
    }); db != nil {
        s.primaryDB = db.(*Database)
    }
    
    // Find secondary by JSON tag
    if db := finder.Find(autoinit.SearchOption{
        ByJSONTag: "secondaryDb",
    }); db != nil {
        s.secondaryDB = db.(*Database)
    }
    
    return nil
}
```

### Example 4: Complex Multi-Level System

```go
type Application struct {
    GlobalLogger *Logger
    GlobalCache  *Cache
    APIModule    *APIModule
    AdminModule  *AdminModule
}

type APIModule struct {
    LocalCache *Cache      // Optional module-level cache
    Endpoints  []*Endpoint
}

type Endpoint struct {
    logger *Logger
    cache  *Cache
}

func (e *Endpoint) Init(ctx context.Context, parent interface{}) error {
    // Will find in order:
    // 1. LocalCache from APIModule (if present)
    // 2. GlobalCache from Application
    e.cache = autoinit.FindByType[*Cache](ctx, e, parent)
    
    // Will find GlobalLogger from Application
    e.logger = autoinit.FindByType[*Logger](ctx, e, parent)
    
    return nil
}
```

## Collections Support

The finder also searches within slices and maps:

```go
type App struct {
    Caches   []*CacheComponent
    Services map[string]*ServiceComponent
}

// Components in collections can be found
cache := finder.Find(autoinit.SearchOption{
    ByType: reflect.TypeOf((*CacheComponent)(nil)),
})
```

## Best Practices

### 1. Use Interfaces for Loose Coupling

```go
type CacheProvider interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}

// Search for interface implementation
cache := autoinit.FindByInterface[CacheProvider](ctx, self, parent)
```

### 2. Provide Fallbacks

```go
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Try to find cache, but don't fail if not found
    if cache := autoinit.FindByType[*Cache](ctx, s, parent); cache != nil {
        s.cache = cache
        s.cacheEnabled = true
    } else {
        // Work without cache
        s.cacheEnabled = false
    }
    return nil
}
```

### 3. Document Dependencies

```go
// ServiceComponent requires:
// - CacheComponent (searches siblings then ancestors)
// - LoggerComponent (optional, searches by type)
type ServiceComponent struct {
    // ...
}
```

### 4. Use Specific Search Methods

- Use `FindSibling()` when you know the component should be at the same level
- Use `FindAncestor()` when looking for parent-level services
- Use `Find()` for flexible discovery with automatic fallback

## How It Works

1. **Parent Chain Tracking**: During initialization, autoinit maintains a chain of parent structs
2. **Hierarchical Search**: The finder uses this chain to search at different levels
3. **Type Matching**: Supports exact type match, interface implementation, and pointer/value conversions
4. **Tag Inspection**: Uses reflection to read struct tags for tag-based searching
5. **Collection Support**: Automatically searches within slices and maps

## Performance Considerations

- Search operations use reflection but are typically called once during initialization
- The parent chain is maintained only when `WithComponentSearch()` is used
- Searches stop as soon as a match is found (early termination)
- No global state or registries - everything is context-based

## Limitations

- Requires `WithComponentSearch(ctx)` to enable the feature
- Components must use `Init(ctx, parent)` signature to access the finder
- Unexported fields cannot be discovered
- Circular dependencies should be handled carefully

## Summary

The Component Finder provides a powerful way to maintain the plug-and-play nature of autoinit while allowing components to discover their dependencies. By searching siblings first and then moving up the hierarchy, it naturally supports local overrides and hierarchical service provision without explicit wiring.