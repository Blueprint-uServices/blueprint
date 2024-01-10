// Package zipkin provides a plugin to generate and include a zipkin collector instance in a Blueprint application.
//
// The package provides a zipkin container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications must use a backend.Tracer (runtime/core/backend) as the interface in the workflow.
package zipkin

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Generates the IRNodes for a zipkin docker container named `collectorName` that uses the latest zipkin container
// and the clients needed by the generated application to communicate with the server.
//
// The returned collectorName must be used as an argument to the opentelemetry.Instrument(spec, serviceName, `collectorName`).
func Collector(spec wiring.WiringSpec, collectorName string) string {
	// The nodes that we are defining
	collectorAddr := collectorName + ".addr"
	collectorCtr := collectorName + ".ctr"
	collectorClient := collectorName + ".client"

	// Define the Zipkin collector container
	spec.Define(collectorCtr, &ZipkinCollectorContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		zipkin, err := newZipkinCollectorContainer(collectorCtr)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*ZipkinCollectorContainer](ns, collectorAddr, zipkin, &zipkin.BindAddr)
		return zipkin, err
	})

	// Create a pointer to the Zipkin collector container
	ptr := pointer.CreatePointer[*ZipkinCollectorClient](spec, collectorName, collectorCtr)

	// Define the address that points to the Zipkin collector container
	address.Define[*ZipkinCollectorContainer](spec, collectorAddr, collectorCtr)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, collectorAddr)

	// Define the Zipkin collector client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, collectorClient)
	spec.Define(collectorClient, &ZipkinCollectorClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*ZipkinCollectorContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newZipkinCollectorClient(collectorClient, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the Zipkin collector instance should do so through the pointer
	return collectorName
}
