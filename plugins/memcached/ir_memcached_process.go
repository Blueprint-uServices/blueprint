package memcached

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type MemcachedAddr struct {
	address.Address
	AddrName string
	Server   *MemcachedProcess
}

type MemcachedProcess struct {
	process.ProcessNode
	backend.Cache
	// TODO: artifact generation

	InstanceName string
	Addr         *MemcachedAddr
}

func newMemcachedProcess(name string, addr blueprint.IRNode) (*MemcachedProcess, error) {
	addrNode, is_addr := addr.(*MemcachedAddr)
	if !is_addr {
		return nil, fmt.Errorf("%s expected %s to be an address but found %s", name, addr.Name(), reflect.TypeOf(addr).String())
	}

	proc := &MemcachedProcess{}
	proc.InstanceName = name
	proc.Addr = addrNode
	return proc, nil
}

func (n *MemcachedProcess) String() string {
	return n.InstanceName + " = MemcachedProcess(" + n.Addr.Name() + ")"
}

func (n *MemcachedProcess) Name() string {
	return n.InstanceName
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
		return fmt.Errorf("address %v should point to a memcached server but got %v", addr.AddrName, node)
	}
	addr.Server = server
	return nil
}

func (addr *MemcachedAddr) ImplementsAddressNode() {}
