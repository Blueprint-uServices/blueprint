package jaeger

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

// Defines the Jaeger collector as a process node. Also creates a pointer to the collector and a client node that are used by clients.
func DefineJaegerCollector(wiring blueprint.WiringSpec, collectorName string) string {
	collectorAddr := collectorName + ".addr"
	collectorProc := collectorName + ".proc"
	collectorDst := collectorName + ".dst"
	collectorClient := collectorName + ".client"

	wiring.Define(collectorProc, &JaegerCollector{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*JaegerCollector](ns, collectorAddr)
		if err != nil {
			return nil, err
		}

		return newJaegerCollector(collectorProc, addr.Bind)
	})

	wiring.Alias(collectorDst, collectorProc)
	pointer.RequireUniqueness(wiring, collectorDst, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, collectorName, &JaegerCollectorClient{}, collectorDst)

	ptr := pointer.GetPointer(wiring, collectorName)

	ptr.AddDstModifier(wiring, collectorAddr)

	clientNext := ptr.AddSrcModifier(wiring, collectorClient)

	wiring.Define(collectorClient, &JaegerCollectorClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*JaegerCollector](ns, clientNext)
		if err != nil {
			return nil, err
		}

		return newJaegerCollectorClient(collectorClient, addr.Dial)
	})

	address.Define[*JaegerCollector](wiring, collectorAddr, collectorProc, &blueprint.ApplicationNode{})
	ptr.AddDstModifier(wiring, collectorAddr)

	return collectorName
}
