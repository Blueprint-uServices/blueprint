package linuxcontainer

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
)

/*
Adds a process to an existing container
*/
func AddProcessToContainer(wiring blueprint.WiringSpec, containerName, childName string) {
	wiring.AddProperty(containerName, "Children", childName)
}

/*
Adds a container that will explicitly instantiate all of the named child processes
The container will also implicitly instantiate any of the dependencies of the children
*/
func CreateContainer(wiring blueprint.WiringSpec, containerName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddProcessToContainer(wiring, containerName, childName)
	}

	wiring.Define(containerName, &Container{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		container := newLinuxNamespace(namespace, wiring, containerName)

		var childNames []string
		if err := namespace.GetProperties(containerName, "Children", &childNames); err != nil {
			return nil, blueprint.Errorf("unable to build Linux container as the \"Children\" property is invalid: %s", err.Error())
		}
		container.Info("%v children to build (%s)", len(childNames), strings.Join(childNames, ", "))

		// Instantiate all of the child nodes.  If the child node hasn't actually been defined, then this will error out
		for _, childName := range childNames {
			ptr := pointer.GetPointer(wiring, childName)
			if ptr == nil {
				// for non-pointer types, just get the child node
				var child blueprint.IRNode
				if err := container.Get(childName, &child); err != nil {
					return nil, err
				}
			} else {
				// for pointer nodes, only instantiate the dst side of the pointer
				_, err := ptr.InstantiateDst(container)
				if err != nil {
					return nil, err
				}
			}
		}

		// Instantiate and return the service
		return container.handler.IRNode, nil

	})
	return containerName
}

// Used during building to accumulate linux process nodes
// Non-linux process nodes will just be recursively fetched from the parent namespace
type LinuxNamespace struct {
	blueprint.SimpleNamespace
	handler *linuxNamespaceHandler
}

type linuxNamespaceHandler struct {
	blueprint.DefaultNamespaceHandler

	IRNode *Container
}

// Creates a process `name` within the provided parent namespace
func newLinuxNamespace(parentNamespace blueprint.Namespace, wiring blueprint.WiringSpec, name string) *LinuxNamespace {
	namespace := &LinuxNamespace{}
	namespace.handler = &linuxNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newLinuxContainerNode(name)
	namespace.Init(name, "Linux", parentNamespace, wiring, namespace.handler)
	return namespace
}

// Golang processes can only contain golang nodes
func (namespace *linuxNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(linux.Node)
	return ok
}

// When a node is added to this namespace, we just attach it to the IRNode representing the linux container
func (handler *linuxNamespaceHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this namespace, we just attach it as an argument to the IRNode representing the linux container
func (handler *linuxNamespaceHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
