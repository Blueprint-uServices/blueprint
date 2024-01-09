// Package latencyinjector provides a Blueprint modifier for the server side of service calls.
//
// The plugin configures the server side to inject a user-defined amount of latency.
// Currently latency is injected for all requests and only a pre-defined duration is supported with no variance/noise added.
// The plugin will generate a wrapper class that will sleep for a fixed amount of time (the specified latency to be injected)
// before invoking the handler for handling the request.
// Example Usage to add 100ms latency to each request:
//    import "github.com/blueprint-uservices/blueprint/plugins/latency"
//    latency.AddFixed(spec, "my_service", "100ms")
//
// TODO: Allow latency to be injected selectively to a subset of requests based using
// random sampling or by matching attributes of a request.
// TODO: Sample injected latency from a configurable distribution and support choosing
// the distribution.
//
// If you need the above feature(s) consider submitting a PR, raising a feature request,
// or posting on the Blueprint slack/mailing list.
package latency

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds fixed-amount of latency on the server side during request processing for the specified sevrice.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that the server adds a fixed amount of `latency` while processing the request.
// The `latency` string must be a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". Negative signed values such as "-1.5h" would result in no explicit latency being added; although it might result in the go runtime re-scheduling the running goroutine.
// Usage:
//   AddFixed(spec, "my_service", "100ms")
func AddFixed(spec wiring.WiringSpec, serviceName string, latency string) {
	serverWrapper := serviceName + ".server.latency"
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add a latencyinjector to " + serviceName + " as it is not a pointer")
	}

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &LatencyInjectorWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(serverNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("LatencyInjector %s expected %s to be a golang.Service, but encountered %s", serverWrapper, serverNext, err)
		}

		return newLatencyInjectorWrapper(serverWrapper, wrapped, latency)
	})
}
