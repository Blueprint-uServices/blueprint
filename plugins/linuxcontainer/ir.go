package linuxcontainer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
)

/*
linuxcontainer.Container is a node that represents a collection of runnable linux processes.
It can contain any number of other process.Node IRNodes.  When it's compiled, the goproc.Process
will generate a run script that instantiates all contained processes.
*/

func init() {
	RegisterBuilders()
}

// to trigger module initialization and register builders
func RegisterBuilders() {
	blueprint.RegisterDefaultNamespace[linux.Node]("linuxcontainer", buildDefaultProcessWorkspace)
}

type Container struct {
	blueprint.IRNode

	/* The implemented build targets for linuxcontainer.Container nodes */
	BasicLinuxContainer  /* Can be deployed as a basic collection of processes; implemented in deploy.go */
	DockerLinuxContainer /* Can be deployed as a docker container; implemented in deploydocker.go */

	InstanceName   string
	ImageName      string
	ArgNodes       []blueprint.IRNode
	ContainedNodes []blueprint.IRNode
}

func newLinuxContainerNode(name string) *Container {
	node := Container{}
	node.InstanceName = name
	node.ImageName = blueprint.CleanName(name)
	return &node
}

func (node *Container) Name() string {
	return node.InstanceName
}

func (node *Container) String() string {
	var b strings.Builder
	b.WriteString(node.InstanceName)
	b.WriteString(" = LinuxContainer(")
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

func (node *Container) AddArg(argnode blueprint.IRNode) {
	node.ArgNodes = append(node.ArgNodes, argnode)
}

func (node *Container) AddChild(child blueprint.IRNode) error {
	node.ContainedNodes = append(node.ContainedNodes, child)
	return nil
}

func (node *Container) AddContainerImage(set docker.ImageSet) error {
	return nil
}

func (node *Container) AddContainerInstance(app docker.DockerApp) error {
	return nil
}

func buildDefaultProcessWorkspace(outputDir string, nodes []blueprint.IRNode) error {
	ctr := newLinuxContainerNode("default")
	ctr.ContainedNodes = nodes
	return ctr.GenerateArtifacts(outputDir)
}
