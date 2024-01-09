// Package mysql provides a plugin to generate and include a mysql instance in a Blueprint application.
//
// The package provides a built-in mysql container that provides the server-side implementation
// and a go-client for connecting to the server.
//
// The applications must use a backend.RelationalDB (runtime/core/backend) as the interface in the workflow.
package mysql

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

var mysql_root_username = "root"
var mysql_root_password = "pass"

// Container generate the IRNodes for a mysql server docker container that uses the latest mysql/mysql image
// and the clients needed by the generated application to communicate with the server.
func Container(spec wiring.WiringSpec, dbName string) string {
	// The nodes that we are defining
	ctrName := dbName + ".ctr"
	clientName := dbName + ".client"
	addrName := dbName + ".addr"

	// Define the MySQL container
	spec.Define(ctrName, &MySQLDBContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		ctr, err := newMySQLDBContainer(ctrName, mysql_root_password)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*MySQLDBContainer](ns, addrName, ctr, &ctr.BindAddr)
		return ctr, err
	})

	// Create a pointer to the MySQL container
	ptr := pointer.CreatePointer[*MySQLDBGoClient](spec, dbName, ctrName)

	// Define the address that points to the MySQL container
	address.Define[*MySQLDBContainer](spec, addrName, ctrName)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, addrName)

	// Define the MySQL client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &MySQLDBGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MySQLDBContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}

		user_val := &ir.IRValue{Value: mysql_root_username}
		pwd_val := &ir.IRValue{Value: mysql_root_password}
		db_val := &ir.IRValue{Value: dbName}

		return newMySQLDBGoClient(clientName, addr.Dial, user_val, pwd_val, db_val)
	})

	return dbName
}
