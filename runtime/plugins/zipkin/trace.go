package zipkin

import (
	"context"

	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type ZipkinTracer struct {
	tp *tracesdk.TracerProvider
}

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

func (t *ZipkinTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
