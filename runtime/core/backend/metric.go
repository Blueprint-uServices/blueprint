package backend

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"golang.org/x/exp/slog"
)

// Represents a metric collector that can be used by the metric/opentelemetry plugin
type MetricCollector interface {
	// Returns a go.opentelemetry.io/otel/metric/MeterProvider
	GetMetricProvider(ctx context.Context) (metric.MeterProvider, error)
}

var metric_collector MetricCollector

type stdoutMetricCollector struct {
	mp *metricsdk.MeterProvider
}

func (s *stdoutMetricCollector) GetMetricProvider(ctx context.Context) (metric.MeterProvider, error) {
	return s.mp, nil
}

func newstdoutMetricCollector(ctx context.Context) (*stdoutMetricCollector, error) {
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
	return &stdoutMetricCollector{mp}, nil
}

func init() {
	coll, err := newstdoutMetricCollector(context.Background())
	if err != nil {
		slog.Error(err.Error())
	}
	metric_collector = coll
}

// Sets the default metric collector to be used by BLueprint applications.
// This should be called from the constructor of a Metric Collector
func SetDefaultMetricCollector(m MetricCollector) {
	metric_collector = m
}

// Meter returns a new metric.Meter with a provided name and configuration
//
// A Meter should be scoped at most to a single package. We recommend a meter being scoped to a single service.
// The name needs to be unique so it does not collide with other names used by
// an application, nor other applications.
//
// If the name is empty, then an implementation defined default name will
// be used instead.
func Meter(ctx context.Context, name string, opts ...metric.MeterOption) (metric.Meter, error) {
	mp, err := metric_collector.GetMetricProvider(ctx)
	if err != nil {
		return nil, err
	}
	return mp.Meter(name, opts...), nil
}
