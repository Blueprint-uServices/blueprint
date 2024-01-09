// Package healthchecker provides a Blueprint modifier for the server side of a service.
//
// The plugin extends the service interface with a `Health` method that returns a success string if the service is healthy.
// Note: The plugin __does not__ check the health of all of the dependencies of the service.
package healthchecker

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Adds a health check API to the server side implementation of the specified service.
// Uses a [blueprint.WiringSpec].
// Usage:
//
//  AddHealthCheckAPI(spec, "serviceA")
//
// Result:
//  Old interface:
//  type ServiceA interface {
//     Method1(ctx context.Context, ...) (..., error)
//     ...
//     MethodN(ctx context.Context, ...) (..., error)
//  }
//  New interface:
//  type ServiceAHealth interface {
//     Method1(ctx context.Context, ...) (..., error)
//     ...
//     MethodN(ctx context.Context, ...) (..., error)
//     Health(ctx context.Context) (string, error)
//  }
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
