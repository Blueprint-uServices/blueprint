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
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	spec.Define(collectorProc, &JaegerCollectorContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*JaegerCollectorContainer](ns, collectorAddr)
		if err != nil {
			return nil, err
		}

		return newJaegerCollectorContainer(collectorProc, addr.Bind)
	})

	spec.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(spec, collectorDst, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, collectorName, &JaegerCollectorClient{}, collectorDst)

	ptr := pointer.GetPointer(spec, collectorName)

	ptr.AddDstModifier(spec, collectorAddr)

	clientNext := ptr.AddSrcModifier(spec, collectorClient)

	spec.Define(collectorClient, &JaegerCollectorClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*JaegerCollectorContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newJaegerCollectorClient(collectorClient, addr.Dial)
	})

	address.Define[*JaegerCollectorContainer](spec, collectorAddr, collectorProc, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, collectorAddr)

	return collectorName
}
