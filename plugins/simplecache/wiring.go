package simplecache

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

/* Creates a simple cache instance with the specified name */
func Define(spec wiring.WiringSpec, cacheName string) string {
	// Define the cache backend
	backendName := cacheName + ".backend"
	spec.Define(backendName, &SimpleCache{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		return newSimpleCache(cacheName)
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := cacheName + ".dst"
	spec.Alias(dstName, backendName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(spec, cacheName, &SimpleCache{}, dstName)

	return cacheName
}
