// Package simple provides basic in-memory implementations of various Blueprint backends including Cache, Queue, NoSQLDB, and RelationalDB.
//
// These simple in-memory implementations are useful when compiling all-in-one applications, and for use during development and testing
// of workflow specs.
//
// For a more fully-fledged microservice application, these simple backends are a poor choice; instead a "proper" implementation
// such as MySQL, Kafka, MongoDB etc. should be used.
package simple

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines an in-memory [backend.NoSQLDatabase] instance with the specified name.
// In the compiled application, uses the [simplenosqldb.SimpleNoSQLDB] implementation from the Blueprint runtime package
// The SimpleNoSQLDB has limited support for query and update operations.
func NoSQLDB(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "NoSQLDatabase", "SimpleNoSQLDB")
}

// Defines an in-memory [backend.RelationalDB] instance with the specified name.
// In the compiled application, uses the [sqlitereldb.SqliteRelDB] implementation from the Blueprint runtime package
// The compiled application might fail to run if gcc is not installed and CGO_ENABLED is not set.
func RelationalDB(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "RelationalDB", "SqliteRelDB")
}

// Defines an in-memory [backend.Queue] instance with the specified name.
// In the compiled application, uses the [simplequeue.SimpleQueue] implementation from the Blueprint runtime package
func Queue(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "Queue", "SimpleQueue")
}

// Defines an in-memory [backend.Cache] instance with the specified name.
// In the compiled application, uses the [simplecache.SimpleCache] implementation from the Blueprint runtime package
func Cache(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "Cache", "SimpleCache")
}

func define(spec wiring.WiringSpec, name, backendType, backendImpl string) string {
	backendName := name + ".backend"
	spec.Define(backendName, &SimpleBackend{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newSimpleBackend(name, backendType, backendImpl)
	})

	// Mandate that this backend with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := name + ".dst"
	spec.Alias(dstName, backendName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(spec, name, &SimpleBackend{}, dstName)

	return name
}
