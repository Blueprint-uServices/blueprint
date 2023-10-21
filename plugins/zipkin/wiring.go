package zipkin

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

// Defines the Zipkin collector as a process node. Also creates a pointer to the collector and a client node that are used by clients.
func DefineZipkinCollector(wiring blueprint.WiringSpec, collectorName string) string {
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	wiring.Define(collectorProc, &ZipkinCollector{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*ZipkinCollector](ns, collectorAddr)
		if err != nil {
			return nil, err
		}

		return newZipkinCollector(collectorProc, addr.Bind)
	})

	wiring.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(wiring, collectorDst, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, collectorName, &ZipkinCollectorClient{}, collectorDst)

	ptr := pointer.GetPointer(wiring, collectorName)

	ptr.AddDstModifier(wiring, collectorAddr)

	clientNext := ptr.AddSrcModifier(wiring, collectorClient)

	wiring.Define(collectorClient, &ZipkinCollectorClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*ZipkinCollector](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newZipkinCollectorClient(collectorClient, addr.Dial)
	})

	address.Define[*ZipkinCollector](wiring, collectorAddr, collectorProc, &blueprint.ApplicationNode{})
	ptr.AddDstModifier(wiring, collectorAddr)

	return collectorName
}
