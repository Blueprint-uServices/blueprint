package circuitbreaker

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds circuit breaker functionality to all clients of the specified service.
// Uses a [blueprint.WiringSpec].
// Circuit breaker trips when `failure_rate` percentage of requests fail. Minimum number of requests for the circuit to break is specified using `min_reqs`.
// The circuit breaker counters are reset after `interval` duration.
// Usage:
//
//	AddCircuitBreaker(wiring, "serviceA", 1000, 0.1, "1s")
func AddCircuitBreaker(wiring blueprint.WiringSpec, serviceName string, min_reqs int64, failure_rate float64, interval string) {
	clientWrapper := serviceName + ".client.cb"

	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to add a circuit breaker to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(wiring, clientWrapper)

	wiring.Define(clientWrapper, &CircuitBreakerClient{Min_Reqs: min_reqs, FailureRate: failure_rate}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("CircuitBreaker %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newCircuitBreakerClient(clientWrapper, wrapped, min_reqs, failure_rate, interval)
	})
}
