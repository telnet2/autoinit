// Package autoinit provides component-based initialization for Go applications.
// It treats any struct with an Init method as a "component" - a self-contained,
// pluggable unit that can be composed into larger systems. Components are
// automatically discovered and initialized when you call AutoInit on a parent struct.
//
// This enables plug-and-play architecture where adding new components requires
// no changes to initialization code - just add the component field and it works.
//
// Supports three initialization patterns: Init(), Init(context.Context), and Init(context.Context, interface{}).
package autoinit

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/rs/zerolog"
)

// SimpleInitializer is the basic interface for components that don't need context
type SimpleInitializer interface {
	Init() error
}

// ContextInitializer is the interface for components that need context during initialization
type ContextInitializer interface {
	Init(ctx context.Context) error
}

// ParentInitializer is the interface for components that need to know their parent component
type ParentInitializer interface {
	Init(ctx context.Context, parent interface{}) error
}

// PreInitializer is the interface for pre-initialization hooks
type PreInitializer interface {
	PreInit(ctx context.Context) error
}

// PostInitializer is the interface for post-initialization hooks
type PostInitializer interface {
	PostInit(ctx context.Context) error
}

// PreFieldHook is the interface for parent components to hook before child component initialization
type PreFieldHook interface {
	PreFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error
}

// PostFieldHook is the interface for parent components to hook after child component initialization
type PostFieldHook interface {
	PostFieldInit(ctx context.Context, fieldName string, fieldValue interface{}) error
}

// Options configures the behavior of AutoInit
type Options struct {
	// Logger for trace logging during traversal. If nil, uses default stdout logger
	Logger *zerolog.Logger
	// DisableCycleDetection disables cycle detection (not recommended for production)
	DisableCycleDetection bool
	// RequireTags when true, only initializes components that have an autoinit tag.
	// This gives explicit control over which components are plugged into the system.
	// Components without tags will be skipped (not initialized).
	RequireTags bool
}

// defaultLogger creates a default logger to stdout with trace level
func defaultLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.TraceLevel)
}

// getParentChain retrieves the parent chain from context
func getParentChain(ctx context.Context) *ParentChain {
	if ctx == nil {
		return nil
	}
	chain, _ := ctx.Value(parentChainKey).(*ParentChain)
	return chain
}

// AutoInit recursively discovers and initializes all components in a struct tree.
// A component is any struct that implements one of the Initializer interfaces.
// Components are initialized depth-first in declaration order, enabling plug-and-play
// architecture where you can add new components without changing initialization code.
// Uses default logger for trace logging.
func AutoInit(ctx context.Context, target interface{}) error {
	return AutoInitWithOptions(ctx, target, nil)
}

// AutoInitWithOptions recursively discovers and initializes all components with custom options.
// Components can be added or removed without changing initialization code - just plug them
// into your struct and they'll be automatically initialized.
// If options is nil or Logger is nil, uses default logger to stdout.
// The context and parent reference are propagated through the component tree.
// Supports: Init(), Init(ctx), and Init(ctx, parent) methods.
// Includes cycle detection to prevent infinite loops in component references.
func AutoInitWithOptions(ctx context.Context, target interface{}, options *Options) error {
	// Setup logger
	var logger zerolog.Logger
	if options != nil && options.Logger != nil {
		logger = *options.Logger
	} else {
		logger = defaultLogger()
	}
	
	logger.Trace().
		Str("target_type", fmt.Sprintf("%T", target)).
		Msg("Starting AutoInit")
	
	if target == nil {
		return fmt.Errorf("cannot initialize nil target")
	}

	v := reflect.ValueOf(target)
	
	// If it's a pointer, get the element
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("cannot initialize nil pointer")
		}
		v = v.Elem()
	}
	
	// Must be a struct
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct or pointer to struct, got %s", v.Kind())
	}
	
	// Create visited map for cycle detection (unless disabled)
	var visited map[uintptr]bool
	if options == nil || !options.DisableCycleDetection {
		visited = make(map[uintptr]bool)
	}
	
	// Add parent chain to context if not already present
	if getParentChain(ctx) == nil {
		ctx = WithComponentSearch(ctx)
	}
	
	// Start recursive initialization with no parent (empty reflect.Value)
	err := initStructWithVisited(ctx, v, reflect.Value{}, []string{}, logger, visited, options)
	
	if err != nil {
		logger.Error().
			Err(err).
			Msg("AutoInit failed")
	} else {
		logger.Trace().
			Msg("AutoInit completed successfully")
	}
	
	return err
}

