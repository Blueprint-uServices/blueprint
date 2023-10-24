package mongodb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

func PrebuiltProcess(wiring blueprint.WiringSpec, dbName string) string {
	procName := dbName + ".process"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"

	wiring.Define(procName, &MongoDBProcess{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*MongoDBProcess](ns, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newMongoDBProcess(procName, addr.Bind)
	})

	dstName := dbName + ".dst"
	wiring.Alias(dstName, procName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, dbName, &MongoDBGoClient{}, dstName)
	ptr := pointer.GetPointer(wiring, dbName)

	address.Define[*MongoDBProcess](wiring, addrName, procName, &blueprint.ApplicationNode{})

	ptr.AddDstModifier(wiring, addrName)

	clientNext := ptr.AddSrcModifier(wiring, clientName)

	wiring.Define(clientName, &MongoDBGoClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*MongoDBProcess](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		return newMongoDBGoClient(clientName, addr.Dial)
	})

	return dbName
}
