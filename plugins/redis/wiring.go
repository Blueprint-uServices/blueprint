// Package memcached provides the Blueprint wiring and IR implementations of a memcached plugin that
// provides a Cache interface implementation via a pre-built redis container image.
//
// Usage: To add a redis container named `fooCache`
//   PrebuiltContainer(spec, "fooCache")
package redis

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Adds a redis container to the application that defines a cache called `cacheName` which uses the pre-built memcached process container
func PrebuiltContainer(spec wiring.WiringSpec, cacheName string) string {
	procName := cacheName + ".process"
	clientName := cacheName + ".client"
	addrName := cacheName + ".addr"

	spec.Define(procName, &RedisContainer{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Bind[*RedisContainer](ns, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newRedisContainer(procName, addr.Bind)
	})

	dstName := cacheName + ".dst"
	spec.Alias(dstName, procName)
	pointer.RequireUniqueness(spec, dstName, &ir.ApplicationNode{})

	pointer.CreatePointer(spec, cacheName, &RedisGoClient{}, dstName)
	ptr := pointer.GetPointer(spec, cacheName)

	address.Define[*RedisContainer](spec, addrName, procName, &ir.ApplicationNode{})

	ptr.AddDstModifier(spec, addrName)

	clientNext := ptr.AddSrcModifier(spec, clientName)

	spec.Define(clientName, &RedisGoClient{}, func(ns wiring.Namespace) (ir.IRNode, error) {
		addr, err := address.Dial[*RedisContainer](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newRedisGoClient(clientName, addr.Dial)
	})

	return cacheName
}
