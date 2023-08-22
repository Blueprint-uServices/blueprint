package memcached

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

/*
Defines a cache called `cacheName` that uses the pre-built memcached process image
*/
func PrebuiltProcess(wiring blueprint.WiringSpec, cacheName string) string {
	// The nodes that we are defining
	procName := cacheName + ".process"
	clientName := cacheName + ".client"
	addrName := cacheName + ".addr"

	// First define the process
	wiring.Define(procName, &MemcachedProcess{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		addr, err := scope.Get(addrName)
		if err != nil {
			return nil, err
		}

		return newMemcachedProcess(procName, addr)
	})

	// Mandate that this cache with this name must be unique within the application (although, this can be changed by scopes)
	dstName := cacheName + ".dst"
	wiring.Alias(dstName, procName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	// Define the pointer to the memcached process for golang clients
	pointer.CreatePointer(wiring, cacheName, &MemcachedGoClient{}, dstName)
	ptr := pointer.GetPointer(wiring, cacheName)

	// Define the address and add the collectorAddr to the pointer dst
	pointer.DefineAddress(wiring, addrName, procName, &blueprint.ApplicationNode{})
	ptr.AddDstModifier(wiring, addrName)

	// Add the client to the pointer
	clientNext := ptr.AddSrcModifier(wiring, clientName)

	// Define the memcached go client
	wiring.Define(clientName, &MemcachedGoClient{}, func(scope blueprint.Scope) (blueprint.IRNode, error) {
		server, err := scope.Get(clientNext)
		if err != nil {
			return nil, err
		}

		return newMemcachedGoClient(clientName, server)
	})

	return cacheName
}
