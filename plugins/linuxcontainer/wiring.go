package linuxcontainer

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

// Adds a process to an existing container
func AddProcessToContainer(spec wiring.WiringSpec, containerName, childName string) {
	namespaceutil.AddNodeTo[Container](spec, containerName, childName)
}

// Wraps serviceName with a modifier that deploys the service inside a container
func Deploy(spec wiring.WiringSpec, serviceName string) string {
	ctrName := serviceName + "_ctr"
	CreateContainer(spec, ctrName, serviceName)
	return serviceName
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
		ctr := newLinuxContainerNode(containerName)
		_, err := namespaceutil.InstantiateNamespace(namespace, &LinuxContainerNamespace{ctr})
		return ctr, err
	})

	return containerName
}

// A [wiring.NamespaceHandler] used to build golang process nodes
type LinuxContainerNamespace struct {
	*Container
}

// Implements [wiring.NamespaceHandler]
func (ctr *Container) Accepts(nodeType any) bool {
	_, isLinuxProcess := nodeType.(linux.Process)
	return isLinuxProcess
}

// Implements [wiring.NamespaceHandler]
func (ctr *Container) AddEdge(name string, edge ir.IRNode) error {
	ctr.Edges = append(ctr.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (ctr *Container) AddNode(name string, node ir.IRNode) error {
	ctr.Nodes = append(ctr.Nodes, node)
	return nil
}
