package mysql

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

func PrebuiltContainer(spec wiring.WiringSpec, dbName string, username string, password string) string {
	cntrName := dbName + ".container"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"

	spec.Define(cntrName, &MySQLDBContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*MySQLDBContainer](ns, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", cntrName, addrName, err)
		}

		return newMySQLDBContainer(cntrName, addr.Bind, username, password)
	})

	dstName := dbName + ".dst"
	spec.Alias(dstName, cntrName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, dbName, &MySQLDBGoClient{}, dstName)
	ptr := pointer.GetPointer(spec, dbName)

	address.Define[*MySQLDBContainer](spec, addrName, cntrName, &ir.ApplicationNode{})

	ptr.AddDstModifier(spec, addrName)

	clientNext := ptr.AddSrcModifier(spec, clientName)

	spec.Define(clientName, &MySQLDBGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MySQLDBContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		return newMySQLDBGoClient(clientName, addr.Dial, username, password)
	})

	return dbName
}
