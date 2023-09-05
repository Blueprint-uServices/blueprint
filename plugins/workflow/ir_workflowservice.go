package workflow

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/golang/goparser"
)

// This Node represents a Golang Workflow spec service in the Blueprint IR.
type WorkflowService struct {
	// IR node types
	golang.Node
	golang.Service
	service.ServiceNode

	// Interfaces for generating Golang artifacts
	golang.ProvidesModule
	golang.RequiresPackages
	golang.Instantiable

	InstanceName string // Name of this instance
	ServiceType  string // The short-name serviceType used to initialize this workflow service

	// Details of the service, including its interface and constructor
	ServiceInfo *WorkflowSpecService

	// IR Nodes of arguments that will be passed in to the generated code
	Args []blueprint.IRNode

	// The workflow spec where this service originated
	Spec *WorkflowSpec
}

func (n WorkflowService) String() string {
	var b strings.Builder
	b.WriteString(n.InstanceName)
	b.WriteString(" = ")
	// b.WriteString(n.ServiceDetails.Interface.Name())

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
	spec, err := getSpec()
	if err != nil {
		return nil, err
	}
	details, err := spec.Get(serviceType)
	if err != nil {
		return nil, err
	}

	node := &WorkflowService{}

	node.InstanceName = name
	node.ServiceType = serviceType
	node.ServiceInfo = details
	node.Args = args
	node.Spec = spec
	// TODO: can eagerly typecheck args here
	return node, nil
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

// Adds the workspace modules containing the interface declaration and implementation
func (node *WorkflowService) AddToWorkspace(builder golang.WorkspaceBuilder) error {
	// Copy the interface module into the workspace
	err := addToWorkspace(builder, node.ServiceInfo.iface.File.Package.Module)
	if err != nil {
		return err
	}

	// Copy the impl module into the workspace (if it's different)
	return addToWorkspace(builder, node.ServiceInfo.constructor.File.Package.Module)
}

func addToModule(builder golang.ModuleBuilder, mod *goparser.ParsedModule) error {
	if builder.Visited(mod.Name) {
		return nil
	}
	return builder.Require(mod.Name, mod.Version)
}

// Adds the 'requires' statements to the module
func (node *WorkflowService) AddToModule(builder golang.ModuleBuilder) error {
	// Make sure we've copied the module into the workspace
	node.AddToWorkspace(builder.Workspace())

	// Add the requires statements
	err := addToModule(builder, node.ServiceInfo.iface.File.Package.Module)
	if err != nil {
		return err
	}
	return addToModule(builder, node.ServiceInfo.constructor.File.Package.Module)
}

type buildFuncArgs struct {
	ImplPkg string
	Node    *WorkflowService
}

var buildFuncTemplate = `func(ctr golang.Container) (any, error) {
	
		{{ range $i, $argName := .Node.Args }}
		arg{{ $i }}, err := ctr.Get("{{ $argName.Name }}")
		if err != nil {
			return nil, err
		}
		{{ end }}

		return {{ .ImplPkg }}.{{ .Node.ServiceDetails.ImplConstructor.Name }}({{ range $i, $argName := .Node.Args }}{{ if $i }},{{end}}arg{{ $i }}{{ end }}), nil
	}`

func (node *WorkflowService) AddInstantiation(builder golang.DICodeBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	// Make sure we've also added requires statements to the module
	err := node.AddToModule(builder.Module())
	if err != nil {
		return err
	}

	// Instantiate the code template
	t, err := template.New(node.InstanceName).Parse(buildFuncTemplate)
	if err != nil {
		return err
	}

	// node.Args[0].Name()

	// node.ServiceDetails.ImplConstructor.Args[0].Type

	// Arguments to the code template
	templateData := buildFuncArgs{
		builder.Import(node.ServiceInfo.constructor.File.Package.Name),
		node,
	}

	// Generate the code
	buf := &bytes.Buffer{}
	err = t.Execute(buf, templateData)
	if err != nil {
		return err
	}

	return builder.Declare(node.InstanceName, buf.String())
}

func (node *WorkflowService) ImplementsGolangNode()         {}
func (node *WorkflowService) ImplementsGolangService()      {}
func (node *WorkflowService) ImplementsGolangInstantiable() {}
func (node *WorkflowService) ImplementsGolangLocalModule()  {}
