// Package redis provides the Blueprint wiring and IR implementations of a redis plugin that
// provides a Cache interface implementation via a pre-built redis container image.
//
// Usage: To add a redis container named `fooCache`
//
//	PrebuiltContainer(spec, "fooCache")
package redis

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/pointer"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
)

// Adds a redis container to the application that defines a cache called `cacheName` which uses
// the pre-built redis process container
func Container(spec wiring.WiringSpec, cacheName string) string {
	// The nodes that we are defining
	ctrName := cacheName + ".ctr"
	clientName := cacheName + ".client"
	addrName := cacheName + ".addr"

	// Define the Redis container
	spec.Define(ctrName, &RedisContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		redis, err := newRedisContainer(ctrName)
		if err != nil {
			return nil, err
		}

		err = address.Bind[*RedisContainer](ns, addrName, redis, &redis.BindAddr)
		return redis, err
	})

	// Create a pointer to the Redis container
	ptr := pointer.CreatePointer[*RedisGoClient](spec, cacheName, ctrName)

	// Define the address that points to the Redis container
	address.Define[*RedisContainer](spec, addrName, ctrName)

	// Add the address to the pointer
	ptr.AddAddrModifier(spec, addrName)

	// Define the Redis client and add it to the client side of the pointer
	clientNext := ptr.AddSrcModifier(spec, clientName)
	spec.Define(clientName, &RedisGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*RedisContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newRedisGoClient(clientName, addr.Dial)
	})

	// Return the pointer; anybody who wants to access the Redis instance should do so through the pointer
	return cacheName
}
