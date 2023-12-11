// Package opentelemetry provides two plugins:
// (i)  a plugin to generate and include an opentelemetry collector instance in a Blueprint application
// (ii) provides a modifier plugin to wrap the service with an OpenTelemetry wrapper to generate OT compatible traces/logs.
//
// The package provides an in-memory trace exporter implementation and a go-client for generating traces on both the server and client side.
// The generated clients handle context propagation correctly on both the server and client sides.
//
// The applications must use a backend.Tracer (runtime/core/backend) as the interface in the workflow.
package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Instruments `serviceName` with OpenTelemetry.  This can only be done if `serviceName` is a service declared in the wiring spec using [workflow.Define]

This call will only export traces to stdout and no external collector will be defined.
Use the InstrumentUsingCustomCollector for exporting traces to a custom collector such as the ones provided by jaeger or zipkin.
*/
func Instrument(spec wiring.WiringSpec, serviceName string) {
	DefineOpenTelemetryCollector(spec, DefaultOpenTelemetryCollectorName)
	InstrumentUsingCustomCollector(spec, serviceName, DefaultOpenTelemetryCollectorName)
}

// Instruments `serviceName` with OpenTelemetry.  This can only be done if `serviceName` is a service declared in the wiring spec using [workflow.Define]
//
// This call will configure the generated clients on server and client side to use the exporter provided by the custom collector indicated by the `collectorName`.
// The `collectorName` must be declared in the wiring spec.
func InstrumentUsingCustomCollector(spec wiring.WiringSpec, serviceName string, collectorName string) {
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

var DefaultOpenTelemetryCollectorName = "ot_collector"

/*
Defines the OpenTelemetry collector as a process node

# Also creates a pointer to the collector and a client node that are used by OT clients

This doesn't need to be explicitly called, although it can if users want to control
the placement of the opentelemetry collector

NOTE: Currently does not include the opentelemetry collector. This might change in the future.
*/
func DefineOpenTelemetryCollector(spec wiring.WiringSpec, collectorName string) string {
	// The nodes that we are defining
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	// Define the collector address

	// Define the collector server
	spec.Define(collectorProc, &OpenTelemetryCollector{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*OpenTelemetryCollector](namespace, collectorAddr)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryCollector(collectorProc, addr.Bind)
	})

	// By default, we should only have one collector globally
	spec.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(spec, collectorDst, &ir.ApplicationNode{})

	// Define the pointer to the collector for golang clients
	pointer.CreatePointer(spec, collectorName, &OpenTelemetryCollectorClient{}, collectorDst)
	ptr := pointer.GetPointer(spec, collectorName)

	// Add the collectorAddr to the pointer dst
	ptr.AddDstModifier(spec, collectorAddr)

	// Add the client to the pointer
	clientNext := ptr.AddSrcModifier(spec, collectorClient)

	// Define the collector client
	spec.Define(collectorClient, &OpenTelemetryCollectorClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*OpenTelemetryCollector](namespace, clientNext)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryCollectorClient(collectorClient, addr.Dial)
	})

	// Define the address and add it to the pointer dst
	address.Define[*OpenTelemetryCollector](spec, collectorAddr, collectorProc, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, collectorAddr)

	// Return the name of the pointer
	return collectorName
}
