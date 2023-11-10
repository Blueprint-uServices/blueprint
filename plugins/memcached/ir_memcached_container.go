package memcached

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type MemcachedContainer struct {
	backend.Cache
	docker.Container

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}

type MemcachedInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (m *MemcachedInterface) GetName() string {
	return "memcached(" + m.Wrapped.GetName() + ")"
}

func (m *MemcachedInterface) GetMethods() []service.Method {
	return m.Wrapped.GetMethods()
}

func newMemcachedContainer(name string, addr *address.BindConfig) (*MemcachedContainer, error) {
	proc := &MemcachedContainer{}
	proc.InstanceName = name
	proc.BindAddr = addr
	err := proc.init(name)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (node *MemcachedContainer) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("Memcached")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	return nil
}

func (n *MemcachedContainer) String() string {
	return n.InstanceName + " = MemcachedProcess(" + n.BindAddr.Name() + ")"
}

func (n *MemcachedContainer) Name() string {
	return n.InstanceName
}

func (node *MemcachedContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &MemcachedInterface{Wrapped: iface}, nil
}

func (node *MemcachedContainer) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	return nil
}

func (node *MemcachedContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	instanceName := ir.CleanName(node.InstanceName)

	node.BindAddr.Hostname = instanceName
	node.BindAddr.Port = 11211

	return target.DeclarePrebuiltInstance(node.InstanceName, "memcached", node.BindAddr)
}
