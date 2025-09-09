# Component Discovery Patterns Comparison

AutoInit provides two complementary patterns for component discovery: the modern **As Pattern** and the classic **Finder Pattern**. This document compares both approaches to help you choose the right one for your needs.

## 📊 Quick Comparison Table

| Feature | As Pattern | Finder Pattern |
|---------|------------|----------------|
| **API Style** | Go CDK-inspired, type-safe | Flexible, option-based |
| **Type Safety** | ✅ Compile-time with generics | ⚠️ Runtime with type assertions |
| **Filtering Logic** | Conjunctive (AND) | Single criteria |
| **Code Verbosity** | Minimal | Moderate |
| **Learning Curve** | Simple | Moderate |
| **Flexibility** | Focused | Very flexible |
| **Best For** | Type-safe dependency injection | Complex discovery scenarios |

## 🎯 As Pattern (Recommended)

The As pattern is inspired by Go CDK's escape hatch design, providing a clean, type-safe way to discover dependencies.

### Advantages

1. **Type Safety**: Full compile-time type checking with generics
2. **Clean API**: Familiar to Go developers who've used Go CDK
3. **Conjunctive Filtering**: Multiple filters work together (AND logic)
4. **Minimal Boilerplate**: Simple, readable syntax
5. **Error Handling**: Clear with `MustAs` for required dependencies

### Example Usage

```go
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Simple type discovery
    var db *Database
    if autoinit.As(ctx, s, parent, &db) {
        s.db = db
    }
    
    // Conjunctive filtering - ALL conditions must match
    var primaryCache *Cache
    if autoinit.As(ctx, s, parent, &primaryCache,
        autoinit.WithFieldName("PrimaryCache"),
        autoinit.WithJSONTag("main")) {
        // Found Cache that is BOTH named "PrimaryCache" AND has json:"main"
        s.cache = primaryCache
    }
    
    // Interface discovery
    var logger Logger
    if autoinit.As(ctx, s, parent, &logger) {
        s.logger = logger
    }
    
    // Required dependency (panics if not found)
    autoinit.MustAs(ctx, s, parent, &s.requiredService)
    
    return nil
}
```

### When to Use As Pattern

- ✅ **Standard dependency injection** - Most common use case
- ✅ **Type-safe requirements** - When compile-time safety is important
- ✅ **Multiple criteria** - When you need to filter by type AND other properties
- ✅ **Clean code preference** - When you want minimal, readable code
- ✅ **Go CDK familiarity** - If your team knows Go CDK patterns

## 🔍 Finder Pattern (Classic)

The original finder pattern provides maximum flexibility for complex discovery scenarios.

### Advantages

1. **Flexibility**: Can search with any combination of criteria
2. **Hierarchical Search**: Explicit control over search scope
3. **Multiple Search Methods**: Dedicated methods for different search types
4. **Context Integration**: Works with parent chain for complex hierarchies
5. **Fine Control**: More control over search behavior

### Example Usage

```go
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Enable component search
    ctx = autoinit.WithComponentSearch(ctx)
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Find by type with helper
    s.cache = autoinit.FindByType[*Cache](ctx, s, parent)
    
    // Find by field name
    if logger := finder.Find(autoinit.SearchOption{
        ByFieldName: "Logger",
    }); logger != nil {
        s.logger = logger.(*LoggerComponent)
    }
    
    // Find by JSON tag
    if db := finder.Find(autoinit.SearchOption{
        ByJSONTag: "primary",
    }); db != nil {
        s.db = db.(*Database)
    }
    
    // Find only among siblings
    if peer := finder.FindSibling(autoinit.SearchOption{
        ByType: reflect.TypeOf((*PeerService)(nil)).Elem(),
    }); peer != nil {
        s.peer = peer.(*PeerService)
    }
    
    return nil
}
```

### When to Use Finder Pattern

- ✅ **Complex hierarchies** - When you need to search ancestors/siblings specifically
- ✅ **Dynamic discovery** - When search criteria is determined at runtime
- ✅ **Custom search logic** - When you need fine control over search behavior
- ✅ **Legacy code** - Already using finder pattern extensively
- ✅ **Advanced scenarios** - Complex component relationships

