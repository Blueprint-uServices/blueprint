package opentelemetry

import (
	"context"
	"time"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type StdoutMetricCollector struct {
	mp *metricsdk.MeterProvider
}

func (s *StdoutMetricCollector) GetMetricProvider(ctx context.Context) (metric.MeterProvider, error) {
	return s.mp, nil
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
	mc := &StdoutMetricCollector{mp}
	backend.SetDefaultMetricCollector(mc)
	return mc, nil
}
