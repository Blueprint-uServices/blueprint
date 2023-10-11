package memcached

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type MemcachedAddr struct {
	address.Address
	AddrName string
	Server   *MemcachedProcess
}

func (addr *MemcachedAddr) Name() string {
	return addr.AddrName
}

func (addr *MemcachedAddr) String() string {
	return addr.AddrName + " = MemcachedAddr()"
}

func (addr *MemcachedAddr) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *MemcachedAddr) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(*MemcachedProcess)
	if !isServer {
		return blueprint.Errorf("address %v should point to a memcached server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *MemcachedAddr) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return addr.Server.GetInterface(ctx)
}

func (addr *MemcachedAddr) ImplementsAddressNode() {}
