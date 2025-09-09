package autoinit

import (
	"fmt"
	"strings"
)

// InitError represents an error that occurred during initialization
type InitError struct {
	Path      []string // Full path to the failing field
	FieldType string   // Type of the field that failed
	Cause     error    // Original error from Init()
}

// Error implements the error interface with detailed context
func (e *InitError) Error() string {
	if len(e.Path) == 0 {
		return fmt.Sprintf("failed to initialize %s: %v", e.FieldType, e.Cause)
	}

	pathStr := strings.Join(e.Path, ".")
	return fmt.Sprintf("failed to initialize field '%s' of type %s: %v", pathStr, e.FieldType, e.Cause)
}

// Unwrap returns the underlying error for error unwrapping support
func (e *InitError) Unwrap() error {
	return e.Cause
}

// GetPath returns the full path to the field that failed initialization
func (e *InitError) GetPath() []string {
	return e.Path
}

// GetFieldType returns the type of the field that failed
func (e *InitError) GetFieldType() string {
	return e.FieldType
}
