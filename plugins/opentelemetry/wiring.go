package opentelemetry

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"golang.org/x/exp/slog"
)

/*
Instruments `serviceName` with OpenTelemetry.  This can only be done if `serviceName` is a
pointer from Golang nodes to Golang nodes.

This call will also define the OpenTelemetry collector.

Instrumenting `serviceName` will add both src and dst-side modifiers to the pointer.
*/
func Instrument(wiring blueprint.WiringSpec, serviceName string) {
	DefineOpenTelemetryCollector(wiring, DefaultOpenTelemetryCollectorName)
	InstrumentUsingCustomCollector(wiring, serviceName, DefaultOpenTelemetryCollectorName)
}

/*
This is the same as the Instrument function, but uses `collectorName` as the OpenTelemetry
collector and does not attempt to define or redefine the collector.
*/
func InstrumentUsingCustomCollector(wiring blueprint.WiringSpec, serviceName string, collectorName string) {
	// The nodes that we are defining
	clientWrapper := serviceName + ".client.ot"
	serverWrapper := serviceName + ".server.ot"

	// Get the pointer metadata
	ptr := pointer.GetPointer(wiring, serviceName)
	if ptr == nil {
		slog.Error("Unable to instrument " + serviceName + " with OpenTelemetry as it is not a pointer")
		return
	}

	// Add the client wrapper to the pointer src
	clientNext := ptr.AddSrcModifier(wiring, clientWrapper)

	// Define the client wrapper
	wiring.Define(clientWrapper, &OpenTelemetryClientWrapper{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var server golang.Service
		err := namespace.Get(clientNext, &server)
		if err != nil {
			return nil, err
		}

		var collectorClient *OpenTelemetryCollectorClient
		err = namespace.Get(collectorName, &collectorClient)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryClientWrapper(clientWrapper, server, collectorClient)
	})

	// Add the server wrapper to the pointer dst
	serverNext := ptr.AddDstModifier(wiring, serverWrapper)

	// Define the server wrapper
	wiring.Define(serverWrapper, &OpenTelemetryServerWrapper{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var wrapped golang.Service
		err := namespace.Get(serverNext, &wrapped)
		if err != nil {
			return nil, err
		}

		var collectorClient *OpenTelemetryCollectorClient
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
func DefineOpenTelemetryCollector(wiring blueprint.WiringSpec, collectorName string) string {
	// The nodes that we are defining
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	// Define the collector address

	// Define the collector server
	wiring.Define(collectorProc, &OpenTelemetryCollector{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *address.Address[*OpenTelemetryCollector]
		err := namespace.Get(collectorAddr, &addr)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryCollector(collectorProc, addr)
	})

	// By default, we should only have one collector globally
	wiring.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(wiring, collectorDst, &blueprint.ApplicationNode{})

	// Define the pointer to the collector for golang clients
	pointer.CreatePointer(wiring, collectorName, &OpenTelemetryCollectorClient{}, collectorDst)
	ptr := pointer.GetPointer(wiring, collectorName)

	// Add the collectorAddr to the pointer dst
	ptr.AddDstModifier(wiring, collectorAddr)

	// Add the client to the pointer
	clientNext := ptr.AddSrcModifier(wiring, collectorClient)

	// Define the collector client
	wiring.Define(collectorClient, &OpenTelemetryCollectorClient{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *address.Address[*OpenTelemetryCollector]
		err := namespace.Get(clientNext, &addr)
		if err != nil {
			return nil, err
		}

		return newOpenTelemetryCollectorClient(collectorClient, addr)
	})

	// Define the address and add it to the pointer dst
	address.Define(wiring, collectorAddr, collectorProc, &blueprint.ApplicationNode{}, func(namespace blueprint.Namespace) (address.Node, error) {
		addr := &address.Address[*OpenTelemetryCollector]{
			AddrName: collectorAddr,
			Server:   nil,
		}
		return addr, nil
	})
	ptr.AddDstModifier(wiring, collectorAddr)

	// Return the name of the pointer
	return collectorName
}
