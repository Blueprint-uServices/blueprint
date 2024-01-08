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

type LogOptions struct {
	Level Priority
}

// Represents a logger that can be used by the logger plugin
type Logger interface {
	// Logf creates a new log record at `INFO` level with `fmt.Sprintf(format, args...)` as the log message. Same interface as fmt.Printf or log.Printf.
	// Returns a context that may-be updated by the logger with some logger specific state. If no state is set, then the passed-in context is returned as is.
	Logf(ctx context.Context, opts LogOptions, format string, args ...any) (context.Context, error)
	// Debug creates a new log record at `DEBUG` level with `fmt.Sprintf(format, args...)` as the log message. Convenience wrapper around Logf
	Debug(ctx context.Context, format string, args ...any) (context.Context, error)
	// Info creates a new log record at `INFO` level with `fmt.Sprintf(format, args...)` as the log message.
	Info(ctx context.Context, format string, args ...any) (context.Context, error)
	// Warn creates a new log record at `WARN` level with `fmt.Sprintf(format, args...)` as the log message.
	Warn(ctx context.Context, format string, args ...any) (context.Context, error)
	// Error creates a new log record at `ERROR` level with `fmt.Sprintf(format, args...)` as the log message.
	Error(ctx context.Context, format string, args ...any) (context.Context, error)
}

var logger Logger

// Blueprint's error out logger. This should never be used.
type errorOutLogger struct{}

func (l *errorOutLogger) Logf(ctx context.Context, opts LogOptions, format string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error) {
	log.Fatal("ERROR: Use of errorOutLogger detected")
	// Unreachable
	return ctx, nil
}

func (l *errorOutLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error) {
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
