// Package mysql provides a plugin to generate and include a mysql instance in a Blueprint application.
//
// The package provides a built-in mysql container that provides the server-side implementation
// and a go-client for connecting to the client.
//
// The applications must use a backend.RelationalDB (runtime/core/backend) as the interface in the workflow.
package mysql

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Container generate the IRNodes for a mysql server docker container that uses the latest mysql/mysql image
// and the clients needed by the generated application to communicate with the server.
//
// The generated container has the name `dbName` with the root password set to `password`.
func Container(spec wiring.WiringSpec, dbName string) string {
	cntrName := dbName + ".container"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"
	username := "root"
	password := "pass"

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

		user_val := &ir.IRValue{Value: username}
		pwd_val := &ir.IRValue{Value: password}
		db_val := &ir.IRValue{Value: dbName}

		return newMySQLDBGoClient(clientName, addr.Dial, user_val, pwd_val, db_val)
	})

	return dbName
}
