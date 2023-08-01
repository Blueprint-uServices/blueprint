package core

import (
	"gitlab.mpi-sws.org/cld/blueprint/pkg/blueprint"
	"golang.org/x/exp/slog"
)

func MakeProcess() {
	blueprint.Wiring()
	slog.Info("Hello from MakeProcess")
}
