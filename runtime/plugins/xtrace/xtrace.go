// Package xtrace provides xtrace-based runtime components to be used by blueprint application workflows and blueprint generated code.
// The package provides the following runtime components:
// (i)  XTracerImpl: a client-wrapper implementation of the [XTracer] interface to a xtrace server. Used by the xtrace plugin for providing context propagation between multiple processes.
// (ii) XTraceLogger: an xtrace-based logger implementation of the [Logger] interface. Once initialized, the logger sets itself as the default logger for logging across blueprint applications.
package xtrace

import (
	"context"

	"github.com/tracingplane/tracingplane-go/tracingplane"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/localbaggage"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
)

// Implementation of the [XTracer] interface
type XTracerImpl struct {
	XTracer
}

// Returns a new instance of [XTracerImpl] that connects to a xtrace server running at `addr`.
// REQUIRED: An xtrace server must be running at `addr`
func NewXTracerImpl(ctx context.Context, addr string) (*XTracerImpl, error) {
	err := connectClient(addr)
	if err != nil {
		return nil, err
	}
	return &XTracerImpl{}, nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) Log(ctx context.Context, msg string) (context.Context, error) {
	return client.Log(ctx, msg), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error) {
	return client.LogWithTags(ctx, msg, tags...), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) StartTask(ctx context.Context, tags ...string) (context.Context, error) {
	return client.StartTask(ctx, tags...), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) StopTask(ctx context.Context) (context.Context, error) {
	return client.StopTask(ctx), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Merge(ctx, other), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Set(ctx, baggage), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) Get(ctx context.Context) (tracingplane.BaggageContext, error) {
	return localbaggage.Get(ctx), nil
}

// Implements the [XTracer] interface
func (xt *XTracerImpl) IsTracing(ctx context.Context) (bool, error) {
	return client.HasTask(ctx), nil
}
