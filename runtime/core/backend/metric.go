package backend

import (
	"context"

	"go.opentelemetry.io/otel/metric"
)

// Represents a metric collector that can be used by the metric/opentelemetry plugin
type MetricCollector interface {
	// Returns a go.opentelemetry.io/otel/metric/MeterProvider
	GetMetricProvider(ctx context.Context) (metric.MeterProvider, error)
}
