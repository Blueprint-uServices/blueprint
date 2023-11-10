package retries

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Modifies the given service such that all clients to that service retry `max_retries` number of times on error.
func AddRetries(spec wiring.WiringSpec, serviceName string, max_retries int64) {
	clientWrapper := serviceName + ".client.retrier"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add retries to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &RetrierClient{Max: max_retries}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Retries %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newRetrierClient(clientWrapper, wrapped, max_retries)
	})
}
