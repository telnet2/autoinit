package autoinit

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// LoggingComponent for testing trace logging
type LoggingComponent struct {
	Name   string
	Nested NestedLogging
	Initialized bool
}

func (l *LoggingComponent) Init(ctx context.Context) error {
	l.Initialized = true
	return nil
}

type NestedLogging struct {
	Value string
	Initialized bool
}

func (n *NestedLogging) Init() error {
	n.Initialized = true
	n.Value = "nested"
	return nil
}

// Test default logger (stdout)
func TestDefaultLogger(t *testing.T) {
	component := &LoggingComponent{
		Name: "test",
	}
	
	ctx := context.Background()
	err := AutoInit(ctx, component)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
	
	if !component.Nested.Initialized {
		t.Error("nested component was not initialized")
	}
}

// Test custom logger with captured output
func TestCustomLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger().Level(zerolog.TraceLevel)
	
	component := &LoggingComponent{
		Name: "test",
	}
	
	ctx := context.Background()
	options := &Options{
		Logger: &logger,
	}
	
	err := AutoInitWithOptions(ctx, component, options)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check that logging occurred
	logOutput := buf.String()
	
	// Verify key log messages are present
	expectedMessages := []string{
		"Starting AutoInit",
		"Processing struct",
		"Traversing field",
		"Calling initializer",
		"completed successfully",
		"AutoInit completed successfully",
	}
	
	for _, msg := range expectedMessages {
		if !strings.Contains(logOutput, msg) {
			t.Errorf("expected log message '%s' not found in output", msg)
		}
	}
	
	// Verify trace level logging
	if !strings.Contains(logOutput, "trace") {
		t.Error("expected trace level logging not found")
	}
	
	// Verify path information
	if !strings.Contains(logOutput, "Nested") {
		t.Error("expected field path 'Nested' not found in logs")
	}
}

// Test logging with different log levels
func TestLogLevels(t *testing.T) {
	testCases := []struct {
		name     string
		level    zerolog.Level
		shouldLog bool
	}{
		{"Trace", zerolog.TraceLevel, true},
		{"Debug", zerolog.DebugLevel, false},
		{"Info", zerolog.InfoLevel, false},
		{"Error", zerolog.ErrorLevel, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := zerolog.New(&buf).With().Timestamp().Logger().Level(tc.level)
			
			component := &LoggingComponent{
				Name: "test",
			}
			
			ctx := context.Background()
			options := &Options{
				Logger: &logger,
			}
			
			err := AutoInitWithOptions(ctx, component, options)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			logOutput := buf.String()
			hasTraceLog := strings.Contains(logOutput, "Traversing field")
			
			if tc.shouldLog && !hasTraceLog {
				t.Errorf("expected trace logs for level %s, but none found", tc.name)
			} else if !tc.shouldLog && hasTraceLog {
				t.Errorf("unexpected trace logs for level %s", tc.name)
			}
		})
	}
}

// Test logging with complex nested structures
func TestLoggingComplexStructure(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger().Level(zerolog.TraceLevel)
	
	// Use existing ComplexApp from tests
	app := &ComplexApp{
		Services: []Service{
			{Name: "service1", Database: &Database{}},
		},
		Cache: &CacheLayer{},
		Plugins: map[string]Plugin{
			"plugin1": {},
		},
	}
	
	ctx := context.Background()
	options := &Options{
		Logger: &logger,
	}
	
	err := AutoInitWithOptions(ctx, app, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	logOutput := buf.String()
	
	// Check for complex path logging
	expectedPaths := []string{
		"Services.[0]",
		"Services.[0].Database",
		"Cache",
		"Plugins.[plugin1]",
	}
	
	for _, path := range expectedPaths {
		if !strings.Contains(logOutput, path) {
			t.Errorf("expected path '%s' not found in logs", path)
		}
	}
}

// Test logging with nil logger in options
func TestNilLoggerInOptions(t *testing.T) {
	component := &LoggingComponent{
		Name: "test",
	}
	
	ctx := context.Background()
	options := &Options{
		Logger: nil, // Explicitly nil logger should use default
	}
	
	err := AutoInitWithOptions(ctx, component, options)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !component.Initialized {
		t.Error("component was not initialized")
	}
}

// Test path formatting
func TestPathFormatting(t *testing.T) {
	testCases := []struct {
		path     []string
		expected string
	}{
		{[]string{}, "<root>"},
		{[]string{"Field"}, "Field"},
		{[]string{"Parent", "Child"}, "Parent.Child"},
		{[]string{"Array", "[0]"}, "Array.[0]"},
		{[]string{"Map", "[key]"}, "Map.[key]"},
		{[]string{"A", "B", "C"}, "A.B.C"},
	}
	
	for _, tc := range testCases {
		result := pathToString(tc.path)
		if result != tc.expected {
			t.Errorf("pathToString(%v) = %s; want %s", tc.path, result, tc.expected)
		}
	}
}