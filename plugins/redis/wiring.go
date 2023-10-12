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
		var addr *RedisAddr
		if err := ns.Get(addrName, &addr); err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", procName, addrName, err)
		}
		return newRedisProcess(procName, addr)
	})

	dstName := cacheName + ".dst"
	wiring.Alias(dstName, procName)
	pointer.RequireUniqueness(wiring, dstName, &blueprint.ApplicationNode{})

	pointer.CreatePointer(wiring, cacheName, &RedisGoClient{}, dstName)
	ptr := pointer.GetPointer(wiring, cacheName)

	address.Define(wiring, addrName, procName, &blueprint.ApplicationNode{}, func(ns blueprint.Namespace) (address.Address, error) {
		addr := &RedisAddr{
			AddrName: addrName,
			Server:   nil,
		}
		return addr, nil
	})

	ptr.AddDstModifier(wiring, addrName)

	clientNext := ptr.AddSrcModifier(wiring, clientName)

	wiring.Define(clientName, &RedisGoClient{}, func(ns blueprint.Namespace) (blueprint.IRNode, error) {
		var addr *RedisAddr
		if err := ns.Get(clientNext, &addr); err != nil {
			return nil, blueprint.Errorf("%s expected %s to be an address but encountered %s", clientName, clientNext, err)
		}
		return newRedisGoClient(clientName, addr)
	})

	return cacheName
}
