package linuxcontainer

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
)

/*
linuxcontainer.Container is a node that represents a collection of runnable linux processes.
It can contain any number of other process.Node IRNodes.  When it's compiled, the goproc.Process
will generate a run script that instantiates all contained processes.
*/

type Container struct {
	ir.IRNode

	/* The implemented build targets for linuxcontainer.Container nodes */
	filesystemDeployer /* Can be deployed as a basic collection of processes; implemented in deploy.go */
	dockerDeployer     /* Can be deployed as a docker container; implemented in deploydocker.go */

	InstanceName   string
	ImageName      string
	ArgNodes       []ir.IRNode
	ContainedNodes []ir.IRNode
}

func newLinuxContainerNode(name string, argNodes, containedNodes []ir.IRNode) *Container {
	node := Container{
		InstanceName:   name,
		ImageName:      ir.CleanName(name),
		ArgNodes:       argNodes,
		ContainedNodes: containedNodes,
	}
	return &node
}

func (node *Container) Name() string {
	return node.InstanceName
}

func (node *Container) String() string {
	return ir.PrettyPrintNamespace(node.InstanceName, NamespaceType, node.ArgNodes, node.ContainedNodes)
}
