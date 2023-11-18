---
title: plugins/docker
---
# plugins/docker
```go
package docker // import "gitlab.mpi-sws.org/cld/blueprint/plugins/docker"
```

## TYPES

The base IRNode interface for docker containers
```go
type Container interface {
	ir.IRNode
	ImplementsDockerContainer()
}
```
```go
type ContainerWorkspace interface {
	ir.BuildContext
```
```go
	Info() ContainerWorkspaceInfo
```
```go
	//			Creates a subdirectory in the workspace dir for a container node
	//			to collect its artifacts
	//
	//			Returns a fully qualified path on the local filesystem where artifacts will be
	//			collected.
```
```go
	CreateImageDir(imageName string) (string, error)
```
```go
	//			Declares an instance of a container with a desired name, using
	//			a pre-existing image.
	//
	//			The instanceName will be used as the container hostname.
	//
	//			The IRnodes provided are considered arguments to the container.  If they
	//			are Config IRNodes, environment variables will be set for the container instance.
	//			If they are addresses, ports will be assigned and mapped.
	//
	//			Returns an error if an instance already exists with this name.
```
```go
	DeclarePrebuiltInstance(instanceName string, image string, args ...ir.IRNode) error
```
```go
	//			Declares an instance of a container using container artifacts
	//			on the local filesystem.  The specified imageName must correspond
	//			to the imageName used in a previous call to CreateImageDir.
	//
	//			The instanceName will be used as the container hostname.
	//
	//			The IRnodes provided are considered arguments to the container.  If they
	//			are Config IRNodes, environment variables will be set for the container instance.
	//			If they are addresses, ports will be assigned and mapped.
	//
	//			Returns an error if an instance already exists with this name.
```
```go
	DeclareLocalImage(instanceName string, imageName string, args ...ir.IRNode) error
```
```go
	//			Indicates that the caller has finished adding images and instances,
	//			and the workspace can generate any subsequent artifacts
	//			e.g. the final docker-compose file can be generated
```
```go
	Finish() error
```
Builders used by the above code and artifact generation interfaces
```go
	ImplementsContainerWorkspace()
}
```
Builders used by the above code and artifact generation interfaces
```go
type ContainerWorkspaceInfo struct {
	Path   string // fully-qualified path on the filesystem to the workspace
	Target string // the type of workspace being built
}
```
```go
type ProcessWorkspace interface {
	linux.ProcessWorkspace
```
```go
	//			Called by a process IRNode to add custom Dockerfile commands to
	//			the Dockerfile.
	//
	//			Specifically the commands are assumed to be part of a Docker
	//			multi-stage build
	//			https://docs.docker.com/build/building/multi-stage/
	//
	//			If a process does NOT add Dockerfile commands, then by default
	//			the process artifacts will simply be copied into the container.
	//
	//			If a process DOES add Dockerfile commands, then it is responsible
	//			for which artifacts make it into the container.
	//
	//			The caller should assume the following:
	//			 * The Dockerfile will reside in the root of the process workspace
	//			 * Thus process artifacts will reside in the appropriate subdirectory
	//			   of the process workspace
	//
	//			The caller must ensure the following:
	//			 * The Dockerfile commands should begin with
	//			 	FROM imagename AS {procname}
	//			 * {procname} must be the process IRnode's name
	//			 * Any build artifacts that should survive into the final container
	//			   must be placed in the /{procname} directory
	//
	//			The generated Dockerfile will automatically copy built artifacts
	//			into the final container as follows:
	//				COPY --from={procname} /{procname} /{procname}
```
```go
	AddDockerfileCommands(procName, commands string)
```
Docker also specializes the build process for linux containers.
```go
	ImplementsDockerProcessWorkspace()
}
```
Concretely, processes can add custom build commands to the Dockerfile.

We support docker multi-stage builds, where a process can build its binaries
in one stage, and then we copy the built output into a different image in
the next stage.

To make use of this, a process that implements
linux.ProvidesProcessArtifacts or linux.InstantiableProcess can typecheck
the workspace argument to see if it is a docker.ProcessWorkspace instance.

For container nodes that want to provide code or other artifacts for their
container. Methods on the `builder` argument are used for collecting the
artifacts
```go
type ProvidesContainerImage interface {
	AddContainerArtifacts(target ContainerWorkspace) error
}
```
Code and artifact generation interfaces that IRNodes can implement to
provide docker images
```go
type ProvidesContainerInstance interface {
	AddContainerInstance(target ContainerWorkspace) error
}
```

