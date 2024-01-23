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

type stdoutLogger struct {
	golang.Node
	service.ServiceNode
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
	// TODO: update this
	return fmt.Errorf("not implemented")
	// workflow.Init("../../runtime")
	// spec, err := workflow.GetSpec()
	// if err != nil {
	// 	return err
	// }

	// details, err := spec.Get("SLogger")
	// if err != nil {
	// 	return err
	// }

	// node.Iface = details.Iface
	// node.Constructor = details.Constructor.AsConstructor()
	// return nil
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
	return node.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Instantiable
func (node *stdoutLogger) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.LoggerName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating SLogger %v in %v/%v", node.LoggerName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.LoggerName, node.Constructor, []ir.IRNode{})
}

func (node *stdoutLogger) ImplementsGolangNode() {}
