package autoinit_test

import (
	"context"
	"testing"

	"github.com/user/autoinit"
)

// Components for testing initialization order
type OrderLogger struct {
	Name        string
	Initialized bool
	foundSibling bool
}

func (l *OrderLogger) Init(ctx context.Context, parent interface{}) error {
	// Try to find a sibling that should be initialized after this component
	finder := autoinit.NewComponentFinder(ctx, l, parent)
	
	if sibling := finder.Find(autoinit.SearchOption{
		ByFieldName: "LaterComponent",
	}); sibling != nil {
		if laterComp, ok := sibling.(*LaterComponent); ok {
			// This should not happen if initialization order is correct
			if laterComp.Initialized {
				l.foundSibling = true
			}
		}
	}
	
	l.Initialized = true
	return nil
}

type LaterComponent struct {
	Name        string  
	Initialized bool
}

func (l *LaterComponent) Init(ctx context.Context) error {
	l.Initialized = true
	return nil
}

// Test that finder only finds already-initialized components
func TestFinderInitializationOrder(t *testing.T) {
	// Since autoinit processes fields in declaration order,
	// OrderLogger should be initialized before LaterComponent
	type App struct {
		Logger         *OrderLogger     // Initialized first
		LaterComponent *LaterComponent  // Initialized second
	}
	
	app := &App{
		Logger:         &OrderLogger{Name: "Logger"},
		LaterComponent: &LaterComponent{Name: "Later"},
	}
	
	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}
	
	// OrderLogger should not have found LaterComponent as initialized
	// since LaterComponent is initialized after OrderLogger
	if app.Logger.foundSibling {
		t.Error("OrderLogger should not have found LaterComponent as initialized (wrong order)")
	}
	
	// Both should be initialized by the end
	if !app.Logger.Initialized {
		t.Error("Logger should be initialized")
	}
	
	if !app.LaterComponent.Initialized {
		t.Error("LaterComponent should be initialized")
	}
}

// Test with early vs late component lookup
type EarlyComponent struct {
	Name             string
	Initialized      bool
	foundLate        bool
	lateWasInitialized bool
}

func (e *EarlyComponent) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, e, parent)
	
	// Look for LateComponent
	if late := finder.Find(autoinit.SearchOption{
		ByFieldName: "LateComponent", 
	}); late != nil {
		e.foundLate = true
		if lateComp, ok := late.(*LateComponent); ok {
			e.lateWasInitialized = lateComp.Initialized
		}
	}
	
	e.Initialized = true
	return nil
}

type LateComponent struct {
	Name             string
	Initialized      bool
	foundEarly       bool
	earlyWasInitialized bool
}

func (l *LateComponent) Init(ctx context.Context, parent interface{}) error {
	finder := autoinit.NewComponentFinder(ctx, l, parent)
	
	// Look for EarlyComponent (should find it and it should be initialized)
	if early := finder.Find(autoinit.SearchOption{
		ByFieldName: "EarlyComponent",
	}); early != nil {
		l.foundEarly = true
		if earlyComp, ok := early.(*EarlyComponent); ok {
			l.earlyWasInitialized = earlyComp.Initialized
		}
	}
	
	l.Initialized = true
	return nil
}

func TestFinderInitializationOrderBidirectional(t *testing.T) {
	type App struct {
		EarlyComponent *EarlyComponent  // Initialized first
		LateComponent  *LateComponent   // Initialized second  
	}
	
	app := &App{
		EarlyComponent: &EarlyComponent{Name: "Early"},
		LateComponent:  &LateComponent{Name: "Late"},
	}
	
	ctx := autoinit.WithComponentSearch(context.Background())
	if err := autoinit.AutoInit(ctx, app); err != nil {
		t.Fatalf("AutoInit failed: %v", err)
	}
	
	// Early component should find Late component but it should not be initialized yet
	if app.EarlyComponent.foundLate {
		if app.EarlyComponent.lateWasInitialized {
			t.Error("Early component found Late component as initialized, but it should not be yet")
		}
	}
	
	// Late component should find Early component and it should be initialized
	if !app.LateComponent.foundEarly {
		t.Error("Late component should have found Early component")
	} else if !app.LateComponent.earlyWasInitialized {
		t.Error("When Late component found Early component, Early should have been initialized")
	}
}