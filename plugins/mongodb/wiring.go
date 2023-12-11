// Package mongodb provides a plugin to generate and include a mongodb instance in a Blueprint application.
//
// The package provides a built-in mongodb container that provides the server-side implementation
// and a go-client for connecting to the client.
//
// The applications must use a backend.NoSQLDatabase (runtime/core/backend) as the interface in the application workflow.
package mongodb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Container generates the IRNodes for a mongodb server docker container that uses the latest mongodb image
// and the clients needed by the generated application to communicate with the server.
//
// The generated container has the name `dbName`.
func Container(spec wiring.WiringSpec, dbName string) string {
	procName := dbName + ".process"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"

	spec.Define(procName, &MongoDBContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*MongoDBContainer](ns, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newMongoDBContainer(procName, addr.Bind)
	})

	dstName := dbName + ".dst"
	spec.Alias(dstName, procName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, dbName, &MongoDBGoClient{}, dstName)
	ptr := pointer.GetPointer(spec, dbName)

	address.Define[*MongoDBContainer](spec, addrName, procName, &ir.ApplicationNode{})

	ptr.AddDstModifier(spec, addrName)

	clientNext := ptr.AddSrcModifier(spec, clientName)

	spec.Define(clientName, &MongoDBGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MongoDBContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		return newMongoDBGoClient(clientName, addr.Dial)
	})

	return dbName
}
