package xtrace

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/xtrace"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to the Xtrace container
type XTraceClient struct {
	golang.Node
	golang.Instantiable

	ClientName     string
	ServerDialAddr *address.DialConfig
	Spec           *workflowspec.Service
}

func newXTraceClient(name string, addr *address.DialConfig) (*XTraceClient, error) {
	spec, err := workflowspec.GetService[xtrace.XTracerImpl]()
	node := &XTraceClient{
		ClientName:     name,
		ServerDialAddr: addr,
		Spec:           spec,
	}
	return node, err
}

// Implements ir.IRNode
func (node *XTraceClient) Name() string {
	return node.ClientName
}

// Implements ir.IRNode
func (node *XTraceClient) String() string {
	return node.Name() + " = XTraceClient(" + node.ServerDialAddr.Name() + ")"
}

// Implements golang.Instantiable
func (node *XTraceClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating XTraceClient %v in %v/%v", node.ClientName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.ClientName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{node.ServerDialAddr})
}

// Implements golang.ProvidesModule
func (node *XTraceClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *XTraceClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements service.ServiceNode
func (node *XTraceClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

func (node *XTraceClient) ImplementsGolangNode() {}
