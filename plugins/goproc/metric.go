package goproc

import (
	"fmt"

	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/Blueprint-uServices/blueprint/blueprint/pkg/ir"
	"github.com/Blueprint-uServices/blueprint/plugins/golang"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/gocode"
	"github.com/Blueprint-uServices/blueprint/plugins/golang/goparser"
	"github.com/Blueprint-uServices/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type stdoutMetricCollector struct {
	golang.Node
	golang.Instantiable

	CollectorName string
	Iface         *goparser.ParsedInterface
	Constructor   *gocode.Constructor
}

func newStdOutMetricCollector(name string) (*stdoutMetricCollector, error) {
	node := &stdoutMetricCollector{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.CollectorName = name
	return node, nil
}

func (node *stdoutMetricCollector) init(name string) error {
	workflow.Init("../../runtime")
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("StdoutMetricCollector")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *stdoutMetricCollector) Name() string {
	return node.CollectorName
}

func (node *stdoutMetricCollector) String() string {
	return node.Name() + " = StdoutMetricCollector()"
}

func (node *stdoutMetricCollector) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *stdoutMetricCollector) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *stdoutMetricCollector) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *stdoutMetricCollector) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.CollectorName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating StdoutMetricCollector %v in %v/%v", node.CollectorName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.CollectorName, node.Constructor, []ir.IRNode{})
}

func (node *stdoutMetricCollector) ImplementsGolangNode() {}
