// Package jaeger provides a plugin to generate and include a jaeger collector instance in a Blueprint application.
//
// The package provides a jaeger container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications must use a backend.Tracer (runtime/core/backend) as the interface in the workflow.
package jaeger

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Generates the IRNodes for a jaeger docker container named `collectorName` that uses the latest jaeger:all-in-one container
// and the clients needed by the generated application to communicate with the server.
//
// The returned collectorName must be used as an argument to the opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, `collectorName`).
func DefineJaegerCollector(spec wiring.WiringSpec, collectorName string) string {
	// The nodes that we are defining
	collectorAddr := collectorName + ".addr"
	collectorCtr := collectorName + ".ctr"
	collectorClient := collectorName + ".client"

	// Define the Jaeger collector
	spec.Define(collectorCtr, &JaegerCollectorContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		collector, err := newJaegerCollectorContainer(collectorCtr)
		if err != nil {
			return nil, err
		}
		err = address.Bind[*JaegerCollectorContainer](ns, collectorAddr, collector, &collector.BindAddr)
		return collector, err
	})

	// Create a pointer to the collector
	ptr := pointer.CreatePointer[*JaegerCollectorClient](spec, collectorName, collectorCtr)

	// Define the address that points to the Jaeger collector
	address.Define[*JaegerCollectorContainer](spec, collectorAddr, collectorCtr)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, collectorAddr)

	// Define the Jaeger client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, collectorClient)
	spec.Define(collectorClient, &JaegerCollectorClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*JaegerCollectorContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newJaegerCollectorClient(collectorClient, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the Jaeger collector should do so through the pointer
	return collectorName
}
