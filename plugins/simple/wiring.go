// Package simple provides a Blueprint plugin for using in-memory backend implementations, e.g. in-memory databases and queues.
//
// In most compiled applications it is preferred to use "proper" implementations such as MySQL, Kafka, MongoDB etc.  However,
// for testing and for all-in-one processes, using the "simple" backend implementations provided here is highly convenient.
package simple

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines an in-memory [backend.NoSQLDatabase] instance with the specified name.
// In the compiled application, uses the [simplenosqldb.SimpleNoSQLDB] implementation from the Blueprint runtime package
func NoSQLDB(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "NoSQLDatabase", "SimpleNoSQLDB")
}

// Defines an in-memory [backend.RelationalDB] instance with the specified name.
// In the compiled application, uses the [simplereldb.SimpleRelationalDB] implementation from the Blueprint runtime package
func RelationalDB(spec wiring.WiringSpec, name string) string {
	return define(spec, name, "RelationalDB", "SimpleRelationalDB")
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