// pathToString converts path slice to dot-separated string
func pathToString(path []string) string {
	if len(path) == 0 {
		return "<root>"
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += "."
		}
		result += p
	}
	return result
}

// initStruct is a wrapper for backward compatibility
func initStruct(ctx context.Context, v reflect.Value, parent reflect.Value, path []string, logger zerolog.Logger) error {
	// Call the new version without cycle detection for backward compatibility
	return initStructWithVisited(ctx, v, parent, path, logger, nil, nil)
}

// initStructWithVisited recursively discovers and initializes all components in a struct.
// Each component (struct with Init method) is initialized after its child components,
// enabling proper dependency order.
func initStructWithVisited(ctx context.Context, v reflect.Value, parent reflect.Value, path []string, logger zerolog.Logger, visited map[uintptr]bool, options *Options) error {
	pathStr := pathToString(path)
	
	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			logger.Trace().
				Str("path", pathStr).
				Msg("Skipping nil pointer")
			return nil // Skip nil pointers
		}
		
		// Check for cycles if cycle detection is enabled
		if visited != nil {
			ptr := v.Pointer()
			if visited[ptr] {
				logger.Trace().
					Str("path", pathStr).
					Msg("Skipping already visited pointer (cycle detected)")
				return nil // Already visited this pointer
			}
			// Mark as visited
			visited[ptr] = true
		}
		
		v = v.Elem()
	}
	
	// Only process structs
	if v.Kind() != reflect.Struct {
		logger.Trace().
			Str("path", pathStr).
			Str("kind", v.Kind().String()).
			Msg("Skipping non-struct field")
		return nil
	}
	
	logger.Trace().
		Str("path", pathStr).
		Str("type", v.Type().String()).
		Msg("Processing struct")
	
	t := v.Type()
	
	// Maintain parent chain for component search
	if chain := getParentChain(ctx); chain != nil {
		// Get the interface value for this struct
		var structInterface interface{}
		if v.CanAddr() {
			structInterface = v.Addr().Interface()
		} else {
			structInterface = v.Interface()
		}
		chain.Push(structInterface)
		defer chain.Pop()
	}
	
	// Call PreInit hook if this struct implements it
	if err := callPreInit(ctx, v, path, logger); err != nil {
		return err
	}
	
	// First, recursively initialize all fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// Skip unexported fields
		if !field.CanInterface() {
			logger.Trace().
				Str("path", pathStr).
				Str("field", fieldType.Name).
				Msg("Skipping unexported field")
			continue
		}
		
		// Check autoinit tag if RequireTags is enabled
		tag := fieldType.Tag.Get("autoinit")
		if tag == "-" {
			// Explicitly skip this field
			logger.Trace().
				Str("path", pathStr).
				Str("field", fieldType.Name).
				Msg("Skipping field with autoinit:\"-\" tag")
			continue
		}
		
		if options != nil && options.RequireTags {
			// When RequireTags is true, only process fields with autoinit tag
			// (empty tag "" or specific values like "init" are OK)
			if _, hasTag := fieldType.Tag.Lookup("autoinit"); !hasTag {
				logger.Trace().
					Str("path", pathStr).
					Str("field", fieldType.Name).
					Msg("Skipping field without autoinit tag (RequireTags enabled)")
				continue
			}
		}
		
		// Create path for error reporting
		fieldPath := append(path, fieldType.Name)
		fieldPathStr := pathToString(fieldPath)
		
		logger.Trace().
			Str("path", fieldPathStr).
			Str("type", field.Type().String()).
			Str("kind", field.Kind().String()).
			Msg("Traversing field")
		
		// Handle different field types
		switch field.Kind() {
		case reflect.Struct:
			// Call parent's PreFieldInit hook if it exists
			if err := callPreFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
				return err
			}
			
			// Recurse into struct fields with current struct as parent
			if err := initStructWithVisited(ctx, field, v, fieldPath, logger, visited, options); err != nil {
				return err
			}
			
			// Call parent's PostFieldInit hook if it exists
			if err := callPostFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
				return err
			}
			
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				// Call parent's PreFieldInit hook if it exists
				if err := callPreFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
				
				// Recurse into pointer to struct with current struct as parent
				if err := initStructWithVisited(ctx, field, v, fieldPath, logger, visited, options); err != nil {
					return err
				}
				
				// Call parent's PostFieldInit hook if it exists
				if err := callPostFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
			}
			
		case reflect.Slice, reflect.Array:
			// Check if this collection contains structs or pointers to structs
			hasInitializableElements := false
			if field.Len() > 0 {
				elemType := field.Type().Elem()
				if elemType.Kind() == reflect.Struct || 
				   (elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) ||
				   elemType.Kind() == reflect.Interface {
					hasInitializableElements = true
				}
			}
			
			// Only call hooks if the collection contains initializable types
			if hasInitializableElements {
				// Call parent's PreFieldInit hook for the collection itself
				if err := callPreFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
			}
			
			// Initialize each element if it's a struct
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				elemPath := append(fieldPath, fmt.Sprintf("[%d]", j))
				if err := initStructWithVisited(ctx, elem, v, elemPath, logger, visited, options); err != nil {
					return err
				}
			}
			
			// Only call hooks if the collection contains initializable types
			if hasInitializableElements {
				// Call parent's PostFieldInit hook for the collection itself
				if err := callPostFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
			}
			
		case reflect.Map:
			// Check if this map contains structs or pointers to structs
			hasInitializableElements := false
			valueType := field.Type().Elem()
			if valueType.Kind() == reflect.Struct || 
			   (valueType.Kind() == reflect.Ptr && valueType.Elem().Kind() == reflect.Struct) ||
			   valueType.Kind() == reflect.Interface {
				hasInitializableElements = true
			}
			
			// Only call hooks if the map contains initializable types
			if hasInitializableElements {
				// Call parent's PreFieldInit hook for the map itself
				if err := callPreFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
			}
			
			// Initialize each map value if it's a struct
			for _, key := range field.MapKeys() {
				elem := field.MapIndex(key)
				elemPath := append(fieldPath, fmt.Sprintf("[%v]", key))
				
				// Map values are not addressable, so we need to handle them specially
				if elem.Kind() == reflect.Struct {
					// For struct values in maps, we need to create a new value,
					// initialize it, and set it back
					newElem := reflect.New(elem.Type()).Elem()
					newElem.Set(elem)
					if err := initStructWithVisited(ctx, newElem.Addr(), v, elemPath, logger, visited, options); err != nil {
						return err
					}
					field.SetMapIndex(key, newElem)
				} else if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					// For pointer values, we can work with them directly
					if err := initStructWithVisited(ctx, elem, v, elemPath, logger, visited, options); err != nil {
						return err
					}
				}
			}
			
			// Only call hooks if the map contains initializable types
			if hasInitializableElements {
				// Call parent's PostFieldInit hook for the map itself
				if err := callPostFieldHook(ctx, v, fieldType.Name, field, logger); err != nil {
					return err
				}
			}
		}
	}
	
	// After initializing all fields, check if this struct itself has Init() method
	if err := callInitIfExists(ctx, v, parent, path, logger); err != nil {
		return err
	}
	
	// Call PostInit hook if this struct implements it
	if err := callPostInit(ctx, v, path, logger); err != nil {
		return err
	}
	
	return nil
}


