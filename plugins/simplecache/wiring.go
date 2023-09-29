package simplecache

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

/* Creates a simple cache instance with the specified name */
func Define(wiring blueprint.WiringSpec, cacheName string) string {
	// Define the cache backend
	backendName := cacheName + ".backend"
	wiring.Define(backendName, &SimpleCache{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		return newSimpleCache(cacheName)
	})

	// Mandate that this service with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := cacheName + ".dst"
	wiring.Alias(dstName, backendName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	// Define the pointer
	pointer.CreatePointer(wiring, cacheName, &SimpleCache{}, dstName)

	return cacheName
}
