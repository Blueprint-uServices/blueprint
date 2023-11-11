package memcached

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

/*
Defines a cache called `cacheName` that uses the pre-built memcached process image
*/
func PrebuiltContainer(spec wiring.WiringSpec, cacheName string) string {
	// The nodes that we are defining
	procName := cacheName + ".process"
	clientName := cacheName + ".client"
	addrName := cacheName + ".addr"

	// First define the process
	spec.Define(procName, &MemcachedContainer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*MemcachedContainer](namespace, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newMemcachedContainer(procName, addr.Bind)
	})

	// Mandate that this cache with this name must be unique within the application (although, this can be changed by namespaces)
	dstName := cacheName + ".dst"
	spec.Alias(dstName, procName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	// Define the pointer to the memcached process for golang clients
	pointer.CreatePointer(spec, cacheName, &MemcachedGoClient{}, dstName)
	ptr := pointer.GetPointer(spec, cacheName)

	// Define the address and add the collectorAddr to the pointer dst
	address.Define[*MemcachedContainer](spec, addrName, procName, &ir.ApplicationNode{})
	ptr.AddDstModifier(spec, addrName)

	// Add the client to the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)

	// Define the memcached go client
	spec.Define(clientName, &MemcachedGoClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MemcachedContainer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newMemcachedGoClient(clientName, addr.Dial)
	})

	return cacheName
}
