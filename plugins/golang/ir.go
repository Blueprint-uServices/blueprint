package golang

import (
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/workflow/parser"
)

var generatedModulePrefix = "gitlab.mpi-sws.org/cld/blueprint/plugins/golang/process"

// Base representation for any application-level golang object
type Node interface {
	blueprint.IRNode
	ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
}

// A golang node that is also a service
type Service interface {
	Node
	service.ServiceNode
	ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
}

// A golang node that can generate and/or package code artifacts
type ProvidesModule interface {
	AddToWorkspace(*WorkspaceBuilder) error
}

type RequiresPackages interface {
	AddToModule(*ModuleBuilder) error
}

type Instantiable interface {
	AddInstantiation(*DICodeBuilder) error
}

// A golang process that instantiates Golang nodes.  This is Blueprint's main implementation of Golang processes
type Process struct {
	blueprint.IRNode
	process.ProcessNode
	process.ArtifactGenerator

	InstanceName   string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

// Code location and interfaces of a service
type GolangServiceDetails struct {
	Interface        service.ServiceInterface         // The interface that is implemented
	InterfacePackage *parser.PackageInfo              // The package containing the constructor method
	ImplName         string                           // The type name of the implementing struct
	ImplConstructor  service.ServiceMethodDeclaration // The constructor method for the implementing struct
	ImplPackage      *parser.PackageInfo              // The package containing the constructor method
}

func (d GolangServiceDetails) String() string {
	var b strings.Builder
	b.WriteString("import \"" + d.InterfacePackage.ImportName + "\"\n")
	b.WriteString("var service " + d.InterfacePackage.ShortName + "." + d.Interface.Name + "\n")
	b.WriteString("service = " + d.ImplConstructor.Name)
	var constructorArgs []string
	for _, arg := range d.ImplConstructor.Args {
		constructorArgs = append(constructorArgs, arg.Name)
	}
	b.WriteString("(")
	b.WriteString(strings.Join(constructorArgs, ", "))
	b.WriteString(")")

	return b.String()
}

// A Golang Process Node can either be given the child nodes ahead of time, or they can be added using AddArtifactNode / AddCodeNode
func newGolangProcessNode(name string) *Process {
	node := Process{}
	node.InstanceName = name
	return &node
}

func (node *Process) Name() string {
	return node.InstanceName
}

func (node *Process) String() string {
	var b strings.Builder
	b.WriteString(node.InstanceName)
	b.WriteString(" = GolangProcessNode(")
	var args []string
	for _, arg := range node.ArgNodes {
		args = append(args, arg.Name())
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(") {\n")
	var children []string
	for _, child := range node.ContainedNodes {
		children = append(children, child.String())
	}
	b.WriteString(blueprint.Indent(strings.Join(children, "\n"), 2))
	b.WriteString("\n}")
	return b.String()
}

func (node *Process) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *Process) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}

func (node *Process) Build(outputDir string) error {
	if isDir(outputDir) {
		return fmt.Errorf("cannot built to %s, directory already exists", outputDir)
	}
	err := checkDir(outputDir, true)
	if err != nil {
		return fmt.Errorf("unable to create %s for process %s due to %s", outputDir, node.Name(), err.Error())
	}

	// TODO: might end up building multiple times which is OK, so need a check here that we haven't already built this artifact, even if it was by a different (but identical) node
	workspaceDir := filepath.Join(outputDir, node.Name())
	workspace, err := NewWorkspaceBuilder(workspaceDir)
	if err != nil {
		return err
	}

	moduleName := generatedModulePrefix + "/" + node.Name()
	module, err := NewModuleBuilder(workspace, node.Name(), moduleName)
	if err != nil {
		return err
	}

	code, err := NewDICodeBuilder(module, "main.go", "pkg/main", "init")
	if err != nil {
		return err
	}

	for _, node := range node.ContainedNodes {
		if instantiable, ok := node.(Instantiable); ok {
			err := instantiable.AddInstantiation(code)
			if err != nil {
				return err
			}
		}
		if packages, ok := node.(RequiresPackages); ok {
			err := packages.AddToModule(module)
			if err != nil {
				return err
			}
		}
		if modules, ok := node.(ProvidesModule); ok {
			err := modules.AddToWorkspace(workspace)
			if err != nil {
				return err
			}
		}
	}

	// TODO:
	//  generate the DI code and main function

	err = module.Finish()
	if err != nil {
		return err
	}

	err = workspace.Finish()
	if err != nil {
		return err
	}

	return nil
}
