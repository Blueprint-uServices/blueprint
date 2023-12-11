// Package zipkin provides a plugin to generate and include a zipkin collector instance in a Blueprint application.
//
// The package provides a zipkin container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications must use a backend.Tracer (runtime/core/backend) as the interface in the workflow.
package zipkin

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Generates the IRNodes for a zipkin docker container named `collectorName` that uses the latest zipkin container
// and the clients needed by the generated application to communicate with the server.
//
// The returned collectorName must be used as an argument to the opentelemetry.InstrumentUsingCustomCollector(spec, serviceName, `collectorName`).
func Collector(spec wiring.WiringSpec, collectorName string) string {
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	spec.Define(collectorProc, &ZipkinCollectorContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*ZipkinCollectorContainer](ns, collectorAddr)
		if err != nil {
			return nil, err
		}

		return newZipkinCollectorContainer(collectorProc, addr.Bind)
	})

	spec.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(spec, collectorDst, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, collectorName, &ZipkinCollectorClient{}, collectorDst)

	ptr := pointer.GetPointer(spec, collectorName)

	ptr.AddDstModifier(spec, collectorAddr)

	clientNext := ptr.AddSrcModifier(spec, collectorClient)

	spec.Define(collectorClient, &ZipkinCollectorClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*ZipkinCollectorContainer](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newZipkinCollectorClient(collectorClient, addr.Dial)
	})

	address.Define[*ZipkinCollectorContainer](spec, collectorAddr, collectorProc, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, collectorAddr)

	return collectorName
}
