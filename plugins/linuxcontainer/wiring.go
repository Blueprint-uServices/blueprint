// Package linuxcontainer is a plugin for instantiating multiple linux process instances in a single
// container deployment.
//
// # Wiring Spec Usage
//
// To use the linuxcontainer plugin in your wiring spec, you can declare a container, giving it a name and
// specifying which process instances to include
//
//	linuxcontainer.CreateContainer(spec, "my_container", "my_process_1", "my_process_2")
//
// You can also add processes to existing linux containers:
//
//	linuxcontainer.AddToContainer(spec, "my_container", "my_process_3")
//
// If you are only deploying a single service within the container, you should use the shorter [Deploy]:
//
//	linuxcontainer.Deploy(spec, "my_service")
//
// When a service is added to a container, the linuxcontainer plugin also adds a modifier to the service,
// so that the service is now converted from a process-level service to a container-level service.  Any
// process-level modifiers should be applied to the service *before* deploying it to a container.
//
// To deploy an application-level service to a container, make sure you first deploy the service to a process
// (e.g. with the [goproc] plugin) and prior to that (if desired) expose it over the network (e.g. with the
// [grpc] plugin)
//
// # Default Builder
//
// Instead of explicitly combining process instances into a linux container, the linuxcontainer plugin can be
// configured as the default builder for process instances, by calling [RegisterAsDefaultBuilder] in your wiring spec.
//
// At compile time Blueprint will combine any process instances that exist in the wiring spec but aren't explicitly added to
// a linux container, and create a default linux container deployment with the name "linux".
//
//	linuxcontainer.RegisterAsDefaultBuilder()
//
// Calling [RegisterAsDefaultBuilder] is optional and usually unnecessary:
//   - If your wiring spec uses Blueprint's [cmdbuilder] then linuxcontainer is already registered as the default
//     process workspace builder.
//   - The default builder only takes effect if there are 1 or more process instances that haven't been added
//     to a linux container.  If your wiring spec manually creates linux containers using [CreateContainer]
//     for all process instances, then the default builder will not have any effect.
//
// # Artifacts Generated
//
// During compilation, the plugin creates a directory to collect the artifacts of all processes contained
// therein.  For example, if one of the processes is a [goproc], then code for that goproc will be collected
// into a subdirectory of the container.
//
// The plugin also gathers run scripts and (optional) build scripts from all processes, and then generates
// scripts that, when invoked, will invoke all the build scripts, and invoke all the run scripts.
//
// The linuxcontainer plugin also implements some of the interfaces defined by the [docker] plugin and
// will generate Dockerfiles in the case when containers are added to a container deployment (e.g. Kubernetes
// or docker-compose)
//
// # Running artifacts
//
// A container's artifacts will be collected in a subdirectory of the build output based on the container
// name.  Navigate to this directory then invoke run.sh.  Only Linux is supported.
//
// Depending on the contents of the container, the run.sh might complain about missing environment variables
// such as addresses to bind to.  These should be set in the calling environment before invoking run.sh.
//
// [docker]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/docker
// [goproc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/goproc
// [grpc]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/grpc
// [cmdbuilder]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/cmdbuilder
package linuxcontainer

import (
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/namespaceutil"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

// AddToContainer can be used by wiring specs to add a process instance to an existing
// container deployment
func AddToContainer(spec wiring.WiringSpec, containerName, childName string) {
	namespaceutil.AddNodeTo[Container](spec, containerName, childName)
}

// Deploy can be used by wiring specs to deploy a process-level service in a linux container.
//
// Adds a modifier to the service that, during compilation, will create the linux container if
// not already created.
//
// The name of the container created is determined by attempting to replace a "_service" suffix
// with "_ctr", or adding "_ctr" if serviceName doesn't end with "_service", e.g.
//
//	user_service => user_ctr
//	user => user_ctr
//	user_srv => user_srv_ctr
//
// After calling [Deploy], serviceName will be a container-level service.
//
// Returns the name of the container
func Deploy(spec wiring.WiringSpec, serviceName string) string {
	servicePrefix, _ := strings.CutSuffix(serviceName, "_service")
	ctrName := servicePrefix + "_ctr"
	CreateContainer(spec, ctrName, serviceName)
	return ctrName
}

// CreateContainer can be used by wiring specs to define a container called containerName and to
// deploy the children processes.  CreateContainer only needs to be used when more than one children
// are being added to the container; otherwise it is more convenient to use [Deploy].
//
// After calling CreateContainer, other processes can still be added to the container by calling
// [AddToContainer] using the same containerName.
//
// After calling CreateContainer, any children that are services will become container-level services
// that can now have container-level modifiers applied to them, or can be added to container deployments
// like kubernetes pods.
func CreateContainer(spec wiring.WiringSpec, containerName string, children ...string) string {
	// If any children were provided in this call, add them to the process via a property
	for _, childName := range children {
		AddToContainer(spec, containerName, childName)
	}

	// A linux container node is simply a namespace that accumulates linux process nodes
	spec.Define(containerName, &Container{}, func(namespace wiring.Namespace) (ir.IRNode, error) {
		ctr := newLinuxContainerNode(containerName)
		_, err := namespaceutil.InstantiateNamespace(namespace, &linuxContainerNamespace{ctr})
		return ctr, err
	})

	return containerName
}

// A [wiring.NamespaceHandler] used to build golang process nodes
type linuxContainerNamespace struct {
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
