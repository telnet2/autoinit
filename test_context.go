package autoinit

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

// ErrComponentNotFound is returned when a component cannot be found
var ErrComponentNotFound = errors.New("component not found")

// TestContext provides isolated dependency discovery for unit testing.
// It allows testing components that use dependency discovery without
// constructing the full application tree.
type TestContext struct {
	dependencies map[reflect.Type][]interface{}
	namedDeps    map[string]interface{}
	taggedDeps   map[string][]interface{}
	mu           sync.RWMutex
}

// NewTestContext creates a new isolated test context for dependency discovery.
func NewTestContext() *TestContext {
	return &TestContext{
		dependencies: make(map[reflect.Type][]interface{}),
		namedDeps:    make(map[string]interface{}),
		taggedDeps:   make(map[string][]interface{}),
	}
}

// Register adds a dependency to the test context that can be discovered by type.
func (tc *TestContext) Register(dep interface{}) *TestContext {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	depType := reflect.TypeOf(dep)

	// Register by concrete type
	tc.dependencies[depType] = append(tc.dependencies[depType], dep)

	return tc
}

// RegisterInterface registers a dependency that implements a specific interface type.
// This allows the dependency to be discovered when searching for the interface type.
func (tc *TestContext) RegisterInterface(interfaceType reflect.Type, dep interface{}) *TestContext {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	depType := reflect.TypeOf(dep)

	// Register by concrete type
	tc.dependencies[depType] = append(tc.dependencies[depType], dep)

	// Register by interface type if it implements it
	if depType.Implements(interfaceType) {
		tc.dependencies[interfaceType] = append(tc.dependencies[interfaceType], dep)
	}

	return tc
}

// RegisterNamed adds a named dependency that can be discovered by field name.
func (tc *TestContext) RegisterNamed(name string, dep interface{}) *TestContext {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.namedDeps[name] = dep
	// Also register by type
	t := reflect.TypeOf(dep)
	tc.dependencies[t] = append(tc.dependencies[t], dep)
	return tc
}

// RegisterTagged adds a dependency that can be discovered by tag value.
func (tc *TestContext) RegisterTagged(tag string, dep interface{}) *TestContext {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.taggedDeps[tag] = append(tc.taggedDeps[tag], dep)
	// Also register by type
	t := reflect.TypeOf(dep)
	tc.dependencies[t] = append(tc.dependencies[t], dep)
	return tc
}

// Context returns a context.Context that can be used with As/MustAs for dependency discovery.
func (tc *TestContext) Context() context.Context {
	return context.WithValue(context.Background(), testCtxKey, tc)
}

// ContextWithParent returns a context with a virtual parent for dependency discovery.
func (tc *TestContext) ContextWithParent(parent interface{}) context.Context {
	ctx := tc.Context()
	return context.WithValue(ctx, parentChainKey, &ParentChain{
		chain: []interface{}{parent},
	})
}

// testCtxKey is the context key for TestContext
type testCtxKeyType string

const testCtxKey testCtxKeyType = "test-context"

// getTestContext retrieves TestContext from context
func getTestContext(ctx context.Context) *TestContext {
	if ctx == nil {
		return nil
	}
	tc, _ := ctx.Value(testCtxKey).(*TestContext)
	return tc
}

// TestAs performs dependency discovery in a test context.
// This is the test-friendly version of As that works with TestContext.
func TestAs[T any](ctx context.Context, target interface{}, dest *T, filters ...Filter) error {
	tc := getTestContext(ctx)
	if tc == nil {
		// Fallback to regular As if no test context
		if As(ctx, target, nil, dest, filters...) {
			return nil
		}
		return ErrComponentNotFound
	}

	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var targetType reflect.Type
	if dest != nil {
		targetType = reflect.TypeOf(dest).Elem()
	}

	// Try to find by type first
	if candidates, exists := tc.dependencies[targetType]; exists {
		for _, candidate := range candidates {
			*dest = candidate.(T)
			return nil
		}
	}

	return ErrComponentNotFound
}

// TestMustAs is like TestAs but panics if the dependency is not found.
func TestMustAs[T any](ctx context.Context, target interface{}, dest *T, filters ...Filter) {
	if err := TestAs(ctx, target, dest, filters...); err != nil {
		panic(err)
	}
}

// TestBuilder provides a fluent interface for building test contexts.
type TestBuilder struct {
	tc *TestContext
}

// NewTestBuilder creates a new test builder.
func NewTestBuilder() *TestBuilder {
	return &TestBuilder{
		tc: NewTestContext(),
	}
}

// WithDependency adds a dependency by type.
func (tb *TestBuilder) WithDependency(dep interface{}) *TestBuilder {
	tb.tc.Register(dep)
	return tb
}

// WithInterfaceDependency adds a dependency that implements a specific interface type.
func (tb *TestBuilder) WithInterfaceDependency(interfaceType reflect.Type, dep interface{}) *TestBuilder {
	tb.tc.RegisterInterface(interfaceType, dep)
	return tb
}

// WithNamedDependency adds a named dependency.
func (tb *TestBuilder) WithNamedDependency(name string, dep interface{}) *TestBuilder {
	tb.tc.RegisterNamed(name, dep)
	return tb
}

// WithTaggedDependency adds a tagged dependency.
func (tb *TestBuilder) WithTaggedDependency(tag string, dep interface{}) *TestBuilder {
	tb.tc.RegisterTagged(tag, dep)
	return tb
}

// Build returns the configured TestContext.
func (tb *TestBuilder) Build() *TestContext {
	return tb.tc
}

// Context returns a context with the configured dependencies.
func (tb *TestBuilder) Context() context.Context {
	return tb.tc.Context()
}
