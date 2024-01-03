package govector

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

// Blueprint IR Node that represents a client to the govec container
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

func (node *GoVecLoggerClient) Name() string {
	return node.ClientName
}

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

func (node *GoVecLoggerClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	if builder.Visited(node.ClientName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating GoVecLoggerClient %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))

	return builder.DeclareConstructor(node.InstanceName, node.Constructor, []ir.IRNode{&ir.IRValue{Value: node.LoggerName}})
}

func (node *GoVecLoggerClient) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *GoVecLoggerClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *GoVecLoggerClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

func (node *GoVecLoggerClient) ImplementsGolangNode() {}
