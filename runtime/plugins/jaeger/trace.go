package jaeger

import (
	"context"

	jaeger_exporter "go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type JaegerTracer struct {
	tp *tracesdk.TracerProvider
}

func NewJaegerTracer(ctx context.Context, addr string) (*JaegerTracer, error) {
	exp, err := jaeger_exporter.New(jaeger_exporter.WithCollectorEndpoint(jaeger_exporter.WithEndpoint("http://" + addr + "/api/traces")))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
	)
	return &JaegerTracer{tp}, nil
}

func (t *JaegerTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
