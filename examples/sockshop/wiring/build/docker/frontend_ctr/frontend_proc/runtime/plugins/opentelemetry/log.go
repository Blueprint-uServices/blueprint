package opentelemetry

import (
	"context"
	"fmt"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Implementation of the [backend.Logger] interface for backend.Tracer
// This logger converts each log statement into an event which is added to a current span.
// Note: This logger should only be used in conjunction with a backend.Tracer. Using this logger without using a backend.Tracer would result in no-op logging behavior.
// Note: This implementation will not be the same as a future OpenTelemetry.Logger which is in beta-testing for select languages (not including Go).
type OTTraceLogger struct {
	backend.Logger
}

// Returns a new OTTraceLogger object
func NewOTTraceLogger(ctx context.Context) (*OTTraceLogger, error) {
	l := &OTTraceLogger{}
	backend.SetDefaultLogger(l)
	return l, nil
}

// Implements backend.Logger
func (l *OTTraceLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error) {
	span := trace.SpanFromContext(ctx)
	all_attributes := []attribute.KeyValue{}
	all_attributes = append(all_attributes, attribute.String("Priority", backend.DEBUG.String()))
	msg := fmt.Sprintf(format, args...)

	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx, nil
}

// Implements backend.Logger
func (l *OTTraceLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error) {
	span := trace.SpanFromContext(ctx)
	all_attributes := []attribute.KeyValue{}
	all_attributes = append(all_attributes, attribute.String("Priority", backend.INFO.String()))
	msg := fmt.Sprintf(format, args...)

	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx, nil
}

// Implements backend.Logger
func (l *OTTraceLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error) {
	span := trace.SpanFromContext(ctx)
	all_attributes := []attribute.KeyValue{}
	all_attributes = append(all_attributes, attribute.String("Priority", backend.WARN.String()))
	msg := fmt.Sprintf(format, args...)

	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx, nil
}

// Implements backend.Logger
func (l *OTTraceLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error) {
	span := trace.SpanFromContext(ctx)
	all_attributes := []attribute.KeyValue{}
	all_attributes = append(all_attributes, attribute.String("Priority", backend.ERROR.String()))
	msg := fmt.Sprintf(format, args...)

	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx, nil
}

// Implements backend.Logger
func (l *OTTraceLogger) Logf(ctx context.Context, opts backend.LogOptions, format string, args ...any) (context.Context, error) {
	msg := fmt.Sprintf(format, args...)
	span := trace.SpanFromContext(ctx)
	all_attributes := []attribute.KeyValue{}
	all_attributes = append(all_attributes, attribute.String("Priority", opts.Level.String()))
	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx, nil
}
