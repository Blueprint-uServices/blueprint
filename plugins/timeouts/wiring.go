// Package timeouts provides a Blueprint modifier for the client side of service calls.
//
// The plugin configures clients with a timeout mechanism using contexts.
// The plugin will generate a wrapper client class that will wait for a fixed amount of time (the specified timeout value) before canceling the context. Once the context is cancelled, the execution returns to the caller.
//
// Example Usage to add a "1s" timeout to each request:
//  timeouts.Add(spec, "my_service", "1s")
package timeouts

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds timeouts to client calls for the specified service.
// Uses a [blueprint.WiringSpec].
// Modifies the given service such that all clients to that service have a user-specified `timeout`.
//
// The `timeout` string must be a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
//
// Usage:
//   Add(spec, "my_service", "1s")
func Add(spec wiring.WiringSpec, serviceName string, timeout string) {
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
