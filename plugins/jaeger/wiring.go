// Package jaeger provides a plugin to generate and include a jaeger collector instance in a Blueprint application.
//
// # Wiring Spec Usage
//
// To instantiate a Jaeger container:
//
//	collector := jaeger.Collector(spec, "jaeger")
//
// The returned `collector` must be used as an argument to the `opentelemetry.Instrument(spec, serviceName, collector)` to ensure the spans generated by instrumented services are correctly exported to the instantiated server.
//
// # Artifacts Generated
//
//  1. The package generates a jaeger docker container that provides the server-side implementation of the Jaeger collector.
//  2. Instantiates a [JaegerTracer] instance for configuring the opentelemetry runtime libraries to export all generated traces to the Jaeger collector.
//
// [JaegerTracer]: https://github.com/Blueprint-uServices/blueprint/tree/main/runtime/plugins/jaeger
package jaeger

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// [Collector] can be used by wiring specs to instantiate a jaeger docker container named `collectorName` that uses the latest jaeger:all-in-one container
// and generates the clients needed by the generated application to communicate with the server.
//
// The returned collectorName must be used as an argument to the opentelemetry.Instrument(spec, serviceName, `collectorName`) to ensure the spans generated by instrumented services are correctly exported to the instantiated server.
//
// # Wiring Spec Usage
//
//	jaeger.Collector(spec, "jaeger")
func Collector(spec wiring.WiringSpec, collectorName string) string {
	// The nodes that we are defining
	collectorAddr := collectorName + ".addr"
	collectorUIAddr := collectorName + ".ui.addr"
	collectorCtr := collectorName + ".ctr"
	collectorClient := collectorName + ".client"

	// Define the Jaeger collector
	spec.Define(collectorCtr, &JaegerCollectorContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		collector, err := newJaegerCollectorContainer(collectorCtr)
		if err != nil {
			return nil, err
		}
		err = address.Bind[*JaegerCollectorContainer](ns, collectorAddr, collector, &collector.BindAddr)
		if err != nil {
			return nil, err
		}
		err = address.Bind[*JaegerCollectorContainer](ns, collectorUIAddr, collector, &collector.UIBindAddr)
		return collector, err
	})

	// Create a pointer to the collector
	ptr := pointer.CreatePointer[*JaegerCollectorClient](spec, collectorName, collectorCtr)

	// Define the address that points to the Jaeger collector
	address.Define[*JaegerCollectorContainer](spec, collectorUIAddr, collectorCtr)
	address.Define[*JaegerCollectorContainer](spec, collectorAddr, collectorCtr)

	// Add the addresses to the pointer
	ptr.AddAddrModifier(spec, collectorUIAddr)
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
