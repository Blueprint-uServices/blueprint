package xtrace

import (
	"context"

	"github.com/tracingplane/tracingplane-go/tracingplane"
)

// Represents the XTrace tracer interface exposed to applications and used by the xtrace plugin.
type XTracer interface {
	// Creates an event in the current trace with `msg` as content.
	Log(ctx context.Context, msg string) (context.Context, error)
	// Creates an event in the current trace with `msg` as content along with `tags` as the keywords for the text.
	LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error)
	// Starts a new trace with given `tags`
	StartTask(ctx context.Context, tags ...string) (context.Context, error)
	// Stops the current running trace.
	StopTask(ctx context.Context) (context.Context, error)
	// Merges the incoming `other` baggage into the current trace's existing baggage.
	Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error)
	// Sets the current trace's `baggage` to the one provided.
	Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error)
	// Returns the current trace's baggage.
	Get(ctx context.Context) (tracingplane.BaggageContext, error)
	// Returns true if there is an ongoing trace.
	IsTracing(ctx context.Context) (bool, error)
}
