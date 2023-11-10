package jaeger

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines the Jaeger collector as a process node. Also creates a pointer to the collector and a client node that are used by clients.
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
