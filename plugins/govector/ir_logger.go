package govector

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/workflow/workflowspec"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/govector"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a GoVector logger instance
type GoVecLoggerClient struct {
	golang.Node
	golang.Instantiable

	ClientName   string
	InstanceName string
	LoggerName   string

	Spec *workflowspec.Service
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

func (node *GoVecLoggerClient) init(name string) (err error) {
	node.Spec, err = workflowspec.GetService[govector.GoVecLogger]()
	return
}

// Implements golang.Instantiable
func (node *GoVecLoggerClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating GoVecLoggerClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	constructor := node.Spec.Constructor.AsConstructor()
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{&ir.IRValue{Value: node.LoggerName}})
}

// Implements golang.ProvidesModule
func (node *GoVecLoggerClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return node.Spec.AddToWorkspace(builder)
}

// Implements golang.ProvidesInterface
func (node *GoVecLoggerClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.Spec.AddToModule(builder)
}

// Implements service.ServiceNode
func (node *GoVecLoggerClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Spec.Iface.ServiceInterface(ctx), nil
}

// Implements golang.Node
func (node *GoVecLoggerClient) ImplementsGolangNode() {}
