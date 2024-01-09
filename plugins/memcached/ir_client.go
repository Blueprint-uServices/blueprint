package memcached

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/backend"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to a memcached container
type MemcachedGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	DialAddr     *address.DialConfig

	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newMemcachedGoClient(name string, addr *address.DialConfig) (*MemcachedGoClient, error) {
	client := &MemcachedGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.DialAddr = addr
	return client, nil
}

func (n *MemcachedGoClient) String() string {
	return n.InstanceName + " = MemcachedClient(" + n.DialAddr.Name() + ")"
}

func (n *MemcachedGoClient) Name() string {
	return n.InstanceName
}

func (node *MemcachedGoClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("Memcached")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()

	return nil
}

func (n *MemcachedGoClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (node *MemcachedGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *MemcachedGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

// Part of code generation compilation pass; provides instantiation snippet
func (node *MemcachedGoClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MemcachedClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{node.DialAddr})
}

func (node *MemcachedGoClient) ImplementsGolangNode()    {}
func (node *MemcachedGoClient) ImplementsGolangService() {}
