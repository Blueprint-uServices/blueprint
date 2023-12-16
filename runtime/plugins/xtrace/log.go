package xtrace

import (
	"context"
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
)

// Implementation of the [backend.Logger] interface
// Note: This logger should only be used in conjunction with the XTracerImpl tracer. Using this logger without using the XTracerImpl tracer would result in no-op logging behavior.
type XTraceLogger struct {
	backend.Logger
}

// Returns a new instance of [XTracerImpl] that connects to a xtrace server running at `addr`.
// REQUIRED: An xtrace server must be running at `addr`
func NewXTraceLogger(ctx context.Context, addr string) (*XTraceLogger, error) {
	err := connectClient(addr)
	if err != nil {
		return nil, err
	}
	l := &XTraceLogger{}
	backend.SetDefaultLogger(l)
	return l, nil
}

func (l *XTraceLogger) Log(ctx context.Context, priority backend.Priority, msg string, attrs ...backend.Attribute) context.Context {
	if !client.HasTask(ctx) {
		// Only do logging if there is an active task running
		return ctx
	}
	formatted_msg := fmt.Sprintf("%s: %s", priority.String(), msg)
	if len(attrs) > 0 {
		formatted_msg = fmt.Sprintf("%s, %v", formatted_msg, attrs)
	}
	return client.Log(ctx, formatted_msg)
}
