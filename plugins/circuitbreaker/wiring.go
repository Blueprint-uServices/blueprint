// Package circuitbreaker provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a circuitbreaker that blocks any new requests from being sent out over a connection if the failure rate exceeds a provided number in a fixed duration. The block is removed after the completion of the fixed duration interval.
package circuitbreaker

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds circuit breaker functionality to all clients of the specified service.
// Uses a [blueprint.WiringSpec].
// Circuit breaker trips when `failure_rate` percentage of requests fail. Minimum number of requests for the circuit to break is specified using `min_reqs`.
// The circuit breaker counters are reset after `interval` duration.
// Usage:
//
//	AddCircuitBreaker(spec, "serviceA", 1000, 0.1, "1s")
func AddCircuitBreaker(spec wiring.WiringSpec, serviceName string, min_reqs int64, failure_rate float64, interval string) {
	clientWrapper := serviceName + ".client.cb"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add a circuit breaker to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &CircuitBreakerClient{Min_Reqs: min_reqs, FailureRate: failure_rate}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("CircuitBreaker %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newCircuitBreakerClient(clientWrapper, wrapped, min_reqs, failure_rate, interval)
	})
}
