// Package retries provides a Blueprint modifier for the client side of service calls.
//
// The plugin wraps clients with a retrier using that retries a request until one of the two conditions is met:
// i)  the requests returns without an error
// ii) the number of failed tries has reached the maximum number of failures.
// Usage:
//
//	import "github.com/blueprint-uservices/blueprint/plugins/retries"
//	 retries.AddRetries(spec, "my_service", 10) // Adds retries with a maximum number of retries
//	 retries.AddRetriesWithTimeouts(spec, "my_service", 10, "1s") // Adds retries and timeouts
//	 retries.AddRetriesWithFixedDelay(spec, "my_service", 10, "50ms") // Adds retries with a maximum number of retries and a fixed delay between any two tries.
//	 retries.AddRetriesWithExponentialBackoff(spec, "my_service", "100ms", "1s") // Adds retries with exponential backoff delay strategy between retries.
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
//
//	AddRetries(spec, "my_service", 10)
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
//
//	workflow -> plugin grpc
//
// After:
//
//	workflow -> retrier -> timeout -> plugin grpc
//
// Usage:
//
//	AddRetriesWithTimeouts(spec, "my_service", 10, "1s")
func AddRetriesWithTimeouts(spec wiring.WiringSpec, serviceName string, max_retries int64, timeout string) {
	AddRetries(spec, serviceName, max_retries)
	timeouts.Add(spec, serviceName, timeout)
}

// Add retrier functionality to all clients of the specified service with a fixed time delay between the tries.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that all clients to that service retry `max_retries` number of times on error with `delay` between any pair of tries.
// Usage:
//
//	AddRetriesWithFixedDelay(spec, "my_service", 10, "50ms")
func AddRetriesWithFixedDelay(spec wiring.WiringSpec, serviceName string, max_retries int64, delay string) {
	clientWrapper := serviceName + ".client.retrierfd"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add retries to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &RetrierExponentialBackoffClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Retries %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newRetrierFixedDelayClient(clientWrapper, wrapped, max_retries, delay)
	})
}

// Add retrier functionality to all clients of the specified service with a fixed time delay between the tries.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that all clients to that service retry with exponential delay.
// The `starting_delay` is the first delay to be used before retrying.
// The retries continue until a `backoff_limit` of delay is reached
// `useJitter` indicates whether to use jitter in the delay or not, 
// jitter is a random value added to the delay.
// Usage:
//
//	AddRetriesWithExponentialBackoff(spec, "my_service", "100ms", "1s", false)
func AddRetriesWithExponentialBackoff(spec wiring.WiringSpec, serviceName string, starting_delay string, backoff_limit string, useJitter bool) {
	clientWrapper := serviceName + ".client.retrierfd"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add retries to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &RetrierExponentialBackoffClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Retries %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newRetrierExponentialBackoffClient(clientWrapper, wrapped, starting_delay, backoff_limit, useJitter)
	})
}

// Add retrier functionality to all clients of the specified service.
// Uses a [blueprint.WiringSpec]
// Modifies the given service such that all clients to that service retry `max_retries` number of times on error.
// The `retry_rate_limit` is the maximum number of retries per second, if `rate_limit` is set to 0, then there is no rate limit.
// Usage:
//
//	AddRetriesRetryRateLimit(spec, "my_service", 10, 100)
func AddRetriesRetryRateLimit(spec wiring.WiringSpec, serviceName string, max_retries int64, retry_rate_limit int64) {
	clientWrapper := serviceName + ".client.retrier"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add retries to " + serviceName + " as it is not a pointer")
		return
	}

	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	spec.Define(clientWrapper, &RetrierRateLimiterClient{Max: max_retries, RetryRateLimit: retry_rate_limit}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service

		if err := ns.Get(clientNext, &wrapped); err != nil {
			return nil, blueprint.Errorf("Retries %s expected %s to be a golang.Service, but encountered %s", clientWrapper, clientNext, err)
		}

		return newRetrierRateLimiterClient(clientWrapper, wrapped, max_retries, retry_rate_limit)
	})
}
