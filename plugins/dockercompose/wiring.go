// Package dockercompose is a plugin for instantiating multiple container instances in a single docker-compose
// deployment.
//
// # Wiring Spec Usage
//
// To use the dockercompose plugin in your wiring spec, you can declare a deployment, giving it a name and
// specifying which container instances to include
//
//	dockercompose.NewDeployment(spec, "my_deployment", "my_container_1", "my_container_2")
//
// You can add containers to existing deployments:
//
//	dockercompose.AddContainerToDeployment(spec, "my_deployment", "my_container_3")
//
// # Artifacts Generated
//
// During compilation, the plugin generates a docker-compose file that instantiates images for the specified
// containers.  The plugin also sets environment variables and ports for the instances.
//
// dockercompose is the default builder for container images: if your wiring spec defines containers but doesn't
// further put them into a namespace, then by default Blueprint will generate a dockercompose file.
//
// # Running Artifacts
//
// If the dockercompose deployment is not further combined by other plugins, then the entry point to running
// your application will be using docker-compose.
//
// After building an application, you can run docker-compose build and docker-compose up to build and start
// the application respectively.  You will probably need to set a number of environment variables for this
// to work, to decide which ports you want to expose locally.  Docker will complain about these environment
// variables not being set, if they are absent.  You can write these to a .env file or set them
//
// # Internals
//
// Internally, the plugin makes use of interfaces defined in the [docker] plugin.  It can combine any
// Container IRNodes including ones that use off-the-shelf container images, and ones that generate their
// own container image (Dockerfile) onto the local filesystem.
//
// [docker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker

// The plugin allows using both off-the-shelf container images, and container images that have been
// auto-generated onto the local file system.  The plugin generates a docker-compose file that instantiates
// images, and sets environment variables and ports.
//
// in your local environment.
package dockercompose

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// Adds a child node to an existing container deployment
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string) {
	namespaceutil.AddNodeTo[Deployment](spec, deploymentName, containerName)
}

// Adds a deployment that explicitly instantiates all of the containers provided.
// The deployment will also implicitly instantiate any of the dependencies of the containers
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, containerName := range containers {
		AddContainerToDeployment(spec, deploymentName, containerName)
	}

	spec.Define(deploymentName, &Deployment{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		deployment := &Deployment{DeploymentName: deploymentName}
		_, err := namespaceutil.InstantiateNamespace(namespace, &DeploymentNamespace{deployment})
		return deployment, err
	})

	return deploymentName
}

// A [wiring.NamespaceHandler] used to build container deployments
type DeploymentNamespace struct {
	*Deployment
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) Accepts(nodeType any) bool {
	_, isDockerContainerNode := nodeType.(docker.Container)
	return isDockerContainerNode
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) AddEdge(name string, edge ir.IRNode) error {
	deployment.Edges = append(deployment.Edges, edge)
	return nil
}

// Implements [wiring.NamespaceHandler]
func (deployment *Deployment) AddNode(name string, node ir.IRNode) error {
	deployment.Nodes = append(deployment.Nodes, node)
	return nil
}
