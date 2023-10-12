package redis

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

type RedisAddr struct {
	address.Address
	AddrName string
	Server   *RedisProcess
}

func (addr *RedisAddr) Name() string {
	return addr.AddrName
}

func (addr *RedisAddr) String() string {
	return addr.AddrName + " = RedisAddr()"
}

func (addr *RedisAddr) GetDestination() blueprint.IRNode {
	if addr.Server != nil {
		return addr.Server
	}
	return nil
}

func (addr *RedisAddr) SetDestination(node blueprint.IRNode) error {
	server, isServer := node.(*RedisProcess)
	if !isServer {
		return blueprint.Errorf("address %v should point to a redis server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *RedisAddr) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return addr.Server.GetInterface(ctx)
}

func (addr *RedisAddr) ImplementsAddressNode() {}
