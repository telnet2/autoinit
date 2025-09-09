package autoinit

import (
	"context"
	"reflect"
	"strings"
)

// SearchOption configures how to search for components
type SearchOption struct {
	// Search by type match
	ByType reflect.Type

	// Search by json tag value
	ByJSONTag string

	// Search by field name
	ByFieldName string

	// Search by custom tag
	ByCustomTag string
	TagKey      string // e.g., "component" for `component:"cache"`
}

// ComponentFinder provides methods to find related components
type ComponentFinder struct {
	ctx    context.Context
	parent interface{}
	self   interface{}
}

// NewComponentFinder creates a finder for the current component
func NewComponentFinder(ctx context.Context, self, parent interface{}) *ComponentFinder {
	return &ComponentFinder{
		ctx:    ctx,
		parent: parent,
		self:   self,
	}
}

// Find searches for a component using the specified options
// Search order:
// 1. Siblings at current level
// 2. Parent's siblings (aunts/uncles)
// 3. Grandparent's siblings, etc.
func (cf *ComponentFinder) Find(opt SearchOption) interface{} {
	return cf.searchHierarchy(cf.parent, cf.self, opt, 0)
}

// FindSibling searches only among siblings at the same level
func (cf *ComponentFinder) FindSibling(opt SearchOption) interface{} {
	if cf.parent == nil {
		return nil
	}
	return cf.searchSiblings(cf.parent, cf.self, opt)
}

// FindAncestor searches up the parent chain for a matching component
func (cf *ComponentFinder) FindAncestor(opt SearchOption) interface{} {
	return cf.searchAncestors(cf.parent, opt)
}

// searchHierarchy implements the full search algorithm
func (cf *ComponentFinder) searchHierarchy(current interface{}, exclude interface{}, opt SearchOption, depth int) interface{} {
	if current == nil || depth > 10 { // Prevent infinite recursion
		return nil
	}

	// Step 1: Search siblings at current level
	if result := cf.searchSiblings(current, exclude, opt); result != nil {
		return result
	}

	// Step 2: Get parent chain from context to search higher levels
	if chain := cf.getParentChain(); chain != nil {
		// Search each level up the hierarchy
		for i := depth + 1; i < len(chain.chain); i++ {
			ancestor := chain.chain[len(chain.chain)-1-i]
			if ancestor == nil {
				continue
			}

			// Search siblings at this ancestor level
			// Exclude the child that we came from
			var excludeAtLevel interface{}
			if i > 0 && i-1 < len(chain.chain) {
				excludeAtLevel = chain.chain[len(chain.chain)-i]
			}

			if result := cf.searchSiblings(ancestor, excludeAtLevel, opt); result != nil {
				return result
			}
		}
	}

	return nil
}

// searchSiblings searches among sibling components in the same parent
func (cf *ComponentFinder) searchSiblings(parent interface{}, exclude interface{}, opt SearchOption) interface{} {
	if parent == nil {
		return nil
	}

	v := reflect.ValueOf(parent)
	if v.Kind() == reflect.Ptr {
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

		// Check if this field matches our search criteria
		if cf.matchesOption(field, &fieldType, opt) {
			// For value types, return a pointer if the field is addressable
			// This allows the found component to be modified
			if field.Kind() != reflect.Ptr && field.CanAddr() {
				return field.Addr().Interface()
			}
			return fieldInterface
		}

		// For embedded structs, search their fields too
		if fieldType.Anonymous && (field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct)) {
			if result := cf.searchSiblings(fieldInterface, exclude, opt); result != nil {
				return result
			}
		}

		// For slices and arrays, check each element
		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				if elem.CanInterface() {
					elemInterface := elem.Interface()
					if elemInterface != exclude && cf.matchesValue(elem, opt) {
						// For value types in collections, return pointer if addressable
						if elem.Kind() != reflect.Ptr && elem.CanAddr() {
							return elem.Addr().Interface()
						}
						return elemInterface
					}
				}
			}
		}

		// For maps, check each value
		if field.Kind() == reflect.Map {
			for _, key := range field.MapKeys() {
				val := field.MapIndex(key)
				if val.CanInterface() {
					valInterface := val.Interface()
					if valInterface != exclude && cf.matchesValue(val, opt) {
						// Map values are not addressable, so we can't return pointers
						// This is a Go limitation
						return valInterface
					}
				}
			}
		}
	}

	return nil
}

