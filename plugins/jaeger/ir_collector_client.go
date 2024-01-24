package jaeger

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/jaeger"
	"golang.org/x/exp/slog"
)

// Blueprint IR node representing a client to the jaeger container
type JaegerCollectorClient struct {
	golang.Node
	service.ServiceNode
	golang.Instantiable
	ClientName string
	ServerDial *address.DialConfig

	InstanceName string
	Spec         *workflowspec.Service
}

func newJaegerCollectorClient(name string, addr *address.DialConfig) (*JaegerCollectorClient, error) {
	spec, err := workflowspec.GetService[jaeger.JaegerTracer]()
	if err != nil {
		return nil, err
	}

	node := &JaegerCollectorClient{
		InstanceName: name,
		ClientName:   name,
		ServerDial:   addr,
		Spec:         spec,
	}
	return node, nil
}

// Implements ir.IRNode
func (node *JaegerCollectorClient) Name() string {
	return node.ClientName
}

// Implements ir.IRNode
func (node *JaegerCollectorClient) String() string {
	return node.Name() + " = JaegerClient(" + node.ServerDial.Name() + ")"
}

// Implements golang.Instantiable
func (node *JaegerCollectorClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating JaegerClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{node.ServerDial})
}

// Implements service.ServiceNode
func (node *JaegerCollectorClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.ProvidesInterface
func (node *JaegerCollectorClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements golang.ProvidesModule
func (node *JaegerCollectorClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

func (node *JaegerCollectorClient) ImplementsGolangNode() {}

func (node *JaegerCollectorClient) ImplementsOTCollectorClient() {}
