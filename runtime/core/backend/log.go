package backend

import (
	"context"
	"log"
)

// The Priority Level at which the message will be recorded
type Priority int

const (
	DEBUG Priority = iota
	INFO
	WARN
	ERROR
)

// String representation for Priority enum
func (p Priority) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[p]
}

// Represents a Key-Value pair associated with a particular log msg
type Attribute struct {
	Key   string
	Value interface{}
}

// Represents a logger that can be used by the logger plugin
type Logger interface {
	// LogWithAttrs creates a new log record at the desired `priority` level with `msg` as the log message and `attr` as list of optional key-value pairs for logging messages.
	// Returns a context that may-be updated by the logger with some logger specific state. If no state is set, then the passed-in context is returned as is.
	LogWithAttrs(ctx context.Context, priority Priority, msg string, attr ...Attribute) (context.Context, error)
	// Debug creates a new log record at `DEBUG` level with `msg` as the log message and `args` as key-value pairs. Same interface as slog's Debug.
	Debug(ctx context.Context, msg string, args ...any) (context.Context, error)
	// Info creates a new log record at `INFO` level with `msg` as the log message and `args` as key-value pairs. Same interface as slog's Info.
	Info(ctx context.Context, msg string, args ...any) (context.Context, error)
	// Warn creates a new log record at `WARN` level with `msg` as the log message and `args` as key-value pairs. Same interface as slog's Warn.
	Warn(ctx context.Context, msg string, args ...any) (context.Context, error)
	// Error creates a new log record at `ERROR` level with `msg` as the log message and `args` as key-value pairs. Same interface as slog's Error.
	Error(ctx context.Context, msg string, args ...any) (context.Context, error)
	// Logf creates a new log record at `INFO` level with the log message constructed from format and args. Same interface as fmt.Printf or log.Printf.
	Logf(ctx context.Context, format string, args ...any) (context.Context, error)
}

var logger Logger

// Blueprint's error out logger. This should never be used.
type errorOutLogger struct{}

func (l *errorOutLogger) LogWithAttrs(ctx context.Context, priority Priority, msg string, attr ...Attribute) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Debug(ctx context.Context, msg string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Info(ctx context.Context, msg string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Warn(ctx context.Context, msg string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Error(ctx context.Context, msg string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Logf(ctx context.Context, format string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

// Set's the default logger to be used by the Blueprint application.
// NOTE: This should not be called in the workflow code. This is called from the various logger plugins.
func SetDefaultLogger(l Logger) {
	logger = l
}

// Returns the default logger
func GetLogger() Logger {
	return logger
}

func init() {
	logger = &errorOutLogger{}
}
