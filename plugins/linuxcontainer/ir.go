package linuxcontainer

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
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

	InstanceName string
	ImageName    string
	Edges        []ir.IRNode
	Nodes        []ir.IRNode
}

func newLinuxContainerNode(name string) *Container {
	node := Container{
		InstanceName: name,
		ImageName:    ir.CleanName(name),
	}
	return &node
}

// Implements ir.IRNode
func (ctr *Container) Name() string {
	return ctr.InstanceName
}

// Implements ir.IRNode
func (ctr *Container) String() string {
	return ir.PrettyPrintNamespace(ctr.InstanceName, NamespaceType, ctr.Edges, ctr.Nodes)
}

// Implements NamespaceHandler
func (ctr *Container) Accepts(nodeType any) bool {
	_, isLinuxProcess := nodeType.(linux.Process)
	return isLinuxProcess
}

// Implements NamespaceHandler
func (ctr *Container) AddEdge(name string, edge ir.IRNode) error {
	ctr.Edges = append(ctr.Edges, edge)
	return nil
}

// Implements NamespaceHandler
func (ctr *Container) AddNode(name string, node ir.IRNode) error {
	ctr.Nodes = append(ctr.Nodes, node)
	return nil
}
