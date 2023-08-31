package memcached

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type MemcachedProcess struct {
	process.ProcessNode
	backend.Cache
	// TODO: artifact generation

	InstanceName string
	Addr         *pointer.Address
}

func newMemcachedProcess(name string, addr blueprint.IRNode) (*MemcachedProcess, error) {
	addrNode, is_addr := addr.(*pointer.Address)
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
