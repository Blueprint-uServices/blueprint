package govector

import (
	"context"

	"github.com/DistributedClocks/GoVector/govec"
)

type GoVecLogger struct {
	logger *govec.GoLog
}

func NewGoVecLogger(ctx context.Context, proc_name string) (*GoVecLogger, error) {
	// TODO Export other options from the config
	config := govec.GetDefaultConfig()
	config.Buffered = true
	return &GoVecLogger{govec.InitGoVector(proc_name, proc_name, config)}, nil
}

func (g *GoVecLogger) GetSendCtx(ctx context.Context, msg string) ([]byte, error) {
	return g.logger.PrepareSend(msg, "", govec.GetDefaultLogOptions()), nil
}

func (g *GoVecLogger) UnpackReceiveCtx(ctx context.Context, msg string, bytes []byte) error {
	incoming := ""
	g.logger.UnpackReceive(msg, bytes, &incoming, govec.GetDefaultLogOptions())
	return nil
}
