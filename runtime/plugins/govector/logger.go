package govector

import (
	"context"
	"errors"
	"fmt"

	"github.com/DistributedClocks/GoVector/govec"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type GoVecLogger struct {
	logger *govec.GoLog
}

var logger *GoVecLogger

func NewGoVecLogger(ctx context.Context, proc_name string) (*GoVecLogger, error) {
	// TODO Export options from the config
	if logger != nil {
		return logger, nil
	}
	config := govec.GetDefaultConfig()
	l := &GoVecLogger{govec.InitGoVector(proc_name, proc_name, config)}
	logger = l
	backend.SetDefaultLogger(logger)
	return logger, nil
}

func GetLogger() *GoVecLogger {
	return logger
}

func (g *GoVecLogger) GetSendCtx(ctx context.Context, msg string) ([]byte, error) {
	return g.logger.PrepareSend(msg, "", govec.GetDefaultLogOptions()), nil
}

func (g *GoVecLogger) UnpackReceiveCtx(ctx context.Context, msg string, bytes []byte) error {
	incoming := ""
	g.logger.UnpackReceive(msg, bytes, &incoming, govec.GetDefaultLogOptions())
	return nil
}

func (g *GoVecLogger) Log(ctx context.Context, priority backend.Priority, msg string, attrs ...backend.Attribute) (context.Context, error) {
	for _, a := range attrs {
		msg += fmt.Sprintf("%v", a)
	}
	ok := g.logger.LogLocalEvent(msg, govec.GetDefaultLogOptions())
	if !ok {
		return ctx, errors.New("Failed to log local event")
	}
	return ctx, nil
}
