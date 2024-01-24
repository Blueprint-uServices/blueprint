package memcached

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/memcached"
)

// Blueprint IR Node that represents a memcached container
type MemcachedContainer struct {
	backend.Cache
	docker.Container
	docker.ProvidesContainerInstance

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
	spec, err := workflowspec.GetService[memcached.Memcached]()
	if err != nil {
		return nil, err
	}

	proc := &MemcachedContainer{
		InstanceName: name,
		Iface:        spec.Iface,
	}
	return proc, nil
}

// Implements ir.IRNode
func (n *MemcachedContainer) String() string {
	return n.InstanceName + " = MemcachedProcess(" + n.BindAddr.Name() + ")"
}

// Implements ir.IRNode
func (n *MemcachedContainer) Name() string {
	return n.InstanceName
}

// Implements service.ServiceNode
func (node *MemcachedContainer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface := node.Iface.ServiceInterface(ctx)
	return &MemcachedInterface{Wrapped: iface}, nil
}

// Implements docker.ProvidesContainerInstance
func (node *MemcachedContainer) AddContainerInstance(target docker.ContainerWorkspace) error {
	instanceName := ir.CleanName(node.InstanceName)

	node.BindAddr.Hostname = instanceName
	node.BindAddr.Port = 11211

	return target.DeclarePrebuiltInstance(node.InstanceName, "memcached", node.BindAddr)
}
