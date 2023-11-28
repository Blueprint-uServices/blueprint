package simplequeue

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/backend"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"golang.org/x/exp/slog"
)

type SimpleQueue struct {
	golang.Service
	backend.Queue

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.Instantiable

	InstanceName string

	Iface       *goparser.ParsedInterface // The Queue interface
	Constructor *gocode.Constructor       // Constructor for this Queue implementation
}

func newSimpleQueue(name string) (*SimpleQueue, error) {
	node := &SimpleQueue{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (node *SimpleQueue) init(name string) error {
	// We use the workflow spec to load the queue interface details
	workflow.Init("../../runtime")

	// Look up the service details; errors out if the service doesn't exist
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}
	details, err := spec.Get("SimpleQueue")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *SimpleQueue) Name() string {
	return node.InstanceName
}

func (node *SimpleQueue) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

/* The queue interface and SimpleQueue implementation exist in the runtime package */
func (node *SimpleQueue) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Add blueprint runtime to the workspace
	return golang.AddRuntimeModule(builder)
}

func (node *SimpleQueue) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *SimpleQueue) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating SimpleQueue %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, node.Constructor, nil)
}

func (node *SimpleQueue) String() string {
	return fmt.Sprintf("%v = SimpleQueue()", node.InstanceName)
}

func (node *SimpleQueue) ImplementsGolangNode()    {}
func (node *SimpleQueue) ImplementsGolangService() {}
