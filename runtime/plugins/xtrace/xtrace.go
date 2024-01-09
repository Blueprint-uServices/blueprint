// Package xtrace provides a client-wrapper implementation of the [backend.XTracer] interface to a xtrace server.
package xtrace

import (
	"context"

	"github.com/Blueprint-uServices/blueprint/runtime/core/backend"
	"github.com/tracingplane/tracingplane-go/tracingplane"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/localbaggage"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
)

// Implementation of the [backend.XTracer] interface
type XTracerImpl struct {
	backend.XTracer
}

// Returns a new instance of [XTracerImpl] that connects to a xtrace server running at `addr`.
// REQUIRED: An xtrace server must be running at `addr`
func NewXTracerImpl(ctx context.Context, addr string) (*XTracerImpl, error) {
	err := client.Connect(addr)
	if err != nil {
		return nil, err
	}
	return &XTracerImpl{}, nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) Log(ctx context.Context, msg string) (context.Context, error) {
	return client.Log(ctx, msg), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error) {
	return client.LogWithTags(ctx, msg, tags...), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) StartTask(ctx context.Context, tags ...string) (context.Context, error) {
	return client.StartTask(ctx, tags...), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) StopTask(ctx context.Context) (context.Context, error) {
	return client.StopTask(ctx), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Merge(ctx, other), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Set(ctx, baggage), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) Get(ctx context.Context) (tracingplane.BaggageContext, error) {
	return localbaggage.Get(ctx), nil
}

// Implements the [backend.XTracer] interface
func (xt *XTracerImpl) IsTracing(ctx context.Context) (bool, error) {
	return client.HasTask(ctx), nil
}
