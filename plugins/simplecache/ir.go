package simplecache

import (
	"fmt"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/gocode"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow"
	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
	"golang.org/x/exp/slog"
)

type SimpleCache struct {
	golang.Service
	backend.Cache

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.Instantiable

	InstanceName string

	Iface       *gocode.ServiceInterface // The Cache interface
	Constructor *gocode.Constructor      // Constructor for this Cache implementation
}

func newSimpleCache(name string) (*SimpleCache, error) {
	node := &SimpleCache{}
	err := node.init(name)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (node *SimpleCache) init(name string) error {
	// We use the workflow spec to load the cache interface details
	workflow.Init("../../runtime")

	// Look up the service details; errors out if the service doesn't exist
	spec, err := workflow.GetSpec()
	if err != nil {
		return err
	}
	details, err := spec.Get("SimpleCache")
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.Iface = details.Iface.ServiceInterface()
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *SimpleCache) Name() string {
	return node.InstanceName
}

func (node *SimpleCache) GetInterface() service.ServiceInterface {
	return node.GetGoInterface()
}

func (node *SimpleCache) GetGoInterface() *gocode.ServiceInterface {
	return node.Iface
}

/* The cache interface and simplecache implementation exist in the runtime package */
func (node *SimpleCache) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Add blueprint runtime to the workspace
	if !builder.Visited("runtime") {
		slog.Info("Copying local module runtime to workspace")
		return builder.AddLocalModuleRelative("runtime", "../../../runtime")
	}
	return nil
}

func (node *SimpleCache) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating SimpleCache %v in %v/%v", node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, node.Constructor, nil)
}

func (node *SimpleCache) String() string {
	return fmt.Sprintf("%v = SimpleCache()", node.InstanceName)
}

func (node *SimpleCache) ImplementsGolangNode()    {}
func (node *SimpleCache) ImplementsGolangService() {}
