package zipkin

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines the Zipkin collector as a process node. Also creates a pointer to the collector and a client node that are used by clients.
func DefineZipkinCollector(spec wiring.WiringSpec, collectorName string) string {
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
