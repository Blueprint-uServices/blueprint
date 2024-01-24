package zipkin

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/zipkin"
	"golang.org/x/exp/slog"
)

// Blueprint IR node representing a client to the zipkin container
type ZipkinCollectorClient struct {
	golang.Node
	golang.Instantiable
	ClientName string
	ServerDial *address.DialConfig
	Spec       *workflowspec.Service
}

func newZipkinCollectorClient(name string, addr *address.DialConfig) (*ZipkinCollectorClient, error) {
	spec, err := workflowspec.GetService[zipkin.ZipkinTracer]()
	node := &ZipkinCollectorClient{
		ClientName: name,
		ServerDial: addr,
		Spec:       spec,
	}
	return node, err
}

// Implements ir.IRNode
func (node *ZipkinCollectorClient) Name() string {
	return node.ClientName
}

// Implements ir.IRNode
func (node *ZipkinCollectorClient) String() string {
	return node.Name() + " = ZipkinClient(" + node.ServerDial.Name() + ")"
}

// Implements golang.Instantiable
func (node *ZipkinCollectorClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating ZipkinClient %v in %v/%v", node.ClientName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.ClientName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{node.ServerDial})
}

// Implements service.ServiceNode
func (node *ZipkinCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesInterface
func (node *ZipkinCollectorClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements golang.ProvidesModule
func (node *ZipkinCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

func (node *ZipkinCollectorClient) ImplementsGolangNode() {}

func (node *ZipkinCollectorClient) ImplementsOTCollectorClient() {}
