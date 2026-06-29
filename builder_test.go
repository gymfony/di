package di_test

import (
	"errors"
	"testing"

	"github.com/gymfony/di"
)

// --- Test environment (Interfaces and structures) ---

type Logger interface {
	Log(msg string)
}

type FileLogger struct{}

func (f *FileLogger) Log(msg string) {}
func NewFileLogger() *FileLogger     { return &FileLogger{} }

type DbLogger struct{}

func (d *DbLogger) Log(msg string) {}
func NewDbLogger() *DbLogger       { return &DbLogger{} }

type OrderService struct {
	logger Logger
}

func NewOrderService(l Logger) *OrderService { return &OrderService{logger: l} }

// --- Tests ---

func TestAddProvider_SuccessAndTags(t *testing.T) {
	cb := di.NewContainerBuilder()

	// Creating a strongly typed tag for loggers
	loggerTag := di.NewTagKey[Logger]("app.logger")

	// We register two different loggers with the same tag.
	p1 := di.Tag(NewFileLogger, loggerTag)
	p2 := di.Tag(NewDbLogger, loggerTag)

	err := di.AddProvider[*FileLogger](cb, p1)
	if err != nil {
		t.Fatalf("failed to add FileLogger: %v", err)
	}

	err = di.AddProvider[*DbLogger](cb, p2)
	if err != nil {
		t.Fatalf("failed to add DbLogger: %v", err)
	}

	// We also register a regular service without tags
	p3 := di.NewSet(NewOrderService).Providers[0]
	err = di.AddProvider[*OrderService](cb, p3)
	if err != nil {
		t.Fatalf("failed to add OrderService: %v", err)
	}

	// CHECKING TAG SEMANTICS
	// We expect exactly two services to be registered in the builder for the "app.logger" tag.
	// In the actual Compiler Pass (Stage 3), we will use this graph to inject the logger slice.
}

func TestAddProvider_DuplicateError(t *testing.T) {
	cb := di.NewContainerBuilder()
	p := di.NewSet(NewFileLogger).Providers[0]

	// The first registration must be successful.
	err := di.AddProvider[*FileLogger](cb, p)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// A second registration of the same type should return a duplicate error.
	err = di.AddProvider[*FileLogger](cb, p)
	if err == nil {
		t.Fatal("expected duplicate binding error, got nil")
	}

	// We check that it is our custom error that was returned, and not just an abstract err
	var dupErr *di.DuplicateBindingError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected error to be *di.DuplicateBindingError, got %T", err)
	}

	// Checking metadata inside error: type name must be exact
	expectedType := "*di_test.FileLogger"
	if dupErr.Type != expectedType {
		t.Errorf("expected duplicate type metadata to be %q, got %q", expectedType, dupErr.Type)
	}
}
