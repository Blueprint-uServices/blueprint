// Package timeouts provides a Blueprint modifier for the client side of service calls.
//
// The plugin configures clients with a timeout mechanism using contexts.
// Usage:
//  AddTimeouts(spec, "my_service", "1s")
package timeouts

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds timeouts to client calls for the specified service.
// Uses a [blueprint.WiringSpec].
// Modifies the given service such that all clients to that service have a user-specified `timeout`.
// Usage:
//   AddTimeouts(spec, "my_service", "1s")
func AddTimeouts(spec wiring.WiringSpec, serviceName string, timeout string) {
	clientWrapper := serviceName + ".client.timeout"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add timeouts to " + serviceName + " as it is not a pointer")
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &TimeoutClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Timeouts %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newTimeoutClient(clientWrapper, wrapped, timeout)
	})
}
