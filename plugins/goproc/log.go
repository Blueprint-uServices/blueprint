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

// Does not implement golang.ProvidesModule or golang.ProvidesInterface
// because the stdout logger is implemented in the Blueprint runtime package
// which is already included in the output by default.

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
