package goproc

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/golang/goparser"
	"golang.org/x/exp/slog"
)

type stdoutMetricCollector struct {
	golang.Node
	service.ServiceNode
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
	// TODO: update this
	return fmt.Errorf("not implemented")
	// workflow.Init("../../runtime")
	// spec, err := workflow.GetSpec()
	// if err != nil {
	// 	return err
	// }

	// details, err := spec.Get("StdoutMetricCollector")
	// if err != nil {
	// 	return err
	// }

	// node.Iface = details.Iface
	// node.Constructor = details.Constructor.AsConstructor()
	// return nil
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
	return node.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *stdoutMetricCollector) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.CollectorName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating StdoutMetricCollector %v in %v/%v", node.CollectorName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.CollectorName, node.Constructor, []ir.IRNode{})
}

func (node *stdoutMetricCollector) ImplementsGolangNode() {}
