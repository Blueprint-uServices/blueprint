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

// Blueprint IR Node that represents a process-level xtrace logger
type XTraceLogger struct {
	golang.Node
	golang.Instantiable

	ServerDialAddr *address.DialConfig

	LoggerName  string
	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newXTraceLogger(name string, addr *address.DialConfig) (*XTraceLogger, error) {
	node := &XTraceLogger{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.LoggerName = name
	node.ServerDialAddr = addr
	return node, nil
}

func (node *XTraceLogger) init(name string) error {
	workflow.Init("../../runtime")
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("XTraceLogger")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *XTraceLogger) Name() string {
	return node.LoggerName
}

func (node *XTraceLogger) String() string {
	return node.Name() + " = XTraceLogger(" + node.ServerDialAddr.Name() + ")"
}

func (node *XTraceLogger) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *XTraceLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *XTraceLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *XTraceLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating XTraceLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Constructor, []ir.IRNode{node.ServerDialAddr})
}

func (node *XTraceLogger) ImplementsGolangNode() {}
