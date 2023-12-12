package linuxcontainer

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/coreplugins/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/wiring"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
)

var NamespaceType = "LinuxContainer"

var prop_CHILDREN = "Children"

/*
Adds a process to an existing container
*/
func AddProcessToContainer(spec wiring.WiringSpec, containerName, childName string) {
	spec.AddProperty(containerName, prop_CHILDREN, childName)
}

/*
Adds a container that will explicitly instantiate all of the named child processes
The container will also implicitly instantiate any of the dependencies of the children
*/
func CreateContainer(spec wiring.WiringSpec, containerName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddProcessToContainer(spec, containerName, childName)
	}

	// A linux container node is simply a namespace that accumulates linux process nodes
	spec.Define(containerName, &Container{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		node := newLinuxContainerNode(containerName)
		ctr := wiring.CreateNamespace[linux.Process](spec, namespace, containerName, NamespaceType, &node.Nodes, &node.Edges)
		_, err := pointer.InstantiateFromProperty(ctr, spec, prop_CHILDREN)
		return node, err
	})

	return containerName
}
