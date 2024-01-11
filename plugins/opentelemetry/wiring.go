// Package opentelemetry provides three plugins:
//
//  1. a plugin to generate and include an opentelemetry collector instance in a Blueprint application.
//  2. a plugin to wrap the service with an OpenTelemetry wrapper to generate OT compatible traces by starting and stopping client spans for remote calls between services and correctly propagating context between services.
//  3. a plugin to install an opentelemetry logger for a go process. The logger adds all the logs as events to the current active span.
//
// In order to generate complete end-to-end traces of the application, all services of the application need to be instrumented with OpenTelemetry.
// If the plugin is only applied to a subset of services, the application will run, but the traces it produces won't be end-to-end and won't be useful.
//
// The package provides a go-client for generating traces on both the server and client side along with the ability to export traces to a custom collector such as jaeger or zipkin.
// The generated clients handle context propagation correctly on both the server and client sides.
//
// The plugin does not support instrumenting clients for backends such as databases, caches, queues, etc. If needed, please consider submitting a PR or contacting the maintainers via google groups or on slack.
//
// Example usage (for complete instrumentation):
//
// # Wiring Example
//
// func applyOTeOptions() {
// 	jaeger.Container(spec, "xtrace_server") // Defines and adds a jaeger server to the wiring spec. More info in Jaeger plugin.
//
// 	for _, service := range serviceNames {
//  	 opentelemetry.Instrument(spec, service) // Instrument service with ottrace instrumentation
// 	}
//
// 	for _, proc := range processNames {
//  	 opentelemetry.Logger(spec, proc) // Define an ottrace-logger for the process
// 	}
// }
package opentelemetry

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

// Instruments `serviceName` with OpenTelemetry.  This can only be done if `serviceName` is a service declared in the wiring spec using [workflow.Define]
//
// This call will configure the generated clients on server and client side to use the exporter provided by the custom collector indicated by the `collectorName`.
// The `collectorName` must be declared in the wiring spec.
func Instrument(spec wiring.WiringSpec, serviceName string, collectorName string) {
	// The nodes that we are defining
	clientWrapper := serviceName + ".client.ot"
	serverWrapper := serviceName + ".server.ot"

	// Get the pointer metadata
	ptr := pointer.GetPointer(spec, serviceName)
	if ptr == nil {
		slog.Error("Unable to instrument " + serviceName + " with OpenTelemetry as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(spec, clientWrapper)

	// Define the client wrapper
	spec.Define(clientWrapper, &OpenTelemetryClientWrapper{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var server golang.Service
		err := namespace.Get(clientNext, &server)
		if err != nil {
			return nil, err
		}

		var collectorClient OpenTelemetryCollectorInterface
		err = namespace.Get(collectorName, &collectorClient)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryClientWrapper(clientWrapper, server, collectorClient)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(spec, serverWrapper)

	// Define the server wrapper
	spec.Define(serverWrapper, &OpenTelemetryServerWrapper{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		var wrapped golang.Service
		err := namespace.Get(serverNext, &wrapped)
		if err != nil {
			return nil, err
		}

		var collectorClient OpenTelemetryCollectorInterface
		err = namespace.Get(collectorName, &collectorClient)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryServerWrapper(serverWrapper, wrapped, collectorClient)
	})

}

// Adds a process-level ot logger for process `processName` to be used in tandem with an OT Tracer.
// Logs are added as `ot.Events` to the current span and will be added as events to the current span and won't appear in stdout.
// If no current span is being recorded, then no events will be generated. Use `Instrument` to ensure that all services in a process are instrumented with OpenTelemetry and are creating active spans.
func Logger(spec wiring.WiringSpec, processName string) string {
	logger := processName + "_ottrace_logger"
	spec.Define(logger, &OTTraceLogger{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		return newOTTraceLogger(logger)
	})
	return logger
}
