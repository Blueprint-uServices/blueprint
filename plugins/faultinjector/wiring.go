// Package faultinjector provides Blueprint modifiers for injecting faults in an application.
package faultinjector

import (
	"log/slog"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/faultinjector/delay"
	"github.com/blueprint-uservices/blueprint/plugins/faultinjector/probabilistic"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
)

// Adds random delay on the server side implementation of the specified service.
//
// @param spec: a [blueprint.WiringSpec]
//
// @param serviceName: Name of the service to add delay to. There must be an instance with this name in the wiring specification.
//
// @param maxDelay: Specifies the maximum amount of delay to be added (in milliseconds)
//
// Usage:
//
//	faultinjector.AddRandomDelay(spec, "serviceA", 100)
func AddRandomDelay(spec wiring.WiringSpec, serviceName string, maxDelay int64) {
	serverWrapper := serviceName + ".server.delay"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add random delay node to " + serviceName)
	}

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &delay.RandomDelayServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("RandomDelay %s expected %s to be a golang.Service, but encountered %v", serverWrapper, serverNext, err)
		}

		return delay.NewRandomDelayServerWrapper(serverWrapper, server, maxDelay)
	})
}

// Fails requests with a given probability on the server side
//
// @param `spec“: a [blueprint.WiringSpec]
//
// @param `serviceName`: Name of the service to add probabilistic failures to. There must be an instance with this name in the wiring specification.
//
// @param `probability`: Probability of a request failing. Must be in the range of [1, 100].
//
// Usage:
//
//	faultinjector.AddProbabilisticFailures(spec, "serviceA", 51)
func AddProbabilisticFailures(spec wiring.WiringSpec, serviceName string, probability int) {
	serverWrapper := serviceName + ".server.probabilistic_failure"

	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to add probabilistic failures node to " + serviceName)
	}

	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	spec.Define(serverWrapper, &probabilistic.ServerWrapper{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		var server golang.Service
		if err := ns.Get(serverNext, &server); err != nil {
			return nil, blueprint.Errorf("ProbabilisticFailures %s expected %s to be a golang.Service, but encountered %v", serverWrapper, serverNext, err)
		}
		return probabilistic.NewServerWrapper(serverWrapper, server, probability)
	})
}
