package govector

import (
	"context"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

// Represents the GoVector logger interface exposed to applications and used by the GoVector plugin
type GoVector interface {
	backend.Logger
	// Gets the govector context (the vector clock) as a bytes array that will be sent from one process to another
	GetSendCtx(ctx context.Context, msg string) ([]byte, error)
	// Unpacks the received context `bytes` and merges the context into the process' current vector clock
	UnpackReceiveCtx(ctx context.Context, msg string, bytes []byte) error
}
