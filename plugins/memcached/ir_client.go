package memcached

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type MemcachedGoClient struct {
	golang.Service
	backend.Cache

	InstanceName string
	Addr         *MemcachedAddr

	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newMemcachedGoClient(name string, addr *MemcachedAddr) (*MemcachedGoClient, error) {
	client := &MemcachedGoClient{}
	err := client.init(name)
	if err != nil {
		return nil, err
	}
	client.InstanceName = name
	client.Addr = addr
	return client, nil
}

func (n *MemcachedGoClient) String() string {
	return n.InstanceName + " = MemcachedClient(" + n.Addr.Name() + ")"
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

func (n *MemcachedGoClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return n.Iface.ServiceInterface(ctx), nil
}

func (node *MemcachedGoClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *MemcachedGoClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

// Part of code generation compilation pass; provides instantiation snippet
func (node *MemcachedGoClient) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating MemcachedClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []blueprint.IRNode{node.Addr})
}

func (node *MemcachedGoClient) ImplementsGolangNode()    {}
func (node *MemcachedGoClient) ImplementsGolangService() {}
