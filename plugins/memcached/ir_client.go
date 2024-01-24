package memcached

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/memcached"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to a memcached container
type MemcachedGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	DialAddr     *address.DialConfig

	Spec *workflowspec.Service
}

func newMemcachedGoClient(name string, addr *address.DialConfig) (*MemcachedGoClient, error) {
	spec, err := workflowspec.GetService[memcached.Memcached]()
	client := &MemcachedGoClient{
		InstanceName: name,
		DialAddr:     addr,
		Spec:         spec,
	}
	return client, err
}

// Implements ir.IRNode
func (n *MemcachedGoClient) String() string {
	return n.InstanceName + " = MemcachedClient(" + n.DialAddr.Name() + ")"
}

// Implements ir.IRNode
func (n *MemcachedGoClient) Name() string {
	return n.InstanceName
}

// Implements service.ServiceNode
func (n *MemcachedGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesModule
func (node *MemcachedGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *MemcachedGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements golang.Instantiable
func (node *MemcachedGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MemcachedClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{node.DialAddr})
}

func (node *MemcachedGoClient) ImplementsGolangNode()    {}
func (node *MemcachedGoClient) ImplementsGolangService() {}
