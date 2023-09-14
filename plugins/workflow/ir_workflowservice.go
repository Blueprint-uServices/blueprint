package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
)

/*
A base IRNode that is responsible for including workflow spec modules in
generated code.

This is used by both the caller side of a service and the service itself,
since both sides need to know the workflow spec's interfaces.

It hooks into artifact generation, but only to add copy the workflow
spec module into the output directory.  It does not have any
runtime instances or nodes.
*/
type WorkflowSpecNode struct {
	// IR node types
	golang.Node

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule

	InstanceName string // Name of this instance
	ServiceType  string // The short-name serviceType used to initialize this workflow service

	// Details of the service, including its interface and constructor
	ServiceInfo *WorkflowSpecService

	// The workflow spec where this service originated
	Spec *WorkflowSpec
}

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type WorkflowService struct {
	WorkflowSpecNode

	// Additional IR node types
	golang.Service

	// Additional interfaces for generating Golang artifacts
	golang.RequiresPackages
	golang.Instantiable

	// IR Nodes of arguments that will be passed in to the generated code
	Args []blueprint.IRNode
}

/*
A node representing the server-side of a workflow service.
*/
func newWorkflowService(name string, serviceType string, args []blueprint.IRNode) (*WorkflowService, error) {
	node := &WorkflowService{}
	err := node.init(name, serviceType)
	if err != nil {
		return nil, err
	}

	// TODO: can eagerly typecheck args here
	if len(node.ServiceInfo.Constructor.Arguments) != len(args) {
		var argStrings []string
		for _, arg := range args {
			argStrings = append(argStrings, arg.Name())
		}
		return nil, fmt.Errorf("mismatched # arguments for %s, constructor is %v but args are (%v)", name, node.ServiceInfo.Constructor, strings.Join(argStrings, ", "))
	}
	node.Args = args

	return node, nil
}

/*
A node representnig the client-side of a workflow service.

All this node does is include the interface definitions of the service
*/
func includeWorkflowDependencies(name string, serviceType string) (*WorkflowSpecNode, error) {
	node := &WorkflowSpecNode{}
	err := node.init(name+".client_dependencies", serviceType)
	return node, err
}

func (node *WorkflowSpecNode) init(name string, serviceType string) error {
	// Look up the service details; errors out if the service doesn't exist
	spec, err := getSpec()
	if err != nil {
		return err
	}
	details, err := spec.Get(serviceType)
	if err != nil {
		return err
	}

	node.InstanceName = name
	node.ServiceType = serviceType
	node.ServiceInfo = details
	node.Spec = spec
	return nil
}

func (node *WorkflowSpecNode) Name() string {
	return node.InstanceName
}

func (node *WorkflowService) Name() string {
	return node.InstanceName
}

func (node *WorkflowService) GetInterface() service.ServiceInterface {
	return node.ServiceInfo.GetInterface()
}

func addToWorkspace(builder golang.WorkspaceBuilder, mod *goparser.ParsedModule) error {
	if builder.Visited(mod.Name) {
		return nil
	}
	_, subdir := filepath.Split(mod.SrcDir)
	return builder.AddLocalModule(subdir, mod.SrcDir)
}

/*
Part of artifact generation.  Used by both the client and server side.
Adds the modules containing the workflow's interfaces to the workspace
*/
func (node *WorkflowSpecNode) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Add the interfaces to the workspace
	return addToWorkspace(builder, node.ServiceInfo.Iface.File.Package.Module)
}

/*
Part of artifact generation.  In addition to the interfaces, adds the constructor
to the workspace.  Most likely the constructor resides in the same module as
the interfaces, but in case it doesn't, it will add the correct module
*/
func (node *WorkflowService) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Add the interfaces to the workspace
	err := node.WorkflowSpecNode.AddToWorkspace(builder)
	if err != nil {
		return err
	}

	// Copy the impl module into the workspace (if it's different)
	return addToWorkspace(builder, node.ServiceInfo.Constructor.File.Package.Module)
}

func addToModule(builder golang.ModuleBuilder, mod *goparser.ParsedModule) error {
	if builder.Visited(mod.Name) {
		return nil
	}
	return builder.Require(mod.Name, mod.Version)
}

// Part of module generation; Adds the 'requires' statements to the module
func (node *WorkflowService) AddRequires(builder golang.ModuleBuilder) error {
	// Add the requires statements
	err := addToModule(builder, node.ServiceInfo.Iface.File.Package.Module)
	if err != nil {
		return err
	}
	return addToModule(builder, node.ServiceInfo.Constructor.File.Package.Module)
}

func (node *WorkflowService) AddInstantiation(builder golang.GraphBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	return builder.DeclareConstructor(node.InstanceName, node.ServiceInfo.GetConstructor(), node.Args)
}

func (n *WorkflowService) String() string {
	var b strings.Builder
	b.WriteString(n.InstanceName)
	b.WriteString(" = ")
	b.WriteString(n.ServiceInfo.Iface.Name)

	var args []string
	for _, arg := range n.Args {
		args = append(args, arg.Name())
	}

	b.WriteString("(")
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")

	return b.String()
}

func (n *WorkflowSpecNode) String() string {
	return "import " + n.ServiceInfo.Iface.Name
}

func (node *WorkflowSpecNode) ImplementsGolangNode()   {}
func (node *WorkflowService) ImplementsGolangNode()    {}
func (node *WorkflowService) ImplementsGolangService() {}
