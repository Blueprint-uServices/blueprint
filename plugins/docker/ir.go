package docker

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint/pkg/ir"
	"gitlab.mpi-sws.org/cld/blueprint/plugins/linux"
)

/*
The base IRNode interface for docker containers
*/
type Container interface {
	ir.IRNode
	ImplementsDockerContainer()
}

/*
Code and artifact generation interfaces that IRNodes
can implement to provide docker images
*/
type (
	/*
		For container nodes that want to provide code or other
		artifacts for their container.  Methods on the `builder` argument
		are used for collecting the artifacts
	*/
	ProvidesContainerImage interface {
		AddContainerArtifacts(target ContainerWorkspace) error
	}

	ProvidesContainerInstance interface {
		AddContainerInstance(target ContainerWorkspace) error
	}
)

/*
Builders used by the above code and artifact generation interfaces
*/
type (
	ContainerWorkspaceInfo struct {
		Path   string // fully-qualified path on the filesystem to the workspace
		Target string // the type of workspace being built
	}

	/*
		A container workspace has commands for adding container artifacts
		to the workspace and instantiating containers by providing

		An example concrete container workspace is a docker-compose file
	*/

	ContainerWorkspace interface {
		ir.BuildContext

		Info() ContainerWorkspaceInfo

		/*
			Creates a subdirectory in the workspace dir for a container node
			to collect its artifacts

			Returns a fully qualified path on the local filesystem where artifacts will be
			collected.
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
