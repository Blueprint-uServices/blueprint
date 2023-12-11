// Package zipkin implements a tracer [backend.Tracer] client interface for the zipkin tracer.
package zipkin

import (
	"context"

	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ZipkinTracer implements the runtime backend instance that implements the backend/trace.Tracer interface.
// REQUIRED: A functional backend running the zipkin collector.
type ZipkinTracer struct {
	tp *tracesdk.TracerProvider
}

// Returns a new instance of ZipkinTracer.
// Configures opentelemetry to export zipkin traces to the zipkin collector hosted at address `addr`.
func NewZipkinTracer(ctx context.Context, addr string) (*ZipkinTracer, error) {
	exp, err := zipkin.New("http://" + addr + "/api/v2/spans")
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
	)
	return &ZipkinTracer{tp}, nil
}

// Implements the backend/trace interface.
func (t *ZipkinTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
