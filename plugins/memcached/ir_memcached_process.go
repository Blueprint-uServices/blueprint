package memcached

import (
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

func newMemcachedProcess(name string, addr *MemcachedAddr) (*MemcachedProcess, error) {
	proc := &MemcachedProcess{}
	proc.InstanceName = name
	proc.Addr = addr
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
