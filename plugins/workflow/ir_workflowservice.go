package workflow

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/service"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
	"golang.org/x/exp/slog"
)

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type WorkflowService struct {
	// IR node types
	golang.Service

	InstanceName string // Name of this instance
	ServiceType  string // The short-name serviceType used to initialize this workflow service

	// Details of the service, including its interface and constructor
	ServiceInfo *WorkflowSpecService

	// The workflow spec where this service originated
	Spec *WorkflowSpec

	// IR Nodes of arguments that will be passed in to the generated code
	Args []ir.IRNode
}

/*
A node representing the server-side of a workflow service.
*/
func newWorkflowService(name string, serviceType string, args []ir.IRNode) (*WorkflowService, error) {
	// Look up the service details; errors out if the service doesn't exist
	spec, err := GetSpec()
	if err != nil {
		return nil, err
	}

	details, err := spec.Get(serviceType)
	if err != nil {
		return nil, err
	}

	node := &WorkflowService{
		InstanceName: name,
		ServiceType:  serviceType,
		ServiceInfo:  details,
		Spec:         spec,
	}

	// TODO: could optionally eagerly typecheck args here
	if len(node.ServiceInfo.Constructor.Arguments) != len(args)+1 {
		var argStrings []string
		for _, arg := range args {
			argStrings = append(argStrings, arg.Name())
		}
		return nil, blueprint.Errorf("mismatched # arguments for %s, constructor is %v but args are (ctx, %v)", name, node.ServiceInfo.Constructor, strings.Join(argStrings, ", "))
	}
	node.Args = args

	return node, nil
}

func (node *WorkflowService) Name() string {
	return node.InstanceName
}

func (node *WorkflowService) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	return node.ServiceInfo.Iface.ServiceInterface(ctx), nil
}

func (node *WorkflowService) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.AddToWorkspace(builder.Workspace())
}

/*
Part of artifact generation.  In addition to the interfaces, adds the constructor
to the workspace.  Most likely the constructor resides in the same module as
the interfaces, but in case it doesn't, it will add the correct module
*/
func (node *WorkflowService) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Add blueprint runtime to the workspace
	if err := golang.AddRuntimeModule(builder); err != nil {
		return err
	}

	// Add interface module to workspace
	if _, err := CopyModuleToOutputWorkspace(builder, node.ServiceInfo.Iface.File.Package.Module); err != nil {
		return err
	}

	// Add constructor module to workspace
	if _, err := CopyModuleToOutputWorkspace(builder, node.ServiceInfo.Constructor.File.Package.Module); err != nil {
		return err
	}

	return nil
}

func CopyModuleToOutputWorkspace(b golang.WorkspaceBuilder, mod *goparser.ParsedModule) (string, error) {
	if b.Visited(mod.Name) {
		return "", nil
	}
	_, subdir := filepath.Split(mod.SrcDir)
	slog.Info(fmt.Sprintf("Copying local module %v to workspace", subdir))
	return b.AddLocalModule(subdir, mod.SrcDir)
}

func (node *WorkflowService) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	slog.Info(fmt.Sprintf("Instantiating %v %v in %v/%v", node.ServiceType, node.InstanceName, builder.Info().Package.PackageName, builder.Info().FileName))
	return builder.DeclareConstructor(node.InstanceName, node.ServiceInfo.Constructor.AsConstructor(), node.Args)
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

func (node *WorkflowService) ImplementsGolangNode()    {}
func (node *WorkflowService) ImplementsGolangService() {}
