package opentelemetry

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

// Blueprint IR Node that represents a process-level OT trace logger
type OTTraceLogger struct {
	golang.Node
	golang.Instantiable

	LoggerName  string
	Iface       *goparser.ParsedInterface
	Constructor *gocode.Constructor
}

func newOTTraceLogger(name string) (*OTTraceLogger, error) {
	node := &OTTraceLogger{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.LoggerName = name
	return node, nil
}

func (node *OTTraceLogger) init(name string) error {
	workflow.Init("../../runtime")
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("OTTraceLogger")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

// Implements ir.IRNode
func (node *OTTraceLogger) Name() string {
	return node.LoggerName
}

// Implements ir.IRNode
func (node *OTTraceLogger) String() string {
	return node.Name() + " = OTTraceLogger()"
}

// Implements golang.ProvidesModule
func (node *OTTraceLogger) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

// Implements golang.ProvidesInterface
func (node *OTTraceLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

// Implements golang.ProvidesInterface
func (node *OTTraceLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *OTTraceLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating OTTraceLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Constructor, []ir.IRNode{})
}

// Implements ir.IRNode
func (node *OTTraceLogger) ImplementsGolangNode() {}
