package redis

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
)

// Defines a cache called `cacheName` that uses the pre-built redis image
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
