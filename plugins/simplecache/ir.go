package simplecache

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

type SimpleCache struct {
	golang.Service
	backend.Cache

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.Instantiable

	InstanceName string

	Iface       *goparser.ParsedInterface // The Cache interface
	Constructor *gocode.Constructor       // Constructor for this Cache implementation
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
	node.Iface = details.Iface
	node.Constructor = details.Constructor.AsConstructor()
	return nil
}

func (node *SimpleCache) Name() string {
	return node.InstanceName
}

func (node *SimpleCache) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.Iface.ServiceInterface(ctx), nil
}

/* The cache interface and simplecache implementation exist in the runtime package */
func (node *SimpleCache) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	return golang.AddRuntimeModule(builder)
}

func (node *SimpleCache) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

func (node *SimpleCache) AddInstantiation(builder golang.NamespaceBuilder) error {
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
