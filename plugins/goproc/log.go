package goproc

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type stdoutLogger struct {
	golang.Node
	golang.Instantiable

	LoggerName  string
	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newStdoutLogger(name string) (*stdoutLogger, error) {
	node := &stdoutLogger{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.LoggerName = name
	return node, nil
}

func (node *stdoutLogger) init(name string) error {
	workflow.Init("../../runtime")
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("SLogger")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *stdoutLogger) Name() string {
	return node.LoggerName
}

func (node *stdoutLogger) String() string {
	return node.Name() + " = SLogger()"
}

func (node *stdoutLogger) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *stdoutLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *stdoutLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *stdoutLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating SLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Constructor, []ir.IRNode{})
}

func (node *stdoutLogger) ImplementsGolangNode() {}
