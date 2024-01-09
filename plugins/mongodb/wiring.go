// Package mongodb provides a plugin to generate and include a mongodb instance in a Blueprint application.
//
// The package provides a built-in mongodb container that provides the server-side implementation
// and a go-client for connecting to the client.
//
// The applications must use a backend.NoSQLDatabase (runtime/core/backend) as the interface in the application workflow.
package mongodb

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Container generates the IRNodes for a mongodb server docker container that uses the latest mongodb image
// and the clients needed by the generated application to communicate with the server.
//
// The generated container has the name `dbName`.
func Container(spec wiring.WiringSpec, dbName string) string {
	// The nodes that we are defining
	ctrName := dbName + ".ctr"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"

	// Define the MongoDB container
	spec.Define(ctrName, &MongoDBContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		ctr, err := newMongoDBContainer(ctrName)
		if err != nil {
			return nil, err
		}
		err = address.Bind[*MongoDBContainer](ns, addrName, ctr, &ctr.BindAddr)
		return ctr, err
	})

	// Create a pointer to the MongoDB container
	ptr := pointer.CreatePointer[*MongoDBGoClient](spec, dbName, ctrName)

	// Define the address that points to the MongoDB container
	address.Define[*MongoDBContainer](spec, addrName, ctrName)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, addrName)

	// Define the MongoDB client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &MongoDBGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MongoDBContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		return newMongoDBGoClient(clientName, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the MongoDB instance should do so through the pointer
	return dbName
}
