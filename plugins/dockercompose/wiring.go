package dockercompose

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/blueprint"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/core/pointer"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
)

// Adds a child node to an existing container deployment
func AddContainerToDeployment(wiring blueprint.WiringSpec, deploymentName, containerName string) {
	wiring.AddProperty(deploymentName, "Children", containerName)
}

// Adds a deployment that explicitly instantiates all of the containers provided.
// The deployment will also implicitly instantiate any of the dependencies of the containers
func NewDeployment(wiring blueprint.WiringSpec, deploymentName string, containers ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, containerName := range containers {
		AddContainerToDeployment(wiring, deploymentName, containerName)
	}

	wiring.Define(deploymentName, &DockerCompose{}, func(namespace blueprint.Namespace) (blueprint.IRNode, error) {
		deployment := newDockerComposeNamespace(namespace, wiring, deploymentName)

		var containerNames []string
		if err := namespace.GetProperties(deploymentName, "Children", &containerNames); err != nil {
			return nil, blueprint.Errorf("unable to build Golang process as the \"Children\" property is invalid: %s", err.Error())
		}
		deployment.Info("%v children to build (%s)", len(containerNames), strings.Join(containerNames, ", "))

		// Instantiate all of the containers.  If the container node hasn't actually been defined, then this will error out
		for _, containerName := range containerNames {
			ptr := pointer.GetPointer(wiring, containerName)
			if ptr == nil {
				// for non-pointer types, just get the child node
				var child blueprint.IRNode
				if err := deployment.Get(containerName, &child); err != nil {
					return nil, err
				}
			} else {
				// for pointer nodes, only instantiate the dst side of the pointer
				_, err := ptr.InstantiateDst(deployment)
				if err != nil {
					return nil, err
				}
			}
		}

		// Instantiate and return the service
		return deployment.handler.IRNode, nil
	})

	return deploymentName
}

// Used during building to accumulate docker container nodes
// Non-container nodes will just be recursively fetched from the parent namespace
type DockerComposeNamespace struct {
	blueprint.SimpleNamespace
	handler *dockerComposeNamespaceHandler
}

type dockerComposeNamespaceHandler struct {
	blueprint.DefaultNamespaceHandler

	IRNode *DockerCompose
}

// Creates a docker compose deployment `name` within the provided parent namespace
func newDockerComposeNamespace(parentNamespace blueprint.Namespace, wiring blueprint.WiringSpec, name string) *DockerComposeNamespace {
	namespace := &DockerComposeNamespace{}
	namespace.handler = &dockerComposeNamespaceHandler{}
	namespace.handler.Init(&namespace.SimpleNamespace)
	namespace.handler.IRNode = newDockerComposeDeployment(name)
	namespace.Init(name, "DockerCompose", parentNamespace, wiring, namespace.handler)
	return namespace
}

// Deployments can only contain container nodes
func (namespace *dockerComposeNamespaceHandler) Accepts(nodeType any) bool {
	_, ok := nodeType.(docker.ContainerNode)
	return ok
}

// When a node is added to this namespace, we just attach it to the IRNode representing the deployment
func (handler *dockerComposeNamespaceHandler) AddNode(name string, node blueprint.IRNode) error {
	return handler.IRNode.AddChild(node)
}

// When an edge is added to this namespace, we just attach it as an argument to the IRNode representing the deployment
func (handler *dockerComposeNamespaceHandler) AddEdge(name string, node blueprint.IRNode) error {
	handler.IRNode.AddArg(node)
	return nil
}
