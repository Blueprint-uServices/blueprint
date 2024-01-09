package zipkin

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

// Blueprint IR node representing a client to the zipkin container
type ZipkinCollectorClient struct {
	golang.Node
	golang.Instantiable
	ClientName string
	ServerDial *address.DialConfig

	InstanceName string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newZipkinCollectorClient(name string, addr *address.DialConfig) (*ZipkinCollectorClient, error) {
	node := &ZipkinCollectorClient{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.ClientName = name
	node.ServerDial = addr
	return node, nil
}

func (node *ZipkinCollectorClient) Name() string {
	return node.ClientName
}

func (node *ZipkinCollectorClient) String() string {
	return node.Name() + " = ZipkinClient(" + node.ServerDial.Name() + ")"
}

func (node *ZipkinCollectorClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("ZipkinTracer")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *ZipkinCollectorClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating ZipkinClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{node.ServerDial})
}

func (node *ZipkinCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *ZipkinCollectorClient) AddInterfaces(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *ZipkinCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *ZipkinCollectorClient) ImplementsGolangNode() {}

func (node *ZipkinCollectorClient) ImplementsOTCollectorClient() {}
