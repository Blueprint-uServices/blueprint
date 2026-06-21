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

func NewStdoutMetricCollector(ctx context.Context, duration string) (*StdoutMetricCollector, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	if duration == "" {
		duration = "1s"
	}

	timerDuration, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}

	mp := metricsdk.NewMeterProvider(
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exp,
			// Default is 1m. Set to 1s for demonstrative purposes.
			metricsdk.WithInterval(timerDuration))),
	)

	otel.SetMeterProvider(mp)
	mc := &StdoutMetricCollector{mp}
	backend.SetDefaultMetricCollector(mc)
	return mc, nil
}
