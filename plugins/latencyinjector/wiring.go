// Package latencyinjector provides a Blueprint modifier for the server side of service calls.
//
// The plugin configures the server side to inject a user-defined amount of latency.
// Note: Currently only a fixed amount of latency can be applied for all the requests to a service. We might extend this in the future to sample the latency injection from a distribution.
package latencyinjector

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds fixed-amount of latency on the server side during request processing for the specified sevrice.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that the server adds a fixed amount of `latency` while processing the request.
// Usage:
//   AddFixedLatency(spec, "my_service", "100ms")
func AddFixedLatency(spec wiring.WiringSpec, serviceName string, latency string) {
	serverWrapper := serviceName + ".server.latency"
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add a latencyinjector to " + serviceName + " as it is not a pointer")
	}

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &LatencyInjector{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("LatencyInjector %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, err)
		}

		return newLatencyInjector(serverWrapper, wrapped, latency)
	})
}
