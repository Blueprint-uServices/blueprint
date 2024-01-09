package jaeger

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR node representing a client to the jaeger container
type JaegerCollectorClient struct {
	golang.Node
	golang.Instantiable
	ClientName string
	ServerDial *address.DialConfig

	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newJaegerCollectorClient(name string, addr *address.DialConfig) (*JaegerCollectorClient, error) {
	node := &JaegerCollectorClient{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.ClientName = name
	node.ServerDial = addr
	return node, nil
}

func (node *JaegerCollectorClient) Name() string {
	return node.ClientName
}

func (node *JaegerCollectorClient) String() string {
	return node.Name() + " = JaegerClient(" + node.ServerDial.Name() + ")"
}

func (node *JaegerCollectorClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("JaegerTracer")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *JaegerCollectorClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating JaegerClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{node.ServerDial})
}

func (node *JaegerCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *JaegerCollectorClient) AddInterfaces(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *JaegerCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *JaegerCollectorClient) ImplementsGolangNode() {}

func (node *JaegerCollectorClient) ImplementsOTCollectorClient() {}
