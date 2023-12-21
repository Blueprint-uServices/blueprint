package backend

import (
	"context"

	"golang.org/x/exp/slog"
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
	// Log creates a new log record at the desired `priority` level with `msg` as the log message and `attr` as list of optional key-value pairs for logging messages.
	// Returns a context that may-be updated by the logger with some logger specific state. If no state is set, then the passed-in context is returned as is.
	Log(ctx context.Context, priority Priority, msg string, attr ...Attribute) (context.Context, error)
}

var logger Logger

// Blueprint's default logger that uses the slog package
type defaultLogger struct{}

func (l *defaultLogger) Log(ctx context.Context, priority Priority, msg string, attrs ...Attribute) (context.Context, error) {
	var args []any
	for _, attr := range attrs {
		args = append(args, attr.Key)
		args = append(args, attr.Value)
	}
	switch priority {
	case DEBUG:
		slog.Debug(msg, args...)
		break
	case INFO:
		slog.Info(msg, args...)
		break
	case WARN:
		slog.Warn(msg, args...)
		break
	case ERROR:
		slog.Error(msg, args...)
	}
	return ctx, nil
}

// Set's the default logger to be used by the Blueprint application.
// NOTE: This should not be called in the workflow code. This is called from the various logger plugins.
func SetDefaultLogger(l Logger) {
	logger = l
}

// Log function exposed to Blueprint workflow applications.
// Wraps around the currently set default logger's Logger API calls.
func Log(ctx context.Context, priority Priority, msg string, attrs ...Attribute) (context.Context, error) {
	return logger.Log(ctx, priority, msg, attrs...)
}

func init() {
	logger = &defaultLogger{}
}
