package xtrace

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type XTraceClient struct {
	golang.Node
	golang.Instantiable

	ClientName string
	ServerAddr *address.Address[*XTraceServer]

	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newXTraceClient(name string, addr *address.Address[*XTraceServer]) (*XTraceClient, error) {
	node := &XTraceClient{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.ClientName = name
	node.ServerAddr = addr
	return node, nil
}

func (node *XTraceClient) Name() string {
	return node.ClientName
}

func (node *XTraceClient) String() string {
	return node.Name() + " = XTraceClient(" + node.ServerAddr.Name() + ")"
}

func (node *XTraceClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("XTracerImpl")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *XTraceClient) AddInstantiation(builder golang.GraphBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating XTraceClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []blueprint.IRNode{node.ServerAddr})
}

func (node *XTraceClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *XTraceClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *XTraceClient) GetInterface(ctx blueprint.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *XTraceClient) ImplementsGolangNode() {}
