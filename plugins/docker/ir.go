// Package docker defines compiler interfaces for use by plugins that generate and instantiate Docker images.
//
// # Prerequisites
//
// In order to compile an application that uses Docker, the build machine must have Docker installed.  Follow
// the instructions on the [Docker website].  The person writing these instructions was using version 24.0.7.
//
// # Wiring Spec Usage
//
// The package does not provide any wiring spec functionality and is not directly used by Blueprint applications;
// only by other Blueprint plugins.
//
// The noteworthy interfaces are as follows:
//   - [Container] is an interface for IRNodes that represent containers.  If an IRNode implements
//     this interface then it will ultimately be instantiated in a namespace that supports containers,
//     such as a docker-compose file or a Kubernetes pod.
//   - If a [Container] wants to generate or define a custom Docker image (e.g. using a Dockerfile),
//     then the IRNode should implement the [ProvidesContainerImage] interface.
//   - If a [Container] wants to instantiate a Docker image (be it a pre-defined image, or a custom
//     image defined using [ProvidesContainerImage]), then the IRNode should implement the
//     [ProvidesContainerInstance] interface.
//
// Consult the following plugins for examples:
//   - Many backend plugins such as the [memcached] plugin provide prebuilt containers for the backends
//   - The [linuxcontainer] plugin generates custom Dockerfile images
//   - The [dockercompose] plugin implements a Container namespace that collects together Container nodes
//     and generates a docker-compose file
//   - The [kubernetes] plugin implements a Container namespace that collects together Container nodes and
//     generates YAML manifests
//
// [memcached]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/memcached
// [linuxcontainer]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/memclinuxcontainerached
// [dockercompose]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/dockercompose
// [kubernetes]: https://github.com/Blueprint-uServices/blueprint/tree/main/plugins/kubernetes
// [Docker website]: https://docs.docker.com/engine/install/
package docker

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/linux"
)

// An IRNode interface that represents containers.  If an IRNode implements this interface
// then it enables that IRNode to be instantiated within container namespaces such as docker-compose
// files and Kubernetes pods.
type Container interface {
	ir.IRNode
	ImplementsDockerContainer()
}

/*
Code and artifact generation interfaces that IRNodes
can implement to provide docker images
*/
type (
	// An optional interface for Container IRNodes to implement if the node needs
	// to generate custom container images (e.g. using a Dockerfile).
	// [target] provides methods for doing so.
	ProvidesContainerImage interface {
		// The IRNode is being compiled into the provided target workspace, and should
		// use methods on target to add its container artifacts into the workspace.
		AddContainerArtifacts(target ContainerWorkspace) error
	}

	// An optional interface for Container IRNodes to implement if the node
	// wants to declare an instance of a container.  The container instance
	// can be of a pre-existing image or of a locally-defined image that
	// was declared with [ProvidesContainerImage].
	ProvidesContainerInstance interface {
		// The IRNode is being compiled into the provided target workspace, and should
		// use methods on target to declare how the container should be instantiated.
		AddContainerInstance(target ContainerWorkspace) error
	}
)

type (
	// Metadata about the local build environment used during the compilation process
	ContainerWorkspaceInfo struct {
		Path   string // fully-qualified path on the filesystem to the workspace
		Target string // the type of workspace being built
	}

	// [ContainerWorkspace] receives container images and instances from [Container] nodes during
	// Blueprint's compilation process.
	//
	// A [ContainerWorkspace] instance will be provided to Container IRNodes that implement
	// either the [ProvidesContainerImage] or [ProvidesContainerInstance] interfaces.
	// The container IRNodes can invoke methods on this workspace in order to add their
	// artifacts to the build output.
	ContainerWorkspace interface {
		ir.BuildContext

		// Provides metadata about the workspace
		Info() ContainerWorkspaceInfo

		// Creates a subdirectory in the workspace dir for a container node
		// to collect its artifacts
		//
		// Returns a fully qualified path on the local filesystem where artifacts will be
		// collected.
		//
		// The caller is responsible for then depositing artifacts in this directory
		// (e.g. generating its Dockerfile there)
		CreateImageDir(imageName string) (string, error)

		// Declares an instance of a container with a desired name, using
		// a pre-existing image.
		//
		// The instanceName will be used as the container hostname.
		//
		// The IRnodes provided are considered arguments to the container.  If they
		// are Config IRNodes, environment variables will be set for the container instance.
		// If they are addresses, ports will be assigned and mapped.
		//
		// Returns an error if an instance already exists with this name.
		DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error

		// Declares an instance of a container using container artifacts
		// on the local filesystem.  The specified imageName must correspond
		// to the imageName used in a previous call to CreateImageDir.
		//
		// The instanceName will be used as the container hostname.
		//
		// The IRnodes provided are considered arguments to the container.  If they
		// are Config IRNodes, environment variables will be set for the container instance.
		// If they are addresses, ports will be assigned and mapped.
		//
		// Returns an error if an instance already exists with this name.
		DeclareLocalImage(instanceName string, imageName string, args ...ir.IRNode) error

		// Adds an environment variable to a container instance.  When instanceName is started,
		// key will be set to val in the container's environment.
		//
		// Returns an error if an instance doesn't exist with the name `instanceName`.
		SetEnvironmentVariable(instanceName string, key string, val string) error

		ImplementsContainerWorkspace()
	}

	// ProcessWorkspace enables [linux.Process] nodes to add custom Dockerfile commands when the process
	// is being added to a Docker container.  ProcessWorkspaces extends [linux.ProcessWorkspace] with the
	// method [ProcessWorkspace.AddDockerfileCommands].
	//
	// ProcessWorkspace can be used by any [linux.Process] node that implements [linux.InstantiableProcess]
	// and/or [linux.ProvidesProcessArtifacts].  The node should typecheck the [linux.ProcessWorkspace] to
	// determine if it is a ProcessWorkspace; if so, [ProcessWorkspace.AddDockerfileCommands] can be used.
	ProcessWorkspace interface {
		linux.ProcessWorkspace

		// Allows a [linux.Process] node to add custom Dockerfile build commands.
		//
		// By default, if a [linux.Process] node *doesn't* add custom Dockerfile commands,
		// then Blueprint will automatically copy all process artifacts into the container.
		//
		// However, if a [linux.Process] node *does* add custom Dockerfile build commands, then
		// the node is entirely responsible for copying its code artifacts into the container
		// image.
		//
		// The commands are assumed to be part of a Docker [multi-stage build].
		// The caller must ensure the following:
		//  * The Dockerfile commands should begin with
		//  	FROM imagename AS {procname}
		//  * {procname} must be the process IRnode's name
		//  * Any build artifacts that should survive into the final container
		//    must be placed in the /{procname} directory.
		//
		// The generated Dockerfile will automatically copy built artifacts
		// into the final container as follows:
		// 	COPY --from={procname} /{procname} /{procname}
		//
		// The caller should assume the following:
		//  * The Dockerfile will reside in the root of the process workspace
		//  * Thus process artifacts will reside in the appropriate subdirectory
		//    of the process workspace
		//
		// [multi-stage build]: https://docs.docker.com/build/building/multi-stage/
		AddDockerfileCommands(procName, commands string)

		ImplementsDockerProcessWorkspace()
	}
)