// initValue handles initialization of a reflect.Value that might be a struct
func initValue(ctx context.Context, v reflect.Value, parent reflect.Value, path []string, logger zerolog.Logger) error {
	// Handle interface values
	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	
	// Only initialize structs
	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && !v.IsNil() && v.Elem().Kind() == reflect.Struct) {
		return initStruct(ctx, v, parent, path, logger)
	}
	
	return nil
}

// callInitIfExists checks if the value has any Init method variant and calls it
// Priority order: Init(ctx, parent) > Init(ctx) > Init()
func callInitIfExists(ctx context.Context, v reflect.Value, parent reflect.Value, path []string, logger zerolog.Logger) error {
	pathStr := pathToString(path)
	
	// Get a pointer to the value if it's not already a pointer
	ptr := v
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		ptr = v.Addr()
	}
	
	// Prepare parent interface{} if parent is valid
	var parentInterface interface{}
	if parent.IsValid() && parent.CanInterface() {
		if parent.Kind() == reflect.Ptr {
			parentInterface = parent.Interface()
		} else if parent.CanAddr() {
			parentInterface = parent.Addr().Interface()
		} else {
			parentInterface = parent.Interface()
		}
	}
	
	// Check for Init(ctx, parent) - highest priority
	if initializer, ok := ptr.Interface().(ParentInitializer); ok {
		logger.Trace().
			Str("path", pathStr).
			Str("type", ptr.Type().String()).
			Str("method", "Init(ctx, parent)").
			Msg("Calling initializer")
		
		if err := initializer.Init(ctx, parentInterface); err != nil {
			logger.Error().
				Str("path", pathStr).
				Err(err).
				Msg("Init(ctx, parent) failed")
			return &InitError{
				Path:      path,
				FieldType: ptr.Type().String(),
				Cause:     err,
			}
		}
		logger.Trace().
			Str("path", pathStr).
			Msg("Init(ctx, parent) completed successfully")
		return nil
	}
	
	// Check for Init(ctx) - second priority
	if initializer, ok := ptr.Interface().(ContextInitializer); ok {
		logger.Trace().
			Str("path", pathStr).
			Str("type", ptr.Type().String()).
			Str("method", "Init(ctx)").
			Msg("Calling initializer")
		
		if err := initializer.Init(ctx); err != nil {
			logger.Error().
				Str("path", pathStr).
				Err(err).
				Msg("Init(ctx) failed")
			return &InitError{
				Path:      path,
				FieldType: ptr.Type().String(),
				Cause:     err,
			}
		}
		logger.Trace().
			Str("path", pathStr).
			Msg("Init(ctx) completed successfully")
		return nil
	}
	
	// Check for Init() - lowest priority
	if initializer, ok := ptr.Interface().(SimpleInitializer); ok {
		logger.Trace().
			Str("path", pathStr).
			Str("type", ptr.Type().String()).
			Str("method", "Init()").
			Msg("Calling initializer")
		
		if err := initializer.Init(); err != nil {
			logger.Error().
				Str("path", pathStr).
				Err(err).
				Msg("Init() failed")
			return &InitError{
				Path:      path,
				FieldType: ptr.Type().String(),
				Cause:     err,
			}
		}
		logger.Trace().
			Str("path", pathStr).
			Msg("Init() completed successfully")
		return nil
	}
	
	// If value receiver, try again with the value itself
	if v.Kind() != reflect.Ptr && v.CanInterface() {
		// Check all three interfaces on the value
		if initializer, ok := v.Interface().(ParentInitializer); ok {
			if v.CanAddr() {
				ptr := v.Addr()
				if init, ok := ptr.Interface().(ParentInitializer); ok {
					if err := init.Init(ctx, parentInterface); err != nil {
						return &InitError{
							Path:      path,
							FieldType: v.Type().String(),
							Cause:     err,
						}
					}
					return nil
				}
			}
			// Can't get address, call on value (won't persist changes)
			if err := initializer.Init(ctx, parentInterface); err != nil {
				return &InitError{
					Path:      path,
					FieldType: v.Type().String(),
					Cause:     err,
				}
			}
			return nil
		}
		
		if initializer, ok := v.Interface().(ContextInitializer); ok {
			if v.CanAddr() {
				ptr := v.Addr()
				if init, ok := ptr.Interface().(ContextInitializer); ok {
					if err := init.Init(ctx); err != nil {
						return &InitError{
							Path:      path,
							FieldType: v.Type().String(),
							Cause:     err,
						}
					}
					return nil
				}
			}
			// Can't get address, call on value (won't persist changes)
			if err := initializer.Init(ctx); err != nil {
				return &InitError{
					Path:      path,
					FieldType: v.Type().String(),
					Cause:     err,
				}
			}
			return nil
		}
		
		if initializer, ok := v.Interface().(SimpleInitializer); ok {
			if v.CanAddr() {
				ptr := v.Addr()
				if init, ok := ptr.Interface().(SimpleInitializer); ok {
					if err := init.Init(); err != nil {
						return &InitError{
							Path:      path,
							FieldType: v.Type().String(),
							Cause:     err,
						}
					}
					return nil
				}
			}
			// Can't get address, call on value (won't persist changes)
			if err := initializer.Init(); err != nil {
				return &InitError{
					Path:      path,
					FieldType: v.Type().String(),
					Cause:     err,
				}
			}
			return nil
		}
	}
	
	return nil
}