## 🔄 Migration Guide

### From Finder to As Pattern

Most finder pattern usage can be simplified with the As pattern:

**Before (Finder Pattern):**
```go
ctx = autoinit.WithComponentSearch(ctx)
finder := autoinit.NewComponentFinder(ctx, s, parent)

// Find by type
if result := finder.Find(autoinit.SearchOption{
    ByType: reflect.TypeOf((*Database)(nil)).Elem(),
}); result != nil {
    s.db = result.(*Database)
}

// Find by field name and type
if result := finder.Find(autoinit.SearchOption{
    ByFieldName: "PrimaryCache",
}); result != nil {
    if cache, ok := result.(*Cache); ok {
        s.cache = cache
    }
}
```

**After (As Pattern):**
```go
// Find by type
autoinit.As(ctx, s, parent, &s.db)

// Find by field name with type safety
var cache *Cache
if autoinit.As(ctx, s, parent, &cache, 
    autoinit.WithFieldName("PrimaryCache")) {
    s.cache = cache
}
```

## 🤝 Using Both Patterns Together

Both patterns can coexist in the same codebase. You might use:

- **As Pattern** for standard dependency injection
- **Finder Pattern** for complex hierarchical searches

```go
func (s *Service) Init(ctx context.Context, parent interface{}) error {
    // Use As for simple dependencies
    autoinit.MustAs(ctx, s, parent, &s.db)
    autoinit.As(ctx, s, parent, &s.cache)
    
    // Use Finder for complex hierarchical search
    ctx = autoinit.WithComponentSearch(ctx)
    finder := autoinit.NewComponentFinder(ctx, s, parent)
    
    // Search only ancestors (not siblings)
    if ancestor := finder.FindAncestor(autoinit.SearchOption{
        ByType: reflect.TypeOf((*RootConfig)(nil)).Elem(),
    }); ancestor != nil {
        s.config = ancestor.(*RootConfig)
    }
    
    return nil
}
```

## 📈 Performance Considerations

Both patterns have similar performance characteristics:

| Aspect | As Pattern | Finder Pattern |
|--------|------------|----------------|
| **Search Speed** | O(n) where n = number of fields | O(n) where n = number of fields |
| **Memory Usage** | Minimal | Slightly higher (context chain) |
| **Reflection Overhead** | Same | Same |
| **Type Assertion** | None (compile-time) | Runtime overhead |

## 🎓 Best Practices

### For As Pattern

1. **Use MustAs for required dependencies** - Makes requirements explicit
2. **Order filters by selectivity** - Most selective filters first
3. **Prefer interfaces** - More flexible than concrete types
4. **Keep filters simple** - Don't over-constrain

### For Finder Pattern

1. **Enable search context early** - Call `WithComponentSearch` at start
2. **Cache finder instance** - Reuse for multiple searches
3. **Check nil returns** - Always handle not-found cases
4. **Use type-safe helpers** - Prefer `FindByType[T]` over manual search

## 📋 Decision Matrix

Choose **As Pattern** when:
- ✅ You want type-safe, compile-time checked code
- ✅ You need to match multiple criteria (AND logic)
- ✅ You prefer clean, minimal syntax
- ✅ You're building new components
- ✅ Your team is familiar with Go CDK patterns

Choose **Finder Pattern** when:
- ✅ You need complex hierarchical searches
- ✅ You require fine control over search scope
- ✅ You have dynamic search requirements
- ✅ You're working with legacy code
- ✅ You need maximum flexibility

## 🚀 Conclusion

**For most use cases, we recommend the As pattern** due to its simplicity, type safety, and clean API. It covers 90% of dependency discovery needs with minimal code.

The Finder pattern remains valuable for advanced scenarios requiring complex hierarchical searches or fine-grained control over discovery behavior.

Both patterns are maintained and supported, allowing you to choose the best tool for your specific needs.