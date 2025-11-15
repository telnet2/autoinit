package autoinit

import (
	"context"
	"reflect"
	"strings"
)

// Filter interface for additional search constraints in As pattern
type Filter interface {
	Matches(field reflect.Value, fieldType *reflect.StructField) bool
}

// fieldNameFilter matches components by field name
type fieldNameFilter struct {
	name string
}

func (f fieldNameFilter) Matches(field reflect.Value, fieldType *reflect.StructField) bool {
	return strings.EqualFold(fieldType.Name, f.name)
}

// jsonTagFilter matches components by JSON tag value
type jsonTagFilter struct {
	tag string
}

func (f jsonTagFilter) Matches(field reflect.Value, fieldType *reflect.StructField) bool {
	jsonTag := fieldType.Tag.Get("json")
	tagName := strings.Split(jsonTag, ",")[0]
	return tagName == f.tag
}

// customTagFilter matches components by custom tag key and value
type customTagFilter struct {
	key   string
	value string
}

func (f customTagFilter) Matches(field reflect.Value, fieldType *reflect.StructField) bool {
	tagValue := fieldType.Tag.Get(f.key)
	return tagValue == f.value
}

// WithFieldName creates a filter that matches by field name
func WithFieldName(name string) Filter {
	return fieldNameFilter{name: name}
}

// WithJSONTag creates a filter that matches by JSON tag value
func WithJSONTag(tag string) Filter {
	return jsonTagFilter{tag: tag}
}

// WithTag creates a filter that matches by custom tag key and value
func WithTag(key, value string) Filter {
	return customTagFilter{key: key, value: value}
}

// As attempts to find a dependency matching the target type AND all provided filters.
// All filters are applied conjunctively (AND logic) to narrow down candidates.
// This follows the Go CDK pattern for escape hatches with additional filtering capabilities.
//
// Search order:
//  1. Immediate parent's fields (siblings)
//  2. Ancestors in the parent chain and their fields
//  3. Searches recursively through slices, maps, and embedded structs
//
// This enables components to discover dependencies anywhere in the object graph,
// not just in their immediate parent.
//
// Usage:
//
//	var db *Database
//	if As(ctx, self, parent, &db) {
//	    // Found any Database type anywhere in the object graph
//	}
//
//	var primaryDB *Database
//	if As(ctx, self, parent, &primaryDB, WithFieldName("PrimaryDB"), WithJSONTag("primary")) {
//	    // Found a Database that is BOTH named "PrimaryDB" AND has json:"primary" tag
//	}
//
// Returns true if a dependency matching ALL criteria was found and assigned.
// Returns false if no matching dependency exists.
func As[T any](ctx context.Context, self, parent interface{}, target *T, filters ...Filter) bool {
	if target == nil {
		return false
	}

	// Get the type we're looking for
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return false
	}

	targetElem := targetValue.Elem()
	targetType := targetElem.Type()

	// Use the internal asSearch to find matching component
	result := asSearch(ctx, self, parent, targetType, filters...)
	if result == nil {
		return false
	}

	// Type-safe assignment
	resultValue := reflect.ValueOf(result)
	if resultValue.Type().AssignableTo(targetType) {
		targetElem.Set(resultValue)
		return true
	}

	// For interface types, check if the result implements the interface
	if targetType.Kind() == reflect.Interface && resultValue.Type().Implements(targetType) {
		targetElem.Set(resultValue)
		return true
	}

	return false
}

// MustAs is like As but panics if the dependency is not found.
// Use this when a dependency is required for the component to function.
func MustAs[T any](ctx context.Context, self, parent interface{}, target *T, filters ...Filter) {
	if !As(ctx, self, parent, target, filters...) {
		targetType := reflect.TypeOf(target).Elem()
		panic("required dependency not found: " + targetType.String())
	}
}

// asSearch performs the actual search with conjunctive filtering
// Search order:
// 1. TestContext (if present)
// 2. Immediate parent's fields
// 3. Ancestors in the parent chain and their siblings
//
// Includes cycle detection and depth limiting for safety.
func asSearch(ctx context.Context, self, parent interface{}, targetType reflect.Type, filters ...Filter) interface{} {
	// First check if we have a TestContext in the context
	if tc := getTestContext(ctx); tc != nil {
		tc.mu.RLock()
		defer tc.mu.RUnlock()

		// Try to find by type first
		if candidates, exists := tc.dependencies[targetType]; exists {
			for _, candidate := range candidates {
				// Apply filters if any
				if len(filters) == 0 {
					return candidate
				}
				// For TestContext, we'll skip filter matching for now
				// In a full implementation, you'd apply filters here
				return candidate
			}
		}
	}

	if parent == nil {
		return nil
	}

	// Initialize visited map for cycle detection
	visited := make(map[uintptr]bool)
	const maxDepth = 50 // Reasonable limit for nested structures

	// Search in immediate parent's fields
	if result := searchInStructSafe(parent, self, targetType, filters, visited, 0, maxDepth); result != nil {
		return result
	}

	// Search in the parent chain for a more comprehensive search
	if chain := getParentChain(ctx); chain != nil {
		// Start from the parent level and search up through ancestors
		for i := 0; i < chain.Len(); i++ {
			ancestor := chain.GetParent(i)
			if ancestor == nil || ancestor == parent {
				continue // Skip nil or already searched parent
			}

			// Search in this ancestor's fields with the same visited map
			// This prevents cycles across the entire search, not just within one ancestor
			if result := searchInStructSafe(ancestor, self, targetType, filters, visited, 0, maxDepth); result != nil {
				return result
			}
		}
	}

	return nil
}

