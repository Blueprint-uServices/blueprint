package backend

import (
	"context"

	"github.com/tracingplane/tracingplane-go/tracingplane"
)

type XTracer interface {
	Log(ctx context.Context, msg string) context.Context
	LogWithTags(ctx context.Context, msg string, tags ...string) context.Context
	StartTask(ctx context.Context, tags ...string) context.Context
	StopTask(ctx context.Context) context.Context
	Merge(ctx context.Context, other tracingplane.BaggageContext) context.Context
	Set(ctx context.Context, baggage tracingplane.BaggageContext) context.Context
	Get(ctx context.Context) tracingplane.BaggageContext
	IsTracing(ctx context.Context) bool
}
