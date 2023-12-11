package xtrace

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/address"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to the Xtrace container
type XTraceClient struct {
	golang.Node
	golang.Instantiable

	ClientName     string
	ServerDialAddr *address.DialConfig

	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newXTraceClient(name string, addr *address.DialConfig) (*XTraceClient, error) {
	node := &XTraceClient{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.ClientName = name
	node.ServerDialAddr = addr
	return node, nil
}

func (node *XTraceClient) Name() string {
	return node.ClientName
}

func (node *XTraceClient) String() string {
	return node.Name() + " = XTraceClient(" + node.ServerDialAddr.Name() + ")"
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

func (node *XTraceClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating XTraceClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{node.ServerDialAddr})
}

func (node *XTraceClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *XTraceClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *XTraceClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *XTraceClient) ImplementsGolangNode() {}
