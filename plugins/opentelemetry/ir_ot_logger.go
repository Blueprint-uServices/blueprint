package opentelemetry

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/opentelemetry"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a process-level OT trace logger
type OTTraceLogger struct {
	golang.Node
	golang.Instantiable

	LoggerName string
	Spec       *workflowspec.Service
}

func newOTTraceLogger(name string) (*OTTraceLogger, error) {
	spec, err := workflowspec.GetService[opentelemetry.OTTraceLogger]()
	node := &OTTraceLogger{
		LoggerName: name,
		Spec:       spec,
	}
	return node, err
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
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *OTTraceLogger) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements golang.ProvidesInterface
func (node *OTTraceLogger) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *OTTraceLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating OTTraceLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{})
}

// Implements ir.IRNode
func (node *OTTraceLogger) ImplementsGolangNode() {}
