package govector

import (
	"context"

	"github.com/DistributedClocks/GoVector/govec"
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
