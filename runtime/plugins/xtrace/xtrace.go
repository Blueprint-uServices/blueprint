package xtrace

import (
	"context"

	"github.com/tracingplane/tracingplane-go/tracingplane"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/localbaggage"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
)

type XTracerImpl struct {
	backend.XTracer
}

func NewXTracerImpl(ctx context.Context, addr string) (*XTracerImpl, error) {
	err := client.Connect(addr)
	if err != nil {
		return nil, err
	}
	return &XTracerImpl{}, nil
}

func (xt *XTracerImpl) Log(ctx context.Context, msg string) (context.Context, error) {
	return client.Log(ctx, msg), nil
}

func (xt *XTracerImpl) LogWithTags(ctx context.Context, msg string, tags ...string) (context.Context, error) {
	return client.LogWithTags(ctx, msg, tags...), nil
}

func (xt *XTracerImpl) StartTask(ctx context.Context, tags ...string) (context.Context, error) {
	return client.StartTask(ctx, tags...), nil
}

func (xt *XTracerImpl) StopTask(ctx context.Context) (context.Context, error) {
	return client.StopTask(ctx), nil
}

func (xt *XTracerImpl) Merge(ctx context.Context, other tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Merge(ctx, other), nil
}

func (xt *XTracerImpl) Set(ctx context.Context, baggage tracingplane.BaggageContext) (context.Context, error) {
	return localbaggage.Set(ctx, baggage), nil
}

func (xt *XTracerImpl) Get(ctx context.Context) (tracingplane.BaggageContext, error) {
	return localbaggage.Get(ctx), nil
}

func (xt *XTracerImpl) IsTracing(ctx context.Context) (bool, error) {
	return client.HasTask(ctx), nil
}
