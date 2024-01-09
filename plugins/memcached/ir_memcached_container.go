package memcached

import (
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/docker"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/goparser"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
)

// Blueprint IR Node that represents a memcached container
type MemcachedContainer struct {
	backend.Cache
	docker.Container

	InstanceName string
	BindAddr     *address.BindConfig
	Iface        *goparser.ParsedInterface
}

// Memcached interface exposed to other services.
// This interface can not be modified further.
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

func newMemcachedContainer(name string) (*MemcachedContainer, error) {
	proc := &MemcachedContainer{}
	proc.InstanceName = name
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
