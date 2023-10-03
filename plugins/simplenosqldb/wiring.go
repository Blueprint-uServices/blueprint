package simplenosqldb

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

/* Creates a simple nosqldb instance with the specified name */
func Define(wiring blueprint.WiringSpec, dbName string) string {
	// Define the nosqldb backend
	backendName := dbName + ".backend"
	wiring.Define(backendName, &SimpleNoSQLDB{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		return newSimpleNoSQLDB(dbName)
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := dbName + ".dst"
	wiring.Alias(dstName, backendName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(wiring, dbName, &SimpleNoSQLDB{}, dstName)

	return dbName
}