// callPreInit calls PreInit hook if the struct implements it
func callPreInit(ctx context.Context, v reflect.Value, path []string, logger zerolog.Logger) error {
	pathStr := pathToString(path)
	
	// Get a pointer to the value if it's not already a pointer
	ptr := v
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		ptr = v.Addr()
	}
	
	if preInit, ok := ptr.Interface().(PreInitializer); ok {
		logger.Trace().
			Str("path", pathStr).
			Str("type", ptr.Type().String()).
			Msg("Calling PreInit")
		
		if err := preInit.PreInit(ctx); err != nil {
			logger.Error().
				Str("path", pathStr).
				Err(err).
				Msg("PreInit failed")
			return &InitError{
				Path:      path,
				FieldType: ptr.Type().String(),
				Cause:     err,
			}
		}
		logger.Trace().
			Str("path", pathStr).
			Msg("PreInit completed successfully")
	}
	
	return nil
}

// callPostInit calls PostInit hook if the struct implements it
func callPostInit(ctx context.Context, v reflect.Value, path []string, logger zerolog.Logger) error {
	pathStr := pathToString(path)
	
	// Get a pointer to the value if it's not already a pointer
	ptr := v
	if v.Kind() != reflect.Ptr && v.CanAddr() {
		ptr = v.Addr()
	}
	
	if postInit, ok := ptr.Interface().(PostInitializer); ok {
		logger.Trace().
			Str("path", pathStr).
			Str("type", ptr.Type().String()).
			Msg("Calling PostInit")
		
		if err := postInit.PostInit(ctx); err != nil {
			logger.Error().
				Str("path", pathStr).
				Err(err).
				Msg("PostInit failed")
			return &InitError{
				Path:      path,
				FieldType: ptr.Type().String(),
				Cause:     err,
			}
		}
		logger.Trace().
			Str("path", pathStr).
			Msg("PostInit completed successfully")
	}
	
	return nil
}

