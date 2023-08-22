package golang

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/process"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/service"
)

// Base representation for any application-level golang object
type Node interface {
	blueprint.IRNode
	ImplementsGolangNode() // Idiomatically necessary in Go for typecasting correctly
}

type Service interface {
	Node
	service.ServiceNode
	ImplementsGolangService() // Idiomatically necessary in Go for typecasting correctly
}

// Represents an application-level golang node that can generate, package, and instantiate code
type CodeGenerator interface {
	// Golang code nodes can create instances
	GenerateInstantiationCode(*GolangCodeGenerator) error
}

// Represents an application-level golang node that wants to include files, code, and dependencies with the generated artifact
type ArtifactGenerator interface {
	// Golang artifact nodes can generate output artifacts like files and code
	CollectArtifacts(*GolangArtifactGenerator) error
}

type Package struct {
	Name string
	Path string
}

type ServiceInterface struct {
	service.ServiceInterface
	Package Package
}

// Code location and interfaces of a service
type GolangServiceDetails struct {
	Name        string                           // The name of the implementing struct
	Package     Package                          // The package containing the implementing struct
	Constructor service.ServiceMethodDeclaration // The constructor method for the implementing struct
	Interface   ServiceInterface                 // The interface that is implemented
}

func (d GolangServiceDetails) String() string {
	var b strings.Builder
	b.WriteString(d.Interface.Name)

	var constructorArgs []string
	for _, arg := range d.Constructor.Args {
		constructorArgs = append(constructorArgs, arg.Type)
	}
	b.WriteString("(")
	b.WriteString(strings.Join(constructorArgs, ", "))
	b.WriteString(")")

	return b.String()
}

// This Node represents a Golang process that internally will instantiate a number of application-level services
type Process struct {
	blueprint.IRNode
	process.ProcessNode
	ArtifactGenerator

	InstanceName           string
	ArgNodes               []blueprint.IRNode
	ContainedNodes         []blueprint.IRNode
	ContainedArtifactNodes []ArtifactGenerator
	ContainedInstanceNodes []CodeGenerator
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
	if artifactNode, ok := child.(ArtifactGenerator); ok {
		node.ContainedArtifactNodes = append(node.ContainedArtifactNodes, artifactNode)
	}
	if instanceNode, ok := child.(CodeGenerator); ok {
		node.ContainedInstanceNodes = append(node.ContainedInstanceNodes, instanceNode)
	}
	return nil
}

func (node *Process) CollectArtifacts(ag *GolangArtifactGenerator) error {
	// Collect all the artifacts of the contained nodes
	for _, n := range node.ContainedArtifactNodes {
		n.CollectArtifacts(ag)
	}

	// Now generate our own artifacts, using code generator
	ca := NewGolangCodeAccumulator()
	for _, n := range node.ContainedInstanceNodes {
		n.GenerateInstantiationCode(ca)
	}

	code := `

	`

	// TODO: correct output path
	ag.AddCode(node.InstanceName, code)
	return nil
}
