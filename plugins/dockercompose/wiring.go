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
// To deploy an application-level service in a container, make sure you first deploy the service to a process
// (with the [goproc] plugin) and to a container image (with the [linuxcontainer] plugin)
//
// # Default Builder
//
// Instead of explicitly combining container instances into a deployment, the dockercompose plugin can be
// configured as the default builder for container instances, by calling [RegisterAsDefaultBuilder] in your wiring spec.
//
// At compile time Blueprint will combine any container instances that exist in the wiring spec but aren't explicitly added to
// a container deployment, and create a default docker-compose deployment with the name "docker".
//
//	dockercompose.RegisterAsDefaultBuilder()
//
// Calling [RegisterAsDefaultBuilder] is optional and usually unnecessary:
//   - If your wiring spec uses Blueprint's [cmdbuilder] then dockercompose is already registered as the default
//     container workspace builder.
//   - The default builder only takes effect if there are 1 or more container instances that haven't been added
//     to a container deployment.  If your wiring spec manually creates container deployments using [NewDeployment]
//     for all container instances, then the default builder will not have any effect.
//
// # Artifacts Generated
//
// During compilation, the plugin generates a docker-compose file that instantiates images for the specified
// containers.  The plugin also sets environment variables and ports for the instances.
//
// If your wiring spec only defines container instances, and dockercompose is registered as the default builder,
// then Blueprint will automatically generate a docker-compose deployment called "docker" that instantiates all
// of the container instances.
//
// # Running Artifacts
//
// If the dockercompose deployment is not further combined by other plugins, then the entry point to running
// your application will be using docker-compose.  You can build or run the deployment with:
//
//	docker compose build
//	docker compose up
//
// Many wiring specs generate .env files to set environment variables; you can point docker to these
// .env files as follows:
//
//	docker compose --env-file=../.local.env build
//
// If you aren't using an env file like the above, Docker might complaion about the absence of necessary
// environment variables.  You can manually write those to a .env file or set them in your local environment.
//
// For a concrete guide on running a generated docker-compose file, see the [SockShop Getting Started]
// documentation.
//
// # Internals
//
// Internally, the plugin makes use of interfaces defined in the [docker] plugin.  It can combine any
// Container IRNodes including ones that use off-the-shelf container images, and ones that generate their
// own container image (Dockerfile) onto the local filesystem.  Internally the plugin assigns hostnames
// to container instances and sets environment variables so that services call to the correct hostnames
// and ports.
//
// [docker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker
// [SockShop Getting Started]: https://github.com/Blueprint-uServices/blueprint/tree/main/examples/sockshop
// [linuxcontainer]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/linuxcontainer
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
// [cmdbuilder]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder
package dockercompose

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/docker"
)

// AddContainerToDeployment can be used by wiring specs to add a container instance to an existing
// container deployment.
func AddContainerToDeployment(spec wiring.WiringSpec, deploymentName, containerName string) {
	namespaceutil.AddNodeTo[Deployment](spec, deploymentName, containerName)
}

// NewDeployment can be used by wiring specs to create a container deployment that instantiates
// a number of containers.
//
// Further container instances can be added to the deployment by calling [AddContainerToDeployment].
//
// During compilation, generates a docker-compose file that instantiates the containers.
//
// Returns deploymentName.
func NewDeployment(spec wiring.WiringSpec, deploymentName string, containers ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, containerName := range containers {
		AddContainerToDeployment(spec, deploymentName, containerName)
	}

	spec.Define(deploymentName, &Deployment{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		deployment := &Deployment{DeploymentName: deploymentName}
		_, err := namespaceutil.InstantiateNamespace(namespace, &deploymentNamespace{deployment})
		return deployment, err
	})

	return deploymentName
}

// A [wiring.NamespaceHandler] used to build container deployments
type deploymentNamespace struct {
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
