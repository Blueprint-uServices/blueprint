package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type StdoutTracer struct {
	tp *tracesdk.TracerProvider
}

func NewStdoutTracer(ctx context.Context, addr string) (*StdoutTracer, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	bsp := tracesdk.NewBatchSpanProcessor(exp)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(bsp),
	)
	return &StdoutTracer{tp}, nil
}

func (t *StdoutTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
