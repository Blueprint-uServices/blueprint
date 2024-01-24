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

// Blueprint IR Node that represents a process-level xtrace logger
type XTraceLogger struct {
	golang.Node
	golang.Instantiable

	ServerDialAddr *address.DialConfig

	LoggerName string
	Spec       *workflowspec.Service
}

func newXTraceLogger(name string, addr *address.DialConfig) (*XTraceLogger, error) {
	spec, err := workflowspec.GetService[xtrace.XTraceLogger]()
	node := &XTraceLogger{
		LoggerName:     name,
		ServerDialAddr: addr,
		Spec:           spec,
	}
	return node, err
}

// Implements ir.IRNode
func (node *XTraceLogger) Name() string {
	return node.LoggerName
}

// Implements ir.IRNode
func (node *XTraceLogger) String() string {
	return node.Name() + " = XTraceLogger(" + node.ServerDialAddr.Name() + ")"
}

// Implements golang.ProvidesModule
func (node *XTraceLogger) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *XTraceLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements service.ServiceNode
func (node *XTraceLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *XTraceLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating XTraceLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{node.ServerDialAddr})
}

func (node *XTraceLogger) ImplementsGolangNode() {}
