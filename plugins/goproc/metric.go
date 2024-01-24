package goproc

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/opentelemetry"
	"golang.org/x/exp/slog"
)

type stdoutMetricCollector struct {
	golang.Node
	service.ServiceNode
	golang.Instantiable

	CollectorName string
	Spec          *workflowspec.Service
}

func newStdOutMetricCollector(name string) (*stdoutMetricCollector, error) {
	spec, err := workflowspec.GetService[opentelemetry.StdoutMetricCollector]()
	node := &stdoutMetricCollector{
		CollectorName: name,
		Spec:          spec,
	}
	return node, err
}

// Implements ir.IRNode
func (node *stdoutMetricCollector) Name() string {
	return node.CollectorName
}

// Implements ir.IRNode
func (node *stdoutMetricCollector) String() string {
	return node.Name() + " = StdoutMetricCollector()"
}

// Does not implement golang.ProvidesModule or golang.ProvidesInterface
// because the stdout logger is implemented in the Blueprint runtime package
// which is already included in the output by default.

// Implements service.ServiceNode
func (node *stdoutMetricCollector) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *stdoutMetricCollector) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.CollectorName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating StdoutMetricCollector %v in %v/%v", node.CollectorName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.CollectorName, node.Spec.Constructor.AsConstructor(), []ir.IRNode{})
}

func (node *stdoutMetricCollector) ImplementsGolangNode() {}
