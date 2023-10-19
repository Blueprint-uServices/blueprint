package zipkin

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type ZipkinTracer struct {
	tp *tracesdk.TracerProvider
}

func NewZipkinTracer(addr string, port string) *ZipkinTracer {
	exp, err := zipkin.New("http://" + addr + ":" + port + "/api/v2/spans")
	if err != nil {
		log.Fatal(err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
	)
	return &ZipkinTracer{tp}
}

func (t *ZipkinTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
