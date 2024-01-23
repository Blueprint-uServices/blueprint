package govector

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

// Blueprint IR Node that represents a GoVector logger instance
type GoVecLoggerClient struct {
	golang.Node
	golang.Instantiable

	ClientName   string
	InstanceName string
	LoggerName   string
	Iface        *goparser.ParsedInterface
	Constructor  *gocode.Constructor
}

func newGoVecLoggerClient(name string) (*GoVecLoggerClient, error) {
	node := &GoVecLoggerClient{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}
	node.ClientName = name
	node.InstanceName = name
	node.LoggerName = name
	return node, nil
}

// Implements ir.IRNode
func (node *GoVecLoggerClient) Name() string {
	return node.ClientName
}

// Implements ir.IRNode
func (node *GoVecLoggerClient) String() string {
	return node.Name() + " = GoVecLogger()"
}

func (node *GoVecLoggerClient) init(name string) error {
	workflow.Init("../../runtime")

	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}

	details, err := spec.Get("GoVecLogger")
	if err != nil {
		return err
	}

	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

// Implements golang.Instantiable
func (node *GoVecLoggerClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating GoVecLoggerClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{&ir.IRValue{Value: node.LoggerName}})
}

// Implements golang.ProvidesModule
func (node *GoVecLoggerClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// TODO: move govec logger implementation into this package
	return fmt.Errorf("not implemented")
	// return golang.AddRuntimeModule(builder)
}

// Implements golang.ProvidesInterface
func (node *GoVecLoggerClient) AddInterfaces(builder golang.ModuleBuilder) error {
	// TODO: move govec logger implementation into this package
	return fmt.Errorf("not implemented")
	// return node.AddToWorkspace(builder.Workspace())
}

// Implements service.ServiceNode
func (node *GoVecLoggerClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Node
func (node *GoVecLoggerClient) ImplementsGolangNode() {}
