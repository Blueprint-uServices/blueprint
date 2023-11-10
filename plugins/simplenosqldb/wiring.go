package simplenosqldb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

/* Creates a simple nosqldb instance with the specified name */
func Define(spec wiring.WiringSpec, dbName string) string {
	// Define the nosqldb backend
	backendName := dbName + ".backend"
	spec.Define(backendName, &SimpleNoSQLDB{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newSimpleNoSQLDB(dbName)
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := dbName + ".dst"
	spec.Alias(dstName, backendName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(spec, dbName, &SimpleNoSQLDB{}, dstName)

	return dbName
}
