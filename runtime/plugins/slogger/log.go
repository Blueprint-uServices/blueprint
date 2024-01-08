package slogger

import (
	"context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"golang.org/x/exp/slog"
)

// SLogger implements the backend.Logger interface by wrapping calls to golang's slog package.
type SLogger struct{}

// Implements backend.Logger
func (l *SLogger) LogWithAttrs(ctx context.Context, priority backend.Priority, msg string, attrs ...backend.Attribute) (context.Context, error) {
	var args []slog.Attr
	for _, attr := range attrs {
		args = append(args, slog.Attr{Key: attr.Key, Value: slog.AnyValue(attr.Value)})
	}
	slog.LogAttrs(ctx, slog.Level(priority), msg, args...)
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Debug(ctx context.Context, msg string, args ...any) (context.Context, error) {
	slog.DebugContext(ctx, msg, args...)
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Info(ctx context.Context, msg string, args ...any) (context.Context, error) {
	slog.InfoContext(ctx, msg, args...)
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Warn(ctx context.Context, msg string, args ...any) (context.Context, error) {
	slog.WarnContext(ctx, msg, args...)
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Error(ctx context.Context, msg string, args ...any) (context.Context, error) {
	slog.ErrorContext(ctx, msg, args...)
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Logf(ctx context.Context, format string, args ...any) (context.Context, error) {
	msg := fmt.Sprintf(format, args...)
	slog.Log(ctx, slog.LevelInfo.Level(), msg)
	return ctx, nil
}

// Returns a new logger object
func NewSLogger(ctx context.Context) (*SLogger, error) {
	return &SLogger{}, nil
}
