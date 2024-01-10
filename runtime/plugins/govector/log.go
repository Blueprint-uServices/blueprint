// package govector provides runtime GoVector components to be used by the govector plugin.
//
// Package provides a GoVector logger that maintains vector clocks for a process and correctly propagates the vector clocks to other processes through its service-level instrumentation.
//
// More info on govector: https://github.com/DistributedClocks/GoVector
package govector

import (
	"context"
	"errors"
	"fmt"

	"github.com/DistributedClocks/GoVector/govec"
	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

// GoVecLogger implements the GoVector interface (including the backend.Logger) by using the GoVector Logger
type GoVecLogger struct {
	logger *govec.GoLog
}

var logger *GoVecLogger

// Returns a new object of type GoVecLogger
func NewGoVecLogger(ctx context.Context, loggerName string) (*GoVecLogger, error) {
	// TODO Export options from the config
	if logger != nil {
		return logger, nil
	}
	config := govec.GetDefaultConfig()
	l := &GoVecLogger{govec.InitGoVector(loggerName, loggerName, config)}
	logger = l
	backend.SetDefaultLogger(logger)
	return logger, nil
}

// Returns the current GoVectorLogger. Used by the govector plugin when instrumenting the server and client side objects of services.
func GetLogger() *GoVecLogger {
	return logger
}

// Implements GoVector interface
func (g *GoVecLogger) GetSendCtx(ctx context.Context, msg string) ([]byte, error) {
	return g.logger.PrepareSend(msg, "", govec.GetDefaultLogOptions()), nil
}

// Implements GoVector interface
func (g *GoVecLogger) UnpackReceiveCtx(ctx context.Context, msg string, bytes []byte) error {
	incoming := ""
	g.logger.UnpackReceive(msg, bytes, &incoming, govec.GetDefaultLogOptions())
	return nil
}

// Implements backend.Logger interface
func (g *GoVecLogger) Debug(ctx context.Context, format string, args ...any) (context.Context, error) {
	opts := govec.GetDefaultLogOptions()
	opts = opts.SetPriority(govec.DEBUG)
	msg := fmt.Sprintf(format, args...)
	ok := g.logger.LogLocalEvent(msg, opts)
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}

// Implements backend.Logger interface
func (g *GoVecLogger) Info(ctx context.Context, format string, args ...any) (context.Context, error) {
	opts := govec.GetDefaultLogOptions()
	opts = opts.SetPriority(govec.INFO)
	msg := fmt.Sprintf(format, args...)
	ok := g.logger.LogLocalEvent(msg, opts)
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}

// Implements backend.Logger interface
func (g *GoVecLogger) Warn(ctx context.Context, format string, args ...any) (context.Context, error) {
	opts := govec.GetDefaultLogOptions()
	opts = opts.SetPriority(govec.WARNING)
	msg := fmt.Sprintf(format, args...)
	ok := g.logger.LogLocalEvent(msg, opts)
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}

// Implements backend.Logger interface
func (g *GoVecLogger) Error(ctx context.Context, format string, args ...any) (context.Context, error) {
	opts := govec.GetDefaultLogOptions()
	opts = opts.SetPriority(govec.ERROR)
	msg := fmt.Sprintf(format, args...)
	ok := g.logger.LogLocalEvent(msg, opts)
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}

// Implements backend.Logger interface
func (g *GoVecLogger) Logf(ctx context.Context, options backend.LogOptions, format string, args ...any) (context.Context, error) {
	msg := fmt.Sprintf(format, args...)
	opts := govec.GetDefaultLogOptions()
	opts = opts.SetPriority(govec.LogPriority(options.Level))
	ok := g.logger.LogLocalEvent(msg, opts)
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}
