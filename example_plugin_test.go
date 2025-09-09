package autoinit_test

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/user/autoinit"
)

// PluginComponent represents a pluggable component
type PluginComponent struct {
	Name    string
	Enabled bool
}

func (p *PluginComponent) Init(ctx context.Context) error {
	p.Enabled = true
	fmt.Printf("Plugin %s activated\n", p.Name)
	return nil
}

// Application with pluggable components
type PluggableApp struct {
	// Core components - always present
	Core *CoreComponent

	// Pluggable components - just add them to enable features!
	Auth      *AuthComponent      // Add authentication
	Cache     *CacheComponent     // Add caching
	Metrics   *MetricsComponent   // Add metrics collection
	RateLimit *RateLimitComponent // Add rate limiting
}

type CoreComponent struct {
	Started bool
}

func (c *CoreComponent) Init(ctx context.Context) error {
	c.Started = true
	fmt.Println("Core system started")
	return nil
}

type AuthComponent struct {
	Active bool
}

func (a *AuthComponent) Init(ctx context.Context) error {
	a.Active = true
	fmt.Println("Authentication enabled")
	return nil
}

type MetricsComponent struct {
	Collecting bool
}

func (m *MetricsComponent) Init(ctx context.Context) error {
	m.Collecting = true
	fmt.Println("Metrics collection started")
	return nil
}

type CacheComponent struct {
	Ready bool
}

func (c *CacheComponent) Init(ctx context.Context) error {
	c.Ready = true
	fmt.Println("Cache initialized")
	return nil
}

type RateLimitComponent struct {
	Enforcing bool
}

func (r *RateLimitComponent) Init(ctx context.Context) error {
	r.Enforcing = true
	fmt.Println("Rate limiting activated")
	return nil
}

// Example_plugAndPlay demonstrates the plug-and-play nature of components.
// You can add or remove components without changing any initialization code!
func Example_plugAndPlay() {
	// Start with minimal app
	app := &PluggableApp{
		Core: &CoreComponent{},
	}

	// Plug in components as needed - no code changes required!
	app.Auth = &AuthComponent{}       // Just add authentication
	app.Metrics = &MetricsComponent{} // Just add metrics
	// app.Cache = &CacheComponent{}   // Commented out - not needed yet
	// app.RateLimit = ...              // Can add later when needed

	// One initialization call handles everything
	ctx := context.Background()
	// Use a silent logger for examples to avoid trace output
	logger := zerolog.Nop()
	options := &autoinit.Options{Logger: &logger}
	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\nAll plugged components are initialized and ready!")

	// Output:
	// Core system started
	// Authentication enabled
	// Metrics collection started
	//
	// All plugged components are initialized and ready!
}

// Example_dynamicComponents shows adding components at runtime
func Example_dynamicComponents() {
	// Start with base configuration
	app := &PluggableApp{
		Core: &CoreComponent{},
		Auth: &AuthComponent{},
	}

	// Conditionally add components based on config/environment
	if needsCaching() {
		app.Cache = &CacheComponent{}
	}

	if isProduction() {
		app.RateLimit = &RateLimitComponent{}
		app.Metrics = &MetricsComponent{}
	}

	// Single initialization point - works regardless of which components are plugged in
	ctx := context.Background()
	// Use a silent logger for examples to avoid trace output
	logger := zerolog.Nop()
	options := &autoinit.Options{Logger: &logger}
	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Application initialized with selected components")

	// Output:
	// Core system started
	// Authentication enabled
	// Cache initialized
	// Metrics collection started
	// Rate limiting activated
	// Application initialized with selected components
}

// Helper functions for the example
func needsCaching() bool {
	return true // Simulate config check
}

func isProduction() bool {
	return true // Simulate environment check
}
