package linuxcontainer

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/process"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/process/procgen"
	"golang.org/x/exp/slog"
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
	blueprint.RegisterDefaultNamespace[process.Node]("linuxcontainer", buildDefaultProcessWorkspace)
}

type Container struct {
	blueprint.IRNode
	docker.Node
	docker.ProvidesContainerImage
	docker.ProvidesContainerInstance

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

func (node *Container) generateArtifacts(procWorkspaceDir string, generateDockerfile bool) error {
	// Create subdirectory for the processes in this ctr image
	slog.Info(fmt.Sprintf("Building linux ctr %s to %s", node.Name(), procWorkspaceDir))
	workspace, err := procgen.NewProcWorkspaceBuilder(procWorkspaceDir)
	if err != nil {
		return err
	}

	// Add all processes artifacts to the workspace
	for _, child := range node.ContainedNodes {
		if n, valid := child.(process.ProvidesProcessArtifacts); valid {
			if err := n.AddProcessArtifacts(workspace); err != nil {
				return err
			}
		}
	}

	// Collect the scripts to run the processes
	graph, err := procgen.NewProcGraphBuilderImpl(workspace, node.Name(), "run.sh")
	if err != nil {
		return err
	}

	for _, child := range node.ContainedNodes {
		if n, valid := child.(process.InstantiableProcess); valid {
			if err := n.AddProcessInstance(graph); err != nil {
				return err
			}
		}
	}

	// TODO: it's possible some metadata / address nodes are residing in this namespace.  They don't
	// get passed in as args, but need to be added to the graph nonetheless

	// Generate the run.sh file
	graph.Build()

	if generateDockerfile {
		// TODO: generate dockerfile not implemented yet
		// TODO: dockerfile will need to also invoke build script from workspace
		return fmt.Errorf("generating dockerfile not yet implemented")
	}

	return workspace.Finish()
}

func buildDefaultProcessWorkspace(outputDir string, nodes []blueprint.IRNode) error {
	ctr := newLinuxContainerNode("default")
	ctr.ContainedNodes = nodes
	return ctr.generateArtifacts(outputDir, false)
}
