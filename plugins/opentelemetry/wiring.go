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
Instruments `serviceName` with OpenTelemetry.  This can only be done if `serviceName` is a
pointer from Golang nodes to Golang nodes.

This call will also define the OpenTelemetry collector.

Instrumenting `serviceName` will add both src and dst-side modifiers to the pointer.
*/
func Instrument(spec wiring.WiringSpec, serviceName string) {
	DefineOpenTelemetryCollector(spec, DefaultOpenTelemetryCollectorName)
	InstrumentUsingCustomCollector(spec, serviceName, DefaultOpenTelemetryCollectorName)
}

/*
This is the same as the Instrument function, but uses `collectorName` as the OpenTelemetry
collector and does not attempt to define or redefine the collector.
*/
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
