package jaeger

import (
	"context"
	"log"

	jaeger_exporter "go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type JaegerTracer struct {
	tp *tracesdk.TracerProvider
}

func NewJaegerTracer(addr string, port string) *JaegerTracer {
	exp, err := jaeger_exporter.New(jaeger_exporter.WithCollectorEndpoint(jaeger_exporter.WithEndpoint("http://" + addr + ":" + port + "/api/traces")))
	if err != nil {
		log.Fatal(err)
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
	)
	return &JaegerTracer{tp}
}

func (t *JaegerTracer) GetTracerProvider(ctx context.Context) (trace.TracerProvider, error) {
	return t.tp, nil
}
