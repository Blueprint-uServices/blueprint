package opentelemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type StdoutMetricCollector struct {
	mp *metricsdk.MeterProvider
}

func NewStdoutMetricCollector(ctx context.Context) (*StdoutMetricCollector, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	mp := metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exp,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metricsdk.WithInterval(3*time.Second))),
	)
	otel.SetMeterProvider(mp)
	return &StdoutMetricCollector{mp}, nil
}

func (m *StdoutMetricCollector) GetMetricProvider(ctx context.Context) (metric.MeterProvider, error) {
	return m.mp, nil
}
