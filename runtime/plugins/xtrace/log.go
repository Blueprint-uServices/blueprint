package xtrace

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
)

// Implementation of the [backend.Logger] interface
// Note: This logger should only be used in conjunction with the XTracerImpl tracer. Using this logger without using the XTracerImpl tracer would result in no-op logging behavior.
type XTraceLogger struct {
	backend.Logger
}

var isConnected bool

func connectClient(addr string) error {
	if !isConnected {
		err := client.Connect(addr)
		if err != nil {
			return err
		}
		isConnected = true
	}
	return nil
}

// Returns a new instance of [XTracerImpl] that connects to a xtrace server running at `addr`.
// REQUIRED: An xtrace server must be running at `addr`
func NewXTraceLogger(ctx context.Context, addr string) (*XTraceLogger, error) {
	err := connectClient(addr)
	if err != nil {
		return nil, err
	}
	l := &XTraceLogger{}
	backend.SetDefaultLogger(l)
	return l, nil
}

// Implements backend.Logger
func (l *XTraceLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error) {
	if !client.HasTask(ctx) {
		// Only do logging if there is an active task running
		return ctx, nil
	}
	msg := "INFO: " + format
	return client.Logf(ctx, msg, args...), nil
}

// Implements backend.Logger
func (l *XTraceLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error) {
	if !client.HasTask(ctx) {
		return ctx, nil
	}
	msg := "DEBUG: " + format
	return client.Logf(ctx, msg, args...), nil
}

// Implements backend.Logger
func (l *XTraceLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error) {
	if !client.HasTask(ctx) {
		return ctx, nil
	}
	msg := "WARN: " + format
	return client.Logf(ctx, msg, args...), nil
}

// Implements backend.Logger
func (l *XTraceLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error) {
	if !client.HasTask(ctx) {
		return ctx, nil
	}
	msg := "ERROR: " + fmt.Sprintf(format, args...)
	return client.LogWithTags(ctx, msg, "Error"), nil
}

// Implements backend.Logger
func (l *XTraceLogger) Logf(ctx context.Context, opts backend.LogOptions, format string, args ...any) (context.Context, error) {
	if !client.HasTask(ctx) {
		return ctx, nil
	}
	format = opts.Level.String() + ": " + format
	return client.Logf(ctx, format, args...), nil
}
