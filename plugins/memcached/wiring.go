// Package memcached provides the Blueprint wiring and IR implementations of a memcached plugin that
// provides a Cache interface implementation via a pre-built memcached container image.
//
// Usage: To add a memcached container named `fooCache`
//
//	PrebuiltContainer(spec, "fooCache")
package memcached

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Adds a memcached container to the application that defines a cache called `cacheName`
// which uses the pre-built memcached process container
func Container(spec wiring.WiringSpec, cacheName string) string {
	// The nodes that we are defining
	ctrName := cacheName + ".ctr"
	addrName := cacheName + ".addr"
	clientName := cacheName + ".client"

	// Define the Memcached container
	spec.Define(ctrName, &MemcachedContainer{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		ctr, err := newMemcachedContainer(ctrName)
		if err != nil {
			return nil, err
		}
		err = address.Bind[*MemcachedContainer](namespace, addrName, ctr, &ctr.BindAddr)
		return ctr, err
	})

	// Create a pointer to the Memcached container
	ptr := pointer.CreatePointer[*MemcachedGoClient](spec, cacheName, ctrName)

	// Define the address that points to the Memcached container
	address.Define[*MemcachedContainer](spec, addrName, ctrName)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, addrName)

	// Define the memcached client add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &MemcachedGoClient{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*MemcachedContainer](namespace, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newMemcachedGoClient(clientName, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the Memcached instance should do so through the pointer
	return cacheName
}
