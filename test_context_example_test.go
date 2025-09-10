package autoinit

import (
	"context"
	"reflect"
	"testing"
)

// Example interfaces for testing
type TestLogger interface {
	Log(message string)
}

type TestDB interface {
	Query(sql string) []string
}

type TestCacheService interface {
	Get(key string) string
	Set(key string, value string)
}

// Mock implementations
type MockTestLogger struct {
	messages []string
}

func (m *MockTestLogger) Log(message string) {
	m.messages = append(m.messages, message)
}

type MockTestDB struct {
	data map[string][]string
}

func (m *MockTestDB) Query(sql string) []string {
	return m.data[sql]
}

type MockTestCache struct {
	data map[string]string
}

func (m *MockTestCache) Get(key string) string {
	return m.data[key]
}

func (m *MockTestCache) Set(key string, value string) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
}

// Service that uses dependency discovery
type TestUserService struct {
	Logger TestLogger       `autoinit:""`
	DB     TestDB           `autoinit:""`
	Cache  TestCacheService `autoinit:""`
}

func (s *TestUserService) Init(ctx context.Context, parent interface{}) error {
	// Use As pattern to discover dependencies
	if !As(ctx, s, parent, &s.Logger) {
		return ErrComponentNotFound
	}
	if !As(ctx, s, parent, &s.DB) {
		return ErrComponentNotFound
	}
	if !As(ctx, s, parent, &s.Cache) {
		return ErrComponentNotFound
	}
	return nil
}

func (s *TestUserService) GetUser(id string) string {
	// Check cache first
	if cached := s.Cache.Get("user:" + id); cached != "" {
		s.Logger.Log("Cache hit for user " + id)
		return cached
	}

	// Query database
	results := s.DB.Query("SELECT name FROM users WHERE id = " + id)
	if len(results) > 0 {
		user := results[0]
		s.Cache.Set("user:"+id, user)
		s.Logger.Log("Loaded user " + id + " from database")
		return user
	}

	s.Logger.Log("User " + id + " not found")
	return ""
}

// Example 1: Testing with TestContext - Basic Usage
func TestUserService_WithTestContext(t *testing.T) {
	// Create mock dependencies
	mockLogger := &MockTestLogger{}
	mockDB := &MockTestDB{
		data: map[string][]string{
			"SELECT name FROM users WHERE id = 123": {"John Doe"},
		},
	}
	mockCache := &MockTestCache{}

	// Create test context with dependencies registered by interface types
	testCtx := NewTestContext().
		RegisterInterface(reflect.TypeOf((*TestLogger)(nil)).Elem(), mockLogger).
		RegisterInterface(reflect.TypeOf((*TestDB)(nil)).Elem(), mockDB).
		RegisterInterface(reflect.TypeOf((*TestCacheService)(nil)).Elem(), mockCache)

	// Create service and initialize with test context
	service := &TestUserService{}
	ctx := testCtx.Context()

	// Initialize the service using autoinit
	err := AutoInit(ctx, service)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Test the service
	user := service.GetUser("123")
	if user != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", user)
	}

	// Verify interactions
	if len(mockLogger.messages) == 0 {
		t.Error("Expected logger to be called")
	}
}

// Example 2: Testing with TestBuilder - Fluent Interface
func TestUserService_WithTestBuilder(t *testing.T) {
	// Create test context using builder pattern
	ctx := NewTestBuilder().
		WithInterfaceDependency(reflect.TypeOf((*TestLogger)(nil)).Elem(), &MockTestLogger{}).
		WithInterfaceDependency(reflect.TypeOf((*TestDB)(nil)).Elem(), &MockTestDB{
			data: map[string][]string{
				"SELECT name FROM users WHERE id = 456": {"Jane Smith"},
			},
		}).
		WithInterfaceDependency(reflect.TypeOf((*TestCacheService)(nil)).Elem(), &MockTestCache{}).
		Context()

	// Create and initialize service
	service := &TestUserService{}
	err := AutoInit(ctx, service)
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Test the service
	user := service.GetUser("456")
	if user != "Jane Smith" {
		t.Errorf("Expected 'Jane Smith', got '%s'", user)
	}
}

// Example 3: Testing individual dependency discovery
func TestDependencyDiscovery_Direct(t *testing.T) {
	// Create test context
	mockLogger := &MockTestLogger{}
	testCtx := NewTestContext().RegisterInterface(reflect.TypeOf((*TestLogger)(nil)).Elem(), mockLogger)
	ctx := testCtx.Context()

	// Test direct dependency discovery
	var logger TestLogger
	err := TestAs(ctx, nil, &logger)
	if err != nil {
		t.Fatalf("Failed to discover logger: %v", err)
	}

	// Test the discovered dependency
	logger.Log("test message")
	if len(mockLogger.messages) != 1 || mockLogger.messages[0] != "test message" {
		t.Error("Logger not working correctly")
	}
}

// Example 4: Testing with MustAs
func TestMustAs_Success(t *testing.T) {
	mockLogger := &MockTestLogger{}
	testCtx := NewTestContext().RegisterInterface(reflect.TypeOf((*TestLogger)(nil)).Elem(), mockLogger)
	ctx := testCtx.Context()

	var logger TestLogger
	// This should not panic
	TestMustAs(ctx, nil, &logger)

	logger.Log("test")
	if len(mockLogger.messages) != 1 {
		t.Error("MustAs did not work correctly")
	}
}

func TestMustAs_Panic(t *testing.T) {
	testCtx := NewTestContext() // Empty context
	ctx := testCtx.Context()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustAs to panic when dependency not found")
		}
	}()

	var logger TestLogger
	TestMustAs(ctx, nil, &logger) // Should panic
}
