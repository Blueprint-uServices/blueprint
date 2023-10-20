package redis

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
)

// Defines a cache called `cacheName` that uses the pre-built redis image
func PrebuiltProcess(wiring blueprint.WiringSpec, cacheName string) string {
	procName := cacheName + ".process"
	clientName := cacheName + ".client"
	addrName := cacheName + ".addr"

	wiring.Define(procName, &RedisProcess{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Bind[*RedisProcess](ns, addrName)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newRedisProcess(procName, addr.Bind)
	})

	dstName := cacheName + ".dst"
	wiring.Alias(dstName, procName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, cacheName, &RedisGoClient{}, dstName)
	ptr := pointer.GetPointer(wiring, cacheName)

	address.Define[*RedisProcess](wiring, addrName, procName, &blueprint.ApplicationNode{})

	ptr.AddDstModifier(wiring, addrName)

	clientNext := ptr.AddSrcModifier(wiring, clientName)

	wiring.Define(clientName, &RedisGoClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		addr, err := address.Dial[*RedisProcess](ns, clientNext)
		if err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newRedisGoClient(clientName, addr.Dial)
	})

	return cacheName
}
