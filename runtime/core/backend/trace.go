package backend

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"go.opentelemetry.io/otel/trace"
)

type Tracer interface {
	GetTracerProvider(ctx context.Context) (trace.TracerProvider, error)
}

type traceCtx struct {
	TraceID    string
	SpanID     string
	TraceFlags string
	TraceState string
	Remote     bool
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
