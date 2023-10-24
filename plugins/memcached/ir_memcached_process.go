package memcached

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
)

type MemcachedProcess struct {
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

func newMemcachedProcess(name string, addr *address.BindConfig) (*MemcachedProcess, error) {
	proc := &MemcachedProcess{}
	proc.InstanceName = name
	proc.BindAddr = addr
	err := proc.init(name)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func (node *MemcachedProcess) init(name string) error {
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

func (n *MemcachedProcess) String() string {
	return n.InstanceName + " = MemcachedProcess(" + n.BindAddr.Name() + ")"
}

func (n *MemcachedProcess) Name() string {
	return n.InstanceName
}

func (node *MemcachedProcess) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &MemcachedInterface{Wrapped: iface}, nil
}

func (node *MemcachedProcess) AddContainerArtifacts(target docker.ContainerWorkspace) error {
	return nil
}

func (node *MemcachedProcess) AddContainerInstance(target docker.ContainerWorkspace) error {
	return target.DeclarePrebuiltInstance(node.InstanceName, "memcached", node.BindAddr)
}
