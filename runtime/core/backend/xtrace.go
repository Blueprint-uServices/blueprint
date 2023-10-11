package backend

import (
	"context"

	"github.com/tracingplane/tracingplane-go/tracingplane"
)

type XTracer interface {
	Log(ctx context.Context, msg string) (context.Context, error)
	LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error)
	StartTask(ctx context.Context, tags ...string) (context.Context, error)
	StopTask(ctx context.Context) (context.Context, error)
	Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error)
	Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error)
	Get(ctx context.Context) (tracingplane.BaggageContext, error)
	IsTracing(ctx context.Context) (bool, error)
}
