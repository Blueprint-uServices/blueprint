package memcached

import (
	"fmt"
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
)

type MemcachedGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	Addr         *MemcachedAddr
}

func newMemcachedGoClient(name string, addr blueprint.IRNode) (*MemcachedGoClient, error) {
	addrNode, is_addr := addr.(*MemcachedAddr)
	if !is_addr {
		return nil, fmt.Errorf("%s expected %s to be an address but found %s", name, addr.Name(), reflect.TypeOf(addr).String())
	}

	client := &MemcachedGoClient{}
	client.InstanceName = name
	client.Addr = addrNode
	return client, nil
}

func (n *MemcachedGoClient) String() string {
	return n.InstanceName + " = MemcachedClient(" + n.Addr.Name() + ")"
}

func (n *MemcachedGoClient) Name() string {
	return n.InstanceName
}

func (n *MemcachedGoClient) GetInterface() *service.ServiceInterface {
	// TODO: return memcached interface
	return nil
}

func (node *MemcachedGoClient) AddInstantiation(builder golang.DICodeBuilder) error {
	// TODO
	return nil
}

func (node *MemcachedGoClient) ImplementsGolangNode()    {}
func (node *MemcachedGoClient) ImplementsGolangService() {}
