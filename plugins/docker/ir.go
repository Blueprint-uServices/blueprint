// Package docker defines compiler interfaces for use by plugins that generate and instantiate Docker images.
// The package does not provide any wiring spec functionality and is not directly used by Blueprint applications;
// only by other Blueprint plugins.
//
// The noteworthy interfaces are as follows:
//   - [Container] is an interface for IRNodes that represent containers.  If an IRNode implements
//     this interface then it will ultimately be instantiated in a namespace that supports containers,
//     such as a docker-compose file or a Kubernetes pod.
//   - If the container IRNode wants to generate or define a custom Docker image (e.g. using a Dockerfile),
//     then the IRNode should implement the [ProvidesContainerImage] interface.
//   - If the container IRNode wants to instantiate a Docker image (be it a pre-defined image, or a custom
//     image defined using [ProvidesContainerImage]), then the IRNode should implement the
//     [ProvidesContainerInstance] interface.
//
// Consult the following plugins for examples:
//   - Many backend plugins such as the memcached plugin provide prebuilt containers for the backends
//   - The linuxcontainer plugin generates custom Dockerfile images
//   - The dockercompose plugin implements a Container namespace that collects together Container nodes
//     and generates a docker-compose file
//   - The kubernetes plugin implements a Container namespace that collects together Container nodes and
//     generates YAML manifests
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
		AddContainerArtifacts(target ContainerWorkspace) error
	}

	// An optional interface for Container IRNodes to implement if the node
	// wants to declare an instance of a container.  The container instance
	// can be of a pre-existing image or of a locally-defined image that
	// was declared with [ProvidesContainerImage].
	ProvidesContainerInstance interface {
		AddContainerInstance(target ContainerWorkspace) error
	}
)

type (
	// Metadata about the local build environment used during the compilation process
	ContainerWorkspaceInfo struct {
		Path   string // fully-qualified path on the filesystem to the workspace
		Target string // the type of workspace being built
	}

	/*
		A container workspace has commands for adding container artifacts
		to the workspace and instantiating containers by providing

		An example concrete container workspace is a docker-compose file
	*/

	// A workspace during the Blueprint compilation process that collects together
	// container images and instances.
	//
	// A [ContainerWorkspace] instance will be provided to Container IRNodes that implement
	// either the [ProvidesContainerImage] or [ProvidesContainerInstance] interfaces.
	// The container IRNodes can invoke methods on this workspace in order to add their
	// artifacts to the build output.
	ContainerWorkspace interface {
		ir.BuildContext

		// Provides metadata about the workspace
		Info() ContainerWorkspaceInfo

		/*
			Creates a subdirectory in the workspace dir for a container node
			to collect its artifacts

			Returns a fully qualified path on the local filesystem where artifacts will be
			collected.

			The caller is responsible for then depositing artifacts in this directory
			(e.g. generating its Dockerfile there)
		*/
		CreateImageDir(imageName string) (string, error)

		/*
			Declares an instance of a container with a desired name, using
			a pre-existing image.

			The instanceName will be used as the container hostname.

			The IRnodes provided are considered arguments to the container.  If they
			are Config IRNodes, environment variables will be set for the container instance.
			If they are addresses, ports will be assigned and mapped.

			Returns an error if an instance already exists with this name.
		*/
		DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error

		/*
			Declares an instance of a container using container artifacts
			on the local filesystem.  The specified imageName must correspond
			to the imageName used in a previous call to CreateImageDir.

			The instanceName will be used as the container hostname.

			The IRnodes provided are considered arguments to the container.  If they
			are Config IRNodes, environment variables will be set for the container instance.
			If they are addresses, ports will be assigned and mapped.

			Returns an error if an instance already exists with this name.
		*/
		DeclareLocalImage(instanceName string, imageName string, args ...ir.IRNode) error

		/*
			Adds an environment variable to the container with name `instanceName`.

			The `instanceName` is the name of the container to add the environment variable to.

			The `key` is the name of the environment variable to be added to the container.

			The `val` is the value of the environment variable.

			Returns an error if an instance doesn't exist with the name `instanceName`.
		*/
		SetEnvironmentVariable(instanceName string, key string, val string) error

		/*
			Indicates that the caller has finished adding images and instances,
			and the workspace can generate any subsequent artifacts
			e.g. the final docker-compose file can be generated
		*/
		Finish() error

		ImplementsContainerWorkspace()
	}

	/*
		Docker also specializes the build process for linux containers.

		Concretely, processes can add custom build commands to the Dockerfile.

		We support docker multi-stage builds, where a process can build its
		binaries in one stage, and then we copy the built output into a
		different image in the next stage.

		To make use of this, a process that implements linux.ProvidesProcessArtifacts
		or linux.InstantiableProcess can typecheck the workspace argument
		to see if it is a docker.ProcessWorkspace instance.
	*/
	ProcessWorkspace interface {
		linux.ProcessWorkspace

		/*
			Called by a process IRNode to add custom Dockerfile commands to
			the Dockerfile.

			Specifically the commands are assumed to be part of a Docker
			multi-stage build
			https://docs.docker.com/build/building/multi-stage/

			If a process does NOT add Dockerfile commands, then by default
			the process artifacts will simply be copied into the container.

			If a process DOES add Dockerfile commands, then it is responsible
			for which artifacts make it into the container.

			The caller should assume the following:
			 * The Dockerfile will reside in the root of the process workspace
			 * Thus process artifacts will reside in the appropriate subdirectory
			   of the process workspace

			The caller must ensure the following:
			 * The Dockerfile commands should begin with
			 	FROM imagename AS {procname}
			 * {procname} must be the process IRnode's name
			 * Any build artifacts that should survive into the final container
			   must be placed in the /{procname} directory

			The generated Dockerfile will automatically copy built artifacts
			into the final container as follows:
				COPY --from={procname} /{procname} /{procname}
		*/
		AddDockerfileCommands(procName, commands string)

		ImplementsDockerProcessWorkspace()
	}
)
