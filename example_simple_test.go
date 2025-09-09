package autoinit_test

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/user/autoinit"
)

// SimpleLogger with Init method
type SimpleLogger struct {
	Enabled bool
	Level   string
}

func (l *SimpleLogger) Init(ctx context.Context) error {
	l.Enabled = true
	l.Level = "INFO"
	fmt.Println("Logger initialized")
	return nil
}

// SimpleDatabase with Init method
type SimpleDatabase struct {
	Connected bool
	Host      string
}

func (d *SimpleDatabase) Init(ctx context.Context) error {
	d.Connected = true
	d.Host = "localhost"
	fmt.Println("Database connected")
	return nil
}

// SimpleApplication with nested components
type SimpleApplication struct {
	Name     string
	Logger   SimpleLogger
	Database SimpleDatabase
	Ready    bool
}

func (a *SimpleApplication) Init(ctx context.Context) error {
	a.Ready = true
	fmt.Println("Application ready")
	return nil
}

// Example demonstrates basic usage of AutoInit
func ExampleSimpleApplication() {
	app := &SimpleApplication{
		Name: "MyApp",
	}

	// Initialize all components automatically with silent logger for examples
	ctx := context.Background()
	silentLogger := zerolog.New(io.Discard)
	options := &autoinit.Options{Logger: &silentLogger}
	if err := autoinit.AutoInitWithOptions(ctx, app, options); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nApp ready: %v", app.Ready)

	// Output:
	// Logger initialized
	// Database connected
	// Application ready
	//
	// App ready: true
}