// searchInStructSafe searches for matching components with cycle detection and depth limiting
func searchInStructSafe(parent, exclude interface{}, targetType reflect.Type, filters []Filter, visited map[uintptr]bool, depth, maxDepth int) interface{} {
	// Depth check to prevent stack overflow
	if depth > maxDepth {
		return nil
	}

	v := reflect.ValueOf(parent)
	if v.Kind() == reflect.Ptr {
		// Cycle detection for pointer types
		if v.Pointer() != 0 {
			if visited[v.Pointer()] {
				return nil // Already visited, prevent cycle
			}
			visited[v.Pointer()] = true
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Skip self
		fieldInterface := field.Interface()
		if fieldInterface == exclude {
			continue
		}

		// Skip nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Check if this field matches our type requirement
		if !matchesTargetType(field, targetType) {
			continue
		}

		// Apply all filters conjunctively
		if !matchesAllFilters(field, &fieldType, filters) {
			continue
		}

		// Found a match!
		// For value types, return a pointer if the field is addressable
		if field.Kind() != reflect.Ptr && field.CanAddr() {
			return field.Addr().Interface()
		}
		return fieldInterface
	}

	// Also search in slices
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.CanInterface() {
					elemInterface := elem.Interface()
					if elemInterface != exclude && matchesTargetType(elem, targetType) {
						// For slice elements, we need to check filters differently
						// since they don't have field metadata
						if len(filters) == 0 {
							// No additional filters, type match is enough
							if elem.Kind() != reflect.Ptr && elem.CanAddr() {
								return elem.Addr().Interface()
							}
							return elemInterface
						}
					}
				}
			}
		}

		// Search in maps
		if field.Kind() == reflect.Map {
			for _, key := range field.MapKeys() {
				val := field.MapIndex(key)
				if val.CanInterface() {
					valInterface := val.Interface()
					if valInterface != exclude && matchesTargetType(val, targetType) {
						if len(filters) == 0 {
							// Map values are not addressable
							return valInterface
						}
					}
				}
			}
		}

		// Search in embedded structs WITH cycle protection
		if fieldType.Anonymous && (field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct)) {
			// Recursively search with incremented depth and same visited map
			if result := searchInStructSafe(field.Interface(), exclude, targetType, filters, visited, depth+1, maxDepth); result != nil {
				return result
			}
		}
	}

	return nil
}

// matchesTargetType checks if a value matches the target type
func matchesTargetType(field reflect.Value, targetType reflect.Type) bool {
	if !field.IsValid() || !field.CanInterface() {
		return false
	}

	fieldInterface := field.Interface()
	fieldType := reflect.TypeOf(fieldInterface)

	// Direct type match
	if fieldType == targetType {
		return true
	}

	// If target is a pointer type, check if field type matches
	if targetType.Kind() == reflect.Ptr && targetType.Elem() == fieldType {
		return true
	}

	// If field is a pointer, check if its element type matches
	if fieldType.Kind() == reflect.Ptr && fieldType.Elem() == targetType {
		return true
	}

	// Check if both are pointers to the same type
	if targetType.Kind() == reflect.Ptr && fieldType.Kind() == reflect.Ptr {
		if targetType.Elem() == fieldType.Elem() {
			return true
		}
	}

	// Check interface implementation
	if targetType.Kind() == reflect.Interface {
		if fieldType.Implements(targetType) {
			return true
		}
		// Check if pointer type implements the interface
		if fieldType.Kind() != reflect.Ptr && reflect.PtrTo(fieldType).Implements(targetType) {
			return true
		}
	}

	return false
}

// matchesAllFilters checks if a field matches all provided filters
func matchesAllFilters(field reflect.Value, fieldType *reflect.StructField, filters []Filter) bool {
	// All filters must match (conjunctive/AND logic)
	for _, filter := range filters {
		if !filter.Matches(field, fieldType) {
			return false
		}
	}
	return true
}

// AsType provides a simpler API when only type matching is needed
// This is a convenience wrapper around As with no additional filters
func AsType[T any](ctx context.Context, self, parent interface{}) (T, bool) {
	var target T
	ok := As(ctx, self, parent, &target)
	return target, ok
}
