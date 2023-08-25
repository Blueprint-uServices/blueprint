package workflow

import (
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow/parser"
)

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type WorkflowService struct {
	// IR node types
	golang.Node
	golang.Service
	service.ServiceNode

	// Artifact generation
	golang.ProvidesModule
	golang.RequiresPackages
	golang.Instantiable

	InstanceName   string
	ServiceType    string
	ServiceDetails *golang.GolangServiceDetails
	Args           []blueprint.IRNode
}

func (n WorkflowService) String() string {
	var b strings.Builder
	b.WriteString(n.InstanceName)
	b.WriteString(" = ")
	b.WriteString(n.ServiceDetails.Interface.Name)

	var args []string
	for _, arg := range n.Args {
		args = append(args, arg.Name())
	}

	b.WriteString("(")
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")

	return b.String()
}

func newWorkflowService(name string, serviceType string, args []blueprint.IRNode) (*WorkflowService, error) {
	// Look up the service details; errors out if the service doesn't exist
	details, err := findService(serviceType)
	if err != nil {
		return nil, err
	}

	node := &WorkflowService{}

	node.InstanceName = name
	node.ServiceType = serviceType
	node.ServiceDetails = details
	node.Args = args
	return node, nil
}

func (node *WorkflowService) Name() string {
	return node.InstanceName
}

func (node *WorkflowService) GetInterface() *service.ServiceInterface {
	return &node.ServiceDetails.Interface
}

func addToWorkspace(builder *golang.WorkspaceBuilder, info *parser.ModuleInfo) error {
	if builder.Visited(info.Name) {
		return nil
	}
	_, subdir := filepath.Split(info.Path)
	return builder.AddLocalModule(subdir, info.Path)
}

// Adds the workspace modules containing the interface declaration and implementation
func (node *WorkflowService) AddToWorkspace(builder *golang.WorkspaceBuilder) error {
	// Copy the interface module into the workspace
	err := addToWorkspace(builder, node.ServiceDetails.InterfacePackage.Module)
	if err != nil {
		return err
	}

	// Copy the impl module into the workspace (if it's different)
	return addToWorkspace(builder, node.ServiceDetails.ImplPackage.Module)
}

func addToModule(builder *golang.ModuleBuilder, info *parser.ModuleInfo) error {
	if builder.Visited(info.Name) {
		return nil
	}
	return builder.Require(info.Name, info.Version)
}

// Adds the 'requires' statements to the module
func (node *WorkflowService) AddToModule(builder *golang.ModuleBuilder) error {
	// Make sure we've copied the module into the workspace
	node.AddToWorkspace(builder.Workspace)

	// Add the requires statements
	err := addToModule(builder, node.ServiceDetails.ImplPackage.Module)
	if err != nil {
		return err
	}
	return addToModule(builder, node.ServiceDetails.InterfacePackage.Module)
}

func (node *WorkflowService) AddInstantiation(builder *golang.DICodeBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// Make sure we've also added requires statements to the module
	err := node.AddToModule(builder.Module)
	if err != nil {
		return err
	}

	implPkg := builder.Import(node.ServiceDetails.ImplPackage.Name)

	return builder.Declare(node.InstanceName, `
		func(ctr Container) (any, error) {
			return `+implPkg+`.`+node.ServiceDetails.ImplConstructor.Name+`()
		}
	`)
}

func (node *WorkflowService) ImplementsGolangNode()         {}
func (node *WorkflowService) ImplementsGolangService()      {}
func (node *WorkflowService) ImplementsGolangInstantiable() {}
func (node *WorkflowService) ImplementsGolangLocalModule()  {}
