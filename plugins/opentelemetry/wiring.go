// Package opentelemetry provides three plugins:
// (i)  a plugin to generate and include an opentelemetry collector instance in a Blueprint application
// (ii) a plugin to wrap the service with an OpenTelemetry wrapper to generate OT compatible traces.
// (iii) a plugin to install an opentelemetry logger for a go process. The logger adds all the logs as events to the current active span.
//
// The package provides an in-memory trace exporter implementation and a go-client for generating traces on both the server and client side.
// The generated clients handle context propagation correctly on both the server and client sides.
//
// Example usage (for complete instrumentation):
//
// import "github.com/blueprint-uservices/blueprint/plugins/opentelemetry"
// import "github.com/blueprint-uservices/blueprint/plugins/jaeger"
// import "github.com/blueprint-uservices/blueprint/plugins/goproc"
//
// jaegerCollector := jaeger.DefineJaegerCollector(spec, "jaeger_collector") // Define a custom opentelemetry collector
//
// for _, service := range serviceNames {
//    opentelemetry.InstrumentUsingCustomCollector(spec, service, jaegerCollector) // Instrument a service with opentelemetry tracing
// }
//
// for _, proc := range procNames {
//    logger := opentelemetry.DefineOTTraceLogger(spec, proc) // Define an OTTrace logger for the desired process
//    goproc.SetLogger(spec, proc, logger) // Set the default logger for the desired process
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
// Note: Logs are added as `ot.Events` to the current span. If no current span is being recorded, then no events will be generated. Use `InstrumentUsingCustomCollector` to ensure that all services in a process are instrumented with OpenTelemetry and are creating active spans.
func DefineOTTraceLogger(spec wiring.WiringSpec, processName string) string {
	logger := processName + "_ottrace_logger"
	spec.Define(logger, &OTTraceLogger{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		return newOTTraceLogger(logger)
	})
	return logger
}
