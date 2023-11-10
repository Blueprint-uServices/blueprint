package healthchecker

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

func AddHealthCheckAPI(spec wiring.WiringSpec, serviceName string) {
	// The node that we are defining
	serverWrapper := serviceName + ".server.hc"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add healthcheck API to " + serviceName + " as it is not a pointer")
		return
	}

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	// Define the server wrapper
	spec.Define(serverWrapper, &HealthCheckerServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("Healthchecker %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, err)
		}

		return newHealthCheckerServerWrapper(serverWrapper, server)
	})
}
