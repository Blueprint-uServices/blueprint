package goproc

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/slogger"
	"golang.org/x/exp/slog"
)

type stdoutLogger struct {
	golang.Node
	service.ServiceNode
	golang.Instantiable

	LoggerName string
	Spec       *workflowspec.Service
}

func newStdoutLogger(name string) (*stdoutLogger, error) {
	spec, err := workflowspec.GetService[slogger.SLogger]()
	node := &stdoutLogger{
		LoggerName: name,
		Spec:       spec,
	}
	return node, err
}

// Implements ir.IRNode
func (node *stdoutLogger) Name() string {
	return node.LoggerName
}

// Implements ir.IRNode
func (node *stdoutLogger) String() string {
	return node.Name() + " = SLogger()"
}

// Implements golang.ProvidesModule
func (node *stdoutLogger) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *stdoutLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements service.ServiceNode
func (node *stdoutLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *stdoutLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating SLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{})
}

func (node *stdoutLogger) ImplementsGolangNode() {}
