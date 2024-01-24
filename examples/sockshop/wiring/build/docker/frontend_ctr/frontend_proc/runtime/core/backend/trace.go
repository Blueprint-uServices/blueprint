package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"go.opentelemetry.io/otel/trace"
)

// Represents a tracer that can be used by the tracer/opentelemetry plugin
type Tracer interface {
	// Returns a go.opentelemetry.io/otel/trace.TracerProvider
	// TracerProvider provides Tracers that are used by instrumentation code to trace computational workflows.
	GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
}

// traceCtx mimics the internal trace context object from OpenTelemetry.
// Included here to be able to implement and provide the `GetSpanContext` function.
type traceCtx struct {
	// ID of the current trace
	TraceID string
	// ID of the current span
	SpanID string
	// Additional flags for the trace
	TraceFlags string
	// Additional state for the trace
	TraceState string
	// If span is a remote span
	Remote bool
}

// Utility function to convert an encoded string into a Span Context
func GetSpanContext(encoded_string string) (trace.SpanContextConfig, error) {
	var tCtx traceCtx
	err := json.Unmarshal([]byte(encoded_string), &tCtx)
	if err != nil {
		return trace.SpanContextConfig{}, err
	}
	tid, err := trace.TraceIDFromHex(tCtx.TraceID)
	if err != nil {
		return trace.SpanContextConfig{}, err
	}
	sid, err := trace.SpanIDFromHex(tCtx.SpanID)
	if err != nil {
		return trace.SpanContextConfig{}, err
	}
	flag_bytes, err := hex.DecodeString(tCtx.TraceFlags)
	if err != nil {
		return trace.SpanContextConfig{}, err
	}
	tFlags := trace.TraceFlags(flag_bytes[0])
	tState, err := trace.ParseTraceState(tCtx.TraceState)
	if err != nil {
		return trace.SpanContextConfig{}, err
	}
	return trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: tFlags, TraceState: tState, Remote: tCtx.Remote}, nil
}