// callPreFieldHook calls parent's PreFieldInit hook if it implements PreFieldHook
func callPreFieldHook(ctx context.Context, parent reflect.Value, fieldName string, fieldValue reflect.Value, logger zerolog.Logger) error {
	if !parent.IsValid() {
		return nil
	}
	
	// Get a pointer to the parent if it's not already a pointer
	parentPtr := parent
	if parent.Kind() != reflect.Ptr && parent.CanAddr() {
		parentPtr = parent.Addr()
	}
	
	// Always pass a pointer to the field to allow modification
	var fieldInterface interface{}
	if fieldValue.CanInterface() {
		if fieldValue.Kind() == reflect.Ptr {
			// Already a pointer
			fieldInterface = fieldValue.Interface()
		} else if fieldValue.CanAddr() {
			// Get pointer to the field
			fieldInterface = fieldValue.Addr().Interface()
		} else {
			// This shouldn't happen for struct fields, but handle it gracefully
			// Log a warning and pass the value itself
			logger.Warn().
				Str("field", fieldName).
				Str("kind", fieldValue.Kind().String()).
				Msg("Cannot get pointer to field, passing value instead")
			fieldInterface = fieldValue.Interface()
		}
	}
	
	if hook, ok := parentPtr.Interface().(PreFieldHook); ok {
		logger.Trace().
			Str("parent_type", parentPtr.Type().String()).
			Str("field", fieldName).
			Msg("Calling PreFieldInit hook")
		
		if err := hook.PreFieldInit(ctx, fieldName, fieldInterface); err != nil {
			logger.Error().
				Str("field", fieldName).
				Err(err).
				Msg("PreFieldInit hook failed")
			return err
		}
		logger.Trace().
			Str("field", fieldName).
			Msg("PreFieldInit hook completed")
	}
	
	return nil
}

// callPostFieldHook calls parent's PostFieldInit hook if it implements PostFieldHook
func callPostFieldHook(ctx context.Context, parent reflect.Value, fieldName string, fieldValue reflect.Value, logger zerolog.Logger) error {
	if !parent.IsValid() {
		return nil
	}
	
	// Get a pointer to the parent if it's not already a pointer
	parentPtr := parent
	if parent.Kind() != reflect.Ptr && parent.CanAddr() {
		parentPtr = parent.Addr()
	}
	
	// Always pass a pointer to the field to allow modification
	var fieldInterface interface{}
	if fieldValue.CanInterface() {
		if fieldValue.Kind() == reflect.Ptr {
			// Already a pointer
			fieldInterface = fieldValue.Interface()
		} else if fieldValue.CanAddr() {
			// Get pointer to the field
			fieldInterface = fieldValue.Addr().Interface()
		} else {
			// This shouldn't happen for struct fields, but handle it gracefully
			// Log a warning and pass the value itself
			logger.Warn().
				Str("field", fieldName).
				Str("kind", fieldValue.Kind().String()).
				Msg("Cannot get pointer to field, passing value instead")
			fieldInterface = fieldValue.Interface()
		}
	}
	
	if hook, ok := parentPtr.Interface().(PostFieldHook); ok {
		logger.Trace().
			Str("parent_type", parentPtr.Type().String()).
			Str("field", fieldName).
			Msg("Calling PostFieldInit hook")
		
		if err := hook.PostFieldInit(ctx, fieldName, fieldInterface); err != nil {
			logger.Error().
				Str("field", fieldName).
				Err(err).
				Msg("PostFieldInit hook failed")
			return err
		}
		logger.Trace().
			Str("field", fieldName).
			Msg("PostFieldInit hook completed")
	}
	
	return nil
}