// matchesOption checks if a field matches the search criteria
func (cf *ComponentFinder) matchesOption(field reflect.Value, fieldType *reflect.StructField, opt SearchOption) bool {
	// Match by type
	if opt.ByType != nil {
		if cf.matchesType(field, opt.ByType) {
			return true
		}
	}

	// Match by field name
	if opt.ByFieldName != "" {
		if strings.EqualFold(fieldType.Name, opt.ByFieldName) {
			return true
		}
	}

	// Match by JSON tag
	if opt.ByJSONTag != "" {
		jsonTag := fieldType.Tag.Get("json")
		tagName := strings.Split(jsonTag, ",")[0]
		if tagName == opt.ByJSONTag {
			return true
		}
	}

	// Match by custom tag
	if opt.ByCustomTag != "" && opt.TagKey != "" {
		tagValue := fieldType.Tag.Get(opt.TagKey)
		if tagValue == opt.ByCustomTag {
			return true
		}
	}

	return false
}

// matchesValue checks if a value matches the search criteria (for elements in collections)
func (cf *ComponentFinder) matchesValue(val reflect.Value, opt SearchOption) bool {
	if opt.ByType != nil {
		return cf.matchesType(val, opt.ByType)
	}
	return false
}

// matchesType checks if a value matches the target type
func (cf *ComponentFinder) matchesType(field reflect.Value, targetType reflect.Type) bool {
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

// searchAncestors searches up the parent chain
func (cf *ComponentFinder) searchAncestors(parent interface{}, opt SearchOption) interface{} {
	if chain := cf.getParentChain(); chain != nil {
		// Skip the first item (self) and search up
		for i := 1; i < len(chain.chain); i++ {
			ancestor := chain.chain[len(chain.chain)-1-i]
			if ancestor == nil {
				continue
			}

			// Check if the ancestor itself matches
			v := reflect.ValueOf(ancestor)
			if cf.matchesType(v, opt.ByType) {
				return ancestor
			}
		}
	}

	return nil
}

// getParentChain retrieves the parent chain from context
func (cf *ComponentFinder) getParentChain() *ParentChain {
	if cf.ctx == nil {
		return nil
	}
	chain, _ := cf.ctx.Value(parentChainKey).(*ParentChain)
	return chain
}

// ParentChain maintains the hierarchy during initialization
type ParentChain struct {
	chain []interface{}
}

// Push adds a parent to the chain
func (pc *ParentChain) Push(parent interface{}) {
	pc.chain = append(pc.chain, parent)
}

// Pop removes the last parent from the chain
func (pc *ParentChain) Pop() {
	if len(pc.chain) > 0 {
		pc.chain = pc.chain[:len(pc.chain)-1]
	}
}

// GetParent returns a parent at the specified level (0 = immediate parent)
func (pc *ParentChain) GetParent(level int) interface{} {
	if level >= len(pc.chain) {
		return nil
	}
	return pc.chain[len(pc.chain)-1-level]
}

// Len returns the depth of the parent chain
func (pc *ParentChain) Len() int {
	return len(pc.chain)
}

// parentChainKey is the context key for the parent chain
type contextKey string

const parentChainKey contextKey = "autoinit:parentChain"

// WithComponentSearch enables component search capabilities by adding a parent chain to context
func WithComponentSearch(ctx context.Context) context.Context {
	chain := &ParentChain{
		chain: make([]interface{}, 0, 10), // Pre-allocate for typical depth
	}
	return context.WithValue(ctx, parentChainKey, chain)
}

// Helper functions for common search patterns

// FindByType searches for a component by its type
func FindByType[T any](ctx context.Context, self, parent interface{}) T {
	var zero T
	finder := NewComponentFinder(ctx, self, parent)
	result := finder.Find(SearchOption{
		ByType: reflect.TypeOf((*T)(nil)).Elem(),
	})
	if result != nil {
		if typed, ok := result.(T); ok {
			return typed
		}
	}
	return zero
}

// FindByInterface searches for a component that implements an interface
func FindByInterface[T any](ctx context.Context, self, parent interface{}) T {
	var zero T
	finder := NewComponentFinder(ctx, self, parent)

	// Get the interface type
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	if interfaceType.Kind() != reflect.Interface {
		return zero
	}

	result := finder.Find(SearchOption{
		ByType: interfaceType,
	})
	if result != nil {
		if typed, ok := result.(T); ok {
			return typed
		}
	}
	return zero
}

// FindByName searches for a component by field name
func FindByName(ctx context.Context, self, parent interface{}, name string) interface{} {
	finder := NewComponentFinder(ctx, self, parent)
	return finder.Find(SearchOption{
		ByFieldName: name,
	})
}

// FindByTag searches for a component by JSON tag
func FindByTag(ctx context.Context, self, parent interface{}, tag string) interface{} {
	finder := NewComponentFinder(ctx, self, parent)
	return finder.Find(SearchOption{
		ByJSONTag: tag,
	})
}

// FindByCustomTag searches for a component by custom tag
func FindByCustomTag(ctx context.Context, self, parent interface{}, tagKey, tagValue string) interface{} {
	finder := NewComponentFinder(ctx, self, parent)
	return finder.Find(SearchOption{
		ByCustomTag: tagValue,
		TagKey:      tagKey,
	})
}
