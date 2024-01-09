// Package retries provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a retrier using that retries a request until one of the two conditions is met:
// i)  the requests returns without an error
// ii) the number of failed tries has reached the maximum number of failures.
// Usage:
//  import "github.com/blueprint-uservices/blueprint/plugins/retries"
//  retries.AddRetries(spec, "my_service", 10) // Only adds retries
//  retries.AddRetriesWithTimeouts(spec, "my_service", 10, "1s") // Adds retries and timeouts
package retries

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/timeouts"
	"golang.org/x/exp/slog"
)

// Add retrier functionality to all clients of the specified service.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that all clients to that service retry `max_retries` number of times on error.
// Usage:
//   AddRetries(spec, "my_service", 10)
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

// Add retrier + timeout functionality to all clients of the specified service.
// Uses a [blueprint.WiringSpec]
// Modifies the given service in the following ways:
// (i)  all clients to that service have a user-specified `timeout` for each request.
// (ii) all clients to that service retry at most `max_retries` number of times on error.
//
// Ordering of functionality depicted via example call-chain:
// Before:
//   workflow -> plugin grpc
// After:
//   workflow -> retrier -> timeout -> plugin grpc
//
// Usage:
//   AddRetriesWithTimeouts(spec, "my_service", 10, "1s")
func AddRetriesWithTimeouts(spec wiring.WiringSpec, serviceName string, max_retries int64, timeout string) {
	AddRetries(spec, serviceName, max_retries)
	timeouts.Add(spec, serviceName, timeout)
}
