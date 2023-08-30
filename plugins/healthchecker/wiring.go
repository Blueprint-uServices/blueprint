package healthchecker

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"golang.org/x/exp/slog"
)

func AddHealthCheckAPI(wiring blueprint.WiringSpec, serviceName string) {
	// The node that we are defining
	serverWrapper := serviceName + ".server.hc"

	// Get the pointer metadata
	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to add healthcheck API to " + serviceName + " as it is not a pointer")
		return
	}

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(wiring, serverWrapper)

	// Define the server wrapper
	wiring.Define(serverWrapper, &HealthCheckerServerWrapper{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		server, err := scope.Get(serverNext)
		if err != nil {
			return nil, err
		}

		return newHealthCheckerServerWrapper(serverWrapper, server)
	})
}
