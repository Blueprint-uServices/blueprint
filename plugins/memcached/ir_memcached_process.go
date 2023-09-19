package memcached

import (
	"reflect"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
)

type MemcachedProcess struct {
	process.ProcessNode
	backend.Cache
	process.ArtifactGenerator

	InstanceName string
	Addr         *MemcachedAddr
}

func newMemcachedProcess(name string, addr blueprint.IRNode) (*MemcachedProcess, error) {
	addrNode, is_addr := addr.(*MemcachedAddr)
	if !is_addr {
		return nil, blueprint.Errorf("%s expected %s to be an address but found %s", name, addr.Name(), reflect.TypeOf(addr).String())
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

func (n *MemcachedProcess) GenerateArtifacts(outputDir string) error {
	// TODO: generate artifacts for the memcached process
	return nil
}
