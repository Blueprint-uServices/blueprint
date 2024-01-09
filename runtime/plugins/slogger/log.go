package slogger

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"golang.org/x/exp/slog"
)

// SLogger implements the backend.Logger interface by wrapping calls to golang's slog package.
type SLogger struct{}

// Implements backend.Logger
func (l *SLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error) {
	slog.DebugContext(ctx, fmt.Sprintf(format, args...))
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error) {
	slog.InfoContext(ctx, fmt.Sprintf(format, args...))
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error) {
	slog.WarnContext(ctx, fmt.Sprintf(format, args...))
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error) {
	slog.ErrorContext(ctx, fmt.Sprintf(format, args...))
	return ctx, nil
}

// Implements backend.Logger
func (l *SLogger) Logf(ctx context.Context, opts backend.LogOptions, format string, args ...any) (context.Context, error) {
	msg := fmt.Sprintf(format, args...)
	slog.Log(ctx, slog.Level(opts.Level), msg)
	return ctx, nil
}

// Returns a new logger object
func NewSLogger(ctx context.Context) (*SLogger, error) {
	l := &SLogger{}
	backend.SetDefaultLogger(l)
	return l, nil
}
