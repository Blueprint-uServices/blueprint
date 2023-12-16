package opentelemetry

import (
	"context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
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

func NewOTTraceLogger(ctx context.Context) (*OTTraceLogger, error) {
	l := &OTTraceLogger{}
	backend.SetDefaultLogger(l)
	return l, nil
}

func (l *OTTraceLogger) Log(ctx context.Context, priority backend.Priority, msg string, attrs ...backend.Attribute) context.Context {
	span := trace.SpanFromContext(ctx)
	var all_attributes []attribute.KeyValue
	all_attributes = append(all_attributes, attribute.String("Priority", priority.String()))
	for _, a := range attrs {
		all_attributes = append(all_attributes, attribute.String(a.Key, fmt.Sprintf("%v", a.Value)))
	}
	span.AddEvent(msg, trace.WithAttributes(all_attributes...))
	return ctx
}
